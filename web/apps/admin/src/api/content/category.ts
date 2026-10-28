/** Author: go-blog — 内容分类 API */

import { http } from '@/utils'

export function tree() {
  return http.get<any[]>('/admin/content/category/tree')
}

export function detail(id: string) {
  return http.get<any>('/admin/content/category/detail', { params: { id } })
}

export function create(body: {
  name: string
  slug?: string | null
  parent_id?: string | null
  sort_code?: number
}) {
  return http.post<any>('/admin/content/category/create', body)
}

export function update(body: {
  id: string
  name: string
  slug?: string | null
  parent_id?: string | null
  sort_code?: number
}) {
  return http.post<any>('/admin/content/category/update', body)
}

export function remove(ids: string[]) {
  return http.post('/admin/content/category/remove', { ids })
}
