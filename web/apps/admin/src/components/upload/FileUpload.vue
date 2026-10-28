<!-- Author: go-blog -->

<script setup lang="ts">
import type { UploadFileInfo } from 'naive-ui'
import { Icon } from '@iconify/vue/offline'
import { fileApi } from '@/api'
import { formatFileSize, normalizeUploadedFile } from '@/utils'
import type { UploadedFileValueType } from '@/utils'
import { computed, reactive, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    value?: string | null
    file?: File | null
    accept?: string
    buttonText?: string
    icon?: string
    mode?: 'button' | 'icon' | 'upload'
    uploadVariant?: 'default' | 'dragger'
    /** mode=upload 时的列表样式；image-card 为封面卡片 */
    listType?: 'text' | 'image' | 'image-card'
    autoUpload?: boolean
    storageProvider?: string | null
    preview?: 'image' | 'video' | 'file'
    compact?: boolean
    valueType?: UploadedFileValueType
    /** 编辑回显时的原始文件名（优先于 object key 解析） */
    displayName?: string | null
  }>(),
  {
    value: '',
    accept: '',
    buttonText: '',
    icon: '',
    mode: 'button',
    uploadVariant: 'dragger',
    listType: 'text',
    autoUpload: true,
    storageProvider: null,
    preview: 'file',
    compact: false,
    valueType: 'auto',
    file: null,
    displayName: '',
  },
)

const emit = defineEmits<{
  'update:value': [value: string]
  'update:file': [file: File | null]
  selected: [file: File]
  cleared: []
  uploaded: [file: any]
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const state = reactive({
  loading: false,
  fileName: '',
  fileUrl: '',
  fileSize: null as number | null,
  contentType: null as string | null,
  selectedFile: null as File | null,
  uploadFileList: [] as UploadFileInfo[],
})

const currentUrl = computed(() => {
  if (state.fileUrl) {
    return state.fileUrl
  }
  const value = String(props.value || '').trim()
  if (!value) {
    return undefined
  }
  if (/^(https?:|data:|blob:)/i.test(value)) {
    return value
  }
  return undefined
})
const currentName = computed(
  () => state.fileName || props.file?.name || props.value || '未选择文件',
)
const uploadText = computed(() => props.buttonText || '上传')
const actionIcon = computed(() => props.icon || 'icon-park-outline:upload')

watch(
  () => [props.value, props.displayName] as const,
  async ([value]) => {
    const raw = String(value || '').trim()
    if (!raw) {
      if (!state.selectedFile) {
        state.fileUrl = ''
        state.fileName = ''
        state.uploadFileList = []
      }
      return
    }
    if (/^(https?:|data:|blob:)/i.test(raw)) {
      state.fileUrl = raw
      syncExistingUploadList(raw, props.displayName?.trim() || raw)
      return
    }
    await resolveExistingFileMeta(raw)
  },
  { immediate: true },
)

async function resolveExistingFileMeta(objectKey: string) {
  const hintedName = props.displayName?.trim()
  const fallbackName = hintedName || displayNameFromObjectKey(objectKey)
  syncExistingUploadList(objectKey, fallbackName)
  try {
    const response = await fileApi.url(objectKey)
    const meta = response.data
    state.fileUrl = meta?.url || ''
    const resolvedName = meta?.original_name?.trim() || fallbackName
    state.fileName = resolvedName
    state.fileSize = meta?.size ?? null
    state.contentType = meta?.content_type ?? null
    syncExistingUploadList(objectKey, resolvedName)
  } catch {
    state.fileUrl = ''
    syncExistingUploadList(objectKey, fallbackName)
  }
}

watch(
  () => props.file,
  (file) => {
    if (file === state.selectedFile) {
      return
    }
    setSelectedFile(file ?? null, false)
  },
  { immediate: true },
)

function triggerUpload() {
  inputRef.value?.click()
}

function clearValue() {
  clearSelectedFile()
  emit('update:value', '')
}

function clearSelectedFile() {
  state.fileName = ''
  state.fileUrl = ''
  state.fileSize = null
  state.contentType = null
  state.selectedFile = null
  state.uploadFileList = []
  emit('update:file', null)
  emit('cleared')
}

async function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) {
    return
  }
  await selectFile(file)
}

async function handleUploadFileListChange(fileList: UploadFileInfo[]) {
  if (!fileList.length) {
    clearSelectedFile()
    emit('update:value', '')
    return
  }
  const fileInfo = [...fileList].reverse().find((item) => item.file)
  if (!fileInfo?.file) {
    state.uploadFileList = fileList
    return
  }
  await selectFile(fileInfo.file, fileInfo)
}

function displayNameFromObjectKey(objectKey: string) {
  const normalized = objectKey.replace(/\\/g, '/')
  const slash = normalized.lastIndexOf('/')
  return slash >= 0 ? normalized.slice(slash + 1) : normalized
}

/** 编辑回显：将已有 object key 同步到 NUpload 文件列表（无本地 File 对象）。 */
function syncExistingUploadList(objectKey: string, displayName: string) {
  if (props.mode !== 'upload' || state.selectedFile) {
    return
  }
  const name = displayName.trim() || displayNameFromObjectKey(objectKey)
  state.fileName = name
  state.uploadFileList = [
    {
      id: `existing-${objectKey}`,
      name,
      status: 'finished',
      percentage: 100,
      // image-card 需要可访问 url 才能缩略图预览
      url: state.fileUrl || undefined,
    },
  ]
}

async function selectFile(file: File, fileInfo?: UploadFileInfo) {
  setSelectedFile(file, true, fileInfo)
  if (props.autoUpload) {
    await uploadSelectedFile()
  }
}

function setSelectedFile(file: File | null, shouldEmit: boolean, fileInfo?: UploadFileInfo) {
  state.selectedFile = file
  state.fileUrl = ''
  state.fileName = file?.name || ''
  state.fileSize = file?.size ?? null
  state.contentType = file?.type || null
  state.uploadFileList = file ? [fileInfo ?? createUploadFileInfo(file)] : []
  if (shouldEmit) {
    emit('update:file', file)
    if (file) {
      emit('selected', file)
    }
  }
}

async function uploadSelectedFile() {
  const file = state.selectedFile
  if (!file) {
    return undefined
  }
  state.loading = true
  updateUploadStatus('uploading')
  try {
    const response = await fileApi.upload(file, {
      storage_provider: props.storageProvider,
    })
    const uploaded = response.data ?? {}
    const normalized = normalizeUploadedFile(uploaded, file, props.valueType)
    state.fileName = normalized.name
    state.fileUrl = normalized.url
    state.fileSize = normalized.size
    state.contentType = normalized.contentType
    state.uploadFileList = state.uploadFileList.map((item) => ({
      ...item,
      name: normalized.name,
      url: normalized.url || item.url,
    }))
    emit('update:value', normalized.value)
    emit('uploaded', { ...normalized, ...uploaded })
    window.$message.success('上传成功')
    updateUploadStatus('finished')
    return uploaded
  } catch (error) {
    updateUploadStatus('error')
    throw error
  } finally {
    state.loading = false
  }
}

function updateUploadStatus(status: UploadFileInfo['status']) {
  state.uploadFileList = state.uploadFileList.map((item) => ({
    ...item,
    status,
    percentage: status === 'finished' ? 100 : item.percentage,
  }))
}

function createUploadFileInfo(file: File): UploadFileInfo {
  return {
    id: `${Date.now()}-${file.name}`,
    name: file.name,
    status: 'pending',
    file,
    type: file.type,
  }
}

defineExpose({
  clear: clearValue,
  upload: uploadSelectedFile,
})
</script>

<template>
  <div
    class="file-upload"
    :class="{
      'file-upload--compact': compact && mode !== 'upload',
      'file-upload--image-row': mode !== 'upload' && preview === 'image',
    }"
  >
    <input
      ref="inputRef"
      class="file-upload__input"
      type="file"
      :accept="accept"
      @change="handleFileChange"
    />

    <div
      v-if="mode !== 'upload' && preview === 'image' && currentUrl"
      class="file-upload__image"
      :class="{ 'file-upload__image--compact': compact }"
    >
      <NImage
        :src="currentUrl"
        object-fit="cover"
        :alt="currentName"
        :width="compact ? 56 : 80"
        :height="compact ? 32 : 45"
      />
    </div>
    <video
      v-else-if="mode !== 'upload' && !compact && preview === 'video' && currentUrl"
      class="file-upload__video"
      controls
      :src="currentUrl"
    />
    <NEllipsis
      v-else-if="mode !== 'upload' && !compact && preview !== 'image'"
      class="file-upload__name"
    >
      {{ currentName }}
    </NEllipsis>
    <div
      v-if="
        mode !== 'upload' &&
        !compact &&
        preview !== 'image' &&
        (state.fileSize !== null || state.contentType)
      "
      class="file-upload__meta"
    >
      <span v-if="state.fileSize !== null">{{ formatFileSize(state.fileSize) }}</span>
      <span v-if="state.contentType">{{ state.contentType }}</span>
    </div>

    <NUpload
      v-if="mode === 'upload'"
      class="file-upload__native"
      :accept="accept"
      :default-upload="false"
      :disabled="state.loading"
      :file-list="state.uploadFileList"
      :list-type="listType"
      :max="1"
      :show-cancel-button="!!props.value"
      :show-retry-button="false"
      @update:file-list="handleUploadFileListChange"
    >
      <template v-if="listType === 'image-card'">
        <NFlex vertical align="center" justify="center" :size="4">
          <NIcon size="24">
            <Icon :icon="actionIcon || 'icon-park-outline:plus'" />
          </NIcon>
          <NText depth="3">
            {{ uploadText || '上传' }}
          </NText>
        </NFlex>
      </template>
      <NUploadDragger v-else-if="uploadVariant === 'dragger'">
        <div class="file-upload__dragger">
          <NIcon size="28">
            <Icon :icon="actionIcon" />
          </NIcon>
          <div>{{ uploadText || '选择文件' }}</div>
        </div>
      </NUploadDragger>
      <NButton v-else :loading="state.loading">
        <template #icon>
          <NIcon>
            <Icon :icon="actionIcon" />
          </NIcon>
        </template>
        {{ uploadText || '选择文件' }}
      </NButton>
    </NUpload>

    <div v-else class="file-upload__actions" :class="{ 'file-upload__actions--compact': compact }">
      <NButton
        v-if="mode === 'icon'"
        text
        :loading="state.loading"
        :title="uploadText"
        :aria-label="uploadText"
        @click="triggerUpload"
      >
        <template #icon>
          <NIcon>
            <Icon :icon="actionIcon" />
          </NIcon>
        </template>
      </NButton>
      <NButton v-else size="small" :loading="state.loading" @click="triggerUpload">
        <template v-if="icon" #icon>
          <NIcon>
            <Icon :icon="icon" />
          </NIcon>
        </template>
        {{ uploadText }}
      </NButton>
      <NButton v-if="value" text type="error" size="small" @click="clearValue"> 清除 </NButton>
    </div>
  </div>
</template>

<style scoped>
.file-upload {
  display: grid;
  gap: 8px;
  width: 100%;
  min-width: 0;
}

.file-upload--image-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.file-upload--compact {
  display: inline-flex;
  align-items: center;
  width: auto;
}

.file-upload__input {
  display: none;
}

.file-upload__native {
  width: 100%;
  min-width: 0;
}

.file-upload__dragger {
  display: grid;
  justify-items: center;
  gap: 6px;
  padding: 6px 0;
  color: var(--text-color-3);
}

.file-upload__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.file-upload__actions--compact {
  gap: 0;
}

.file-upload__image,
.file-upload__video {
  flex-shrink: 0;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
  background: var(--body-color);
  line-height: 0;
}

.file-upload__image--compact {
  border-radius: 4px;
}

.file-upload__video {
  width: 160px;
  max-width: 100%;
  aspect-ratio: 16 / 9;
}

.file-upload__name {
  max-width: 100%;
  color: var(--text-color-3);
}

.file-upload__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  color: var(--text-color-3);
  font-size: 12px;
  line-height: 1.3;
}
</style>
