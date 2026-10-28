/** Author: go-blog — 项目进度枚举（与后端 / 前台角标配色对齐） */

export type ProjectProgress = 'planned' | 'active' | 'done' | 'maintained' | 'paused'

/** NTag color：与 Blog 进度徽章一致 */
export type ProjectProgressTagColor = {
  color: string
  textColor: string
  borderColor: string
}

export const PROJECT_PROGRESS_OPTIONS: Array<{ label: string; value: ProjectProgress }> = [
  { label: '规划中', value: 'planned' },
  { label: '进行中', value: 'active' },
  { label: '已完成', value: 'done' },
  { label: '维护中', value: 'maintained' },
  { label: '已搁置', value: 'paused' },
]

const PROGRESS_TAG_COLOR: Record<ProjectProgress, ProjectProgressTagColor> = {
  planned: { color: '#6d5ce7', textColor: '#fff', borderColor: '#6d5ce7' },
  active: { color: '#2f6fed', textColor: '#fff', borderColor: '#2f6fed' },
  done: { color: '#2a9a6e', textColor: '#fff', borderColor: '#2a9a6e' },
  maintained: { color: '#1a96a8', textColor: '#fff', borderColor: '#1a96a8' },
  paused: { color: '#6b7280', textColor: '#fff', borderColor: '#6b7280' },
}

export function projectProgressLabel(value: string | null | undefined): string {
  const hit = PROJECT_PROGRESS_OPTIONS.find((o) => o.value === value)
  return hit?.label || '进行中'
}

export function normalizeProjectProgress(value: string | null | undefined): ProjectProgress {
  const v = String(value || '').trim()
  if (PROJECT_PROGRESS_OPTIONS.some((o) => o.value === v)) {
    return v as ProjectProgress
  }
  return 'active'
}

export function projectProgressTagColor(value: string | null | undefined): ProjectProgressTagColor {
  return PROGRESS_TAG_COLOR[normalizeProjectProgress(value)]
}
