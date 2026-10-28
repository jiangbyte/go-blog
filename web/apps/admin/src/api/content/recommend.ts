/** Author: go-blog — 内容推荐位 API */

import { http } from '@/utils'

export function page(params?: { current?: number; size?: number }) {
  return http.get<{
    records: any[]
    total: string
    current: string
    size: string
  }>('/admin/content/recommend/page', { params })
}

export function replace(items: Array<{ article_id: string; sort_code: number }>) {
  return http.post('/admin/content/recommend/replace', { items })
}

export function add(article_id: string) {
  return http.post('/admin/content/recommend/add', { article_id })
}

export function remove(article_id: string) {
  return http.post('/admin/content/recommend/remove', { article_id })
}
