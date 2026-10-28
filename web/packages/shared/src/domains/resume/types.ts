export type ResumeProfileField = {
  key: string
  label: string
  value: string
  /** 信息展示为「图标」时使用；空则回退 lucide:circle */
  icon?: string
}

export type ResumeProfile = {
  name: string
  tagline: string
  avatar_url: string
  github_url: string
  fields: ResumeProfileField[]
}

export type ResumeThemeSpacing = {
  moduleGap: number
  lineHeight: number
  padX: number
  padTop: number
  padBottom: number
}

export type ResumeFontFamily =
  'default' | 'source-sans' | 'source-serif' | 'wenkai' | 'inter-sans' | string

export type ResumeTheme = {
  themeColor: string
  layout: 'left' | 'center' | 'tiled' | string
  infoDisplay: 'text' | 'icon' | 'plain' | string
  language: 'zh' | 'en' | string
  fontFamily: ResumeFontFamily
  /** 字号缩放，1 = 100% */
  fontSizeScale: number
  spacing: ResumeThemeSpacing
}

export type ResumeEntry = {
  title: string
  role: string
  period: string
  body_html: string
  tags?: string[]
}

export type ResumeSection = {
  id: string
  resume_id: string
  type: string
  title: string
  icon: string
  sort_code: string
  /** meta_json.visible === false 时纸面不展示；缺省为显示 */
  meta_json: Record<string, unknown> & { visible?: boolean }
  body_html: string
  entries_json: ResumeEntry[]
}

/** 分节是否在纸面展示；仅 visible === false 时隐藏。 */
export function isSectionVisible(section: { meta_json?: Record<string, unknown> | null }): boolean {
  return section?.meta_json?.visible !== false
}

export type ResumeDocument = {
  id: string
  title: string
  template_id: string
  theme_json: ResumeTheme
  custom_css: string
  profile_json: ResumeProfile
  status: string
  published_at: string | null
  sort_code: string
  sections?: ResumeSection[]
}

export type ResumeTemplateMeta = {
  id: string
  name: string
  description: string
  defaultColor?: string
}
