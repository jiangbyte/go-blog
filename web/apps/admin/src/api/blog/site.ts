/** Author: go-blog — 博客站点配置与分类选项 API */

import { http } from '@/utils'

export type TaxonomyKind = 'category' | 'tag' | 'project_tag'

export function getSite() {
  return http.get<Record<string, any>>('/admin/blog/site')
}

export function updateSite(body: Record<string, any>) {
  return http.post<Record<string, any>>('/admin/blog/site', body)
}

export function taxonomyOptions(kind: TaxonomyKind) {
  return http.get<Array<{ uid: string; name: string }>>('/admin/blog/taxonomy/options', {
    params: { kind },
  })
}
