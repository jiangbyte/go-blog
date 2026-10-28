<!-- Author: go-blog — 简历全屏可视化工作台（侧栏编辑 + 即时预览） -->

<script setup lang="ts">
import { resumeApi } from '@/api'
import { hasPermission } from '@/utils'
import { NCard, NEmpty, NScrollbar, NSpin, NText } from 'naive-ui'
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ResumeBackendPreview from '../preview/ResumeBackendPreview.vue'
import { downloadResumeDocument } from '../preview/resumePrint'
import { sanitizeFontScale, sanitizeLineHeight } from '@jiangbyte/shared'
import ModuleRail, { type RailKey } from './ModuleRail.vue'
import ProfilePanel from './panels/ProfilePanel.vue'
import SectionPanel from './panels/SectionPanel.vue'
import ThemePanel from './panels/ThemePanel.vue'
import WorkbenchHeader from './WorkbenchHeader.vue'
import './workbench.css'

const route = useRoute()
const router = useRouter()
const canUpdate = hasPermission('admin:resume:update')
const canSaveSection = hasPermission('admin:resume:savesection')
const canDeleteSection = hasPermission('admin:resume:deletesection')
const canCopySection = hasPermission('admin:resume:copysection')
const canReorderSection = hasPermission('admin:resume:reordersection')
const canPublish = hasPermission('admin:resume:publish')
const canUnpublish = hasPermission('admin:resume:unpublish')
const canExport = hasPermission('admin:resume:export')
/** 资料/主题可写 */
const canWriteProfile = canUpdate
/** 分节内容可写 */
const canWriteSection = canSaveSection

const resumeId = computed(() => String(route.query.id || ''))

const state = reactive({
  loading: false,
  saving: false,
  dirty: false,
  title: '',
  template_id: 'classic',
  status: 'unpublished',
  custom_css: '',
  theme: {
    themeColor: 'blue',
    layout: 'left',
    infoDisplay: 'text',
    language: 'zh',
    fontFamily: 'default',
    fontSizeScale: 1,
    spacing: { moduleGap: 12, lineHeight: 1.55, padX: 48, padTop: 48, padBottom: 60 },
  },
  profile: {
    name: '',
    tagline: '',
    avatar_url: '',
    github_url: '',
    fields: [] as { key: string; label: string; value: string; icon?: string }[],
  },
  sections: [] as any[],
  templates: [] as {
    id: string
    name: string
    description?: string
    defaultColor?: string
    structure?: string
  }[],
})

const activeKey = ref<RailKey>('profile')
let saveTimer: ReturnType<typeof setTimeout> | null = null
let skipDirtyWatch = false
let hydrated = false
let lastSavedFp = ''

const activeSection = computed(() => {
  if (!String(activeKey.value).startsWith('section:')) return null
  const id = String(activeKey.value).slice('section:'.length)
  return state.sections.find((s) => s.id === id) || null
})

const activeSectionIndex = computed(() => {
  if (!String(activeKey.value).startsWith('section:')) return -1
  const id = String(activeKey.value).slice('section:'.length)
  return state.sections.findIndex((s) => s.id === id)
})

const sideTitle = computed(() => {
  if (activeKey.value === 'profile') return '基本信息'
  if (activeKey.value === 'theme') return '外观'
  return activeSection.value?.title || '分节'
})

const previewDoc = computed(() => ({
  id: resumeId.value,
  title: state.title,
  template_id: state.template_id,
  theme_json: state.theme,
  custom_css: state.custom_css,
  profile_json: state.profile,
  sections: state.sections,
}))

/** 内容指纹，用于跳过无实质变更的自动保存。 */
function contentFingerprint() {
  return JSON.stringify({
    title: state.title,
    template_id: state.template_id,
    custom_css: state.custom_css,
    theme: state.theme,
    profile: state.profile,
    sections: state.sections.map((s) => ({
      id: s.id,
      type: s.type,
      title: s.title,
      icon: s.icon,
      sort_code: s.sort_code,
      meta_json: s.meta_json,
      body_html: s.body_html,
      entries_json: s.entries_json,
    })),
  })
}

/** 跳转 Admin 内全屏预览页。 */
async function goAdminPreview() {
  if (state.dirty) {
    await saveAll({ silent: true })
  }
  void router.push({ path: '/resume/preview', query: { id: resumeId.value } })
}

/** 下载与当前预览一致的简历 PDF。 */
async function onDownload() {
  if (!canExport || !resumeId.value) return
  if (state.dirty) {
    await saveAll({ silent: true })
  }
  try {
    await downloadResumeDocument(resumeId.value, state.profile.name || state.title || '简历')
    window.$message.success('简历已下载')
  } catch (err: any) {
    window.$message.error(err?.message || '下载失败')
  }
}

/** 加载详情到本地 state。 */
async function reload() {
  if (!resumeId.value) {
    window.$message.error('缺少简历 ID')
    return
  }
  state.loading = true
  skipDirtyWatch = true
  try {
    // 1. 并行拉详情与模板列表
    const [detailRes, tplRes] = await Promise.all([
      resumeApi.detail(resumeId.value),
      resumeApi.templates(),
    ])
    const data = detailRes.data
    // 2. 写入本地编辑 state
    state.title = data.title || ''
    state.template_id = data.template_id || 'classic'
    state.status = data.status || 'unpublished'
    state.custom_css = data.custom_css || ''
    state.theme = {
      themeColor: data.theme_json?.themeColor || 'blue',
      layout: data.theme_json?.layout || 'left',
      infoDisplay: data.theme_json?.infoDisplay || 'text',
      language: data.theme_json?.language === 'en' ? 'en' : 'zh',
      fontFamily: data.theme_json?.fontFamily || 'default',
      fontSizeScale: sanitizeFontScale(data.theme_json?.fontSizeScale),
      spacing: {
        moduleGap: data.theme_json?.spacing?.moduleGap ?? 12,
        lineHeight: sanitizeLineHeight(data.theme_json?.spacing?.lineHeight),
        padX: data.theme_json?.spacing?.padX ?? 48,
        padTop: data.theme_json?.spacing?.padTop ?? 48,
        padBottom: data.theme_json?.spacing?.padBottom ?? 60,
      },
    }
    state.profile = {
      name: data.profile_json?.name || '',
      tagline: data.profile_json?.tagline || '',
      avatar_url: data.profile_json?.avatar_url || '',
      github_url: data.profile_json?.github_url || '',
      fields: Array.isArray(data.profile_json?.fields) ? [...data.profile_json.fields] : [],
    }
    state.sections = Array.isArray(data.sections)
      ? data.sections.map((s: any) => ({
          ...s,
          meta_json: s.meta_json || {},
          entries_json: Array.isArray(s.entries_json) ? s.entries_json : [],
          body_html: s.body_html || '',
        }))
      : []
    state.templates = (tplRes.data || []).map((t: any) => ({
      id: t.id,
      name: t.name,
      description: t.description || '',
      defaultColor: t.defaultColor || '',
      structure: t.structure || 'sheet-v1',
    }))
    // 3. 若当前分节已不存在，回落到基本信息
    if (String(activeKey.value).startsWith('section:')) {
      const sid = String(activeKey.value).slice('section:'.length)
      if (!state.sections.some((s) => s.id === sid)) {
        activeKey.value = 'profile'
      }
    }
    state.dirty = false
    hydrated = true
    lastSavedFp = contentFingerprint()
  } finally {
    state.loading = false
    // 等 Vue flush 完成后再打开脏监听，避免赋值触发误保存
    window.setTimeout(() => {
      skipDirtyWatch = false
    }, 120)
  }
}

/** 保存主信息与全部分节。 */
async function saveAll(opts?: { silent?: boolean }) {
  if ((!canUpdate && !canSaveSection) || !resumeId.value || state.saving) return
  const fp = contentFingerprint()
  if (opts?.silent && fp === lastSavedFp) {
    state.dirty = false
    return
  }
  state.saving = true
  try {
    // 1. 落库前钳制行高与字号，避免异常值写回
    state.theme.spacing.lineHeight = sanitizeLineHeight(state.theme.spacing.lineHeight)
    state.theme.fontSizeScale = sanitizeFontScale(state.theme.fontSizeScale)
    if (canUpdate) {
      await resumeApi.update({
        id: resumeId.value,
        title: state.title,
        template_id: state.template_id,
        theme_json: state.theme,
        custom_css: state.custom_css,
        profile_json: state.profile,
      })
    }
    // 2. 同步全部分节，避免切轨后未落库
    if (canSaveSection) {
      for (const s of state.sections) {
        await resumeApi.saveSection({
          resume_id: resumeId.value,
          id: s.id,
          type: s.type,
          title: s.title,
          icon: s.icon,
          sort_code: Number(s.sort_code || 0),
          meta_json: s.meta_json || {},
          body_html: s.body_html || '',
          entries_json: s.entries_json || [],
        })
      }
    }
    lastSavedFp = contentFingerprint()
    state.dirty = false
    if (!opts?.silent) {
      window.$message.success('已保存')
    }
  } finally {
    state.saving = false
  }
}

function scheduleSave() {
  if ((!canUpdate && !canSaveSection) || !hydrated) return
  state.dirty = true
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    void saveAll({ silent: true })
  }, 800)
}

function markDirty() {
  if (skipDirtyWatch || !hydrated || (!canUpdate && !canSaveSection) || state.saving) return
  if (contentFingerprint() === lastSavedFp) {
    state.dirty = false
    return
  }
  scheduleSave()
}

watch(
  () => [
    state.title,
    state.template_id,
    state.custom_css,
    state.theme,
    state.profile,
    state.sections,
  ],
  () => markDirty(),
  { deep: true },
)

function onSelect(key: RailKey) {
  activeKey.value = key
}

async function addSection() {
  if (!canSaveSection) return
  // 1. 先落库新分节
  const res = await resumeApi.saveSection({
    resume_id: resumeId.value,
    type: 'custom',
    title: '新分节',
    icon: 'lucide:file-text',
    sort_code: (state.sections.length + 1) * 10,
    meta_json: { visible: true },
    body_html: '',
    entries_json: [],
  })
  // 2. 刷新并选中
  await reload()
  activeKey.value = `section:${res.data.id}`
  window.$message.success('已添加模块')
}

async function removeActiveSection() {
  if (!canDeleteSection || !activeSection.value) return
  await resumeApi.removeSections(resumeId.value, [activeSection.value.id])
  activeKey.value = 'profile'
  window.$message.success('已删除分节')
  await reload()
}

/** 复制当前分节并选中副本。 */
async function copyActiveSection() {
  if (!canCopySection || !activeSection.value) return
  // 1. 先落库当前脏数据，避免复制到旧内容
  if (state.dirty) await saveAll({ silent: true })
  // 2. 后端复制
  const res = await resumeApi.copySection(resumeId.value, activeSection.value.id)
  window.$message.success('已复制本节')
  await reload()
  if (res.data?.id) {
    activeKey.value = `section:${res.data.id}`
  }
}

async function moveSection(dir: -1 | 1) {
  if (!canReorderSection || !activeSection.value) return
  const ids = state.sections.map((s) => s.id)
  const idx = ids.indexOf(activeSection.value.id)
  const j = idx + dir
  if (idx < 0 || j < 0 || j >= ids.length) return
  ;[ids[idx], ids[j]] = [ids[j], ids[idx]]
  await resumeApi.reorderSections(resumeId.value, ids)
  await reload()
  activeKey.value = `section:${ids[j]}`
}

/** 发布 / 更新并发布：先落库再走 publish（会重传 OSS 公开 PDF）。 */
async function onPublish() {
  if (!canPublish) return
  if (state.dirty) await saveAll()
  const wasPublished = state.status === 'published'
  await resumeApi.publish(resumeId.value)
  window.$message.success(wasPublished ? '已更新并发布' : '已发布')
  await reload()
}

async function onUnpublish() {
  if (!canUnpublish) return
  await resumeApi.unpublish(resumeId.value)
  window.$message.success('已取消发布')
  await reload()
}

function onBeforeUnload(e: BeforeUnloadEvent) {
  if (!state.dirty) return
  e.preventDefault()
  e.returnValue = ''
}

onMounted(() => {
  void reload()
  window.addEventListener('beforeunload', onBeforeUnload)
})

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', onBeforeUnload)
  if (saveTimer) clearTimeout(saveTimer)
})
</script>

<template>
  <div class="resume-wb">
    <WorkbenchHeader
      :title="state.title"
      :dirty="state.dirty"
      :saving="state.saving"
      :can-save="canUpdate || canSaveSection"
      :can-publish="canPublish"
      :can-unpublish="canUnpublish"
      :can-export="canExport"
      :status="state.status"
      @back="router.push('/resume/list')"
      @save="saveAll"
      @publish="onPublish"
      @unpublish="onUnpublish"
      @preview="goAdminPreview"
      @download="onDownload"
    />

    <div class="resume-wb__body">
      <div
        v-if="state.loading"
        class="resume-wb__loading"
      >
        <NSpin size="large" />
      </div>

      <!-- 左侧：模块轨 + 编辑面板（固定宽度，禁止撑破后台内容区） -->
      <aside class="resume-wb__editor">
        <div class="resume-wb__rail">
          <ModuleRail
            :active="activeKey"
            :sections="state.sections"
            :can-write="canSaveSection"
            @select="onSelect"
            @add="addSection"
          />
        </div>
        <NCard
          :bordered="false"
          class="resume-wb__panel"
          size="small"
          :segmented="{ content: true }"
          content-style="padding: 0; flex: 1 1 0%; min-height: 0; min-width: 0; max-width: 100%; display: flex; flex-direction: column; overflow: hidden"
          style="width: 100%; max-width: 100%; min-width: 0; overflow: hidden"
        >
          <template #header>
            <NText strong>
              {{ sideTitle }}
            </NText>
          </template>
          <NScrollbar class="resume-wb__panel-scroll">
            <div
              class="resume-edit-panel"
              :style="{
                '--line-height': String(sanitizeLineHeight(state.theme.spacing.lineHeight)),
                '--font-scale': String(sanitizeFontScale(state.theme.fontSizeScale)),
              }"
            >
              <ProfilePanel
                v-if="activeKey === 'profile'"
                v-model:profile="state.profile"
                v-model:title="state.title"
                :can-write="canWriteProfile"
              />
              <ThemePanel
                v-else-if="activeKey === 'theme'"
                v-model:theme="state.theme"
                v-model:custom-css="state.custom_css"
                v-model:template-id="state.template_id"
                :can-write="canWriteProfile"
                :templates="state.templates"
              />
              <SectionPanel
                v-else-if="activeSectionIndex >= 0"
                v-model:section="state.sections[activeSectionIndex]"
                :can-write="canWriteSection"
                :can-reorder="canReorderSection"
                :can-copy="canCopySection"
                :can-remove="canDeleteSection"
                @move-up="moveSection(-1)"
                @move-down="moveSection(1)"
                @remove="removeActiveSection"
                @copy="copyActiveSection"
              />
              <NEmpty v-else description="请选择模块" size="small" />
            </div>
          </NScrollbar>
        </NCard>
      </aside>

      <!-- 右侧：简历预览 -->
      <div class="resume-wb__preview">
        <NScrollbar style="max-height: 100%">
          <div class="resume-wb__preview-inner">
            <ResumeBackendPreview :doc="previewDoc" />
          </div>
        </NScrollbar>
      </div>
    </div>
  </div>
</template>
