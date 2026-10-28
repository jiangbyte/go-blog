/**
 * Author: go-blog — 简历模板清单（与 styles/templates/manifest.json 同源）
 *
 * 加皮肤：manifest 增一条 + templates/{css}；structure 仍用 sheet-v1。
 * 加结构：后端注册 STRUCTURE_BUILDERS 新族 + manifest.structure 指向该族；
 *         非 sheet-v1 预览走后端 HTML，无需再写 Vue/React DOM。
 */
import manifest from './styles/templates/manifest.json'

export const RESUME_STRUCTURE_SHEET_V1 = 'sheet-v1'

export type ResumeTemplateMeta = {
  id: string
  name: string
  description?: string
  defaultColor?: string
  structure: string
  /** 皮肤文件名；API 列表可省略，仅 manifest 完整条目含此字段 */
  css?: string
}

export const RESUME_TEMPLATE_MANIFEST = manifest as ResumeTemplateMeta[]

const BY_ID = new Map(RESUME_TEMPLATE_MANIFEST.map((t) => [t.id, t]))

/** 按 id 查模板元数据。 */
export function getResumeTemplateMeta(
  id: string | undefined | null,
): ResumeTemplateMeta | undefined {
  const key = (id || '').trim()
  if (!key) return undefined
  return BY_ID.get(key)
}

/** sheet-v1 可用前端活 DOM 预览；其它结构应走后端 HTML。 */
export function usesLiveDomPreview(templateId: string | undefined | null): boolean {
  const meta = getResumeTemplateMeta(templateId)
  return (meta?.structure || RESUME_STRUCTURE_SHEET_V1) === RESUME_STRUCTURE_SHEET_V1
}
