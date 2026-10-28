<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listCategories } from '@/api'
import { useRoute } from 'vue-router'

defineProps<{ active?: string }>()
const route = useRoute()
const categories = ref<any[]>([])

onMounted(async () => {
  const { list } = await listCategories({})
  categories.value = list
})
</script>

<template>
  <div class="mb-4 flex gap-2 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
    <RouterLink
      to="/"
      class="cat-chip shrink-0 rounded-full border px-3 py-1.5 text-sm transition"
      :class="route.path === '/' ? 'cat-chip--on' : 'cat-chip--off'"
    >
      首页
    </RouterLink>
    <RouterLink
      to="/archives"
      class="cat-chip shrink-0 rounded-full border px-3 py-1.5 text-sm transition"
      :class="route.path.startsWith('/archives') ? 'cat-chip--on' : 'cat-chip--off'"
    >
      全部文章
    </RouterLink>
    <RouterLink
      v-for="c in categories"
      :key="c.uid"
      :to="`/categories/${c.uid}`"
      class="cat-chip shrink-0 rounded-full border px-3 py-1.5 text-sm transition"
      :class="active === c.name ? 'cat-chip--on' : 'cat-chip--off'"
    >
      {{ c.name }}
    </RouterLink>
  </div>
</template>

<style scoped>
.cat-chip--on {
  border-color: var(--milk-ink);
  background: var(--milk-ink);
  color: #fff;
}
.cat-chip--off {
  border-color: var(--efu-card-border);
  background: var(--efu-card-bg);
  color: var(--milk-ink-soft);
}
.cat-chip--off:hover {
  background: var(--efu-ahoverbg);
  border-color: var(--efu-card-border);
  color: var(--milk-ink);
}
</style>
