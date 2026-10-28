<!-- Author: go-blog — 博客站点配置（作者信息走个人中心） -->

<script setup lang="ts">
import DefaultCoversUpload from '@/views/blog/components/DefaultCoversUpload.vue'
import NavEditor from '@/views/blog/site/components/NavEditor.vue'
import NavMenuPreview from '@/views/blog/site/components/NavMenuPreview.vue'
import type { NavGroupItem, NavItem } from '@/views/blog/site/components/NavEditor.vue'
import ImageUpload from '@/components/upload/ImageUpload.vue'
import { blogSiteApi } from '@/api'
import { hasPermission } from '@/utils'
import { ProCard } from 'pro-naive-ui'
import { onMounted, reactive } from 'vue'

type BeianItem = { name: string; url: string }

function asBool(v: unknown) {
  return v === true || v === 'true' || v === 1 || v === '1'
}

function normalizeNav(raw: unknown): NavItem[] {
  if (!Array.isArray(raw)) return []
  return raw.map((item: any) => ({
    name: String(item?.name || '').trim(),
    href: String(item?.href || '').trim(),
    icon: String(item?.icon || '').trim(),
    external: asBool(item?.external),
    children: Array.isArray(item?.children)
      ? item.children.map((c: any) => ({
          name: String(c?.name || '').trim(),
          href: String(c?.href || '').trim(),
          icon: String(c?.icon || '').trim(),
          external: asBool(c?.external),
        }))
      : [],
  }))
}

function normalizeNavGroup(raw: unknown): NavGroupItem[] {
  if (!Array.isArray(raw)) return []
  return raw.map((item: any) => ({
    name: String(item?.name || '').trim(),
    url: String(item?.url || '').trim(),
    icon: String(item?.icon || '').trim(),
  }))
}

function serializeNav(list: NavItem[]) {
  return list
    .filter((item) => item.name)
    .map((item) => {
      const children = (item.children || [])
        .filter((c) => c.name)
        .map((c) => {
          const row: Record<string, any> = { name: c.name }
          if (c.href) row.href = c.href
          if (c.icon) row.icon = c.icon
          if (c.external) row.external = true
          return row
        })
      const row: Record<string, any> = { name: item.name }
      if (item.href) row.href = item.href
      if (item.icon) row.icon = item.icon
      if (item.external) row.external = true
      if (children.length) row.children = children
      return row
    })
}

function serializeNavGroup(list: NavGroupItem[]) {
  return list
    .filter((item) => item.name && item.url)
    .map((item) => {
      const row: Record<string, any> = { name: item.name, url: item.url }
      if (item.icon) row.icon = item.icon
      return row
    })
}

const state = reactive({
  loading: false,
  saving: false,
  activeTab: 'site',
  /** 完整站点 JSON，保存时与表单字段合并 */
  rawConfig: {} as Record<string, any>,
  site_name: '',
  title: '',
  description: '',
  icon: '',
  started_at_ts: null as number | null,
  default_covers: [] as string[],
  nav: [] as NavItem[],
  nav_group: [] as NavGroupItem[],
  footer_copyright: '',
  footer_beian_name: '',
  footer_beian_url: 'https://beian.miit.gov.cn/',
  /** 公安备案：粘贴官网生成的 HTML 片段 */
  footer_police_html: '',
  /** 联系方式 */
  footer_email: '',
  footer_wechat_name: '',
  footer_wechat_qr: '',
})

function pickString(obj: any, key: string, fallback = '') {
  const v = obj?.[key]
  return typeof v === 'string' ? v : fallback
}

function startedAtToTs(raw: string): number | null {
  const text = String(raw || '').trim()
  if (!text) return null
  const d = new Date(text)
  if (Number.isNaN(d.getTime())) {
    const day = new Date(`${text.slice(0, 10)}T00:00:00+08:00`)
    return Number.isNaN(day.getTime()) ? null : day.getTime()
  }
  return d.getTime()
}

function tsToStartedAt(value: number | null): string {
  if (value == null) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}T00:00:00+08:00`
}

function pickCoverList(raw: unknown): string[] {
  if (!Array.isArray(raw)) return []
  const out: string[] = []
  for (const item of raw) {
    const url = String(item || '').trim()
    if (url && !out.includes(url)) out.push(url)
  }
  return out
}

function fillFromConfig(config: Record<string, any>) {
  state.rawConfig = { ...(config || {}) }
  state.site_name = pickString(config, 'site_name')
  state.title = pickString(config, 'title')
  state.description = pickString(config, 'description')
  state.icon = pickString(config, 'icon')
  state.started_at_ts = startedAtToTs(pickString(config, 'started_at'))
  state.default_covers = pickCoverList(config.default_covers)
  state.nav = normalizeNav(config.nav)
  state.nav_group = normalizeNavGroup(config.nav_group)

  const footer = config.footer || {}
  state.footer_copyright = pickString(footer, 'copyright')
  state.footer_police_html = pickString(footer, 'police_html')
  state.footer_email = pickString(footer, 'email')
  state.footer_wechat_name = pickString(footer, 'wechat_name')
  state.footer_wechat_qr = pickString(footer, 'wechat_qr')
  const beian = Array.isArray(footer.beian) ? footer.beian : []
  const first = (beian[0] || {}) as BeianItem
  state.footer_beian_name = pickString(first, 'name')
  state.footer_beian_url = pickString(first, 'url', 'https://beian.miit.gov.cn/')
}

function buildPayload(): Record<string, any> {
  const beian: BeianItem[] = []
  const name = state.footer_beian_name.trim()
  if (name) {
    beian.push({
      name,
      url: state.footer_beian_url.trim() || 'https://beian.miit.gov.cn/',
    })
  }
  const prevFooter =
    state.rawConfig.footer && typeof state.rawConfig.footer === 'object'
      ? { ...state.rawConfig.footer }
      : {}
  const prevAuthor =
    state.rawConfig.author && typeof state.rawConfig.author === 'object'
      ? { ...state.rawConfig.author }
      : {}
  // 作者身份字段不落库；仅保留侧栏展示壳
  delete prevAuthor.name
  delete prevAuthor.avatar_url
  delete prevAuthor.description
  delete prevAuthor.content
  return {
    ...state.rawConfig,
    site_name: state.site_name.trim(),
    title: state.title.trim(),
    description: state.description.trim(),
    icon: state.icon.trim(),
    started_at: tsToStartedAt(state.started_at_ts),
    default_covers: pickCoverList(state.default_covers),
    nav: serializeNav(state.nav),
    nav_group: serializeNavGroup(state.nav_group),
    author: prevAuthor,
    footer: {
      ...prevFooter,
      copyright: state.footer_copyright.trim(),
      beian,
      police_html: state.footer_police_html.trim(),
      email: state.footer_email.trim(),
      wechat_name: state.footer_wechat_name.trim(),
      wechat_qr: state.footer_wechat_qr.trim(),
    },
  }
}

async function reload() {
  state.loading = true
  try {
    const res = await blogSiteApi.getSite()
    fillFromConfig(res.data || {})
  } finally {
    state.loading = false
  }
}

async function save() {
  state.saving = true
  try {
    const payload = buildPayload()
    const res = await blogSiteApi.updateSite(payload)
    fillFromConfig(res.data || payload)
    window.$message.success('已保存')
  } finally {
    state.saving = false
  }
}

onMounted(() => {
  void reload()
})
</script>

<template>
  <div class="min-h-full">
    <ProCard title="站点配置" :show-collapse="false" :trigger-areas="[]">
      <NSpin :show="state.loading">
        <NTabs v-model:value="state.activeTab" type="line" animated>
          <NTabPane name="site" tab="站点信息" display-directive="show">
            <NForm label-placement="top" :disabled="state.saving">
              <NGrid :cols="2" :x-gap="16">
                <NGi>
                  <NFormItem label="站点名称">
                    <NInput v-model:value="state.site_name" placeholder="Demo Blog" />
                  </NFormItem>
                </NGi>
                <NGi>
                  <NFormItem label="标题">
                    <NInput v-model:value="state.title" placeholder="页面标题" />
                  </NFormItem>
                </NGi>
                <NGi :span="2">
                  <NFormItem label="描述">
                    <NInput
                      v-model:value="state.description"
                      type="textarea"
                      :autosize="{ minRows: 2, maxRows: 4 }"
                    />
                  </NFormItem>
                </NGi>
                <NGi>
                  <NFormItem label="图标 URL">
                    <NInput v-model:value="state.icon" placeholder="https://..." />
                  </NFormItem>
                </NGi>
                <NGi>
                  <NFormItem label="上线时间">
                    <NDatePicker
                      v-model:value="state.started_at_ts"
                      class="w-full"
                      type="date"
                      clearable
                    />
                  </NFormItem>
                </NGi>
              </NGrid>
            </NForm>
          </NTabPane>

          <NTabPane name="nav" tab="导航菜单" display-directive="show">
            <div class="nav-tab-layout">
              <div class="nav-tab-layout__editor">
                <NavEditor
                  v-model:nav="state.nav"
                  v-model:nav-group="state.nav_group"
                  :disabled="state.saving"
                />
              </div>
              <aside class="nav-tab-layout__preview">
                <NavMenuPreview
                  :site-name="state.site_name"
                  :site-icon="state.icon"
                  :nav="state.nav"
                  :nav-group="state.nav_group"
                />
              </aside>
            </div>
          </NTabPane>

          <NTabPane name="covers" tab="默认封面" display-directive="show">
            <NForm label-placement="top" :disabled="state.saving">
              <NFormItem label="默认封面池">
                <DefaultCoversUpload
                  v-model:value="state.default_covers"
                  :max="16"
                  :disabled="state.saving"
                />
              </NFormItem>
            </NForm>
          </NTabPane>

          <NTabPane name="footer" tab="页脚" display-directive="show">
            <NForm label-placement="top" :disabled="state.saving">
              <NText depth="3" class="footer-section-title">联系方式</NText>
              <NGrid :cols="2" :x-gap="16">
                <NGi>
                  <NFormItem label="邮箱">
                    <NInput
                      v-model:value="state.footer_email"
                      placeholder="hello@example.com"
                    />
                  </NFormItem>
                </NGi>
                <NGi>
                  <NFormItem label="公众号名称">
                    <NInput
                      v-model:value="state.footer_wechat_name"
                      placeholder="不填则显示「公众号」"
                    />
                  </NFormItem>
                </NGi>
                <NGi :span="2">
                  <NFormItem label="公众号二维码">
                    <NFlex vertical :size="6">
                      <ImageUpload
                        v-model:value="state.footer_wechat_qr"
                        variant="card"
                        value-type="object_name"
                        button-text="上传二维码"
                      />
                      <NText depth="3" style="font-size: 12px">
                        前台页脚悬停「公众号」时展示；清空则不显示公众号入口
                      </NText>
                    </NFlex>
                  </NFormItem>
                </NGi>
              </NGrid>

              <NText depth="3" class="footer-section-title">备案与版权</NText>
              <NGrid :cols="2" :x-gap="16">
                <NGi>
                  <NFormItem label="ICP 备案号">
                    <NInput v-model:value="state.footer_beian_name" placeholder="粤ICP备xxxxxxxx号" />
                  </NFormItem>
                </NGi>
                <NGi>
                  <NFormItem label="ICP 备案链接">
                    <NInput v-model:value="state.footer_beian_url" />
                  </NFormItem>
                </NGi>
                <NGi :span="2">
                  <NFormItem label="公安备案">
                    <NInput
                      v-model:value="state.footer_police_html"
                      type="textarea"
                      :rows="4"
                      placeholder="粘贴公安备案官网生成的 HTML，例如：&#10;<a href=&quot;https://beian.mps.gov.cn/#/query/webSearch?code=...&quot; target=&quot;_blank&quot;>...</a>"
                    />
                  </NFormItem>
                </NGi>
                <NGi :span="2">
                  <NFormItem label="版权（显示在页脚最底部）">
                    <NInput v-model:value="state.footer_copyright" />
                  </NFormItem>
                </NGi>
              </NGrid>
            </NForm>
          </NTabPane>
        </NTabs>

        <NFlex class="mt-4" justify="end">
          <NButton :loading="state.loading" :disabled="state.saving" @click="reload">重新加载</NButton>
          <NButton
            v-if="hasPermission('admin:blog:site:update')"
            type="primary"
            :loading="state.saving"
            :disabled="state.loading"
            @click="save"
          >
            保存
          </NButton>
        </NFlex>
      </NSpin>
    </ProCard>
  </div>
</template>

<style scoped>
.footer-section-title {
  display: block;
  margin: 4px 0 12px;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.footer-section-title + :deep(.n-grid) {
  margin-bottom: 8px;
}

.nav-tab-layout {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.nav-tab-layout__editor {
  flex: 1;
  min-width: 0;
}

.nav-tab-layout__preview {
  position: sticky;
  top: 12px;
  flex: none;
  width: min(520px, 42vw);
  max-width: 100%;
  min-height: 300px;
  padding: 12px;
  overflow: visible;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: var(--card-color);
}

@media (max-width: 1280px) {
  .nav-tab-layout {
    flex-direction: column;
  }

  .nav-tab-layout__preview {
    position: static;
    width: 100%;
  }
}
</style>
