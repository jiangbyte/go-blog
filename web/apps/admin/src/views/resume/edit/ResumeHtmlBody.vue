<!-- Author: go-blog — 简历纸面正文：统一转 HTML 消毒渲染（外观变量可作用到全部分节） -->

<script setup lang="ts">
import DOMPurify from 'dompurify'
import type { Config } from 'dompurify'
import { computed } from 'vue'
import { toResumeHtml } from '@jiangbyte/shared'

const props = withDefaults(
  defineProps<{
    value?: string | null
  }>(),
  {
    value: '',
  },
)

const PURIFY_CONFIG: Config = {
  ALLOWED_TAGS: ['p', 'br', 'strong', 'b', 'em', 'i', 'ul', 'ol', 'li'],
  ALLOWED_ATTR: [],
}

const raw = computed(() => String(props.value ?? '').trim())
const safeHtml = computed(() => {
  if (!raw.value) return ''
  return DOMPurify.sanitize(toResumeHtml(raw.value), PURIFY_CONFIG)
})
</script>

<template>
  <div v-if="safeHtml" class="resume-html-body" v-html="safeHtml" />
</template>

<style scoped>
.resume-html-body {
  line-height: inherit;
  color: inherit;
  font-size: inherit;
  font-family: inherit;
  word-break: break-word;
}

.resume-html-body :deep(p),
.resume-html-body :deep(ul),
.resume-html-body :deep(ol),
.resume-html-body :deep(li) {
  margin: 0 0 0.35em;
  line-height: var(--line-height, 1.55);
  font-size: inherit;
  font-family: inherit;
}

.resume-html-body :deep(ul),
.resume-html-body :deep(ol) {
  padding-left: 1.2em;
}

.resume-html-body :deep(li:last-child),
.resume-html-body :deep(p:last-child) {
  margin-bottom: 0;
}

.resume-html-body :deep(strong),
.resume-html-body :deep(b) {
  font-weight: 700;
}
</style>
