<!--
  Author: go-blog
  分类/标签多选：下拉选择已有 + 输入回车创建（Naive NSelect tag / on-create）
  v-model 为 uid[]；新建走 taxonomy create（后端 ensure）
-->
<script setup lang="ts">
import type { SelectOption } from 'naive-ui'
import type { TaxonomyKind } from '@/api/blog/site'
import { blogTaxonomyApi } from '@/api'
import { NSelect } from 'naive-ui'
import { ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    kind: TaxonomyKind
    value?: string[] | null
    options?: SelectOption[]
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    value: () => [],
    options: () => [],
    placeholder: '选择已有，或输入后回车创建',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:value': [string[]]
  'update:options': [SelectOption[]]
}>()

const localOptions = ref<SelectOption[]>([...(props.options || [])])
const creating = ref(false)
/** 输入创建中的临时 value（名称）→ 待替换为 uid */
const pendingNames = ref(new Set<string>())

watch(
  () => props.options,
  (list) => {
    const incoming = list || []
    const byValue = new Map(localOptions.value.map((o) => [String(o.value), o]))
    for (const o of incoming) {
      byValue.set(String(o.value), o)
    }
    localOptions.value = [...byValue.values()]
  },
  { deep: true },
)

function knownUid(v: string): boolean {
  return localOptions.value.some((o) => String(o.value) === v && !pendingNames.value.has(v))
}

function findByName(name: string): SelectOption | undefined {
  const n = name.trim().toLowerCase()
  return localOptions.value.find(
    (o) => String(o.label ?? '').trim().toLowerCase() === n && !String(o.value).startsWith('__new__:'),
  )
}

function upsertOption(opt: SelectOption) {
  const uid = String(opt.value)
  const idx = localOptions.value.findIndex((o) => String(o.value) === uid)
  if (idx >= 0) {
    localOptions.value[idx] = opt
  } else {
    localOptions.value = [...localOptions.value, opt]
  }
  emit('update:options', [...localOptions.value])
}

/**
 * tag 模式新建：先插入临时选项（value=名称），回车后由 onUpdateValue ensure 成真实 uid。
 * Naive on-create 必须同步返回 SelectOption。
 */
function onCreate(label: string): SelectOption {
  const name = String(label || '').trim()
  const existing = findByName(name)
  if (existing) return existing
  pendingNames.value.add(name)
  const opt: SelectOption = { label: name, value: name }
  // 不立刻 upsert 进 options 列表由 Select 内部持有；ensure 后再写入
  return opt
}

/** 把选中项中的「新建名称」换成 uid，已有 uid 原样保留。 */
async function onUpdateValue(vals: Array<string | number> | null) {
  const list = (vals || []).map((v) => String(v).trim()).filter(Boolean)
  const next: string[] = []
  const seen = new Set<string>()

  for (const v of list) {
    if (knownUid(v)) {
      if (!seen.has(v)) {
        seen.add(v)
        next.push(v)
      }
      continue
    }

    // 临时 tag 名称，或用户直接输入的同名
    const name = v
    const cached = findByName(name)
    if (cached && knownUid(String(cached.value))) {
      const uid = String(cached.value)
      if (!seen.has(uid)) {
        seen.add(uid)
        next.push(uid)
      }
      pendingNames.value.delete(name)
      continue
    }

    creating.value = true
    try {
      const res = await blogTaxonomyApi.create({ kind: props.kind, name })
      const uid = String(res.data?.uid || '')
      const label = String(res.data?.name || name)
      if (!uid) continue
      upsertOption({ label, value: uid })
      pendingNames.value.delete(name)
      if (!seen.has(uid)) {
        seen.add(uid)
        next.push(uid)
      }
    } catch (e: any) {
      window.$message?.error(e?.message || `创建「${name}」失败`)
      pendingNames.value.delete(name)
    } finally {
      creating.value = false
    }
  }

  emit('update:value', next)
}
</script>

<template>
  <NSelect
    :value="value || []"
    :options="localOptions"
    multiple
    filterable
    tag
    clearable
    :disabled="disabled"
    :loading="creating"
    :placeholder="placeholder"
    :on-create="onCreate"
    @update:value="onUpdateValue"
  />
</template>
