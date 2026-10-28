<!-- Author: go-blog — JSON 格式化 / 压缩 / 校验（左右分栏） -->

<script setup lang="ts">
import { computed, ref } from 'vue'

const input = ref('{\n  "hello": "world"\n}')
const output = ref('')
const error = ref('')
const message = ref('')

const inputCount = computed(() => input.value.length)
const outputCount = computed(() => output.value.length)

function clearFeedback() {
  error.value = ''
  message.value = ''
}

function parseInput(): unknown | null {
  clearFeedback()
  try {
    return JSON.parse(input.value)
  } catch (e: any) {
    error.value = e?.message || 'JSON 解析失败'
    output.value = ''
    return null
  }
}

function formatPretty() {
  const data = parseInput()
  if (data === null) return
  output.value = JSON.stringify(data, null, 2)
  message.value = '已格式化'
}

function minify() {
  const data = parseInput()
  if (data === null) return
  output.value = JSON.stringify(data)
  message.value = '已压缩'
}

function validate() {
  const data = parseInput()
  if (data === null) return
  output.value = JSON.stringify(data, null, 2)
  message.value = 'JSON 合法'
}

async function copyOutput() {
  clearFeedback()
  const text = output.value
  if (!text.trim()) {
    error.value = '右侧暂无输出，请先格式化或校验'
    return
  }
  try {
    await navigator.clipboard.writeText(text)
    message.value = '已复制输出到剪贴板'
  } catch {
    error.value = '复制失败，请手动选择复制'
  }
}

function clearAll() {
  input.value = ''
  output.value = ''
  clearFeedback()
}

function syncToInput() {
  if (!output.value.trim()) return
  input.value = output.value
  message.value = '已用输出覆盖左侧输入'
  error.value = ''
}
</script>

<template>
  <div class="tool-panel glass-card">
    <div class="tool-toolbar">
      <button type="button" class="tool-btn tool-btn--primary" @click="formatPretty">格式化</button>
      <button type="button" class="tool-btn" @click="minify">压缩</button>
      <button type="button" class="tool-btn" @click="validate">校验</button>
      <button type="button" class="tool-btn" @click="copyOutput">复制输出</button>
      <button type="button" class="tool-btn" @click="syncToInput">输出→输入</button>
      <button type="button" class="tool-btn tool-btn--ghost" @click="clearAll">清空</button>
    </div>

    <p v-if="error" class="tool-feedback tool-feedback--error">{{ error }}</p>
    <p v-else-if="message" class="tool-feedback tool-feedback--ok">{{ message }}</p>

    <div class="tool-split">
      <section class="tool-pane">
        <header class="tool-pane__head">
          <span>输入</span>
          <span class="tool-meta">{{ inputCount }} 字符</span>
        </header>
        <textarea
          v-model="input"
          class="tool-textarea"
          spellcheck="false"
          placeholder="粘贴 JSON…"
        />
      </section>
      <section class="tool-pane">
        <header class="tool-pane__head">
          <span>输出</span>
          <span class="tool-meta">{{ outputCount }} 字符</span>
        </header>
        <textarea
          v-model="output"
          class="tool-textarea"
          spellcheck="false"
          readonly
          placeholder="格式化 / 压缩结果将显示在这里"
        />
      </section>
    </div>
  </div>
</template>

<style scoped>
.tool-panel {
  padding: 1rem 1.1rem 1.15rem;
}

.tool-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.tool-btn {
  min-height: 2rem;
  padding: 0.3rem 0.85rem;
  border: 1px solid var(--efu-card-border);
  border-radius: 8px;
  background: var(--efu-secondbg);
  color: var(--milk-ink);
  font-size: 0.85rem;
  cursor: pointer;
  transition:
    background 0.15s ease,
    border-color 0.15s ease;
}

.tool-btn:hover {
  background: var(--efu-ahoverbg);
}

.tool-btn--primary {
  border-color: transparent;
  background: var(--milk-ink);
  color: #fff;
}

.tool-btn--primary:hover {
  opacity: 0.9;
  background: var(--milk-ink);
  color: #fff;
}

.tool-btn--ghost {
  background: transparent;
}

.tool-meta {
  font-size: 0.75rem;
  color: var(--efu-secondtext);
}

.tool-feedback {
  margin: 0 0 0.65rem;
  font-size: 0.85rem;
}

.tool-feedback--error {
  color: #d4483a;
}

.tool-feedback--ok {
  color: #2f8f5b;
}

.tool-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.85rem;
  min-height: 22rem;
}

.tool-pane {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.tool-pane__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.4rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--milk-ink);
}

.tool-textarea {
  display: block;
  flex: 1;
  width: 100%;
  min-height: 22rem;
  box-sizing: border-box;
  padding: 0.85rem 1rem;
  border: 1px solid var(--efu-card-border);
  border-radius: 10px;
  background: var(--efu-secondbg);
  color: var(--milk-ink);
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 0.86rem;
  line-height: 1.55;
  resize: vertical;
}

.tool-textarea:focus {
  outline: 2px solid color-mix(in srgb, var(--milk-ink) 22%, transparent);
  outline-offset: 1px;
}

.tool-textarea[readonly] {
  cursor: default;
}

@media (max-width: 839px) {
  .tool-split {
    grid-template-columns: 1fr;
  }
}
</style>
