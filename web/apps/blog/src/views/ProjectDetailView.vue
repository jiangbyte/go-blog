<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getProject } from '@/api'
import PostArticle from '@/components/post/PostArticle.vue'

const route = useRoute()
const project = ref<any>(null)

watch(
  () => String(route.params.uid),
  async (uid) => {
    const { project: p } = await getProject({ uid })
    project.value = p
  },
  { immediate: true },
)
</script>

<template>
  <div v-if="project" class="flex flex-col gap-4">
    <article class="glass-card p-5 md:p-8">
      <PostArticle :body="project.body" />
    </article>
  </div>
  <div v-else class="glass-card p-8 text-center text-muted">项目不存在</div>
</template>
