/**
 * Author: go-blog — 文章 Markdown 渲染：
 * - code → figure.code-block (+ hl_lines)
 * - mermaid fence → div.mermaid
 * - KaTeX for \( \) \[ \] $$ $$
 */
import hljs from 'highlight.js/lib/core'
import bash from 'highlight.js/lib/languages/bash'
import go from 'highlight.js/lib/languages/go'
import javascript from 'highlight.js/lib/languages/javascript'
import json from 'highlight.js/lib/languages/json'
import python from 'highlight.js/lib/languages/python'
import typescript from 'highlight.js/lib/languages/typescript'
import xml from 'highlight.js/lib/languages/xml'
import yaml from 'highlight.js/lib/languages/yaml'
import katex from 'katex'
import MarkdownIt from 'markdown-it'
import markdownItAnchor from 'markdown-it-anchor'
import markdownItFootnote from 'markdown-it-footnote'

hljs.registerLanguage('go', go)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('yml', yaml)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('shell', bash)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('json', json)
hljs.registerLanguage('js', javascript)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('ts', typescript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('py', python)

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function slugifyHeading(s: string): string {
  return (
    s
      .trim()
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^\w\u4e00-\u9fff-]/g, '')
      .replace(/-+/g, '-') || 'section'
  )
}

function parseHlLines(info: string): Set<number> {
  const set = new Set<number>()
  const m = info.match(/hl_lines=\s*\[([^\]]*)\]/)
  if (!m) return set
  for (const part of m[1].split(',')) {
    const raw = part.trim().replace(/^["']|["']$/g, '')
    if (!raw) continue
    const range = raw.match(/^(\d+)-(\d+)$/)
    if (range) {
      const a = Number(range[1])
      const b = Number(range[2])
      for (let n = Math.min(a, b); n <= Math.max(a, b); n++) set.add(n)
    } else {
      const n = Number(raw)
      if (!Number.isNaN(n)) set.add(n)
    }
  }
  return set
}

function renderCodeBlock(info: string, content: string, sourceLine?: number): string {
  const lang = (info.split(/\s+/)[0] || '').replace(/[{[].*$/, '') || 'text'
  const lineAttr = sourceLine ? ` data-source-line="${sourceLine}"` : ''

  if (lang === 'mermaid') {
    return `<div class="md-tag tag-mermaid mermaid"${lineAttr}>${escapeHtml(content.trim())}</div>\n`
  }

  const hlLines = parseHlLines(info)
  let highlighted = ''
  try {
    if (lang && hljs.getLanguage(lang)) {
      highlighted = hljs.highlight(content.replace(/\n$/, ''), { language: lang }).value
    } else {
      highlighted = escapeHtml(content.replace(/\n$/, ''))
    }
  } catch {
    highlighted = escapeHtml(content.replace(/\n$/, ''))
  }

  const lines = highlighted.split('\n')
  const body = lines
    .map((line, i) => {
      const n = i + 1
      const cls = hlLines.has(n) ? ' class="hl-line"' : ''
      return `<span${cls}><span class="line-no">${n}</span>${line || ' '}</span>`
    })
    .join('\n')

  return `<figure class="code-block" data-code-block data-language="${escapeHtml(lang)}"${lineAttr}>
  <figcaption class="code-block__toolbar">
    <span class="code-block__language">${escapeHtml(lang)}</span>
    <button class="code-block__copy" type="button" data-code-copy>复制</button>
  </figcaption>
  <div class="code-block__viewport">
    <pre class="hljs" tabindex="0"><code class="language-${escapeHtml(lang)}">${body}</code></pre>
  </div>
</figure>\n`
}

/** 把 token.map 起始行写成 data-source-line（1-based，对齐编辑器行号）。 */
function attachSourceLineAttrs(instance: MarkdownIt) {
  instance.core.ruler.push('attach_source_line', (state) => {
    for (const token of state.tokens) {
      if (!token.map) continue
      if (
        token.nesting === 1 ||
        token.type === 'fence' ||
        token.type === 'code_block' ||
        token.type === 'hr' ||
        token.type === 'html_block'
      ) {
        token.attrSet('data-source-line', String(token.map[0] + 1))
      }
    }
  })
}

/** 占位替换时补足换行，避免后续 token.map 相对原文行号漂移。 */
function padLineCount(replacement: string, original: string): string {
  const need = (original.match(/\n/g) || []).length
  const have = (replacement.match(/\n/g) || []).length
  return have >= need ? replacement : replacement + '\n'.repeat(need - have)
}

const md = new MarkdownIt({
  html: true,
  linkify: true,
  breaks: false,
})

md.use(markdownItFootnote)
md.use(markdownItAnchor, {
  slugify: slugifyHeading,
  permalink: markdownItAnchor.permalink.ariaHidden({
    symbol: '',
    class: 'headerlink',
    placement: 'after',
  }),
})
attachSourceLineAttrs(md)

/** 脚注引用始终显示 [n]，重复引用不显示 [n:m]。 */
md.renderer.rules.footnote_ref = (tokens, idx, options, env, self) => {
  const nameRule = md.renderer.rules.footnote_anchor_name
  const captionRule = md.renderer.rules.footnote_caption
  const id = nameRule ? nameRule(tokens, idx, options, env, self) : String(idx + 1)
  const caption = captionRule ? captionRule(tokens, idx, options, env, self) : String(idx + 1)
  // caption is like "[1]" or "[1:1]" — keep only the footnote number
  const num = String(caption).replace(/[\[\]]/g, '').split(':')[0] || '1'
  const meta = tokens[idx].meta as { subId?: number } | undefined
  let refid = id
  if (meta && meta.subId && meta.subId > 0) refid += `:${meta.subId}`
  return `<sup class="footnote-ref"><a href="#fn${id}" id="fnref${refid}">[${num}]</a></sup>`
}

md.renderer.rules.fence = (tokens, idx) => {
  const token = tokens[idx]
  const sourceLine = token.map ? token.map[0] + 1 : undefined
  return renderCodeBlock(token.info.trim(), token.content, sourceLine)
}

function applyMath(src: string): { text: string; slots: string[] } {
  const slots: string[] = []
  const put = (html: string, original: string) => {
    slots.push(html)
    return padLineCount(`@@MATH${slots.length - 1}@@`, original)
  }
  let s = src
  s = s.replace(/\$\$([\s\S]+?)\$\$/g, (full, tex) =>
    put(
      `<div class="katex-block"><div class="katex-display">${katex.renderToString(tex.trim(), { displayMode: true, throwOnError: false })}</div></div>`,
      full,
    ),
  )
  s = s.replace(/\\\[([\s\S]+?)\\\]/g, (full, tex) =>
    put(
      `<div class="katex-block"><div class="katex-display">${katex.renderToString(tex.trim(), { displayMode: true, throwOnError: false })}</div></div>`,
      full,
    ),
  )
  s = s.replace(/\\\((.+?)\\\)/g, (full, tex) =>
    put(katex.renderToString(tex.trim(), { displayMode: false, throwOnError: false }), full),
  )
  return { text: s, slots }
}

function restoreMath(html: string, slots: string[]): string {
  return html.replace(/@@MATH(\d+)@@/g, (_, i) => slots[Number(i)] ?? '')
}

function convertTables(src: string): string {
  const lines = src.split('\n')
  const out: string[] = []
  let i = 0
  while (i < lines.length) {
    const line = lines[i]
    const next = lines[i + 1]
    if (line.includes('|') && next && /^\s*\|?[\s:|-]+\|[\s:|-]+/.test(next)) {
      const rows: string[] = []
      while (i < lines.length && lines[i].includes('|') && lines[i].trim() !== '') {
        rows.push(lines[i])
        i++
      }
      if (rows.length >= 2) {
        // 保留行数，避免后续 data-source-line 相对原文偏移
        out.push(padLineCount(renderTable(rows).replace(/\n$/, ''), rows.join('\n')))
        continue
      }
      out.push(...rows)
      continue
    }
    out.push(line)
    i++
  }
  return out.join('\n')
}

function splitRow(row: string): string[] {
  let s = row.trim()
  if (s.startsWith('|')) s = s.slice(1)
  if (s.endsWith('|')) s = s.slice(0, -1)
  return s.split('|').map((c) => c.trim())
}

function renderTable(rows: string[]): string {
  const aligns = splitRow(rows[1]).map((cell) => {
    const left = cell.startsWith(':')
    const right = cell.endsWith(':')
    if (left && right) return 'center'
    if (right) return 'right'
    return 'left'
  })
  const head = splitRow(rows[0])
  const body = rows.slice(2).map(splitRow)
  // Keep cell markdown for later md.render — inject as plain; use inline renderer carefully
  const th = head
    .map((c, i) => `<th style="text-align:${aligns[i] || 'left'}">${md.renderInline(c)}</th>`)
    .join('')
  const trs = body
    .map(
      (cells) =>
        `<tr>${cells
          .map((c, i) => `<td style="text-align:${aligns[i] || 'left'}">${md.renderInline(c)}</td>`)
          .join('')}</tr>`,
    )
    .join('')
  return `<div class="table-wrap"><table><thead><tr>${th}</tr></thead><tbody>${trs}</tbody></table></div>\n`
}

function convertTaskLists(src: string): string {
  return src
    .replace(/^(\s*)- \[x\]\s+/gim, '$1- <input type="checkbox" checked disabled> ')
    .replace(/^(\s*)- \[ \]\s+/gim, '$1- <input type="checkbox" disabled> ')
    .replace(/^(\s*)- \(x\)\s+/gim, '$1- <input type="radio" checked disabled> ')
    .replace(/^(\s*)- \( \)\s+/gim, '$1- <input type="radio" disabled> ')
}

function convertStrike(src: string): string {
  return src.replace(/~~([^~\n]+?)~~/g, '<del>$1</del>')
}

export function renderMarkdown(source: string): string {
  const { text, slots } = applyMath(source)
  let s = text
  s = convertStrike(s)
  s = convertTaskLists(s)
  s = convertTables(s)
  let html = md.render(s)
  html = restoreMath(html, slots)
  return html
}
