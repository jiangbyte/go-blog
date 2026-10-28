<script setup lang="ts">
import { getSite } from '@/api'
import AppIcon from '@/components/chrome/AppIcon.vue'
import { computed, ref } from 'vue'

defineProps<{
  onHero?: boolean
  menuOpen?: boolean
}>()
defineEmits<{ toggleMenu: []; openSearch: [] }>()

const site = computed(() => getSite() || {})
const openGroup = ref<string | null>(null)

function isExternal(item: { external?: unknown; href?: string }) {
  if (item.external === true || item.external === 'true' || item.external === 1) return true
  return /^https?:\/\//i.test(String(item.href || ''))
}

/** 站内相对路径（非协议相对 //）走 RouterLink。 */
function isInternalPath(url: string | undefined) {
  const u = String(url || '')
  return u.startsWith('/') && !u.startsWith('//')
}
</script>

<template>
  <nav class="page-gutter flex h-[60px] w-full items-center justify-center">
    <div class="page-shell relative flex items-center gap-3">
      <div class="relative flex items-center gap-2">
        <div
          class="relative hidden md:block"
          @mouseenter="openGroup = 'links'"
          @mouseleave="openGroup = null"
        >
          <button
            type="button"
            class="nav-chip inline-flex h-9 w-9 items-center justify-center rounded-lg"
            :class="onHero ? 'nav-chip--ghost' : 'nav-chip--solid'"
            aria-label="快捷链接"
            :aria-expanded="openGroup === 'links'"
          >
            <AppIcon name="apps" :size="20" />
          </button>
          <!-- pt 作为悬停桥，避免一级→二级空隙导致菜单收起 -->
          <div
            v-if="openGroup === 'links'"
            class="nav-flyout absolute left-0 top-full z-50 w-52 pt-2"
          >
            <div
              class="rounded-xl border border-card-border bg-glass-strong p-2 text-ink shadow-lg backdrop-blur-xl"
            >
              <p class="px-2 py-1 text-xs text-muted">链接</p>
              <template v-for="link in site.nav_group" :key="link.name">
                <RouterLink
                  v-if="isInternalPath(link.url)"
                  :to="link.url"
                  class="flex items-center gap-2 rounded-lg px-2 py-2 text-sm hover:bg-hover"
                >
                  <AppIcon v-if="link.icon" :name="link.icon" :size="16" />
                  {{ link.name }}
                </RouterLink>
                <a
                  v-else
                  :href="link.url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="flex items-center gap-2 rounded-lg px-2 py-2 text-sm hover:bg-hover"
                >
                  <AppIcon v-if="link.icon" :name="link.icon" :size="16" />
                  {{ link.name }}
                </a>
              </template>
            </div>
          </div>
        </div>
        <RouterLink
          to="/"
          class="nav-chip rounded-lg px-2.5 py-1 text-lg font-bold tracking-wide"
          :class="onHero ? 'nav-chip--ghost' : 'nav-chip--solid'"
        >
          {{ site.site_name }}
        </RouterLink>
      </div>

      <div class="ml-4 hidden flex-1 items-center gap-1 md:flex">
        <div
          v-for="item in site.nav"
          :key="item.name"
          class="relative"
          @mouseenter="openGroup = item.name"
          @mouseleave="openGroup = null"
        >
          <a
            v-if="isExternal(item) && item.href"
            :href="item.href"
            target="_blank"
            rel="noopener noreferrer"
            class="nav-chip inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm"
            :class="onHero ? 'nav-chip--ghost' : 'nav-chip--solid'"
          >
            <AppIcon v-if="item.icon" :name="item.icon" :size="16" />
            {{ item.name }}
          </a>
          <RouterLink
            v-else-if="item.href && !item.children"
            :to="item.href"
            class="nav-chip inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm"
            :class="onHero ? 'nav-chip--ghost' : 'nav-chip--solid'"
          >
            <AppIcon v-if="item.icon" :name="item.icon" :size="16" />
            {{ item.name }}
          </RouterLink>
          <button
            v-else
            type="button"
            class="nav-chip inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm"
            :class="onHero ? 'nav-chip--ghost' : 'nav-chip--solid'"
            :aria-expanded="openGroup === item.name"
          >
            {{ item.name }}
            <AppIcon name="down" :size="14" />
          </button>
          <div
            v-if="item.children && openGroup === item.name"
            class="nav-flyout absolute left-0 top-full z-50 min-w-[168px] pt-2"
          >
            <div
              class="rounded-xl border border-card-border bg-glass-strong p-2 text-ink shadow-lg backdrop-blur-xl"
            >
              <RouterLink
                v-for="child in item.children"
                :key="child.name"
                :to="child.href || '/'"
                class="flex items-center gap-2 rounded-lg px-3 py-2 text-sm hover:bg-hover"
              >
                <AppIcon v-if="child.icon" :name="child.icon" :size="16" />
                {{ child.name }}
              </RouterLink>
            </div>
          </div>
        </div>
      </div>

      <div class="ml-auto flex items-center gap-1">
        <button
          type="button"
          class="nav-chip inline-flex h-9 w-9 items-center justify-center rounded-lg"
          :class="onHero ? 'nav-chip--ghost' : 'nav-chip--solid'"
          aria-label="搜索"
          title="搜索 (Ctrl/⌘ K)"
          @click="$emit('openSearch')"
        >
          <AppIcon name="search" :size="20" />
        </button>
        <button
          type="button"
          class="nav-chip inline-flex h-9 w-9 items-center justify-center rounded-lg md:hidden"
          :class="onHero ? 'nav-chip--ghost' : 'nav-chip--solid'"
          aria-label="菜单"
          @click="$emit('toggleMenu')"
        >
          <AppIcon :name="menuOpen ? 'close' : 'menu'" :size="22" />
        </button>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.nav-chip {
  transition:
    background 0.2s ease,
    color 0.2s ease,
    box-shadow 0.2s ease;
}

.nav-chip--solid:hover,
.nav-chip--solid:focus-visible {
  background: var(--milk-glass);
  color: var(--milk-ink);
  box-shadow: 0 2px 10px rgba(20, 24, 32, 0.05);
  backdrop-filter: blur(12px) saturate(1.2);
  -webkit-backdrop-filter: blur(12px) saturate(1.2);
}

.nav-chip--ghost:hover,
.nav-chip--ghost:focus-visible {
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
}
</style>
