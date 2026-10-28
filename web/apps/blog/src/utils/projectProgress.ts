/** 项目进度展示（与后端 progress 枚举对齐） */

const LABELS: Record<string, string> = {
  planned: '规划中',
  active: '进行中',
  done: '已完成',
  maintained: '维护中',
  paused: '已搁置',
}

export function projectProgressLabel(value: string | null | undefined): string {
  const key = String(value || 'active').trim()
  return LABELS[key] || LABELS.active
}

/** CSS 修饰类后缀：progress--active 等 */
export function projectProgressMod(value: string | null | undefined): string {
  const key = String(value || 'active').trim()
  return LABELS[key] ? key : 'active'
}
