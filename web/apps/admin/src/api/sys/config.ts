/** Author: go-blog */

import { http } from '@/utils'

const configPrefix = '/admin/sys/config'

export function page(params: any) {
  return http.get<any>(`${configPrefix}/page`, {
    params,
  })
}

export function list(params: { category?: string; scope?: string }) {
  return http.get<any[]>(`${configPrefix}/list`, { params })
}

export function detail(params: any) {
  return http.get<any>(`${configPrefix}/detail`, {
    params,
  })
}

export function create(data: any) {
  return http.post<any>(`${configPrefix}/create`, data)
}

export function update(data: any) {
  return http.post<any>(`${configPrefix}/update`, data)
}

export function remove(data: any) {
  return http.post<any>(`${configPrefix}/delete`, data)
}

export function batchSave(data: {
  items: Array<{
    config_key: string
    config_value: string | null
    category?: string | null
    remark?: string | null
    value_type?: string | null
    label?: string | null
    scope?: string | null
    scene?: string | null
    is_builtin?: boolean | null
  }>
}) {
  return http.post<any>(`${configPrefix}/batch-save`, data)
}
