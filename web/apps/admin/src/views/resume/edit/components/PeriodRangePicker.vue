<!-- Author: go-blog — 简历时间区间：月选择 + 至今 -->

<script setup lang="ts">
import { NCheckbox, NDatePicker } from 'naive-ui'
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    value?: string | null
    disabled?: boolean
    /** zh → 至今；en → Present */
    language?: string
  }>(),
  {
    value: '',
    disabled: false,
    language: 'zh',
  },
)

const emit = defineEmits<{
  'update:value': [value: string]
}>()

const presentLabel = computed(() => (props.language === 'en' ? 'Present' : '至今'))

const startTs = ref<number | null>(null)
const endTs = ref<number | null>(null)
const isPresent = ref(false)
/** 避免外部 value ↔ 内部状态互相回写死循环 */
let syncingFromProp = false

watch(
  () => props.value,
  (raw) => {
    syncingFromProp = true
    applyParsed(parsePeriod(String(raw || '')))
    syncingFromProp = false
  },
  { immediate: true },
)

function applyParsed(parsed: { start: number | null; end: number | null; present: boolean }) {
  startTs.value = parsed.start
  endTs.value = parsed.present ? null : parsed.end
  isPresent.value = parsed.present
}

/** 解析已有文案：`2022.09 - 2026.06` / `2026.03 - 至今`。 */
function parsePeriod(raw: string): {
  start: number | null
  end: number | null
  present: boolean
} {
  const text = raw.trim()
  if (!text) {
    return { start: null, end: null, present: false }
  }
  const presentRe =
    /^(\d{4})[./\-年](\d{1,2})月?\s*[-–—~～至到]+\s*(至今|现在|当前|present|now|current)\s*$/i
  const presentMatch = presentRe.exec(text)
  if (presentMatch) {
    return {
      start: toMonthTs(Number(presentMatch[1]), Number(presentMatch[2])),
      end: null,
      present: true,
    }
  }
  const rangeRe = /^(\d{4})[./\-年](\d{1,2})月?\s*[-–—~～至到]+\s*(\d{4})[./\-年](\d{1,2})月?\s*$/
  const rangeMatch = rangeRe.exec(text)
  if (rangeMatch) {
    return {
      start: toMonthTs(Number(rangeMatch[1]), Number(rangeMatch[2])),
      end: toMonthTs(Number(rangeMatch[3]), Number(rangeMatch[4])),
      present: false,
    }
  }
  // 单月：视为开始月
  const singleRe = /^(\d{4})[./\-年](\d{1,2})月?\s*$/
  const singleMatch = singleRe.exec(text)
  if (singleMatch) {
    return {
      start: toMonthTs(Number(singleMatch[1]), Number(singleMatch[2])),
      end: null,
      present: false,
    }
  }
  return { start: null, end: null, present: false }
}

function toMonthTs(year: number, month: number): number | null {
  if (!Number.isFinite(year) || !Number.isFinite(month) || month < 1 || month > 12) {
    return null
  }
  return new Date(year, month - 1, 1).getTime()
}

function formatYm(ts: number | null): string {
  if (ts == null) return ''
  const d = new Date(ts)
  return `${d.getFullYear()}.${String(d.getMonth() + 1).padStart(2, '0')}`
}

/** 写回统一文案，供纸面直接展示。 */
function emitFormatted() {
  if (syncingFromProp) return
  const start = formatYm(startTs.value)
  if (!start) {
    emit('update:value', '')
    return
  }
  if (isPresent.value) {
    emit('update:value', `${start} - ${presentLabel.value}`)
    return
  }
  const end = formatYm(endTs.value)
  if (!end) {
    emit('update:value', start)
    return
  }
  emit('update:value', `${start} - ${end}`)
}

function onStartUpdate(v: number | null) {
  startTs.value = v
  // 结束早于开始时纠正
  if (v != null && endTs.value != null && endTs.value < v) {
    endTs.value = v
  }
  emitFormatted()
}

function onEndUpdate(v: number | null) {
  endTs.value = v
  if (v != null) {
    isPresent.value = false
  }
  if (v != null && startTs.value != null && v < startTs.value) {
    startTs.value = v
  }
  emitFormatted()
}

function onPresentUpdate(checked: boolean) {
  isPresent.value = checked
  if (checked) {
    endTs.value = null
  }
  emitFormatted()
}

watch(presentLabel, () => {
  if (isPresent.value) emitFormatted()
})
</script>

<template>
  <div class="period-range">
    <NDatePicker
      class="period-range__picker"
      type="month"
      clearable
      :disabled="disabled"
      :value="startTs"
      :actions="['clear', 'confirm']"
      @update:value="onStartUpdate"
    />
    <span class="period-range__sep">-</span>
    <NDatePicker
      class="period-range__picker"
      type="month"
      clearable
      :disabled="disabled || isPresent"
      :value="endTs"
      :actions="['clear', 'confirm']"
      :placeholder="isPresent ? presentLabel : undefined"
      @update:value="onEndUpdate"
    />
    <NCheckbox
      class="period-range__present"
      :checked="isPresent"
      :disabled="disabled || startTs == null"
      @update:checked="onPresentUpdate"
    >
      {{ presentLabel }}
    </NCheckbox>
  </div>
</template>

<style scoped>
.period-range {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 6px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
}

.period-range__picker {
  flex: 1 1 0%;
  min-width: 0;
  max-width: 100%;
}

.period-range__picker :deep(.n-input),
.period-range__picker :deep(.n-date-picker) {
  width: 100%;
  min-width: 0;
  max-width: 100%;
}

.period-range__sep {
  flex: none;
  color: var(--text-color-3);
  line-height: 1;
}

.period-range__present {
  flex: none;
  white-space: nowrap;
}
</style>
