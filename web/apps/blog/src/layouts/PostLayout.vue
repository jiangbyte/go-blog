<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getPost, getSite } from '@/api'
import MobileToc from '@/components/aside/MobileToc.vue'
import BackToTop from '@/components/chrome/BackToTop.vue'
import SiteAside from '@/components/chrome/SiteAside.vue'
import SiteFooter from '@/components/chrome/SiteFooter.vue'
import SiteHeader from '@/components/chrome/SiteHeader.vue'
import PostCoverHero from '@/components/post/PostCoverHero.vue'
import { applySeo, pageAbsoluteUrl } from '@/composables/useSeo'
import { coverOf } from '@/utils/markdown'

const route = useRoute()
const site = getSite()
const post = ref<any>(null)

watch(
  () => String(route.params.uid),
  async (uid) => {
    // 1. 拉取文章详情
    const { post: p } = await getPost({ uid })
    post.value = p
    // 2. 写入 title / OG / Article JSON-LD，供搜索引擎与分享抓取
    if (!p) {
      applySeo({ title: '文章不存在', noindex: true, url: pageAbsoluteUrl(route.path) })
      return
    }
    const siteName = String(site?.site_name || site?.title || 'Demo Blog')
    const cover = coverOf(p.cover, site?.default_covers)
    applySeo({
      title: `${p.title} · ${siteName}`,
      description: String(p.description || '').trim() || site?.description || '',
      image: cover || undefined,
      url: pageAbsoluteUrl(`/p/${p.uid}`),
      type: 'article',
      jsonLd: {
        '@context': 'https://schema.org',
        '@type': 'BlogPosting',
        headline: p.title,
        description: p.description || '',
        image: cover || undefined,
        datePublished: p.date || undefined,
        dateModified: p.updated || p.date || undefined,
        mainEntityOfPage: pageAbsoluteUrl(`/p/${p.uid}`),
      },
    })
  },
  { immediate: true },
)

const coverUrl = computed(() => coverOf(post.value?.cover, site.default_covers))
const headerStyle = computed(() =>
  post.value ? { '--post-cover-url': `url('${coverUrl.value}')` } : undefined,
)
</script>

<template>
  <div class="flex min-h-screen flex-col bg-surface text-ink">
    <SiteHeader :on-hero="!!post" />

    <div v-if="post" class="post-hero relative w-full" :style="headerStyle">
      <div
        class="absolute inset-0 bg-cover bg-center"
        :style="{ backgroundImage: 'var(--post-cover-url)' }"
      />
      <div class="absolute inset-0 bg-gradient-to-b from-black/40 via-black/28 to-black/62" />
      <PostCoverHero :post="post" />
    </div>

    <div class="page-gutter flex-1 pb-8" :class="post ? 'pt-5' : 'pt-[76px]'">
      <div
        class="page-shell layout-shell grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_300px] lg:gap-[15px]"
      >
        <main class="min-w-0 overflow-x-clip">
          <RouterView />
        </main>
        <SiteAside class="hidden lg:flex" show-toc />
      </div>
    </div>
    <MobileToc />
    <BackToTop />
    <SiteFooter />
  </div>
</template>

<style scoped>
.post-hero {
  min-height: 0;
}
</style>
