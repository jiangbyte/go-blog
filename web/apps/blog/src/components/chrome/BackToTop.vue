<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import AppIcon from '@/components/chrome/AppIcon.vue'

const visible = ref(false)
const THRESHOLD = 200

function onScroll() {
  visible.value = (window.scrollY || document.documentElement.scrollTop) > THRESHOLD
}

function goTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true })
  onScroll()
})

onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="backtop">
      <button
        v-if="visible"
        type="button"
        class="back-to-top"
        aria-label="回到顶部"
        title="回到顶部"
        @click="goTop"
      >
        <AppIcon name="up" :size="20" />
      </button>
    </Transition>
  </Teleport>
</template>

<style scoped>
.back-to-top {
  position: fixed;
  right: 1rem;
  /* 移动端叠在「目录」FAB 上方，避免遮挡 */
  bottom: calc(4.5rem + env(safe-area-inset-bottom, 0px));
  z-index: 60;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.6rem;
  height: 2.6rem;
  border: 1px solid var(--efu-card-border);
  border-radius: 999px;
  background: var(--milk-glass-strong);
  color: var(--milk-ink);
  box-shadow: 0 8px 24px rgba(20, 24, 32, 0.14);
  backdrop-filter: blur(16px) saturate(1.2);
  -webkit-backdrop-filter: blur(16px) saturate(1.2);
  cursor: pointer;
}

@media (min-width: 1024px) {
  .back-to-top {
    bottom: calc(1.25rem + env(safe-area-inset-bottom, 0px));
  }
}

.backtop-enter-active,
.backtop-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}
.backtop-enter-from,
.backtop-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
</style>
