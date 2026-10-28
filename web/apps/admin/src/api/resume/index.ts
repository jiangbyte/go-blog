/** Author: go-blog — 简历管理 API */

import { http } from '@/utils'

export function page(params: { current?: number; size?: number; status?: string; title?: string }) {
  return http.get<{
    records: any[]
    total: string
    current: string
    size: string
  }>('/admin/resume/resume/page', { params })
}

export function detail(id: string) {
  return http.get<any>('/admin/resume/resume/detail', { params: { id } })
}

export function create(body: Record<string, unknown>) {
  return http.post<any>('/admin/resume/resume/create', body)
}

export function update(body: Record<string, unknown>) {
  return http.post<any>('/admin/resume/resume/update', body)
}

export function remove(ids: string[]) {
  return http.post('/admin/resume/resume/remove', { ids })
}

export function publish(id: string) {
  return http.post<any>('/admin/resume/resume/publish', { id })
}

export function unpublish(id: string) {
  return http.post<any>('/admin/resume/resume/unpublish', { id })
}

export function templates() {
  return http.get<{ id: string; name: string; description: string }[]>(
    '/admin/resume/resume/templates',
  )
}

export function copy(id: string) {
  return http.post<any>('/admin/resume/resume/copy', { id })
}

export function saveSection(body: Record<string, unknown>) {
  return http.post<any>('/admin/resume/resume/section/save', body)
}

export function copySection(resumeId: string, sectionId: string) {
  return http.post<any>('/admin/resume/resume/section/copy', {
    resume_id: resumeId,
    section_id: sectionId,
  })
}

export function reorderSections(resumeId: string, ids: string[]) {
  return http.post<any[]>('/admin/resume/resume/section/reorder', {
    resume_id: resumeId,
    ids,
  })
}

export function removeSections(resumeId: string, ids: string[]) {
  return http.post('/admin/resume/resume/section/remove', {
    resume_id: resumeId,
    ids,
  })
}

/** 后端导出 PDF（真实文字层，可选中）。 */
export function exportPdf(id: string) {
  return http.get<Blob>('/admin/resume/resume/export-pdf', {
    params: { id },
    responseType: 'blob',
  })
}

/** 文档 → 纸面 HTML（与 PDF 同源）。 */
export function renderHtml(body: {
  title?: string
  template_id?: string
  theme_json?: Record<string, unknown>
  custom_css?: string | null
  profile_json?: Record<string, unknown>
  sections?: unknown[]
}) {
  return http.post<{ html: string }>('/admin/resume/resume/render-html', body)
}

/** 文档 → PDF（chromedp，与下载分页一致；编辑态预览用）。 */
export function renderPdf(body: {
  title?: string
  template_id?: string
  theme_json?: Record<string, unknown>
  custom_css?: string | null
  profile_json?: Record<string, unknown>
  sections?: unknown[]
}) {
  return http.post<Blob>('/admin/resume/resume/render-pdf', body, {
    responseType: 'blob',
    timeout: 60000,
  })
}

/** 分页查询纸面可用图标（与后端渲染同源）。 */
export function pageIcons(params?: {
  prefix?: string
  current?: number
  size?: number
  q?: string
}) {
  return http.get<{
    records: { name: string; id: string }[]
    total: string
    current: string
    size: string
  }>('/admin/resume/resume/icons', { params })
}
