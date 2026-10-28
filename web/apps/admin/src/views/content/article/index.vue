<!-- Author: go-blog — 内容文章列表 -->

<script setup lang="tsx">
import type { PaginationProps } from 'naive-ui'
import type { ProDataTableColumns, ProSearchFormColumns } from 'pro-naive-ui'
import { Icon } from '@iconify/vue/offline'
import { contentArticleApi, contentCategoryApi, contentRecommendApi, contentTagApi } from '@/api'
import { formatDateTime, hasPermission, normalizeSearchValues, renderButtonIcon } from '@/utils'
import StorageImage from '@/components/upload/StorageImage.vue'
import { NButton, NFlex, NIcon, NTag } from 'naive-ui'
import { createProSearchForm, ProCard, ProDataTable, ProSearchForm } from 'pro-naive-ui'
import { computed, onMounted, reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const router = useRouter()
const route = useRoute()
const state = reactive({
  records: [] as any[],
  total: 0,
  loading: false,
  searchValues: {} as {
    title?: string
    status?: string
    category_id?: string
    tag_id?: string
  },
  page: 1,
  pageSize: 20,
  categoryOptions: [] as { label: string; value: string }[],
  tagOptions: [] as { label: string; value: string }[],
})

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

const searchForm = createProSearchForm<any>({
  defaultCollapsed: true,
  onSubmit(values) {
    state.searchValues = normalizeSearchValues(values, {
      title: (value) => String(value).trim(),
      status: (value) => String(value).trim(),
      category_id: (value) => String(value).trim(),
      tag_id: (value) => String(value).trim(),
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
  { title: '标题', path: 'title', field: 'input' },
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
    title: '分类',
    path: 'category_id',
    field: 'select',
    fieldProps: { options: state.categoryOptions, clearable: true, filterable: true },
  },
  {
    title: '标签',
    path: 'tag_id',
    field: 'select',
    fieldProps: { options: state.tagOptions, clearable: true, filterable: true },
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

async function loadFilters() {
  const [treeRes, tagRes] = await Promise.all([contentCategoryApi.tree(), contentTagApi.options()])
  state.categoryOptions = flattenCategories(treeRes.data || [])
  state.tagOptions = (tagRes.data || []).map((t) => ({ label: t.name, value: t.id }))
}

async function fetchPage() {
  state.loading = true
  try {
    const res = await contentArticleApi.page({
      current: state.page,
      size: state.pageSize,
      title: state.searchValues.title,
      status: state.searchValues.status,
      category_id: state.searchValues.category_id,
      tag_id: state.searchValues.tag_id,
    })
    state.records = res.data?.records || []
    state.total = Number(res.data?.total || 0)
  } finally {
    state.loading = false
  }
}

function openCreatePage() {
  void router.push('/content/article/create')
}

function openEditPage(id: string) {
  void router.push({ path: '/content/article/edit', query: { id } })
}

async function onRemove(row: any) {
  window.$dialog.warning({
    title: '删除文章',
    content: `确定删除「${row.title}」？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await contentArticleApi.remove([row.id])
      window.$message.success('已删除')
      await fetchPage()
    },
  })
}

/** 列表快捷改状态：拉详情后整单更新。 */
async function setStatus(row: any, status: 'draft' | 'published') {
  if (!hasPermission('admin:article:update')) return
  try {
    const detail = (await contentArticleApi.detail(row.id)).data
    if (!detail) {
      window.$message.error('文章不存在')
      return
    }
    if (status === 'published' && !detail.category_id) {
      window.$message.warning('请先编辑文章并选择分类后再发布')
      return
    }
    await contentArticleApi.update({
      id: row.id,
      category_id: detail.category_id || null,
      title: detail.title,
      summary: detail.summary || null,
      cover_url: detail.cover_object_name || detail.cover_url || null,
      body_md: detail.body_md,
      status,
      tag_ids: Array.isArray(detail.tag_ids) ? detail.tag_ids : [],
      sort_code: Number(detail.sort_code ?? 0) || 0,
    })
    window.$message.success(status === 'published' ? '已发布' : '已撤回为草稿')
    await fetchPage()
  } catch (e: any) {
    window.$message.error(e?.message || '状态更新失败')
  }
}

/** 已发布 → 草稿（撤回）；会同步移出推荐位。 */
function onWithdraw(row: any) {
  if (!hasPermission('admin:article:update')) return
  window.$dialog.warning({
    title: '撤回文章',
    content: `确定将「${row.title}」撤回为草稿？撤回后前台不可见，若在推荐位中将一并移除。`,
    positiveText: '撤回',
    negativeText: '取消',
    onPositiveClick: async () => {
      await setStatus(row, 'draft')
    },
  })
}

/** 已发布文章加入 / 移出推荐位。 */
async function toggleRecommend(row: any) {
  const need = row.in_recommend ? 'admin:contentrecommend:delete' : 'admin:contentrecommend:create'
  if (!hasPermission(need)) return
  if (row.status !== 'published') {
    window.$message.warning('仅已发布文章可加入推荐位')
    return
  }
  try {
    if (row.in_recommend) {
      await contentRecommendApi.remove(row.id)
      window.$message.success('已移出推荐位')
    } else {
      await contentRecommendApi.add(row.id)
      window.$message.success('已加入推荐位')
    }
    await fetchPage()
  } catch (e: any) {
    window.$message.error(e?.message || '操作失败')
  }
}

function renderCover(row: any) {
  const objectName = String(row.cover_object_name || '').trim()
  const src = String(row.cover_url || '').trim()
  if (!objectName && !src) return '—'
  return (
    <StorageImage
      objectName={objectName || null}
      src={src || null}
      alt={row.title || '封面'}
      width={48}
      height={32}
      objectFit="cover"
    />
  )
}

const tableColumns = computed<ProDataTableColumns<any>>(() => [
  {
    title: '封面',
    key: 'cover',
    width: 72,
    render: (row) => renderCover(row),
  },
  { title: '标题', path: 'title', minWidth: 180, ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 88,
    render: (row) => (
      <NTag size="small" type={row.status === 'published' ? 'success' : 'default'} bordered={false}>
        {row.status === 'published' ? '已发布' : '草稿'}
      </NTag>
    ),
  },
  { title: '浏览', path: 'view_count', width: 80 },
  {
    title: '更新时间',
    path: 'updated_at',
    width: 180,
    render: (row) => formatDateTime(row.updated_at),
  },
  {
    title: '推荐',
    key: 'in_recommend',
    width: 72,
    render: (row) =>
      row.in_recommend ? (
        <NTag size="small" type="warning" bordered={false}>
          是
        </NTag>
      ) : (
        '—'
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    fixed: 'right',
    render: (row) => (
      <NFlex size={12} align="center">
        {hasPermission('admin:article:update') && row.status !== 'published' ? (
          <NButton
            type="success"
            text
            title="发布"
            onClick={() => void setStatus(row, 'published')}
          >
            {renderButtonIcon('icon-park-outline:send')}
          </NButton>
        ) : null}
        {hasPermission('admin:article:update') && row.status === 'published' ? (
          <NButton type="warning" text title="撤回为草稿" onClick={() => void onWithdraw(row)}>
            {renderButtonIcon('icon-park-outline:undo')}
          </NButton>
        ) : null}
        {(
          row.in_recommend
            ? hasPermission('admin:contentrecommend:delete')
            : hasPermission('admin:contentrecommend:create')
        ) ? (
          <NButton
            type={row.in_recommend ? 'warning' : 'primary'}
            text
            disabled={row.status !== 'published'}
            title={
              row.status !== 'published'
                ? '请先发布后再加入推荐位'
                : row.in_recommend
                  ? '移出推荐位'
                  : '加入推荐位'
            }
            onClick={() => void toggleRecommend(row)}
          >
            {renderButtonIcon('icon-park-outline:like')}
          </NButton>
        ) : null}
        {hasPermission('admin:article:update') ? (
          <NButton type="primary" text title="编辑" onClick={() => openEditPage(row.id)}>
            {renderButtonIcon('icon-park-outline:edit')}
          </NButton>
        ) : null}
        {hasPermission('admin:article:delete') ? (
          <NButton type="error" text title="删除" onClick={() => void onRemove(row)}>
            {renderButtonIcon('icon-park-outline:delete')}
          </NButton>
        ) : null}
      </NFlex>
    ),
  },
])

onMounted(() => {
  void loadFilters()
  void fetchPage()
})

/** 同页 query 变化或再次进入时补拉，避免仍看到离开前的旧 records */
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
      title="内容文章"
      row-key="id"
      :scroll-x="1100"
      :columns="tableColumns"
      :data="state.records"
      :loading="state.loading"
      :pagination="pagination"
    >
      <template #toolbar>
        <NFlex>
          <NButton
            v-if="hasPermission('admin:article:create')"
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
