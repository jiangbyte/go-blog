<!-- Author: go-blog — 存储引擎状态弹窗（只读，配置来自 .env） -->

<script setup lang="ts">
import { formatFileSize } from '@/utils'
import { NFlex, NModal, NTag, NText } from 'naive-ui'
import { computed, ref } from 'vue'

export type EngineInfo = {
  engine: string
  provider: string
  display_name: string
  configured: boolean
  is_default: boolean
  bucket?: string | null
  endpoint?: string | null
  bucket_public?: boolean
}

const props = defineProps<{
  engines: EngineInfo[]
  hint?: string
  defaultEngine?: string
  uploadMaxBytes?: number
  presignExpireSeconds?: number
}>()

const show = ref(false)

const hintText = computed(
  () =>
    props.hint ||
    '切换默认引擎：修改服务器环境变量 DEFAULT_FILE_ENGINE 后重启。单次上传可在列表上方指定存储。',
)

const limitsText = computed(() => {
  const parts: string[] = []
  if (props.defaultEngine) {
    parts.push(`默认引擎：${props.defaultEngine}`)
  }
  if (props.uploadMaxBytes != null && props.uploadMaxBytes > 0) {
    parts.push(`单文件上限：${formatFileSize(props.uploadMaxBytes)}`)
  }
  if (props.presignExpireSeconds != null && props.presignExpireSeconds > 0) {
    parts.push(`签名链接有效期：${props.presignExpireSeconds}s`)
  }
  return parts.join(' · ')
})

function open() {
  show.value = true
}

defineExpose({ open })
</script>

<template>
  <NModal
    v-model:show="show"
    preset="card"
    title="存储引擎"
    :style="{ width: '560px' }"
    :mask-closable="true"
  >
    <NText v-if="limitsText" depth="3" class="mb-3 block text-xs">
      {{ limitsText }}
    </NText>
    <NFlex v-if="engines?.length" :size="12" wrap>
      <div
        v-for="engine in engines"
        :key="engine.engine"
        class="engine-card"
        :class="{ 'engine-card--default': engine.is_default }"
      >
        <div class="engine-card__title">
          <span>{{ engine.display_name }}</span>
          <NTag v-if="engine.is_default" size="tiny" type="success">默认</NTag>
          <NTag :type="engine.configured ? 'info' : 'warning'" size="tiny">
            {{ engine.configured ? '已配置' : '未配置' }}
          </NTag>
        </div>
        <div class="engine-card__meta">
          <div>引擎：{{ engine.engine }}</div>
          <div>Provider：{{ engine.provider }}</div>
          <div>桶：{{ engine.bucket || '—' }}</div>
          <div>端点：{{ engine.endpoint || '—' }}</div>
          <div>公开桶：{{ engine.bucket_public ? '是' : '否' }}</div>
        </div>
      </div>
    </NFlex>
    <NText v-else depth="3">暂无存储状态</NText>
    <NText depth="3" class="mt-3 block text-xs">
      {{ hintText }}
    </NText>
  </NModal>
</template>

<style scoped>
.engine-card {
  min-width: 200px;
  flex: 1 1 200px;
  max-width: 260px;
  padding: 10px 12px;
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  background: var(--n-color);
}
.engine-card--default {
  border-color: var(--n-primary-color);
}
.engine-card__title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  margin-bottom: 6px;
}
.engine-card__meta {
  font-size: 12px;
  line-height: 1.6;
  color: var(--n-text-color-3);
  word-break: break-all;
}
</style>
