<!-- Author: go-blog — 基于 CodeMirror 6 的轻量 Markdown 源码编辑（预览走 ArticleMdPreview） -->

<script setup lang="ts">
import type { EditorUploadedFile } from './shared'
import { markdown } from '@codemirror/lang-markdown'
import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands'
import { Compartment, EditorState } from '@codemirror/state'
import {
  EditorView,
  keymap,
  lineNumbers,
  drawSelection,
  dropCursor,
  highlightActiveLine,
  highlightActiveLineGutter,
  placeholder as cmPlaceholder,
} from '@codemirror/view'
import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { toCssSize, uploadEditorFile, warmEditorMediaUrls } from './shared'

const props = withDefaults(
  defineProps<{
    value?: string | null
    height?: string | number
    placeholder?: string
    readOnly?: boolean
    disabled?: boolean
    noUploadImg?: boolean
    /** 是否展示滚动同步勾选（分栏写作页开启） */
    showSyncScroll?: boolean
    syncScroll?: boolean
  }>(),
  {
    value: '',
    height: 360,
    placeholder: '请输入内容',
    readOnly: false,
    disabled: false,
    noUploadImg: false,
    showSyncScroll: false,
    syncScroll: true,
  },
)

const emit = defineEmits<{
  'update:value': [value: string]
  'update:syncScroll': [value: boolean]
  change: [value: string]
  uploaded: [file: EditorUploadedFile]
  'upload-error': [error: unknown]
}>()

const hostRef = ref<HTMLElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const viewRef = shallowRef<EditorView | null>(null)
const uploading = ref(false)
const editableCompartment = new Compartment()
/** 避免外部 v-model 回写再触发 emit */
let applyingExternal = false

const hostStyle = computed(() => ({
  height: toCssSize(props.height),
}))

const charCount = computed(() => (props.value ?? '').length)
const locked = computed(() => props.disabled || props.readOnly)

type ToolbarItem = {
  key: string
  icon: string
  title: string
  run: () => void
}

function emitValue(next: string) {
  emit('update:value', next)
  emit('change', next)
}

/** 在选区两侧包裹标记；无选区时插入占位并选中。 */
function wrapSelection(before: string, after = before, sample = '文本') {
  const view = viewRef.value
  if (!view || locked.value) return
  const { from, to } = view.state.selection.main
  const selected = view.state.sliceDoc(from, to)
  const inner = selected || sample
  const insert = `${before}${inner}${after}`
  view.dispatch({
    changes: { from, to, insert },
    selection: {
      anchor: from + before.length,
      head: from + before.length + inner.length,
    },
    scrollIntoView: true,
  })
  view.focus()
}

/** 当前行（或选中各行）前缀插入。 */
function prefixLines(prefix: string) {
  const view = viewRef.value
  if (!view || locked.value) return
  const { from, to } = view.state.selection.main
  const startLine = view.state.doc.lineAt(from)
  const endLine = view.state.doc.lineAt(to)
  const changes: { from: number; to: number; insert: string }[] = []
  for (let n = startLine.number; n <= endLine.number; n++) {
    const line = view.state.doc.line(n)
    changes.push({ from: line.from, to: line.from, insert: prefix })
  }
  view.dispatch({ changes, scrollIntoView: true })
  view.focus()
}

function insertBlock(snippet: string) {
  const view = viewRef.value
  if (!view || locked.value) return
  const { from, to } = view.state.selection.main
  const needsNl = from > 0 && view.state.sliceDoc(from - 1, from) !== '\n'
  const insert = `${needsNl ? '\n' : ''}${snippet}`
  view.dispatch({
    changes: { from, to, insert },
    selection: { anchor: from + insert.length },
    scrollIntoView: true,
  })
  view.focus()
}

function insertAtCursor(text: string) {
  const view = viewRef.value
  if (!view || locked.value) return
  const { from, to } = view.state.selection.main
  view.dispatch({
    changes: { from, to, insert: text },
    selection: { anchor: from + text.length },
    scrollIntoView: true,
  })
  view.focus()
}

const toolbar: ToolbarItem[] = [
  { key: 'bold', icon: 'lucide:bold', title: '加粗', run: () => wrapSelection('**') },
  { key: 'italic', icon: 'lucide:italic', title: '斜体', run: () => wrapSelection('*') },
  {
    key: 'strike',
    icon: 'lucide:strikethrough',
    title: '删除线',
    run: () => wrapSelection('~~'),
  },
  { key: 'h2', icon: 'lucide:heading-2', title: '二级标题', run: () => prefixLines('## ') },
  { key: 'h3', icon: 'lucide:heading-3', title: '三级标题', run: () => prefixLines('### ') },
  { key: 'quote', icon: 'lucide:quote', title: '引用', run: () => prefixLines('> ') },
  { key: 'ul', icon: 'lucide:list', title: '无序列表', run: () => prefixLines('- ') },
  {
    key: 'ol',
    icon: 'lucide:list-ordered',
    title: '有序列表',
    run: () => prefixLines('1. '),
  },
  {
    key: 'task',
    icon: 'lucide:list-checks',
    title: '任务列表',
    run: () => prefixLines('- [ ] '),
  },
  { key: 'code', icon: 'lucide:code', title: '行内代码', run: () => wrapSelection('`') },
  {
    key: 'codeblock',
    icon: 'lucide:square-code',
    title: '代码块',
    run: () => insertBlock('```\n\n```\n'),
  },
  {
    key: 'link',
    icon: 'lucide:link',
    title: '链接',
    run: () => wrapSelection('[', '](https://)', '链接文字'),
  },
  {
    key: 'table',
    icon: 'lucide:table',
    title: '表格',
    run: () => insertBlock('| 列1 | 列2 |\n| --- | --- |\n|  |  |\n'),
  },
  { key: 'hr', icon: 'lucide:minus', title: '分隔线', run: () => insertBlock('---\n') },
]

async function uploadImages(files: File[]) {
  if (props.noUploadImg || locked.value || !files.length) return
  uploading.value = true
  try {
    for (const file of files) {
      if (!file.type.startsWith('image/')) continue
      try {
        const uploaded = await uploadEditorFile(file)
        const stableRef = uploaded.objectName || uploaded.value
        if (stableRef) {
          const alt = uploaded.name || 'image'
          insertAtCursor(`![${alt}](${stableRef})\n`)
        }
        emit('uploaded', uploaded)
      } catch (error) {
        emit('upload-error', error)
        window.$message?.error('上传失败')
      }
    }
  } finally {
    uploading.value = false
  }
}

function onPickImage() {
  if (props.noUploadImg || locked.value) return
  fileInputRef.value?.click()
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const files = [...(input.files ?? [])]
  input.value = ''
  void uploadImages(files)
}

function onPaste(event: ClipboardEvent) {
  if (props.noUploadImg || locked.value) return false
  const items = event.clipboardData?.items
  if (!items) return false
  const files: File[] = []
  for (const item of items) {
    if (item.type.startsWith('image/')) {
      const file = item.getAsFile()
      if (file) files.push(file)
    }
  }
  if (!files.length) return false
  event.preventDefault()
  void uploadImages(files)
  return true
}

function onDrop(event: DragEvent) {
  if (props.noUploadImg || locked.value) return false
  const files = [...(event.dataTransfer?.files ?? [])].filter((f) =>
    f.type.startsWith('image/'),
  )
  if (!files.length) return false
  event.preventDefault()
  void uploadImages(files)
  return true
}

function createEditor(doc: string) {
  if (!hostRef.value) return

  const state = EditorState.create({
    doc,
    extensions: [
      lineNumbers(),
      highlightActiveLineGutter(),
      highlightActiveLine(),
      drawSelection(),
      dropCursor(),
      history(),
      markdown(),
      cmPlaceholder(props.placeholder),
      keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab]),
      EditorView.lineWrapping,
      editableCompartment.of(EditorView.editable.of(!locked.value)),
      EditorView.updateListener.of((update) => {
        if (!update.docChanged || applyingExternal) return
        emitValue(update.state.doc.toString())
      }),
      EditorView.domEventHandlers({
        paste: (event) => onPaste(event),
        drop: (event) => onDrop(event),
      }),
      EditorView.theme({
        '&': {
          height: '100%',
          fontSize: '14px',
        },
        '.cm-scroller': {
          fontFamily:
            'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
          lineHeight: '1.65',
        },
        '.cm-content': {
          padding: '12px 0',
          caretColor: 'var(--primary-color, #1677ff)',
        },
        '&.cm-focused': {
          outline: 'none',
        },
        '.cm-gutters': {
          backgroundColor: 'var(--app-muted-bg, #fafafa)',
          color: 'var(--text-color-3, #8c8c8c)',
          border: 'none',
          borderRight: '1px solid var(--border-color, #e5e7eb)',
        },
        '.cm-activeLine': {
          backgroundColor: 'rgba(22, 119, 255, 0.04)',
        },
        '.cm-activeLineGutter': {
          backgroundColor: 'rgba(22, 119, 255, 0.06)',
        },
      }),
    ],
  })

  viewRef.value = new EditorView({
    state,
    parent: hostRef.value,
  })
}

onMounted(() => {
  createEditor(props.value ?? '')
  void warmEditorMediaUrls(props.value ?? '')
})

watch(
  () => props.value,
  (value) => {
    const view = viewRef.value
    const next = value ?? ''
    if (!view || view.state.doc.toString() === next) return
    applyingExternal = true
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: next },
    })
    applyingExternal = false
    void warmEditorMediaUrls(next)
  },
)

watch(locked, (isLocked) => {
  viewRef.value?.dispatch({
    effects: editableCompartment.reconfigure(EditorView.editable.of(!isLocked)),
  })
})

onBeforeUnmount(() => {
  viewRef.value?.destroy()
  viewRef.value = null
})

/** 供分栏滚动同步：CodeMirror 实际滚动在 scrollDOM。 */
function getScrollEl(): HTMLElement | null {
  return viewRef.value?.scrollDOM ?? null
}

/** 视口顶部对应的源码行号（1-based）。 */
function getTopVisibleLine(): number {
  const view = viewRef.value
  if (!view) return 1
  const block = view.lineBlockAtHeight(view.scrollDOM.scrollTop + 2)
  return view.state.doc.lineAt(block.from).number
}

/** 将指定源码行滚到视口顶部。 */
function scrollToLine(line: number) {
  const view = viewRef.value
  if (!view) return
  const safe = Math.min(Math.max(1, Math.floor(line)), view.state.doc.lines)
  const lineObj = view.state.doc.line(safe)
  const block = view.lineBlockAt(lineObj.from)
  view.scrollDOM.scrollTop = block.top
}

defineExpose({ getScrollEl, getTopVisibleLine, scrollToLine })
</script>

<template>
  <div class="md-editor" :class="{ 'md-editor--disabled': locked }" :style="hostStyle">
    <div class="md-editor__toolbar">
      <button
        v-for="item in toolbar"
        :key="item.key"
        type="button"
        class="md-editor__btn"
        :title="item.title"
        :aria-label="item.title"
        :disabled="locked"
        @click="item.run"
      >
        <NovaIcon :icon="item.icon" :size="16" />
      </button>
      <button
        v-if="!noUploadImg"
        type="button"
        class="md-editor__btn"
        title="上传图片"
        aria-label="上传图片"
        :disabled="locked || uploading"
        @click="onPickImage"
      >
        <NovaIcon
          :icon="uploading ? 'lucide:loader-circle' : 'lucide:image'"
          :size="16"
          :class="{ 'md-editor__spin': uploading }"
        />
      </button>
      <input
        ref="fileInputRef"
        type="file"
        accept="image/*"
        multiple
        class="md-editor__file"
        @change="onFileChange"
      />
      <span class="md-editor__spacer" />
      <button
        v-if="showSyncScroll"
        type="button"
        class="md-editor__btn"
        :class="{ 'md-editor__btn--active': syncScroll }"
        :title="syncScroll ? '滚动同步：开' : '滚动同步：关'"
        :aria-label="syncScroll ? '关闭滚动同步' : '开启滚动同步'"
        :aria-pressed="syncScroll"
        @click="emit('update:syncScroll', !syncScroll)"
      >
        <NovaIcon
          :icon="syncScroll ? 'lucide:square-check' : 'lucide:square'"
          :size="16"
        />
      </button>
      <span class="md-editor__count">字数: {{ charCount }}</span>
    </div>
    <div ref="hostRef" class="md-editor__host" />
  </div>
</template>

<style scoped>
.md-editor {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-width: 0;
  min-height: 0;
  box-sizing: border-box;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: var(--border-radius, 0);
  background: var(--card-color, #fff);
  overflow: hidden;
}

.md-editor--disabled {
  opacity: 0.72;
}

.md-editor__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  padding: 6px 8px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
  background: var(--app-muted-bg, #fafafa);
}

.md-editor__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: 1px solid transparent;
  border-radius: var(--border-radius, 0);
  background: transparent;
  color: var(--text-color-2, #595959);
  line-height: 1;
  cursor: pointer;
}

.md-editor__spin {
  animation: md-editor-spin 0.8s linear infinite;
}

@keyframes md-editor-spin {
  to {
    transform: rotate(360deg);
  }
}

.md-editor__btn:hover:not(:disabled) {
  background: var(--hover-color, #f0f0f0);
  border-color: var(--border-color, #e5e7eb);
}

.md-editor__btn--active {
  color: var(--primary-color, #1677ff);
}

.md-editor__btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.md-editor__file {
  display: none;
}

.md-editor__spacer {
  flex: 1 1 auto;
}

.md-editor__count {
  font-size: 12px;
  color: var(--text-color-3, #8c8c8c);
  padding: 0 4px;
}

.md-editor__host {
  flex: 1 1 0;
  min-height: 0;
  overflow: hidden;
}

.md-editor__host :deep(.cm-editor) {
  height: 100%;
}
</style>
