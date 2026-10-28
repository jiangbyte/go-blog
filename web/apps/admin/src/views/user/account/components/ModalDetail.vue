<!-- Author: go-blog — 用户详情 -->

<script setup lang="ts">
import { accountApi } from '@/api'
import { displayValue, formatDateTime, wireBool } from '@/utils'
import { reactive } from 'vue'

const state = reactive({
  showModal: false,
  loading: false,
  user: {} as any,
})

async function openModal(id: string, row?: any) {
  state.user = row ?? {}
  state.showModal = true
  state.loading = true
  try {
    const res = await accountApi.detail({ user_id: id })
    state.user = res.data ?? row ?? {}
  } catch {
    // 详情失败仍展示行数据
  } finally {
    state.loading = false
  }
}

defineExpose({ openModal })
</script>

<template>
  <DetailContainer
    v-model:show="state.showModal"
    title="用户详情"
    :width="560"
    :mask-closable="false"
  >
    <NSpin :show="state.loading">
      <NDescriptions label-placement="left" bordered :column="1">
        <NDescriptionsItem label="用户ID">
          {{ displayValue(state.user.id) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="账号">
          {{ displayValue(state.user.username) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="昵称">
          {{ displayValue(state.user.nickname) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="类型">
          {{ state.user.user_type === 'ADMIN' ? '管理员' : '前台' }}
        </NDescriptionsItem>
        <NDescriptionsItem label="状态">
          <NTag
            :type="wireBool(state.user.enabled) ? 'success' : 'default'"
            :bordered="false"
            size="small"
          >
            {{ wireBool(state.user.enabled) ? '启用' : '禁用' }}
          </NTag>
        </NDescriptionsItem>
        <NDescriptionsItem label="手机">
          {{ displayValue(state.user.phone) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="邮箱">
          {{ displayValue(state.user.email) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="签名">
          {{ displayValue(state.user.signature) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="更新时间">
          {{ formatDateTime(state.user.update_time || state.user.updated_at) }}
        </NDescriptionsItem>
      </NDescriptions>
    </NSpin>
  </DetailContainer>
</template>
