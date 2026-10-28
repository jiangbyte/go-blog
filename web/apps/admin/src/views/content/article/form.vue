<!--
  Author: go-blog

  内容文章编辑页（对齐掘金：顶栏标题 + 全屏 Markdown，元数据在发布弹窗；默认草稿 + 自动保存）。
-->
<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import ImageUpload from '@/components/upload/ImageUpload.vue'
import { MdEditor } from '@/components/editor'
import { contentArticleApi, contentCategoryApi, contentTagApi } from '@/api'
import { createRequiredRule } from '@/utils'
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const publishFormRef = ref<FormInst | null>(null)
const editorHostRef = ref<HTMLElement | null>(null)
const listPath = '/content/article'
const AUTO_SAVE_MS = 1200
/** 串行化自动保存 / 手动发布，避免草稿自动保存覆盖已发布状态 */
let saveChain: Promise<void> = Promise.resolve()

function enqueueSave<T>(task: () => Promise<T>): Promise<T> {
  const run = saveChain.then(task, task)
  saveChain = run.then(
    () => undefined,
    () => undefined,
  )
  return run
}

const defaultFormData = {
  category_id: null as string | null,
  title: '',
  slug: '',
  summary: '',
  cover_url: '',
  body_md: '',
  status: 'draft' as 'draft' | 'published',
  tag_ids: [] as string[],
  sort_code: 0,
}

const state = reactive({
  loading: false,
  submitLoading: false,
  autoSaving: false,
  dirty: false,
  dataId: null as string | null,
  formModel: { ...defaultFormData, tag_ids: [] as string[] },
  categoryOptions: [] as { label: string; value: string }[],
  tagOptions: [] as { label: string; value: string }[],
  /** 发布弹窗：draft=保存草稿时补全元数据；publish=确认发布 */
  publishModalShow: false,
  publishModalMode: 'publish' as 'draft' | 'publish',
  editorHeight: 480,
  lastSavedAt: null as string | null,
  saveError: '',
})

/** 初始化未完成时不触发自动保存。 */
let hydrated = false
let skipDirtyWatch = false
let saveTimer: ReturnType<typeof setTimeout> | null = null
let lastSavedFp = ''
let resizeObserver: ResizeObserver | null = null

const statusHint = computed(() => {
  if (state.autoSaving || state.submitLoading) return '保存中…'
  if (state.saveError) return state.saveError
  if (state.dirty) return '有未保存的更改'
  if (state.lastSavedAt) {
    return state.formModel.status === 'published'
      ? `已自动保存 ${state.lastSavedAt}`
      : `已自动保存至草稿 ${state.lastSavedAt}`
  }
  if (state.formModel.status === 'published') return '已发布，修改将自动保存'
  return '文章将自动保存至草稿'
})

const publishModalTitle = computed(() =>
  state.publishModalMode === 'publish' ? '发布文章' : '完善文章信息',
)

const publishConfirmText = computed(() =>
  state.publishModalMode === 'publish' ? '确定并发布' : '保存草稿',
)

const publishRules = computed<FormRules>(() => ({
  category_id: [createRequiredRule('分类', 'change')],
}))

/** 内容指纹，用于跳过无实质变更的自动保存。 */
function contentFingerprint() {
  return JSON.stringify({
    category_id: state.formModel.category_id,
    title: state.formModel.title,
    summary: state.formModel.summary,
    cover_url: state.formModel.cover_url,
    body_md: state.formModel.body_md,
    tag_ids: state.formModel.tag_ids,
    sort_code: state.formModel.sort_code,
    status: state.formModel.status,
  })
}

function flattenCategories(nodes: any[], prefix = ''): { label: string; value: string }[] {
  const out: { label: string; value: string }[] = []
  for (const n of nodes || []) {
    out.push({ label: `${prefix}${n.name}`, value: n.id })
    if (n.children?.length) {
      out.push(...flattenCategories(n.children, `${prefix}${n.name} / `))
    }
  }
  return out
}

function resolveQueryId() {
  const id = route.query.id
  return typeof id === 'string' && id ? id : null
}

function syncEditorHeight() {
  const el = editorHostRef.value
  if (!el) return
  const h = Math.floor(el.clientHeight)
  if (h > 120) state.editorHeight = h
}

function bindEditorResize() {
  resizeObserver?.disconnect()
  const el = editorHostRef.value
  if (!el || typeof ResizeObserver === 'undefined') return
  resizeObserver = new ResizeObserver(() => syncEditorHeight())
  resizeObserver.observe(el)
  syncEditorHeight()
}

function clearSaveTimer() {
  if (saveTimer) {
    clearTimeout(saveTimer)
    saveTimer = null
  }
}

function formatSavedAt(d = new Date()) {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

async function loadOptions() {
  const [treeRes, tagRes] = await Promise.all([contentCategoryApi.tree(), contentTagApi.options()])
  state.categoryOptions = flattenCategories(treeRes.data || [])
  state.tagOptions = (tagRes.data || []).map((t) => ({ label: t.name, value: t.id }))
}

function fillFromRow(row: any) {
  if (!row) return
  state.formModel = {
    category_id: row.category_id || null,
    title: String(row.title ?? ''),
    slug: String(row.slug ?? ''),
    summary: String(row.summary ?? ''),
    // 优先稳定 object_name，避免预签名 URL 回写后再过期
    cover_url: String(row.cover_object_name || row.cover_url || ''),
    body_md: String(row.body_md ?? ''),
    status: row.status === 'published' ? 'published' : 'draft',
    tag_ids: Array.isArray(row.tag_ids) ? [...row.tag_ids] : [],
    sort_code: Number(row.sort_code ?? 0) || 0,
  }
}

async function fetchDetail(id: string) {
  state.loading = true
  try {
    const res = await contentArticleApi.detail(id)
    fillFromRow(res.data ?? {})
  } finally {
    state.loading = false
  }
}

async function initPage() {
  // 1. 暂停脏监听，避免回填触发保存
  clearSaveTimer()
  hydrated = false
  skipDirtyWatch = true
  state.dirty = false
  state.saveError = ''
  state.lastSavedAt = null
  // 2. 解析路由 id，区分新建 / 编辑
  const id = resolveQueryId()
  state.dataId = id
  state.formModel = { ...defaultFormData, tag_ids: [] }
  state.publishModalShow = false
  // 3. 拉分类与标签选项
  await loadOptions()
  // 4. 编辑态加载详情
  if (id) {
    await fetchDetail(id)
  }
  lastSavedFp = contentFingerprint()
  hydrated = true
  await nextTick()
  bindEditorResize()
  window.setTimeout(() => {
    skipDirtyWatch = false
  }, 120)
}

function goBack() {
  void router.push(listPath)
}

/** 是否有可落库内容（标题或正文任一非空即可自动建草稿）。 */
function hasSavableContent() {
  return Boolean(state.formModel.title.trim() || String(state.formModel.body_md || '').trim())
}

/** 发布前校验标题与正文。 */
function assertPublishReady(): boolean {
  const title = state.formModel.title.trim()
  if (!title || title === '无标题') {
    window.$message.warning('请输入文章标题')
    return false
  }
  if (!String(state.formModel.body_md || '').trim()) {
    window.$message.warning('请输入正文')
    return false
  }
  return true
}

function buildPayload(status: 'draft' | 'published') {
  return {
    category_id: state.formModel.category_id,
    title: state.formModel.title.trim(),
    summary: state.formModel.summary.trim() || null,
    cover_url: state.formModel.cover_url || null,
    body_md: state.formModel.body_md,
    status,
    tag_ids: state.formModel.tag_ids,
    sort_code: state.formModel.sort_code,
  }
}

/**
 * 静默自动保存：草稿保持 draft；已发布保持 published。
 * 新建无内容时跳过；成功后更新指纹与提示。
 */
async function autoSave() {
  if (!hydrated || state.loading || state.submitLoading || state.autoSaving) return
  const fp = contentFingerprint()
  if (fp === lastSavedFp) {
    state.dirty = false
    return
  }
  if (!state.dataId && !hasSavableContent()) {
    state.dirty = false
    lastSavedFp = fp
    return
  }
  await enqueueSave(async () => {
    // 手动提交进行中则让路，避免把 published 打回 draft
    if (state.submitLoading) return
    state.autoSaving = true
    state.saveError = ''
    try {
      // 1. 已发布文章自动保存不降级为草稿（以提交瞬间的 form 状态为准）
      const status = state.formModel.status === 'published' ? 'published' : 'draft'
      const payload = buildPayload(status)
      if (state.submitLoading) return
      // 2. 更新或首次创建草稿
      if (state.dataId) {
        await contentArticleApi.update({ id: state.dataId, ...payload })
      } else {
        const res = await contentArticleApi.create(payload)
        const newId = res.data?.id ? String(res.data.id) : null
        if (newId) {
          state.dataId = newId
          if (res.data?.slug) state.formModel.slug = String(res.data.slug)
          await router.replace({ path: '/content/article/edit', query: { id: newId } })
        }
      }
      // 3. 若期间已手动发布，勿把本地状态写回 draft
      if (status === 'draft' && state.formModel.status === 'published') {
        return
      }
      state.formModel.status = status
      lastSavedFp = contentFingerprint()
      state.dirty = false
      state.lastSavedAt = formatSavedAt()
    } catch (err: any) {
      state.saveError = err?.message || '自动保存失败'
    } finally {
      state.autoSaving = false
      if (state.dirty && !state.submitLoading && contentFingerprint() !== lastSavedFp) {
        scheduleAutoSave()
      }
    }
  })
}

function scheduleAutoSave() {
  if (!hydrated || skipDirtyWatch) return
  state.dirty = true
  clearSaveTimer()
  saveTimer = setTimeout(() => {
    void autoSave()
  }, AUTO_SAVE_MS)
}

function markDirty() {
  if (skipDirtyWatch || !hydrated || state.submitLoading) return
  if (contentFingerprint() === lastSavedFp) {
    state.dirty = false
    return
  }
  scheduleAutoSave()
}

function openPublishModal(mode: 'draft' | 'publish') {
  if (mode === 'publish' && !assertPublishReady()) return
  if (mode === 'draft' && !hasSavableContent()) {
    window.$message.warning('请先输入标题或正文')
    return
  }
  state.publishModalMode = mode
  state.publishModalShow = true
}

/**
 * 按指定状态落库（手动）；发布后回列表，草稿留页。
 */
async function submitWithStatus(status: 'draft' | 'published') {
  if (status === 'published' && !assertPublishReady()) return
  if (status === 'published' && !state.formModel.category_id) {
    window.$message.warning('请选择分类')
    return
  }
  if (status === 'draft' && !hasSavableContent()) {
    window.$message.warning('请先输入标题或正文')
    return
  }
  clearSaveTimer()
  await enqueueSave(async () => {
    state.submitLoading = true
    state.saveError = ''
    try {
      // 先写入本地状态，避免并发自动保存仍按 draft 组装 payload
      state.formModel.status = status
      const payload = buildPayload(status)
      if (state.dataId) {
        await contentArticleApi.update({ id: state.dataId, ...payload })
        window.$message.success(status === 'published' ? '已发布' : '已保存草稿')
      } else {
        const res = await contentArticleApi.create(payload)
        const newId = res.data?.id ? String(res.data.id) : null
        window.$message.success(status === 'published' ? '已发布' : '已保存草稿')
        if (newId) {
          state.dataId = newId
          if (res.data?.slug) state.formModel.slug = String(res.data.slug)
          await router.replace({ path: '/content/article/edit', query: { id: newId } })
        }
      }
      lastSavedFp = contentFingerprint()
      state.dirty = false
      state.lastSavedAt = formatSavedAt()
      state.publishModalShow = false
      if (status === 'published') {
        goBack()
      }
    } finally {
      state.submitLoading = false
    }
  })
}

/**
 * 顶栏保存：已发布保持 published；草稿保持 draft。
 * 避免已发布文章误点「保存草稿」把库状态打回 draft。
 */
function saveNow() {
  const status = state.formModel.status === 'published' ? 'published' : 'draft'
  void submitWithStatus(status)
}

function openPublish() {
  openPublishModal('publish')
}

async function confirmPublishModal() {
  await publishFormRef.value?.validate()
  const status = state.publishModalMode === 'publish' ? 'published' : 'draft'
  await submitWithStatus(status)
}

watch(
  () => [
    state.formModel.title,
    state.formModel.body_md,
    state.formModel.summary,
    state.formModel.cover_url,
    state.formModel.category_id,
    state.formModel.tag_ids,
    state.formModel.sort_code,
  ],
  () => markDirty(),
  { deep: true },
)

onMounted(() => {
  void initPage()
})

onUnmounted(() => {
  clearSaveTimer()
  resizeObserver?.disconnect()
  resizeObserver = null
})

watch(
  () => route.query.id,
  (id) => {
    // 新建草稿后 replace 到编辑路由时 dataId 已写入，跳过重复拉详情
    const next = typeof id === 'string' && id ? id : null
    if (next === state.dataId) return
    void initPage()
  },
)
</script>

<template>
  <NCard
    class="h-full min-h-0"
    :bordered="false"
    content-style="padding: 0; height: 100%; display: flex; flex-direction: column; min-height: 0;"
  >
    <NFlex
      class="px-4 h-14 shrink-0"
      align="center"
      justify="space-between"
      :size="16"
      style="border-bottom: 1px solid var(--n-divider-color)"
    >
      <NInput
        v-model:value="state.formModel.title"
        class="flex-1 min-w-0"
        :bordered="false"
        size="large"
        maxlength="200"
        placeholder="输入文章标题..."
        :disabled="state.loading || state.submitLoading"
        style="font-size: 20px; font-weight: 600"
      />
      <NFlex class="shrink-0" align="center" :size="10">
        <NText depth="3" style="white-space: nowrap">
          {{ statusHint }}
        </NText>
        <NButton quaternary :disabled="state.submitLoading || state.autoSaving" @click="goBack">
          返回
        </NButton>
        <NButton
          :loading="state.submitLoading"
          :disabled="state.loading || state.autoSaving"
          @click="saveNow"
        >
          {{ state.formModel.status === 'published' ? '保存' : '保存草稿' }}
        </NButton>
        <NButton
          type="primary"
          :loading="state.submitLoading"
          :disabled="state.loading || state.autoSaving"
          @click="openPublish"
        >
          {{ state.formModel.status === 'published' ? '更新发布' : '发布' }}
        </NButton>
      </NFlex>
    </NFlex>

    <NSpin
      class="flex-1 min-h-0"
      :show="state.loading"
      content-style="height: 100%; display: flex; flex-direction: column; min-height: 0;"
    >
      <div ref="editorHostRef" class="flex-1 min-h-0 w-full">
        <MdEditor
          v-model:value="state.formModel.body_md"
          :height="state.editorHeight"
          placeholder="开始写作..."
          :disabled="state.loading || state.submitLoading"
        />
      </div>
    </NSpin>

    <NModal
      v-model:show="state.publishModalShow"
      preset="card"
      :title="publishModalTitle"
      style="width: 640px; max-width: calc(100vw - 32px)"
      :mask-closable="!state.submitLoading"
      :closable="!state.submitLoading"
    >
      <NForm
        ref="publishFormRef"
        :model="state.formModel"
        :rules="publishRules"
        label-placement="left"
        label-width="88"
        :disabled="state.submitLoading"
      >
        <NFormItem label="分类" path="category_id" required>
          <NSelect
            v-model:value="state.formModel.category_id"
            :options="state.categoryOptions"
            filterable
            clearable
            placeholder="请选择分类"
          />
        </NFormItem>

        <NFormItem label="添加标签" path="tag_ids">
          <NSelect
            v-model:value="state.formModel.tag_ids"
            :options="state.tagOptions"
            multiple
            filterable
            clearable
            placeholder="请搜索添加标签"
          />
        </NFormItem>

        <NFormItem label="文章封面" path="cover_url">
          <NFlex vertical :size="6">
            <ImageUpload
              v-model:value="state.formModel.cover_url"
              variant="card"
              button-text="上传封面"
            />
            <NText depth="3" style="font-size: 12px">
              建议尺寸：192×128px（封面展示在列表信息流中）
            </NText>
          </NFlex>
        </NFormItem>

        <NFormItem label="编辑摘要" path="summary">
          <NInput
            v-model:value="state.formModel.summary"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 5 }"
            maxlength="500"
            show-count
            placeholder="请输入文章摘要"
          />
        </NFormItem>

        <NFormItem v-if="state.dataId" label="标识" path="slug">
          <NInput :value="state.formModel.slug" disabled />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton :disabled="state.submitLoading" @click="state.publishModalShow = false">
            取消
          </NButton>
          <NButton type="primary" :loading="state.submitLoading" @click="confirmPublishModal">
            {{ publishConfirmText }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>
