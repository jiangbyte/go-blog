/** Author: go-blog — 项目管理 API */

import { http } from '@/utils'

export function page(params: {
  current?: number
  size?: number
  name?: string
  status?: string
  visibility?: string
  category_id?: string
}) {
  return http.get<{
    records: any[]
    total: string
    current: string
    size: string
  }>('/admin/project/project/page', { params })
}

export function detail(id: string) {
  return http.get<any>('/admin/project/project/detail', { params: { id } })
}

export function create(body: {
  name: string
  slug?: string | null
  tech_stack?: string[]
  role?: string | null
  description?: string
  images?: string[]
  start_date?: string | null
  end_date?: string | null
  status?: string
  visibility?: string
  category_id?: string | null
  sort_code?: number
}) {
  return http.post<any>('/admin/project/project/create', body)
}

export function update(body: {
  id: string
  name: string
  slug?: string | null
  tech_stack?: string[]
  role?: string | null
  description?: string
  images?: string[]
  start_date?: string | null
  end_date?: string | null
  status?: string
  visibility?: string
  category_id?: string | null
  sort_code?: number
}) {
  return http.post<any>('/admin/project/project/update', body)
}

export function remove(ids: string[]) {
  return http.post('/admin/project/project/remove', { ids })
}
