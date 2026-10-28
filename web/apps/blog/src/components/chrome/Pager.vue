<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { usePagination } from '@/composables/usePagination'

const props = withDefaults(
  defineProps<{
    current: number
    pages: number
    total: number
    size?: number
  }>(),
  { size: 12 },
)

const emit = defineEmits<{ change: [page: number] }>()

const currentRef = computed(() => props.current)
const pagesRef = computed(() => props.pages)
const { pages: pageList, hasPrev, hasNext } = usePagination(currentRef, pagesRef)

const jump = ref(String(props.current || 1))
watch(
  () => props.current,
  (v) => {
    jump.value = String(v)
  },
)

const summary = computed(() => {
  const total = Math.max(0, Number(props.total) || 0)
  const pages = Math.max(0, Number(props.pages) || 0)
  const cur = Math.max(1, Number(props.current) || 1)
  const size = Math.max(1, Number(props.size) || 12)
  if (total === 0) return '共 0 条'
  const from = (cur - 1) * size + 1
  const to = Math.min(cur * size, total)
  return `共 ${total} 条 · 第 ${from}-${to} 条 · ${cur}/${pages || 1} 页`
})

function go(page: number) {
  const total = Math.max(0, Number(props.pages) || 0)
  if (total <= 0) return
  const next = Math.min(Math.max(1, Math.floor(page)), total)
  if (next === props.current) return
  emit('change', next)
}

function submitJump() {
  const n = Number(jump.value)
  if (!Number.isFinite(n)) {
    jump.value = String(props.current)
    return
  }
  go(n)
}
</script>

<template>
  <div v-if="pages > 0" class="pager">
    <div class="pager__summary">{{ summary }}</div>
    <div class="pager__row">
      <button
        type="button"
        class="pager__btn"
        :disabled="!hasPrev"
        @click="go(current - 1)"
      >
        上一页
      </button>

      <template v-for="(p, i) in pageList" :key="`${p}-${i}`">
        <span v-if="p === '…'" class="pager__ellipsis">…</span>
        <button
          v-else
          type="button"
          class="pager__num"
          :class="{ 'is-active': p === current }"
          @click="go(p)"
        >
          {{ p }}
        </button>
      </template>

      <button
        type="button"
        class="pager__btn"
        :disabled="!hasNext"
        @click="go(current + 1)"
      >
        下一页
      </button>

      <form class="pager__jump" @submit.prevent="submitJump">
        <span class="pager__jump-label">跳至</span>
        <input
          v-model="jump"
          type="number"
          min="1"
          :max="Math.max(1, pages)"
          class="pager__jump-input"
          inputmode="numeric"
        />
        <span class="pager__jump-label">页</span>
        <button type="submit" class="pager__btn pager__btn--go">确定</button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.pager {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.65rem;
  margin-top: 1.25rem;
}

.pager__summary {
  color: var(--efu-secondtext);
  font-size: 0.82rem;
}

.pager__row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
}

.pager__btn,
.pager__num {
  border: 0;
  border-radius: 999px;
  background: var(--efu-card-bg);
  color: var(--milk-ink);
  box-shadow: 0 1px 4px rgba(20, 24, 32, 0.05);
  font: inherit;
  font-size: 0.875rem;
  cursor: pointer;
}

.pager__btn {
  padding: 0.45rem 1rem;
}

.pager__btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.pager__num {
  min-width: 2.4rem;
  height: 2.4rem;
  padding: 0 0.55rem;
}

.pager__num.is-active,
.pager__num.is-active:hover {
  background: var(--milk-ink);
  color: #fff;
}

.pager__num:hover,
.pager__btn:not(:disabled):hover {
  background: var(--efu-ahoverbg);
}

.pager__ellipsis {
  color: var(--efu-secondtext);
  padding: 0 0.15rem;
  user-select: none;
}

.pager__jump {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  margin-left: 0.35rem;
}

.pager__jump-label {
  color: var(--efu-secondtext);
  font-size: 0.82rem;
}

.pager__jump-input {
  width: 3.2rem;
  height: 2.2rem;
  border: 1px solid var(--efu-card-border);
  border-radius: 8px;
  background: var(--efu-card-bg);
  color: inherit;
  text-align: center;
  font: inherit;
  font-size: 0.875rem;
  outline: none;
}

.pager__jump-input:focus {
  border-color: color-mix(in srgb, var(--milk-ink) 35%, var(--efu-card-border));
}

.pager__jump-input::-webkit-outer-spin-button,
.pager__jump-input::-webkit-inner-spin-button {
  appearance: none;
  margin: 0;
}

.pager__jump-input[type='number'] {
  appearance: textfield;
  -moz-appearance: textfield;
}

.pager__btn--go {
  padding: 0.4rem 0.85rem;
}
</style>
