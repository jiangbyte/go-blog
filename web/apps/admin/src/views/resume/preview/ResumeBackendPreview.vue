<!-- Author: go-blog — 与导出同源的 PDF 预览：pdf.js 纯纸面，无浏览器阅读器嵌套 -->

<script setup lang="ts">
import { resumeApi } from '@/api'
import { getDocument, GlobalWorkerOptions, type PDFDocumentProxy } from 'pdfjs-dist'
import pdfWorker from 'pdfjs-dist/build/pdf.worker.min.mjs?url'
import { NEmpty, NSpin, NText } from 'naive-ui'
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

GlobalWorkerOptions.workerSrc = pdfWorker

const props = withDefaults(
  defineProps<{
    /** 纸面文档（与详情结构一致） */
    doc: Record<string, any> | null
    /** 防抖毫秒；PDF 较慢，编辑态宜 ≥ 800 */
    debounceMs?: number
  }>(),
  {
    doc: null,
    debounceMs: 900,
  },
)

const loading = ref(false)
const error = ref('')
const hasPdf = ref(false)
const pagesEl = ref<HTMLDivElement | null>(null)

let timer: ReturnType<typeof setTimeout> | null = null
let reqSeq = 0
let paintSeq = 0
let pdfDoc: PDFDocumentProxy | null = null
let lastDocFp = ''
let lastPaintWidth = 0
let painting = false
let resizeObserver: ResizeObserver | null = null
let resizeTimer: ReturnType<typeof setTimeout> | null = null

/** 内容指纹：避免 deep watch 因无关引用抖动反复打 PDF。 */
function docFingerprint(doc: Record<string, any> | null) {
  if (!doc) return ''
  return JSON.stringify({
    title: doc.title,
    template_id: doc.template_id,
    theme_json: doc.theme_json,
    custom_css: doc.custom_css,
    profile_json: doc.profile_json,
    sections: doc.sections,
  })
}

/** 释放 PDFDocumentProxy。 */
async function destroyPdfDoc() {
  if (!pdfDoc) return
  try {
    await pdfDoc.destroy()
  } catch {
    /* ignore */
  }
  pdfDoc = null
}

/** 清空纸面。 */
function clearPages() {
  pagesEl.value?.replaceChildren()
  lastPaintWidth = 0
}

/**
 * 把已加载的 pdfDoc 画到 canvas。
 * 先离屏拼好再一次性替换，避免闪白；仅宽度变化才重画。
 */
async function paintPages(force = false) {
  const host = pagesEl.value
  if (!host || !pdfDoc) return

  const cssWidth = Math.max(Math.floor(host.clientWidth || 0), 320)
  if (!force && lastPaintWidth > 0 && Math.abs(cssWidth - lastPaintWidth) < 2) {
    return
  }
  if (painting) return

  const seq = ++paintSeq
  painting = true
  try {
    // 1. 按宿主宽度算 scale；dpr 保证清晰
    const dpr = Math.min(window.devicePixelRatio || 1, 2)
    const frag = document.createDocumentFragment()

    for (let i = 1; i <= pdfDoc.numPages; i++) {
      if (seq !== paintSeq) return
      const page = await pdfDoc.getPage(i)
      const base = page.getViewport({ scale: 1 })
      const scale = cssWidth / base.width
      const viewport = page.getViewport({ scale: scale * dpr })
      const canvas = document.createElement('canvas')
      canvas.className = 'resume-be-preview__page'
      canvas.width = Math.floor(viewport.width)
      canvas.height = Math.floor(viewport.height)
      canvas.style.width = `${cssWidth}px`
      canvas.style.height = `${Math.floor(viewport.height / dpr)}px`
      // 2. 渲染到离屏节点，整页完成后再挂载
      await page.render({ canvas, viewport }).promise
      if (seq !== paintSeq) return
      frag.appendChild(canvas)
    }

    // 3. 一次性替换，旧纸面保留到新纸面就绪
    if (seq !== paintSeq) return
    host.replaceChildren(frag)
    lastPaintWidth = cssWidth
  } catch (err: any) {
    if (seq !== paintSeq) return
    error.value = err?.message || 'PDF 渲染失败'
    clearPages()
  } finally {
    if (seq === paintSeq) {
      painting = false
    }
  }
}

/** 从 ArrayBuffer 加载 PDF 并绘制。 */
async function loadAndPaint(bytes: ArrayBuffer, seq: number) {
  await destroyPdfDoc()
  const task = getDocument({ data: new Uint8Array(bytes) })
  const pdf = await task.promise
  if (seq !== reqSeq) {
    await pdf.destroy()
    return
  }
  pdfDoc = pdf
  hasPdf.value = true
  await nextTick()
  if (seq !== reqSeq) return
  bindResizeObserver()
  await paintPages(true)
}

/** 仅在宽度变化时重绘；高度因换页变化一律忽略，打断闪烁环。 */
function bindResizeObserver() {
  const el = pagesEl.value
  resizeObserver?.disconnect()
  resizeObserver = null
  if (!el) return
  resizeObserver = new ResizeObserver((entries) => {
    const w = Math.floor(entries[0]?.contentRect?.width || 0)
    if (!pdfDoc || w < 1) return
    if (lastPaintWidth > 0 && Math.abs(w - lastPaintWidth) < 2) return
    if (resizeTimer) clearTimeout(resizeTimer)
    resizeTimer = setTimeout(() => {
      void paintPages(false)
    }, 200)
  })
  resizeObserver.observe(el)
}

/** 用当前文档打 PDF，再画成干净纸面。 */
async function renderNow(doc: Record<string, any>) {
  const seq = ++reqSeq
  loading.value = true
  error.value = ''
  try {
    // 1. 请求与导出同源的 PDF
    const res = await resumeApi.renderPdf({
      title: doc.title || '',
      template_id: doc.template_id || 'classic',
      theme_json: doc.theme_json || {},
      custom_css: doc.custom_css || '',
      profile_json: doc.profile_json || {},
      sections: Array.isArray(doc.sections) ? doc.sections : [],
    })
    if (seq !== reqSeq) return

    // 2. 校验 blob / 误返回的 JSON 错误
    const data = res.data
    if (!(data instanceof Blob) || data.size === 0) {
      error.value = '预览为空'
      hasPdf.value = false
      clearPages()
      await destroyPdfDoc()
      return
    }
    if (data.type && data.type.includes('application/json')) {
      const text = await data.text()
      let msg = '预览失败'
      try {
        const json = JSON.parse(text)
        msg = String(json?.message || json?.msg || msg)
      } catch {
        /* ignore */
      }
      error.value = msg
      hasPdf.value = false
      clearPages()
      await destroyPdfDoc()
      return
    }

    // 3. 加载并绘制（旧纸面保留到新 PDF 就绪）
    await loadAndPaint(await data.arrayBuffer(), seq)
  } catch (err: any) {
    if (seq !== reqSeq) return
    error.value = err?.message || '预览失败'
    hasPdf.value = false
    clearPages()
    await destroyPdfDoc()
  } finally {
    if (seq === reqSeq) {
      loading.value = false
    }
  }
}

function scheduleRender(doc: Record<string, any> | null) {
  if (timer) clearTimeout(timer)
  const fp = docFingerprint(doc)
  if (!doc || !fp) {
    lastDocFp = ''
    error.value = ''
    hasPdf.value = false
    clearPages()
    void destroyPdfDoc()
    return
  }
  // 指纹未变则跳过，避免 deep 抖动
  if (fp === lastDocFp && hasPdf.value) return
  timer = setTimeout(() => {
    lastDocFp = fp
    void renderNow(doc)
  }, Math.max(0, props.debounceMs))
}

watch(
  () => docFingerprint(props.doc),
  () => scheduleRender(props.doc),
  { immediate: true },
)

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
  if (resizeTimer) clearTimeout(resizeTimer)
  reqSeq += 1
  paintSeq += 1
  resizeObserver?.disconnect()
  resizeObserver = null
  clearPages()
  void destroyPdfDoc()
})
</script>

<template>
  <div class="resume-be-preview">
    <div v-if="loading && !hasPdf" class="resume-be-preview__spin">
      <NSpin size="large" />
      <NText depth="3" class="mt-3 text-12px">正在生成 PDF 预览…</NText>
    </div>
    <NEmpty v-else-if="error && !hasPdf" :description="error" class="py-16" />
    <NEmpty v-else-if="!hasPdf && !loading" description="暂无预览" class="py-16" />
    <!-- 始终挂载，避免 v-if 导致 ref 丢失、反复重绑 ResizeObserver -->
    <div
      ref="pagesEl"
      class="resume-be-preview__pages"
      :class="{ 'is-hidden': !hasPdf }"
    />
    <div v-if="loading && hasPdf" class="resume-be-preview__badge">更新中…</div>
  </div>
</template>

<style scoped>
.resume-be-preview {
  position: relative;
  width: min(100%, 210mm);
  max-width: 100%;
  margin: 0 auto;
  min-height: 297mm;
  background: #e8eef5;
  box-sizing: border-box;
  overflow: visible;
}

.resume-be-preview__pages {
  display: flex;
  flex-direction: column;
  gap: 14px;
  width: 100%;
  padding: 0;
  box-sizing: border-box;
}

.resume-be-preview__pages.is-hidden {
  display: none;
}

.resume-be-preview__pages :deep(.resume-be-preview__page) {
  display: block;
  width: 100%;
  height: auto;
  background: #fff;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.12);
  border: 1px solid var(--border-color, #e5e7eb);
}

.resume-be-preview__spin {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 297mm;
  padding: 48px 0;
  background: #fff;
  border: 1px solid var(--border-color, #e5e7eb);
}

.resume-be-preview__badge {
  position: absolute;
  top: 8px;
  right: 10px;
  z-index: 2;
  padding: 2px 8px;
  font-size: 12px;
  color: var(--text-color-3, #6b7280);
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 4px;
  pointer-events: none;
}
</style>
