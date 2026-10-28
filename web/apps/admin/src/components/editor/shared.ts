/** Author: go-blog — 编辑器媒体上传与预览 URL 签发（不改动预览 DOM） */

import { fileApi } from '@/api'
import { normalizeUploadedFile } from '@/utils'
import type { NormalizedUploadedFile } from '@/utils'

export type EditorUploadedFile = NormalizedUploadedFile

export const MEDIA_PREFIXES = ['uploads/', 'avatars/', 'files/', 'resume/', 'content/'] as const
const resolvedUrlCache = new Map<string, string>()

const MD_IMG_RE = /!\[[^\]]*]\(\s*<?([^)\s>]+)>?(?:\s+["'][^"']*["'])?\s*\)/g
const HTML_IMG_SRC_RE = /\bsrc=(["'])([^"']+)\1/gi

export async function uploadEditorFile(file: File): Promise<EditorUploadedFile> {
  const response = await fileApi.upload(file)
  return normalizeUploadedFile(response.data, file, 'object_name')
}

export function toCssSize(value: string | number) {
  return typeof value === 'number' ? `${value}px` : value
}

function looksLikePresignedUrl(value: string) {
  const lower = value.toLowerCase()
  return (
    lower.includes('x-amz-') ||
    lower.includes('x-oss-') ||
    lower.includes('signature=') ||
    lower.includes('token=') ||
    lower.includes('x-qiniu-')
  )
}

function isStableHttpUrl(value: string) {
  return /^https?:\/\//i.test(value) && !looksLikePresignedUrl(value)
}

/** 从存储直链 / 预签名 URL 推导 object_name 候选。 */
function objectKeyCandidates(value: string): string[] {
  const raw = String(value || '').trim()
  if (!raw) return []
  const out: string[] = []
  const seen = new Set<string>()

  const add = (item?: string | null) => {
    const key = String(item || '')
      .replace(/\\/g, '/')
      .replace(/^\/+/, '')
      .trim()
    if (!key || seen.has(key)) return
    seen.add(key)
    out.push(key)
  }

  if (/^(https?:|data:|blob:)/i.test(raw)) {
    try {
      const path = new URL(raw).pathname.replace(/^\/+/, '')
      add(path)
      if (path.includes('/')) {
        const [, ...rest] = path.split('/')
        add(rest.join('/'))
      }
    } catch {
      /* ignore */
    }
    return out
  }

  add(raw)
  if (raw.includes('/')) {
    const [, ...rest] = raw.split('/')
    add(rest.join('/'))
  }
  return out
}

function pickObjectName(value: string): string {
  const candidates = objectKeyCandidates(value)
  const preferred = candidates.find((c) => MEDIA_PREFIXES.some((p) => c.startsWith(p)))
  return preferred || candidates[0] || ''
}

function isMediaObjectRef(src: string): boolean {
  const raw = String(src || '')
    .trim()
    .replace(/^\/+/, '')
  return MEDIA_PREFIXES.some((p) => raw.startsWith(p))
}

/** 从 Markdown / HTML 片段收集站内对象键。 */
export function collectEditorMediaRefs(text: string): string[] {
  const raw = String(text || '')
  if (!raw) return []
  const out: string[] = []
  const seen = new Set<string>()
  const add = (ref: string) => {
    const key = pickObjectName(ref)
    if (!key || !isMediaObjectRef(key) || seen.has(key)) return
    seen.add(key)
    out.push(key)
  }
  for (const match of raw.matchAll(MD_IMG_RE)) add(match[1] || '')
  for (const match of raw.matchAll(HTML_IMG_SRC_RE)) add(match[2] || '')
  return out
}

/**
 * 预览用：把 object_name / 过期预签名链解析为当前可访问 URL。
 * Markdown 源文仍写稳定 object_name，避免落库后裂图。
 */
export async function transformEditorImgUrl(src: string): Promise<string> {
  const raw = String(src || '').trim()
  if (!raw) return raw
  if (/^(data:|blob:)/i.test(raw)) return raw
  if (isStableHttpUrl(raw)) return raw

  const objectName = pickObjectName(raw)
  if (!objectName) return raw

  const cached = resolvedUrlCache.get(objectName)
  if (cached) return cached

  try {
    const res = await fileApi.url(objectName)
    const url = String(res.data?.url || '').trim()
    if (url) {
      resolvedUrlCache.set(objectName, url)
      return url
    }
  } catch {
    /* 解析失败时回退原值，避免预览空白 */
  }
  return raw
}

/** 预热缓存：编辑正文变化时批量签发，供 sanitize 同步替换 img src。 */
export async function warmEditorMediaUrls(text: string): Promise<number> {
  const refs = collectEditorMediaRefs(text)
  let filled = 0
  await Promise.all(
    refs.map(async (ref) => {
      if (resolvedUrlCache.has(ref)) return
      const url = await transformEditorImgUrl(ref)
      if (url && url !== ref) filled += 1
    }),
  )
  return filled
}

/**
 * sanitize 钩子内同步替换：只改 HTML 字符串里的 img src，不碰预览 DOM / mermaid。
 */
export function sanitizeEditorHtml(html: string): string {
  let out = String(html || '').replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, '')
  out = out.replace(/\bsrc=(["'])([^"']+)\1/gi, (full, quote: string, src: string) => {
    if (!isMediaObjectRef(src) && !looksLikePresignedUrl(src)) return full
    const objectName = pickObjectName(src)
    const resolved = objectName ? resolvedUrlCache.get(objectName) : undefined
    if (!resolved) return full
    return `src=${quote}${resolved}${quote}`
  })
  return out
}
