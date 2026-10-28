<!-- Author: go-blog — 精简个人中心：公开资料 + 修改密码 -->

<script setup lang="ts">
import { ProCard } from 'pro-naive-ui'
import { reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import './profile.css'
import BasicInfoPanel from './components/BasicInfoPanel.vue'
import PasswordPanel from './components/PasswordPanel.vue'
import ProfileSummary from './components/ProfileSummary.vue'

const route = useRoute()
const router = useRouter()

const TAB_KEYS = ['basic_info', 'password'] as const
type TabKey = (typeof TAB_KEYS)[number]

const state = reactive({
  activeTab: resolveInitialTab(),
})

function resolveInitialTab(): TabKey {
  const tab = typeof route.query.tab === 'string' ? route.query.tab : ''
  if (tab === 'password' || tab === 'security') {
    return 'password'
  }
  return 'basic_info'
}

function onTabChange(key: string) {
  if (!TAB_KEYS.includes(key as TabKey) || state.activeTab === key) {
    return
  }
  state.activeTab = key as TabKey
  void router.replace({ query: { ...route.query, tab: key } })
}

watch(
  () => route.query.tab,
  (tab) => {
    const next = tab === 'password' || tab === 'security' ? 'password' : 'basic_info'
    if (state.activeTab !== next) {
      state.activeTab = next
    }
  },
)
</script>

<template>
  <div class="profile w-full min-w-0">
    <div class="profile__layout">
      <ProCard
        class="profile__summary-card"
        content-class="profile__summary-body"
        :show-collapse="false"
      >
        <ProfileSummary />
      </ProCard>

      <ProCard
        class="profile__content-card"
        content-class="profile__content-body"
        :show-collapse="false"
      >
        <!-- 面板放进 TabPane，避免导航与内容脱节导致改密区不可见 -->
        <NTabs
          class="profile__tabs"
          type="line"
          :value="state.activeTab"
          @update:value="onTabChange"
        >
          <NTabPane name="basic_info" tab="公开资料" display-directive="show:lazy">
            <div class="profile__panel">
              <BasicInfoPanel />
            </div>
          </NTabPane>
          <NTabPane name="password" tab="修改密码" display-directive="show:lazy">
            <div class="profile__panel">
              <PasswordPanel />
            </div>
          </NTabPane>
        </NTabs>
      </ProCard>
    </div>
  </div>
</template>
