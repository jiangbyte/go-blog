/** 系统角色码（数量少且稳定，允许常量；权限码禁止全量枚举） */
export const ROLE_ADMIN = 'admin' as const
export const ROLE_PORTAL = 'portal' as const

export type RoleCode = typeof ROLE_ADMIN | typeof ROLE_PORTAL

/** 是否拥有全部所需权限（AND）。权限码用字面量传入。 */
export function hasPermission(owned: string[] | undefined | null, ...required: string[]): boolean {
  if (!required.length) return true
  const set = new Set(owned ?? [])
  return required.every((code) => set.has(code))
}
