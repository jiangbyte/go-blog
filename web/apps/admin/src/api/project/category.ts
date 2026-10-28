/** Author: go-blog — 项目分类 API */

import { http } from '@/utils'

export function tree() {
  return http.get<any[]>('/admin/project/category/tree')
}

export function detail(id: string) {
  return http.get<any>('/admin/project/category/detail', { params: { id } })
}

export function create(body: {
  name: string
  slug?: string | null
  parent_id?: string | null
  sort_code?: number
}) {
  return http.post<any>('/admin/project/category/create', body)
}

export function update(body: {
  id: string
  name: string
  slug?: string | null
  parent_id?: string | null
  sort_code?: number
}) {
  return http.post<any>('/admin/project/category/update', body)
}

export function remove(ids: string[]) {
  return http.post('/admin/project/category/remove', { ids })
}
