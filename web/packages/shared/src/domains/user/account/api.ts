import type { AxiosInstance } from 'axios'
import type {
  AdminCreateUserBody,
  AdminUpdateUserBody,
  ApiEnvelope,
  AuthTokenData,
  CaptchaCreateData,
  CaptchaVerifyBody,
  CaptchaVerifyData,
  ForgotPasswordBody,
  ForgotPasswordData,
  Gender,
  LoginBody,
  PublicUser,
  RegisterBody,
  ResetPasswordBody,
  UserPage,
  UserProfile,
  UserRoles,
} from './types'

/** user/account 领域 API（portal / admin 共用契约） */
export function createAccountApi(http: AxiosInstance) {
  return {
    getPublicKey() {
      return http.get<ApiEnvelope<{ public_key: string }>>('/auth/public-key', {
        public: true,
      })
    },
    createCaptcha() {
      return http.get<ApiEnvelope<CaptchaCreateData>>('/auth/captcha/create', {
        public: true,
      })
    },
    verifyCaptcha(body: CaptchaVerifyBody) {
      return http.post<ApiEnvelope<CaptchaVerifyData>>('/auth/captcha/verify', body, {
        public: true,
      })
    },
    register(body: RegisterBody) {
      return http.post<ApiEnvelope<AuthTokenData>>('/auth/register', body, {
        public: true,
      })
    },
    login(body: LoginBody) {
      return http.post<ApiEnvelope<AuthTokenData>>('/auth/login', body, {
        public: true,
      })
    },
    forgotPassword(body: ForgotPasswordBody) {
      return http.post<ApiEnvelope<ForgotPasswordData>>('/auth/password/forgot', body, {
        public: true,
      })
    },
    resetPassword(body: ResetPasswordBody) {
      return http.post<ApiEnvelope<null>>('/auth/password/reset', body, {
        public: true,
      })
    },
    logout() {
      return http.post<ApiEnvelope<null>>('/auth/logout')
    },
    me() {
      return http.get<ApiEnvelope<UserProfile>>('/auth/me')
    },
    updateProfile(body: { nickname: string; gender?: Gender; signature?: string | null }) {
      return http.post<ApiEnvelope<UserProfile>>('/auth/profile/update', {
        nickname: body.nickname,
        gender: body.gender || 'UNKNOWN',
        signature: body.signature ?? null,
      })
    },
    updatePassword(body: { old_password: string; new_password: string }) {
      return http.post<ApiEnvelope<null>>('/auth/profile/password/update', body)
    },
    uploadAvatar(file: File) {
      const form = new FormData()
      form.append('file', file)
      return http.post<ApiEnvelope<UserProfile>>('/auth/profile/avatar/upload', form)
    },
    publicUser(userId: string) {
      return http.get<ApiEnvelope<PublicUser>>('/portal/user/account/public', {
        params: { user_id: userId },
        public: true,
      })
    },
    listUsers(page = '1', pageSize = '20', filters?: { username?: string; enabled?: string }) {
      return http.get<ApiEnvelope<UserPage>>('/admin/user/account/page', {
        params: {
          page,
          page_size: pageSize,
          username: filters?.username || undefined,
          enabled: filters?.enabled || undefined,
        },
      })
    },
    userDetail(userId: string) {
      return http.get<ApiEnvelope<UserProfile>>('/admin/user/account/detail', {
        params: { user_id: userId },
      })
    },
    createUser(body: AdminCreateUserBody) {
      return http.post<ApiEnvelope<UserProfile>>('/admin/user/account/create', body)
    },
    updateUser(body: AdminUpdateUserBody) {
      return http.post<ApiEnvelope<UserProfile>>('/admin/user/account/update', body)
    },
    changeEnabled(userId: string, enabled: boolean) {
      return http.post<ApiEnvelope<UserProfile>>('/admin/user/account/change-enabled', {
        user_id: userId,
        enabled: enabled ? 'true' : 'false',
      })
    },
    userRoles(userId: string) {
      return http.get<ApiEnvelope<UserRoles>>('/admin/user/account/roles', {
        params: { user_id: userId },
      })
    },
  }
}

export type AccountApi = ReturnType<typeof createAccountApi>
