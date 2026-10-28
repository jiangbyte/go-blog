<script setup lang="ts">
import { computed } from 'vue'
import { getSite } from '@/api'
import { coverOf } from '@/utils/markdown'

const props = defineProps<{
  prev: any | null
  next: any | null
}>()

const site = getSite()

function coverStyle(post: any | null, fallbackIndex: number) {
  if (!post) return undefined
  const url = coverOf(post.cover, site.default_covers, fallbackIndex)
  return { '--nav-cover': `url('${url}')` } as Record<string, string>
}

const prevStyle = computed(() => coverStyle(props.prev, 0))
const nextStyle = computed(() => coverStyle(props.next, 1))
</script>

<template>
  <nav class="post-nav" aria-label="文章翻页">
    <RouterLink
      v-if="prev"
      :to="`/p/${prev.uid}`"
      class="post-nav__card post-nav__card--prev has-cover"
      :style="prevStyle"
    >
      <span class="post-nav__label">&lt; 上一篇</span>
      <span class="post-nav__title">{{ prev.title }}</span>
    </RouterLink>
    <div v-else class="post-nav__card post-nav__card--prev is-empty">
      <span class="post-nav__label">&lt; 上一篇</span>
      <span class="post-nav__title">没有了</span>
    </div>

    <RouterLink
      v-if="next"
      :to="`/p/${next.uid}`"
      class="post-nav__card post-nav__card--next has-cover"
      :style="nextStyle"
    >
      <span class="post-nav__label">下一篇 &gt;</span>
      <span class="post-nav__title">{{ next.title }}</span>
    </RouterLink>
    <div v-else class="post-nav__card post-nav__card--next is-empty">
      <span class="post-nav__label">下一篇 &gt;</span>
      <span class="post-nav__title">没有了</span>
    </div>
  </nav>
</template>

<style scoped>
.post-nav {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}

@media (min-width: 768px) {
  .post-nav {
    grid-template-columns: 1fr 1fr;
    gap: 0.85rem;
  }
}

.post-nav__card {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.35rem;
  min-height: 7.5rem;
  padding: 1.15rem 1.25rem;
  overflow: hidden;
  border: 1px solid var(--efu-card-border);
  border-radius: 12px;
  background: var(--efu-card-bg);
  color: var(--milk-ink);
  text-decoration: none;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.post-nav__card--prev {
  align-items: flex-start;
  text-align: left;
}

.post-nav__card--next {
  align-items: flex-end;
  text-align: right;
}

.post-nav__card.has-cover {
  color: #fff;
  border-color: transparent;
  background-color: #2a2622;
  background-image: linear-gradient(rgba(20, 18, 16, 0.48), rgba(20, 18, 16, 0.58)),
    var(--nav-cover);
  background-size: cover;
  background-position: center;
}

.post-nav__card.has-cover:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 28px rgba(20, 24, 32, 0.18);
  background-image: linear-gradient(rgba(20, 18, 16, 0.38), rgba(20, 18, 16, 0.48)),
    var(--nav-cover);
}

.post-nav__card.is-empty {
  background: var(--efu-secondbg);
  color: var(--milk-ink-soft);
  cursor: default;
}

.post-nav__label {
  font-size: 0.8rem;
  opacity: 0.85;
  line-height: 1.2;
}

.post-nav__card.has-cover .post-nav__label {
  opacity: 0.9;
}

.post-nav__title {
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.post-nav__card.is-empty .post-nav__title {
  font-weight: 500;
}
</style>
