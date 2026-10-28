import { onMounted, onUnmounted, ref } from 'vue'

/** True after scrolling past the hero / a small threshold. */
export function useNavScroll(threshold = 48) {
  const scrolled = ref(false)

  function onScroll() {
    scrolled.value = window.scrollY > threshold
  }

  onMounted(() => {
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
  })
  onUnmounted(() => {
    window.removeEventListener('scroll', onScroll)
  })

  return { scrolled }
}
