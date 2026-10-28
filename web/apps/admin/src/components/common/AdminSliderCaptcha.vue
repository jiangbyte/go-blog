<!-- Author: go-blog — 点击弹出滑块，对接本后端 create/verify（对齐 mypage） -->

<script setup lang="ts">
import { PopupCaptcha } from '@captcha-pro/vue'
import '@captcha-pro/vue/style.css'
import { computed, ref } from 'vue'
import { authApi } from '@/api'

const captchaToken = defineModel<string>('captchaToken', { default: '' })

const remountKey = ref(0)
const triggerId = `admin-captcha-trigger-${Math.random().toString(36).slice(2, 9)}`

type CaptchaPayload = {
  captchaId?: string
  target?: Array<number | { x: number; y?: number }>
}

const backendVerify = {
  // 字段映射到 captcha-pro 约定
  getCaptcha: async () => {
    captchaToken.value = ''
    const res = await authApi.createCaptcha()
    const data = res.data
    return {
      data: {
        captchaId: data.captcha_id,
        bgImage: data.bg_url,
        sliderImage: data.puzzle_url,
        sliderY: data.y,
        width: data.bg_width,
        height: data.bg_height,
      },
    }
  },
  verify: async (payload: CaptchaPayload) => {
    const captchaId = String(payload.captchaId || '')
    const first = Array.isArray(payload.target) ? payload.target[0] : 0
    const x =
      typeof first === 'number'
        ? first
        : typeof first === 'object' && first && 'x' in first
          ? Number(first.x)
          : 0
    const res = await authApi.verifyCaptcha({ captcha_id: captchaId, x })
    captchaToken.value = res.data.captcha_token
    return { success: true }
  },
}

const captchaOptions = computed(() => ({
  width: 320,
  height: 160,
  sliderWidth: 60,
  sliderHeight: 60,
  precision: 5,
  verifyMode: 'backend' as const,
  locale: 'zh-CN' as const,
  backendVerify,
}))

function onFail() {
  captchaToken.value = ''
}

function onRefresh() {
  captchaToken.value = ''
  remountKey.value += 1
}

defineExpose({
  refresh: onRefresh,
  hasToken: () => Boolean(captchaToken.value),
})
</script>

<template>
  <div class="admin-slider-captcha">
    <!-- 已通过：展示状态条，可点重验 -->
    <button
      v-if="captchaToken"
      type="button"
      class="admin-slider-captcha__trigger is-passed"
      @click="onRefresh"
    >
      验证通过，点击可重验
    </button>

    <!-- 未通过：点击后弹出拼图，避免常驻占位 -->
    <PopupCaptcha
      v-else
      :key="remountKey"
      type="slider"
      :trigger="`#${triggerId}`"
      :auto-close="true"
      :close-delay="400"
      :modal="{ title: '安全验证', showClose: true, maskClosable: true, escClosable: true }"
      :captcha-options="captchaOptions"
      @fail="onFail"
    >
      <button :id="triggerId" type="button" class="admin-slider-captcha__trigger">
        点击完成安全验证
      </button>
    </PopupCaptcha>
  </div>
</template>

<style scoped>
.admin-slider-captcha {
  width: 100%;
  margin-bottom: 12px;
}

.admin-slider-captcha__trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 42px;
  border: 1px solid var(--n-border-color, #e4e7eb);
  border-radius: 6px;
  background: var(--n-color, #f7f9fa);
  color: var(--n-text-color, #45494c);
  font-size: 14px;
  cursor: pointer;
  transition:
    border-color 0.15s,
    color 0.15s,
    background 0.15s;
}

.admin-slider-captcha__trigger:hover {
  border-color: var(--primary-color, #18a058);
  color: var(--primary-color, #18a058);
}

.admin-slider-captcha__trigger.is-passed {
  border-color: #52ccba;
  background: #d2f4ef;
  color: #1f8f7f;
}
</style>
