<!-- Author: go-blog — 简历自研轻量正文编辑器（样式对齐纸面） -->

<script setup lang="ts">
import { Icon } from '@iconify/vue/offline'
import { NButton, NFlex, NIcon } from 'naive-ui'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { normalizeResumeHtml, toEditorHtml } from '@jiangbyte/shared'

const props = withDefaults(
  defineProps<{
    value?: string | null
    height?: number
    disabled?: boolean
    placeholder?: string
  }>(),
  {
    value: '',
    height: 200,
    disabled: false,
    placeholder: '输入正文…',
  },
)

const emit = defineEmits<{
  'update:value': [value: string]
  change: [value: string]
}>()

const editorRef = ref<HTMLElement | null>(null)
const focused = ref(false)
const isEmpty = ref(true)
const activeBold = ref(false)
const activeOl = ref(false)
const activeUl = ref(false)
let syncingFromProp = false

const editorStyle = computed(() => ({
  minHeight: `${props.height}px`,
  maxHeight: `${Math.max(props.height, 280)}px`,
}))

function refreshEmpty() {
  const text = (editorRef.value?.textContent || '').replace(/\u00a0/g, ' ').trim()
  const html = (editorRef.value?.innerHTML || '').replace(/<br\s*\/?>/gi, '').trim()
  isEmpty.value = !text && !html
}

function emitHtml() {
  if (!editorRef.value || props.disabled || syncingFromProp) return
  const html = normalizeResumeHtml(editorRef.value.innerHTML)
  emit('update:value', html)
  emit('change', html)
  refreshEmpty()
}

function applyCommand(command: string) {
  if (props.disabled) return
  editorRef.value?.focus()
  document.execCommand(command, false)
  refreshActiveStates()
  emitHtml()
}

function refreshActiveStates() {
  try {
    activeBold.value = document.queryCommandState('bold')
    activeOl.value = document.queryCommandState('insertOrderedList')
    activeUl.value = document.queryCommandState('insertUnorderedList')
  } catch {
    activeBold.value = false
    activeOl.value = false
    activeUl.value = false
  }
}

function onEditorInput() {
  if (syncingFromProp) return
  refreshEmpty()
  emitHtml()
  refreshActiveStates()
}

function onSelectionChange() {
  if (!editorRef.value || !focused.value) return
  const sel = document.getSelection()
  if (!sel || !sel.anchorNode) return
  if (!editorRef.value.contains(sel.anchorNode)) return
  refreshActiveStates()
}

function setEditorContent(html: string) {
  if (!editorRef.value) return
  syncingFromProp = true
  editorRef.value.innerHTML = html || ''
  refreshEmpty()
  void nextTick(() => {
    syncingFromProp = false
  })
}

watch(
  () => props.value,
  (val) => {
    // 编辑中不回写，避免光标乱跳、结构被反复重置
    if (focused.value) return
    const next = toEditorHtml(String(val || ''))
    const current = normalizeResumeHtml(editorRef.value?.innerHTML || '')
    if (next === current) return
    setEditorContent(next)
  },
)

watch(
  () => props.disabled,
  (disabled) => {
    if (editorRef.value) {
      editorRef.value.contentEditable = disabled ? 'false' : 'true'
    }
  },
)

onMounted(() => {
  setEditorContent(toEditorHtml(String(props.value || '')))
  if (editorRef.value) {
    editorRef.value.contentEditable = props.disabled ? 'false' : 'true'
  }
  document.addEventListener('selectionchange', onSelectionChange)
})

onBeforeUnmount(() => {
  document.removeEventListener('selectionchange', onSelectionChange)
})

function onFocus() {
  focused.value = true
}

function onBlur() {
  focused.value = false
  // 失焦再规范一次 DOM，并与外部 value 对齐
  if (editorRef.value) {
    const html = normalizeResumeHtml(editorRef.value.innerHTML)
    setEditorContent(html)
    emit('update:value', html)
    emit('change', html)
  }
}
</script>

<template>
  <div class="resume-body-editor" :class="{ 'is-disabled': disabled, 'is-empty': isEmpty }">
    <NFlex class="resume-body-editor__toolbar" :size="4" align="center">
      <NButton
        size="tiny"
        quaternary
        :type="activeBold ? 'primary' : 'default'"
        :disabled="disabled"
        title="加粗"
        aria-label="加粗"
        @mousedown.prevent
        @click="applyCommand('bold')"
      >
        <strong style="font-size: 13px; line-height: 1">B</strong>
      </NButton>
      <NButton
        size="tiny"
        quaternary
        :type="activeOl ? 'primary' : 'default'"
        :disabled="disabled"
        title="有序列表"
        aria-label="有序列表"
        @mousedown.prevent
        @click="applyCommand('insertOrderedList')"
      >
        <template #icon>
          <NIcon :size="16">
            <Icon icon="icon-park-outline:list" />
          </NIcon>
        </template>
      </NButton>
      <NButton
        size="tiny"
        quaternary
        :type="activeUl ? 'primary' : 'default'"
        :disabled="disabled"
        title="无序列表"
        aria-label="无序列表"
        @mousedown.prevent
        @click="applyCommand('insertUnorderedList')"
      >
        <template #icon>
          <NIcon :size="16">
            <Icon icon="icon-park-outline:list-two" />
          </NIcon>
        </template>
      </NButton>
      <span class="resume-body-editor__sep" />
      <NButton
        size="tiny"
        quaternary
        :disabled="disabled"
        title="减少缩进"
        aria-label="减少缩进"
        @mousedown.prevent
        @click="applyCommand('outdent')"
      >
        <template #icon>
          <NIcon :size="16">
            <Icon icon="icon-park-outline:indent-left" />
          </NIcon>
        </template>
      </NButton>
      <NButton
        size="tiny"
        quaternary
        :disabled="disabled"
        title="增加缩进"
        aria-label="增加缩进"
        @mousedown.prevent
        @click="applyCommand('indent')"
      >
        <template #icon>
          <NIcon :size="16">
            <Icon icon="icon-park-outline:indent-right" />
          </NIcon>
        </template>
      </NButton>
    </NFlex>

    <div
      ref="editorRef"
      class="resume-body-editor__body resume-md"
      :style="editorStyle"
      :data-placeholder="placeholder"
      role="textbox"
      aria-multiline="true"
      @input="onEditorInput"
      @focus="onFocus"
      @blur="onBlur"
      @keyup="refreshActiveStates"
      @mouseup="refreshActiveStates"
    />
  </div>
</template>

<style scoped>
.resume-body-editor {
  width: 100%;
  min-width: 0;
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius, 3px);
  overflow: hidden;
  background: var(--card-color, #fff);
}

.resume-body-editor.is-disabled {
  opacity: 0.65;
  pointer-events: none;
}

.resume-body-editor__toolbar {
  padding: 6px 8px;
  border-bottom: 1px solid var(--border-color);
  background: var(--app-muted-bg, var(--body-color));
}

.resume-body-editor__sep {
  width: 1px;
  height: 16px;
  margin: 0 2px;
  background: var(--border-color);
  flex: none;
}

/* 编辑区视觉对齐简历纸面 classic */
.resume-body-editor__body {
  padding: 12px 14px;
  outline: none;
  overflow-y: auto;
  box-sizing: border-box;
  background: #fff;
  color: #222;
  font-size: 12.5px;
  line-height: var(--line-height, 1.55);
  font-family: 'PingFang SC', 'Microsoft YaHei', 'Noto Sans SC', sans-serif;
  word-break: break-word;
}

.resume-body-editor.is-empty .resume-body-editor__body::before {
  content: attr(data-placeholder);
  color: #999;
  pointer-events: none;
  float: left;
  height: 0;
}

.resume-body-editor__body :deep(p) {
  margin: 0 0 0.35em;
  line-height: var(--line-height, 1.55);
}

.resume-body-editor__body :deep(ul),
.resume-body-editor__body :deep(ol) {
  margin: 0 0 0.35em;
  padding-left: 1.2em;
  line-height: var(--line-height, 1.55);
}

.resume-body-editor__body :deep(ul) {
  list-style: disc;
}

.resume-body-editor__body :deep(ol) {
  list-style: decimal;
}

.resume-body-editor__body :deep(li) {
  margin: 0 0 0.28em;
  padding-left: 1px;
  line-height: var(--line-height, 1.55);
}

.resume-body-editor__body :deep(li:last-child),
.resume-body-editor__body :deep(p:last-child) {
  margin-bottom: 0;
}

.resume-body-editor__body :deep(strong),
.resume-body-editor__body :deep(b) {
  font-weight: 700;
}
</style>
