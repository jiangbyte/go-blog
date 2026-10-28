/** Author: go-blog — 博客文章 API */

import { http } from '@/utils'

export function page(params: {
  current?: number
  size?: number
  title?: string
  status?: string
}) {
  return http.get<{
    records: any[]
    total: string
    current: string
    size: string
  }>('/admin/blog/posts/page', { params })
}

export function detail(uid: string) {
  return http.get<any>('/admin/blog/posts/detail', { params: { uid } })
}

export function create(body: {
  title: string
  description?: string | null
  date?: string | null
  status?: 'draft' | 'published'
  cover?: string | null
  category_uids?: string[]
  tag_uids?: string[]
  reprint?: boolean
  sticky?: boolean
  home?: boolean
  body: string
}) {
  return http.post<any>('/admin/blog/posts/create', body)
}

export function update(body: {
  uid: string
  title: string
  description?: string | null
  date?: string | null
  status?: 'draft' | 'published'
  cover?: string | null
  category_uids?: string[]
  tag_uids?: string[]
  reprint?: boolean
  sticky?: boolean
  home?: boolean
  body: string
}) {
  return http.post<any>('/admin/blog/posts/update', body)
}

export function remove(uids: string[]) {
  return http.post('/admin/blog/posts/remove', { uids })
}

export function batchStatus(body: { uids: string[]; status: 'draft' | 'published' }) {
  return http.post<{
    updated: number
    skipped: number
    failed: Array<{ uid: string; title?: string; reason: string }>
  }>('/admin/blog/posts/batch-status', body)
}
