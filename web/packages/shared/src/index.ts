export { clearToken, getToken, setToken } from './auth/token'
export { encryptPassword } from './auth/rsa'
export {
  ApiError,
  createMypageHttp,
  createMypageHttp as createJiangbyteHttp,
  type CreateMypageHttpOptions,
} from './http/client'
export { mypageSeedToken, mypageTheme } from './theme/tokens'

export { ROLE_ADMIN, ROLE_PORTAL, hasPermission, type RoleCode } from './domains/user/rbac/codes'
export { createRbacApi, type RbacApi } from './domains/user/rbac/api'
export { createAccountApi, type AccountApi } from './domains/user/account/api'
export { createSiteApi, type SiteApi } from './domains/sys/site/api'
export { createStatsApi, type StatsApi } from './domains/sys/stats/api'
export { createResumeApi, type ResumeApi } from './domains/resume/api'
export { createContentApi, type ContentApi } from './domains/content/api'
export { createProjectApi, type ProjectApi } from './domains/project/api'
export type * from './domains/user/account/types'
export type * from './domains/sys/site/types'
export type * from './domains/sys/stats/types'
export type * from './domains/resume/types'
export type * from './domains/content/types'
export type * from './domains/project/types'
export {
  looksLikeHtml,
  markdownishToHtml,
  normalizeResumeHtml,
  toResumeHtml,
  toEditorHtml,
} from './domains/resume/bodyFormat'
export { isSectionVisible } from './domains/resume/types'
export { RESUME_ICON_FALLBACK, RESUME_ICON_PREFIX, resolveResumeIcon } from './domains/resume/icons'
export { balanceProfileRows } from './domains/resume/contactLayout'
export { resolveContactHref } from './domains/resume/contactHref'
export {
  fontStylesheetUrl,
  resolveFontFamily,
  resolveThemeColor,
  sanitizeFontScale,
  sanitizeLineHeight,
  themeCssVars,
  themeTokens,
  type ResumeThemeLike,
} from './domains/resume/themeTokens'
export {
  RESUME_STRUCTURE_SHEET_V1,
  RESUME_TEMPLATE_MANIFEST,
  getResumeTemplateMeta,
  usesLiveDomPreview,
  type ResumeTemplateMeta,
} from './domains/resume/templates'
