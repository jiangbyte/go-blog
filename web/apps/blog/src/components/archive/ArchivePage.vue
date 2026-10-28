<script setup lang="ts">
import Pager from '@/components/chrome/Pager.vue'
import PostGrid from '@/components/home/PostGrid.vue'
import { useArchiveFilter } from '@/composables/useArchiveFilter'

const { year, years, result, setYear, setCurrent } = useArchiveFilter()
</script>

<template>
  <section>
    <div class="mb-4">
      <h1 class="text-xl font-bold">归档</h1>
      <p class="page-desc mt-1 text-sm text-muted">按年份筛选全部文章</p>
      <div class="mt-3 flex flex-wrap gap-2">
        <button
          type="button"
          class="rounded-full px-3 py-1 text-sm"
          :class="year === 'all' ? 'chip-ink--on' : 'milk-hover bg-hover'"
          @click="setYear('all')"
        >
          全部
        </button>
        <button
          v-for="y in years"
          :key="y"
          type="button"
          class="rounded-full px-3 py-1 text-sm"
          :class="year === y ? 'chip-ink--on' : 'milk-hover bg-hover'"
          @click="setYear(y)"
        >
          {{ y }}
        </button>
      </div>
    </div>

    <PostGrid :posts="result.records" />
    <p v-if="!result.records.length" class="py-8 text-center text-sm text-muted">暂无文章</p>

    <Pager
      :current="result.current"
      :pages="result.pages"
      :total="result.total"
      :size="result.size"
      @change="setCurrent"
    />
  </section>
</template>
