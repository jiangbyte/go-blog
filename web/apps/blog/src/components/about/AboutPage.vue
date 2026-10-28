<script setup lang="ts">
import { computed } from 'vue'
import PostArticle from '@/components/post/PostArticle.vue'

const props = defineProps<{ doc: any }>()

const body = computed(() => {
  const content = typeof props.doc?.content === 'string' ? props.doc.content : ''
  return content.trim()
})
</script>

<template>
  <section v-if="body" class="glass-card p-6 md:p-8">
    <PostArticle :body="body" />
  </section>
</template>
