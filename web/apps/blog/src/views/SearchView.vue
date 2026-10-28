<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getSite, searchSite } from '@/api'
import AppIcon from '@/components/chrome/AppIcon.vue'
import Pager from '@/components/chrome/Pager.vue'

const route = useRoute()
const router = useRouter()
const size = getSite()?.page_size ?? 12

const query = ref(String(route.query.q || ''))
const activeQuery = computed(() => String(route.query.q || '').trim())
const current = ref(Math.max(1, Number(route.query.current) || 1))
const loading = ref(false)
const result = ref<any>({
  records: [],
  total: 0,
  size,
  current: 1,
  pages: 0,
  took_ms: 0,
  suggestions: { tags: [], categories: [] },
})
const inputEl = ref<HTMLInputElement | null>(null)

let seq = 0

async function runSearch(q: string, page = 1) {
  const id = ++seq
  loading.value = true
  try {
    const res = await searchSite({ q, current: page, size })
    if (id !== seq) return
    result.value = res
    current.value = res.current || page
  } finally {
    if (id === seq) loading.value = false
  }
}

function syncUrl(q: string, page: number) {
  const nextQ = q.trim()
  const nextQuery: Record<string, string> = {}
  if (nextQ) nextQuery.q = nextQ
  if (page > 1) nextQuery.current = String(page)
  const same =
    String(route.query.q || '') === (nextQuery.q || '') &&
    String(route.query.current || '') === (nextQuery.current || '')
  if (!same) router.replace({ name: 'search', query: nextQuery })
  else runSearch(nextQ, page)
}

/** 主动触发：回车 / 点击搜索按钮 / 热门词 */
function submitSearch() {
  syncUrl(query.value, 1)
}

function resetSearch() {
  query.value = ''
  syncUrl('', 1)
  nextTick(() => inputEl.value?.focus())
}

function setPage(page: number) {
  syncUrl(activeQuery.value, page)
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

watch(
  () => [String(route.query.q || ''), String(route.query.current || '')] as const,
  ([q, c]) => {
    if (q !== query.value) query.value = q
    const page = Math.max(1, Number(c) || 1)
    current.value = page
    runSearch(q.trim(), page)
  },
  { immediate: true },
)

onMounted(async () => {
  await nextTick()
  inputEl.value?.focus()
})
</script>

<template>
  <div class="search-page">
    <header class="mb-5">
      <h1 class="text-2xl font-bold md:text-[1.75rem]">搜索</h1>
      <p class="mt-1 text-sm text-muted">按标题与摘要搜索文章与项目；分类 / 标签请点下方入口浏览</p>
    </header>

    <form class="search-page__form" @submit.prevent="submitSearch">
      <div class="search-page__box glass-card">
        <AppIcon name="search" :size="20" class="search-page__icon" />
        <input
          ref="inputEl"
          v-model="query"
          type="search"
          class="search-page__input"
          placeholder="输入关键词后回车或点击搜索"
          autocomplete="off"
          spellcheck="false"
        />
      </div>
      <button type="submit" class="search-page__submit">搜索</button>
      <button type="button" class="search-page__reset" @click="resetSearch">重置</button>
    </form>

    <div v-if="!activeQuery" class="mt-5 flex flex-col gap-5">
      <section v-if="result.suggestions?.categories?.length">
        <div class="mb-2 text-sm text-muted">热门分类</div>
        <div class="flex flex-wrap gap-2">
          <RouterLink
            v-for="cat in result.suggestions.categories"
            :key="cat.uid || cat"
            :to="`/categories/${cat.uid || cat}`"
            class="search-chip"
          >
            {{ cat.name || cat }}
          </RouterLink>
        </div>
      </section>
      <section v-if="result.suggestions?.tags?.length">
        <div class="mb-2 text-sm text-muted">热门标签</div>
        <div class="flex flex-wrap gap-2">
          <RouterLink
            v-for="tag in result.suggestions.tags"
            :key="tag.uid || tag"
            :to="`/tags/${tag.uid || tag}`"
            class="search-chip"
          >
            #{{ tag.name || tag }}
          </RouterLink>
        </div>
      </section>
    </div>

    <template v-else>
      <div class="mt-4 text-sm text-muted">
        <span v-if="loading">搜索中…</span>
        <span v-else-if="result.total">找到 {{ result.total }} 条结果，用时 {{ result.took_ms }} ms</span>
        <span v-else>找不到你查询的内容：{{ activeQuery }}</span>
      </div>

      <ul v-if="result.records?.length" class="mt-4 flex flex-col gap-2">
        <li v-for="hit in result.records" :key="`${hit.type}-${hit.uid}`">
          <RouterLink :to="hit.url" class="search-hit glass-card">
            <img :src="hit.cover" alt="" class="search-hit__cover" loading="lazy" />
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span class="search-hit__type">{{ hit.type === 'project' ? '项目' : '文章' }}</span>
                <span class="search-hit__title">{{ hit.title }}</span>
              </div>
              <p class="search-hit__snippet">{{ hit.snippet || hit.description }}</p>
              <div v-if="hit.tags?.length" class="mt-2 flex flex-wrap gap-1.5">
                <span
                  v-for="tag in hit.tags.slice(0, 4)"
                  :key="typeof tag === 'string' ? tag : tag.uid || tag.name"
                  class="search-hit__tag"
                >
                  #{{ typeof tag === 'string' ? tag : tag.name }}
                </span>
              </div>
            </div>
          </RouterLink>
        </li>
      </ul>

      <Pager
        :current="result.current"
        :pages="result.pages"
        :total="result.total"
        :size="result.size"
        @change="setPage"
      />
    </template>
  </div>
</template>

<style scoped>
.search-page__form {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: nowrap;
}

.search-page__box {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
}

.search-page__icon {
  flex-shrink: 0;
  opacity: 0.55;
}

.search-page__input {
  width: 100%;
  min-width: 0;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  font-size: 1.05rem;
  outline: none;
}

.search-page__input::-webkit-search-cancel-button {
  appearance: none;
}

.search-page__reset,
.search-page__submit {
  flex-shrink: 0;
  border-radius: 999px;
  padding: 0.65rem 1.15rem;
  font: inherit;
  font-size: 0.9rem;
  cursor: pointer;
  white-space: nowrap;
}

.search-page__reset {
  border: 1px solid var(--efu-card-border);
  background: var(--efu-secondbg);
  color: var(--milk-ink);
}

.search-page__reset:hover {
  background: var(--efu-ahoverbg);
}

.search-page__submit {
  border: 0;
  background: var(--milk-ink);
  color: #fff;
  font-weight: 600;
}

.search-page__submit:hover {
  opacity: 0.9;
}

.search-chip {
  display: inline-flex;
  align-items: center;
  border: 1px solid var(--efu-card-border);
  border-radius: 999px;
  background: var(--efu-card-bg);
  color: var(--milk-ink);
  padding: 0.3rem 0.75rem;
  font: inherit;
  font-size: 0.85rem;
  text-decoration: none;
  cursor: pointer;
}

.search-chip:hover {
  background: var(--efu-ahoverbg);
}

.search-hit {
  display: flex;
  gap: 0.9rem;
  padding: 0.9rem 1rem;
  text-decoration: none;
  color: inherit;
  transition:
    transform 0.15s ease,
    box-shadow 0.15s ease;
}

.search-hit:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 20px rgba(44, 52, 68, 0.06);
}

.search-hit__cover {
  width: 5.5rem;
  height: 3.8rem;
  flex-shrink: 0;
  border-radius: 10px;
  object-fit: cover;
  background: var(--efu-secondbg);
}

.search-hit__type {
  flex-shrink: 0;
  padding: 0.05rem 0.4rem;
  border-radius: 4px;
  background: var(--efu-secondbg);
  color: var(--efu-secondtext);
  font-size: 0.72rem;
}

.search-hit__title {
  font-weight: 600;
  font-size: 1rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-hit__snippet {
  margin: 0.3rem 0 0;
  color: var(--milk-ink-soft);
  font-size: 0.88rem;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.search-hit__tag {
  font-size: 0.75rem;
  color: var(--efu-secondtext);
}
</style>
