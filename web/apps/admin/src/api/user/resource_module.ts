/** Author: go-blog — 资源模块 API */

import { http } from '@/utils'

export function selector(_params?: any) {
  return http.get<any[]>('/admin/user/resource/modules/selector')
}

export function list() {
  return http.get<any[]>('/admin/user/resource/modules')
}
