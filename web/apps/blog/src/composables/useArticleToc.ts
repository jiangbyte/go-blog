import { nextTick, onUnmounted, ref, watch, type Ref } from 'vue'
import { useRoute } from 'vue-router'

const OFFSET = 80

export type TocHeading = { id: string; text: string; level: number }

type TocApi = {
  headings: Ref<TocHeading[]>
  activeId: Ref<string>
  go: (id: string) => void
  updateActive: () => void
}

let shared: TocApi | null = null
let users = 0
let observer: MutationObserver | null = null
let collectTimer: number | undefined
let ticking = false
let routeWatchStop: (() => void) | null = null
let headingsRef: Ref<TocHeading[]> | null = null
let activeIdRef: Ref<string> | null = null

function getTop(el: HTMLElement) {
  return el.getBoundingClientRect().top + window.scrollY
}

function updateActive() {
  if (!headingsRef || !activeIdRef) return
  if (!headingsRef.value.length) {
    activeIdRef.value = ''
    return
  }
  const top = window.scrollY || document.documentElement.scrollTop
  let current = headingsRef.value[0].id
  for (const h of headingsRef.value) {
    const el = document.getElementById(h.id)
    if (!el) continue
    if (top >= getTop(el) - OFFSET) current = h.id
    else break
  }
  if (current) activeIdRef.value = current
}

function onScroll() {
  if (ticking) return
  ticking = true
  requestAnimationFrame(() => {
    updateActive()
    ticking = false
  })
}

function collect() {
  if (!headingsRef) return
  const nodes = document.querySelectorAll('.prose-body h2, .prose-body h3, .prose-body h4')
  const next = [...nodes].map((el, i) => {
    if (!el.id) el.id = `h-${i}`
    const text = (el.textContent || '').replace(/^#\s*/, '').trim()
    return {
      id: el.id,
      text,
      level: Number(el.tagName.slice(1)),
    }
  })
  const prev = headingsRef.value
  const same =
    next.length === prev.length &&
    next.every(
      (h, i) =>
        h.id === prev[i]?.id && h.text === prev[i]?.text && h.level === prev[i]?.level,
    )
  if (!same) headingsRef.value = next
  nextTick(() => updateActive())
}

function scheduleCollect() {
  if (collectTimer) window.clearTimeout(collectTimer)
  collectTimer = window.setTimeout(() => {
    collectTimer = undefined
    collect()
  }, 40)
}

function bindObserver() {
  observer?.disconnect()
  observer = new MutationObserver(() => scheduleCollect())
  observer.observe(document.body, { childList: true, subtree: true })
  collect()
}

function go(id: string) {
  if (!activeIdRef) return
  const el = document.getElementById(id)
  if (!el) return
  activeIdRef.value = id
  const y = Math.max(0, getTop(el) - OFFSET + 1)
  window.scrollTo({ top: y, behavior: 'smooth' })
}

function createShared(): TocApi {
  const headings = ref<TocHeading[]>([])
  const activeId = ref('')
  headingsRef = headings
  activeIdRef = activeId

  window.addEventListener('scroll', onScroll, { passive: true })

  const route = useRoute()
  routeWatchStop?.()
  routeWatchStop = watch(
    () => route.fullPath,
    async () => {
      activeId.value = ''
      headings.value = []
      await nextTick()
      bindObserver()
    },
    { immediate: true },
  )

  return { headings, activeId, go, updateActive }
}

function destroyShared() {
  if (collectTimer) window.clearTimeout(collectTimer)
  collectTimer = undefined
  observer?.disconnect()
  observer = null
  routeWatchStop?.()
  routeWatchStop = null
  window.removeEventListener('scroll', onScroll)
  headingsRef = null
  activeIdRef = null
  shared = null
}

/** 从文章 `.prose-body` 收集目录（单例，侧栏 + 移动端共用） */
export function useArticleToc(): TocApi {
  if (!shared) shared = createShared()
  users += 1

  onUnmounted(() => {
    users -= 1
    if (users <= 0) {
      users = 0
      destroyShared()
    }
  })

  return shared
}
