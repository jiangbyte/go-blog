<!-- Author: go-blog — 内容标签列表 -->

<script setup lang="tsx">
import type { PaginationProps } from 'naive-ui'
import type { ProDataTableColumns, ProSearchFormColumns } from 'pro-naive-ui'
import { Icon } from '@iconify/vue/offline'
import { contentTagApi } from '@/api'
import { formatDateTime, hasPermission, normalizeSearchValues, renderButtonIcon } from '@/utils'
import { NButton, NFlex, NIcon } from 'naive-ui'
import { createProSearchForm, ProCard, ProDataTable, ProSearchForm } from 'pro-naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import ModalForm from './components/ModalForm.vue'

const formRef = ref<InstanceType<typeof ModalForm> | null>(null)
const state = reactive({
  records: [] as any[],
  total: 0,
  loading: false,
  searchValues: {} as { name?: string },
  page: 1,
  pageSize: 20,
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
    const res = await contentTagApi.page({
      current: state.page,
      size: state.pageSize,
      name: state.searchValues.name,
    })
    state.records = res.data?.records || []
    state.total = Number(res.data?.total || 0)
  } finally {
    state.loading = false
  }
}

async function onRemove(row: any) {
  window.$dialog.warning({
    title: '删除标签',
    content: `确定删除「${row.name}」？仍被文章引用时将失败。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await contentTagApi.remove([row.id])
      window.$message.success('已删除')
      await fetchPage()
    },
  })
}

const tableColumns = computed<ProDataTableColumns<any>>(() => [
  { title: '名称', path: 'name', minWidth: 140 },
  { title: '标识', path: 'slug', minWidth: 140 },
  { title: '排序', path: 'sort_code', width: 80 },
  {
    title: '更新时间',
    path: 'updated_at',
    width: 180,
    render: (row) => formatDateTime(row.updated_at),
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    fixed: 'right',
    render: (row) => (
      <NFlex size={12}>
        {hasPermission('admin:contenttag:update') ? (
          <NButton
            type="primary"
            text
            title="编辑"
            onClick={() => formRef.value?.openModal(row.id, row)}
          >
            {renderButtonIcon('icon-park-outline:edit')}
          </NButton>
        ) : null}
        {hasPermission('admin:contenttag:delete') ? (
          <NButton type="error" text title="删除" onClick={() => void onRemove(row)}>
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
      title="内容标签"
      row-key="id"
      :scroll-x="800"
      :columns="tableColumns"
      :data="state.records"
      :loading="state.loading"
      :pagination="pagination"
    >
      <template #toolbar>
        <NFlex>
          <NButton
            v-if="hasPermission('admin:contenttag:create')"
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

    <ModalForm ref="formRef" @saved="fetchPage" />
  </NFlex>
</template>
