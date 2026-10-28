<script setup lang="ts">
import TocList from '@/components/aside/TocList.vue'
import { useArticleToc } from '@/composables/useArticleToc'

const { headings } = useArticleToc()
</script>

<template>
  <section
    v-if="headings.length"
    class="glass-card flex max-h-[calc(100vh-5.5rem)] flex-col overflow-hidden p-3"
  >
    <div class="mb-1.5 shrink-0 px-1 text-sm font-semibold">目录</div>
    <TocList class="min-h-0 flex-1" auto-scroll />
  </section>
</template>
