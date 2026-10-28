<!-- Author: go-blog — 单一 kind 的分类/标签列表（对齐 mypage 标签管理） -->

<script setup lang="tsx">
import type { PaginationProps } from 'naive-ui'
import type { ProDataTableColumns, ProSearchFormColumns } from 'pro-naive-ui'
import type { TaxonomyKind } from '@/api/blog/taxonomy'
import { Icon } from '@iconify/vue/offline'
import { blogTaxonomyApi } from '@/api'
import { formatDateTime, hasPermission, normalizeSearchValues, renderButtonIcon } from '@/utils'
import { NButton, NFlex, NIcon } from 'naive-ui'
import { createProSearchForm, ProCard, ProDataTable, ProSearchForm } from 'pro-naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import ModalForm from './components/ModalForm.vue'

const props = defineProps<{
  kind: TaxonomyKind
  title: string
}>()

const formRef = ref<InstanceType<typeof ModalForm> | null>(null)
const state = reactive({
  records: [] as any[],
  total: 0,
  loading: false,
  searchValues: {} as { name?: string },
  page: 1,
  pageSize: 20,
  checkedKeys: [] as string[],
})

const searchForm = createProSearchForm<any>({
  defaultCollapsed: true,
  onSubmit(values) {
    state.searchValues = normalizeSearchValues(values, {
      name: (value) => String(value).trim(),
    })
    state.page = 1
    void fetchPage()
  },
  onReset() {
    state.searchValues = {}
    state.page = 1
    void fetchPage()
  },
})

const searchColumns = computed<ProSearchFormColumns<any>>(() => [
  { title: '名称', path: 'name', field: 'input' },
])

const pagination = computed<PaginationProps>(() => ({
  page: state.page,
  pageSize: state.pageSize,
  itemCount: state.total,
  showSizePicker: true,
  pageSizes: [10, 20, 30, 50],
  prefix: ({ itemCount }) => `${itemCount} 条`,
  onUpdatePage: (value) => {
    state.page = value
    void fetchPage()
  },
  onUpdatePageSize: (value) => {
    state.pageSize = value
    state.page = 1
    void fetchPage()
  },
}))

async function fetchPage() {
  state.loading = true
  try {
    const res = await blogTaxonomyApi.page({
      kind: props.kind,
      current: state.page,
      size: state.pageSize,
      name: state.searchValues.name,
    })
    state.records = res.data?.records || []
    state.total = Number(res.data?.total || 0)
    state.checkedKeys = []
  } finally {
    state.loading = false
  }
}

async function onRemove(uids: string[]) {
  if (!uids.length) return
  const single = uids.length === 1 ? state.records.find((r) => r.uid === uids[0]) : null
  window.$dialog.warning({
    title: uids.length > 1 ? '批量删除' : `删除${props.title}`,
    content: single
      ? `确定删除「${single.name}」？仍被文章或项目引用时将失败。`
      : `确定删除选中的 ${uids.length} 个${props.title}？仍被文章或项目引用时将失败。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await blogTaxonomyApi.remove(uids)
      state.checkedKeys = []
      window.$message.success('已删除')
      await fetchPage()
    },
  })
}

const usageColumnTitle = computed(() => (props.kind === 'project_tag' ? '项目数' : '文章数'))

const tableColumns = computed<ProDataTableColumns<any>>(() => [
  { type: 'selection', width: 48 },
  { title: '名称', path: 'name', minWidth: 160 },
  {
    title: usageColumnTitle.value,
    key: 'usage_count',
    width: 100,
    render: (row) => Number(row.usage_count ?? 0),
  },
  { title: 'UID', path: 'uid', minWidth: 220, ellipsis: { tooltip: true } },
  {
    title: '创建时间',
    path: 'created_at',
    width: 180,
    render: (row) => formatDateTime(row.created_at),
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    fixed: 'right',
    render: (row) => (
      <NFlex size={12}>
        {hasPermission('admin:blog:taxonomy:update') ? (
          <NButton
            type="primary"
            text
            title="编辑"
            onClick={() => formRef.value?.openModal(row)}
          >
            {renderButtonIcon('icon-park-outline:edit')}
          </NButton>
        ) : null}
        {hasPermission('admin:blog:taxonomy:delete') ? (
          <NButton type="error" text title="删除" onClick={() => void onRemove([row.uid])}>
            {renderButtonIcon('icon-park-outline:delete')}
          </NButton>
        ) : null}
      </NFlex>
    ),
  },
])

onMounted(() => {
  void fetchPage()
})
</script>

<template>
  <NFlex class="h-full min-h-0" vertical>
    <ProCard>
      <ProSearchForm
        :form="searchForm"
        :columns="searchColumns"
        :reset-button-props="{ content: '重置' }"
        :search-button-props="{ content: '查询' }"
        :collapse-button-props="{ content: searchForm.collapsed.value ? '展开' : '收起' }"
      />
    </ProCard>

    <ProDataTable
      class="min-h-0 flex-1"
      remote
      :title="title"
      row-key="uid"
      :scroll-x="948"
      :columns="tableColumns"
      :data="state.records"
      :loading="state.loading"
      :pagination="pagination"
      :checked-row-keys="state.checkedKeys"
      :on-update-checked-row-keys="(keys) => (state.checkedKeys = keys as string[])"
    >
      <template #toolbar>
        <NFlex>
          <NButton
            v-if="hasPermission('admin:blog:taxonomy:create')"
            type="primary"
            text
            title="新增"
            aria-label="新增"
            @click="formRef?.openModal()"
          >
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:plus" />
              </NIcon>
            </template>
          </NButton>
          <NButton
            v-if="hasPermission('admin:blog:taxonomy:delete')"
            type="error"
            text
            title="批量删除"
            aria-label="批量删除"
            :disabled="!state.checkedKeys.length"
            @click="onRemove(state.checkedKeys)"
          >
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:delete" />
              </NIcon>
            </template>
          </NButton>
          <NButton text title="刷新" aria-label="刷新" :loading="state.loading" @click="fetchPage">
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:reload" />
              </NIcon>
            </template>
          </NButton>
        </NFlex>
      </template>
    </ProDataTable>

    <ModalForm ref="formRef" :kind="kind" :kind-label="title" @saved="fetchPage" />
  </NFlex>
</template>
