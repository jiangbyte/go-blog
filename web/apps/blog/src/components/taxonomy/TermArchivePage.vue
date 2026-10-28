<script setup lang="ts">
import PostGrid from '@/components/home/PostGrid.vue'

defineProps<{
  title: string
  desc?: string
  items: any[]
}>()
</script>

<template>
  <section>
    <div class="mb-4">
      <h1 class="text-xl font-bold">{{ title }}</h1>
      <p v-if="desc" class="mt-1 text-sm text-muted">{{ desc }}</p>
    </div>
    <PostGrid :posts="items" />
    <p v-if="!items.length" class="py-8 text-center text-sm text-muted">暂无内容</p>
  </section>
</template>
