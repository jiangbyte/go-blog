/** Author: go-blog */

/**
 * Web Admin 会话 token：与 Cookie 双通道。
 * - Cookie：服务端 HttpOnly `token`
 * - Header：本地 token 以 X-Token 发送
 */
export const tokenKey = 'token'

export function getToken(): string {
  const fromLocal = localStorage.getItem(tokenKey)
  if (fromLocal) {
    return fromLocal
  }
  return sessionStorage.getItem(tokenKey) || ''
}

/** rememberMe=true 用 localStorage，否则 sessionStorage（关页即丢）。 */
export function setToken(token: string, rememberMe = true) {
  const value = token?.trim()
  if (!value) {
    clearToken()
    return
  }
  if (rememberMe) {
    localStorage.setItem(tokenKey, value)
    sessionStorage.removeItem(tokenKey)
  } else {
    sessionStorage.setItem(tokenKey, value)
    localStorage.removeItem(tokenKey)
  }
}

export function clearToken() {
  localStorage.removeItem(tokenKey)
  sessionStorage.removeItem(tokenKey)
}
