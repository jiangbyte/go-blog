<script setup lang="ts">
import BackToTop from '@/components/chrome/BackToTop.vue'
import SiteFooter from '@/components/chrome/SiteFooter.vue'
import SiteHeader from '@/components/chrome/SiteHeader.vue'
</script>

<template>
  <div class="flex min-h-screen flex-col bg-surface text-ink">
    <SiteHeader />
    <div class="page-gutter flex-1 pb-8 pt-[76px]">
      <main class="page-shell min-w-0 overflow-x-clip">
        <RouterView />
      </main>
    </div>
    <BackToTop />
    <SiteFooter />
  </div>
</template>
