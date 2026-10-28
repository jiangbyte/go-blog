<script setup lang="ts">
import BackToTop from '@/components/chrome/BackToTop.vue'
import SiteAside from '@/components/chrome/SiteAside.vue'
import SiteFooter from '@/components/chrome/SiteFooter.vue'
import SiteHeader from '@/components/chrome/SiteHeader.vue'
</script>

<template>
  <div class="flex min-h-screen flex-col bg-surface text-ink">
    <SiteHeader />
    <div class="page-gutter flex-1 pb-8 pt-[76px]">
      <div
        class="page-shell layout-shell grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_300px] lg:gap-[15px]"
      >
        <main class="min-w-0 overflow-x-clip">
          <RouterView />
        </main>
        <SiteAside class="hidden lg:flex" />
      </div>
    </div>
    <BackToTop />
    <SiteFooter />
  </div>
</template>
