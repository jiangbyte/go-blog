/** Author: go-blog — 简历主题色 / 字体 → CSS 变量（令牌来自 themeTokens.json，与后端同源） */

import themeTokens from './themeTokens.json'

export type ResumeThemeLike = {
  themeColor?: string
  layout?: string
  infoDisplay?: string
  language?: string
  fontFamily?: string
  fontSizeScale?: number
  spacing?: {
    moduleGap?: number
    lineHeight?: number
    padX?: number
    padTop?: number
    padBottom?: number
  }
}

const PRESETS = themeTokens.colors as Record<string, string>
const FONT_STACKS = themeTokens.fontStacks as Record<string, string>
const FONT_CDN = themeTokens.fontCdns as Record<string, string>

/** 解析主题色为 hex（预设名或 #hex）。 */
export function resolveThemeColor(color: string): string {
  const value = (color || '').trim()
  if (/^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/.test(value)) return value
  return PRESETS[value] ?? PRESETS.blue ?? '#3b82f6'
}

const DEFAULT_LINE_HEIGHT = 1.55

/** 行高钳制：0 / NaN / 过小会导致正文叠字。 */
export function sanitizeLineHeight(value: unknown): number {
  const n = Number(value)
  if (!Number.isFinite(n) || n < 1) return DEFAULT_LINE_HEIGHT
  if (n > 3) return 3
  return n
}

/** 字号缩放钳制。 */
export function sanitizeFontScale(value: unknown): number {
  const n = Number(value)
  if (!Number.isFinite(n) || n < 0.8) return 1
  if (n > 1.2) return 1.2
  return Math.round(n * 100) / 100
}

/** 字体族 id → CSS font-family。 */
export function resolveFontFamily(id: string | undefined): string {
  const key = (id || 'default').trim()
  return FONT_STACKS[key] ?? FONT_STACKS.default
}

/** 需要外链加载的字体 CDN；系统默认返回空。 */
export function fontStylesheetUrl(id: string | undefined): string | null {
  const key = (id || 'default').trim()
  return FONT_CDN[key] ?? null
}

/** theme_json → CSS 变量对象。 */
export function themeCssVars(theme: ResumeThemeLike | null | undefined): Record<string, string> {
  const t = theme ?? {}
  const primary = resolveThemeColor(t.themeColor || 'blue')
  const spacing = t.spacing ?? {}
  return {
    '--color-primary': primary,
    '--color-primary-soft': `${primary}1a`,
    '--color-text': '#111827',
    '--color-text-secondary': '#4b5563',
    '--color-border': '#e5e7eb',
    '--space-module': `${spacing.moduleGap ?? 12}px`,
    '--line-height': String(sanitizeLineHeight(spacing.lineHeight)),
    '--pad-x': `${spacing.padX ?? 48}px`,
    '--pad-top': `${spacing.padTop ?? 48}px`,
    '--pad-bottom': `${spacing.padBottom ?? 60}px`,
    '--font-family': resolveFontFamily(t.fontFamily),
    '--font-scale': String(sanitizeFontScale(t.fontSizeScale)),
  }
}

export { themeTokens }
