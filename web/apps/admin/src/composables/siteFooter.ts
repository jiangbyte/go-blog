/** Author: go-blog — 站点公开信息：仅来自公开接口 / 传入 props，不读 Vite env */

export const DEFAULT_APP_NAME = 'Demo Blog'

export interface SiteFooterInfo {
  appName: string
  appSubtitle: string
  copyrightText: string
  copyrightUrl: string
  icpNumber: string
  icpUrl: string
  psbNumber: string
  psbUrl: string
  homeProfileAvatar: string
  homeProfileName: string
  homeProfileTags: string[]
  homeProfileSignature: string
}

export const emptySiteFooter = (): SiteFooterInfo => ({
  appName: '',
  appSubtitle: '',
  copyrightText: '',
  copyrightUrl: '',
  icpNumber: '',
  icpUrl: '',
  psbNumber: '',
  psbUrl: '',
  homeProfileAvatar: '',
  homeProfileName: '',
  homeProfileTags: [],
  homeProfileSignature: '',
})

export function externalHref(url: string): string {
  const value = (url || '').trim()
  if (!value) return ''
  if (/^https?:\/\//i.test(value)) return value
  return `https://${value}`
}

/** 规范化个人标签列表。 */
function normalizeTags(raw: unknown): string[] {
  if (!Array.isArray(raw)) return []
  return raw
    .filter((item): item is string => typeof item === 'string')
    .map((item) => item.trim())
    .filter(Boolean)
}

/** 从 site-footer 接口响应解析站点信息。 */
export function parseSiteFooter(data: Record<string, unknown> | null | undefined): SiteFooterInfo {
  if (!data) return emptySiteFooter()
  const nested = (data.site_footer ?? data.siteFooter) as Record<string, unknown> | undefined
  const source = nested && typeof nested === 'object' ? nested : data
  return {
    appName: String(source.app_name ?? source.appName ?? '').trim(),
    appSubtitle: String(source.app_subtitle ?? source.appSubtitle ?? '').trim(),
    copyrightText: String(source.copyright_text ?? source.copyrightText ?? '').trim(),
    copyrightUrl: String(source.copyright_url ?? source.copyrightUrl ?? '').trim(),
    icpNumber: String(source.icp_number ?? source.icpNumber ?? '').trim(),
    icpUrl: String(source.icp_url ?? source.icpUrl ?? '').trim(),
    psbNumber: String(source.psb_number ?? source.psbNumber ?? '').trim(),
    psbUrl: String(source.psb_url ?? source.psbUrl ?? '').trim(),
    homeProfileAvatar: String(source.home_profile_avatar ?? source.homeProfileAvatar ?? '').trim(),
    homeProfileName: String(source.home_profile_name ?? source.homeProfileName ?? '').trim(),
    homeProfileTags: normalizeTags(source.home_profile_tags ?? source.homeProfileTags),
    homeProfileSignature: String(
      source.home_profile_signature ?? source.homeProfileSignature ?? '',
    ).trim(),
  }
}

export function hasSiteFooterContent(footer: SiteFooterInfo): boolean {
  return Boolean(footer.copyrightText || footer.icpNumber || footer.psbNumber)
}

/** 展示用应用名：接口为空时回退默认名。 */
export function resolveAppName(name: string | null | undefined): string {
  const value = (name || '').trim()
  return value || DEFAULT_APP_NAME
}
