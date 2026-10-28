import axios, { type AxiosInstance, type CreateAxiosDefaults } from 'axios'
import { clearToken, getToken } from '../auth/token'

export class ApiError extends Error {
  code: string
  constructor(code: string, message: string) {
    super(message)
    this.code = code
    this.name = 'ApiError'
  }
}

export type CreateMypageHttpOptions = {
  /** 401 时回调（如跳转登录）；公开接口不会触发 */
  onUnauthorized?: () => void
  /** axios 额外配置 */
  axiosConfig?: CreateAxiosDefaults
}

declare module 'axios' {
  export interface AxiosRequestConfig {
    /** 公开接口：业务/HTTP 401 不触发登出跳转 */
    public?: boolean
  }
}

/**
 * 创建 HTTP 客户端：
 * - baseURL `/api`（Vite 代理去掉前缀到后端）
 * - Header `X-Token` + withCredentials（Cookie）
 * - 解包 R 信封，业务码非 0000 抛 ApiError
 */
export function createMypageHttp(options: CreateMypageHttpOptions = {}): AxiosInstance {
  const http = axios.create({
    baseURL: '/api',
    timeout: 30000,
    withCredentials: true,
    ...options.axiosConfig,
  })

  http.interceptors.request.use((config) => {
    const token = getToken()
    if (token) {
      config.headers = config.headers ?? {}
      config.headers['X-Token'] = token
    }
    return config
  })

  http.interceptors.response.use(
    (response) => {
      const payload = response.data
      if (payload && typeof payload === 'object' && 'code' in payload) {
        if (payload.code !== '0000') {
          const isPublic = Boolean(response.config.public)
          if (payload.code === '4010' && !isPublic) {
            clearToken()
            options.onUnauthorized?.()
          }
          throw new ApiError(String(payload.code), payload.message || '请求失败')
        }
      }
      return response
    },
    (error) => {
      const isPublic = Boolean(error?.config?.public)
      const status = error?.response?.status
      if (status === 401 && !isPublic) {
        clearToken()
        options.onUnauthorized?.()
      }
      const data = error?.response?.data
      if (data?.code && data?.message) {
        throw new ApiError(String(data.code), data.message)
      }
      throw error
    },
  )

  return http
}
