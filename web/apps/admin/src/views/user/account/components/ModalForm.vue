<!-- Author: go-blog — 用户表单 -->

<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { accountApi } from '@/api'
import { createRequiredRule } from '@/utils'
import { encryptPasswords } from '@/utils/security'
import { wireBool } from '@/utils/wire'
import { computed, reactive, ref } from 'vue'

const emit = defineEmits<{
  saved: []
}>()

const formRef = ref<FormInst | null>(null)
const defaultFormData = {
  username: '',
  nickname: '',
  password: '',
  user_type: 'ADMIN',
  phone: '',
  email: '',
  gender: 'UNKNOWN',
  signature: '',
  enabled: true,
}

const state = reactive({
  showModal: false,
  loading: false,
  submitLoading: false,
  dataId: null as string | null,
  formModel: { ...defaultFormData },
})

const modalTitle = computed(() => (state.dataId ? '编辑用户' : '新增用户'))

const rules = computed<FormRules>(() => ({
  username: state.dataId ? [] : [createRequiredRule('账号', 'input')],
  password: state.dataId
    ? []
    : [
        createRequiredRule('密码', 'input'),
        { min: 6, message: '至少 6 位', trigger: ['blur', 'input'] },
      ],
}))

/**
 * 打开表单：编辑时优先 detail，失败则用行数据回填。
 */
async function openModal(id?: string, row?: any) {
  // 1. 重置并打开
  state.dataId = id ?? null
  state.formModel = { ...defaultFormData }
  state.showModal = true
  // 2. 编辑态拉详情或用行数据
  if (id) {
    state.loading = true
    try {
      try {
        const res = await accountApi.detail({ user_id: id })
        fillFromRow(res.data ?? row)
      } catch {
        fillFromRow(row)
      }
    } finally {
      state.loading = false
    }
  }
}

function fillFromRow(row: any) {
  if (!row) return
  state.formModel = {
    username: String(row.username ?? ''),
    nickname: String(row.nickname ?? ''),
    password: '',
    user_type: String(row.user_type ?? 'ADMIN'),
    phone: String(row.phone ?? ''),
    email: String(row.email ?? ''),
    gender: String(row.gender ?? 'UNKNOWN'),
    signature: String(row.signature ?? ''),
    enabled: wireBool(row.enabled, true),
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
    // 1. 密码按需 RSA 加密
    let password_cipher: string | undefined
    if (state.formModel.password.trim()) {
      const enc = await encryptPasswords({ password: state.formModel.password })
      password_cipher = enc.values.password || undefined
    }
    // 2. 创建或更新
    if (state.dataId) {
      await accountApi.update({
        user_id: state.dataId,
        nickname: state.formModel.nickname || state.formModel.username,
        phone: state.formModel.phone || null,
        email: state.formModel.email || null,
        gender: state.formModel.gender,
        signature: state.formModel.signature || null,
        enabled: state.formModel.enabled ? 'true' : 'false',
        password_cipher,
      })
      window.$message.success('更新成功')
    } else {
      if (!password_cipher) {
        window.$message.error('请输入密码')
        return
      }
      await accountApi.create({
        username: state.formModel.username.trim(),
        nickname: state.formModel.nickname,
        password_cipher,
        user_type: state.formModel.user_type,
        phone: state.formModel.phone || null,
        email: state.formModel.email || null,
        gender: state.formModel.gender,
        signature: state.formModel.signature || null,
      })
      window.$message.success('创建成功')
    }
    closeModal()
    emit('saved')
  } finally {
    state.submitLoading = false
  }
}

defineExpose({ openModal })
</script>

<template>
  <FormContainer
    v-model:show="state.showModal"
    :title="modalTitle"
    :width="560"
    :mask-closable="false"
  >
    <NSpin :show="state.loading">
      <NForm
        ref="formRef"
        :model="state.formModel"
        :rules="rules"
        label-placement="left"
        label-width="90"
        :disabled="state.loading || state.submitLoading"
      >
        <NFormItem label="账号" path="username">
          <NInput v-model:value="state.formModel.username" :disabled="Boolean(state.dataId)" />
        </NFormItem>
        <NFormItem label="密码" path="password">
          <NInput
            v-model:value="state.formModel.password"
            type="password"
            show-password-on="click"
            :placeholder="state.dataId ? '留空则不修改' : undefined"
          />
        </NFormItem>
        <NFormItem v-if="!state.dataId" label="类型" path="user_type">
          <NSelect
            v-model:value="state.formModel.user_type"
            :options="[
              { label: '管理员', value: 'ADMIN' },
              { label: '前台', value: 'PORTAL' },
            ]"
          />
        </NFormItem>
        <NFormItem label="昵称" path="nickname">
          <NInput v-model:value="state.formModel.nickname" />
        </NFormItem>
        <NFormItem label="状态" path="enabled">
          <NSwitch v-model:value="state.formModel.enabled" />
        </NFormItem>
        <NFormItem label="性别" path="gender">
          <NSelect
            v-model:value="state.formModel.gender"
            :options="[
              { label: '未知', value: 'UNKNOWN' },
              { label: '男', value: 'MALE' },
              { label: '女', value: 'FEMALE' },
            ]"
          />
        </NFormItem>
        <NFormItem label="手机" path="phone">
          <NInput v-model:value="state.formModel.phone" />
        </NFormItem>
        <NFormItem label="邮箱" path="email">
          <NInput v-model:value="state.formModel.email" />
        </NFormItem>
        <NFormItem label="签名" path="signature">
          <NInput
            v-model:value="state.formModel.signature"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 4 }"
          />
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
