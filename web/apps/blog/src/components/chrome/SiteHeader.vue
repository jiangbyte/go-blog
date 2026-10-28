<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MobileNavDrawer from '@/components/chrome/MobileNavDrawer.vue'
import SiteNav from '@/components/chrome/SiteNav.vue'
import { useNavScroll } from '@/composables/useNavScroll'

const props = defineProps<{
  /** Transparent over cover hero until scrolled */
  onHero?: boolean
}>()

const route = useRoute()
const router = useRouter()
const menuOpen = ref(false)
const { scrolled } = useNavScroll(56)

const solid = computed(() => !props.onHero || scrolled.value)

function goSearch() {
  menuOpen.value = false
  if (route.name === 'search') {
    const input = document.querySelector<HTMLInputElement>('.search-page__input')
    input?.focus()
    return
  }
  router.push({ name: 'search' })
}

function onKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    goSearch()
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <header
    class="site-header fixed inset-x-0 top-0 z-50 h-[60px] w-full transition-[background,box-shadow,color,backdrop-filter] duration-300"
    :class="[
      onHero && !solid ? 'text-white' : 'text-ink',
      solid ? 'site-header--solid' : 'site-header--ghost',
    ]"
  >
    <SiteNav
      :on-hero="onHero && !solid"
      :menu-open="menuOpen"
      @toggle-menu="menuOpen = !menuOpen"
      @open-search="goSearch"
    />
  </header>

  <MobileNavDrawer :open="menuOpen" @close="menuOpen = false" @open-search="goSearch" />
</template>

<style scoped>
.site-header--ghost {
  background: transparent;
  box-shadow: none;
}

.site-header--solid {
  background: var(--milk-glass-strong);
  border-bottom: 1px solid var(--efu-card-border);
  box-shadow: 0 4px 18px rgba(20, 24, 32, 0.06);
  backdrop-filter: blur(18px) saturate(1.25);
  -webkit-backdrop-filter: blur(18px) saturate(1.25);
}
</style>
