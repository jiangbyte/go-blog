/** 与后端 R 信封 / 用户资料对齐的 snake_case 类型 */

export type ApiEnvelope<T> = {
  code: string
  message: string
  data: T
}

export type UserType = 'PORTAL' | 'ADMIN'
export type Gender = 'MALE' | 'FEMALE' | 'UNKNOWN'

export type UserProfile = {
  id: string
  username: string
  nickname: string
  phone: string | null
  email: string | null
  gender: string
  signature: string | null
  avatar: string | null
  user_type: string
  enabled: string
  create_time: string | null
  update_time: string | null
  permissions: string[]
}

export type AuthTokenData = {
  token: string
  profile: UserProfile
}

export type UserPage = {
  records: UserProfile[]
  total: string
  page: string
  page_size: string
}

export type RoleItem = {
  id: string
  code: string
  name: string
  enabled: string
}

export type PermissionItem = {
  id: string
  code: string
  name: string
  enabled: string
}

export type UserRoles = {
  user_id: string
  role_codes: string[]
}

export type PublicUser = {
  id: string
  username: string
  nickname: string
  gender: string
  signature: string | null
  user_type: string
}

export type RegisterBody = {
  username: string
  nickname?: string
  password_cipher: string
  captcha_token: string
  phone?: string | null
  email?: string | null
  gender?: Gender
  signature?: string | null
}

export type LoginBody = {
  username: string
  password_cipher: string
  captcha_token: string
  client_type: UserType
}

export type CaptchaCreateData = {
  captcha_id: string
  bg_url: string
  puzzle_url: string
  y: number
  bg_width: number
  bg_height: number
  puzzle_size: number
}

export type CaptchaVerifyBody = {
  captcha_id: string
  x: number
}

export type CaptchaVerifyData = {
  captcha_token: string
}

export type ForgotPasswordBody = {
  email: string
  captcha_token: string
  client_type: UserType
}

export type ForgotPasswordData = {
  message: string
  reset_url?: string | null
}

export type ResetPasswordBody = {
  token: string
  password_cipher: string
  captcha_token: string
}

export type AdminCreateUserBody = {
  username: string
  nickname?: string
  password_cipher: string
  user_type?: UserType
  phone?: string | null
  email?: string | null
  gender?: Gender
  signature?: string | null
}

export type AdminUpdateUserBody = {
  user_id: string
  nickname: string
  phone?: string | null
  email?: string | null
  gender?: Gender
  signature?: string | null
  enabled?: string | null
  password_cipher?: string | null
}

export type CreateRoleBody = {
  code: string
  name: string
  enabled?: string
}

export type UpdateRoleBody = {
  role_id: string
  name: string
  enabled?: string
}

export type CreatePermissionBody = {
  code: string
  name: string
  enabled?: string
}

export type UpdatePermissionBody = {
  permission_id: string
  name: string
  enabled?: string
}

export type RolePermissions = {
  role_id: string
  permission_codes: string[]
}
