<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getProject, getSite } from '@/api'
import MobileToc from '@/components/aside/MobileToc.vue'
import BackToTop from '@/components/chrome/BackToTop.vue'
import SiteAside from '@/components/chrome/SiteAside.vue'
import SiteFooter from '@/components/chrome/SiteFooter.vue'
import SiteHeader from '@/components/chrome/SiteHeader.vue'
import ProjectCoverHero from '@/components/project/ProjectCoverHero.vue'
import { applySeo, pageAbsoluteUrl } from '@/composables/useSeo'
import { coverOf } from '@/utils/markdown'

const route = useRoute()
const site = getSite()
const project = ref<any>(null)

watch(
  () => String(route.params.uid),
  async (uid) => {
    // 1. 拉取项目详情
    const { project: p } = await getProject({ uid })
    project.value = p
    // 2. 写入 SEO head
    if (!p) {
      applySeo({ title: '项目不存在', noindex: true, url: pageAbsoluteUrl(route.path) })
      return
    }
    const siteName = String(site?.site_name || site?.title || 'Demo Blog')
    const cover = coverOf(p.cover, site?.default_covers)
    applySeo({
      title: `${p.title} · ${siteName}`,
      description: String(p.description || '').trim() || site?.description || '',
      image: cover || undefined,
      url: pageAbsoluteUrl(`/projects/${p.uid}`),
      type: 'article',
      jsonLd: {
        '@context': 'https://schema.org',
        '@type': 'SoftwareApplication',
        name: p.title,
        description: p.description || '',
        image: cover || undefined,
        url: pageAbsoluteUrl(`/projects/${p.uid}`),
      },
    })
  },
  { immediate: true },
)

const coverUrl = computed(() => coverOf(project.value?.cover, site.default_covers))
const headerStyle = computed(() =>
  project.value ? { '--post-cover-url': `url('${coverUrl.value}')` } : undefined,
)
</script>

<template>
  <div class="flex min-h-screen flex-col bg-surface text-ink" data-project="true">
    <SiteHeader :on-hero="!!project" />

    <div v-if="project" class="project-hero relative w-full" :style="headerStyle">
      <div
        class="absolute inset-0 bg-cover bg-center"
        :style="{ backgroundImage: 'var(--post-cover-url)' }"
      />
      <div class="absolute inset-0 bg-gradient-to-b from-black/32 via-black/30 to-black/58" />
      <ProjectCoverHero :project="project" />
    </div>

    <div class="page-gutter flex-1 pb-8" :class="project ? 'pt-5' : 'pt-[76px]'">
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
.project-hero {
  min-height: 0;
}
</style>
