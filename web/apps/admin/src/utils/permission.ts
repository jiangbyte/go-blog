/** Author: go-blog */

import { useAuthStore } from '@/stores'

/** 单管理员控制台：拥有 admin:console:access 即放行各业务按钮权限。 */
export function hasPermission(permissionKey: string) {
  const authStore = useAuthStore()
  if (authStore.hasPermission('admin:console:access')) {
    return true
  }
  return authStore.hasPermission(permissionKey)
}
