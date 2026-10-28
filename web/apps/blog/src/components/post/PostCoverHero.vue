<script setup lang="ts">
import AppIcon from '@/components/chrome/AppIcon.vue'
import { formatDate } from '@/utils/markdown'

defineProps<{ post: any }>()
</script>

<template>
  <div id="post-info" class="post-head page-gutter">
    <div class="page-shell post-head__inner">
      <div class="post-head__main">
        <div class="post-head__badges">
          <span v-if="post.sticky" class="badge badge--pin">置顶</span>
          <span class="badge">{{ post.reprint ? '转载' : '原创' }}</span>
          <RouterLink
            v-if="post.categories?.[0]"
            :to="`/categories/${post.categories[0].uid}`"
            class="badge badge--link"
          >
            {{ post.categories[0].name }}
          </RouterLink>
          <RouterLink
            v-for="tag in (post.tags || []).slice(0, 3)"
            :key="tag.uid"
            :to="`/tags/${tag.uid}`"
            class="badge badge--link"
          >
            <AppIcon name="tags" :size="12" />
            {{ tag.name }}
          </RouterLink>
        </div>

        <h1 class="post-head__title">{{ post.title }}</h1>

        <p v-if="post.description" class="post-head__desc">
          {{ post.description }}
        </p>

        <div class="post-head__meta">
          <span class="pill">
            <AppIcon name="view" :size="14" />
            {{ Number(post.view_count) || 0 }}
          </span>
          <span class="pill">
            <AppIcon name="calendar" :size="14" />
            {{ formatDate(post.date) }}
          </span>
          <span v-if="post.updated" class="pill">
            <AppIcon name="refresh" :size="14" />
            {{ formatDate(post.updated) }}
          </span>
          <span v-if="post.reading_time_minutes" class="pill">
            <AppIcon name="time" :size="14" />
            {{ post.reading_time_minutes }} 分钟
          </span>
          <span v-if="post.locate" class="pill">
            <AppIcon name="local" :size="14" />
            {{ post.locate }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.post-head {
  position: relative;
  z-index: 10;
  padding-top: calc(4.75rem + env(safe-area-inset-top, 0px));
  padding-bottom: 1.35rem;
  color: #fff;
}

.post-head__inner {
  display: flex;
  align-items: flex-end;
  min-height: 0;
}

.post-head__main {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  width: 100%;
  max-width: 44rem;
  min-width: 0;
}

.post-head__badges {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.15rem 0.5rem;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.16);
  color: inherit;
  font-size: 0.75rem;
  line-height: 1.35;
  text-decoration: none;
}

.badge--link:hover {
  background: rgba(255, 255, 255, 0.26);
}

.badge--pin {
  background: linear-gradient(135deg, #e85d4c 0%, #d4483a 100%);
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.post-head__title {
  margin: 0;
  font-size: clamp(1.35rem, 3.2vw, 1.85rem);
  font-weight: 700;
  line-height: 1.3;
}

.post-head__desc {
  margin: 0;
  max-width: 38rem;
  font-size: 0.9rem;
  line-height: 1.55;
  color: rgba(255, 255, 255, 0.92);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.post-head__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  height: 1.7rem;
  padding: 0 0.65rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.16);
  font-size: 0.78rem;
  white-space: nowrap;
}

@media (min-width: 768px) {
  .post-head {
    padding-top: calc(4.5rem + env(safe-area-inset-top, 0px));
    padding-bottom: 1.6rem;
  }

  .post-head__inner {
    min-height: 8.5rem;
  }

  .post-head__desc {
    -webkit-line-clamp: 3;
  }
}

@media (max-width: 767px) {
  .post-head__title {
    font-size: 1.28rem;
  }
}
</style>
