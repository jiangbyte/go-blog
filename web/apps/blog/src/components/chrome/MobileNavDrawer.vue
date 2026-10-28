<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getSite } from '@/api'
import AppIcon from '@/components/chrome/AppIcon.vue'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: []; openSearch: [] }>()

const site = computed(() => getSite() || {})
const brandIcon = computed(
  () =>
    String(site.value.author?.avatar_url || '').trim() ||
    String(site.value.icon || '').trim(),
)
const route = useRoute()
const expanded = ref<Record<string, boolean>>({})

watch(
  () => site.value.nav,
  (nav) => {
    const next: Record<string, boolean> = { ...expanded.value }
    for (const n of nav || []) {
      if (n?.children?.length && next[n.name] === undefined) next[n.name] = true
    }
    expanded.value = next
  },
  { immediate: true, deep: true },
)

function toggle(name: string) {
  expanded.value[name] = !expanded.value[name]
}

function close() {
  emit('close')
}

function isInternalPath(url: string | undefined) {
  const u = String(url || '')
  return u.startsWith('/') && !u.startsWith('//')
}

function isExternal(item: { external?: unknown; href?: string }) {
  if (item.external === true || item.external === 'true' || item.external === 1) return true
  return /^https?:\/\//i.test(String(item.href || ''))
}

/** 与目录一致的激活判断；`/` 仅精确匹配 */
function isNavActive(href?: string) {
  if (!href) return false
  const path = href.split('?')[0] || '/'
  if (path === '/') return route.path === '/'
  return route.path === path || route.path.startsWith(`${path}/`)
}

watch(
  () => props.open,
  (v) => {
    document.body.style.overflow = v ? 'hidden' : ''
  },
)

watch(
  () => route.fullPath,
  () => {
    if (props.open) close()
  },
)

onUnmounted(() => {
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <div
      class="mobile-drawer md:hidden"
      :class="{ 'is-visible': open }"
      role="dialog"
      aria-modal="true"
      aria-label="站点导航"
    >
      <button type="button" class="mobile-drawer__mask" aria-label="关闭菜单" @click="close" />

      <aside class="mobile-drawer__panel">
        <header class="mobile-drawer__head">
          <RouterLink to="/" class="mobile-drawer__brand" @click="close">
            <img
              v-if="brandIcon"
              :src="brandIcon"
              alt=""
              class="h-8 w-8 rounded-full object-cover"
              referrerpolicy="no-referrer"
            />
            <span class="truncate font-bold">{{ site.site_name }}</span>
          </RouterLink>
          <button
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded-lg hover:bg-hover"
            aria-label="关闭"
            @click="close"
          >
            <AppIcon name="close" :size="20" />
          </button>
        </header>

        <nav class="mobile-drawer__tree">
          <button
            type="button"
            class="mobile-drawer__node mobile-drawer__node--leaf mobile-drawer__node--root"
            @click="emit('openSearch'); close()"
          >
            <span class="mobile-drawer__spacer" />
            <AppIcon name="search" :size="16" />
            <span>搜索</span>
            <kbd class="mobile-drawer__kbd">K</kbd>
          </button>

          <template v-for="item in site.nav || []" :key="item.name">
            <div v-if="item.children?.length" class="mobile-drawer__group">
              <button
                type="button"
                class="mobile-drawer__node mobile-drawer__node--branch"
                :aria-expanded="!!expanded[item.name]"
                @click="toggle(item.name)"
              >
                <AppIcon
                  name="down"
                  :size="14"
                  class="mobile-drawer__caret"
                  :class="{ 'is-open': expanded[item.name] }"
                />
                <span class="font-semibold">{{ item.name }}</span>
              </button>
              <div v-show="expanded[item.name]" class="mobile-drawer__children">
                <template v-for="child in item.children" :key="child.name">
                  <a
                    v-if="isExternal(child) && child.href"
                    :href="child.href"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="mobile-drawer__node mobile-drawer__node--leaf"
                  >
                    <AppIcon v-if="child.icon" :name="child.icon" :size="16" />
                    <span>{{ child.name }}</span>
                  </a>
                  <RouterLink
                    v-else
                    :to="child.href || '/'"
                    class="mobile-drawer__node mobile-drawer__node--leaf"
                    :class="{ 'is-active': isNavActive(child.href) }"
                    @click="close"
                  >
                    <AppIcon v-if="child.icon" :name="child.icon" :size="16" />
                    <span>{{ child.name }}</span>
                  </RouterLink>
                </template>
              </div>
            </div>

            <a
              v-else-if="isExternal(item) && item.href"
              :href="item.href"
              target="_blank"
              rel="noopener noreferrer"
              class="mobile-drawer__node mobile-drawer__node--leaf mobile-drawer__node--root"
            >
              <span class="mobile-drawer__spacer" />
              <AppIcon v-if="item.icon" :name="item.icon" :size="16" />
              <span>{{ item.name }}</span>
            </a>
            <RouterLink
              v-else
              :to="item.href || '/'"
              class="mobile-drawer__node mobile-drawer__node--leaf mobile-drawer__node--root"
              :class="{ 'is-active': isNavActive(item.href) }"
              @click="close"
            >
              <span class="mobile-drawer__spacer" />
              <AppIcon v-if="item.icon" :name="item.icon" :size="16" />
              <span>{{ item.name }}</span>
            </RouterLink>
          </template>

          <div v-if="site.nav_group?.length" class="mt-3 border-t border-card-border pt-3">
            <div class="mobile-drawer__section-label">链接</div>
            <template v-for="link in site.nav_group" :key="link.name">
              <RouterLink
                v-if="isInternalPath(link.url)"
                :to="link.url"
                class="mobile-drawer__node mobile-drawer__node--leaf mobile-drawer__node--root"
                @click="close"
              >
                <span class="mobile-drawer__spacer" />
                <AppIcon v-if="link.icon" :name="link.icon" :size="16" />
                <span>{{ link.name }}</span>
              </RouterLink>
              <a
                v-else
                :href="link.url"
                target="_blank"
                rel="noopener noreferrer"
                class="mobile-drawer__node mobile-drawer__node--leaf mobile-drawer__node--root"
              >
                <span class="mobile-drawer__spacer" />
                <AppIcon v-if="link.icon" :name="link.icon" :size="16" />
                <span>{{ link.name }}</span>
              </a>
            </template>
          </div>
        </nav>
      </aside>
    </div>
  </Teleport>
</template>

<style scoped>
.mobile-drawer {
  position: fixed;
  inset: 0;
  z-index: 80;
  pointer-events: none;
  visibility: hidden;
}

.mobile-drawer.is-visible {
  pointer-events: auto;
  visibility: visible;
}

.mobile-drawer__mask {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(20, 24, 32, 0.42);
  opacity: 0;
  transition: opacity 0.28s ease;
  cursor: pointer;
}

.mobile-drawer.is-visible .mobile-drawer__mask {
  opacity: 1;
}

.mobile-drawer__panel {
  position: absolute;
  top: 0;
  right: 0;
  left: auto;
  display: flex;
  flex-direction: column;
  width: min(82vw, 300px);
  height: 100%;
  background: var(--milk-glass-strong);
  color: var(--milk-ink);
  box-shadow: -8px 0 28px rgba(20, 24, 32, 0.14);
  backdrop-filter: blur(20px) saturate(1.25);
  -webkit-backdrop-filter: blur(20px) saturate(1.25);
  transform: translateX(105%);
  transition: transform 0.28s ease;
}

.mobile-drawer.is-visible .mobile-drawer__panel {
  transform: translateX(0);
}

.mobile-drawer__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  height: 60px;
  padding: 0 0.85rem 0 1rem;
  border-bottom: 1px solid var(--efu-card-border);
  flex-shrink: 0;
}

.mobile-drawer__brand {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  min-width: 0;
  color: inherit;
  text-decoration: none;
}

.mobile-drawer__tree {
  flex: 1;
  overflow-y: auto;
  padding: 0.65rem 0.55rem 1.25rem;
  overscroll-behavior: contain;
}

.mobile-drawer__section-label {
  padding: 0.25rem 0.75rem 0.4rem;
  font-size: 0.75rem;
  color: var(--efu-secondtext);
}

.mobile-drawer__node {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  width: 100%;
  margin: 0;
  padding: 0.65rem 0.75rem;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
  text-decoration: none;
}

.mobile-drawer__node:hover,
.mobile-drawer__node:focus-visible {
  background: var(--efu-ahoverbg);
}

.mobile-drawer__node--branch {
  font-size: 0.95rem;
}

.mobile-drawer__caret {
  transition: transform 0.2s ease;
  transform: rotate(-90deg);
  opacity: 0.65;
}

.mobile-drawer__caret.is-open {
  transform: rotate(0deg);
}

.mobile-drawer__spacer {
  width: 14px;
  flex-shrink: 0;
}

.mobile-drawer__children {
  padding-left: 0.2rem;
  margin: 0.1rem 0 0.35rem 0.95rem;
  border-left: 1px solid var(--efu-card-border);
}

.mobile-drawer__node--leaf {
  font-size: 0.92rem;
  color: var(--milk-ink-soft);
}

.mobile-drawer__node--root {
  color: var(--milk-ink);
}

/* 与目录激活一致：圆角灰底 + 加粗 */
.mobile-drawer__node--leaf.is-active {
  background: var(--efu-secondbg);
  color: var(--milk-ink);
  font-weight: 600;
}

.mobile-drawer__kbd {
  margin-left: auto;
  padding: 0.05em 0.4em;
  border: 1px solid var(--efu-card-border);
  border-radius: 4px;
  background: var(--efu-secondbg);
  color: var(--efu-secondtext);
  font-size: 0.72rem;
  font-family: inherit;
}
</style>
