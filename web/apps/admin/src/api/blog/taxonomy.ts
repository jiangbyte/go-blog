/** Author: go-blog — 博客分类 / 标签 / 项目标签 */

import { http } from '@/utils'
import type { TaxonomyKind } from './site'

export type { TaxonomyKind }

export function page(params: {
  kind: TaxonomyKind
  current?: number
  size?: number
  name?: string
}) {
  return http.get<{
    records: Array<{ uid: string; kind: string; name: string; created_at?: string }>
    total: string
    current: string
    size: string
  }>('/admin/blog/taxonomy/page', { params })
}

export function create(body: { kind: TaxonomyKind; name: string }) {
  return http.post<{ uid: string; kind: string; name: string }>('/admin/blog/taxonomy/create', body)
}

export function update(body: { uid: string; name: string }) {
  return http.post<{ uid: string; kind: string; name: string }>('/admin/blog/taxonomy/update', body)
}

export function remove(uids: string[]) {
  return http.post('/admin/blog/taxonomy/remove', { uids })
}
