<script setup lang="ts">
import SiteFooter from '@/components/chrome/SiteFooter.vue'
import SiteHeader from '@/components/chrome/SiteHeader.vue'
</script>

<template>
  <div class="flex min-h-screen flex-col bg-surface text-ink">
    <SiteHeader />
    <div class="page-gutter flex flex-1 items-center pb-10 pt-[76px]">
      <main class="page-shell w-full min-w-0">
        <RouterView />
      </main>
    </div>
    <SiteFooter />
  </div>
</template>
