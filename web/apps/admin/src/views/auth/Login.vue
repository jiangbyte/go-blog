<!-- Author: go-blog -->

<script setup lang="ts">
import type { FormItemInst, FormItemRule } from 'naive-ui'
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores'
import { encryptPasswords } from '@/utils/security'
import AdminSliderCaptcha from '@/components/common/AdminSliderCaptcha.vue'
import AuthLayout from './AuthLayout.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const accountItemRef = ref<FormItemInst | null>(null)
const passwordItemRef = ref<FormItemInst | null>(null)
const loading = ref(false)
const captchaToken = ref('')
const captchaRef = ref<InstanceType<typeof AdminSliderCaptcha> | null>(null)

const form = reactive({
  username: '',
  password: '',
  remember: true,
})

const accountRule: FormItemRule = {
  trigger: ['input', 'blur'],
  validator() {
    if (!form.username.trim()) return new Error('请输入管理员账号')
    return true
  },
}

const passwordRule: FormItemRule = {
  trigger: ['input', 'blur'],
  validator() {
    if (!form.password) return new Error('请输入密码')
    return true
  },
}

async function handleSubmit() {
  try {
    await Promise.all([accountItemRef.value?.validate(), passwordItemRef.value?.validate()])
  } catch {
    return
  }

  if (!captchaToken.value) {
    window.$message?.warning?.('请先完成滑块验证')
    return
  }

  loading.value = true
  try {
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : undefined
    const encrypted = await encryptPasswords({ password: form.password })
    const cipher = encrypted.values.password || ''
    await authStore.login(form.username.trim(), cipher, redirect, form.remember, captchaToken.value)
    window.$message.success('登录成功')
  } catch (error: unknown) {
    captchaToken.value = ''
    captchaRef.value?.refresh()
    const msg = error instanceof Error ? error.message : '登录失败'
    window.$message.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthLayout title="欢迎登录" description="使用管理员账号进入控制台">
    <form class="auth-form" @submit.prevent="handleSubmit">
      <n-form-item ref="accountItemRef" :show-label="false" :rule="accountRule">
        <n-input v-model:value="form.username" size="large" clearable placeholder="管理员账号" />
      </n-form-item>
      <n-form-item ref="passwordItemRef" :show-label="false" :rule="passwordRule">
        <n-input
          v-model:value="form.password"
          size="large"
          type="password"
          show-password-on="click"
          placeholder="密码"
        />
      </n-form-item>

      <AdminSliderCaptcha ref="captchaRef" v-model:captcha-token="captchaToken" />

      <n-button type="primary" size="large" block attr-type="submit" :loading="loading">
        登录
      </n-button>

      <n-flex justify="center" style="margin-top: 12px">
        <n-button text type="primary" @click="router.push('/auth/forgot-password')">
          忘记密码？
        </n-button>
      </n-flex>
    </form>
  </AuthLayout>
</template>
