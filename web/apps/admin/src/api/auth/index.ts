/** Author: go-blog */

import { http } from '@/utils'

/** mypage 鉴权：/auth/* */

export function publicKey() {
  return http.get<{ public_key: string }>('/auth/public-key', {
    public: true,
  })
}

/** @deprecated 兼容旧调用名，等同 publicKey */
export function passwordKey() {
  return publicKey()
}

export function createCaptcha() {
  return http.get<{
    captcha_id: string
    bg_url: string
    puzzle_url: string
    y: number
    bg_width: number
    bg_height: number
    puzzle_size: number
  }>('/auth/captcha/create', { public: true })
}

export function verifyCaptcha(data: { captcha_id: string; x: number }) {
  return http.post<{ captcha_token: string }>('/auth/captcha/verify', data, {
    public: true,
  })
}

export function login(data: {
  username: string
  password_cipher: string
  captcha_token: string
  client_type?: string
}) {
  return http.post<{
    profile: Record<string, unknown>
  }>(
    '/auth/login',
    {
      username: data.username,
      password_cipher: data.password_cipher,
      captcha_token: data.captcha_token,
      client_type: data.client_type || 'ADMIN',
    },
    {
      public: true,
    },
  )
}

export function forgotPassword(data: {
  email: string
  captcha_token: string
  client_type?: string
}) {
  return http.post<{ message: string; reset_url?: string | null }>(
    '/auth/password/forgot',
    {
      email: data.email,
      captcha_token: data.captcha_token,
      client_type: data.client_type || 'ADMIN',
    },
    { public: true },
  )
}

export function resetPassword(data: {
  token: string
  password_cipher: string
  captcha_token: string
}) {
  return http.post<null>('/auth/password/reset', data, { public: true })
}

export function logout() {
  return http.post<null>('/auth/logout')
}

export function me() {
  return http.get<Record<string, unknown>>('/auth/me')
}

export function updateProfile(data: {
  nickname?: string | null
  gender?: string | null
  signature?: string | null
}) {
  return http.post<Record<string, unknown>>('/auth/profile/update', {
    nickname: data.nickname ?? '',
    gender: data.gender || 'UNKNOWN',
    signature: data.signature ?? null,
  })
}

export function updatePassword(data: { old_password?: string; new_password: string }) {
  return http.post<null>('/auth/profile/password/update', {
    old_password: data.old_password || '',
    new_password: data.new_password,
  })
}

export function uploadAvatar(file: File) {
  const form = new FormData()
  form.append('file', file)
  return http.post<Record<string, unknown>>('/auth/profile/avatar/upload', form)
}

/** @deprecated 字图验证码已废弃，保留空实现以免旧引用编译失败 */
export function captcha() {
  return Promise.resolve({ data: { captcha_id: '', image_base64: '' } }) as any
}
