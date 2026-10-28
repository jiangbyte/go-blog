<script setup lang="ts">
import ProjectCard from '@/components/project/ProjectCard.vue'

defineProps<{ projects: any[] }>()
</script>

<template>
  <div class="projects-grid">
    <ProjectCard v-for="(p, i) in projects" :key="p.uid" :project="p" :index="i" />
  </div>
</template>

<style scoped>
.projects-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.85rem;
  width: 100%;
}

@media (min-width: 640px) {
  .projects-grid {
    /* 固定两列：单条时也不拉成通栏大卡（auto-fit 会撑满） */
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.9rem 1rem;
  }
}

@media (min-width: 1100px) {
  .projects-grid {
    /* auto-fill 保留空轨，条目少时卡片保持合理宽度 */
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem 1.1rem;
  }
}
</style>
