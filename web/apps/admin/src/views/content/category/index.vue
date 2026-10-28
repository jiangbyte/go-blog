<!-- Author: go-blog — 内容分类树表（对齐标签/素材管理页） -->

<script setup lang="tsx">
import type { ProDataTableColumns, ProSearchFormColumns } from 'pro-naive-ui'
import { Icon } from '@iconify/vue/offline'
import { contentCategoryApi } from '@/api'
import { formatDateTime, hasPermission, normalizeSearchValues, renderButtonIcon } from '@/utils'
import { NButton, NFlex, NIcon } from 'naive-ui'
import { createProSearchForm, ProCard, ProDataTable, ProSearchForm } from 'pro-naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import ModalForm from './components/ModalForm.vue'

const formRef = ref<InstanceType<typeof ModalForm> | null>(null)
const state = reactive({
  tree: [] as any[],
  loading: false,
  searchValues: {} as { name?: string },
})

const searchForm = createProSearchForm<any>({
  defaultCollapsed: true,
  onSubmit(values) {
    state.searchValues = normalizeSearchValues(values, {
      name: (value) => String(value).trim(),
    })
  },
  onReset() {
    state.searchValues = {}
  },
})

const searchColumns = computed<ProSearchFormColumns<any>>(() => [
  { title: '名称', path: 'name', field: 'input' },
])

/** 按名称过滤树，命中祖先时保留其子树。 */
function filterTree(nodes: any[], keyword: string): any[] {
  const kw = keyword.trim().toLowerCase()
  if (!kw) return nodes
  const out: any[] = []
  for (const node of nodes || []) {
    const children = filterTree(node.children || [], kw)
    if (
      String(node.name || '')
        .toLowerCase()
        .includes(kw) ||
      children.length
    ) {
      out.push({ ...node, children })
    }
  }
  return out
}

const displayRecords = computed(() => filterTree(state.tree, state.searchValues.name || ''))

async function fetchTree() {
  state.loading = true
  try {
    const res = await contentCategoryApi.tree()
    state.tree = res.data || []
  } finally {
    state.loading = false
  }
}

async function onRemove(row: any) {
  window.$dialog.warning({
    title: '删除分类',
    content: `确定删除「${row.name}」？有子分类或文章时将失败。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await contentCategoryApi.remove([row.id])
      window.$message.success('已删除')
      await fetchTree()
    },
  })
}

const tableColumns = computed<ProDataTableColumns<any>>(() => [
  { title: '名称', path: 'name', minWidth: 180 },
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
    width: 140,
    fixed: 'right',
    render: (row) => (
      <NFlex size={12}>
        {hasPermission('admin:contentcategory:create') ? (
          <NButton
            type="primary"
            text
            title="新增子级"
            onClick={() => formRef.value?.openModal(undefined, undefined, row.id)}
          >
            {renderButtonIcon('icon-park-outline:add-one')}
          </NButton>
        ) : null}
        {hasPermission('admin:contentcategory:update') ? (
          <NButton
            type="primary"
            text
            title="编辑"
            onClick={() => formRef.value?.openModal(row.id, row)}
          >
            {renderButtonIcon('icon-park-outline:edit')}
          </NButton>
        ) : null}
        {hasPermission('admin:contentcategory:delete') ? (
          <NButton type="error" text title="删除" onClick={() => void onRemove(row)}>
            {renderButtonIcon('icon-park-outline:delete')}
          </NButton>
        ) : null}
      </NFlex>
    ),
  },
])

onMounted(() => {
  void fetchTree()
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
      title="内容分类"
      row-key="id"
      children-key="children"
      default-expand-all
      :scroll-x="800"
      :columns="tableColumns"
      :data="displayRecords"
      :loading="state.loading"
      :pagination="false"
    >
      <template #toolbar>
        <NFlex>
          <NButton
            v-if="hasPermission('admin:contentcategory:create')"
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
          <NButton text title="刷新" aria-label="刷新" :loading="state.loading" @click="fetchTree">
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:reload" />
              </NIcon>
            </template>
          </NButton>
        </NFlex>
      </template>
    </ProDataTable>

    <ModalForm ref="formRef" @saved="fetchTree" />
  </NFlex>
</template>
