/** Author: go-blog — 博客侧 Markdown 工具；渲染委托 @jiangbyte/markdown */

export { renderMarkdown } from '@jiangbyte/markdown'

export function formatDate(iso: string): string {
  return iso.slice(0, 10)
}

export function slugify(name: string): string {
  return encodeURIComponent(name)
}

export function coverOf(cover: string | undefined, fallbacks: string[], index = 0): string {
  if (cover) return cover
  return fallbacks[index % fallbacks.length] ?? 'https://t.alcy.cc/fj?u=default'
}
