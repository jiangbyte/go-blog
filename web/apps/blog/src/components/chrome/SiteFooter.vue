<script setup lang="ts">
/**
 * 站点页脚：联系方式 → 统计 → 版权 / 备案。
 */
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { getSite, getSiteStats } from '@/api'

const site = getSite()
const stats = ref<any | null>(null)
const now = ref(Date.now())
let timer: number | undefined

function formatCount(n: number) {
  const v = Number(n) || 0
  if (v >= 10000) return `${(v / 10000).toFixed(1).replace(/\.0$/, '')}万`
  if (v >= 1000) return `${(v / 1000).toFixed(1).replace(/\.0$/, '')}k`
  return String(v)
}

const runtimeLabel = computed(() => {
  const start = Date.parse(stats.value?.started_at || site?.started_at || '')
  if (!Number.isFinite(start)) return '—'
  const ms = Math.max(0, now.value - start)
  const days = Math.floor(ms / 86400000)
  const hours = Math.floor((ms % 86400000) / 3600000)
  if (days >= 365) {
    const years = Math.floor(days / 365)
    const rest = days % 365
    return rest ? `${years} 年 ${rest} 天` : `${years} 年`
  }
  if (days > 0) return `${days} 天 ${hours} 时`
  return `${hours} 时`
})

const items = computed(() => {
  const s = stats.value
  if (!s) {
    return [
      { label: '访问', value: '—' },
      { label: '阅读', value: '—' },
      { label: '文章', value: '—' },
      { label: '运行', value: '—' },
    ]
  }
  return [
    { label: '访问', value: formatCount(s.visit_count) },
    { label: '阅读', value: formatCount(s.read_count) },
    { label: '文章', value: formatCount(s.post_count) },
    { label: '运行', value: runtimeLabel.value },
  ]
})

/** ICP 备案项（不含版权） */
const beianParts = computed(() => {
  const parts: Array<{ text: string; href?: string }> = []
  for (const b of site?.footer?.beian || []) {
    const name = String(b?.name || '').trim()
    if (!name) continue
    parts.push({ text: name, href: b.url || undefined })
  }
  return parts
})

/** 公安备案官方 HTML；去 script，并在链接内文案前插入备案图标（若尚无 img） */
const policeHtml = computed(() => {
  const raw = String(site?.footer?.police_html || '').trim()
  if (!raw) return ''
  let html = raw.replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, '')
  if (!/<img[\s>]/i.test(html)) {
    const icon =
      '<img src="/beian-police.png" alt="" width="14" height="14" class="footer-beian__police-icon" />'
    if (/^<a\b/i.test(html)) {
      html = html.replace(/^(<a\b[^>]*>)/i, `$1${icon}`)
    } else {
      html = icon + html
    }
  }
  return html
})

const copyright = computed(() => String(site?.footer?.copyright || '').trim())
const email = computed(() => String(site?.footer?.email || '').trim())
const wechatName = computed(() => {
  const name = String(site?.footer?.wechat_name || '').trim()
  return name || '公众号'
})
const wechatQr = computed(() => String(site?.footer?.wechat_qr || '').trim())
const showContact = computed(() => !!email.value || !!wechatQr.value)
const showBeianRow = computed(() => beianParts.value.length > 0 || !!policeHtml.value)
const mailtoHref = computed(() => (email.value ? `mailto:${email.value}` : ''))

onMounted(async () => {
  stats.value = await getSiteStats({})
  timer = window.setInterval(() => {
    now.value = Date.now()
  }, 60_000)
})

onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <footer class="site-footer page-gutter">
    <div class="page-shell site-footer__inner">
      <!-- 1. PC 左右：联系方式 | 统计；移动端上下居中 -->
      <div class="footer-main" :class="{ 'footer-main--solo': !showContact }">
        <div v-if="showContact" class="footer-contact" aria-label="联系方式">
          <a
            v-if="email"
            class="footer-contact__item"
            :href="mailtoHref"
          >
            <span class="footer-contact__label">邮箱</span>
            <span class="footer-contact__value">{{ email }}</span>
          </a>
          <div
            v-if="wechatQr"
            class="footer-contact__item footer-contact__wechat"
          >
            <span class="footer-contact__label">{{ wechatName }}</span>
            <span class="footer-contact__value">扫码关注</span>
            <div class="footer-contact__qr" role="img" :aria-label="`${wechatName}二维码`">
              <img :src="wechatQr" :alt="`${wechatName}二维码`" />
            </div>
          </div>
        </div>

        <ul class="footer-stats" aria-label="站点统计">
          <li v-for="item in items" :key="item.label" class="footer-stats__item">
            <span class="footer-stats__value">{{ item.value }}</span>
            <span class="footer-stats__label">{{ item.label }}</span>
          </li>
        </ul>
      </div>

      <!-- 2. 左版权、右备案 -->
      <div v-if="copyright || showBeianRow" class="footer-meta">
        <p v-if="copyright" class="footer-copyright">{{ copyright }}</p>
        <div v-if="showBeianRow" class="footer-beian">
          <template v-for="(part, i) in beianParts" :key="`${part.text}-${i}`">
            <a
              v-if="part.href"
              :href="part.href"
              class="footer-beian__link"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ part.text }}
            </a>
            <span v-else>{{ part.text }}</span>
          </template>
          <span v-if="policeHtml" class="footer-beian__police" v-html="policeHtml" />
        </div>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.site-footer {
  margin-top: auto;
  padding-top: 1.25rem;
  padding-bottom: calc(1.25rem + env(safe-area-inset-bottom, 0px));
  border-top: 1px solid var(--efu-card-border);
  background: color-mix(in srgb, var(--efu-secondbg) 55%, transparent);
}

.site-footer__inner {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 0.85rem;
}

.footer-main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.85rem 1.5rem;
  width: 100%;
}

.footer-main--solo {
  justify-content: center;
}

.footer-contact {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-start;
  align-items: stretch;
  gap: 0.75rem 1.5rem;
}

.footer-contact__item {
  position: relative;
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.12rem;
  padding: 0.15rem 0;
  color: inherit;
  text-decoration: none;
  min-width: 8rem;
}

.footer-contact__label {
  color: var(--efu-secondtext);
  font-size: 0.72rem;
  line-height: 1.2;
}

.footer-contact__value {
  color: var(--milk-ink);
  font-size: 0.84rem;
  font-weight: 500;
  line-height: 1.35;
  word-break: break-all;
}

a.footer-contact__item:hover .footer-contact__value {
  text-decoration: underline;
  text-underline-offset: 0.18em;
}

.footer-contact__wechat {
  cursor: default;
}

.footer-contact__qr {
  position: absolute;
  left: 50%;
  bottom: calc(100% + 10px);
  z-index: 20;
  display: none;
  width: 132px;
  padding: 8px;
  border: 1px solid var(--efu-card-border);
  background: var(--efu-card-bg, #fff);
  box-shadow: 0 8px 24px rgba(26, 24, 22, 0.1);
  transform: translateX(-50%);
}

.footer-contact__qr img {
  display: block;
  width: 100%;
  height: auto;
  aspect-ratio: 1;
  object-fit: contain;
}

.footer-contact__wechat:hover .footer-contact__qr,
.footer-contact__wechat:focus-within .footer-contact__qr {
  display: block;
}

.footer-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem 1.25rem;
  width: 100%;
}

.footer-stats {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: stretch;
  gap: 0;
  margin: 0;
  margin-left: auto;
  padding: 0;
  list-style: none;
  width: auto;
  max-width: 100%;
}

.footer-main--solo .footer-stats {
  margin-left: 0;
  justify-content: center;
}

.footer-stats__item {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.15rem;
  min-width: 4.5rem;
  padding: 0.15rem 1.1rem;
}

.footer-stats__item + .footer-stats__item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 18%;
  bottom: 18%;
  width: 1px;
  background: color-mix(in srgb, var(--efu-card-border) 90%, transparent);
}

.footer-stats__value {
  color: var(--milk-ink);
  font-size: 0.92rem;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
  line-height: 1.15;
  letter-spacing: 0.01em;
}

.footer-stats__label {
  color: var(--efu-secondtext);
  font-size: 0.72rem;
  line-height: 1.2;
}

.footer-beian {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: center;
  gap: 0.35rem 0.75rem;
  margin-left: auto;
  color: var(--efu-secondtext);
  font-size: 0.75rem;
  line-height: 1.5;
  text-align: right;
}

.footer-beian__link {
  color: inherit;
  text-decoration: none;
}

.footer-beian__link:hover {
  color: var(--milk-ink);
}

.footer-beian__police {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  color: inherit;
}

.footer-beian__police :deep(a) {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  color: inherit;
  text-decoration: none;
}

.footer-beian__police :deep(a:hover) {
  color: var(--milk-ink);
}

.footer-beian__police :deep(img),
.footer-beian__police :deep(.footer-beian__police-icon) {
  width: 14px;
  height: 14px;
  object-fit: contain;
  vertical-align: middle;
  flex-shrink: 0;
}

.footer-copyright {
  margin: 0;
  color: var(--efu-secondtext);
  font-size: 0.75rem;
  line-height: 1.5;
  text-align: left;
}

@media (max-width: 639px) {
  .footer-main {
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .footer-contact {
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.65rem;
  }

  .footer-contact__item {
    align-items: center;
    text-align: center;
  }

  .footer-stats {
    margin-left: 0;
    justify-content: center;
  }

  .footer-meta {
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .footer-copyright {
    text-align: center;
  }

  .footer-beian {
    justify-content: center;
    margin-left: 0;
    text-align: center;
  }
}

@media (max-width: 479px) {
  .footer-stats__item {
    min-width: 3.5rem;
    padding: 0.15rem 0.7rem;
  }

  .footer-stats__value {
    font-size: 0.86rem;
  }
}
</style>
