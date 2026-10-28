<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getPost, getPostNeighbors } from '@/api'
import PostArticle from '@/components/post/PostArticle.vue'
import PostNav from '@/components/post/PostNav.vue'

const route = useRoute()
const post = ref<any>(null)
const prev = ref<any | null>(null)
const next = ref<any | null>(null)

watch(
  () => String(route.params.uid),
  async (uid) => {
    const [{ post: p }, neighbors] = await Promise.all([
      getPost({ uid }),
      getPostNeighbors({ uid }),
    ])
    post.value = p
    prev.value = neighbors.prev
    next.value = neighbors.next
  },
  { immediate: true },
)
</script>

<template>
  <div v-if="post" class="flex flex-col gap-4">
    <article class="glass-card p-5 md:p-8">
      <PostArticle :body="post.body" />
    </article>
    <PostNav :prev="prev" :next="next" />
  </div>
  <div v-else class="glass-card p-8 text-center text-muted">文章不存在</div>
</template>
