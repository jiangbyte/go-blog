/** Author: go-blog — 联系字段 value → 超链接（与后端 resume_html._field_href 对齐） */

/** 仅按展示值推断 mailto / tel / https；无法识别则返回 undefined。 */
export function resolveContactHref(value: string | null | undefined): string | undefined {
  const text = String(value || '').trim()
  if (!text) return undefined
  // 1. 已是绝对 URL
  if (/^https?:\/\//i.test(text)) return text
  // 2. 邮箱
  if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(text)) return `mailto:${text}`
  // 3. 手机号（可选 +86 / 空格横线）
  let digits = text.replace(/[\s\-()]/g, '')
  if (/^\+?86/.test(digits)) digits = digits.replace(/^\+?86/, '')
  if (/^1\d{10}$/.test(digits)) return `tel:${digits}`
  // 4. 形如 domain/path
  if (
    /^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+(\/[\w\-./?%&=+#]*)?$/i.test(
      text,
    )
  ) {
    return `https://${text}`
  }
  return undefined
}
