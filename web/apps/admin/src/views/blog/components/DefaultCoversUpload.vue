<!-- Author: go-blog — 站点 default_covers（本地上传 + 第三方 URL） -->

<script setup lang="ts">
import type { UploadCustomRequestOptions, UploadFileInfo } from 'naive-ui'
import { fileApi } from '@/api'
import { normalizeUploadedFile } from '@/utils'
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    value?: string[] | null
    max?: number
    disabled?: boolean
  }>(),
  {
    value: () => [],
    max: 16,
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:value': [value: string[]]
}>()

const fileList = ref<UploadFileInfo[]>([])
const pendingList = ref<UploadFileInfo[]>([])
const urlDraft = ref('')

const covers = computed(() => normalizeList(props.value))
const atLimit = computed(() => covers.value.length >= props.max)

function normalizeList(raw: string[] | null | undefined): string[] {
  const out: string[] = []
  for (const item of Array.isArray(raw) ? raw : []) {
    const url = String(item || '').trim()
    if (url && !out.includes(url)) out.push(url)
  }
  return out
}

function looksLikePresignedUrl(value: string) {
  const lower = value.toLowerCase()
  return (
    lower.includes('x-amz-') ||
    lower.includes('x-oss-') ||
    lower.includes('signature=') ||
    lower.includes('token=') ||
    lower.includes('x-qiniu-')
  )
}

function pickObjectKey(ref: string): string {
  const raw = ref.trim()
  if (!raw) return ''
  if (!/^(https?:)/i.test(raw)) {
    return raw.replace(/^\/+/, '')
  }
  try {
    return new URL(raw).pathname.replace(/^\/+/, '')
  } catch {
    return ''
  }
}

function isStableExternalUrl(raw: string): boolean {
  const url = raw.trim()
  if (!url || url.startsWith('/')) return !!url
  try {
    const parsed = new URL(url)
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') return false
    return !looksLikePresignedUrl(url)
  } catch {
    return false
  }
}

function isCoverRef(raw: string): boolean {
  const value = raw.trim()
  if (!value) return false
  if (isStableExternalUrl(value) || value.startsWith('/')) return true
  // 对象键（uploads/...）或可剥键的预签名链
  return !!pickObjectKey(value)
}

async function resolvePreview(ref: string): Promise<string | undefined> {
  const value = ref.trim()
  if (!value) return undefined
  if (value.startsWith('/') || (isStableExternalUrl(value) && !looksLikePresignedUrl(value))) {
    return value
  }
  const key = pickObjectKey(value)
  if (!key) return undefined
  try {
    const res = await fileApi.url(key)
    return String(res.data?.url || '').trim() || undefined
  } catch {
    return undefined
  }
}

function toFinishedItem(ref: string, preview?: string): UploadFileInfo {
  // 无扩展名的第三方 URL（如 https://t.alcy.cc/fj?u=d00）会被 NUpload
  // 判成非图片而显示文件图标；显式 type + thumbnailUrl 强制缩略图。
  const baseName = ref.replace(/\\/g, '/').split('/').pop()?.split(/[?#]/)[0] || 'cover'
  const name = /\.(webp|svg|png|gif|jpe?g|jfif|bmp|ico)$/i.test(baseName)
    ? baseName
    : `${baseName || 'cover'}.jpg`
  return {
    id: ref,
    name,
    type: 'image/jpeg',
    status: 'finished',
    percentage: 100,
    url: preview,
    thumbnailUrl: preview,
  }
}

function applyFileList(finished: UploadFileInfo[]) {
  fileList.value = [...finished, ...pendingList.value]
}

watch(
  covers,
  (list) => {
    void (async () => {
      const items = await Promise.all(
        list.map(async (ref) => toFinishedItem(ref, await resolvePreview(ref))),
      )
      applyFileList(items)
    })()
  },
  { immediate: true },
)

function emitCovers(next: string[]) {
  emit('update:value', normalizeList(next).slice(0, props.max))
}

function appendCover(url: string): boolean {
  const next = url.trim()
  if (!isCoverRef(next)) {
    window.$message.warning('请输入有效的图片 URL（http/https 或站内路径）')
    return false
  }
  // 外链原样；预签名链剥成 object_name
  const stored =
    isStableExternalUrl(next) || next.startsWith('/')
      ? next
      : pickObjectKey(next) || next
  if (covers.value.includes(stored)) {
    window.$message.warning('该封面已在列表中')
    return false
  }
  if (atLimit.value) {
    window.$message.warning(`最多 ${props.max} 张默认封面`)
    return false
  }
  emitCovers([...covers.value, stored])
  return true
}

function addExternalUrl() {
  if (props.disabled || atLimit.value) return
  if (!appendCover(urlDraft.value)) return
  urlDraft.value = ''
  window.$message.success('已添加')
}

function onUpdateFileList(list: UploadFileInfo[]) {
  pendingList.value = list.filter((f) => f.status === 'pending' || f.status === 'uploading')
  void (async () => {
    const items = await Promise.all(
      covers.value.map(async (ref) => toFinishedItem(ref, await resolvePreview(ref))),
    )
    applyFileList(items)
  })()
}

async function customRequest({ file, onFinish, onError }: UploadCustomRequestOptions) {
  const raw = file.file
  if (!(raw instanceof File)) {
    onError()
    return
  }
  try {
    const response = await fileApi.upload(raw)
    const normalized = normalizeUploadedFile(response.data ?? {}, raw, 'object_name')
    const objectName = String(normalized.value || normalized.object_name || '').trim()
    if (!objectName) {
      onError()
      window.$message.error('上传结果缺少对象键，请重试')
      return
    }
    if (covers.value.length >= props.max) {
      onError()
      window.$message.warning(`最多 ${props.max} 张默认封面`)
      return
    }
    const next = covers.value.includes(objectName)
      ? [...covers.value]
      : [...covers.value, objectName]
    emitCovers(next)
    pendingList.value = pendingList.value.filter((f) => String(f.id) !== String(file.id))
    const items = await Promise.all(
      normalizeList(next).map(async (ref) => toFinishedItem(ref, await resolvePreview(ref))),
    )
    applyFileList(items)
    onFinish()
    window.$message.success('上传成功')
  } catch {
    onError()
  }
}

function onRemove(options: { file: UploadFileInfo }) {
  const id = String(options.file.id || '').trim()
  if (options.file.status === 'pending' || options.file.status === 'uploading') {
    pendingList.value = pendingList.value.filter((f) => String(f.id) !== id)
    void (async () => {
      const items = await Promise.all(
        covers.value.map(async (ref) => toFinishedItem(ref, await resolvePreview(ref))),
      )
      applyFileList(items)
    })()
    return false
  }
  emitCovers(covers.value.filter((url) => url !== id))
  return false
}
</script>

<template>
  <div class="default-covers-upload">
    <NUpload
      class="default-covers-upload__upload"
      :file-list="fileList"
      multiple
      list-type="image-card"
      accept="image/jpeg,image/png,image/webp,image/gif"
      :max="max"
      :disabled="disabled || atLimit"
      :should-use-thumbnail-url="() => true"
      :custom-request="customRequest"
      :on-remove="onRemove"
      @update:file-list="onUpdateFileList"
    />

    <NInputGroup class="default-covers-upload__url mt-2">
      <NInput
        v-model:value="urlDraft"
        :disabled="disabled || atLimit"
        clearable
        placeholder="第三方图片 URL，如 https://..."
        @keyup.enter="addExternalUrl"
      />
      <NButton
        type="primary"
        ghost
        :disabled="disabled || atLimit || !urlDraft.trim()"
        @click="addExternalUrl"
      >
        添加 URL
      </NButton>
    </NInputGroup>

    <NText depth="3" style="font-size: 12px; display: block; margin-top: 6px">
      支持本地上传或填写第三方 URL；文章/项目未设封面时按序轮询；最多 {{ max }} 张
    </NText>
  </div>
</template>

<style scoped>
.default-covers-upload {
  display: block;
  width: 100%;
  min-width: 0;
  max-width: 100%;
}

.default-covers-upload__upload {
  display: block;
  width: 100%;
  min-width: 0;
  max-width: 100%;
}

.default-covers-upload__url {
  width: 100%;
  max-width: 640px;
}

.default-covers-upload__upload :deep(.n-upload) {
  width: 100%;
}

.default-covers-upload__upload :deep(.n-upload-file-list--grid) {
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  grid-template-columns: unset;
  gap: 8px;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

.default-covers-upload__upload :deep(.n-upload-file-list--grid > *) {
  flex: 0 0 96px;
  width: 96px;
  max-width: 96px;
  box-sizing: border-box;
}
</style>
