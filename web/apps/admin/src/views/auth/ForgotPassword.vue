<!-- Author: go-blog -->

<script setup lang="ts">
import type { FormItemInst, FormItemRule } from 'naive-ui'
import { authApi } from '@/api'
import AdminSliderCaptcha from '@/components/common/AdminSliderCaptcha.vue'
import { isValidEmail } from '@/utils'
import { encryptPasswords } from '@/utils/security'
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AuthLayout from './AuthLayout.vue'

const route = useRoute()
const router = useRouter()
const emailItemRef = ref<FormItemInst | null>(null)
const passwordItemRef = ref<FormItemInst | null>(null)
const confirmItemRef = ref<FormItemInst | null>(null)
const captchaRef = ref<InstanceType<typeof AdminSliderCaptcha> | null>(null)
const captchaToken = ref('')
const loading = ref(false)
const debugUrl = ref('')

const form = reactive({
  email: '',
  token: typeof route.query.token === 'string' ? route.query.token : '',
  password: '',
  confirmPassword: '',
})

const isResetMode = computed(() => Boolean(form.token))

const emailRule: FormItemRule = {
  trigger: ['input', 'blur'],
  validator() {
    const text = form.email.trim()
    if (!text) return new Error('请输入登录邮箱')
    if (!isValidEmail(text)) return new Error('请输入有效邮箱')
    return true
  },
}

const passwordRule: FormItemRule = {
  trigger: ['input', 'blur'],
  validator() {
    if (!form.password) return new Error('请输入新密码')
    if (form.password.length < 6) return new Error('新密码至少 6 位')
    return true
  },
}

const confirmPasswordRule: FormItemRule = {
  trigger: ['input', 'blur'],
  validator() {
    if (!form.confirmPassword) return new Error('请确认密码')
    if (form.confirmPassword !== form.password) return new Error('两次输入的密码不一致')
    return true
  },
}

async function validateFields(items: Array<FormItemInst | null>) {
  try {
    await Promise.all(items.map((item) => item?.validate()))
    return true
  } catch {
    return false
  }
}

async function sendLink() {
  const ok = await validateFields([emailItemRef.value])
  if (!ok) return
  if (!captchaToken.value) {
    window.$message?.warning?.('请先完成滑块验证')
    return
  }
  loading.value = true
  try {
    const res = await authApi.forgotPassword({
      email: form.email.trim(),
      captcha_token: captchaToken.value,
      client_type: 'ADMIN',
    })
    window.$message.success(res.data.message || '若邮箱存在，重置链接已发送')
    if (res.data.reset_url) {
      debugUrl.value = res.data.reset_url
    }
    captchaToken.value = ''
    captchaRef.value?.refresh()
  } catch {
    captchaToken.value = ''
    captchaRef.value?.refresh()
  } finally {
    loading.value = false
  }
}

async function resetPassword() {
  const ok = await validateFields([passwordItemRef.value, confirmItemRef.value])
  if (!ok) return
  if (!captchaToken.value) {
    window.$message?.warning?.('请先完成滑块验证')
    return
  }
  loading.value = true
  try {
    const encrypted = await encryptPasswords({ password: form.password })
    await authApi.resetPassword({
      token: form.token,
      password_cipher: encrypted.values.password || '',
      captcha_token: captchaToken.value,
    })
    window.$message.success('密码已重置，请重新登录')
    router.push('/login')
  } catch {
    captchaToken.value = ''
    captchaRef.value?.refresh()
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthLayout
    :title="isResetMode ? '重置密码' : '找回密码'"
    :description="
      isResetMode
        ? '请设置新密码。重置链接在过期前仅可使用一次。'
        : '请输入已启用管理端登录的邮箱，系统将发送密码重置链接。'
    "
  >
    <template #headerExtra>
      <n-button text type="primary" @click="router.push('/login')"> 返回登录 </n-button>
    </template>

    <form class="auth-form" @submit.prevent="isResetMode ? resetPassword() : sendLink()">
      <n-form-item v-if="!isResetMode" ref="emailItemRef" :show-label="false" :rule="emailRule">
        <n-input v-model:value="form.email" size="large" clearable placeholder="登录邮箱" />
      </n-form-item>

      <template v-if="isResetMode">
        <n-form-item ref="passwordItemRef" :show-label="false" :rule="passwordRule">
          <n-input
            v-model:value="form.password"
            size="large"
            type="password"
            show-password-on="click"
            placeholder="新密码（至少 6 位）"
          />
        </n-form-item>
        <n-form-item ref="confirmItemRef" :show-label="false" :rule="confirmPasswordRule">
          <n-input
            v-model:value="form.confirmPassword"
            size="large"
            type="password"
            show-password-on="click"
            placeholder="确认新密码"
          />
        </n-form-item>
      </template>

      <AdminSliderCaptcha ref="captchaRef" v-model:captcha-token="captchaToken" />

      <n-button type="primary" size="large" block attr-type="submit" :loading="loading">
        {{ isResetMode ? '重置密码' : '发送重置链接' }}
      </n-button>

      <n-text
        v-if="debugUrl"
        depth="3"
        style="display: block; margin-top: 12px; word-break: break-all"
      >
        调试链接：
        <n-a :href="debugUrl" target="_blank" rel="noopener noreferrer">
          {{ debugUrl }}
        </n-a>
      </n-text>
    </form>
  </AuthLayout>
</template>
