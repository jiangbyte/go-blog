/** Author: go-blog */

import { defineStore } from 'pinia'
import { router } from '@/router'
import { authApi } from '@/api'
import { clearDict, refreshDict, syncDictTree } from '@/utils/dict'
import { clearToken } from '@/utils/session'
import { useRouteStore } from './route'
import { useTabStore } from './tab'

/** 对齐 mypage UserProfile 的精简会话用户 */
interface AuthUserInfo {
  accountId: string
  account: string
  accountType: string
  nickname?: string | null
  avatar?: string | null
  gender?: string | null
  signature?: string | null
  permissionKeys: string[]
  loginAt: number
}

interface AuthState {
  userInfo: AuthUserInfo | null
  sessionChecked: boolean
}

const userInfoKey = 'user_info'
const loginPath = '/login'

function getStoredUserInfo() {
  const raw = localStorage.getItem(userInfoKey)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as AuthUserInfo
  } catch {
    localStorage.removeItem(userInfoKey)
    return null
  }
}

function getSafeRedirect(redirect?: string) {
  const home = import.meta.env.VITE_HOME_PATH
  const value = String(redirect ?? '').trim()
  if (!value) {
    return home
  }
  if (
    !value.startsWith('/') ||
    value.startsWith('//') ||
    value.startsWith('/login') ||
    value.startsWith('/auth')
  ) {
    return home
  }
  if (value.includes('://')) {
    return home
  }
  return value
}

function mapProfile(data: Record<string, any>): AuthUserInfo {
  return {
    accountId: String(data.id ?? data.account_id ?? ''),
    account: String(data.username ?? data.account ?? ''),
    accountType: String(data.user_type ?? data.account_type ?? 'ADMIN'),
    nickname: data.nickname ?? null,
    avatar: data.avatar ?? null,
    gender: data.gender ?? 'UNKNOWN',
    signature: data.signature ?? null,
    permissionKeys: Array.isArray(data.permissions)
      ? data.permissions.map(String)
      : Array.isArray(data.permission_keys)
        ? data.permission_keys.map(String)
        : [],
    loginAt: Date.now(),
  }
}

export function resolveSecurityWallPath(_user: AuthUserInfo | null | undefined): string | null {
  return null
}

export function isAllowedUnderSecurityWall(
  _path: string,
  _queryTab: string | undefined | null,
  _user: AuthUserInfo | null | undefined,
): boolean {
  return true
}

export const useAuthStore = defineStore('auth-store', {
  state: (): AuthState => ({
    userInfo: getStoredUserInfo(),
    sessionChecked: false,
  }),
  getters: {
    isLogin: (state) => Boolean(state.userInfo?.accountId),
  },
  actions: {
    async ensureSession() {
      if (this.sessionChecked) {
        return this.isLogin
      }
      this.sessionChecked = true
      try {
        await this.refreshUserInfo()
        return true
      } catch {
        this.clearAuthStorage()
        return false
      }
    },

    /**
     * 账号密码登录（已 RSA 加密的 password_cipher）。
     */
    async login(
      username: string,
      passwordCipher: string,
      redirect?: string,
      rememberMe?: boolean,
      captchaToken?: string,
    ) {
      const response = await authApi.login({
        username,
        password_cipher: passwordCipher,
        captcha_token: captchaToken || '',
        client_type: 'ADMIN',
      })

      // 会话走 HttpOnly Cookie（withCredentials）；清掉旧的本地 token，避免 XSS 可读副本
      clearToken()
      void rememberMe

      this.sessionChecked = true

      // 优先用 login 返回的 profile，减少一次 me 往返
      if (response.data?.profile && typeof response.data.profile === 'object') {
        const userInfo = mapProfile(response.data.profile as Record<string, any>)
        userInfo.loginAt = Date.now()
        localStorage.setItem(userInfoKey, JSON.stringify(userInfo))
        this.userInfo = userInfo
      }

      await this.finishLogin(redirect)
    },

    async finishLogin(redirect?: string) {
      if (!this.userInfo?.accountId) {
        await this.refreshUserInfo()
      }

      const routeStore = useRouteStore()
      await routeStore.initAuthRoute()
      try {
        syncDictTree()
        await refreshDict()
      } catch {
        // mypage 无字典服务，忽略
      }
      await router.push(getSafeRedirect(redirect))
    },

    async refreshUserInfo() {
      const meResponse = await authApi.me()
      const data = (meResponse.data ?? {}) as Record<string, any>
      const userInfo = mapProfile(data)
      userInfo.loginAt = this.userInfo?.loginAt ?? Date.now()

      localStorage.setItem(userInfoKey, JSON.stringify(userInfo))
      this.userInfo = userInfo
      return data
    },

    hasPermission(permissionKey: string) {
      const keys = this.userInfo?.permissionKeys ?? []
      return keys.includes(permissionKey)
    },

    clearAuthStorage() {
      clearToken()
      localStorage.removeItem(userInfoKey)
      this.userInfo = null
    },

    resetSession() {
      this.clearAuthStorage()
      this.sessionChecked = true

      const routeStore = useRouteStore()
      routeStore.resetRouteStore()

      const tabStore = useTabStore()
      tabStore.clearAllTabs()

      clearDict()
    },

    async logout(redirect?: string) {
      const currentRoute = router.currentRoute.value
      const finalRedirect = redirect ?? currentRoute.fullPath

      try {
        await authApi.logout()
      } catch {
        // 后端登出失败不阻塞本地会话清理。
      } finally {
        this.resetSession()
      }

      await router.push({
        path: loginPath,
        query:
          finalRedirect.startsWith('/auth') || finalRedirect.startsWith('/login')
            ? undefined
            : { redirect: finalRedirect },
      })
    },
  },
})
