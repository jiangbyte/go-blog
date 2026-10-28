<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getSite, listCategories, paginateArchive } from '@/api'
import CategoryBar from '@/components/chrome/CategoryBar.vue'
import Pager from '@/components/chrome/Pager.vue'
import PostGrid from '@/components/home/PostGrid.vue'
import { applySeo, pageAbsoluteUrl } from '@/composables/useSeo'

const route = useRoute()
const uid = computed(() => String(route.params.uid || ''))
const name = ref('')
const current = ref(1)
const size = getSite()?.page_size ?? 12
const result = ref<any>({ records: [], total: 0, size, current: 1, pages: 0 })

watch(
  uid,
  async (id) => {
    current.value = 1
    const { list } = await listCategories({})
    name.value = list.find((c: any) => c.uid === id)?.name || id
    const siteName = String(getSite()?.site_name || getSite()?.title || 'Demo Blog')
    applySeo({
      title: `${name.value} · 分类 · ${siteName}`,
      description: `分类「${name.value}」下的文章`,
      url: pageAbsoluteUrl(`/categories/${id}`),
    })
  },
  { immediate: true },
)

watch(
  [uid, current],
  async ([id, c]) => {
    if (!id) return
    result.value = await paginateArchive({
      current: c,
      size,
      by: 'category',
      uid: id,
    })
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <CategoryBar :active="name" />
    <div class="mb-4">
      <h1 class="text-xl font-bold">{{ name }}</h1>
      <p class="page-desc mt-1 text-sm text-muted">分类「{{ name }}」下的文章</p>
    </div>
    <PostGrid :posts="result.records" />
    <p v-if="!result.records.length" class="py-8 text-center text-sm text-muted">暂无文章</p>
    <Pager
      :current="result.current"
      :pages="result.pages"
      :total="result.total"
      :size="result.size"
      @change="current = $event"
    />
  </div>
</template>
