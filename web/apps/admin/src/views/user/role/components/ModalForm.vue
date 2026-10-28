<!-- Author: go-blog — 角色表单 -->

<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { roleApi } from '@/api'
import { createRequiredRule } from '@/utils'
import { computed, reactive, ref } from 'vue'
import { wireBool } from '@/utils/wire'

const emit = defineEmits<{
  saved: []
}>()

const formRef = ref<FormInst | null>(null)
const defaultFormData = {
  code: '',
  name: '',
  enabled: true,
}
const state = reactive({
  showModal: false,
  loading: false,
  submitLoading: false,
  dataId: null as string | null,
  formModel: { ...defaultFormData },
})

const modalTitle = computed(() => (state.dataId ? '编辑角色' : '新增角色'))

const rules = computed<FormRules>(() => ({
  code: createRequiredRule('角色编码', 'input'),
  name: createRequiredRule('角色名称', 'input'),
}))

/**
 * 打开表单：无 id 为新增；有 id 时用列表行回填（mypage 无 roles/detail）。
 */
async function openModal(id?: string, row?: any) {
  // 1. 重置并打开
  state.dataId = id ?? null
  state.formModel = { ...defaultFormData }
  state.showModal = true
  // 2. 编辑态用行数据回填
  if (id && row) {
    state.formModel = {
      code: String(row.code ?? ''),
      name: String(row.name ?? ''),
      enabled: wireBool(row.enabled),
    }
  }
}

function closeModal() {
  state.showModal = false
  state.submitLoading = false
}

async function submitForm() {
  await formRef.value?.validate()
  state.submitLoading = true
  try {
    // 1. 组装 mypage 载荷（enabled 用字符串）
    const payload = {
      code: state.formModel.code.trim(),
      name: state.formModel.name.trim(),
      enabled: state.formModel.enabled ? 'true' : 'false',
    }
    // 2. 创建或更新
    if (state.dataId) {
      await roleApi.update({
        role_id: state.dataId,
        name: payload.name,
        enabled: payload.enabled,
      })
      window.$message.success('更新成功')
    } else {
      await roleApi.create(payload)
      window.$message.success('创建成功')
    }
    closeModal()
    emit('saved')
  } finally {
    state.submitLoading = false
  }
}

defineExpose({
  openModal,
})
</script>

<template>
  <FormContainer
    v-model:show="state.showModal"
    :title="modalTitle"
    :width="520"
    :mask-closable="false"
  >
    <NSpin :show="state.loading">
      <NForm
        ref="formRef"
        :model="state.formModel"
        :rules="rules"
        label-placement="left"
        label-width="110"
        :disabled="state.loading || state.submitLoading"
      >
        <NFormItem label="角色名称" path="name">
          <NInput v-model:value="state.formModel.name" />
        </NFormItem>
        <NFormItem label="角色编码" path="code">
          <NInput v-model:value="state.formModel.code" :disabled="Boolean(state.dataId)" />
        </NFormItem>
        <NFormItem label="状态" path="enabled">
          <NSwitch v-model:value="state.formModel.enabled" />
        </NFormItem>
      </NForm>
    </NSpin>

    <template #action>
      <NSpace justify="end" align="center">
        <NButton @click="closeModal"> 取消 </NButton>
        <NButton type="primary" :loading="state.submitLoading" @click="submitForm"> 确认 </NButton>
      </NSpace>
    </template>
  </FormContainer>
</template>
