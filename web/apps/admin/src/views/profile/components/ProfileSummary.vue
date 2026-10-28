<!-- Author: go-blog — 左侧摘要：头像 + 账号信息 -->

<script setup lang="ts">
import { useAuthStore } from '@/stores'
import { computed, onMounted, reactive } from 'vue'
import { displayValue, genderLabel } from '../composables/useProfile'
import '../profile.css'
import AvatarUploadModal from './AvatarUploadModal.vue'

const authStore = useAuthStore()
const avatarImgProps = { referrerPolicy: 'no-referrer' } as any

const state = reactive({
  loading: false,
  avatarModalShow: false,
  me: null as any,
})

/** 优先 authStore（保存公开资料后会刷新），其次详情接口缓存 */
const avatarUrl = computed(() => authStore.userInfo?.avatar || state.me?.avatar || undefined)
const displayName = computed(() => {
  const nickname = String(authStore.userInfo?.nickname ?? state.me?.nickname ?? '').trim()
  return nickname || '-'
})
const accountText = computed(() =>
  String(authStore.userInfo?.account ?? state.me?.username ?? state.me?.account ?? '').trim(),
)
const userId = computed(() => String(authStore.userInfo?.accountId ?? state.me?.id ?? '').trim())
const genderText = computed(() => genderLabel(authStore.userInfo?.gender ?? state.me?.gender))
const signatureText = computed(() => {
  const text = String(authStore.userInfo?.signature ?? state.me?.signature ?? '').trim()
  return text || '暂无签名'
})
const userTypeText = computed(() =>
  displayValue(authStore.userInfo?.accountType || state.me?.user_type),
)

onMounted(async () => {
  await refresh()
})

async function refresh() {
  state.loading = true
  try {
    state.me = await authStore.refreshUserInfo()
  } finally {
    state.loading = false
  }
}

defineExpose({ refresh })
</script>

<template>
  <aside class="profile__summary">
    <NSpin :show="state.loading">
      <div class="profile__avatar-card">
        <button
          class="profile__avatar-edit"
          type="button"
          title="更换头像"
          @click="state.avatarModalShow = true"
        >
          <NAvatar
            v-if="avatarUrl"
            round
            :size="160"
            :src="avatarUrl"
            :img-props="avatarImgProps"
          />
          <NAvatar v-else round :size="160">
            <NovaIcon icon="icon-park-outline:user" :size="64" />
          </NAvatar>
          <span class="profile__avatar-badge">
            <NovaIcon icon="icon-park-outline:edit" :size="14" />
            编辑
          </span>
        </button>
        <div class="profile__avatar-name">
          {{ displayName }}
        </div>
        <div class="profile__avatar-account">
          {{ accountText || '-' }}
        </div>
        <NDescriptions class="profile__avatar-desc" :column="1" label-placement="left" size="small">
          <NDescriptionsItem label="用户 ID">
            {{ userId || '-' }}
          </NDescriptionsItem>
          <NDescriptionsItem label="类型">
            {{ userTypeText }}
          </NDescriptionsItem>
          <NDescriptionsItem label="性别">
            {{ genderText }}
          </NDescriptionsItem>
          <NDescriptionsItem label="签名">
            {{ signatureText }}
          </NDescriptionsItem>
        </NDescriptions>
      </div>
    </NSpin>

    <AvatarUploadModal
      v-model:show="state.avatarModalShow"
      :avatar="avatarUrl"
      @uploaded="refresh"
    />
  </aside>
</template>
