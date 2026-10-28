/** Author: go-blog — 简历正文字符串：HTML 判定、结构规范化 */

const BLOCK_TAGS = new Set(['P', 'UL', 'OL', 'LI', 'DIV', 'BR'])

/** 是否已是 HTML（富文本存储）。 */
export function looksLikeHtml(src: string): boolean {
  return /^\s*</.test(String(src || '').trim())
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

/** 行内 **加粗** → <strong>。 */
function inlineMarkdown(text: string): string {
  const escaped = escapeHtml(text)
  return escaped.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
}

/**
 * 将近似 Markdown 的纯文本转为 HTML（仅兼容存量种子；新数据应直接存 HTML）。
 * 仅覆盖简历常用：段落、无序/有序列表、加粗。
 */
export function markdownishToHtml(src: string): string {
  const raw = String(src || '')
    .replace(/\r\n/g, '\n')
    .trim()
  if (!raw) return ''
  if (looksLikeHtml(raw)) return raw

  const lines = raw.split('\n')
  const parts: string[] = []
  let i = 0

  while (i < lines.length) {
    const line = lines[i]
    if (!line.trim()) {
      i += 1
      continue
    }

    if (/^\s*[-*]\s+/.test(line)) {
      const items: string[] = []
      while (i < lines.length && /^\s*[-*]\s+/.test(lines[i])) {
        const text = lines[i].replace(/^\s*[-*]\s+/, '').replace(/^\s*[-–—·•]+\s*/, '')
        items.push(`<li>${inlineMarkdown(text)}</li>`)
        i += 1
      }
      parts.push(`<ul>${items.join('')}</ul>`)
      continue
    }

    if (/^\s*\d+[.)]\s+/.test(line)) {
      const items: string[] = []
      while (i < lines.length && /^\s*\d+[.)]\s+/.test(lines[i])) {
        const text = lines[i]
          .replace(/^\s*\d+[.)]\s+/, '')
          .replace(/^\s*(?:\d+[.)、]|[-–—·•]+)\s*/, '')
        items.push(`<li>${inlineMarkdown(text)}</li>`)
        i += 1
      }
      parts.push(`<ol>${items.join('')}</ol>`)
      continue
    }

    const para: string[] = []
    while (
      i < lines.length &&
      lines[i].trim() &&
      !/^\s*[-*]\s+/.test(lines[i]) &&
      !/^\s*\d+[.)]\s+/.test(lines[i])
    ) {
      para.push(inlineMarkdown(lines[i].trim()))
      i += 1
    }
    parts.push(`<p>${para.join('<br>')}</p>`)
  }

  return parts.join('')
}

function unwrapNode(node: Element) {
  const parent = node.parentNode
  if (!parent) return
  while (node.firstChild) {
    parent.insertBefore(node.firstChild, node)
  }
  parent.removeChild(node)
}

function renameTag(el: Element, tagName: string): Element {
  const next = el.ownerDocument.createElement(tagName)
  while (el.firstChild) {
    next.appendChild(el.firstChild)
  }
  el.parentNode?.replaceChild(next, el)
  return next
}

function isEmptyBlock(el: Element): boolean {
  const text = (el.textContent || '').replace(/\u00a0/g, ' ').trim()
  if (text) return false
  if (el.querySelector('img, br')) {
    const html = el.innerHTML.replace(/<br\s*\/?>/gi, '').trim()
    return !html
  }
  return true
}

/**
 * 规范化 contenteditable 产出：去无用包装、统一标签，贴近简历纸面结构。
 * 需在浏览器环境调用（DOMParser）。
 */
export function normalizeResumeHtml(html: string): string {
  const raw = String(html || '').trim()
  if (!raw) return ''
  if (typeof DOMParser === 'undefined') return raw

  const doc = new DOMParser().parseFromString(`<div id="__resume_root">${raw}</div>`, 'text/html')
  const root = doc.getElementById('__resume_root')
  if (!root) return ''

  // 1. 深度优先处理子节点
  const walk = (parent: Element) => {
    const children = Array.from(parent.childNodes)
    for (const child of children) {
      if (child.nodeType === Node.TEXT_NODE) {
        if (parent.id === '__resume_root') {
          const text = (child.textContent || '').replace(/\u00a0/g, ' ')
          if (!text.trim()) {
            parent.removeChild(child)
            continue
          }
          const p = doc.createElement('p')
          p.textContent = text
          parent.replaceChild(p, child)
        }
        continue
      }
      if (child.nodeType !== Node.ELEMENT_NODE) {
        parent.removeChild(child)
        continue
      }
      let el = child as Element
      const tag = el.tagName

      if (tag === 'B') {
        el = renameTag(el, 'strong')
      } else if (tag === 'I') {
        el = renameTag(el, 'em')
      } else if (tag === 'DIV') {
        const hasBlock = Array.from(el.children).some((c) => BLOCK_TAGS.has(c.tagName))
        if (hasBlock) {
          unwrapNode(el)
          walk(parent)
          return
        }
        el = renameTag(el, 'p')
      } else if (tag === 'SPAN' || tag === 'FONT') {
        unwrapNode(el)
        walk(parent)
        return
      }

      for (const attr of Array.from(el.attributes)) {
        el.removeAttribute(attr.name)
      }

      if (el.tagName === 'P' || el.tagName === 'LI' || el.tagName === 'UL' || el.tagName === 'OL') {
        walk(el)
      } else if (el.tagName === 'STRONG' || el.tagName === 'EM' || el.tagName === 'BR') {
        // 行内保留
      } else {
        unwrapNode(el)
        walk(parent)
        return
      }
    }
  }

  walk(root)

  // 2. 去掉 <li> 开头重复的 - / · / 1.，避免与列表符叠床架屋
  for (const li of Array.from(root.querySelectorAll('li'))) {
    const html = li.innerHTML
    const cleaned = html.replace(
      /^(\s*(?:<(?:strong|b|em|span)[^>]*>)*)\s*(?:\d+[.)、]|[-–—·•]+)\s*/i,
      '$1',
    )
    if (cleaned !== html) li.innerHTML = cleaned
  }

  // 3. 删空段落
  for (const el of Array.from(root.querySelectorAll('p, li'))) {
    if (isEmptyBlock(el) && el.tagName === 'P') {
      el.remove()
    }
  }

  // 4. 空 ul/ol 删除
  for (const el of Array.from(root.querySelectorAll('ul, ol'))) {
    if (!el.querySelector('li')) el.remove()
  }

  const out = root.innerHTML.trim()
  if (!out || out === '<br>' || out === '<p><br></p>') return ''
  return out
}

/** 纯文本或 HTML → 规范化 HTML（纸面渲染与编辑器入载共用）。 */
export function toResumeHtml(raw: string): string {
  const text = String(raw || '')
  if (!text.trim()) return ''
  const html = looksLikeHtml(text) ? text : markdownishToHtml(text)
  return normalizeResumeHtml(html)
}

/** @deprecated 使用 toResumeHtml */
export const toEditorHtml = toResumeHtml
