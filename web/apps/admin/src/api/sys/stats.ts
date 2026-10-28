/** Author: go-blog — 博客运营台概览统计 */

import { http } from '@/utils'

export type StatsOverview = {
  days: number
  counts: {
    post_published: number
    post_draft: number
    project_published: number
    project_draft: number
    category: number
    tag: number
    project_tag: number
    article?: number
    project?: number
  }
  period: { site_pv: number; site_uv: number; article_pv: number }
  total: { site_pv: number; site_uv: number; article_pv: number }
  traffic: Array<{ date: string; site_pv: number; site_uv: number; article_pv: number }>
  article: {
    pv_series: Array<{ date: string; count: number }>
    top: Array<{ id: string; uid?: string; title: string; view_count: number; slug: string }>
  }
  drafts?: Array<{ uid: string; title: string; updated_at: string }>
  generated_at?: string
}

export function overview(params: { days?: 7 | 14 | 30 } = {}) {
  return http.get<StatsOverview>('/admin/sys/stats/overview', { params })
}
