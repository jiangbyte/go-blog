<script setup lang="ts">
import AuthorCard from '@/components/aside/AuthorCard.vue'
import RecentPostsWidget from '@/components/aside/RecentPostsWidget.vue'
import TocWidget from '@/components/aside/TocWidget.vue'

defineProps<{ showToc?: boolean }>()
</script>

<template>
  <!-- display 由外层 class（如 hidden lg:flex）控制，避免 scoped 覆盖掉 hidden -->
  <aside class="site-aside w-full flex-col gap-3">
    <AuthorCard />
    <div class="sticky-stack sticky top-[68px] flex w-full flex-col gap-3">
      <TocWidget v-if="showToc" />
      <RecentPostsWidget />
    </div>
  </aside>
</template>
