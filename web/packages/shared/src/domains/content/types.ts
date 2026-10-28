/** 公开内容 API 类型 */

export type ContentTagBrief = {
  id: string
  name: string
  slug: string
}

export type ContentArticleListItem = {
  id: string
  title: string
  slug: string
  summary: string
  cover_url: string
  category_id: string | null
  category_name: string
  category_slug: string
  tags: ContentTagBrief[]
  published_at: string | null
  view_count: number
  updated_at: string | null
}

export type ContentArticleDetail = ContentArticleListItem & {
  body_md: string
  body_html: string
}

export type ContentArticlePage = {
  records: ContentArticleListItem[]
  total: string
  current: string
  size: string
}

export type ContentArticleNavItem = {
  id: string
  title: string
  slug: string
  cover_url: string
}

export type ContentArticleNeighbors = {
  prev: ContentArticleNavItem | null
  next: ContentArticleNavItem | null
}

export type ContentCategoryNode = {
  id: string
  parent_id: string | null
  name: string
  slug: string
  sort_code: number
  article_count: number
  children: ContentCategoryNode[]
}

export type ContentTagOption = {
  id: string
  name: string
  slug: string
  sort_code: number
  article_count: number
}

export type ContentArticleSort = 'latest' | 'hot' | 'recommend'
