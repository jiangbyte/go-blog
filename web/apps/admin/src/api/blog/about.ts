/** Author: go-blog — 博客关于页配置 API */

import { http } from '@/utils'

export function getAbout() {
  return http.get<Record<string, any>>('/admin/blog/about')
}

export function updateAbout(body: Record<string, any>) {
  return http.post<Record<string, any>>('/admin/blog/about', body)
}
