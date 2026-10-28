<!-- Author: go-blog — 博客文章列表 -->

<script setup lang="tsx">
import type { PaginationProps } from 'naive-ui'
import type { ProDataTableColumns, ProSearchFormColumns } from 'pro-naive-ui'
import { Icon } from '@iconify/vue/offline'
import { blogPostsApi } from '@/api'
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
  searchValues: {} as { title?: string; status?: string },
  page: 1,
  pageSize: 20,
  checkedKeys: [] as string[],
})

const searchForm = createProSearchForm<any>({
  defaultCollapsed: true,
  onSubmit(values) {
    state.searchValues = normalizeSearchValues(values, {
      title: (value) => String(value).trim(),
      status: (value) => String(value).trim(),
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
    const res = await blogPostsApi.page({
      current: state.page,
      size: state.pageSize,
      title: state.searchValues.title,
      status: state.searchValues.status,
    })
    state.records = res.data?.records || []
    state.total = Number(res.data?.total || 0)
    state.checkedKeys = []
  } finally {
    state.loading = false
  }
}

function openCreatePage() {
  void router.push('/blog/posts/create')
}

function openEditPage(uid: string) {
  void router.push({ path: '/blog/posts/edit', query: { uid } })
}

async function onRemove(uids: string[]) {
  if (!uids.length) return
  const single = uids.length === 1 ? state.records.find((r) => r.uid === uids[0]) : null
  window.$dialog.warning({
    title: uids.length > 1 ? '批量删除' : '删除文章',
    content: single
      ? `确定删除「${single.title}」？`
      : `确定删除选中的 ${uids.length} 篇文章？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await blogPostsApi.remove(uids)
      state.checkedKeys = []
      window.$message.success('已删除')
      await fetchPage()
    },
  })
}

function summarizeBatchStatus(
  data: { updated?: number; skipped?: number; failed?: Array<{ title?: string; uid: string; reason: string }> },
  unit: string,
) {
  const updated = Number(data?.updated || 0)
  const skipped = Number(data?.skipped || 0)
  const failed = Array.isArray(data?.failed) ? data.failed : []
  const parts = [`成功 ${updated} ${unit}`]
  if (skipped) parts.push(`跳过 ${skipped} ${unit}`)
  if (failed.length) parts.push(`失败 ${failed.length} ${unit}`)
  if (failed.length) {
    const detail = failed
      .slice(0, 3)
      .map((f) => `${f.title || f.uid}（${f.reason}）`)
      .join('；')
    window.$message.warning(`${parts.join('，')}：${detail}`)
  } else {
    window.$message.success(parts.join('，'))
  }
}

/** 列表行内快捷改状态：走后端 batch-status（单 uid）。 */
async function setStatus(row: any, status: 'draft' | 'published') {
  if (!hasPermission('admin:blog:post:update')) return
  try {
    const res = await blogPostsApi.batchStatus({ uids: [row.uid], status })
    const data = res.data || { updated: 0, skipped: 0, failed: [] }
    if (data.failed?.length) {
      window.$message.warning(data.failed[0].reason || '状态更新失败')
      return
    }
    window.$message.success(status === 'published' ? '已发布' : '已撤回为草稿')
    await fetchPage()
  } catch (e: any) {
    window.$message.error(e?.message || '状态更新失败')
  }
}

function onWithdraw(row: any) {
  if (!hasPermission('admin:blog:post:update')) return
  window.$dialog.warning({
    title: '撤回文章',
    content: `确定将「${row.title}」撤回为草稿？撤回后前台不可见。`,
    positiveText: '撤回',
    negativeText: '取消',
    onPositiveClick: async () => {
      await setStatus(row, 'draft')
    },
  })
}

/** 批量发布 / 撤回：一次请求后端 batch-status。 */
async function batchSetStatus(status: 'draft' | 'published') {
  if (!hasPermission('admin:blog:post:update')) return
  const uids = [...state.checkedKeys]
  if (!uids.length) return
  const publish = status === 'published'
  window.$dialog.warning({
    title: publish ? '批量发布' : '批量撤回',
    content: publish
      ? `确定发布选中的 ${uids.length} 篇文章？已发布的将跳过，无正文的无法发布。`
      : `确定将选中的 ${uids.length} 篇文章撤回为草稿？已是草稿的将跳过。`,
    positiveText: publish ? '发布' : '撤回',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await blogPostsApi.batchStatus({ uids, status })
        state.checkedKeys = []
        summarizeBatchStatus(res.data || {}, '篇')
        await fetchPage()
      } catch (e: any) {
        window.$message.error(e?.message || '批量操作失败')
      }
    },
  })
}

function renderCover(row: any) {
  const src = String(row.cover || '').trim()
  if (!src) return '—'
  return (
    <StorageImage
      objectName={/^(https?:|\/)/i.test(src) ? undefined : src}
      src={src}
      alt={row.title || '封面'}
      width={48}
      height={32}
      objectFit="cover"
    />
  )
}

function renderTaxonomyTags(items: any[], opts?: { nowrap?: boolean; max?: number }) {
  if (!items.length) return '—'
  const max = opts?.max ?? 4
  const nowrap = opts?.nowrap ?? false
  const shown = items.slice(0, max)
  const tip = items.map((item: any) => item.name || item).join('、')
  return (
    <NFlex
      size={4}
      wrap={!nowrap}
      style={
        nowrap
          ? { flexWrap: 'nowrap', overflow: 'hidden', whiteSpace: 'nowrap', maxWidth: '100%' }
          : undefined
      }
      title={tip}
    >
      {shown.map((item: any) => (
        <NTag key={item.uid || item.name} size="small" bordered={false} style={{ flexShrink: 0 }}>
          {item.name || item}
        </NTag>
      ))}
      {items.length > max ? (
        <NTag size="small" bordered={false} style={{ flexShrink: 0 }}>
          +{items.length - max}
        </NTag>
      ) : null}
    </NFlex>
  )
}

function renderCategories(row: any) {
  const items = Array.isArray(row.categories)
    ? row.categories
    : Array.isArray(row.category_names)
      ? row.category_names.map((name: string) => ({ name }))
      : []
  return renderTaxonomyTags(items, { max: 3 })
}

function renderTags(row: any) {
  const items = Array.isArray(row.tags)
    ? row.tags
    : Array.isArray(row.tag_names)
      ? row.tag_names.map((name: string) => ({ name }))
      : []
  return renderTaxonomyTags(items, { nowrap: true, max: 6 })
}

const tableColumns = computed<ProDataTableColumns<any>>(() => [
  { type: 'selection', width: 48 },
  {
    title: '封面',
    key: 'cover',
    width: 72,
    render: (row) => renderCover(row),
  },
  { title: '标题', path: 'title', minWidth: 180, ellipsis: { tooltip: true } },
  {
    title: '发布日期',
    key: 'date',
    width: 120,
    render: (row) => row.date || row.published_at || '—',
  },
  {
    title: '状态',
    key: 'status',
    width: 96,
    render: (row) => (
      <NTag size="small" type={row.status === 'published' ? 'success' : 'default'} bordered={false}>
        {row.status === 'published' ? '已发布' : '草稿'}
      </NTag>
    ),
  },
  {
    title: '分类',
    key: 'categories',
    minWidth: 120,
    render: (row) => renderCategories(row),
  },
  {
    title: '标签',
    key: 'tags',
    width: 320,
    ellipsis: { tooltip: true },
    render: (row) => renderTags(row),
  },
  {
    title: '置顶',
    key: 'sticky',
    width: 80,
    render: (row) =>
      row.sticky ? (
        <NTag size="small" type="error" bordered={false}>
          置顶
        </NTag>
      ) : (
        '—'
      ),
  },
  {
    title: '首页',
    key: 'home',
    width: 72,
    render: (row) =>
      row.home || row.home_visible ? (
        <NTag size="small" type="success" bordered={false}>
          是
        </NTag>
      ) : (
        '—'
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    fixed: 'right',
    render: (row) => (
      <NFlex size={12} align="center">
        {hasPermission('admin:blog:post:update') && row.status !== 'published' ? (
          <NButton
            type="success"
            text
            title="发布"
            onClick={() => void setStatus(row, 'published')}
          >
            {renderButtonIcon('icon-park-outline:send')}
          </NButton>
        ) : null}
        {hasPermission('admin:blog:post:update') && row.status === 'published' ? (
          <NButton type="warning" text title="撤回为草稿" onClick={() => void onWithdraw(row)}>
            {renderButtonIcon('icon-park-outline:undo')}
          </NButton>
        ) : null}
        {hasPermission('admin:blog:post:update') ? (
          <NButton type="primary" text title="编辑" onClick={() => openEditPage(row.uid)}>
            {renderButtonIcon('icon-park-outline:edit')}
          </NButton>
        ) : null}
        {hasPermission('admin:blog:post:delete') ? (
          <NButton type="error" text title="删除" onClick={() => void onRemove([row.uid])}>
            {renderButtonIcon('icon-park-outline:delete')}
          </NButton>
        ) : null}
      </NFlex>
    ),
  },
])

/** 从路由 query 同步状态筛选（运营台「草稿箱/已发布」快捷入口） */
function applyStatusFromRoute() {
  const status = String(route.query.status || '').trim()
  if (status !== 'draft' && status !== 'published') return
  state.searchValues = { ...state.searchValues, status }
  searchForm.values.value = { ...searchForm.values.value, status }
}

onMounted(() => {
  applyStatusFromRoute()
  void fetchPage()
})

watch(
  () => route.fullPath,
  () => {
    applyStatusFromRoute()
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
      title="博客文章"
      row-key="uid"
      :scroll-x="1360"
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
            v-if="hasPermission('admin:blog:post:create')"
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
          <NButton
            v-if="hasPermission('admin:blog:post:update')"
            type="success"
            text
            title="批量发布"
            aria-label="批量发布"
            :disabled="!state.checkedKeys.length"
            @click="batchSetStatus('published')"
          >
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:send" />
              </NIcon>
            </template>
          </NButton>
          <NButton
            v-if="hasPermission('admin:blog:post:update')"
            type="warning"
            text
            title="批量撤回"
            aria-label="批量撤回"
            :disabled="!state.checkedKeys.length"
            @click="batchSetStatus('draft')"
          >
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:undo" />
              </NIcon>
            </template>
          </NButton>
          <NButton
            v-if="hasPermission('admin:blog:post:delete')"
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
  </NFlex>
</template>
