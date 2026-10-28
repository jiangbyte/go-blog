<!-- Author: go-blog — 角色详情 -->

<script setup lang="ts">
import { displayValue, formatDateTime, wireBool } from '@/utils'
import { reactive } from 'vue'

const state = reactive({
  showModal: false,
  loading: false,
  role: {} as any,
})

/** mypage 无 detail 接口，直接展示列表行 */
async function openModal(row: any) {
  state.role = row ?? {}
  state.showModal = true
}

defineExpose({
  openModal,
})
</script>

<template>
  <DetailContainer
    v-model:show="state.showModal"
    title="角色详情"
    :width="520"
    :mask-closable="false"
  >
    <NSpin :show="state.loading">
      <NDescriptions label-placement="left" bordered :column="1">
        <NDescriptionsItem label="角色ID">
          {{ displayValue(state.role.id) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="角色名称">
          {{ displayValue(state.role.name) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="角色编码">
          {{ displayValue(state.role.code) }}
        </NDescriptionsItem>
        <NDescriptionsItem label="状态">
          <NTag
            :type="wireBool(state.role.enabled) ? 'success' : 'default'"
            :bordered="false"
            size="small"
          >
            {{ wireBool(state.role.enabled) ? '启用' : '禁用' }}
          </NTag>
        </NDescriptionsItem>
        <NDescriptionsItem label="更新时间">
          {{ formatDateTime(state.role.update_time || state.role.updated_at) }}
        </NDescriptionsItem>
      </NDescriptions>
    </NSpin>
  </DetailContainer>
</template>
