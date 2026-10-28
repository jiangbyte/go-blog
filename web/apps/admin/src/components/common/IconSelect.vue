<!-- Author: go-blog -->

<script setup lang="ts">
import { icons as iconParkOutline } from '@iconify-json/icon-park-outline'
import { icons as lucide } from '@iconify-json/lucide'
import { resumeApi } from '@/api'
import { resolveResumeIcon } from '@jiangbyte/shared'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const ICON_COLLECTIONS: Record<string, { icons: Record<string, unknown> }> = {
  'icon-park-outline': iconParkOutline,
  lucide,
}

const props = withDefaults(
  defineProps<{
    value?: string | null
    placeholder?: string
    clearable?: boolean
    disabled?: boolean
    size?: 'tiny' | 'small' | 'medium' | 'large'
    prefix?: string
    columns?: number
    pageSize?: number
    modalWidth?: string | number
    /** 仅预览+选图标按钮，适合表单行内嵌 */
    compact?: boolean
  }>(),
  {
    value: '',
    placeholder: '请选择图标',
    clearable: true,
    disabled: false,
    size: 'medium',
    prefix: 'icon-park-outline',
    columns: 8,
    pageSize: 64,
    modalWidth: 720,
    compact: false,
  },
)

const emit = defineEmits<{
  'update:value': [value: string]
  change: [value: string]
}>()

/** 后端图标名（lucide）；null 表示未加载或回退本地 */
const backendNames = ref<string[] | null>(null)

const localIconNames = computed(() => {
  const collection = ICON_COLLECTIONS[props.prefix] || iconParkOutline
  return Object.keys(collection.icons).sort((a, b) => a.localeCompare(b))
})

const iconNames = computed(() => {
  if (props.prefix === 'lucide' && backendNames.value?.length) {
    return backendNames.value
  }
  return localIconNames.value
})

/** 模块级缓存：lucide 图标名只拉一次 */
let lucideNamesCache: string[] | null = null
let lucideNamesPromise: Promise<string[] | null> | null = null

/** 拉取与后端渲染同源的 lucide 名称列表。 */
async function fetchBackendLucideNames(): Promise<string[] | null> {
  if (lucideNamesCache) return lucideNamesCache
  if (lucideNamesPromise) return lucideNamesPromise
  lucideNamesPromise = (async () => {
    try {
      // 1. 分页拉全量（每页 ≤100）
      const names: string[] = []
      let current = 1
      let total = Number.POSITIVE_INFINITY
      while (names.length < total && current <= 40) {
        const res = await resumeApi.pageIcons({
          prefix: 'lucide',
          current,
          size: 100,
        })
        const records = res.data?.records || []
        total = Number(res.data?.total || 0)
        for (const row of records) {
          if (row?.name) names.push(String(row.name))
        }
        if (!records.length) break
        current += 1
      }
      // 2. 有数据才缓存
      lucideNamesCache = names.length ? names : null
      return lucideNamesCache
    } catch {
      return null
    } finally {
      lucideNamesPromise = null
    }
  })()
  return lucideNamesPromise
}

async function loadBackendLucideNames() {
  if (props.prefix !== 'lucide') return
  backendNames.value = await fetchBackendLucideNames()
}

const showPanel = ref(false)
const searchKey = ref('')
const deferredSearchKey = ref('')
const page = ref(1)
let searchTimer: number | null = null
const currentIcon = computed(() => normalizeIconValue(props.value))
const columnCount = computed(() => Math.max(1, Math.floor(Number(props.columns) || 1)))
const normalizedPageSize = computed(() => Math.max(1, Math.floor(Number(props.pageSize) || 1)))
const modalStyle = computed(() => ({
  width: toCssSize(props.modalWidth),
  '--icon-select-columns': String(columnCount.value),
}))
const filteredIcons = computed(() => {
  const keyword = deferredSearchKey.value.trim().toLowerCase()
  if (!keyword) {
    return iconNames.value
  }

  return iconNames.value.filter((name) => {
    const fullName = toFullIconName(name)
    return name.includes(keyword) || fullName.includes(keyword)
  })
})
const pageCount = computed(() =>
  Math.max(1, Math.ceil(filteredIcons.value.length / normalizedPageSize.value)),
)
const pagedIcons = computed(() => {
  const currentPage = Math.min(page.value, pageCount.value)
  const start = (currentPage - 1) * normalizedPageSize.value
  return filteredIcons.value.slice(start, start + normalizedPageSize.value)
})
const resultText = computed(() => {
  const total = filteredIcons.value.length
  if (!total) {
    return '0 个图标'
  }
  const start = (Math.min(page.value, pageCount.value) - 1) * normalizedPageSize.value + 1
  const end = Math.min(start + pagedIcons.value.length - 1, total)
  return `${start}-${end} / ${total}`
})

function toFullIconName(name: string) {
  return name.includes(':') ? name : `${props.prefix}:${name}`
}

function normalizeIconValue(value?: string | null) {
  const rawValue = String(value ?? '').trim()
  if (!rawValue) {
    return ''
  }
  // lucide 与后端纸面规则一致
  if (props.prefix === 'lucide') {
    return resolveResumeIcon(rawValue)
  }
  return toFullIconName(rawValue)
}

function handleUpdateValue(value: string | null) {
  const icon = normalizeIconValue(value)
  emit('update:value', icon)
  emit('change', icon)
}

function selectIcon(name: string) {
  handleUpdateValue(toFullIconName(name))
  showPanel.value = false
}

function clearIcon() {
  if (props.disabled || !props.clearable) {
    return
  }
  handleUpdateValue('')
}

/** 弹窗内清除并关闭，避免清空后仍停留在选择态 */
function clearIconAndClose() {
  clearIcon()
  showPanel.value = false
}

function openPanel() {
  if (!props.disabled) {
    showPanel.value = true
  }
}

function handleUpdatePage(value: number) {
  page.value = value
}

function toCssSize(value: string | number) {
  return typeof value === 'number' ? `${value}px` : value
}

watch(searchKey, (value) => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
  searchTimer = window.setTimeout(() => {
    deferredSearchKey.value = value
    page.value = 1
  }, 120)
})

watch(showPanel, (value) => {
  if (value) {
    deferredSearchKey.value = searchKey.value
    page.value = 1
    void loadBackendLucideNames()
  }
})

watch(
  () => props.prefix,
  () => {
    backendNames.value = null
  },
)

onMounted(() => {
  if (props.prefix === 'lucide') {
    void loadBackendLucideNames()
  }
})

onBeforeUnmount(() => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
})
</script>

<template>
  <div class="icon-select" :class="{ 'icon-select--compact': compact }">
    <template v-if="compact">
      <button
        type="button"
        class="icon-select__preview icon-select__preview--btn"
        :disabled="disabled"
        :title="currentIcon || placeholder"
        aria-label="选择图标"
        @click="openPanel"
      >
        <NovaIcon v-if="currentIcon" :icon="currentIcon" :size="18" />
        <NovaIcon v-else icon="icon-park-outline:all-application" :size="16" />
      </button>
      <button
        v-if="clearable && currentIcon && !disabled"
        type="button"
        class="icon-select__clear"
        title="清除图标"
        aria-label="清除图标"
        @click.stop="clearIcon"
      >
        <NovaIcon icon="icon-park-outline:close" :size="12" />
      </button>
    </template>
    <template v-else>
      <span class="icon-select__preview">
        <NovaIcon v-if="currentIcon" :icon="currentIcon" :size="18" />
      </span>
      <NInput
        class="icon-select__input"
        :value="currentIcon"
        :placeholder="placeholder"
        :disabled="disabled"
        :size="size"
        :clearable="clearable"
        @update:value="handleUpdateValue"
        @clear="clearIcon"
      />
      <NButton
        class="icon-select__button"
        :size="size"
        :disabled="disabled"
        title="选择图标"
        aria-label="选择图标"
        @click="openPanel"
      >
        <template #icon>
          <NovaIcon icon="icon-park-outline:all-application" />
        </template>
      </NButton>
    </template>
  </div>

  <NModal
    v-model:show="showPanel"
    preset="card"
    draggable
    :mask-closable="true"
    :title="'选择图标'"
    :style="modalStyle"
    :segmented="{ content: true, footer: 'soft' }"
  >
    <div class="icon-select-panel">
      <div class="icon-select-panel__header">
        <NInput v-model:value="searchKey" size="small" placeholder="搜索图标名称" clearable />
        <div class="icon-select-panel__meta">
          <NText depth="3">
            {{ resultText }}
          </NText>
          <NPagination
            v-if="pageCount > 1"
            :page="page"
            :page-count="pageCount"
            size="small"
            simple
            @update:page="handleUpdatePage"
          />
        </div>
      </div>
      <div class="icon-select-panel__scroll">
        <div class="icon-select-panel__grid">
          <button
            v-for="name in pagedIcons"
            :key="name"
            class="icon-select-panel__item"
            :class="{ 'icon-select-panel__item--active': currentIcon === toFullIconName(name) }"
            type="button"
            :title="toFullIconName(name)"
            @click="selectIcon(name)"
          >
            <NovaIcon :icon="toFullIconName(name)" :size="20" />
            <span>{{ name }}</span>
          </button>
        </div>
        <NEmpty v-if="!filteredIcons.length" size="small" description="未找到图标" />
      </div>
    </div>
    <template v-if="clearable && currentIcon && !disabled" #footer>
      <NFlex justify="end">
        <NButton @click="clearIconAndClose">清除</NButton>
      </NFlex>
    </template>
  </NModal>
</template>

<style scoped>
.icon-select {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-width: 0;
}

.icon-select--compact {
  width: auto;
  flex: none;
}

.icon-select__clear {
  display: inline-flex;
  flex: none;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  padding: 0;
  border: 1px solid var(--border-color);
  border-radius: 50%;
  background: var(--body-color);
  color: var(--text-color-3);
  cursor: pointer;
}

.icon-select__clear:hover {
  border-color: var(--error-color);
  color: var(--error-color);
}

.icon-select__preview {
  display: inline-flex;
  flex: 0 0 32px;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-color-2);
}

.icon-select__preview--btn {
  flex: none;
  padding: 0;
  background: var(--body-color);
  cursor: pointer;
}

.icon-select__preview--btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.icon-select__preview--btn:not(:disabled):hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.icon-select__input {
  flex: 1;
  min-width: 0;
}

.icon-select__button {
  flex: 0 0 auto;
}

.icon-select-panel {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  gap: 10px;
  min-width: 280px;
  min-height: 0;
  max-height: min(640px, calc(100vh - 180px));
  max-width: 100%;
  overflow-x: hidden;
}

.icon-select-panel__header {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.icon-select-panel__grid {
  display: grid;
  grid-template-columns: repeat(var(--icon-select-columns), minmax(0, 1fr));
  gap: 6px;
  min-width: 0;
  padding-right: 4px;
}

.icon-select-panel__scroll {
  min-height: 0;
  max-height: min(520px, calc(100vh - 300px));
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior: contain;
}

.icon-select-panel__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-width: 0;
}

.icon-select-panel__item {
  display: grid;
  grid-template-rows: 24px 16px;
  gap: 4px;
  align-items: center;
  justify-items: center;
  min-width: 0;
  height: 54px;
  padding: 5px 4px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--body-color);
  color: var(--text-color-2);
  cursor: pointer;
}

.icon-select-panel__item:hover,
.icon-select-panel__item--active {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.icon-select-panel__item span {
  max-width: 100%;
  overflow: hidden;
  font-size: 11px;
  line-height: 16px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:global(.n-modal .icon-select-panel) {
  max-height: calc(100vh - 180px);
}
</style>
