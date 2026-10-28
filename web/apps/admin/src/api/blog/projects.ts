/** Author: go-blog — 博客项目 API */

import { http } from '@/utils'

export function page(params: {
  current?: number
  size?: number
  title?: string
  status?: string
  progress?: string
}) {
  return http.get<{
    records: any[]
    total: string
    current: string
    size: string
  }>('/admin/blog/projects/page', { params })
}

export function detail(uid: string) {
  return http.get<any>('/admin/blog/projects/detail', { params: { uid } })
}

export function create(body: {
  title: string
  description?: string | null
  date?: string | null
  status?: 'draft' | 'published'
  progress?: string
  weight?: number
  sticky?: boolean
  cover?: string | null
  git?: string | null
  site?: string | null
  tag_uids?: string[]
  body: string
}) {
  return http.post<any>('/admin/blog/projects/create', body)
}

export function update(body: {
  uid: string
  title: string
  description?: string | null
  date?: string | null
  status?: 'draft' | 'published'
  progress?: string
  weight?: number
  sticky?: boolean
  cover?: string | null
  git?: string | null
  site?: string | null
  tag_uids?: string[]
  body: string
}) {
  return http.post<any>('/admin/blog/projects/update', body)
}

export function remove(uids: string[]) {
  return http.post('/admin/blog/projects/remove', { uids })
}

export function batchStatus(body: { uids: string[]; status: 'draft' | 'published' }) {
  return http.post<{
    updated: number
    skipped: number
    failed: Array<{ uid: string; title?: string; reason: string }>
  }>('/admin/blog/projects/batch-status', body)
}
