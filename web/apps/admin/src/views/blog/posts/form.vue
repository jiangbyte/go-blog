<!--
  Author: go-blog
  文章编辑（对齐 mypage 掘金式：保留 admin 布局；内容区顶栏标题 + 铺满 Markdown；元数据在弹窗）。
-->

<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { MdSplitWorkspace } from '@/components/editor'
import BlogCoverUpload from '@/views/blog/components/BlogCoverUpload.vue'
import TaxonomySelect from '@/views/blog/components/TaxonomySelect.vue'
import { blogPostsApi, blogSiteApi } from '@/api'
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const metaFormRef = ref<FormInst | null>(null)
const editorHostRef = ref<HTMLElement | null>(null)
const listPath = '/blog/posts'

const defaultFormData = {
  title: '',
  description: '',
  cover: '',
  date: null as number | null,
  status: 'draft' as 'draft' | 'published',
  category_uids: [] as string[],
  tag_uids: [] as string[],
  reprint: false,
  sticky: false,
  home: true,
  body: '',
}

const state = reactive({
  loading: false,
  submitLoading: false,
  dataUid: null as string | null,
  formModel: { ...defaultFormData, category_uids: [] as string[], tag_uids: [] as string[] },
  categoryOptions: [] as { label: string; value: string }[],
  tagOptions: [] as { label: string; value: string }[],
  metaModalShow: false,
  editorHeight: 480,
})

let resizeObserver: ResizeObserver | null = null

const metaRules: FormRules = {}

const statusHint = computed(() => {
  if (!state.dataUid) return '未保存的草稿'
  return state.formModel.status === 'published' ? '已发布' : '草稿'
})

function resolveQueryUid() {
  const uid = route.query.uid
  return typeof uid === 'string' && uid ? uid : null
}

function dateStringToTs(raw: string | null | undefined): number | null {
  const text = String(raw || '').trim()
  if (!text) return null
  const d = new Date(`${text.slice(0, 10)}T00:00:00`)
  if (Number.isNaN(d.getTime())) return null
  return d.getTime()
}

function tsToDateString(value: number | null | undefined): string | null {
  if (value == null) return null
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return null
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function taxonUids(row: any, listKey: string, uidsKey: string): string[] {
  if (Array.isArray(row[uidsKey])) return [...row[uidsKey]]
  const list = row[listKey]
  if (Array.isArray(list)) {
    return list.map((item: any) => String(item.uid || '')).filter(Boolean)
  }
  return []
}

function fillFromRow(row: any) {
  if (!row) return
  state.formModel = {
    title: String(row.title ?? ''),
    description: String(row.description ?? ''),
    cover: String(row.cover ?? ''),
    date: dateStringToTs(row.date || row.published_at),
    status: row.status === 'published' ? 'published' : 'draft',
    category_uids: taxonUids(row, 'categories', 'category_uids'),
    tag_uids: taxonUids(row, 'tags', 'tag_uids'),
    reprint: Boolean(row.reprint),
    sticky: Boolean(row.sticky),
    home: row.home !== undefined ? Boolean(row.home) : row.home_visible !== false,
    body: String(row.body ?? ''),
  }
}

function buildPayload(status: 'draft' | 'published') {
  return {
    title: state.formModel.title.trim(),
    description: state.formModel.description.trim() || null,
    date: tsToDateString(state.formModel.date),
    status,
    cover: state.formModel.cover.trim() || null,
    category_uids: state.formModel.category_uids,
    tag_uids: state.formModel.tag_uids,
    reprint: state.formModel.reprint,
    sticky: state.formModel.sticky,
    home: state.formModel.home,
    body: state.formModel.body,
  }
}

async function loadOptions() {
  const [catRes, tagRes] = await Promise.all([
    blogSiteApi.taxonomyOptions('category'),
    blogSiteApi.taxonomyOptions('tag'),
  ])
  state.categoryOptions = (catRes.data || []).map((t) => ({ label: t.name, value: t.uid }))
  state.tagOptions = (tagRes.data || []).map((t) => ({ label: t.name, value: t.uid }))
}

async function loadDetail(uid: string) {
  state.loading = true
  try {
    const res = await blogPostsApi.detail(uid)
    fillFromRow(res.data)
  } finally {
    state.loading = false
  }
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

async function initPage() {
  // 1. 解析 uid
  const uid = resolveQueryUid()
  state.dataUid = uid
  state.formModel = { ...defaultFormData, category_uids: [], tag_uids: [] }
  state.metaModalShow = false
  // 2. 选项 + 详情
  await loadOptions()
  if (uid) await loadDetail(uid)
  // 3. 按内容区剩余高度铺满编辑器
  await nextTick()
  bindEditorResize()
}

async function submitWithStatus(
  status: 'draft' | 'published',
  opts?: { leaveAfterPublish?: boolean },
) {
  // 1. 校验：草稿只需标题；发布需标题与正文
  const title = state.formModel.title.trim()
  const body = state.formModel.body.trim()
  if (!title) {
    window.$message.warning('请填写标题')
    return
  }
  if (status === 'published' && !body) {
    window.$message.warning('发布前请填写正文')
    return
  }
  // 2. 按目标状态写入；草稿留页；点发布则回列表
  state.submitLoading = true
  try {
    const payload = buildPayload(status)
    if (state.dataUid) {
      const res = await blogPostsApi.update({ uid: state.dataUid, ...payload })
      fillFromRow(res.data)
      window.$message.success(status === 'published' ? '已发布' : '已保存草稿')
    } else {
      const res = await blogPostsApi.create(payload)
      const newUid = String(res.data?.uid || '')
      if (newUid) {
        state.dataUid = newUid
        fillFromRow(res.data)
        if (!(status === 'published' && opts?.leaveAfterPublish)) {
          await router.replace({ path: '/blog/posts/edit', query: { uid: newUid } })
        }
      }
      window.$message.success(status === 'published' ? '已发布' : '已保存草稿')
    }
    state.metaModalShow = false
    if (status === 'published' && opts?.leaveAfterPublish) {
      goBack()
    }
  } finally {
    state.submitLoading = false
  }
}

function saveDraft() {
  void submitWithStatus('draft')
}

function publish() {
  void submitWithStatus('published', { leaveAfterPublish: true })
}

/** 元数据弹窗确认：按当前状态落库（新建按草稿） */
function saveMeta() {
  const status = state.formModel.status === 'published' ? 'published' : 'draft'
  void submitWithStatus(status)
}

function openMeta() {
  state.metaModalShow = true
}

function goBack() {
  void router.push(listPath)
}

onMounted(() => {
  void initPage()
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
})

watch(
  () => route.query.uid,
  (uid) => {
    const next = typeof uid === 'string' && uid ? uid : null
    if (next === state.dataUid) return
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
      :wrap="false"
      :size="16"
      style="border-bottom: 1px solid var(--n-divider-color)"
    >
      <NInput
        v-model:value="state.formModel.title"
        class="flex-1 min-w-0"
        :bordered="false"
        size="large"
        maxlength="255"
        placeholder="输入文章标题..."
        :disabled="state.loading || state.submitLoading"
        style="font-size: 20px; font-weight: 600"
      />
      <NFlex class="shrink-0" align="center" :wrap="false" :size="10">
        <NText depth="3" style="white-space: nowrap">
          {{ statusHint }}
        </NText>
        <NButton quaternary :disabled="state.submitLoading" @click="goBack">返回</NButton>
        <NButton :disabled="state.loading || state.submitLoading" @click="openMeta">
          完善信息
        </NButton>
        <NButton
          :loading="state.submitLoading"
          :disabled="state.loading"
          @click="saveDraft"
        >
          保存草稿
        </NButton>
        <NButton
          type="primary"
          :loading="state.submitLoading"
          :disabled="state.loading"
          @click="publish"
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
        <MdSplitWorkspace
          v-model:value="state.formModel.body"
          :height="state.editorHeight"
          placeholder="开始写作..."
          :disabled="state.loading || state.submitLoading"
        />
      </div>
    </NSpin>

    <NModal
      v-model:show="state.metaModalShow"
      preset="card"
      title="完善文章信息"
      class="blog-meta-modal"
      style="width: 640px; max-width: calc(100vw - 32px); max-height: min(680px, calc(100vh - 48px))"
      content-style="max-height: min(480px, calc(100vh - 220px)); overflow-y: auto;"
      :mask-closable="!state.submitLoading"
      :closable="!state.submitLoading"
    >
      <NForm
        ref="metaFormRef"
        :model="state.formModel"
        :rules="metaRules"
        label-placement="left"
        label-width="88"
        :disabled="state.submitLoading"
      >
        <NFormItem label="发布日期" path="date">
          <NDatePicker
            v-model:value="state.formModel.date"
            class="w-full"
            type="date"
            clearable
          />
        </NFormItem>
        <NFormItem label="摘要" path="description">
          <NInput
            v-model:value="state.formModel.description"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 5 }"
            placeholder="文章描述"
          />
        </NFormItem>
        <NFormItem label="封面" path="cover">
          <BlogCoverUpload v-model:value="state.formModel.cover" />
        </NFormItem>
        <NFormItem label="分类" path="category_uids">
          <TaxonomySelect
            v-model:value="state.formModel.category_uids"
            v-model:options="state.categoryOptions"
            kind="category"
            placeholder="选择或输入后回车创建"
          />
        </NFormItem>
        <NFormItem label="标签" path="tag_uids">
          <TaxonomySelect
            v-model:value="state.formModel.tag_uids"
            v-model:options="state.tagOptions"
            kind="tag"
            placeholder="选择或输入后回车创建"
          />
        </NFormItem>
        <NFormItem label="置顶">
          <NSwitch v-model:value="state.formModel.sticky" />
        </NFormItem>
        <NFormItem label="首页展示">
          <NSwitch v-model:value="state.formModel.home" />
        </NFormItem>
        <NFormItem label="转载">
          <NSwitch v-model:value="state.formModel.reprint" />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton :disabled="state.submitLoading" @click="state.metaModalShow = false">
            取消
          </NButton>
          <NButton type="primary" :loading="state.submitLoading" @click="saveMeta">保存</NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>

<style>
/* NModal 挂到 body，需非 scoped */
.blog-meta-modal.n-card {
  display: flex;
  flex-direction: column;
  height: auto !important;
  max-height: min(680px, calc(100vh - 48px));
}

.blog-meta-modal .n-card__content {
  flex: 0 1 auto;
  min-height: 0;
}
</style>
