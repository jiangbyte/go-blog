<!-- Author: go-blog — 旧密码改密 -->

<script setup lang="ts">
import { authApi } from '@/api'
import { encryptPasswords } from '@/utils/security'
import { reactive } from 'vue'
import '../profile.css'

const state = reactive({
  savingPassword: false,
  passwordForm: {
    old_password: '',
    new_password: '',
    confirm_password: '',
  },
})

async function savePassword() {
  if (state.passwordForm.new_password !== state.passwordForm.confirm_password) {
    window.$message.warning('两次输入的新密码不一致')
    return
  }
  if (!state.passwordForm.old_password) {
    window.$message.warning('请输入旧密码')
    return
  }
  if (!state.passwordForm.new_password) {
    window.$message.warning('请输入新密码')
    return
  }
  state.savingPassword = true
  try {
    const encrypted = await encryptPasswords({
      old_password: state.passwordForm.old_password,
      new_password: state.passwordForm.new_password,
    })
    await authApi.updatePassword({
      old_password: encrypted.values.old_password || undefined,
      new_password: encrypted.values.new_password || '',
    })
    state.passwordForm.old_password = ''
    state.passwordForm.new_password = ''
    state.passwordForm.confirm_password = ''
    window.$message.success('密码已更新')
  } finally {
    state.savingPassword = false
  }
}
</script>

<template>
  <NForm class="profile-form profile-form--narrow w-full min-w-0" label-placement="top">
    <NFormItem label="旧密码">
      <NInput
        v-model:value="state.passwordForm.old_password"
        type="password"
        show-password-on="click"
        placeholder="请输入旧密码"
      />
    </NFormItem>
    <NFormItem label="新密码">
      <NInput
        v-model:value="state.passwordForm.new_password"
        type="password"
        show-password-on="click"
        placeholder="请输入新密码"
      />
    </NFormItem>
    <NFormItem label="确认密码">
      <NInput
        v-model:value="state.passwordForm.confirm_password"
        type="password"
        show-password-on="click"
        placeholder="再次输入新密码"
      />
    </NFormItem>
    <NFormItem :show-label="false">
      <NButton type="primary" :loading="state.savingPassword" @click="savePassword">
        更新密码
      </NButton>
    </NFormItem>
  </NForm>
</template>
