<!-- Author: go-blog — 非写作页的轻量 MD 预览，委托 ArticleMdPreview -->

<script setup lang="ts">
import ArticleMdPreview from './ArticleMdPreview.vue'

withDefaults(
  defineProps<{
    value?: string | null
    height?: string | number
  }>(),
  {
    value: '',
    height: 360,
  },
)
</script>

<template>
  <ArticleMdPreview :value="value" :height="height" />
</template>
