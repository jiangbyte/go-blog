<!-- Author: go-blog — 服务条款 / 隐私政策 Markdown -->

<script setup lang="ts">
import { MdEditor } from '@/components/editor'
import { onMounted, reactive } from 'vue'
import ConfigSectionLayout from './ConfigSectionLayout.vue'
import { loadByCategory, saveByKeys } from '../composables/useConfigForm'

const CATEGORY = 'LEGAL'

const state = reactive({
  loading: false,
  saving: false,
  terms: '',
  privacy: '',
  snapshot: '',
})

onMounted(() => {
  void reload()
})

function snapshotOf() {
  return JSON.stringify({
    terms: state.terms,
    privacy: state.privacy,
  })
}

async function reload() {
  state.loading = true
  try {
    // 1. 按 LEGAL 分类拉 Markdown 正文
    const map = await loadByCategory(CATEGORY)
    state.terms = map.TERMS_OF_SERVICE || ''
    state.privacy = map.PRIVACY_POLICY || ''
    state.snapshot = snapshotOf()
  } finally {
    state.loading = false
  }
}

function reset() {
  if (!state.snapshot) return
  Object.assign(state, JSON.parse(state.snapshot))
}

async function save() {
  state.saving = true
  try {
    // 2. 批量写回 Markdown（内置键）
    await saveByKeys([
      {
        config_key: 'TERMS_OF_SERVICE',
        config_value: state.terms,
        category: CATEGORY,
        value_type: 'STRING',
        label: '服务条款',
        is_builtin: true,
      },
      {
        config_key: 'PRIVACY_POLICY',
        config_value: state.privacy,
        category: CATEGORY,
        value_type: 'STRING',
        label: '隐私政策',
        is_builtin: true,
      },
    ])
    window.$message.success('保存成功')
    state.snapshot = snapshotOf()
  } finally {
    state.saving = false
  }
}
</script>

<template>
  <NSpin :show="state.loading">
    <ConfigSectionLayout
      description="编辑前台与管理端展示的服务条款、隐私政策（Markdown）。公开接口 GET /portal/sys/site/legal 读取。"
      :saving="state.saving"
      @save="save"
      @reset="reset"
    >
      <NForm class="sys-config-form sys-config-form--fluid" label-placement="top">
        <NFormItem label="服务条款">
          <MdEditor v-model:value="state.terms" :height="360" />
        </NFormItem>
        <NFormItem label="隐私政策">
          <MdEditor v-model:value="state.privacy" :height="360" />
        </NFormItem>
      </NForm>
    </ConfigSectionLayout>
  </NSpin>
</template>
