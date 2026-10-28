import { computed, type Ref } from 'vue'

/**
 * UI 分页控件，字段语义对齐 MyBatis-Plus IPage：
 * - `current` ← IPage.current
 * - `pages` ← IPage.pages（为 0 时不渲染页码）
 */
export function usePagination(current: Ref<number>, pages: Ref<number>, windowSize = 2) {
  /** 带省略的页码序列，元素为页码 number 或 '…' */
  const pageList = computed(() => {
    const total = Math.max(0, Number(pages.value) || 0)
    const cur = Math.min(Math.max(1, Number(current.value) || 1), Math.max(1, total))
    if (total === 0) return [] as Array<number | '…'>
    if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)

    const set = new Set<number>()
    set.add(1)
    set.add(total)
    for (let i = cur - windowSize; i <= cur + windowSize; i++) {
      if (i >= 1 && i <= total) set.add(i)
    }
    const sorted = [...set].sort((a, b) => a - b)
    const out: Array<number | '…'> = []
    for (let i = 0; i < sorted.length; i++) {
      if (i > 0 && sorted[i] - sorted[i - 1] > 1) out.push('…')
      out.push(sorted[i])
    }
    return out
  })

  const hasPrev = computed(() => current.value > 1)
  const hasNext = computed(() => {
    const total = Math.max(0, Number(pages.value) || 0)
    return total > 0 && current.value < total
  })

  return { pages: pageList, hasPrev, hasNext }
}
