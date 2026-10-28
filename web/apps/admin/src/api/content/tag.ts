/** Author: go-blog — 内容标签 API */

import { http } from '@/utils'

export function page(params: { current?: number; size?: number; name?: string }) {
  return http.get<{
    records: any[]
    total: string
    current: string
    size: string
  }>('/admin/content/tag/page', { params })
}

export function options() {
  return http.get<{ id: string; name: string; slug: string }[]>('/admin/content/tag/options')
}

export function detail(id: string) {
  return http.get<any>('/admin/content/tag/detail', { params: { id } })
}

export function create(body: { name: string; slug?: string | null; sort_code?: number }) {
  return http.post<any>('/admin/content/tag/create', body)
}

export function update(body: {
  id: string
  name: string
  slug?: string | null
  sort_code?: number
}) {
  return http.post<any>('/admin/content/tag/update', body)
}

export function remove(ids: string[]) {
  return http.post('/admin/content/tag/remove', { ids })
}
