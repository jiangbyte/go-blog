<!-- Author: go-blog — 项目新建 / 编辑（可见性；私有仅保存无发布） -->

<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import ImageGalleryUpload from '@/components/upload/ImageGalleryUpload.vue'
import { projectApi, projectCategoryApi } from '@/api'
import { createRequiredRule } from '@/utils'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInst | null>(null)
const listPath = '/project/manage'

const defaultFormData = {
  name: '',
  tech_stack: [] as string[],
  role: '',
  description: '',
  images: [] as string[],
  start_date: null as number | null,
  end_date: null as number | null,
  end_present: false,
  status: 'draft' as 'draft' | 'published',
  visibility: 'public' as 'public' | 'private',
  category_id: null as string | null,
  sort_code: 0,
}

const state = reactive({
  loading: false,
  submitLoading: false,
  dataId: null as string | null,
  /** 仅提交时带回，界面不展示 */
  slug: '',
  categoryOptions: [] as { label: string; value: string }[],
  formModel: {
    ...defaultFormData,
    tech_stack: [] as string[],
    images: [] as string[],
  },
})

const pageTitle = computed(() => (state.dataId ? '编辑项目' : '新建项目'))

const isPrivate = computed(() => state.formModel.visibility === 'private')
const isPublished = computed(() => !isPrivate.value && state.formModel.status === 'published')

/** 顶栏状态提示。 */
const statusHint = computed(() => {
  if (!state.dataId) return '未保存'
  if (isPrivate.value) return '私有（仅保存）'
  return isPublished.value ? '已发布' : '草稿'
})

const rules = computed<FormRules>(() => ({
  name: [createRequiredRule('项目名称', 'input')],
  category_id: [createRequiredRule('分类', 'change')],
}))

const visibilityOptions = [
  { label: '公开', value: 'public' },
  { label: '私有', value: 'private' },
]

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

/** YYYY-MM-DD → 时间戳（本地日初），供 DatePicker。 */
function dateStringToTs(raw: string | null | undefined): number | null {
  const text = String(raw || '').trim()
  if (!text) return null
  const d = new Date(`${text}T00:00:00`)
  if (Number.isNaN(d.getTime())) return null
  return d.getTime()
}

/** DatePicker 值 → YYYY-MM-DD。 */
function tsToDateString(value: number | null | undefined): string | null {
  if (value == null) return null
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return null
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function fillFromRow(row: any) {
  if (!row) return
  const start = dateStringToTs(row.start_date)
  const end = dateStringToTs(row.end_date)
  // 有开始、无结束 → 视为至今
  const endPresent = Boolean(start) && !end
  const visibility = row.visibility === 'private' ? 'private' : 'public'
  state.slug = String(row.slug ?? '')
  state.formModel = {
    name: String(row.name ?? ''),
    tech_stack: Array.isArray(row.tech_stack) ? [...row.tech_stack] : [],
    role: String(row.role ?? ''),
    description: String(row.description ?? ''),
    images: Array.isArray(row.images) ? [...row.images] : [],
    start_date: start,
    end_date: endPresent ? null : end,
    end_present: endPresent,
    status: visibility === 'private' ? 'draft' : row.status === 'published' ? 'published' : 'draft',
    visibility,
    category_id: row.category_id || null,
    sort_code: Number(row.sort_code ?? 0) || 0,
  }
}

function buildPayload(status: 'draft' | 'published') {
  const visibility = state.formModel.visibility
  return {
    name: state.formModel.name.trim(),
    // 编辑保留原标识；新建不传由后端随机分配
    slug: state.dataId ? state.slug || null : null,
    tech_stack: state.formModel.tech_stack,
    role: state.formModel.role.trim() || null,
    description: state.formModel.description,
    images: state.formModel.images,
    start_date: tsToDateString(state.formModel.start_date),
    end_date: state.formModel.end_present ? null : tsToDateString(state.formModel.end_date),
    status: visibility === 'private' ? 'draft' : status,
    visibility,
    category_id: state.formModel.category_id,
    sort_code: Number(state.formModel.sort_code ?? 0) || 0,
  }
}

async function loadCategories() {
  const res = await projectCategoryApi.tree()
  state.categoryOptions = flattenCategories(res.data || [])
}

async function loadDetail() {
  const id = resolveQueryId()
  state.dataId = id
  if (!id) return
  state.loading = true
  try {
    const res = await projectApi.detail(id)
    fillFromRow(res.data)
  } finally {
    state.loading = false
  }
}

/**
 * 按指定状态落库；发布后回列表，草稿/已发布「保存」留页。
 * 私有一律按 draft 保存。
 */
async function submitWithStatus(
  status: 'draft' | 'published',
  opts?: { leaveAfterPublish?: boolean },
) {
  await formRef.value?.validate()
  const effectiveStatus = isPrivate.value ? 'draft' : status
  const leaveAfterPublish = opts?.leaveAfterPublish ?? effectiveStatus === 'published'
  state.submitLoading = true
  try {
    // 先写入本地状态，保证 payload 与后续 fill 一致
    state.formModel.status = effectiveStatus
    const payload = buildPayload(effectiveStatus)
    if (state.dataId) {
      const res = await projectApi.update({ id: state.dataId, ...payload })
      fillFromRow(res.data)
    } else {
      const res = await projectApi.create(payload)
      const id = String(res.data?.id || '')
      if (id) {
        state.dataId = id
        fillFromRow(res.data)
        if (!(effectiveStatus === 'published' && leaveAfterPublish)) {
          void router.replace({ path: '/project/manage/edit', query: { id } })
        }
      }
    }
    if (effectiveStatus === 'published') {
      window.$message.success(leaveAfterPublish ? '已发布' : '已保存')
      if (leaveAfterPublish) {
        goBack()
      }
    } else {
      window.$message.success(isPrivate.value ? '已保存' : '已保存草稿')
    }
  } finally {
    state.submitLoading = false
  }
}

/** 顶栏保存：私有/草稿保持 draft；公开已发布保持 published。 */
function saveNow() {
  if (isPrivate.value) {
    void submitWithStatus('draft', { leaveAfterPublish: false })
    return
  }
  const status = state.formModel.status === 'published' ? 'published' : 'draft'
  void submitWithStatus(status, { leaveAfterPublish: false })
}

/** 顶栏发布 / 更新发布（仅公开）。 */
function publishNow() {
  if (isPrivate.value) return
  void submitWithStatus('published', { leaveAfterPublish: true })
}

function goBack() {
  void router.push(listPath)
}

/** 勾选至今时清空并禁用结束日期。 */
function onEndPresentUpdate(checked: boolean) {
  state.formModel.end_present = checked
  if (checked) {
    state.formModel.end_date = null
  }
}

/** 切到私有时强制本地 status=draft。 */
function onVisibilityUpdate(value: 'public' | 'private') {
  state.formModel.visibility = value
  if (value === 'private') {
    state.formModel.status = 'draft'
  }
}

watch(
  () => state.formModel.start_date,
  (start) => {
    // 无开始日期时不能勾「至今」
    if (start == null && state.formModel.end_present) {
      state.formModel.end_present = false
    }
    // 结束早于开始时纠正
    if (start != null && state.formModel.end_date != null && state.formModel.end_date < start) {
      state.formModel.end_date = start
    }
  },
)

onMounted(() => {
  void loadCategories()
  void loadDetail()
})
</script>

<template>
  <NSpin class="project-form h-full min-h-0" :show="state.loading">
    <NCard class="project-form__card h-full min-h-0" :title="pageTitle" :bordered="false">
      <template #header-extra>
        <NSpace align="center">
          <NText depth="3" style="white-space: nowrap">
            {{ statusHint }}
          </NText>
          <NButton quaternary :disabled="state.submitLoading" @click="goBack"> 返回 </NButton>
          <NButton :loading="state.submitLoading" :disabled="state.loading" @click="saveNow">
            {{ isPrivate || !isPublished ? (isPrivate ? '保存' : '保存草稿') : '保存' }}
          </NButton>
          <NButton
            v-if="!isPrivate"
            type="primary"
            :loading="state.submitLoading"
            :disabled="state.loading"
            @click="publishNow"
          >
            {{ isPublished ? '更新发布' : '发布' }}
          </NButton>
        </NSpace>
      </template>

      <NForm
        ref="formRef"
        class="project-form__body"
        :model="state.formModel"
        :rules="rules"
        label-placement="top"
        :disabled="state.loading || state.submitLoading"
      >
        <NGrid :cols="4" :x-gap="16" :y-gap="0">
          <NGi :span="2">
            <NFormItem label="项目名称" path="name">
              <NInput
                v-model:value="state.formModel.name"
                maxlength="200"
                show-count
                placeholder="项目名称"
              />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="可见性" path="visibility">
              <NSelect
                :value="state.formModel.visibility"
                :options="visibilityOptions"
                @update:value="onVisibilityUpdate"
              />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="分类" path="category_id">
              <NSelect
                v-model:value="state.formModel.category_id"
                :options="state.categoryOptions"
                filterable
                placeholder="请选择分类"
              />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="负责身份" path="role">
              <NInput
                v-model:value="state.formModel.role"
                maxlength="128"
                show-count
                placeholder="如：负责人 / 核心开发"
              />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="排序" path="sort_code">
              <NInputNumber
                v-model:value="state.formModel.sort_code"
                class="w-full"
                :show-button="false"
              />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="开始日期" path="start_date">
              <NDatePicker
                v-model:value="state.formModel.start_date"
                class="w-full"
                type="date"
                clearable
              />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="结束日期" path="end_date">
              <NFlex class="w-full" align="center" :size="8">
                <NDatePicker
                  v-model:value="state.formModel.end_date"
                  class="flex-1 min-w-0"
                  type="date"
                  clearable
                  :disabled="state.formModel.end_present"
                  :placeholder="state.formModel.end_present ? '至今' : '选择结束日期'"
                />
                <NCheckbox
                  :checked="state.formModel.end_present"
                  :disabled="state.formModel.start_date == null"
                  @update:checked="onEndPresentUpdate"
                >
                  至今
                </NCheckbox>
              </NFlex>
            </NFormItem>
          </NGi>
          <NGi :span="2">
            <NFormItem label="技术栈" path="tech_stack">
              <NDynamicTags v-model:value="state.formModel.tech_stack" :max="32" />
            </NFormItem>
          </NGi>
          <NGi :span="4">
            <NFormItem label="图片合集" path="images">
              <ImageGalleryUpload
                v-model:value="state.formModel.images"
                class="w-full"
                :max="12"
                :disabled="state.loading || state.submitLoading"
              />
            </NFormItem>
          </NGi>
          <NGi :span="4">
            <NFormItem label="项目描述" path="description">
              <NInput
                v-model:value="state.formModel.description"
                type="textarea"
                :rows="10"
                maxlength="10000"
                show-count
                placeholder="项目描述（纯文本）"
              />
            </NFormItem>
          </NGi>
        </NGrid>
      </NForm>
    </NCard>
  </NSpin>
</template>

<style scoped>
.project-form :deep(.n-spin-content) {
  height: 100%;
  min-height: 0;
}

.project-form__card {
  display: flex;
  flex-direction: column;
}

.project-form__card :deep(.n-card__content) {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.project-form__body {
  width: 100%;
  max-width: none;
}
</style>
