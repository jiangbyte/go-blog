/** Author: go-blog */

import { http } from '@/utils'

/** mypage 角色 API */

export function list() {
  return http.get<any[]>('/admin/user/role/page')
}

/** 兼容旧 page 调用：全量列表当一页返回 */
export function page(_params?: any) {
  return list().then((res) => {
    const records = Array.isArray(res.data) ? res.data : []
    return {
      ...res,
      data: {
        records,
        total: String(records.length),
        page: '1',
        page_size: String(records.length || 20),
      },
    }
  })
}

export function create(data: { code: string; name: string; enabled?: string }) {
  return http.post<any>('/admin/user/role/create', data)
}

export function update(data: { role_id: string; name: string; enabled?: string }) {
  return http.post<any>('/admin/user/role/update', data)
}

export function changeEnabled(data: { role_id: string; enabled: string }) {
  return http.post<any>('/admin/user/role/change-enabled', data)
}

/** 返回 modules + grant_info_list */
export function ownResources(roleId: string, _accountType?: string) {
  return http.get<{ modules: any[]; grant_info_list: any[] }>('/admin/user/role/own-resource', {
    params: { role_id: roleId },
  })
}

/** 覆盖式授予角色资源 */
export function grantResources(data: {
  id?: string
  role_id?: string
  account_type?: string
  grant_info_list?: any[]
  resource_ids?: string[]
}) {
  return http.post<{ modules?: any[]; grant_info_list?: any[]; resource_ids?: string[] }>(
    '/admin/user/role/grant-resource',
    {
      role_id: data.role_id || data.id,
      id: data.id || data.role_id,
      grant_info_list: data.grant_info_list ?? [],
      resource_ids: data.resource_ids ?? [],
    },
  )
}
