<!-- Author: go-blog — 工作台顶栏（占文档流，禁止 absolute 撑出横向滚动） -->

<script setup lang="ts">
import { Icon } from '@iconify/vue/offline'
import { NButton, NFlex, NIcon, NTag, NText } from 'naive-ui'

defineProps<{
  title: string
  dirty: boolean
  saving: boolean
  canSave: boolean
  canPublish: boolean
  canUnpublish: boolean
  canExport: boolean
  status: string
}>()

const emit = defineEmits<{
  back: []
  save: []
  publish: []
  unpublish: []
  preview: []
  download: []
}>()
</script>

<template>
  <header class="resume-wb__header px-3">
    <NFlex class="h-full" align="center" justify="space-between" :size="12" :wrap="false">
      <NFlex align="center" :size="10" class="min-w-0" style="flex: 1 1 0%; overflow: hidden">
        <NButton quaternary size="small" @click="emit('back')">
          <template #icon>
            <NIcon>
              <Icon icon="icon-park-outline:left" />
            </NIcon>
          </template>
          返回
        </NButton>
        <NText strong class="truncate" style="max-width: 200px">
          {{ title || '未命名简历' }}
        </NText>
        <NTag v-if="dirty" size="small" type="warning" :bordered="false"> 未保存 </NTag>
        <NTag v-else-if="status === 'published'" size="small" type="success" :bordered="false">
          已发布
        </NTag>
      </NFlex>
      <NFlex align="center" :size="8" style="flex: none">
        <NButton size="small" quaternary @click="emit('preview')"> 预览 </NButton>
        <NButton v-if="canExport" size="small" quaternary @click="emit('download')"> 下载 </NButton>
        <NButton
          v-if="canPublish && status !== 'published'"
          size="small"
          type="success"
          @click="emit('publish')"
        >
          发布
        </NButton>
        <NButton
          v-if="canPublish && status === 'published'"
          size="small"
          type="success"
          @click="emit('publish')"
        >
          更新并发布
        </NButton>
        <NButton
          v-if="canUnpublish && status === 'published'"
          size="small"
          type="warning"
          @click="emit('unpublish')"
        >
          取消发布
        </NButton>
        <NButton v-if="canSave" size="small" type="primary" :loading="saving" @click="emit('save')">
          保存
        </NButton>
      </NFlex>
    </NFlex>
  </header>
</template>
