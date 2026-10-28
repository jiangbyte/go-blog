<!-- Author: go-blog — 简历管理列表（卡片缩略预览 + 下载） -->

<script setup lang="ts">
import type { PaginationProps } from 'naive-ui'
import { Icon } from '@iconify/vue/offline'
import { resumeApi } from '@/api'
import { hasPermission } from '@/utils'
import {
  NButton,
  NCard,
  NEmpty,
  NFlex,
  NGrid,
  NGridItem,
  NIcon,
  NPagination,
  NSpin,
  NTag,
  NText,
} from 'naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import ResumeThumb from '../preview/ResumeThumb.vue'
import { downloadResumeDocument } from '../preview/resumePrint'

const router = useRouter()
const canCreate = hasPermission('admin:resume:create')
const canUpdate = hasPermission('admin:resume:update')
const canDelete = hasPermission('admin:resume:delete')
const canCopy = hasPermission('admin:resume:copy')
const canPublish = hasPermission('admin:resume:publish')
const canUnpublish = hasPermission('admin:resume:unpublish')
const downloadingId = ref('')

const state = reactive({
  records: [] as any[],
  total: 0,
  loading: false,
  page: 1,
  pageSize: 12,
})

const statusMap: Record<string, { label: string; type: 'default' | 'success' | 'warning' }> = {
  unpublished: { label: '未发布', type: 'default' },
  published: { label: '已发布', type: 'success' },
}

/** 列表卡片相对页面背景加深描边，避免白底糊在一起 */
const resumeCardTheme = {
  borderColor: 'rgba(15, 23, 42, 0.16)',
}

const pagination = computed<PaginationProps>(() => ({
  page: state.page,
  pageSize: state.pageSize,
  itemCount: state.total,
  showSizePicker: true,
  pageSizes: [8, 12, 20, 30],
  prefix: ({ itemCount }) => `${itemCount} 条`,
}))

/** 拉取分页并渲染缩略纸面。 */
async function fetchPage() {
  state.loading = true
  try {
    const response = await resumeApi.page({ current: state.page, size: state.pageSize })
    state.records = response.data?.records ?? []
    state.total = Number(response.data?.total ?? 0)
  } finally {
    state.loading = false
  }
}

function onPageChange(page: number) {
  state.page = page
  void fetchPage()
}

function onPageSizeChange(size: number) {
  state.pageSize = size
  state.page = 1
  void fetchPage()
}

async function onCreate() {
  const res = await resumeApi.create({
    title: '未命名简历',
    template_id: 'classic',
    theme_json: {
      themeColor: 'blue',
      layout: 'left',
      infoDisplay: 'text',
      language: 'zh',
      fontFamily: 'default',
      fontSizeScale: 1,
      spacing: { moduleGap: 12, lineHeight: 1.55, padX: 48, padTop: 48, padBottom: 60 },
    },
    profile_json: { name: '', tagline: '', avatar_url: '', github_url: '', fields: [] },
  })
  window.$message.success('已创建')
  void router.push({ path: '/resume/edit', query: { id: res.data.id } })
}

function goEdit(id: string) {
  void router.push({ path: '/resume/edit', query: { id } })
}

async function onPublish(row: any) {
  await resumeApi.publish(row.id)
  window.$message.success('已发布（其它已发布将自动取消）')
  await fetchPage()
}

async function onUnpublish(row: any) {
  await resumeApi.unpublish(row.id)
  window.$message.success('已取消发布')
  await fetchPage()
}

async function onCopy(row: any) {
  if (!canCopy || !row?.id) return
  await resumeApi.copy(row.id)
  window.$message.success('已复制')
  await fetchPage()
}

async function onRemove(row: any) {
  if (row.status === 'published') {
    window.$message.warning('请先取消发布')
    return
  }
  await resumeApi.remove([row.id])
  window.$message.success('已删除')
  await fetchPage()
}

function goPreview(id: string) {
  void router.push({ path: '/resume/preview', query: { id } })
}

/** 下载与纸面一致的 PDF 文件。 */
async function onDownload(row: any) {
  if (!row?.id || downloadingId.value) return
  downloadingId.value = row.id
  try {
    await downloadResumeDocument(row.id, row.profile_json?.name || row.title || '简历')
    window.$message.success('简历已下载')
  } catch (err: any) {
    window.$message.error(err?.message || '下载失败')
  } finally {
    downloadingId.value = ''
  }
}

function onCardClick(row: any) {
  if (canUpdate) {
    goEdit(row.id)
    return
  }
  goPreview(row.id)
}

onMounted(() => {
  void fetchPage()
})
</script>

<template>
  <NFlex class="h-full min-h-0" vertical :size="16">
    <NFlex justify="space-between" align="center">
      <NText strong style="font-size: 18px"> 简历管理 </NText>
      <NFlex :size="8">
        <NButton v-if="canCreate" type="primary" @click="onCreate">
          <template #icon>
            <NIcon>
              <Icon icon="icon-park-outline:plus" />
            </NIcon>
          </template>
          新建简历
        </NButton>
        <NButton :loading="state.loading" @click="fetchPage">
          <template #icon>
            <NIcon>
              <Icon icon="icon-park-outline:reload" />
            </NIcon>
          </template>
          刷新
        </NButton>
      </NFlex>
    </NFlex>

    <NSpin class="min-h-0 flex-1" :show="state.loading">
      <NGrid
        v-if="state.records.length"
        cols="1 s:2 m:3 l:4 xl:5"
        responsive="screen"
        :x-gap="12"
        :y-gap="12"
      >
        <NGridItem v-for="row in state.records" :key="row.id">
          <NCard
            hoverable
            bordered
            size="small"
            segmented
            content-style="padding: 0"
            :theme-overrides="resumeCardTheme"
          >
            <template #cover>
              <NFlex
                justify="center"
                style="padding: 12px; cursor: pointer; background: var(--hover-color)"
                :title="canUpdate ? '编辑' : '预览'"
                role="button"
                tabindex="0"
                @click="onCardClick(row)"
                @keydown.enter="onCardClick(row)"
              >
                <!-- 内层卡片框住纸面，与外层 NCard 底色区分 -->
                <NCard
                  size="small"
                  :bordered="true"
                  content-style="padding: 0; line-height: 0"
                  style="width: fit-content"
                >
                  <ResumeThumb :doc="row" :scale="0.28" />
                </NCard>
              </NFlex>
            </template>
            <template #header>
              <NText strong class="truncate" :title="row.title">
                {{ row.title || '未命名' }}
              </NText>
            </template>
            <template #header-extra>
              <NTag
                size="small"
                :bordered="false"
                :type="(statusMap[row.status] || statusMap.unpublished).type"
              >
                {{ (statusMap[row.status] || statusMap.unpublished).label }}
              </NTag>
            </template>
            <template #action>
              <NFlex :size="4" style="flex-wrap: wrap">
                <NButton
                  v-if="canUpdate"
                  size="tiny"
                  quaternary
                  type="primary"
                  title="编辑"
                  @click="goEdit(row.id)"
                >
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:edit" />
                    </NIcon>
                  </template>
                </NButton>
                <NButton size="tiny" quaternary title="预览" @click="goPreview(row.id)">
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:preview-open" />
                    </NIcon>
                  </template>
                </NButton>
                <NButton
                  size="tiny"
                  quaternary
                  title="下载"
                  :loading="downloadingId === row.id"
                  @click="onDownload(row)"
                >
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:download" />
                    </NIcon>
                  </template>
                </NButton>
                <NButton v-if="canCopy" size="tiny" quaternary title="复制" @click="onCopy(row)">
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:copy" />
                    </NIcon>
                  </template>
                </NButton>
                <NButton
                  v-if="canPublish && row.status !== 'published'"
                  size="tiny"
                  quaternary
                  type="success"
                  title="发布"
                  @click="onPublish(row)"
                >
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:send" />
                    </NIcon>
                  </template>
                </NButton>
                <NButton
                  v-if="canUnpublish && row.status === 'published'"
                  size="tiny"
                  quaternary
                  type="warning"
                  title="取消发布"
                  @click="onUnpublish(row)"
                >
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:undo" />
                    </NIcon>
                  </template>
                </NButton>
                <NButton
                  v-if="canDelete"
                  size="tiny"
                  quaternary
                  type="error"
                  title="删除"
                  @click="onRemove(row)"
                >
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:delete" />
                    </NIcon>
                  </template>
                </NButton>
              </NFlex>
            </template>
          </NCard>
        </NGridItem>
      </NGrid>
      <NEmpty v-else-if="!state.loading" description="暂无简历，点击新建开始" class="py-24" />
    </NSpin>

    <NFlex v-if="state.total > 0" justify="flex-end">
      <NPagination
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :item-count="pagination.itemCount"
        :page-sizes="pagination.pageSizes"
        show-size-picker
        :prefix="pagination.prefix"
        @update:page="onPageChange"
        @update:page-size="onPageSizeChange"
      />
    </NFlex>
  </NFlex>
</template>
