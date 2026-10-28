<script setup lang="ts">
import { computed } from 'vue'
import { getSite } from '@/api'
import AppIcon from '@/components/chrome/AppIcon.vue'

const site = getSite()

const avatarUrl = computed(() => {
  const url = String(site.author?.avatar_url || '').trim()
  return url || String(site.icon || '').trim()
})

const displayName = computed(() => String(site.author?.name || site.site_name || '').trim())

const description = computed(() => String(site.author?.description || '').trim())

const links = computed(() => {
  const list = site.author?.links
  return Array.isArray(list) ? list.filter((l: any) => String(l?.name || '').trim()) : []
})

function isInternalPath(url: string | undefined) {
  const u = String(url || '')
  return u.startsWith('/') && !u.startsWith('//')
}
</script>

<template>
  <section class="glass-card author-card relative overflow-hidden">
    <div
      class="absolute inset-0 bg-cover bg-center"
      :style="{ backgroundImage: `url('${site.author?.background_url || ''}')` }"
    />
    <div class="author-card__shade absolute inset-0" />
    <div class="author-card__body relative z-10 flex flex-col items-center text-center text-white">
      <img
        v-if="avatarUrl"
        :src="avatarUrl"
        alt=""
        class="author-card__avatar"
        width="80"
        height="80"
        referrerpolicy="no-referrer"
      />
      <h3 class="author-card__name">{{ displayName }}</h3>
      <p v-if="description" class="author-card__desc">{{ description }}</p>
      <div v-if="links.length" class="author-card__links">
        <template v-for="link in links" :key="link.name">
          <RouterLink
            v-if="isInternalPath(link.url)"
            :to="link.url"
            class="author-card__link"
          >
            <AppIcon v-if="link.icon" :name="link.icon" :size="14" />
            {{ link.name }}
          </RouterLink>
          <a
            v-else
            :href="link.url"
            target="_blank"
            rel="noopener noreferrer"
            class="author-card__link"
          >
            <AppIcon v-if="link.icon" :name="link.icon" :size="14" />
            {{ link.name }}
          </a>
        </template>
      </div>
    </div>
  </section>
</template>

<style scoped>
.author-card {
  min-height: 260px;
}

.author-card__shade {
  background: linear-gradient(
    180deg,
    rgb(0 0 0 / 28%) 0%,
    rgb(0 0 0 / 42%) 45%,
    rgb(0 0 0 / 62%) 100%
  );
}

.author-card__body {
  box-sizing: border-box;
  min-height: 260px;
  padding: 2.25rem 1rem 1.1rem;
}

.author-card__avatar {
  width: 80px;
  height: 80px;
  border: 3px solid rgb(255 255 255 / 55%);
  border-radius: 9999px;
  object-fit: cover;
  box-shadow: 0 8px 24px rgb(0 0 0 / 28%);
}

.author-card__name {
  margin: 0.85rem 0 0;
  font-size: 1.15rem;
  font-weight: 700;
  line-height: 1.25;
  letter-spacing: 0.02em;
  text-shadow: 0 1px 8px rgb(0 0 0 / 35%);
}

.author-card__desc {
  margin: 0.45rem 0 0;
  max-width: 16rem;
  color: rgb(255 255 255 / 88%);
  font-size: 0.78rem;
  line-height: 1.45;
  text-shadow: 0 1px 6px rgb(0 0 0 / 40%);
}

.author-card__links {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.5rem;
  margin-top: auto;
  padding-top: 1.5rem;
}

.author-card__link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.4rem 0.85rem;
  border: 1px solid rgb(255 255 255 / 35%);
  border-radius: 0.55rem;
  background: rgb(0 0 0 / 28%);
  color: inherit;
  font-size: 0.78rem;
  line-height: 1.2;
  text-decoration: none;
  backdrop-filter: blur(8px);
  transition:
    background 0.2s ease,
    border-color 0.2s ease;
}

.author-card__link:hover {
  border-color: rgb(255 255 255 / 55%);
  background: rgb(0 0 0 / 42%);
}
</style>
