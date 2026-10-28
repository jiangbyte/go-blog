/** Author: go-blog — 联系字段分行（与后端 resume_html._balance_contact_rows 对齐） */

import type { ResumeProfileField } from './types'

/**
 * 按展示模式分行。
 * 图标模式每行最多 3 项，避免长邮箱把同行挤换行。
 */
export function balanceProfileRows(
  fields: ResumeProfileField[],
  infoDisplay: string = 'text',
): ResumeProfileField[][] {
  const n = fields.length
  if (n === 0) return []
  if (infoDisplay === 'icon') {
    if (n <= 3) return [fields]
    const size = 3
    const rows: ResumeProfileField[][] = []
    for (let i = 0; i < n; i += size) {
      rows.push(fields.slice(i, i + size))
    }
    return rows
  }
  if (n <= 4) return [fields]
  const rowCount = n <= 6 ? 2 : n <= 9 ? 3 : Math.ceil(n / 4)
  const size = Math.ceil(n / rowCount)
  const rows: ResumeProfileField[][] = []
  for (let i = 0; i < n; i += size) {
    rows.push(fields.slice(i, i + size))
  }
  return rows
}
