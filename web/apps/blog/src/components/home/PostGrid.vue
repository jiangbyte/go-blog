<script setup lang="ts">
import PostCard from '@/components/home/PostCard.vue'

defineProps<{
  /** Post 或归档条目（含 uid/title/description/cover/date/tags/categories） */
  posts: any[]
}>()
</script>

<template>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 sm:gap-5">
    <PostCard
      v-for="(post, i) in posts"
      :key="post.uid"
      :post="post"
      :index="Number(i)"
    />
  </div>
</template>
