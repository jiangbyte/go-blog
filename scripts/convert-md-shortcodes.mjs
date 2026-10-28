#!/usr/bin/env node
/**
 * Author: go-blog
 * 将历史短代码转为普通 Markdown / HTML（不再依赖 shortcode 引擎）。
 *
 * 用法：
 *   node scripts/convert-md-shortcodes.mjs --file path.md
 *   node scripts/convert-md-shortcodes.mjs --stdin
 *   node scripts/convert-md-shortcodes.mjs --db   # 更新 go_blog 中含 {{< 的文章
 */
import { readFileSync, writeFileSync } from 'node:fs'
import { spawnSync } from 'node:child_process'

function attr(attrs, key) {
  return attrs.match(new RegExp(`${key}="([^"]*)"`))?.[1]
}

/** 短代码 → MD/HTML */
export function convertMdShortcodes(body) {
  let s = String(body || '')

  s = s.replace(/\{\{<\s*note([^>]*)>\}\}([\s\S]*?)\{\{<\s*\/note\s*>\}\}/gi, (_, a, inner) => {
    const type = attr(a, 'type') || 'info'
    const style = attr(a, 'style') || 'flat'
    return `\n<aside class="md-tag note ${type} ${style} has-icon"><div class="note-content">\n\n${inner.trim()}\n\n</div></aside>\n`
  })

  s = s.replace(/\{\{<\s*subnote([^>]*)>\}\}([\s\S]*?)\{\{<\s*\/subnote\s*>\}\}/gi, (_, a, inner) => {
    const type = attr(a, 'type') || 'warning'
    return `\n<aside class="md-tag note ${type} simple has-icon"><div class="note-content">\n\n${inner.trim()}\n\n</div></aside>\n`
  })

  s = s.replace(/\{\{<\s*img\s+([^>]+?)\s*>\}\}/gi, (_, a) => {
    const src = attr(a, 'src') || ''
    const alt = attr(a, 'alt') || ''
    const caption = attr(a, 'caption')
    return caption ? `![${alt}](${src})\n*${caption}*\n` : `![${alt}](${src})\n`
  })
  s = s.replace(/\{\{<\s*inlineImg\s+([^>]+?)\s*>\}\}/gi, (_, a) => {
    const src = attr(a, 'src') || ''
    const alt = attr(a, 'alt') || ''
    const h = attr(a, 'height') || '22px'
    return `<img class="inline-img" src="${src}" alt="${alt}" style="height:${h}" />`
  })

  s = s.replace(/\{\{<\s*mermaid\s*>\}\}([\s\S]*?)\{\{<\s*\/mermaid\s*>\}\}/gi, (_, inner) => {
    return `\n\`\`\`mermaid\n${inner.trim()}\n\`\`\`\n`
  })

  s = s.replace(/\{\{<\s*p\s+color="([^"]*)"\s*>\}\}([\s\S]*?)\{\{<\s*\/p\s*>\}\}/gi, (_, c, inner) => {
    return `\n<p class="md-tag tag-p ${c}">${inner.trim()}</p>\n`
  })
  s = s.replace(/\{\{<\s*span\s+color="([^"]*)"\s*>\}\}([\s\S]*?)\{\{<\s*\/span\s*>\}\}/gi, (_, c, inner) => {
    return `<span class="md-tag tag-span ${c}">${inner.trim()}</span>`
  })
  s = s.replace(/\{\{<\s*label\s+color="([^"]*)"\s*>\}\}([\s\S]*?)\{\{<\s*\/label\s*>\}\}/gi, (_, c, inner) => {
    return `<span class="md-tag tag-label ${c}">${inner.trim()}</span>`
  })

  s = s.replace(
    /\{\{<\s*checkbox\s+checked=(true|false)[^>]*>\}\}([\s\S]*?)\{\{<\s*\/checkbox\s*>\}\}/gi,
    (_, checked, inner) => `\n- [${checked === 'true' ? 'x' : ' '}] ${inner.trim()}\n`,
  )
  s = s.replace(
    /\{\{<\s*radio\s+checked=(true|false)[^>]*>\}\}([\s\S]*?)\{\{<\s*\/radio\s*>\}\}/gi,
    (_, checked, inner) => `\n- (${checked === 'true' ? 'x' : ' '}) ${inner.trim()}\n`,
  )

  s = s.replace(/\{\{<\s*keyboard\s*>\}\}([\s\S]*?)\{\{<\s*\/keyboard\s*>\}\}/gi, (_, inner) => {
    return `<kbd>${inner.trim()}</kbd>`
  })
  s = s.replace(/\{\{<\s*spoiler\s*>\}\}([\s\S]*?)\{\{<\s*\/spoiler\s*>\}\}/gi, (_, inner) => {
    return `<span class="md-tag tag-spoiler" title="spoiler">${inner.trim()}</span>`
  })
  s = s.replace(/\{\{<\s*bubble[^>]*>\}\}([\s\S]*?)\{\{<\s*\/bubble\s*>\}\}/gi, (_, inner) => {
    return `<span class="md-tag tag-bubble">${inner.trim()}</span>`
  })

  s = s.replace(/\{\{<\s*fold\s+([^>]*)>\}\}([\s\S]*?)\{\{<\s*\/fold\s*>\}\}/gi, (_, a, inner) => {
    const title = attr(a, 'title') || '折叠'
    const open = /\bopen\b/.test(a) ? ' open' : ''
    return `\n<details class="md-tag tag-fold"${open}><summary class="tag-fold-summary"><span class="tag-fold-title">${title}</span></summary><div class="tag-fold-content">\n\n${inner.trim()}\n\n</div></details>\n`
  })

  s = s.replace(
    /\{\{<\s*hideInline\s+text="([^"]*)"\s*>\}\}([\s\S]*?)\{\{<\s*\/hideInline\s*>\}\}/gi,
    (_, text, inner) =>
      `<details class="md-tag tag-hide-inline"><summary>${text}</summary>${inner.trim()}</details>`,
  )
  s = s.replace(
    /\{\{<\s*hideBlock\s+text="([^"]*)"\s*>\}\}([\s\S]*?)\{\{<\s*\/hideBlock\s*>\}\}/gi,
    (_, text, inner) =>
      `\n<details class="md-tag tag-hide-block"><summary>${text}</summary><div>\n\n${inner.trim()}\n\n</div></details>\n`,
  )
  s = s.replace(
    /\{\{<\s*hideToggle\s+title="([^"]*)"\s*>\}\}([\s\S]*?)\{\{<\s*\/hideToggle\s*>\}\}/gi,
    (_, title, inner) =>
      `\n<details class="md-tag tag-hide-toggle"><summary>${title}</summary><div>\n\n${inner.trim()}\n\n</div></details>\n`,
  )

  s = s.replace(/\{\{<\s*tabs\s*>\}\}([\s\S]*?)\{\{<\s*\/tabs\s*>\}\}/gi, (_, inner) => {
    let i = 0
    const panels = String(inner).replace(
      /\{\{%?\s*tab\s+name="([^"]*)"\s*%?\}\}([\s\S]*?)\{\{%?\s*\/tab\s*%?\}\}/gi,
      (___, name, tabInner) => {
        const active = i === 0 ? ' active' : ''
        const id = `tab-${i++}`
        return `<section class="tab-item-content${active}" id="${id}" data-tab-title="${name}">\n\n${tabInner.trim()}\n\n</section>`
      },
    )
    return `\n<div class="md-tag tag-tabs tabs" data-md-tabs><div class="tab-contents">${panels}</div></div>\n`
  })

  s = s.replace(/\{\{<\s*button\s+([^>]+?)\s*>\}\}/gi, (_, a) => {
    const url = attr(a, 'url') || '#'
    const text = attr(a, 'text') || 'Button'
    const option = attr(a, 'option') || ''
    return `<a class="md-tag st-btn tag-button ${option}" href="${url}"><span>${text}</span></a>`
  })

  s = s.replace(/\{\{<\s*link\s+([^>]+?)\s*>\}\}/gi, (_, a) => {
    const url = attr(a, 'url') || '#'
    const title = attr(a, 'title') || url
    const desc = attr(a, 'desc') || ''
    return `\n<div class="md-tag tag-link-card"><a class="tag-link" href="${url}"><div class="tag-link-tips">链接</div><div class="tag-link-bottom"><div class="tag-link-right"><div class="tag-link-title">${title}</div><div class="tag-link-sitename">${desc}</div></div></div></a></div>\n`
  })

  s = s.replace(/\{\{<\s*card\s+([^>]+?)\s*>\}\}([\s\S]*?)\{\{<\s*\/card\s*>\}\}/gi, (_, a, inner) => {
    const title = attr(a, 'title') || 'Card'
    const url = attr(a, 'url') || '#'
    const cover = attr(a, 'cover')
    const tag = attr(a, 'tag') || ''
    const coverHtml = cover
      ? `<div class="card_cover" style="background-image:url('${cover}')"></div>`
      : ''
    return `\n<article class="md-tag card_box tag-card ${cover ? 'has-cover' : 'no-cover'}"><a href="${url}">${coverHtml}<div class="card_body"><span class="card_tag">${tag}</span><span class="card_title">${title}</span><span class="card_description">${inner.trim()}</span></div></a></article>\n`
  })

  s = s.replace(/\{\{<\s*timeline\s*>\}\}([\s\S]*?)\{\{<\s*\/timeline\s*>\}\}/gi, (_, inner) => {
    const nodes = String(inner).replace(
      /\{\{<\s*timenode\s+title="([^"]*)"\s+date="([^"]*)"\s*>\}\}([\s\S]*?)\{\{<\s*\/timenode\s*>\}\}/gi,
      (___, title, date, nodeInner) =>
        `<div class="timeline-item md-tag timenode tag-timenode"><div class="timeline-item-title meta"><div class="item-circle"><p>${date}</p></div></div><div class="timeline-item-content body"><strong>${title}</strong><p>${nodeInner.trim()}</p></div></div>`,
    )
    return `\n<div class="md-tag timeline tag-timeline">${nodes}</div>\n`
  })

  s = s.replace(/\{\{<\s*chartjs([^>]*)>\}\}([\s\S]*?)\{\{<\s*\/chartjs\s*>\}\}/gi, (_, a, inner) => {
    const desc = attr(a, 'description') || ''
    return `\n<div class="md-tag chartjs-container tag-chart" data-chartjs-id="chart-ks"><pre class="chartjs-src" hidden>${inner.trim()}</pre>${desc ? `<div class="chartjs-desc">${desc}</div>` : ''}</div>\n`
  })

  s = s.replace(/\{\{<\s*typeit([^>]*)>\}\}([\s\S]*?)\{\{<\s*\/typeit\s*>\}\}/gi, (_, a, inner) => {
    const speed = attr(a, 'speed') || '55'
    return `\n<p class="md-tag tag-typeit" data-typeit data-speed="${speed}">${inner.trim()}</p>\n`
  })

  s = s.replace(/\{\{<\s*github\s+repo="([^"]+)"\s*>\}\}/gi, (_, repo) => {
    const name = repo.split('/').pop()
    return `\n<article class="md-tag repo-card tag-repo" data-repo-provider="github" data-repo-endpoint="https://api.github.com/repos/${repo}"><div class="repo-card-header"><span class="repo-provider">GitHub</span></div><a class="repo-title" href="https://github.com/${repo}" target="_blank" rel="noopener noreferrer">${name}</a><p class="repo-desc">github.com/${repo}</p><div class="repo-card-footer"><span class="repo-status">仓库</span></div></article>\n`
  })

  s = s.replace(/\{\{<\s*score\s*>\}\}([\s\S]*?)\{\{<\s*\/score\s*>\}\}/gi, (_, inner) => {
    return `\n<pre class="abc-music-sheet md-tag tag-score">${inner.trim()}</pre>\n`
  })

  // 残留短代码直接剔除，避免前台再露出 {{< … >}}
  s = s.replace(/\{\{[<%]\s*\/?[A-Za-z0-9_-]+[\s\S]*?[%>]\}\}/g, '')

  return s.trim() + '\n'
}

function stripHugoFrontMatter(raw) {
  if (!raw.startsWith('---')) return raw
  const end = raw.indexOf('\n---', 3)
  if (end < 0) return raw
  return raw.slice(end + 4).replace(/^\r?\n/, '')
}

function mysqlEscape(s) {
  return s.replace(/\\/g, '\\\\').replace(/'/g, "\\'")
}

function updateDb() {
  const dump = spawnSync(
    'docker',
    [
      'exec',
      'mysql',
      'mysql',
      '-uroot',
      "-pinfra123!",
      '--default-character-set=utf8mb4',
      '-N',
      'go_blog',
      '-e',
      "SET NAMES utf8mb4; SELECT uid, body FROM blog_post WHERE body LIKE '%{{<%' OR body LIKE '%{{%';",
    ],
    { encoding: 'utf8', maxBuffer: 50 * 1024 * 1024 },
  )
  if (dump.status !== 0) {
    console.error(dump.stderr || dump.stdout)
    process.exit(1)
  }
  // mysql -N with TAB: uid\tbody — body may contain newlines, so use a safer approach
  const list = spawnSync(
    'docker',
    [
      'exec',
      'mysql',
      'mysql',
      '-uroot',
      "-pinfra123!",
      '--default-character-set=utf8mb4',
      '-N',
      'go_blog',
      '-e',
      "SET NAMES utf8mb4; SELECT uid FROM blog_post WHERE body LIKE '%{{<%' OR body LIKE '%{{%';",
    ],
    { encoding: 'utf8' },
  )
  const uids = list.stdout
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l && !l.startsWith('mysql:'))
  if (!uids.length) {
    console.log('no posts with shortcodes')
    return
  }
  for (const uid of uids) {
    const bodyRes = spawnSync(
      'docker',
      [
        'exec',
        'mysql',
        'mysql',
        '-uroot',
        "-pinfra123!",
        '--default-character-set=utf8mb4',
        '-N',
        '-r',
        'go_blog',
        '-e',
        `SET NAMES utf8mb4; SELECT body FROM blog_post WHERE uid='${uid}'`,
      ],
      { encoding: 'utf8', maxBuffer: 50 * 1024 * 1024 },
    )
    let body = bodyRes.stdout
    // strip mysql password warning lines
    body = body.replace(/^mysql: \[Warning\].*\n/gm, '')
    if (body.endsWith('\n')) body = body.slice(0, -1)
    const converted = convertMdShortcodes(body)
    const left = (converted.match(/\{\{[<%]/g) || []).length
    const sql = `SET NAMES utf8mb4; UPDATE blog_post SET body='${mysqlEscape(converted)}' WHERE uid='${uid}';`
    const up = spawnSync(
      'docker',
      ['exec', '-i', 'mysql', 'mysql', '-uroot', "-pinfra123!", '--default-character-set=utf8mb4', 'go_blog'],
      { input: sql, encoding: 'utf8', maxBuffer: 50 * 1024 * 1024 },
    )
    if (up.status !== 0) {
      console.error(up.stderr || up.stdout)
      process.exit(1)
    }
    console.log(`updated ${uid}, residual_shortcodes=${left}, bytes=${converted.length}`)
  }
}

const args = process.argv.slice(2)
if (args.includes('--db')) {
  updateDb()
} else if (args.includes('--stdin')) {
  const raw = readFileSync(0, 'utf8')
  process.stdout.write(convertMdShortcodes(stripHugoFrontMatter(raw)))
} else {
  const fi = args.indexOf('--file')
  if (fi < 0 || !args[fi + 1]) {
    console.error('usage: --file path.md | --stdin | --db')
    process.exit(1)
  }
  const path = args[fi + 1]
  const raw = readFileSync(path, 'utf8')
  const hasFm = raw.startsWith('---')
  const fmEnd = hasFm ? raw.indexOf('\n---', 3) : -1
  const fm = hasFm && fmEnd >= 0 ? raw.slice(0, fmEnd + 4) + '\n\n' : ''
  const body = hasFm && fmEnd >= 0 ? raw.slice(fmEnd + 4).replace(/^\r?\n/, '') : raw
  const out = fm + convertMdShortcodes(body)
  const outPath = args.includes('--in-place') ? path : args[args.indexOf('--out') + 1] || path
  if (!args.includes('--in-place') && !args.includes('--out')) {
    process.stdout.write(out)
  } else {
    writeFileSync(outPath, out)
    console.log('wrote', outPath)
  }
}
