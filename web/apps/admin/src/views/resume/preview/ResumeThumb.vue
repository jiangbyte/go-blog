<!-- Author: go-blog — 列表缩略纸面（HTML 轻量渲染；分页以编辑态 PDF 为准） -->

<script setup lang="ts">
import { computed } from 'vue'
import ResumeHtmlPreview from './ResumeHtmlPreview.vue'
import type { ResumePreviewDoc } from './ResumeSheetPreview.vue'

const props = withDefaults(
  defineProps<{
    doc: ResumePreviewDoc
    /** 相对 A4 宽度的缩放比例 */
    scale?: number
  }>(),
  { scale: 0.28 },
)

/** A4 宽约 794px@96dpi；高度按比例估算容器 */
const SHEET_W = 794
const SHEET_H = 1123

const hostStyle = computed(() => ({
  width: `${Math.round(SHEET_W * props.scale)}px`,
  height: `${Math.round(SHEET_H * props.scale)}px`,
}))

const scalerStyle = computed(() => ({
  width: `${SHEET_W}px`,
  transform: `scale(${props.scale})`,
}))
</script>

<template>
  <div class="resume-thumb" :style="hostStyle">
    <div class="resume-thumb__scaler" :style="scalerStyle">
      <ResumeHtmlPreview :doc="doc" :debounce-ms="0" />
    </div>
  </div>
</template>

<style scoped>
.resume-thumb {
  overflow: hidden;
  pointer-events: none;
  background: #fff;
}
.resume-thumb__scaler {
  transform-origin: top left;
}
</style>
