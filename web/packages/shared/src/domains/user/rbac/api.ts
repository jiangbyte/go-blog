import type { AxiosInstance } from 'axios'
import type {
  ApiEnvelope,
  CreatePermissionBody,
  CreateRoleBody,
  PermissionItem,
  RoleItem,
  UpdatePermissionBody,
  UpdateRoleBody,
  UserRoles,
} from '../account/types'

/** user/rbac 领域 API（角色授资源见 admin 端 resource/role API） */
export function createRbacApi(http: AxiosInstance) {
  return {
    listRoles() {
      return http.get<ApiEnvelope<RoleItem[]>>('/admin/user/role/page')
    },
    createRole(body: CreateRoleBody) {
      return http.post<ApiEnvelope<RoleItem>>('/admin/user/role/create', body)
    },
    updateRole(body: UpdateRoleBody) {
      return http.post<ApiEnvelope<RoleItem>>('/admin/user/role/update', body)
    },
    changeRoleEnabled(roleId: string, enabled: boolean) {
      return http.post<ApiEnvelope<RoleItem>>('/admin/user/role/change-enabled', {
        role_id: roleId,
        enabled: enabled ? 'true' : 'false',
      })
    },
    listPermissions() {
      return http.get<ApiEnvelope<PermissionItem[]>>('/admin/user/permission/page')
    },
    createPermission(body: CreatePermissionBody) {
      return http.post<ApiEnvelope<PermissionItem>>('/admin/user/permission/create', body)
    },
    updatePermission(body: UpdatePermissionBody) {
      return http.post<ApiEnvelope<PermissionItem>>('/admin/user/permission/update', body)
    },
    changePermissionEnabled(permissionId: string, enabled: boolean) {
      return http.post<ApiEnvelope<PermissionItem>>('/admin/user/permission/change-enabled', {
        permission_id: permissionId,
        enabled: enabled ? 'true' : 'false',
      })
    },
    bindUserRoles(userId: string, roleCodes: string[]) {
      return http.post<ApiEnvelope<UserRoles>>('/admin/user/account/bind-roles', {
        user_id: userId,
        role_codes: roleCodes,
      })
    },
    ownRoleResources(roleId: string) {
      return http.get<ApiEnvelope<{ role_id: string; resource_ids: string[] }>>(
        '/admin/user/role/own-resource',
        { params: { role_id: roleId } },
      )
    },
    grantRoleResources(roleId: string, resourceIds: string[]) {
      return http.post<ApiEnvelope<{ role_id: string; resource_ids: string[] }>>(
        '/admin/user/role/grant-resource',
        { role_id: roleId, resource_ids: resourceIds },
      )
    },
  }
}

export type RbacApi = ReturnType<typeof createRbacApi>
