<!-- Author: go-blog — 文件详情 -->

<script setup lang="ts">
import type { SysFileItem } from '@/api/sys/file'
import { fileApi } from '@/api'
import {
  displayValue,
  formatDateTime,
  formatFileSize,
  hasPermission,
} from '@/utils'
import { NButton, NFlex, NSpin, NTag, NText } from 'naive-ui'
import { computed, reactive } from 'vue'

const emit = defineEmits<{
  deleted: []
}>()

const props = defineProps<{
  providerLabel?: (provider: string) => string
}>()

const state = reactive({
  showModal: false,
  loading: false,
  refreshing: false,
  previewLoading: false,
  previewError: false,
  previewUrl: '',
  file: {} as SysFileItem,
})

const isImage = computed(() => String(state.file.content_type || '').startsWith('image/'))
const storageLabel = computed(() => {
  const provider = state.file.storage_provider || ''
  return props.providerLabel?.(provider) || provider || '—'
})

async function resolvePreviewUrl(file: SysFileItem) {
  state.previewError = false
  state.previewUrl = ''
  if (!String(file.content_type || '').startsWith('image/')) return

  const objectName = String(file.object_name || '').trim()
  state.previewLoading = true
  try {
    if (objectName) {
      const res = await fileApi.url(objectName)
      const signed = String(res.data?.url || '').trim()
      if (signed) {
        state.previewUrl = signed
        state.file.url = signed
        return
      }
    }
    const fallback = String(file.url || '').trim()
    state.previewUrl = fallback
    if (!fallback) state.previewError = true
  } catch {
    const fallback = String(file.url || '').trim()
    state.previewUrl = fallback
    if (!fallback) state.previewError = true
  } finally {
    state.previewLoading = false
  }
}

async function openModal(id: string, row?: Partial<SysFileItem>) {
  state.file = (row as SysFileItem) ?? ({} as SysFileItem)
  state.previewUrl = String(state.file.url || '')
  state.previewError = false
  state.showModal = true
  state.loading = true
  try {
    const res = await fileApi.detail({ id })
    state.file = res.data ?? (row as SysFileItem) ?? ({} as SysFileItem)
  } catch {
    // 详情失败仍展示行数据
  } finally {
    state.loading = false
  }
  await resolvePreviewUrl(state.file)
}

async function copyText(text?: string | null, label = '已复制') {
  const value = String(text || '')
  if (!value) {
    window.$message.warning('无可复制内容')
    return
  }
  try {
    await navigator.clipboard.writeText(value)
    window.$message.success(label)
  } catch {
    window.$message.error('复制失败')
  }
}

async function onDownload() {
  if (!state.file.id && !state.file.object_name) return
  try {
    await fileApi.downloadFile(state.file)
    window.$message.success('开始下载')
  } catch {
    window.$message.error('下载失败')
  }
}

function onOpenUrl() {
  const url = String(state.previewUrl || state.file.url || '').trim()
  if (!url) {
    window.$message.warning('暂无访问链接')
    return
  }
  window.open(url, '_blank', 'noopener,noreferrer')
}

async function onRefreshUrl() {
  const objectName = String(state.file.object_name || '').trim()
  if (!objectName) return
  state.refreshing = true
  try {
    await resolvePreviewUrl(state.file)
    if (state.previewUrl) {
      window.$message.success('链接已刷新')
    } else {
      window.$message.error('刷新链接失败')
    }
  } finally {
    state.refreshing = false
  }
}

function onPreviewError() {
  state.previewError = true
}

function onRemove() {
  const id = String(state.file.id || '')
  if (!id) return
  window.$dialog.warning({
    title: '删除文件',
    content: `确定删除「${state.file.original_name || id}」？将同时删除对象存储中的对象。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await fileApi.remove({ ids: [id] })
      window.$message.success('已删除')
      state.showModal = false
      emit('deleted')
    },
  })
}

defineExpose({ openModal })
</script>

<template>
  <DetailContainer
    v-model:show="state.showModal"
    title="文件详情"
    :width="640"
    :mask-closable="false"
  >
    <NSpin :show="state.loading || state.previewLoading">
      <img
        v-if="isImage && state.previewUrl && !state.previewError"
        :src="state.previewUrl"
        :alt="state.file.original_name || ''"
        class="file-detail__img"
        referrerpolicy="no-referrer"
        draggable="false"
        @error="onPreviewError"
      />
      <NText v-else-if="isImage" depth="3" class="mb-3 block">
        {{ state.previewError ? '预览加载失败' : '加载预览中…' }}
      </NText>
      <NText v-else depth="3" class="mb-3 block">
        {{ state.file.content_type || '无预览' }}
      </NText>

      <NFlex class="mb-3" :size="8" wrap>
        <NButton size="small" secondary @click="copyText(state.file.id, '已复制 ID')">
          复制 ID
        </NButton>
        <NButton size="small" secondary @click="copyText(state.file.object_name, '已复制对象键')">
          复制对象键
        </NButton>
        <NButton
          size="small"
          secondary
          @click="copyText(state.previewUrl || state.file.url, '已复制链接')"
        >
          复制链接
        </NButton>
        <NButton size="small" secondary @click="onDownload">下载</NButton>
        <NButton size="small" secondary @click="onOpenUrl">新窗口打开</NButton>
        <NButton size="small" secondary :loading="state.refreshing" @click="onRefreshUrl">
          刷新链接
        </NButton>
        <NButton
          v-if="hasPermission('admin:file:delete')"
          size="small"
          type="error"
          secondary
          @click="onRemove"
        >
          删除
        </NButton>
      </NFlex>

      <NDescriptions label-placement="left" bordered :column="1" size="small">
        <NDescriptionsItem label="文件 ID">
          {{ displayValue(state.file.id) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="原始文件名">
          {{ displayValue(state.file.original_name) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="对象键">
          <span class="break-all">{{ displayValue(state.file.object_name) }}</span>
        </NDescriptionsItem>
        <NDescriptionsItem label="访问链接">
          <span class="break-all">{{ displayValue(state.previewUrl || state.file.url) }}</span>
        </NDescriptionsItem>
        <NDescriptionsItem label="存储引擎">
          <NTag size="small">{{ storageLabel }}</NTag>
        </NDescriptionsItem>
        <NDescriptionsItem label="存储桶">
          {{ displayValue(state.file.bucket) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="内容类型">
          {{ displayValue(state.file.content_type) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="大小">
          {{ formatFileSize(state.file.size) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="创建人">
          {{ displayValue(state.file.created_by) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="创建时间">
          {{ formatDateTime(state.file.created_at) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="更新人">
          {{ displayValue(state.file.updated_by) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="更新时间">
          {{ formatDateTime(state.file.updated_at) }}
        </NDescriptionsItem>
      </NDescriptions>
    </NSpin>
  </DetailContainer>
</template>

<style scoped>
.file-detail__img {
  display: block;
  max-width: 100%;
  max-height: 280px;
  width: auto;
  height: auto;
  margin: 0 0 12px;
  padding: 0;
  border: 0;
  outline: none;
  background: none;
  box-shadow: none;
  vertical-align: top;
}
.break-all {
  word-break: break-all;
}
</style>
