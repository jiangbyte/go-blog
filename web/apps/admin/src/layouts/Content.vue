<!-- Author: go-blog -->

<script setup lang="ts">
import { useAppStore } from '@/stores'

const appStore = useAppStore()
</script>

<template>
  <n-el class="layout-page h-full min-h-0 min-w-0 w-full overflow-auto p-8px">
    <router-view v-slot="{ Component, route }">
      <transition name="fade-slide" mode="out-in">
        <!-- 不使用 keep-alive：每次进入页面重新挂载并拉最新数据 -->
        <component
          :is="Component"
          v-if="appStore.loadFlag"
          :key="`${route.path}::${String(route.query.uid ?? route.query.id ?? route.params.id ?? '')}`"
        />
      </transition>
    </router-view>
  </n-el>
</template>

<style scoped>
.layout-page {
  background: transparent;
  max-width: 100%;
  box-sizing: border-box;
}
</style>
