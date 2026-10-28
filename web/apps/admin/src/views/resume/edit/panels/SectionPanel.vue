<!-- Author: go-blog — 分节侧栏面板（按类型分流正文 / 条目；条目用折叠） -->

<script setup lang="ts">
import { Icon } from '@iconify/vue/offline'
import IconSelect from '@/components/common/IconSelect.vue'
import PeriodRangePicker from '../components/PeriodRangePicker.vue'
import ResumeBodyEditor from '../ResumeBodyEditor.vue'
import {
  NButton,
  NCollapse,
  NCollapseItem,
  NEmpty,
  NFlex,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NSelect,
  NSpace,
  NSwitch,
  NText,
} from 'naive-ui'
import { computed, ref, watch } from 'vue'

const section = defineModel<{
  id: string
  type: string
  title: string
  icon: string
  meta_json: Record<string, any>
  body_html: string
  entries_json: {
    title: string
    role: string
    period: string
    body_html: string
    tags?: string[]
  }[]
}>('section', { required: true })

const props = defineProps<{
  canWrite: boolean
  canReorder?: boolean
  canCopy?: boolean
  canRemove?: boolean
}>()

const emit = defineEmits<{
  moveUp: []
  moveDown: []
  remove: []
  copy: []
}>()

const canReorder = computed(() => props.canReorder ?? props.canWrite)
const canCopySec = computed(() => props.canCopy ?? props.canWrite)
const canRemoveSec = computed(() => props.canRemove ?? props.canWrite)

/** 当前展开的条目 name（允许多开） */
const expandedNames = ref<string[]>([])

const SECTION_TITLES: Record<string, string> = {
  education: '教育背景',
  experience: '工作经历',
  skills: '专业技能',
  projects: '项目经历',
  evaluation: '自我评价',
  custom: '自定义',
}

const SECTION_TYPE_OPTIONS = [
  { label: '教育背景', value: 'education' },
  { label: '工作经历', value: 'experience' },
  { label: '专业技能', value: 'skills' },
  { label: '项目经历', value: 'projects' },
  { label: '自我评价', value: 'evaluation' },
  { label: '自定义', value: 'custom' },
]

/** 取分节默认中文标题。 */
function defaultTitleForType(type: string) {
  return SECTION_TITLES[type] || SECTION_TITLES.custom
}

/** 切换类型时，若标题仍是旧默认标题则同步为新默认。 */
function onTypeChange(type: string) {
  const prevDefaults = Object.values(SECTION_TITLES)
  if (!section.value.title || prevDefaults.includes(section.value.title)) {
    section.value.title = defaultTitleForType(type)
  }
  if (
    (type === 'experience' || type === 'projects') &&
    !Array.isArray(section.value.entries_json)
  ) {
    section.value.entries_json = []
  }
}

/** 是否走条目列表（工作经历 / 项目）。 */
const isEntryType = computed(
  () => section.value.type === 'experience' || section.value.type === 'projects',
)

/** 是否展示节级正文（HTML）。 */
const showSectionBody = computed(() => !isEntryType.value)

/** 条目标题字段文案随类型变化。 */
const entryTitleLabel = computed(() => (section.value.type === 'projects' ? '项目名' : '公司'))

const bodyDividerLabel = computed(() => {
  if (section.value.type === 'education') return '教育说明 / 成绩获奖等'
  if (section.value.type === 'skills') return '技能说明'
  if (section.value.type === 'evaluation') return '自我评价'
  return '正文'
})

/** 纸面是否展示本节；缺省 true，写入 meta_json.visible（仅 boolean）。 */
const sectionVisible = computed({
  get: () => section.value.meta_json?.visible !== false,
  set: (v: boolean) => {
    if (!section.value.meta_json || typeof section.value.meta_json !== 'object') {
      section.value.meta_json = {}
    }
    section.value.meta_json.visible = Boolean(v)
  },
})

function entryName(idx: number) {
  return `entry-${idx}`
}

/** 折叠头摘要：标题 · 角色 · 时间。 */
function entrySummary(entry: {
  title?: string
  role?: string
  period?: string
}, idx: number) {
  const title = String(entry.title || '').trim() || `${entryTitleLabel.value}${idx + 1}`
  const parts = [title]
  const role = String(entry.role || '').trim()
  const period = String(entry.period || '').trim()
  if (role) parts.push(role)
  if (period) parts.push(period)
  return parts.join(' · ')
}

function addEntry() {
  if (!Array.isArray(section.value.entries_json)) {
    section.value.entries_json = []
  }
  section.value.entries_json.push({
    title: '',
    role: '',
    period: '',
    body_html: '',
    tags: [],
  })
  // 1. 新条目默认展开，方便立刻填写
  const name = entryName(section.value.entries_json.length - 1)
  if (!expandedNames.value.includes(name)) {
    expandedNames.value = [...expandedNames.value, name]
  }
}

function removeEntry(idx: number) {
  section.value.entries_json?.splice(idx, 1)
  // 2. 删除后重建展开 name，避免索引错位
  expandedNames.value = expandedNames.value
    .map((n) => {
      const m = /^entry-(\d+)$/.exec(n)
      if (!m) return n
      const i = Number(m[1])
      if (i === idx) return ''
      if (i > idx) return entryName(i - 1)
      return n
    })
    .filter(Boolean)
}

/** 条目上下调序。 */
function moveEntry(idx: number, dir: -1 | 1) {
  const list = section.value.entries_json
  if (!Array.isArray(list)) return
  const j = idx + dir
  if (j < 0 || j >= list.length) return
  ;[list[idx], list[j]] = [list[j], list[idx]]
  // 3. 展开态跟随条目移动
  const a = entryName(idx)
  const b = entryName(j)
  expandedNames.value = expandedNames.value.map((n) => {
    if (n === a) return b
    if (n === b) return a
    return n
  })
}

function tagsText(entry: { tags?: string[] }) {
  return Array.isArray(entry.tags) ? entry.tags.join(',') : ''
}

function setTags(entry: { tags?: string[] }, v: string) {
  entry.tags = String(v || '')
    .split(/[,，]/)
    .map((s) => s.trim())
    .filter(Boolean)
}

// 切到条目型分节时，若尚无展开项则默认展开第一条
watch(
  () => [isEntryType.value, section.value.entries_json?.length] as const,
  ([isEntry, len]) => {
    if (!isEntry || !len) {
      expandedNames.value = []
      return
    }
    if (!expandedNames.value.length) {
      expandedNames.value = [entryName(0)]
    }
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <NFlex class="mb-3" justify="space-between" align="center" :size="8">
      <NSpace :size="4">
        <NButton
          size="small"
          quaternary
          title="上移"
          aria-label="上移"
          :disabled="!canReorder"
          @click="emit('moveUp')"
        >
          <template #icon>
            <NIcon>
              <Icon icon="icon-park-outline:up" />
            </NIcon>
          </template>
        </NButton>
        <NButton
          size="small"
          quaternary
          title="下移"
          aria-label="下移"
          :disabled="!canReorder"
          @click="emit('moveDown')"
        >
          <template #icon>
            <NIcon>
              <Icon icon="icon-park-outline:down" />
            </NIcon>
          </template>
        </NButton>
        <NButton
          size="small"
          quaternary
          title="复制本节"
          aria-label="复制本节"
          :disabled="!canCopySec"
          @click="emit('copy')"
        >
          <template #icon>
            <NIcon>
              <Icon icon="icon-park-outline:copy" />
            </NIcon>
          </template>
        </NButton>
        <NButton
          size="small"
          quaternary
          type="error"
          title="删除本节"
          aria-label="删除本节"
          :disabled="!canRemoveSec"
          @click="emit('remove')"
        >
          <template #icon>
            <NIcon>
              <Icon icon="icon-park-outline:delete" />
            </NIcon>
          </template>
        </NButton>
      </NSpace>
      <NFlex align="center" :size="8">
        <span class="text-13px opacity-70">显示在简历</span>
        <NSwitch v-model:value="sectionVisible" :disabled="!canWrite" size="small" />
      </NFlex>
    </NFlex>

    <NForm label-placement="top" size="small">
      <NFlex :size="8" align="flex-end" style="width: 100%">
        <IconSelect
          v-model:value="section.icon"
          prefix="lucide"
          size="small"
          compact
          :disabled="!canWrite"
          placeholder="图标"
        />
        <NFormItem label="类型" :show-feedback="false" class="section-panel__row-item">
          <NSelect
            v-model:value="section.type"
            :disabled="!canWrite"
            :options="SECTION_TYPE_OPTIONS"
            @update:value="onTypeChange"
          />
        </NFormItem>
        <NFormItem label="标题" :show-feedback="false" class="section-panel__row-item">
          <NInput v-model:value="section.title" :disabled="!canWrite" />
        </NFormItem>
      </NFlex>

      <template v-if="section.type === 'education'">
        <div class="section-panel__heading">教育元数据</div>
        <NFormItem label="学校">
          <NInput v-model:value="section.meta_json.school" :disabled="!canWrite" />
        </NFormItem>
        <NFormItem label="专业">
          <NInput v-model:value="section.meta_json.major" :disabled="!canWrite" />
        </NFormItem>
        <NFormItem label="时间">
          <PeriodRangePicker
            :value="section.meta_json.period"
            :disabled="!canWrite"
            @update:value="(v) => (section.meta_json.period = v)"
          />
        </NFormItem>
      </template>

      <!-- 教育 / 技能 / 评价 / 自定义：节级正文 -->
      <template v-if="showSectionBody">
        <div class="section-panel__heading">{{ bodyDividerLabel }}</div>
        <ResumeBodyEditor v-model:value="section.body_html" :disabled="!canWrite" :height="220" />
      </template>

      <!-- 工作 / 项目：折叠条目 -->
      <template v-if="isEntryType">
        <div class="section-panel__heading">
          {{ section.type === 'projects' ? '项目条目' : '工作条目' }}
        </div>

        <NEmpty
          v-if="!(section.entries_json && section.entries_json.length)"
          description="暂无条目，点击下方添加"
          size="small"
          class="mb-3"
        />

        <NCollapse
          v-else
          v-model:expanded-names="expandedNames"
          class="section-panel__entries"
        >
          <NCollapseItem
            v-for="(entry, eidx) in section.entries_json"
            :key="eidx"
            :name="entryName(Number(eidx))"
          >
            <template #header>
              <NText class="section-panel__entry-title truncate">
                {{ entrySummary(entry, Number(eidx)) }}
              </NText>
            </template>
            <template #header-extra>
              <NFlex
                v-if="canWrite"
                :size="0"
                align="center"
                class="section-panel__entry-actions"
                @click.stop
              >
                <NButton
                  quaternary
                  size="tiny"
                  title="上移"
                  aria-label="上移"
                  :disabled="eidx === 0"
                  @click="moveEntry(Number(eidx), -1)"
                >
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:up" />
                    </NIcon>
                  </template>
                </NButton>
                <NButton
                  quaternary
                  size="tiny"
                  title="下移"
                  aria-label="下移"
                  :disabled="eidx === (section.entries_json?.length || 0) - 1"
                  @click="moveEntry(Number(eidx), 1)"
                >
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:down" />
                    </NIcon>
                  </template>
                </NButton>
                <NButton
                  quaternary
                  type="error"
                  size="tiny"
                  title="删除"
                  aria-label="删除"
                  @click="removeEntry(Number(eidx))"
                >
                  <template #icon>
                    <NIcon>
                      <Icon icon="icon-park-outline:delete" />
                    </NIcon>
                  </template>
                </NButton>
              </NFlex>
            </template>

            <NFormItem :label="entryTitleLabel">
              <NInput
                v-model:value="entry.title"
                :placeholder="entryTitleLabel"
                :disabled="!canWrite"
              />
            </NFormItem>
            <NFormItem label="角色">
              <NInput v-model:value="entry.role" placeholder="角色" :disabled="!canWrite" />
            </NFormItem>
            <NFormItem label="时间">
              <PeriodRangePicker
                :value="entry.period"
                :disabled="!canWrite"
                @update:value="(v) => (entry.period = v)"
              />
            </NFormItem>
            <NFormItem label="标签">
              <NInput
                placeholder="逗号分隔"
                :disabled="!canWrite"
                :value="tagsText(entry)"
                @update:value="(v: string) => setTags(entry, v)"
              />
            </NFormItem>
            <NFormItem label="说明">
              <ResumeBodyEditor
                v-model:value="entry.body_html"
                :disabled="!canWrite"
                :height="180"
              />
            </NFormItem>
          </NCollapseItem>
        </NCollapse>

        <NButton v-if="canWrite" dashed block class="mt-2" @click="addEntry"> 添加条目 </NButton>
      </template>
    </NForm>
  </div>
</template>

<style scoped>
.section-panel__row-item {
  flex: 1 1 0;
  min-width: 0;
  margin-bottom: 0;
}

.section-panel__heading {
  margin: 16px 0 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color, #efeff5);
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-1, #333);
  line-height: 1.4;
}

.section-panel__entries {
  width: 100%;
  max-width: 100%;
  min-width: 0;
}

.section-panel__entries :deep(.n-collapse-item__header) {
  padding: 8px 0;
  min-width: 0;
}

.section-panel__entries :deep(.n-collapse-item__header-main) {
  min-width: 0;
  overflow: hidden;
}

.section-panel__entry-title {
  display: block;
  max-width: 100%;
  font-size: 13px;
  line-height: 1.4;
}

.section-panel__entry-actions {
  flex: none;
  margin-left: 4px;
}

:deep(.n-form-item) {
  grid-template-rows: auto auto auto !important;
}
</style>
