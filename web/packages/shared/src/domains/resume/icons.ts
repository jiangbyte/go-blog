/** Author: go-blog — 简历纸面图标约定（与后端 resume_html 对齐） */

/** 纸面唯一图标集；其它前缀一律回退。 */
export const RESUME_ICON_PREFIX = 'lucide'
export const RESUME_ICON_FALLBACK = 'lucide:circle'

/**
 * 规范化纸面图标 id。
 * 规则与后端 `_icon_name` 一致：只认 lucide:xxx；空/其它集 → lucide:circle。
 */
export function resolveResumeIcon(raw?: string | null): string {
  const value = String(raw ?? '').trim()
  if (!value) return RESUME_ICON_FALLBACK
  const colon = value.indexOf(':')
  if (colon < 0) {
    const name = value.trim()
    return name ? `${RESUME_ICON_PREFIX}:${name}` : RESUME_ICON_FALLBACK
  }
  const prefix = value.slice(0, colon).trim().toLowerCase()
  const name = value.slice(colon + 1).trim()
  if (prefix !== RESUME_ICON_PREFIX || !name) return RESUME_ICON_FALLBACK
  return `${RESUME_ICON_PREFIX}:${name}`
}
