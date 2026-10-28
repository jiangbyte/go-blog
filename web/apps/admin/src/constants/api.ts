/** Author: go-blog */

/**
 * mypage Admin：请求直连 FastAPI 路径（/auth、/admin），无 /api/v1/admin 前缀。
 */
export const API_ROOT = ''
export const API_VERSION = ''
export const API_CLIENT = 'admin' as const

/** 空前缀：业务 API 自行拼 /auth、/admin/... */
export const API_PREFIX = ''

export function buildApiUrl(path: string) {
  const normalized = path.startsWith('/') ? path : `/${path}`
  return `${API_PREFIX}${normalized}`
}
