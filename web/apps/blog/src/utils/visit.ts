/** Author: go-blog — 前台 PV/UV 上报（localStorage 访客 ID） */

const VISITOR_KEY = 'jb_vid'

function randomHex32(): string {
  const bytes = new Uint8Array(16)
  crypto.getRandomValues(bytes)
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('')
}

/** 稳定访客 ID（32 hex），用于 UV 去重。 */
export function getVisitorId(): string {
  try {
    const existing = String(localStorage.getItem(VISITOR_KEY) || '').trim()
    if (/^[a-f0-9]{32}$/i.test(existing)) return existing.toLowerCase()
    const next = randomHex32()
    localStorage.setItem(VISITOR_KEY, next)
    return next
  } catch {
    return randomHex32()
  }
}

function uidFromParam(params?: Record<string, any>): string {
  const fromParam = String(params?.uid || '').trim()
  return /^[a-f0-9]{32}$/i.test(fromParam) ? fromParam.toLowerCase() : ''
}

/** 从路由提取文章 uid（仅 `/p/:uid`）。 */
export function postUidFromRoute(path: string, params?: Record<string, any>): string {
  const fromParam = uidFromParam(params)
  if (fromParam && /^\/p\//.test(path)) return fromParam
  const m = path.match(/^\/p\/([a-f0-9]{32})\/?$/i)
  return m ? m[1].toLowerCase() : ''
}

/** 从路由提取项目 uid（仅 `/projects/:uid`，排除列表与标签页）。 */
export function projectUidFromRoute(path: string, params?: Record<string, any>): string {
  const fromParam = uidFromParam(params)
  if (fromParam && /^\/projects\/[a-f0-9]{32}/i.test(path)) return fromParam
  const m = path.match(/^\/projects\/([a-f0-9]{32})\/?$/i)
  return m ? m[1].toLowerCase() : ''
}
