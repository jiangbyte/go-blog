<!-- Author: go-blog -->

<script setup lang="ts">
import FileUpload from './FileUpload.vue'
import type { UploadedFileValueType } from '@/utils'

withDefaults(
  defineProps<{
    value?: string | null
    valueType?: UploadedFileValueType
    /** 单行紧凑：小缩略图 + 操作按钮横排 */
    compact?: boolean
    /**
     * 展示形态：
     * - default：缩略图 + 上传/清除按钮
     * - card：NUpload image-card（虚线框 + 上传文案）
     */
    variant?: 'default' | 'card'
    buttonText?: string
  }>(),
  {
    value: '',
    valueType: 'object_name',
    compact: false,
    variant: 'default',
    buttonText: '',
  },
)

defineEmits<{
  'update:value': [value: string]
  uploaded: [file: any]
}>()
</script>

<template>
  <FileUpload
    :value="value"
    :value-type="valueType"
    :compact="compact"
    :mode="variant === 'card' ? 'upload' : 'button'"
    :list-type="variant === 'card' ? 'image-card' : 'text'"
    :button-text="buttonText || (variant === 'card' ? '上传封面' : '')"
    :icon="variant === 'card' ? 'icon-park-outline:plus' : ''"
    accept="image/jpeg,image/png,image/webp,image/gif"
    preview="image"
    @update:value="$emit('update:value', $event)"
    @uploaded="$emit('uploaded', $event)"
  />
</template>
