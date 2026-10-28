/** Author: go-blog — 内容文章 API */

import { http } from '@/utils'

export function page(params: {
  current?: number
  size?: number
  title?: string
  status?: string
  category_id?: string
  tag_id?: string
}) {
  return http.get<{
    records: any[]
    total: string
    current: string
    size: string
  }>('/admin/content/article/page', { params })
}

export function detail(id: string) {
  return http.get<any>('/admin/content/article/detail', { params: { id } })
}

export function create(body: {
  category_id?: string | null
  title: string
  slug?: string | null
  summary?: string | null
  cover_url?: string | null
  body_md: string
  status?: string
  tag_ids?: string[]
  sort_code?: number
}) {
  return http.post<any>('/admin/content/article/create', body)
}

export function update(body: {
  id: string
  category_id?: string | null
  title: string
  slug?: string | null
  summary?: string | null
  cover_url?: string | null
  body_md: string
  status?: string
  tag_ids?: string[]
  sort_code?: number
}) {
  return http.post<any>('/admin/content/article/update', body)
}

export function remove(ids: string[]) {
  return http.post('/admin/content/article/remove', { ids })
}

export function recordView(id: string) {
  return http.post<{ counted: boolean; view_count: number }>('/admin/content/article/record-view', {
    id,
  })
}
