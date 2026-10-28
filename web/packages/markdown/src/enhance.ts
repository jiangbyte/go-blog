/** Author: go-blog — 正文 DOM 增强（tabs / 复制 / mermaid / chart / abc） */

export type EnhanceMarkdownOptions = {
  /** 点击正文图片时回调；不传则不绑定灯箱 */
  onImageClick?: (src: string, alt: string) => void
}

/**
 * 对 renderMarkdown 产出的 DOM 做运行时增强。
 * 幂等：已处理节点带 data-* 标记，重复调用可跳过。
 */
export async function enhanceMarkdownRoot(
  el: HTMLElement,
  opts: EnhanceMarkdownOptions = {},
): Promise<void> {
  // 1. 图片点击预览
  if (opts.onImageClick) {
    el.querySelectorAll<HTMLImageElement>('img').forEach((img) => {
      if (img.dataset.previewBound) return
      img.dataset.previewBound = '1'
      img.addEventListener('click', (e) => {
        e.preventDefault()
        e.stopPropagation()
        const src = img.currentSrc || img.src
        if (src) opts.onImageClick?.(src, img.alt || '')
      })
    })
  }

  // 2. 代码块复制
  el.querySelectorAll<HTMLButtonElement>('[data-code-copy]').forEach((btn) => {
    if (btn.dataset.bound) return
    btn.dataset.bound = '1'
    btn.addEventListener('click', async () => {
      const block = btn.closest('.code-block')
      const text = block?.querySelector('code')?.innerText ?? ''
      try {
        await navigator.clipboard.writeText(text)
        btn.textContent = '已复制'
        window.setTimeout(() => {
          btn.textContent = '复制'
        }, 1200)
      } catch {
        btn.textContent = '失败'
      }
    })
  })

  // 3. Tabs
  el.querySelectorAll<HTMLElement>('[data-md-tabs]:not([data-tabs-ready])').forEach(
    (container) => {
      const panels = [
        ...container.querySelectorAll<HTMLElement>(':scope > .tab-contents > .tab-item-content'),
      ]
      if (!panels.length) return
      const nav = document.createElement('ul')
      nav.className = 'nav-tabs'
      nav.setAttribute('role', 'tablist')
      const activate = (active: number) => {
        nav.querySelectorAll('.tab').forEach((t, i) => t.classList.toggle('active', i === active))
        panels.forEach((p, i) => {
          p.classList.toggle('active', i === active)
          p.hidden = i !== active
        })
      }
      panels.forEach((panel, index) => {
        const li = document.createElement('li')
        li.className = `tab${index === 0 ? ' active' : ''}`
        const button = document.createElement('button')
        button.type = 'button'
        button.textContent = panel.dataset.tabTitle || `Tab ${index + 1}`
        button.addEventListener('click', () => activate(index))
        li.appendChild(button)
        nav.appendChild(li)
      })
      container.prepend(nav)
      activate(0)
      container.dataset.tabsReady = 'true'
    },
  )

  // 4. TypeIt 简易打字
  el.querySelectorAll<HTMLElement>('[data-typeit]:not([data-typeit-ready])').forEach((node) => {
    const full = node.textContent || ''
    const speed = Number(node.dataset.speed) || 55
    node.textContent = ''
    node.dataset.typeitReady = 'true'
    let i = 0
    const tick = () => {
      i++
      node.textContent = full.slice(0, i)
      if (i < full.length) window.setTimeout(tick, speed)
    }
    tick()
  })

  // 5. Mermaid
  const diagrams = [...el.querySelectorAll<HTMLElement>('.mermaid:not([data-processed])')]
  if (diagrams.length) {
    const mermaid = (await import('mermaid')).default
    mermaid.initialize({
      startOnLoad: false,
      theme: 'default',
      securityLevel: 'loose',
      fontFamily: 'Noto Sans SC, sans-serif',
    })
    diagrams.forEach((d, idx) => {
      if (!d.dataset.mermaidSource) d.dataset.mermaidSource = d.textContent || ''
      d.id = d.id || `mermaid-${idx}-${Date.now()}`
    })
    try {
      await mermaid.run({ nodes: diagrams })
    } catch (e) {
      console.warn('[markdown] mermaid failed', e)
    }
  }

  // 6. Chart.js
  const charts = [...el.querySelectorAll<HTMLElement>('.chartjs-container:not([data-chart-ready])')]
  if (charts.length) {
    const { Chart } = await import('chart.js/auto')
    charts.forEach((container, index) => {
      try {
        const source = container.querySelector('.chartjs-src')
        if (!source) return
        container.querySelector('.chartjs-wrap')?.remove()
        const wrap = document.createElement('div')
        wrap.className = 'chartjs-wrap'
        const canvas = document.createElement('canvas')
        canvas.id = container.dataset.chartjsId || `chartjs-${index}`
        wrap.appendChild(canvas)
        source.insertAdjacentElement('afterend', wrap)
        const definition = JSON.parse(source.textContent || '{}')
        new Chart(canvas, definition)
        container.dataset.chartReady = 'true'
      } catch (e) {
        console.warn('[markdown] chart failed', e)
      }
    })
  }

  // 7. ABC 乐谱
  const sheets = [...el.querySelectorAll<HTMLElement>('.abc-music-sheet:not([data-rendered])')]
  if (sheets.length) {
    try {
      const ABCJS = (await import('abcjs')).default
      sheets.forEach((sheet) => {
        const src = sheet.textContent || ''
        sheet.textContent = ''
        ABCJS.renderAbc(sheet, src, { responsive: 'resize' })
        sheet.dataset.rendered = 'true'
      })
    } catch (e) {
      console.warn('[markdown] abcjs failed', e)
    }
  }
}
