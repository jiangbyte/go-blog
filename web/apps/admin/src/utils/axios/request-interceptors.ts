/** Author: go-blog */

import type { AxiosInstance } from 'axios'
import { getToken } from '@/utils/session'
import { stringifyScalars } from '@/utils/wire'

declare module 'axios' {
  interface AxiosRequestConfig {
    /** 公开接口：401 时不跳转登录。会话：Cookie（withCredentials）和/或 X-Token 头。 */
    public?: boolean
    skipErrorMessage?: boolean
    customErrorMessage?: string
  }
}

/** 按 wire 约定将 JSON body/params 标量序列化为字符串；附带本地会话 token。 */
export function setupRequestInterceptor(http: AxiosInstance) {
  http.interceptors.request.use((config) => {
    if (config.data && !(config.data instanceof FormData)) {
      config.data = stringifyScalars(config.data)
    }
    if (config.params) {
      config.params = stringifyScalars(config.params) as typeof config.params
    }
    config.headers = config.headers ?? {}
    // 禁止浏览器缓存 GET，避免列表/详情一直看到旧数据
    const method = String(config.method || 'get').toLowerCase()
    if (method === 'get' || method === 'head') {
      config.headers['Cache-Control'] = 'no-cache'
      config.headers.Pragma = 'no-cache'
    }
    const token = getToken()
    if (token) {
      // mypage：X-Token + Cookie token 双通道
      config.headers['X-Token'] = token
    }
    return config
  })
}
