<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getHomePosts, getSite } from '@/api'
import AppIcon from '@/components/chrome/AppIcon.vue'
import { coverOf } from '@/utils/markdown'

const site = getSite()
const recent = ref<any[]>([])

onMounted(async () => {
  const { list } = await getHomePosts({})
  recent.value = list.slice(0, 5)
})
</script>

<template>
  <section class="glass-card p-2.5">
    <div class="mb-1 flex items-center gap-2 px-1.5 text-sm font-semibold">
      <AppIcon name="book" :size="16" />
      <span>最近文章</span>
    </div>
    <div class="flex flex-col">
      <RouterLink
        v-for="(post, i) in recent"
        :key="post.uid"
        :to="`/p/${post.uid}`"
        class="aside-item flex min-w-0 items-center gap-2.5 rounded-xl border border-transparent px-1.5 py-1.5 transition hover:border-card-border hover:bg-hover"
      >
        <div class="h-[2.1em] w-[2.1em] shrink-0 overflow-hidden rounded-lg border border-card-border">
          <img
            :src="coverOf(post.cover, site.default_covers, i)"
            alt=""
            class="h-full w-full object-cover"
            loading="lazy"
          />
        </div>
        <div class="min-w-0 flex-1 overflow-hidden">
          <div class="line-clamp-1 text-[95%] font-bold">{{ post.title }}</div>
          <div class="mt-0.5 text-xs font-semibold text-muted">
            {{ post.categories?.[0]?.name ?? post.categories?.[0] ?? '未分类' }}
          </div>
        </div>
      </RouterLink>
    </div>
  </section>
</template>
