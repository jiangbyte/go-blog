<!-- Author: go-blog — 左编辑右预览工作区；滚动同步按源码行对齐（非比例） -->

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue'
import ArticleMdPreview from './ArticleMdPreview.vue'
import MdEditor from './MdEditor.vue'
import { toCssSize } from './shared'

const props = withDefaults(
  defineProps<{
    value?: string | null
    height?: string | number
    placeholder?: string
    disabled?: boolean
    noUploadImg?: boolean
    /** 默认开启左右滚动同步 */
    defaultSyncScroll?: boolean
  }>(),
  {
    value: '',
    height: 480,
    placeholder: '请输入内容',
    disabled: false,
    noUploadImg: false,
    defaultSyncScroll: true,
  },
)

const emit = defineEmits<{
  'update:value': [value: string]
  change: [value: string]
}>()

type EditorScrollExpose = {
  getScrollEl: () => HTMLElement | null
  getTopVisibleLine: () => number
  scrollToLine: (line: number) => void
}

type PreviewScrollExpose = {
  getScrollEl: () => HTMLElement | null
  getTopSourceLine: () => number
  scrollToSourceLine: (line: number) => void
}

const editorRef = useTemplateRef<EditorScrollExpose>('editorRef')
const previewRef = useTemplateRef<PreviewScrollExpose>('previewRef')
const syncScroll = ref(props.defaultSyncScroll)
const rootStyle = computed(() => ({
  height: toCssSize(props.height),
}))
/** 同步过程中忽略对端 scroll，避免回环 */
let syncing = false
let unbindEditor: (() => void) | null = null
let unbindPreview: (() => void) | null = null

function updateValue(value: string) {
  emit('update:value', value)
  emit('change', value)
}

function bindEditorToPreview(): (() => void) | null {
  const editor = editorRef.value
  const preview = previewRef.value
  const el = editor?.getScrollEl()
  if (!el || !editor || !preview) return null
  const onScroll = () => {
    if (!syncScroll.value || syncing) return
    syncing = true
    preview.scrollToSourceLine(editor.getTopVisibleLine())
    requestAnimationFrame(() => {
      syncing = false
    })
  }
  el.addEventListener('scroll', onScroll, { passive: true })
  return () => el.removeEventListener('scroll', onScroll)
}

function bindPreviewToEditor(): (() => void) | null {
  const editor = editorRef.value
  const preview = previewRef.value
  const el = preview?.getScrollEl()
  if (!el || !editor || !preview) return null
  const onScroll = () => {
    if (!syncScroll.value || syncing) return
    syncing = true
    editor.scrollToLine(preview.getTopSourceLine())
    requestAnimationFrame(() => {
      syncing = false
    })
  }
  el.addEventListener('scroll', onScroll, { passive: true })
  return () => el.removeEventListener('scroll', onScroll)
}

async function rebind() {
  unbindEditor?.()
  unbindPreview?.()
  unbindEditor = null
  unbindPreview = null
  await nextTick()
  unbindEditor = bindEditorToPreview()
  unbindPreview = bindPreviewToEditor()
}

watch(syncScroll, (on) => {
  if (on) void rebind()
})

watch(
  () => props.height,
  () => {
    void rebind()
  },
)

watch(
  () => props.value,
  () => {
    // 预览重渲后需重新抓 scroll 节点；短暂延迟等 debounce 渲染
    window.setTimeout(() => {
      void rebind()
    }, 320)
  },
)

onMounted(() => {
  void rebind()
  requestAnimationFrame(() => {
    void rebind()
  })
  window.setTimeout(() => {
    void rebind()
  }, 80)
})

onBeforeUnmount(() => {
  unbindEditor?.()
  unbindPreview?.()
})
</script>

<template>
  <div class="md-split" :style="rootStyle">
    <div class="md-split__panes">
      <div class="md-split__pane">
        <MdEditor
          ref="editorRef"
          :value="value"
          height="100%"
          :placeholder="placeholder"
          :disabled="disabled"
          :no-upload-img="noUploadImg"
          show-sync-scroll
          v-model:sync-scroll="syncScroll"
          @update:value="updateValue"
        />
      </div>
      <div class="md-split__pane">
        <ArticleMdPreview ref="previewRef" :value="value" height="100%" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.md-split {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-width: 0;
  min-height: 0;
  box-sizing: border-box;
}

.md-split__panes {
  display: flex;
  flex: 1 1 0;
  gap: 12px;
  min-width: 0;
  min-height: 0;
}

.md-split__pane {
  flex: 1 1 0;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.md-split__pane > :deep(*) {
  flex: 1 1 0;
  min-height: 0;
}
</style>
