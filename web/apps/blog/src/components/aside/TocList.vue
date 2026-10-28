<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { useArticleToc } from '@/composables/useArticleToc'

const props = withDefaults(
  defineProps<{
    /** 激活项变化时是否自动滚入可视区 */
    autoScroll?: boolean
  }>(),
  { autoScroll: false },
)
const emit = defineEmits<{ pick: [id: string] }>()

const { headings, activeId, go } = useArticleToc()
const navEl = ref<HTMLElement | null>(null)
const itemEls = new Map<string, HTMLButtonElement>()

function setItemEl(id: string, el: unknown) {
  if (el instanceof HTMLButtonElement) itemEls.set(id, el)
  else itemEls.delete(id)
}

function onClick(id: string) {
  go(id)
  emit('pick', id)
}

watch(activeId, (id) => {
  if (!props.autoScroll || !id) return
  nextTick(() => {
    const btn = itemEls.get(id)
    const nav = navEl.value
    if (!btn || !nav) return
    const btnRect = btn.getBoundingClientRect()
    const navRect = nav.getBoundingClientRect()
    if (btnRect.bottom > navRect.bottom - 12) {
      nav.scrollTop += btnRect.bottom - navRect.bottom + 28
    } else if (btnRect.top < navRect.top + 12) {
      nav.scrollTop -= navRect.top - btnRect.top + 28
    }
  })
})
</script>

<template>
  <nav ref="navEl" class="toc-list">
    <button
      v-for="h in headings"
      :key="h.id"
      :ref="(el) => setItemEl(h.id, el)"
      type="button"
      class="toc-link"
      :class="[`is-h${h.level}`, h.id === activeId ? 'is-active' : '']"
      :aria-current="h.id === activeId ? 'location' : undefined"
      @click="onClick(h.id)"
    >
      {{ h.text }}
    </button>
  </nav>
</template>

<style scoped>
.toc-list {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  color: var(--milk-ink-soft);
  font-size: 0.875rem;
}

.toc-link {
  display: block;
  width: 100%;
  margin: 0;
  padding-top: 0.4rem;
  padding-right: 0.5rem;
  padding-bottom: 0.4rem;
  padding-left: 0.5rem;
  border: 0;
  border-radius: 0.375rem;
  background: transparent;
  color: var(--milk-ink-soft);
  font: inherit;
  font-size: 0.875rem;
  font-weight: 400;
  line-height: 1.35;
  text-align: left;
  cursor: pointer;
  transition:
    background 0.15s ease,
    color 0.15s ease;
}

/* 层级只靠缩进 */
.toc-link.is-h2 {
  padding-left: 0.5rem;
}
.toc-link.is-h3 {
  padding-left: 1.35rem;
}
.toc-link.is-h4 {
  padding-left: 2.2rem;
}

.toc-link:hover,
.toc-link:focus-visible {
  background: var(--efu-ahoverbg);
  color: var(--milk-ink);
}

.toc-link.is-active {
  background: var(--efu-secondbg);
  color: var(--milk-ink);
  font-weight: 600;
}
</style>
