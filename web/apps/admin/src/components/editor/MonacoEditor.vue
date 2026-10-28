<!-- Author: go-blog -->

<script setup lang="ts">
import type * as monaco from 'monaco-editor'
import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { toCssSize } from './shared'
import './monacoWorkers'

const props = withDefaults(
  defineProps<{
    value?: string | null
    language?: string
    theme?: string
    height?: string | number
    readOnly?: boolean
    options?: monaco.editor.IStandaloneEditorConstructionOptions
  }>(),
  {
    value: '',
    language: 'typescript',
    theme: 'vs',
    height: 360,
    readOnly: false,
    options: () => ({}),
  },
)

const emit = defineEmits<{
  'update:value': [value: string]
  change: [value: string]
}>()

const containerRef = ref<HTMLElement | null>(null)
const monacoRef = shallowRef<typeof monaco | null>(null)
const editorRef = shallowRef<monaco.editor.IStandaloneCodeEditor | null>(null)
let disposed = false
const containerStyle = computed(() => ({
  height: toCssSize(props.height),
}))

async function loadMonaco() {
  if (!monacoRef.value) {
    monacoRef.value = await import('monaco-editor')
  }
  return monacoRef.value
}

function syncValue(value: string) {
  const editor = editorRef.value
  if (!editor || editor.getValue() === value) {
    return
  }
  editor.setValue(value)
}

onMounted(async () => {
  if (!containerRef.value) {
    return
  }

  const monacoInstance = await loadMonaco()
  if (disposed || !containerRef.value) {
    return
  }
  editorRef.value = monacoInstance.editor.create(containerRef.value, {
    ...props.options,
    value: props.value ?? '',
    language: props.language,
    theme: props.theme,
    readOnly: props.readOnly,
    automaticLayout: true,
    minimap: {
      enabled: false,
    },
    scrollBeyondLastLine: false,
  })

  editorRef.value.onDidChangeModelContent(() => {
    const value = editorRef.value?.getValue() ?? ''
    emit('update:value', value)
    emit('change', value)
  })
})

watch(
  () => props.value,
  (value) => {
    syncValue(value ?? '')
  },
)

watch(
  () => props.language,
  (language) => {
    const monacoInstance = monacoRef.value
    const model = editorRef.value?.getModel()
    if (monacoInstance && model) {
      monacoInstance.editor.setModelLanguage(model, language)
    }
  },
)

watch(
  () => props.theme,
  (theme) => {
    monacoRef.value?.editor.setTheme(theme)
  },
)

watch(
  () => props.readOnly,
  (readOnly) => {
    editorRef.value?.updateOptions({ readOnly })
  },
)

watch(
  () => props.options,
  (options) => {
    editorRef.value?.updateOptions(options)
  },
)

onBeforeUnmount(() => {
  disposed = true
  const model = editorRef.value?.getModel()
  editorRef.value?.dispose()
  model?.dispose()
  editorRef.value = null
})

/** 供 Markdown 工具栏等上层封装调用 */
function getEditor() {
  return editorRef.value
}

function focus() {
  editorRef.value?.focus()
}

/**
 * 在选区插入/包裹文本。
 * before/after 为空时等价于纯插入；有选区时用选区内容作为中间段。
 */
function wrapSelection(before: string, after = '', placeholder = '') {
  const editor = editorRef.value
  if (!editor) return
  const selection = editor.getSelection()
  if (!selection) return
  const model = editor.getModel()
  if (!model) return
  const selected = model.getValueInRange(selection)
  const mid = selected || placeholder
  const text = `${before}${mid}${after}`
  editor.executeEdits('wrap-selection', [
    {
      range: selection,
      text,
      forceMoveMarkers: true,
    },
  ])
  // 无原选区时，把光标放到 placeholder 位置便于继续输入
  if (!selected && placeholder) {
    const start = selection.getStartPosition()
    const startOffset = model.getOffsetAt(start) + before.length
    const endOffset = startOffset + placeholder.length
    const startPos = model.getPositionAt(startOffset)
    const endPos = model.getPositionAt(endOffset)
    editor.setSelection({
      startLineNumber: startPos.lineNumber,
      startColumn: startPos.column,
      endLineNumber: endPos.lineNumber,
      endColumn: endPos.column,
    })
  }
  editor.focus()
}

/** 在当前行行首插入前缀（标题 / 列表等） */
function prefixLines(prefix: string) {
  const editor = editorRef.value
  if (!editor) return
  const selection = editor.getSelection()
  if (!selection) return
  const model = editor.getModel()
  if (!model) return
  const startLine = selection.startLineNumber
  const endLine = selection.endLineNumber
  const edits: monaco.editor.IIdentifiedSingleEditOperation[] = []
  for (let line = startLine; line <= endLine; line++) {
    edits.push({
      range: {
        startLineNumber: line,
        startColumn: 1,
        endLineNumber: line,
        endColumn: 1,
      },
      text: prefix,
      forceMoveMarkers: true,
    })
  }
  editor.executeEdits('prefix-lines', edits)
  editor.focus()
}

defineExpose({
  getEditor,
  focus,
  wrapSelection,
  prefixLines,
})
</script>

<template>
  <div ref="containerRef" class="monaco-editor" :style="containerStyle" />
</template>

<style scoped>
.monaco-editor {
  width: 100%;
  min-width: 0;
  border: 1px solid var(--border-color);
}
</style>
