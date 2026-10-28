<!-- Author: go-blog — 裁剪后上传到文件存储，回传可访问 URL（供站点配置等复用） -->

<script setup lang="ts">
import { fileApi } from '@/api'
import { computed, reactive, ref, watch } from 'vue'
import { Cropper } from 'vue-advanced-cropper'
import 'vue-advanced-cropper/dist/style.css'

const props = withDefaults(
  defineProps<{
    show: boolean
    avatar?: string | null
    /** 裁剪框宽高比；默认 1（圆形头像），简历证件照可用 3/4 */
    aspectRatio?: number
    /** 弹窗标题 */
    title?: string
    /** 预览是否圆形；默认 true，证件照传 false */
    previewRound?: boolean
    /** 导出画布宽；高度按 aspectRatio 推算 */
    outputWidth?: number
  }>(),
  {
    avatar: null,
    aspectRatio: 1,
    title: '上传头像',
    previewRound: true,
    outputWidth: 320,
  },
)

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'uploaded', url: string): void
}>()

const fileInputRef = ref<HTMLInputElement | null>(null)
const cropperRef = ref<any>(null)

const state = reactive({
  source: '',
  previewUrl: '',
  fileName: '',
  uploading: false,
})

const modalShow = computed({
  get: () => props.show,
  set: (value: boolean) => emit('update:show', value),
})

const safeAspect = computed(() => {
  const r = Number(props.aspectRatio)
  return Number.isFinite(r) && r > 0 ? r : 1
})

const canvasSize = computed(() => {
  const w = Math.max(64, Math.floor(Number(props.outputWidth) || 320))
  const h = Math.max(64, Math.round(w / safeAspect.value))
  return { width: w, height: h }
})

const previewBoxStyle = computed(() => {
  if (props.previewRound) {
    return { width: '128px', height: '128px', borderRadius: '999px' }
  }
  const w = 128
  const h = Math.round(w / safeAspect.value)
  return { width: `${w}px`, height: `${h}px`, borderRadius: '4px' }
})

watch(
  () => props.show,
  (show) => {
    if (!show) {
      resetSource()
    }
  },
)

function openFilePicker() {
  fileInputRef.value?.click()
}

function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) {
    return
  }
  if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
    window.$message.warning('仅支持 JPG、PNG 和 WebP 图片，裁剪后上传')
    return
  }
  resetSource()
  state.fileName = file.name
  state.source = URL.createObjectURL(file)
  updatePreview()
}

/** 裁剪画布导出 PNG，上传文件存储后回传 url。 */
async function uploadCropped() {
  // 1. 从裁剪器取画布；未选图则提示
  const canvas = cropperRef.value?.getResult?.().canvas
  if (!canvas) {
    window.$message.warning('请先选择头像图片')
    return
  }

  // 2. 导出 blob 并上传到通用文件接口
  state.uploading = true
  try {
    const blob = await canvasToBlob(canvas)
    const file = new File([blob], 'avatar-cropped.png', { type: 'image/png' })
    const res = await fileApi.upload(file)
    // 优先落 object_name，避免把会过期的预签名 URL 写入配置
    const objectName = String(res.data?.object_name || '').trim()
    const url = String(res.data?.url || '').trim()
    const value = objectName || url
    if (!value) {
      window.$message.error('上传成功但未返回访问地址')
      return
    }
    // 3. 回传稳定引用并关闭弹窗
    emit('uploaded', value)
    window.$message.success('头像已上传')
    modalShow.value = false
  } finally {
    state.uploading = false
  }
}

function onCropperChange() {
  updatePreview()
}

function updatePreview() {
  window.requestAnimationFrame(() => {
    const canvas = cropperRef.value?.getResult?.().canvas
    state.previewUrl = canvas ? canvas.toDataURL('image/png') : ''
  })
}

function canvasToBlob(canvas: HTMLCanvasElement) {
  return new Promise<Blob>((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (blob) {
          resolve(blob)
          return
        }
        reject(new Error('头像图片导出失败'))
      },
      'image/png',
      0.92,
    )
  })
}

function resetSource() {
  if (state.source) {
    URL.revokeObjectURL(state.source)
  }
  state.source = ''
  state.previewUrl = ''
  state.fileName = ''
}
</script>

<template>
  <NModal
    v-model:show="modalShow"
    preset="card"
    :title="title"
    class="max-w-150"
    :bordered="false"
    :mask-closable="!state.uploading"
  >
    <input
      ref="fileInputRef"
      class="hidden"
      type="file"
      accept="image/jpeg,image/png,image/webp"
      @change="onFileChange"
    />

    <div class="avatar-upload-modal">
      <div v-if="state.source" class="avatar-editor">
        <div class="avatar-cropper-wrap">
          <Cropper
            ref="cropperRef"
            class="avatar-cropper"
            :src="state.source"
            :stencil-props="{ aspectRatio: safeAspect }"
            :canvas="canvasSize"
            :cross-origin="true"
            @change="onCropperChange"
          />
        </div>
        <div class="avatar-preview-panel">
          <div class="avatar-preview" :style="previewBoxStyle">
            <img v-if="state.previewUrl" :src="state.previewUrl" alt="" />
          </div>
          <div class="mt-3 text-sm text-[var(--text-color-3)]">实时预览</div>
        </div>
      </div>
      <div v-else class="avatar-empty">
        <template v-if="previewRound">
          <NAvatar v-if="avatar" round :size="96" :src="avatar" />
          <NAvatar v-else round :size="96">
            <NovaIcon icon="icon-park-outline:user" :size="40" />
          </NAvatar>
        </template>
        <div v-else class="avatar-empty__rect" :style="previewBoxStyle">
          <img v-if="avatar" :src="avatar" alt="" />
          <NovaIcon v-else icon="icon-park-outline:user" :size="40" />
        </div>
        <NButton secondary type="primary" class="mt-4" @click="openFilePicker">
          <template #icon>
            <NovaIcon icon="icon-park-outline:upload-picture" />
          </template>
          选择图片
        </NButton>
      </div>
    </div>

    <template #footer>
      <NSpace justify="space-between" align="center">
        <div class="max-w-70 truncate text-sm text-[var(--text-color-3)]">
          {{ state.fileName || '仅支持 JPG、PNG 和 WebP 图片，裁剪后上传' }}
        </div>
        <NSpace>
          <NButton @click="modalShow = false"> 取消 </NButton>
          <NButton v-if="state.source" secondary @click="openFilePicker"> 重新选择 </NButton>
          <NButton type="primary" :loading="state.uploading" @click="uploadCropped">
            裁剪并上传
          </NButton>
        </NSpace>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped>
.avatar-upload-modal {
  min-height: 360px;
}

.avatar-editor {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 180px;
  gap: 16px;
  align-items: stretch;
}

.avatar-cropper-wrap {
  height: min(56vh, 420px);
  overflow: hidden;
  border: 1px solid var(--border-color);
}

.avatar-cropper {
  height: 100%;
  width: 100%;
  background: var(--body-color);
}

.avatar-preview-panel {
  min-height: 360px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-color);
  padding: 16px;
}

.avatar-preview {
  overflow: hidden;
  background: var(--body-color);
  box-shadow: inset 0 0 0 1px var(--border-color);
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.avatar-empty {
  min-height: 360px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 1px dashed var(--border-color);
}

.avatar-empty__rect {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: var(--body-color);
  box-shadow: inset 0 0 0 1px var(--border-color);
  color: var(--text-color-3);
}

.avatar-empty__rect img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

@media (max-width: 640px) {
  .avatar-editor {
    grid-template-columns: 1fr;
  }

  .avatar-preview-panel {
    min-height: 190px;
  }
}
</style>
