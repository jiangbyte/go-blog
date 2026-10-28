/** Author: go-blog — 管理端资源树 API */

import { http } from '@/utils'

export function tree(params?: { module_id?: string }) {
  return http.get<any[]>('/admin/user/resource/tree', { params })
}

export function current() {
  return http.get<any[]>('/admin/user/resource/current')
}

export function detail(params: { id: string }) {
  return http.get<any>('/admin/user/resource/detail', { params })
}

export function create(data: any) {
  return http.post<any>('/admin/user/resource/create', data)
}

export function update(data: any) {
  return http.post<any>('/admin/user/resource/update', data)
}

export function remove(data: { ids: string[] }) {
  return http.post('/admin/user/resource/delete', data)
}

export function permissionRegistry() {
  return http.get<any[]>('/admin/user/permission/registry')
}

export function buttonPage(params?: { parent_id: string; current?: number; size?: number }) {
  return http.get<{ records: any[]; total: string }>('/admin/user/resource/buttons/page', {
    params,
  })
}

export function buttonCreate(data: any) {
  return http.post<any>('/admin/user/resource/buttons/create', data)
}

export function buttonUpdate(data: any) {
  return http.post<any>('/admin/user/resource/buttons/update', data)
}

export function buttonRemove(data: { ids: string[] }) {
  return http.post('/admin/user/resource/buttons/delete', data)
}
