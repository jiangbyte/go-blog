<script setup lang="ts">
import { ref, watch } from 'vue'
import { getSite, paginatePosts } from '@/api'
import Pager from '@/components/chrome/Pager.vue'
import PostGrid from '@/components/home/PostGrid.vue'

const current = ref(1)
const size = getSite()?.page_size ?? 12
const result = ref<any>({ records: [], total: 0, size, current: 1, pages: 0 })

watch(
  current,
  async (c) => {
    result.value = await paginatePosts({ current: c, size })
  },
  { immediate: true },
)
</script>

<template>
  <div class="flex flex-col gap-5">
    <PostGrid :posts="result.records" />
    <Pager
      :current="result.current"
      :pages="result.pages"
      :total="result.total"
      :size="result.size"
      @change="current = $event"
    />
  </div>
</template>
