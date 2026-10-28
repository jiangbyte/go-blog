/** Author: go-blog */

import { http } from '@/utils'

/** mypage 用户管理 API */

export function page(params: {
  page?: string | number
  page_size?: string | number
  username?: string
  enabled?: string
}) {
  return http.get<{
    records: any[]
    total: string
    page: string
    page_size: string
  }>('/admin/user/account/page', { params })
}

export function detail(params: { user_id: string }) {
  return http.get<any>('/admin/user/account/detail', { params })
}

export function create(data: Record<string, unknown>) {
  return http.post<any>('/admin/user/account/create', data)
}

export function update(data: Record<string, unknown>) {
  return http.post<any>('/admin/user/account/update', data)
}

export function changeEnabled(data: { user_id: string; enabled: string }) {
  return http.post<any>('/admin/user/account/change-enabled', data)
}

export function ownRoles(userId: string) {
  return http.get<{ user_id: string; role_codes: string[] }>('/admin/user/account/roles', {
    params: { user_id: userId },
  })
}

export function grantRoles(data: { user_id: string; role_codes: string[] }) {
  return http.post<{ user_id: string; role_codes: string[] }>(
    '/admin/user/account/bind-roles',
    data,
  )
}
