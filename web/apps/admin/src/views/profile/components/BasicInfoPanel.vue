<!-- Author: go-blog — 公开资料（昵称 / 性别 / 签名） -->

<script setup lang="ts">
import { authApi } from '@/api'
import { useAuthStore } from '@/stores'
import { onMounted, reactive } from 'vue'
import '../profile.css'

const authStore = useAuthStore()

const state = reactive({
  loading: false,
  savingProfile: false,
  me: null as any,
  profileForm: {
    nickname: '',
    gender: 'UNKNOWN',
    signature: '',
  },
})

onMounted(async () => {
  await refresh()
})

async function refresh() {
  state.loading = true
  try {
    const data = await authStore.refreshUserInfo()
    state.me = data
    syncForms(data)
  } finally {
    state.loading = false
  }
}

function syncForms(data: any) {
  state.profileForm.nickname = data?.nickname ?? ''
  state.profileForm.gender = data?.gender || 'UNKNOWN'
  state.profileForm.signature = data?.signature ?? ''
}

async function saveProfile() {
  state.savingProfile = true
  try {
    await authApi.updateProfile({
      nickname: state.profileForm.nickname || null,
      gender: state.profileForm.gender || 'UNKNOWN',
      signature: state.profileForm.signature || null,
    })
    await refresh()
    window.$message.success('保存成功')
  } finally {
    state.savingProfile = false
  }
}

defineExpose({ refresh })
</script>

<template>
  <NSpin :show="state.loading">
    <NForm class="profile-form profile-form--narrow w-full min-w-0" label-placement="top">
      <NFormItem label="账号">
        <NInput :value="state.me?.account || state.me?.username" disabled />
        <template #feedback>
          <span class="profile__hint">登录账号不可修改。</span>
        </template>
      </NFormItem>
      <NFormItem label="昵称">
        <NInput v-model:value="state.profileForm.nickname" />
      </NFormItem>
      <NFormItem label="性别">
        <NSelect
          v-model:value="state.profileForm.gender"
          :options="[
            { label: '未知', value: 'UNKNOWN' },
            { label: '男', value: 'MALE' },
            { label: '女', value: 'FEMALE' },
          ]"
        />
      </NFormItem>
      <NFormItem label="个性签名">
        <NInput
          v-model:value="state.profileForm.signature"
          type="textarea"
          :rows="3"
          placeholder="一句话介绍自己"
        />
      </NFormItem>
      <NFormItem :show-label="false">
        <NButton type="primary" :loading="state.savingProfile" @click="saveProfile"> 保存 </NButton>
      </NFormItem>
    </NForm>
  </NSpin>
</template>
