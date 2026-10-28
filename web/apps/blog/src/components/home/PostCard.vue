<script setup lang="ts">
import AppIcon from '@/components/chrome/AppIcon.vue'
import { coverOf, formatDate } from '@/utils/markdown'
import { getSite } from '@/api'

defineProps<{ post: any; index: number }>()
const site = getSite()
</script>

<template>
  <article class="post-card glass-card group h-full overflow-hidden transition hover:-translate-y-0.5 hover:shadow-[0_8px_20px_rgba(44,52,68,0.06)]">
    <RouterLink :to="`/p/${post.uid}`" class="post-card__cover relative block overflow-hidden">
      <img
        :src="coverOf(post.cover, site.default_covers, index)"
        :alt="post.title"
        class="h-[132px] w-full object-cover transition duration-300 group-hover:scale-[1.02] sm:h-[148px]"
        loading="lazy"
      />
      <span v-if="post.sticky" class="post-card__pin" aria-label="置顶">置顶</span>
    </RouterLink>

    <div class="post-card__body">
      <RouterLink
        :to="`/p/${post.uid}`"
        class="post-card__title text-lg font-bold leading-snug text-ink transition hover:opacity-80"
        :title="post.title"
      >
        {{ post.title }}
      </RouterLink>

      <p class="post-card__desc text-sm leading-relaxed text-muted" :title="post.description">
        {{ post.description }}
      </p>

      <div v-if="post.tags?.length" class="post-card__tags flex flex-wrap gap-1.5">
        <RouterLink
          v-for="tag in post.tags.slice(0, 4)"
          :key="tag.uid || tag"
          :to="`/tags/${tag.uid || tag}`"
          class="milk-hover rounded-md bg-second-bg px-2 py-0.5 text-xs text-ink-soft"
        >
          {{ tag.name || tag }}
        </RouterLink>
      </div>

      <div class="post-card__meta flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted">
        <span class="inline-flex items-center gap-1">
          <AppIcon name="user" :size="13" />
          {{ site.author.name }}
        </span>
        <RouterLink
          v-if="post.categories?.[0]"
          :to="`/categories/${post.categories[0].uid || post.categories[0]}`"
          class="inline-flex items-center gap-1 hover:text-ink"
        >
          <AppIcon name="folder" :size="13" />
          {{ post.categories[0].name || post.categories[0] }}
        </RouterLink>
        <time class="inline-flex items-center gap-1">
          <AppIcon name="calendar" :size="13" />
          {{ formatDate(post.date) }}
        </time>
      </div>
    </div>
  </article>
</template>

<style scoped>
.post-card {
  display: grid;
  grid-template-rows: auto 1fr;
}

/* 封面左上角置顶角标，与灰标签区分开 */
.post-card__pin {
  position: absolute;
  top: 0.65rem;
  left: 0.65rem;
  z-index: 2;
  padding: 0.15rem 0.55rem;
  border-radius: 999px;
  background: linear-gradient(135deg, #e85d4c 0%, #d4483a 100%);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.35;
  box-shadow: 0 2px 8px rgba(212, 72, 58, 0.35);
}

.post-card__body {
  display: grid;
  grid-template-rows: auto auto 1fr auto;
  gap: 0.65rem;
  min-height: 0;
  padding: 1rem 1.1rem 1.1rem;
}

@media (min-width: 768px) {
  .post-card__body {
    padding: 1.15rem 1.25rem 1.25rem;
  }
}

/* 标题一行省略 */
.post-card__title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  line-clamp: 1;
  overflow: hidden;
  overflow-wrap: anywhere;
}

/* 概要两行省略 */
.post-card__desc {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
  min-height: calc(1.625em * 2);
}

.post-card__tags {
  align-content: start;
}

.post-card__meta {
  align-self: end;
}
</style>
