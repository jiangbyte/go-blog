import axios from 'axios'

/**
 * Axios 实例。
 * - `baseURL`：`VITE_API_BASE`，默认 `/api`
 * - 经 Vite 代理或网关访问 Gin 后端
 *
 * 后端统一响应包：
 * ```json
 * { "code": "0000", "data": {}, "message": "ok" }
 * ```
 * 兼容历史数字码 `0` / `200`。
 */
export const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 15000,
  withCredentials: true,
})

http.interceptors.response.use(
  (res) => res,
  (err) => {
    const message =
      err?.response?.data?.message || err?.message || '网络请求失败'
    return Promise.reject(new Error(message))
  },
)

function isOkCode(code: unknown) {
  return code === '0000' || code === 0 || code === 200 || code === '0' || code === '200'
}

/**
 * 统一请求入口（业务层只调这个）。
 */
export async function request(options: {
  url: string
  method?: string
  params?: any
  data?: any
}): Promise<any> {
  const res = await http.request({
    url: options.url,
    method: (options.method || 'get') as any,
    params: options.params,
    data: options.data,
  })
  const body = res.data
  if (body && typeof body === 'object' && 'code' in body && 'data' in body) {
    if (!isOkCode(body.code)) {
      throw new Error(body.message || `请求失败 code=${body.code}`)
    }
    return body.data
  }
  return body
}
