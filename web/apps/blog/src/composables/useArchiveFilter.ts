import { ref, watch } from 'vue'
import { getSite, listArchiveYears, paginateArchive } from '@/api'

export function useArchiveFilter(initialYear = 'all') {
  const year = ref(initialYear)
  const current = ref(1)
  const years = ref<string[]>([])
  const size = getSite()?.page_size ?? 12
  const result = ref<any>({
    records: [],
    total: 0,
    size,
    current: 1,
    pages: 0,
  })

  async function load() {
    const [y, pageResult] = await Promise.all([
      listArchiveYears({}),
      paginateArchive({ current: current.value, size, year: year.value }),
    ])
    years.value = y.list
    result.value = pageResult
  }

  watch([year, current], load, { immediate: true })

  function setYear(next: string) {
    year.value = next
    current.value = 1
  }

  function setCurrent(next: number) {
    current.value = next
  }

  return { year, current, years, result, setYear, setCurrent }
}
