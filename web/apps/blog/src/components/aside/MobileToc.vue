<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import TocList from '@/components/aside/TocList.vue'
import AppIcon from '@/components/chrome/AppIcon.vue'
import { useArticleToc } from '@/composables/useArticleToc'

const { headings } = useArticleToc()
const open = ref(false)
const route = useRoute()

function close() {
  open.value = false
}

watch(
  () => route.fullPath,
  () => close(),
)

watch(open, (v) => {
  document.body.style.overflow = v ? 'hidden' : ''
})

onUnmounted(() => {
  document.body.style.overflow = ''
})
</script>

<template>
  <div v-if="headings.length" class="mobile-toc lg:hidden">
    <button
      type="button"
      class="mobile-toc__fab"
      aria-label="打开目录"
      title="目录"
      @click="open = true"
    >
      <AppIcon name="book" :size="20" />
    </button>

    <Teleport to="body">
      <div
        class="mobile-toc__layer"
        :class="{ 'is-open': open }"
        role="dialog"
        aria-modal="true"
        aria-label="文章目录"
      >
        <button type="button" class="mobile-toc__mask" aria-label="关闭目录" @click="close" />
        <div class="mobile-toc__sheet">
          <header class="mobile-toc__head">
            <span class="font-semibold">目录</span>
            <button
              type="button"
              class="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-hover"
              aria-label="关闭"
              @click="close"
            >
              <AppIcon name="close" :size="18" />
            </button>
          </header>
          <TocList class="mobile-toc__list" @pick="close" />
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.mobile-toc__fab {
  position: fixed;
  right: 1rem;
  bottom: calc(1rem + env(safe-area-inset-bottom, 0px));
  z-index: 60;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.6rem;
  height: 2.6rem;
  padding: 0;
  border: 1px solid var(--efu-card-border);
  border-radius: 999px;
  background: var(--milk-glass-strong);
  color: var(--milk-ink);
  box-shadow: 0 8px 24px rgba(20, 24, 32, 0.14);
  backdrop-filter: blur(16px) saturate(1.2);
  -webkit-backdrop-filter: blur(16px) saturate(1.2);
  cursor: pointer;
}

.mobile-toc__layer {
  position: fixed;
  inset: 0;
  z-index: 85;
  pointer-events: none;
  visibility: hidden;
}

.mobile-toc__layer.is-open {
  pointer-events: auto;
  visibility: visible;
}

.mobile-toc__mask {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(20, 24, 32, 0.42);
  opacity: 0;
  transition: opacity 0.25s ease;
  cursor: pointer;
}

.mobile-toc__layer.is-open .mobile-toc__mask {
  opacity: 1;
}

.mobile-toc__sheet {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  max-height: min(72vh, 560px);
  border-radius: 16px 16px 0 0;
  /* 与 PC glass-card 同底，激活态 secondbg 对比度一致 */
  background: var(--efu-card-bg);
  color: var(--milk-ink);
  box-shadow: 0 -8px 28px rgba(20, 24, 32, 0.14);
  backdrop-filter: blur(20px) saturate(1.25);
  -webkit-backdrop-filter: blur(20px) saturate(1.25);
  transform: translateY(105%);
  transition: transform 0.28s ease;
  padding-bottom: env(safe-area-inset-bottom, 0px);
}

.mobile-toc__layer.is-open .mobile-toc__sheet {
  transform: translateY(0);
}

.mobile-toc__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--efu-card-border);
  flex-shrink: 0;
}

.mobile-toc__list {
  flex: 1;
  padding: 0.5rem;
}
</style>
