<!-- Author: go-blog — 项目图片合集（以 value 为唯一数据源；禁止临时项污染） -->

<script setup lang="ts">
import type { UploadCustomRequestOptions, UploadFileInfo } from 'naive-ui'
import { fileApi } from '@/api'
import { normalizeUploadedFile } from '@/utils'
import { computed, nextTick, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    value?: string[] | null
    max?: number
    disabled?: boolean
  }>(),
  {
    value: () => [],
    max: 12,
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:value': [value: string[]]
}>()

const fileList = ref<UploadFileInfo[]>([])
/** object_name → 预览 URL，避免每次回显重复签发 */
const urlCache = ref<Record<string, string>>({})
/** 丢弃过期的异步回显结果 */
let syncSeq = 0
/** 正在上传的本地卡片（临时 id），与正式 value 隔离 */
const pendingList = ref<UploadFileInfo[]>([])

const objectNames = computed(() => dedupeNames(props.value))

function dedupeNames(raw: string[] | null | undefined): string[] {
  const out: string[] = []
  for (const item of Array.isArray(raw) ? raw : []) {
    const name = String(item || '').trim()
    if (name && name.includes('/') && !out.includes(name)) {
      out.push(name)
    }
  }
  return out
}

function displayNameFromKey(objectKey: string) {
  const normalized = objectKey.replace(/\\/g, '/')
  const slash = normalized.lastIndexOf('/')
  return slash >= 0 ? normalized.slice(slash + 1) : normalized
}

async function resolvePreviewUrl(objectName: string): Promise<string> {
  if (urlCache.value[objectName]) {
    return urlCache.value[objectName]
  }
  try {
    const res = await fileApi.url(objectName)
    const url = String(res.data?.url || '').trim()
    if (url) {
      urlCache.value = { ...urlCache.value, [objectName]: url }
    }
    return url
  } catch {
    return ''
  }
}

function toFinishedItem(objectName: string, url?: string): UploadFileInfo {
  return {
    id: objectName,
    name: displayNameFromKey(objectName),
    status: 'finished',
    percentage: 100,
    url: url || urlCache.value[objectName] || undefined,
  }
}

/** 合并：正式图（来自 value）+ 上传中临时卡片 */
function applyFileList(finished: UploadFileInfo[]) {
  fileList.value = [...finished, ...pendingList.value]
}

function finishedFromValue(): UploadFileInfo[] {
  return objectNames.value.map((objectName) => {
    const hit = fileList.value.find((f) => f.status === 'finished' && String(f.id) === objectName)
    return toFinishedItem(objectName, hit?.url)
  })
}

/**
 * NUpload 回写列表时：只吸收「上传中」临时项；
 * finished 一律按 value 重建，杜绝临时 id / 重复卡片。
 */
function onUpdateFileList(list: UploadFileInfo[]) {
  pendingList.value = list.filter(
    (f) =>
      (f.status === 'pending' || f.status === 'uploading') && !String(f.id || '').includes('/'),
  )
  applyFileList(finishedFromValue())
}

/** 按 value 重建正式列表；并发回显用序号丢弃过期结果。 */
async function syncFromValue(names: string[]) {
  const seq = ++syncSeq
  const finished: UploadFileInfo[] = []
  for (const objectName of names) {
    const cached = fileList.value.find(
      (f) => f.status === 'finished' && String(f.id) === objectName && f.url,
    )
    if (cached?.url) {
      finished.push(toFinishedItem(objectName, cached.url))
      continue
    }
    const url = await resolvePreviewUrl(objectName)
    if (seq !== syncSeq) return
    finished.push(toFinishedItem(objectName, url || undefined))
  }
  if (seq !== syncSeq) return
  applyFileList(finished)
}

watch(
  objectNames,
  (names) => {
    void syncFromValue(names)
  },
  { immediate: true },
)

function emitNames(names: string[]) {
  emit('update:value', dedupeNames(names).slice(0, props.max))
}

/** 上传成功只追加 object_name；临时卡片由 onUpdateFileList / sync 清掉。 */
async function customRequest({ file, onFinish, onError }: UploadCustomRequestOptions) {
  const raw = file.file
  if (!(raw instanceof File)) {
    onError()
    return
  }
  try {
    const response = await fileApi.upload(raw)
    const uploaded = response.data ?? {}
    const normalized = normalizeUploadedFile(uploaded, raw, 'object_name')
    const objectName = String(normalized.value || normalized.objectName || '').trim()
    if (!objectName || !objectName.includes('/')) {
      onError()
      window.$message.error('上传结果缺少对象键，请重试')
      return
    }
    if (normalized.url) {
      urlCache.value = { ...urlCache.value, [objectName]: normalized.url }
    }
    const next = objectNames.value.includes(objectName)
      ? [...objectNames.value]
      : [...objectNames.value, objectName]
    if (next.length > props.max) {
      onError()
      window.$message.warning(`最多上传 ${props.max} 张图片`)
      return
    }
    // 1. 先写入 value 并等 DOM 同步
    emitNames(next)
    await nextTick()
    // 2. 清掉对应临时上传项，再通知 NUpload 结束
    pendingList.value = pendingList.value.filter((f) => String(f.id) !== String(file.id))
    applyFileList(finishedFromValue())
    onFinish()
    window.$message.success('上传成功')
  } catch {
    onError()
  }
}

/**
 * 删除拦截：只从 value 去掉当前 object_name，由 watch 重建列表。
 * 返回 false 阻止 NUpload 自己改 fileList，避免同 key 误删后续项。
 */
function onRemove(options: { file: UploadFileInfo }) {
  const id = String(options.file.id || '').trim()
  if (!id.includes('/')) {
    pendingList.value = pendingList.value.filter((f) => String(f.id) !== id)
    applyFileList(finishedFromValue())
    return false
  }
  emitNames(objectNames.value.filter((n) => n !== id))
  return false
}
</script>

<template>
  <div class="project-image-gallery">
    <NUpload
      class="project-image-gallery__upload"
      :file-list="fileList"
      multiple
      list-type="image-card"
      accept="image/jpeg,image/png,image/webp,image/gif"
      :max="max"
      :disabled="disabled"
      :custom-request="customRequest"
      :on-remove="onRemove"
      @update:file-list="onUpdateFileList"
    />
    <NText depth="3" style="font-size: 12px; display: block; margin-top: 6px">
      最多 {{ max }} 张；第一张将作为封面展示在列表与详情
    </NText>
  </div>
</template>

<style scoped>
/* 横向流式排列：占满表单宽度，到边界自动换行 */
.project-image-gallery {
  display: block;
  width: 100%;
  min-width: 0;
  max-width: 100%;
}

.project-image-gallery__upload {
  display: block;
  width: 100%;
  min-width: 0;
  max-width: 100%;
}

.project-image-gallery__upload :deep(.n-upload) {
  width: 100%;
}

.project-image-gallery__upload :deep(.n-upload-file-list--grid) {
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  grid-template-columns: unset;
  gap: 8px;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

.project-image-gallery__upload :deep(.n-upload-file-list--grid > *) {
  flex: 0 0 96px;
  width: 96px;
  max-width: 96px;
  box-sizing: border-box;
}
</style>
