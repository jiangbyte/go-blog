<!-- Author: go-blog — 存储对象封面：按 object_name 重新签发，避免预签名过期裂图 -->

<script setup lang="ts">
import { fileApi } from '@/api'
import { NImage, NSpin } from 'naive-ui'
import { ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 稳定对象键（优先，每次展示重新签发） */
    objectName?: string | null
    /** 可能过期的预签名 / 外链兜底 */
    src?: string | null
    alt?: string
    width?: number | string
    height?: number | string
    objectFit?: 'fill' | 'contain' | 'cover' | 'none' | 'scale-down'
  }>(),
  {
    alt: '',
    width: 48,
    height: 32,
    objectFit: 'cover',
  },
)

const MEDIA_PREFIXES = ['uploads/', 'avatars/', 'files/', 'resume/', 'content/']

const previewSrc = ref('')
const loading = ref(false)

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

/** 从 object_name 或过期预签名 URL 取出可签发的对象键。 */
function pickObjectKey(objectName?: string | null, src?: string | null): string {
  const key = String(objectName || '').trim()
  if (key && !/^(https?:|data:|blob:)/i.test(key)) {
    return key.replace(/^\/+/, '')
  }

  const raw = String(src || '').trim()
  if (!raw) return ''
  if (!/^(https?:)/i.test(raw)) {
    return raw.replace(/^\/+/, '')
  }

  try {
    const path = new URL(raw).pathname.replace(/^\/+/, '')
    const candidates = [path]
    if (path.includes('/')) {
      const [, ...rest] = path.split('/')
      candidates.push(rest.join('/'))
    }
    const preferred = candidates.find((c) => MEDIA_PREFIXES.some((p) => c.startsWith(p)))
    return preferred || candidates[0] || ''
  } catch {
    return ''
  }
}

/**
 * 列表展示每次重新请求可访问 URL：
 * 1. 有 object_name（或能从过期链还原）→ POST /file/url 签发
 * 2. 稳定外链（非预签名）直接用
 */
async function loadPreview() {
  loading.value = true
  previewSrc.value = ''
  try {
    const objectKey = pickObjectKey(props.objectName, props.src)
    if (objectKey) {
      const res = await fileApi.url(objectKey)
      previewSrc.value = String(res.data?.url || '').trim()
      return
    }
    const direct = String(props.src || '').trim()
    if (/^https?:\/\//i.test(direct) && !looksLikePresignedUrl(direct)) {
      previewSrc.value = direct
    }
  } catch {
    previewSrc.value = ''
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.objectName, props.src] as const,
  () => {
    void loadPreview()
  },
  { immediate: true },
)
</script>

<template>
  <NSpin :show="loading" size="small">
    <NImage
      v-if="previewSrc"
      :src="previewSrc"
      :alt="alt"
      :width="width"
      :height="height"
      :object-fit="objectFit"
    />
    <span v-else>—</span>
  </NSpin>
</template>
