<!-- Author: go-blog — 与前台同源的 Markdown 预览（非 md-editor 内置预览） -->

<script setup lang="ts">
import { enhanceMarkdownRoot, renderMarkdown } from '@jiangbyte/markdown'
import { computed, nextTick, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import { sanitizeEditorHtml, warmEditorMediaUrls } from './shared'
import '@fontsource/noto-sans-sc/400.css'
import '@fontsource/noto-sans-sc/500.css'
import '@fontsource/noto-sans-sc/700.css'
import '@jiangbyte/markdown/styles/prose-base.css'
import '@jiangbyte/markdown/styles/prose.css'
import 'katex/dist/katex.min.css'
import 'highlight.js/styles/github.css'

const props = withDefaults(
  defineProps<{
    value?: string | null
    /** 预览区高度，与编辑器对齐 */
    height?: string | number
    /** 输入防抖毫秒；左右分栏时避免每键全量重渲 */
    debounceMs?: number
  }>(),
  {
    value: '',
    height: 360,
    debounceMs: 280,
  },
)

const scrollHost = useTemplateRef<HTMLElement>('scrollHost')
const root = useTemplateRef<HTMLElement>('root')
const ready = ref(false)
const previewSrc = ref<string | null>(null)
const previewAlt = ref('')
const html = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null
let refreshSeq = 0

const hostStyle = computed(() => ({
  height: typeof props.height === 'number' ? `${props.height}px` : props.height,
}))

function openPreview(src: string, alt = '') {
  previewSrc.value = src
  previewAlt.value = alt
  document.body.style.overflow = 'hidden'
}

function closePreview() {
  previewSrc.value = null
  previewAlt.value = ''
  document.body.style.overflow = ''
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') closePreview()
}

/** 签发媒体 → 渲染 → sanitize 替换 img → enhance。 */
async function refresh() {
  const seq = ++refreshSeq
  const md = props.value ?? ''
  ready.value = false
  try {
    // 1. 预热 object_name → 可访问 URL
    await warmEditorMediaUrls(md)
    if (seq !== refreshSeq) return
    // 2. 与前台同一 renderMarkdown，再同步替换 img src
    html.value = sanitizeEditorHtml(renderMarkdown(md))
    await nextTick()
    if (seq !== refreshSeq) return
    // 3. DOM 增强（tabs / mermaid / 复制等）
    if (root.value) {
      await enhanceMarkdownRoot(root.value, { onImageClick: openPreview })
    }
  } finally {
    if (seq === refreshSeq) ready.value = true
  }
}

function scheduleRefresh(immediate = false) {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
    debounceTimer = null
  }
  if (immediate || props.debounceMs <= 0) {
    void refresh()
    return
  }
  debounceTimer = setTimeout(() => {
    debounceTimer = null
    void refresh()
  }, props.debounceMs)
}

watch(
  () => props.value,
  () => {
    scheduleRefresh(false)
  },
  { immediate: true },
)

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
  window.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
})

/** 供分栏滚动同步。 */
function getScrollEl(): HTMLElement | null {
  return scrollHost.value
}

function offsetTopWithin(host: HTMLElement, el: HTMLElement): number {
  return el.getBoundingClientRect().top - host.getBoundingClientRect().top + host.scrollTop
}

/** 视口顶部最近的 data-source-line（不超过顶部的最大行号）。 */
function getTopSourceLine(): number {
  const host = scrollHost.value
  if (!host) return 1
  const nodes = host.querySelectorAll<HTMLElement>('[data-source-line]')
  const top = host.scrollTop + 4
  let best = 1
  for (const node of nodes) {
    const line = Number(node.dataset.sourceLine)
    if (!Number.isFinite(line) || line <= 0) continue
    if (offsetTopWithin(host, node) <= top) best = line
    else break
  }
  return best
}

/** 滚到 <= line 的最后一个带 data-source-line 的块顶对齐。 */
function scrollToSourceLine(line: number) {
  const host = scrollHost.value
  if (!host) return
  const nodes = host.querySelectorAll<HTMLElement>('[data-source-line]')
  let target: HTMLElement | null = null
  for (const node of nodes) {
    const n = Number(node.dataset.sourceLine)
    if (!Number.isFinite(n) || n <= 0) continue
    if (n <= line) target = node
    else break
  }
  if (!target) return
  host.scrollTop = offsetTopWithin(host, target)
}

defineExpose({ getScrollEl, getTopSourceLine, scrollToSourceLine })
</script>

<template>
  <div ref="scrollHost" class="article-md-preview" :style="hostStyle">
    <div
      ref="root"
      class="prose-body article-container article-md-preview__body"
      v-html="html"
    />
    <div v-if="!ready" class="article-md-preview__loading">预览生成中…</div>
  </div>

  <Teleport to="body">
    <div
      v-if="previewSrc"
      class="article-img-preview"
      role="dialog"
      aria-modal="true"
      aria-label="图片预览"
      @click="closePreview"
    >
      <button
        type="button"
        class="article-img-preview__close"
        aria-label="关闭"
        @click="closePreview"
      >
        ×
      </button>
      <img
        :src="previewSrc"
        :alt="previewAlt"
        class="article-img-preview__img"
        @click.stop
      />
    </div>
  </Teleport>
</template>

<style scoped>
.article-md-preview {
  position: relative;
  width: 100%;
  min-width: 0;
  overflow: auto;
  box-sizing: border-box;
  padding: 16px 20px 24px;
  background: var(--card-color, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: var(--border-radius, 0);
}

.article-md-preview__body {
  max-width: 720px;
  margin: 0 auto;
}

.article-md-preview__loading {
  position: absolute;
  top: 12px;
  right: 16px;
  font-size: 12px;
  color: rgba(26, 24, 22, 0.45);
  pointer-events: none;
}

.article-img-preview {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background: rgba(12, 12, 14, 0.82);
  backdrop-filter: blur(6px);
  cursor: zoom-out;
}

.article-img-preview__img {
  max-width: min(96vw, 1200px);
  max-height: 90vh;
  width: auto;
  height: auto;
  object-fit: contain;
  border-radius: var(--border-radius, 0);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.35);
  cursor: default;
}

.article-img-preview__close {
  position: absolute;
  top: 1rem;
  right: 1.25rem;
  width: 2.5rem;
  height: 2.5rem;
  border: 0;
  border-radius: var(--border-radius, 0);
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
  font-size: 1.6rem;
  line-height: 1;
  cursor: pointer;
}

.article-img-preview__close:hover {
  background: rgba(255, 255, 255, 0.28);
}
</style>

<style>
/* v-html 内容不受 scoped 影响；对齐前台链接色 */
.article-md-preview .prose-body a {
  color: inherit;
  text-decoration: none;
}
.article-md-preview .prose-body ::selection {
  background: color-mix(in srgb, var(--efu-theme) 22%, transparent);
  color: var(--milk-ink);
}
</style>
