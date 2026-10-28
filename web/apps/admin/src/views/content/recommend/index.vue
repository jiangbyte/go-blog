<!-- Author: go-blog — 推荐位排序管理（加入/移出在文章列表操作列） -->

<script setup lang="tsx">
import type { ProDataTableColumns } from 'pro-naive-ui'
import { Icon } from '@iconify/vue/offline'
import { contentRecommendApi } from '@/api'
import { formatDateTime, hasPermission, renderButtonIcon } from '@/utils'
import StorageImage from '@/components/upload/StorageImage.vue'
import { NButton, NFlex, NIcon, NInputNumber } from 'naive-ui'
import { ProCard, ProDataTable } from 'pro-naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'

const state = reactive({
  records: [] as any[],
  loading: false,
  saving: false,
})

const dirty = ref(false)

async function fetchList() {
  state.loading = true
  try {
    const res = await contentRecommendApi.page({ current: 1, size: 100 })
    state.records = (res.data?.records || []).map((r: any, idx: number) => ({
      ...r,
      sort_code: Number(r.sort_code ?? (idx + 1) * 10) || (idx + 1) * 10,
    }))
    dirty.value = false
  } finally {
    state.loading = false
  }
}

function markDirty() {
  dirty.value = true
}

function moveRow(index: number, delta: number) {
  const next = index + delta
  if (next < 0 || next >= state.records.length) return
  const list = [...state.records]
  const tmp = list[index]
  list[index] = list[next]
  list[next] = tmp
  state.records = list.map((r, i) => ({ ...r, sort_code: (i + 1) * 10 }))
  markDirty()
}

async function onRemove(articleId: string) {
  if (!hasPermission('admin:contentrecommend:delete')) return
  await contentRecommendApi.remove(articleId)
  window.$message.success('已移出推荐位')
  await fetchList()
}

async function onSave() {
  if (!hasPermission('admin:contentrecommend:replace')) return
  state.saving = true
  try {
    // 按当前表格顺序全量 replace
    await contentRecommendApi.replace(
      state.records.map((r, i) => ({
        article_id: r.article_id,
        sort_code: Number(r.sort_code ?? (i + 1) * 10) || (i + 1) * 10,
      })),
    )
    window.$message.success('已保存排序')
    await fetchList()
  } finally {
    state.saving = false
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
    title: '排序',
    key: 'sort_code',
    width: 100,
    render: (row) => (
      <NInputNumber
        value={Number(row.sort_code ?? 0)}
        showButton={false}
        class="w-full"
        onUpdateValue={(v) => {
          row.sort_code = Number(v ?? 0)
          markDirty()
        }}
      />
    ),
  },
  {
    title: '封面',
    key: 'cover',
    width: 72,
    render: (row) => renderCover(row),
  },
  { title: '标题', path: 'title', minWidth: 180, ellipsis: { tooltip: true } },
  { title: '标识', path: 'slug', width: 100 },
  { title: '浏览', path: 'view_count', width: 80 },
  {
    title: '发布时间',
    path: 'published_at',
    width: 170,
    render: (row) => formatDateTime(row.published_at) || '—',
  },
  {
    title: '操作',
    key: 'actions',
    width: 140,
    fixed: 'right',
    render: (row, index) => (
      <NFlex size={8} align="center">
        <NButton text title="上移" disabled={index === 0} onClick={() => moveRow(index, -1)}>
          {renderButtonIcon('icon-park-outline:up')}
        </NButton>
        <NButton
          text
          title="下移"
          disabled={index >= state.records.length - 1}
          onClick={() => moveRow(index, 1)}
        >
          {renderButtonIcon('icon-park-outline:down')}
        </NButton>
        {hasPermission('admin:contentrecommend:delete') ? (
          <NButton type="error" text title="移出" onClick={() => void onRemove(row.article_id)}>
            {renderButtonIcon('icon-park-outline:delete')}
          </NButton>
        ) : null}
      </NFlex>
    ),
  },
])

onMounted(() => {
  void fetchList()
})
</script>

<template>
  <NFlex class="h-full min-h-0" vertical>
    <ProCard>
      <NFlex align="center" justify="space-between" :size="12" wrap>
        <span class="text-13px text-[var(--n-text-color-3)]">
          在「文章管理」操作列将已发布文章加入推荐；此处调整排序（最多 20 篇）
        </span>
        <NFlex align="center" :size="12">
          <NButton
            v-if="hasPermission('admin:contentrecommend:replace')"
            type="primary"
            :loading="state.saving"
            :disabled="!dirty"
            @click="onSave"
          >
            保存排序
          </NButton>
          <NButton quaternary :loading="state.loading" @click="fetchList"> 刷新 </NButton>
        </NFlex>
      </NFlex>
    </ProCard>

    <ProDataTable
      class="min-h-0 flex-1"
      title="推荐位"
      row-key="article_id"
      :scroll-x="900"
      :columns="tableColumns"
      :data="state.records"
      :loading="state.loading"
    >
      <template #toolbar>
        <NButton text title="刷新" aria-label="刷新" :loading="state.loading" @click="fetchList">
          <template #icon>
            <NIcon>
              <Icon icon="icon-park-outline:reload" />
            </NIcon>
          </template>
        </NButton>
      </template>
    </ProDataTable>
  </NFlex>
</template>
