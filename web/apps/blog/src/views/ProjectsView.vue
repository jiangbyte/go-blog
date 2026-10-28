<script setup lang="ts">
import { ref, watch } from 'vue'
import { getSite, paginateProjects } from '@/api'
import Pager from '@/components/chrome/Pager.vue'
import ProjectGrid from '@/components/project/ProjectGrid.vue'

const current = ref(1)
const size = getSite()?.page_size ?? 12
const result = ref<any>({ records: [], total: 0, size, current: 1, pages: 0 })

watch(
  current,
  async (c) => {
    result.value = await paginateProjects({ current: c, size })
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <div class="mb-4">
      <h1 class="text-2xl font-bold">项目</h1>
      <p class="mt-1 text-sm text-muted">实践与工具项目一览</p>
    </div>
    <ProjectGrid :projects="result.records" />
    <Pager
      :current="result.current"
      :pages="result.pages"
      :total="result.total"
      :size="result.size"
      @change="current = $event"
    />
  </div>
</template>
