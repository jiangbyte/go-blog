<!-- Author: go-blog -->

<script setup lang="ts">
import type { Component } from 'vue'
import { ProCard } from 'pro-naive-ui'
import { computed, reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import './config.css'
import SysDefaultsForm from './components/SysDefaultsForm.vue'
import LegalContentForm from './components/LegalContentForm.vue'
import OtherConfigPanel from './components/OtherConfigPanel.vue'

const route = useRoute()
const router = useRouter()

const navItems: Array<{ key: string; label: string }> = [
  { key: 'SYS', label: '站点信息' },
  { key: 'LEGAL', label: '服务条款' },
  { key: 'OTHER', label: '其他配置' },
]

const panelMap: Record<string, Component> = {
  SYS: SysDefaultsForm,
  LEGAL: LegalContentForm,
  OTHER: OtherConfigPanel,
}

const state = reactive({
  activeTab: resolveInitialTab(),
})

const activeNav = computed(
  () => navItems.find((item) => item.key === state.activeTab) ?? navItems[0],
)

function resolveInitialTab() {
  const tab = typeof route.query.tab === 'string' ? route.query.tab : ''
  if (tab && navItems.some((item) => item.key === tab)) {
    return tab
  }
  return 'SYS'
}

function selectTab(key: string) {
  if (!key || state.activeTab === key) return
  if (!navItems.some((item) => item.key === key)) return
  state.activeTab = key
  void router.replace({ query: { ...route.query, tab: key } })
}

watch(
  () => route.query.tab,
  (tab) => {
    if (typeof tab !== 'string' || !navItems.some((item) => item.key === tab)) return
    if (state.activeTab !== tab) state.activeTab = tab
  },
)
</script>

<template>
  <div class="sys-config h-full min-h-0">
    <ProCard
      class="sys-config__main"
      content-class="sys-config__main-body"
      :title="activeNav.label"
      :show-collapse="false"
      :trigger-areas="[]"
    >
      <NTabs
        class="sys-config__tabs"
        type="line"
        :value="state.activeTab"
        @update:value="selectTab"
      >
        <NTabPane v-for="item in navItems" :key="item.key" :name="item.key" :tab="item.label">
          <div class="sys-config__panel">
            <KeepAlive>
              <component :is="panelMap[item.key]" :key="item.key" />
            </KeepAlive>
          </div>
        </NTabPane>
      </NTabs>
    </ProCard>
  </div>
</template>
