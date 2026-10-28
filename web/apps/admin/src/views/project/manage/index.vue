<!-- Author: go-blog — 项目管理列表（可见性 + 分类；私有无发布操作） -->

<script setup lang="tsx">
import type { PaginationProps } from 'naive-ui'
import type { ProDataTableColumns, ProSearchFormColumns } from 'pro-naive-ui'
import { Icon } from '@iconify/vue/offline'
import { projectApi, projectCategoryApi } from '@/api'
import { formatDateTime, hasPermission, normalizeSearchValues, renderButtonIcon } from '@/utils'
import { NButton, NFlex, NIcon, NTag } from 'naive-ui'
import { createProSearchForm, ProCard, ProDataTable, ProSearchForm } from 'pro-naive-ui'
import { computed, onMounted, reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import StorageImage from '@/components/upload/StorageImage.vue'

const router = useRouter()
const route = useRoute()
const state = reactive({
  records: [] as any[],
  total: 0,
  loading: false,
  categoryOptions: [] as { label: string; value: string }[],
  searchValues: {} as {
    name?: string
    status?: string
    visibility?: string
    category_id?: string
  },
  page: 1,
  pageSize: 20,
})

const searchForm = createProSearchForm<any>({
  defaultCollapsed: true,
  onSubmit(values) {
    state.searchValues = normalizeSearchValues(values, {
      name: (value) => String(value).trim(),
      status: (value) => String(value).trim(),
      visibility: (value) => String(value).trim(),
      category_id: (value) => String(value).trim(),
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
  {
    title: '状态',
    path: 'status',
    field: 'select',
    fieldProps: {
      options: [
        { label: '草稿', value: 'draft' },
        { label: '已发布', value: 'published' },
      ],
      clearable: true,
    },
  },
  {
    title: '可见性',
    path: 'visibility',
    field: 'select',
    fieldProps: {
      options: [
        { label: '公开', value: 'public' },
        { label: '私有', value: 'private' },
      ],
      clearable: true,
    },
  },
  {
    title: '分类',
    path: 'category_id',
    field: 'select',
    fieldProps: {
      options: state.categoryOptions,
      clearable: true,
      filterable: true,
    },
  },
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

function flattenCategories(nodes: any[], prefix = ''): { label: string; value: string }[] {
  const out: { label: string; value: string }[] = []
  for (const n of nodes || []) {
    out.push({ label: `${prefix}${n.name}`, value: n.id })
    if (n.children?.length) {
      out.push(...flattenCategories(n.children, `${prefix}${n.name} / `))
    }
  }
  return out
}

async function loadCategories() {
  const res = await projectCategoryApi.tree()
  state.categoryOptions = flattenCategories(res.data || [])
}

async function fetchPage() {
  state.loading = true
  try {
    const res = await projectApi.page({
      current: state.page,
      size: state.pageSize,
      name: state.searchValues.name,
      status: state.searchValues.status,
      visibility: state.searchValues.visibility,
      category_id: state.searchValues.category_id,
    })
    state.records = res.data?.records || []
    state.total = Number(res.data?.total || 0)
  } finally {
    state.loading = false
  }
}

function openCreatePage() {
  void router.push('/project/manage/create')
}

function openEditPage(id: string) {
  void router.push({ path: '/project/manage/edit', query: { id } })
}

function formatRange(row: any) {
  const start = String(row.start_date || '').trim()
  const end = String(row.end_date || '').trim()
  if (!start && !end) return '—'
  if (start && !end) return `${start} ~ 至今`
  if (!start && end) return `~ ${end}`
  return `${start} ~ ${end}`
}

function isPrivate(row: any) {
  return String(row.visibility || 'public') === 'private'
}

async function onRemove(row: any) {
  window.$dialog.warning({
    title: '删除项目',
    content: `确定删除「${row.name}」？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await projectApi.remove([row.id])
      window.$message.success('已删除')
      await fetchPage()
    },
  })
}

/** 列表快捷改状态：拉详情后整单更新（对齐内容文章）。 */
async function setStatus(row: any, status: 'draft' | 'published') {
  if (!hasPermission('admin:project:update')) return
  if (isPrivate(row)) {
    window.$message.warning('私有项目不可发布')
    return
  }
  try {
    const detail = (await projectApi.detail(row.id)).data
    if (!detail) {
      window.$message.error('项目不存在')
      return
    }
    // 发布前校验必填，避免后端抛错、误把草稿直接顶上线
    if (status === 'published') {
      if (!String(detail.name || '').trim() || detail.name === '未命名项目') {
        window.$message.warning('请先编辑并填写项目名称后再发布')
        return
      }
      if (!String(detail.description || '').trim()) {
        window.$message.warning('请先编辑并填写项目描述后再发布')
        return
      }
      if (!detail.category_id) {
        window.$message.warning('请先编辑并选择分类后再发布')
        return
      }
    }
    await projectApi.update({
      id: row.id,
      name: detail.name,
      slug: detail.slug,
      tech_stack: Array.isArray(detail.tech_stack) ? detail.tech_stack : [],
      role: detail.role || null,
      description: detail.description || '',
      images: Array.isArray(detail.images) ? detail.images : [],
      start_date: detail.start_date || null,
      end_date: detail.end_date || null,
      status,
      visibility: detail.visibility || 'public',
      category_id: detail.category_id || null,
      sort_code: Number(detail.sort_code ?? 0) || 0,
    })
    window.$message.success(status === 'published' ? '已发布' : '已撤回为草稿')
    await fetchPage()
  } catch (e: any) {
    window.$message.error(e?.message || '状态更新失败')
  }
}

/** 已发布 → 草稿（撤回）。 */
function onWithdraw(row: any) {
  if (!hasPermission('admin:project:update')) return
  window.$dialog.warning({
    title: '撤回项目',
    content: `确定将「${row.name}」撤回为草稿？撤回后前台不可见。`,
    positiveText: '撤回',
    negativeText: '取消',
    onPositiveClick: async () => {
      await setStatus(row, 'draft')
    },
  })
}

const tableColumns = computed<ProDataTableColumns<any>>(() => [
  {
    title: '封面',
    key: 'cover',
    width: 72,
    render: (row) => {
      const objectName = String(row.cover_object_name || '').trim()
      const src = String(row.cover_url || '').trim()
      if (!objectName && !src) return '—'
      return (
        <StorageImage
          objectName={objectName || null}
          src={src || null}
          alt={row.name || '封面'}
          width={48}
          height={32}
          objectFit="cover"
        />
      )
    },
  },
  { title: '名称', path: 'name', minWidth: 160, ellipsis: { tooltip: true } },
  {
    title: '分类',
    path: 'category_name',
    width: 120,
    ellipsis: { tooltip: true },
    render: (row) => row.category_name || '—',
  },
  {
    title: '技术栈',
    key: 'tech_stack',
    width: 180,
    ellipsis: { tooltip: true },
    render: (row) => {
      const tags = Array.isArray(row.tech_stack) ? row.tech_stack : []
      if (!tags.length) return '—'
      return tags.slice(0, 4).join('、') + (tags.length > 4 ? ` +${tags.length - 4}` : '')
    },
  },
  { title: '负责身份', path: 'role', width: 110, ellipsis: { tooltip: true } },
  {
    title: '时间范围',
    key: 'date_range',
    width: 200,
    render: (row) => formatRange(row),
  },
  {
    title: '可见性',
    key: 'visibility',
    width: 80,
    render: (row) => (
      <NTag size="small" type={isPrivate(row) ? 'warning' : 'info'} bordered={false}>
        {isPrivate(row) ? '私有' : '公开'}
      </NTag>
    ),
  },
  {
    title: '状态',
    key: 'status',
    width: 88,
    render: (row) => {
      if (isPrivate(row)) {
        return (
          <NTag size="small" bordered={false}>
            仅保存
          </NTag>
        )
      }
      return (
        <NTag
          size="small"
          type={row.status === 'published' ? 'success' : 'default'}
          bordered={false}
        >
          {row.status === 'published' ? '已发布' : '草稿'}
        </NTag>
      )
    },
  },
  { title: '排序', path: 'sort_code', width: 70 },
  {
    title: '更新时间',
    path: 'updated_at',
    width: 170,
    render: (row) => formatDateTime(row.updated_at),
  },
  {
    title: '操作',
    key: 'actions',
    width: 140,
    fixed: 'right',
    render: (row) => (
      <NFlex size={12} align="center" wrap={false}>
        {hasPermission('admin:project:update') && !isPrivate(row) && row.status !== 'published' ? (
          <NButton
            type="success"
            text
            title="发布"
            onClick={() => void setStatus(row, 'published')}
          >
            {renderButtonIcon('icon-park-outline:send')}
          </NButton>
        ) : null}
        {hasPermission('admin:project:update') && !isPrivate(row) && row.status === 'published' ? (
          <NButton type="warning" text title="撤回为草稿" onClick={() => void onWithdraw(row)}>
            {renderButtonIcon('icon-park-outline:undo')}
          </NButton>
        ) : null}
        {hasPermission('admin:project:update') ? (
          <NButton type="primary" text title="编辑" onClick={() => openEditPage(row.id)}>
            {renderButtonIcon('icon-park-outline:edit')}
          </NButton>
        ) : null}
        {hasPermission('admin:project:delete') ? (
          <NButton type="error" text title="删除" onClick={() => void onRemove(row)}>
            {renderButtonIcon('icon-park-outline:delete')}
          </NButton>
        ) : null}
      </NFlex>
    ),
  },
])

onMounted(() => {
  void loadCategories()
  void fetchPage()
})

watch(
  () => route.fullPath,
  () => {
    void fetchPage()
  },
)
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
      title="项目管理"
      row-key="id"
      :scroll-x="1400"
      :columns="tableColumns"
      :data="state.records"
      :loading="state.loading"
      :pagination="pagination"
    >
      <template #toolbar>
        <NFlex>
          <NButton
            v-if="hasPermission('admin:project:create')"
            type="primary"
            text
            title="新增"
            aria-label="新增"
            @click="openCreatePage"
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
  </NFlex>
</template>
