<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getHomePosts } from '@/api'

const recent = ref<any[]>([])

onMounted(async () => {
  const { list } = await getHomePosts({})
  recent.value = list.slice(0, 3)
})
</script>

<template>
  <div class="glass-card w-full overflow-hidden p-0 md:flex">
    <div class="flex min-h-[180px] items-center justify-center bg-ink px-6 py-10 text-5xl font-bold text-white md:w-2/5">
      404
    </div>
    <div class="flex flex-1 flex-col justify-center p-6 md:p-8">
      <h1 class="text-2xl font-bold">页面走丢了</h1>
      <p class="mt-2 text-muted">链接可能已失效，或页面不存在。</p>
      <RouterLink
        to="/"
        class="chip-ink--on mt-5 inline-flex w-fit rounded-full px-5 py-2 text-sm"
      >
        返回首页
      </RouterLink>
      <div v-if="recent.length" class="mt-6 border-t border-card-border pt-4">
        <p class="mb-2 text-sm font-semibold">看看这些</p>
        <RouterLink
          v-for="p in recent"
          :key="p.uid"
          :to="`/p/${p.uid}`"
          class="mb-1 block line-clamp-1 text-sm text-muted hover:text-ink"
        >
          {{ p.title }}
        </RouterLink>
      </div>
    </div>
  </div>
</template>
