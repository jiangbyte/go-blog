import { getSite, trackVisit } from '@/api'
import { applySeo, pageAbsoluteUrl } from '@/composables/useSeo'
import { getVisitorId, postUidFromRoute, projectUidFromRoute } from '@/utils/visit'
import { createRouter, createWebHistory } from 'vue-router'

/** 列表/静态页 SEO 文案（详情页由布局在数据就绪后覆盖） */
const ROUTE_SEO: Record<string, { title: string; description: string; noindex?: boolean }> = {
  home: { title: '', description: '' },
  archives: { title: '归档', description: '按时间浏览全部文章' },
  search: { title: '搜索', description: '搜索文章与项目', noindex: true },
  categories: { title: '分类', description: '按分类浏览文章' },
  tags: { title: '标签', description: '按标签浏览文章' },
  projects: { title: '项目', description: '已发布的开源与个人项目' },
  'project-tags': { title: '项目标签', description: '按标签浏览项目' },
  about: { title: '关于', description: '关于本站与作者' },
  'tools-json': { title: 'JSON 工具', description: '在线 JSON 格式化与校验', noindex: true },
  'tools-subnet': { title: '子网计算', description: 'CIDR / 子网掩码计算', noindex: true },
  'not-found': { title: '页面不存在', description: '', noindex: true },
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior() {
    return { top: 0 }
  },
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/DefaultLayout.vue'),
      children: [
        { path: '', name: 'home', component: () => import('@/views/HomeView.vue') },
        { path: 'archives', name: 'archives', component: () => import('@/views/ArchivesView.vue') },
        { path: 'search', name: 'search', component: () => import('@/views/SearchView.vue') },
        {
          path: 'categories',
          name: 'categories',
          component: () => import('@/views/CategoriesView.vue'),
        },
        {
          path: 'categories/:uid',
          name: 'category-term',
          component: () => import('@/views/CategoryTermView.vue'),
        },
        { path: 'tags', name: 'tags', component: () => import('@/views/TagsView.vue') },
        {
          path: 'tags/:uid',
          name: 'tag-term',
          component: () => import('@/views/TagTermView.vue'),
        },
      ],
    },
    {
      path: '/',
      component: () => import('@/layouts/WideLayout.vue'),
      children: [
        { path: 'projects', name: 'projects', component: () => import('@/views/ProjectsView.vue') },
        {
          path: 'project-tags',
          name: 'project-tags',
          component: () => import('@/views/ProjectTagsView.vue'),
        },
        {
          path: 'project-tags/:uid',
          name: 'project-tag-term',
          component: () => import('@/views/ProjectTagTermView.vue'),
        },
        { path: 'about', name: 'about', component: () => import('@/views/AboutView.vue') },
        {
          path: 'tools/json',
          name: 'tools-json',
          component: () => import('@/views/ToolsJsonView.vue'),
        },
        {
          path: 'tools/subnet',
          name: 'tools-subnet',
          component: () => import('@/views/ToolsSubnetView.vue'),
        },
      ],
    },
    {
      path: '/',
      component: () => import('@/layouts/PostLayout.vue'),
      children: [
        { path: 'p/:uid', name: 'post', component: () => import('@/views/PostDetailView.vue') },
      ],
    },
    {
      path: '/',
      component: () => import('@/layouts/ProjectLayout.vue'),
      children: [
        {
          path: 'projects/:uid',
          name: 'project',
          component: () => import('@/views/ProjectDetailView.vue'),
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      component: () => import('@/layouts/ErrorLayout.vue'),
      children: [
        { path: '', name: 'not-found', component: () => import('@/views/NotFoundView.vue') },
      ],
    },
  ],
})

/** 列表页默认 SEO；文章/项目详情交给布局在内容加载后写入 */
function applyListSeo(name: string | symbol | undefined | null, path: string) {
  const key = typeof name === 'string' ? name : ''
  // 详情与 term 页由页面自身写入，避免空白标题抢先覆盖
  if (
    key === 'post' ||
    key === 'project' ||
    key === 'category-term' ||
    key === 'tag-term' ||
    key === 'project-tag-term'
  ) {
    return
  }
  const site = getSite() || {}
  const siteName = String(site.site_name || site.title || 'Demo Blog')
  const siteDesc = String(site.description || '')
  const preset = ROUTE_SEO[key] || { title: '', description: siteDesc }
  const pageTitle = preset.title ? `${preset.title} · ${siteName}` : siteName
  applySeo({
    title: pageTitle,
    description: preset.description || siteDesc,
    url: pageAbsoluteUrl(path),
    type: 'website',
    noindex: !!preset.noindex,
    jsonLd:
      key === 'home'
        ? {
            '@context': 'https://schema.org',
            '@type': 'WebSite',
            name: siteName,
            description: siteDesc,
            url: pageAbsoluteUrl('/'),
            potentialAction: {
              '@type': 'SearchAction',
              target: `${pageAbsoluteUrl('/search')}?q={search_term_string}`,
              'query-input': 'required name=search_term_string',
            },
          }
        : null,
  })
}

/** 路由切换上报 PV/UV（失败静默；同页 30 分钟内去重，避免刷新刷量） */
const TRACK_DEDUP_MS = 30 * 60 * 1000
router.afterEach((to) => {
  const path = (to.path || '/').split(/[?#]/)[0] || '/'
  applyListSeo(to.name, path)
  if (to.name === 'not-found') return
  const params = to.params as Record<string, any>
  const postUID = postUidFromRoute(path, params)
  const projectUID = postUID ? '' : projectUidFromRoute(path, params)
  const dedupeKey = `jb_track:${postUID || projectUID || path}`
  try {
    const prev = Number(sessionStorage.getItem(dedupeKey) || 0)
    if (prev && Date.now() - prev < TRACK_DEDUP_MS) return
    sessionStorage.setItem(dedupeKey, String(Date.now()))
  } catch {
    // sessionStorage 不可用时仍上报，由后端 Redis 去重
  }
  void trackVisit({
    visitor_id: getVisitorId(),
    path,
    post_uid: postUID || undefined,
    project_uid: projectUID || undefined,
  }).catch(() => {})
})

export default router
