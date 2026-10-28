/**
 * 轻量 SEO：写 document title / meta / canonical / JSON-LD，不引入额外依赖。
 */

export type SeoInput = {
  title?: string
  description?: string
  image?: string
  url?: string
  type?: 'website' | 'article'
  noindex?: boolean
  jsonLd?: Record<string, unknown> | Record<string, unknown>[] | null
}

const META_ATTR = 'data-jb-seo'
const LD_ATTR = 'data-jb-seo-ld'

function upsertMeta(attr: 'name' | 'property', key: string, content: string) {
  if (!content) return
  let el = document.head.querySelector<HTMLMetaElement>(
    `meta[${attr}="${key}"][${META_ATTR}]`,
  )
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute(attr, key)
    el.setAttribute(META_ATTR, '1')
    document.head.appendChild(el)
  }
  el.content = content
}

function upsertLink(rel: string, href: string) {
  if (!href) return
  let el = document.head.querySelector<HTMLLinkElement>(`link[rel="${rel}"][${META_ATTR}]`)
  if (!el) {
    el = document.createElement('link')
    el.rel = rel
    el.setAttribute(META_ATTR, '1')
    document.head.appendChild(el)
  }
  el.href = href
}

function setJsonLd(data: SeoInput['jsonLd']) {
  document.head.querySelectorAll(`script[${LD_ATTR}]`).forEach((n) => n.remove())
  if (!data) return
  const list = Array.isArray(data) ? data : [data]
  for (const item of list) {
    const el = document.createElement('script')
    el.type = 'application/ld+json'
    el.setAttribute(LD_ATTR, '1')
    el.textContent = JSON.stringify(item)
    document.head.appendChild(el)
  }
}

/** 应用一页 SEO；空字段会尽量保留站点默认而不强行清空。 */
export function applySeo(input: SeoInput) {
  const title = (input.title || '').trim()
  if (title) document.title = title

  if (input.description != null) {
    upsertMeta('name', 'description', input.description.trim())
    upsertMeta('property', 'og:description', input.description.trim())
    upsertMeta('name', 'twitter:description', input.description.trim())
  }
  if (title) {
    upsertMeta('property', 'og:title', title)
    upsertMeta('name', 'twitter:title', title)
  }
  const type = input.type || 'website'
  upsertMeta('property', 'og:type', type)
  upsertMeta('name', 'twitter:card', input.image ? 'summary_large_image' : 'summary')

  if (input.image) {
    upsertMeta('property', 'og:image', input.image)
    upsertMeta('name', 'twitter:image', input.image)
  }
  if (input.url) {
    upsertMeta('property', 'og:url', input.url)
    upsertLink('canonical', input.url)
  }
  upsertMeta('name', 'robots', input.noindex ? 'noindex,follow' : 'index,follow')
  setJsonLd(input.jsonLd ?? null)
}

/** 当前页绝对 URL（SPA 分享/canonical 用）。 */
export function pageAbsoluteUrl(path?: string): string {
  const origin = typeof window !== 'undefined' ? window.location.origin : ''
  const p = path ?? (typeof window !== 'undefined' ? window.location.pathname : '/')
  if (!p.startsWith('/')) return `${origin}/${p}`
  return `${origin}${p}`
}
