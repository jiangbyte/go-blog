<!-- Author: go-blog — 轻量 HTML 预览（列表缩略用；分页以 PDF 预览为准） -->

<script setup lang="ts">
import { resumeApi } from '@/api'
import { NEmpty, NSpin } from 'naive-ui'
import { onBeforeUnmount, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 纸面文档（与详情结构一致） */
    doc: Record<string, any> | null
    /** 防抖毫秒 */
    debounceMs?: number
  }>(),
  {
    doc: null,
    debounceMs: 320,
  },
)

const loading = ref(false)
const error = ref('')
const htmlSrc = ref('')
const frameHeight = ref('1123px')
const frameRef = ref<HTMLIFrameElement | null>(null)
let timer: ReturnType<typeof setTimeout> | null = null
let reqSeq = 0

/** 按内容高度撑开 iframe。 */
function syncFrameHeight() {
  const win = frameRef.value?.contentWindow
  const doc = win?.document
  if (!doc?.body) return
  const h = Math.max(doc.body.scrollHeight, doc.documentElement?.scrollHeight || 0, 1123)
  frameHeight.value = `${h + 8}px`
}

/** 请求后端渲染纸面 HTML。 */
async function renderNow(doc: Record<string, any>) {
  const seq = ++reqSeq
  loading.value = true
  error.value = ''
  try {
    // 1. 提交文档取 HTML
    const res = await resumeApi.renderHtml({
      title: doc.title || '',
      template_id: doc.template_id || 'classic',
      theme_json: doc.theme_json || {},
      custom_css: doc.custom_css || '',
      profile_json: doc.profile_json || {},
      sections: Array.isArray(doc.sections) ? doc.sections : [],
    })
    // 2. 丢弃过期响应
    if (seq !== reqSeq) return
    htmlSrc.value = String(res.data?.html || '')
    if (!htmlSrc.value) {
      error.value = '预览为空'
    }
  } catch (err: any) {
    if (seq !== reqSeq) return
    error.value = err?.message || '预览失败'
    htmlSrc.value = ''
  } finally {
    if (seq === reqSeq) {
      loading.value = false
    }
  }
}

function scheduleRender(doc: Record<string, any> | null) {
  if (timer) clearTimeout(timer)
  if (!doc) {
    htmlSrc.value = ''
    error.value = ''
    return
  }
  timer = setTimeout(
    () => {
      void renderNow(doc)
    },
    Math.max(0, props.debounceMs),
  )
}

watch(
  () => props.doc,
  (doc) => scheduleRender(doc),
  { deep: true, immediate: true },
)

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
  reqSeq += 1
})
</script>

<template>
  <div class="resume-html-preview">
    <div v-if="loading && !htmlSrc" class="resume-html-preview__spin">
      <NSpin size="large" />
    </div>
    <NEmpty v-else-if="error && !htmlSrc" :description="error" class="py-16" />
    <iframe
      v-else-if="htmlSrc"
      ref="frameRef"
      class="resume-html-preview__frame"
      title="简历缩略"
      sandbox="allow-same-origin"
      referrerpolicy="no-referrer"
      :srcdoc="htmlSrc"
      :style="{ height: frameHeight }"
      @load="syncFrameHeight"
    />
    <NEmpty v-else description="暂无预览" class="py-16" />
  </div>
</template>

<style scoped>
.resume-html-preview {
  position: relative;
  width: 100%;
  background: #fff;
}

.resume-html-preview__frame {
  display: block;
  width: 100%;
  border: 0;
  background: #fff;
}

.resume-html-preview__spin {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  padding: 24px 0;
}
</style>
