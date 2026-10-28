<script setup lang="ts">
/**
 * 公开简历 PDF：同源 stream + 浏览器原生查看器；元数据与 iframe 两段加载都有等待反馈。
 */
import { computed, nextTick, onMounted, ref } from 'vue'

type Status = 'loading' | 'ready' | 'empty' | 'error'

const PDF_STREAM_BASE = '/api/portal/resume/resume/pdf?stream=1'

const status = ref<Status>('loading')
const message = ref('')
/** 刷新时改 query，逼 iframe 重载 */
const nonce = ref(0)
/** iframe 仍在拉 PDF / 渲染时盖住灰底 */
const iframePending = ref(false)

const pdfSrc = computed(() => `${PDF_STREAM_BASE}&t=${nonce.value}`)
const showWait = computed(
  () => status.value === 'loading' || (status.value === 'ready' && iframePending.value),
)
const waitText = computed(() =>
  status.value === 'loading' ? '正在检查公开简历…' : '正在加载 PDF…',
)

/**
 * 用元数据判断是否有公开 PDF，再交给原生查看器加载。
 */
async function loadPdf() {
  status.value = 'loading'
  message.value = ''
  iframePending.value = false
  try {
    // 1. 查公开简历；无简历 / 无 pdf 走空态
    const res = await fetch('/api/portal/resume/resume', {
      method: 'GET',
      credentials: 'same-origin',
    })

    if (res.status === 404) {
      status.value = 'empty'
      message.value = '暂无公开简历'
      return
    }

    let body: {
      code?: string
      message?: string
      data?: { pdf_url?: string; pdf_file_id?: string }
    } = {}
    try {
      body = await res.json()
    } catch {
      /* ignore */
    }

    if (!res.ok) {
      const detail = body.message || `请求失败（${res.status}）`
      if (detail.includes('暂无')) {
        status.value = 'empty'
        message.value = detail
        return
      }
      status.value = 'error'
      message.value = detail
      return
    }

    const hasPdf =
      Boolean(String(body.data?.pdf_url || '').trim()) ||
      Boolean(String(body.data?.pdf_file_id || '').trim())
    if (!hasPdf) {
      status.value = 'empty'
      message.value = '暂无公开简历'
      return
    }

    // 2. 先盖等待层，再挂 iframe（PDF 流可能要数秒）
    iframePending.value = true
    nonce.value += 1
    status.value = 'ready'
    await nextTick()
  } catch (e) {
    iframePending.value = false
    status.value = 'error'
    message.value = e instanceof Error ? e.message : '加载失败'
  }
}

function onIframeLoad() {
  iframePending.value = false
}

onMounted(() => {
  void loadPdf()
})
</script>

<template>
  <div class="resume-page">
    <header class="resume-bar">
      <span class="resume-bar__title">Resume</span>
      <div class="resume-bar__actions">
        <a
          v-if="status === 'ready' && !iframePending"
          class="resume-btn"
          :href="PDF_STREAM_BASE"
          target="_blank"
          rel="noopener noreferrer"
        >
          下载 PDF
        </a>
        <button
          type="button"
          class="resume-btn resume-btn--ghost"
          :disabled="showWait"
          @click="loadPdf"
        >
          刷新
        </button>
      </div>
    </header>

    <main class="resume-main">
      <div v-if="status === 'empty'" class="resume-state">{{ message || '暂无公开简历' }}</div>
      <div v-else-if="status === 'error'" class="resume-state resume-state--error">
        {{ message || '加载失败' }}
      </div>
      <template v-else>
        <iframe
          v-if="status === 'ready'"
          class="resume-pdf"
          :class="{ 'resume-pdf--hidden': iframePending }"
          :src="pdfSrc"
          title="Resume PDF"
          @load="onIframeLoad"
        />
        <div v-if="showWait" class="resume-wait" aria-live="polite" aria-busy="true">
          <span class="resume-wait__spinner" aria-hidden="true" />
          <span class="resume-wait__text">{{ waitText }}</span>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
.resume-page {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-height: 100vh;
  overflow: hidden;
}

.resume-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-shrink: 0;
  height: 48px;
  padding: 0 16px;
  border-bottom: 1px solid #e4e4e7;
  background: #fff;
}

.resume-bar__title {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.resume-bar__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.resume-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 12px;
  border: 1px solid #18181b;
  background: #18181b;
  color: #fff;
  font-size: 13px;
  line-height: 1.2;
  text-decoration: none;
  cursor: pointer;
}

.resume-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.resume-btn--ghost {
  background: #fff;
  color: #18181b;
}

.resume-btn:hover:not(:disabled) {
  opacity: 0.88;
}

.resume-main {
  position: relative;
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
  overflow: hidden;
  background: #52525b;
}

.resume-pdf {
  display: block;
  width: 100%;
  height: calc(100vh - 48px);
  border: 0;
  background: #52525b;
}

.resume-pdf--hidden {
  opacity: 0;
  pointer-events: none;
}

.resume-wait {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  background: #52525b;
  color: #f4f4f5;
}

.resume-wait__spinner {
  width: 28px;
  height: 28px;
  border: 2px solid rgba(244, 244, 245, 0.25);
  border-top-color: #f4f4f5;
  border-radius: 50%;
  animation: resume-spin 0.7s linear infinite;
}

.resume-wait__text {
  font-size: 14px;
  letter-spacing: 0.02em;
}

@keyframes resume-spin {
  to {
    transform: rotate(360deg);
  }
}

.resume-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 50vh;
  padding: 24px;
  color: #e4e4e7;
  font-size: 14px;
  background: #52525b;
}

.resume-state--error {
  color: #fecaca;
}
</style>
