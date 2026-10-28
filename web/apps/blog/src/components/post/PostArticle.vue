<script setup lang="ts">
/**
 * 正文 DOM 后处理（渲染与 enhance 来自 @jiangbyte/markdown）。
 */
import { enhanceMarkdownRoot, renderMarkdown } from '@jiangbyte/markdown'
import { computed, nextTick, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'

const props = defineProps<{ body: string }>()
const html = computed(() => renderMarkdown(props.body))
const root = useTemplateRef<HTMLElement>('root')

const previewSrc = ref<string | null>(null)
const previewAlt = ref('')

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

async function run() {
  await nextTick()
  if (root.value) {
    await enhanceMarkdownRoot(root.value, { onImageClick: openPreview })
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
  void run()
})
onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
})
watch(html, run, { flush: 'post' })
</script>

<template>
  <div ref="root" class="prose-body article-container" v-html="html" />

  <Teleport to="body">
    <div
      v-if="previewSrc"
      class="img-preview"
      role="dialog"
      aria-modal="true"
      aria-label="图片预览"
      @click="closePreview"
    >
      <button type="button" class="img-preview__close" aria-label="关闭" @click="closePreview">
        ×
      </button>
      <img
        :src="previewSrc"
        :alt="previewAlt"
        class="img-preview__img"
        @click.stop
      />
    </div>
  </Teleport>
</template>

<style scoped>
.img-preview {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background: rgba(12, 12, 14, 0.82);
  backdrop-filter: blur(6px);
  cursor: zoom-out;
}
.img-preview__img {
  max-width: min(96vw, 1200px);
  max-height: 90vh;
  width: auto;
  height: auto;
  object-fit: contain;
  border-radius: 8px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.35);
  cursor: default;
}
.img-preview__close {
  position: absolute;
  top: 1rem;
  right: 1.25rem;
  width: 2.5rem;
  height: 2.5rem;
  border: 0;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
  font-size: 1.6rem;
  line-height: 1;
  cursor: pointer;
}
.img-preview__close:hover {
  background: rgba(255, 255, 255, 0.28);
}
</style>
