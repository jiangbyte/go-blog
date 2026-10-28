<!-- Author: go-blog — 简历头像：仅此处 3:4 裁剪上传，不影响全局圆形头像 -->

<script setup lang="ts">
import CropAvatarUploadModal from '@/components/upload/CropAvatarUploadModal.vue'
import StorageImage from '@/components/upload/StorageImage.vue'
import { NButton, NFlex, NText } from 'naive-ui'
import { computed, ref } from 'vue'

const value = defineModel<string>('value', { default: '' })

defineProps<{
  disabled?: boolean
}>()

const showCrop = ref(false)

const hasAvatar = computed(() => Boolean(String(value.value || '').trim()))

/** 裁剪回传 object_name / url，写入 profile.avatar_url。 */
function onUploaded(v: string) {
  value.value = String(v || '').trim()
}

function clearAvatar() {
  value.value = ''
}
</script>

<template>
  <div class="resume-avatar-upload">
    <NFlex align="center" :size="12" :wrap="false">
      <button
        type="button"
        class="resume-avatar-upload__thumb"
        :disabled="disabled"
        title="裁剪上传头像"
        @click="showCrop = true"
      >
        <StorageImage
          v-if="hasAvatar"
          :object-name="value"
          :src="value"
          alt="简历头像"
          :width="82"
          :height="106"
          object-fit="cover"
        />
        <span v-else class="resume-avatar-upload__placeholder">
          <NText depth="3">82×106</NText>
        </span>
      </button>

      <NFlex vertical :size="8">
        <NButton :disabled="disabled" @click="showCrop = true"> 裁剪上传 </NButton>
        <NButton v-if="hasAvatar" :disabled="disabled" quaternary type="error" @click="clearAvatar">
          清除
        </NButton>
      </NFlex>
    </NFlex>

    <CropAvatarUploadModal
      v-model:show="showCrop"
      :avatar="hasAvatar ? value : null"
      title="裁剪简历头像（3:4）"
      :aspect-ratio="3 / 4"
      :preview-round="false"
      :output-width="246"
      @uploaded="onUploaded"
    />
  </div>
</template>

<style scoped>
.resume-avatar-upload {
  width: 100%;
  max-width: 100%;
  min-width: 0;
}

.resume-avatar-upload__thumb {
  flex: none;
  display: block;
  width: 82px;
  height: 106px;
  padding: 0;
  border: 1px solid var(--border-color, #efeff5);
  border-radius: 2px;
  background: var(--body-color, #f5f5f5);
  overflow: hidden;
  cursor: pointer;
}

.resume-avatar-upload__thumb:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.resume-avatar-upload__thumb :deep(img) {
  width: 82px !important;
  height: 106px !important;
  object-fit: cover;
  display: block;
}

.resume-avatar-upload__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  font-size: 12px;
}
</style>
