/**
 * 站点前端 API 层（对接 Gin `/portal/blog/*`）。
 *
 * ## 约定
 * - 入参、出参均为普通对象
 * - 列表：`{ list: T[] }`；详情：`{ post | project | doc | site | … }`
 * - 分页：MyBatis-Plus **IPage**（`current` / `size` / `records` / `total` / `pages`）
 * - 传输层：`{ code, data, message }`，`request()` 已解包为业务 `data`
 * - GET → `params`；写操作 → `data`（目前均为读接口）
 * - 环境：`VITE_API_BASE`（默认 `/api`）
 *
 * ## 分页（MyBatis-Plus `Page` / `IPage`）
 *
 * **请求 query**
 * | 字段 | 说明 |
 * |---|---|
 * | current | 当前页，从 1 起，默认 1 |
 * | size | 每页条数，默认站点 `page_size` |
 *
 * **响应（IPage）**
 * | 字段 | 说明 |
 * |---|---|
 * | records | 当前页记录 |
 * | total | 总记录数 |
 * | size | 每页条数 |
 * | current | 当前页 |
 * | pages | 总页数（`total===0` → `0`） |
 *
 * ## 实体字段
 *
 * **Post**
 * | 字段 | 说明 |
 * |---|---|
 * | uid | 对外主键（UUIDv7 无连字符），路由 `/p/:uid` |
 * | title | 标题 |
 * | description? | 摘要 |
 * | date | 发布日 `YYYY-MM-DD` |
 * | updated? | 更新日期 |
 * | cover? | 封面 URL |
 * | categories | `{uid,name}[]` |
 * | tags | `{uid,name}[]` |
 * | tags | 标签名数组 |
 * | reprint? | 是否转载 |
 * | sticky? | 是否置顶 |
 * | home? | `false` 时不进首页流 / `/posts/page` |
 * | reading_time_minutes? | 阅读分钟 |
 * | locate? | 地点 |
 * | body | Markdown 正文 |
 *
 * **Project**
 * | 字段 | 说明 |
 * |---|---|
 * | uid | 对外主键（UUIDv7 无连字符），路由 `/projects/:uid` |
 * | title | 标题 |
 * | description? | 摘要 |
 * | date | `YYYY-MM-DD` |
 * | weight? | 排序权重，越小越前 |
 * | cover? | 封面 URL |
 * | git? | 仓库链接 |
 * | site? | 站点链接 |
 * | project_tags | `{uid,name}[]` |
 * | body | Markdown 正文 |
 *
 * **ArchiveEntry**（兼容列表/grid；含 PostCard 所需字段）
 * | 字段 | 说明 |
 * |---|---|
 * | uid | 文章 uid |
 * | title | 标题 |
 * | url | 前端路径，如 `/p/{uid}` |
 * | cover | 封面 URL（已解析默认图） |
 * | primary_category | 主分类 |
 * | date_label | 展示用日期文案 |
 * | year | 四位年份 |
 * | description | 摘要 |
 * | date | 原始日期 `YYYY-MM-DD` |
 * | tags / categories | 标签与分类数组 |
 *
 * **TaxonomyTerm**
 * | 字段 | 说明 |
 * |---|---|
 * | uid | 对外主键（UUIDv7 无连字符），路由 `/categories|:tags|:project-tags/:uid` |
 * | name | 显示名 |
 * | count | 数量 |
 *
 * **Site**
 * | 字段 | 说明 |
 * |---|---|
 * | title / description / site_name | 站点文案 |
 * | icon | 站点图标 URL |
 * | nav | 导航项数组，见下 |
 * | nav_group | 顶栏外链 `{ name, url, icon? }[]` |
 * | author | 侧栏作者卡，见下 |
 * | default_covers | 默认封面 URL 列表 |
 * | footer | `{ beian[], police_html?, copyright, email?, wechat_name?, wechat_qr? }`；联系方式在上，统计居中，版权/备案置底 |
 * | started_at | 上线时间 ISO，页脚运行时长 |
 * | page_size | 默认分页 size（请求映射为 IPage.`size`） |
 *
 * **Site.nav[] 项**
 * | 字段 | 说明 |
 * |---|---|
 * | name | 显示名 |
 * | href? | 站内路径或外链 |
 * | icon? | IconPark 键名 |
 * | external? | 是否外链 |
 * | children? | 子导航，结构同 nav 项 |
 *
 * **Site.author**
 * | 字段 | 说明 |
 * |---|---|
 * | name / avatar_url / background_url? | 身份与背景 |
 * | description | 侧栏作者卡简介 |
 * | content | 关于页等长文（侧栏不展示） |
 * | witty_words | 已废弃，侧栏不展示 |
 * | links | `{ name, url, icon }[]` |
 *
 * **About.doc**
 * | 字段 | 说明 |
 * |---|---|
 * | title | 页面标题 |
 * | content | 整页 Markdown 正文 |
 */

import { shallowRef } from 'vue'
import { request } from './http'

/** bootstrap 后填充；chrome 壳层同步读取（shallowRef 便于前台随配置刷新） */
const siteCache = shallowRef<any>(null)

/**
 * 应用启动时拉取站点配置并写入缓存。
 *
 * `GET /site`
 *
 * @returns {{ site: object }}
 */
export async function bootstrapApi() {
  const { site } = await request({ url: '/portal/blog/site' })
  siteCache.value = site
  return { site }
}

/**
 * 同步读取站点配置（须先 `bootstrapApi()`）。
 *
 * @param {object} [_params]
 * @returns {object|null} Site
 */
export function getSite(_params: any = {}) {
  return siteCache.value
}

/**
 * 重新拉取站点配置（异步，并刷新缓存）。
 *
 * `GET /site`
 *
 * @param {object} [_params]
 * @returns {Promise<{ site: object }>}
 */
export async function fetchSite(_params: any = {}) {
  const res = await request({ url: '/portal/blog/site', params: _params })
  siteCache.value = res.site
  return res
}

/**
 * 站点运营统计（访问 / 阅读 / 文章数 / 运行起点等）。
 *
 * `GET /site/stats`
 *
 * @param {object} [_params]
 * @returns {Promise<{ visit_count: number, read_count: number, post_count: number, project_count: number, started_at: string }>}
 */
export async function getSiteStats(_params: any = {}) {
  return request({ url: '/portal/blog/site/stats', params: _params })
}

/**
 * 上报一次页面访问（PV；带 visitor_id 做 UV；文章详情带 post_uid 记阅读）。
 *
 * `POST /portal/blog/visit`
 */
export async function trackVisit(body: {
  visitor_id: string
  path: string
  post_uid?: string
  project_uid?: string
}): Promise<{ ok: boolean; counted?: boolean; view_count?: number }> {
  return request({ url: '/portal/blog/visit', method: 'post', data: body })
}

/**
 * 全部文章（按 date 降序）。
 *
 * `GET /posts`
 *
 * @param {object} [_params]
 * @returns {Promise<{ list: object[] }>} list = Post[]
 */
export async function listPosts(_params: any = {}) {
  return request({ url: '/portal/blog/posts', params: _params })
}

/**
 * 首页文章流（排除 `home === false`）。
 *
 * `GET /posts/home`
 *
 * @param {object} [_params]
 * @returns {Promise<{ list: object[] }>} list = Post[]
 */
export async function getHomePosts(_params: any = {}) {
  return request({ url: '/portal/blog/posts/home', params: _params })
}

/**
 * 文章详情。
 *
 * `GET /posts/detail`
 *
 * @param {object} params
 * @param {string} params.uid - 文章 uid（UUIDv7 无连字符）
 * @returns {Promise<{ post: object|null }>}
 */
export async function getPost(params: any = {}) {
  return request({ url: '/portal/blog/posts/detail', params })
}

/**
 * 相关文章（标签/分类加权）。
 *
 * `GET /posts/related`
 *
 * @param {object} params
 * @param {string} params.uid - 当前文章 uid
 * @param {number} [params.limit=3] - 返回条数
 * @returns {Promise<{ list: object[] }>} list = Post[]
 */
export async function getRelatedPosts(params: any = {}) {
  return request({ url: '/portal/blog/posts/related', params })
}

/**
 * 上一篇 / 下一篇（按发布日降序：prev 更旧，next 更新）。
 *
 * `GET /posts/neighbors`
 *
 * @param {object} params
 * @param {string} params.uid - 当前文章 uid
 * @returns {Promise<{ prev: {uid,title,cover}|null, next: {uid,title,cover}|null }>}
 */
export async function getPostNeighbors(params: any = {}) {
  return request({ url: '/portal/blog/posts/neighbors', params })
}

/**
 * 首页文章分页（IPage，仅 `home !== false`）。
 *
 * `GET /posts/page`
 *
 * @param {object} [params]
 * @param {number} [params.current=1]
 * @param {number} [params.size]
 * @returns {Promise<{ records: object[], total: number, size: number, current: number, pages: number }>}
 */
export async function paginatePosts(params: any = {}) {
  return request({ url: '/portal/blog/posts/page', params })
}

/**
 * 按分类名筛选文章（全量 list，分页请用 `paginateArchive`）。
 *
 * `GET /posts/by-category`
 *
 * @param {object} params
 * @param {string} params.uid - 分类 uid
 * @returns {Promise<{ list: object[] }>} list = Post[]
 */
export async function postsByCategory(params: any = {}) {
  return request({ url: '/portal/blog/posts/by-category', params })
}

/**
 * 按标签 uid 筛选文章（全量 list，分页请用 `paginateArchive`）。
 *
 * `GET /posts/by-tag`
 *
 * @param {object} params
 * @param {string} params.uid - 标签 uid
 * @returns {Promise<{ list: object[] }>} list = Post[]
 */
export async function postsByTag(params: any = {}) {
  return request({ url: '/portal/blog/posts/by-tag', params })
}

/**
 * 全部项目（按 weight 升序）。
 *
 * `GET /projects`
 *
 * @param {object} [_params]
 * @returns {Promise<{ list: object[] }>} list = Project[]
 */
export async function listProjects(_params: any = {}) {
  return request({ url: '/portal/blog/projects', params: _params })
}

/**
 * 项目分页（IPage）。
 *
 * `GET /projects/page`
 *
 * @param {object} [params]
 * @param {number} [params.current=1]
 * @param {number} [params.size]
 * @param {string} [params.uid] - 按项目标签 uid 过滤
 * @returns {Promise<{ records: object[], total: number, size: number, current: number, pages: number }>}
 */
export async function paginateProjects(params: any = {}) {
  return request({ url: '/portal/blog/projects/page', params })
}

/**
 * 项目详情。
 *
 * `GET /projects/detail`
 *
 * @param {object} params
 * @param {string} params.uid - 项目 uid（UUIDv7 无连字符）
 * @returns {Promise<{ project: object|null }>}
 */
export async function getProject(params: any = {}) {
  return request({ url: '/portal/blog/projects/detail', params })
}

/**
 * 按项目标签筛选（全量 list，分页请用 `paginateProjects({ tag })`）。
 *
 * `GET /projects/by-tag`
 *
 * @param {object} params
 * @param {string} params.uid - 项目标签 uid
 * @returns {Promise<{ list: object[] }>} list = Project[]
 */
export async function projectsByTag(params: any = {}) {
  return request({ url: '/portal/blog/projects/by-tag', params })
}

/**
 * 关于页文档。
 *
 * `GET /about`
 *
 * @param {object} [_params]
 * @returns {Promise<{ doc: object }>} doc 见文件头 About.doc
 */
export async function getAbout(_params: any = {}) {
  return request({ url: '/portal/blog/about', params: _params })
}

/**
 * 站内搜索（仅匹配标题；文章 + 项目，IPage 分页）。
 *
 * `GET /search`
 *
 * @param {object} params
 * @param {string} params.q - 关键词（只搜 title）
 * @param {number} [params.current=1]
 * @param {number} [params.size]
 * @returns {Promise<{ records: object[], total: number, size: number, current: number, pages: number, took_ms: number, suggestions: { tags: string[], categories: string[] } }>}
 */
export async function searchSite(params: any = {}) {
  return request({ url: '/portal/blog/search', params })
}

/**
 * 归档列表项（全量；分页请用 `paginateArchive`）。
 *
 * `GET /archive/entries`
 *
 * @param {object} [params]
 * @param {string} [params.uid] - 分类或标签 uid
 * @param {'category'|'tag'} [params.by] - 与 uid 联用
 * @returns {Promise<{ list: object[] }>} list = ArchiveEntry[]
 */
export async function toArchiveEntries(params: any = {}) {
  return request({ url: '/portal/blog/archive/entries', params })
}

/**
 * 归档年份列表（降序）。
 *
 * `GET /archive/years`
 *
 * @param {object} [_params]
 * @returns {Promise<{ list: string[] }>}
 */
export async function listArchiveYears(_params: any = {}) {
  return request({ url: '/portal/blog/archive/years', params: _params })
}

/**
 * 归档分页（IPage）。
 *
 * `GET /archive/page`
 *
 * @param {object} [params]
 * @param {number} [params.current=1]
 * @param {number} [params.size]
 * @param {string} [params.year='all'] - 四位年份或 `'all'`
 * @param {'category'|'tag'} [params.by] - 与 uid 联用过滤
 * @param {string} [params.uid] - 分类/标签 uid
 * @returns {Promise<{ records: object[], total: number, size: number, current: number, pages: number }>}
 */
export async function paginateArchive(params: any = {}) {
  return request({ url: '/portal/blog/archive/page', params })
}

/**
 * 文章分类云。
 *
 * `GET /taxonomy/categories`
 *
 * @param {object} [_params]
 * @returns {Promise<{ list: object[] }>} list = TaxonomyTerm[]
 */
export async function listCategories(_params: any = {}) {
  return request({ url: '/portal/blog/taxonomy/categories', params: _params })
}

/**
 * 文章标签云。
 *
 * `GET /taxonomy/tags`
 *
 * @param {object} [_params]
 * @returns {Promise<{ list: object[] }>} list = TaxonomyTerm[]
 */
export async function listTags(_params: any = {}) {
  return request({ url: '/portal/blog/taxonomy/tags', params: _params })
}

/**
 * 项目标签云。
 *
 * `GET /taxonomy/project-tags`
 *
 * @param {object} [_params]
 * @returns {Promise<{ list: object[] }>} list = TaxonomyTerm[]
 */
export async function listProjectTags(_params: any = {}) {
  return request({ url: '/portal/blog/taxonomy/project-tags', params: _params })
}

export { request, http } from './http'
