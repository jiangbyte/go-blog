<!-- Author: go-blog — 基本信息侧栏面板（自绘分区标题，避开 NDivider 叠层） -->

<script setup lang="ts">
import { Icon } from '@iconify/vue/offline'
import IconSelect from '@/components/common/IconSelect.vue'
import ResumeAvatarUpload from '../components/ResumeAvatarUpload.vue'
import { NButton, NFlex, NForm, NFormItem, NIcon, NInput, NText } from 'naive-ui'

const profile = defineModel<{
  name: string
  tagline: string
  avatar_url: string
  github_url: string
  fields: { key: string; label: string; value: string; icon?: string }[]
}>('profile', { required: true })

const title = defineModel<string>('title', { required: true })

defineProps<{ canWrite: boolean }>()

function addField() {
  profile.value.fields.push({
    key: `f_${Date.now()}`,
    label: '',
    value: '',
    icon: 'lucide:circle',
  })
}

function removeField(idx: number) {
  profile.value.fields.splice(idx, 1)
}

/** 上下调换联系字段顺序。 */
function moveField(idx: number, dir: -1 | 1) {
  const j = idx + dir
  const list = profile.value.fields
  if (j < 0 || j >= list.length) return
  ;[list[idx], list[j]] = [list[j], list[idx]]
}
</script>

<template>
  <NForm class="profile-panel" label-placement="top" size="small" :show-feedback="false">
    <NFormItem label="管理标题">
      <NInput v-model:value="title" :disabled="!canWrite" placeholder="仅后台可见" />
    </NFormItem>

    <div class="profile-panel__section">
      <div class="profile-panel__heading">个人资料</div>
      <NFormItem label="姓名">
        <NInput v-model:value="profile.name" :disabled="!canWrite" placeholder="纸面显示的姓名" />
      </NFormItem>
      <NFormItem label="头像">
        <ResumeAvatarUpload v-model:value="profile.avatar_url" :disabled="!canWrite" />
      </NFormItem>
    </div>

    <div class="profile-panel__section">
      <div class="profile-panel__heading">联系字段</div>
      <NText depth="3" class="profile-panel__hint">
        信息展示选「图标」时，使用此处为每项选择的图标
      </NText>
      <div
        v-for="(field, idx) in profile.fields"
        :key="field.key || idx"
        class="profile-panel__field"
      >
        <IconSelect
          v-model:value="field.icon"
          prefix="lucide"
          size="small"
          compact
          :clearable="false"
          :disabled="!canWrite"
          placeholder="图标"
        />
        <NInput
          v-model:value="field.label"
          placeholder="标签"
          :disabled="!canWrite"
          class="profile-panel__field-label"
        />
        <NInput
          v-model:value="field.value"
          placeholder="值"
          :disabled="!canWrite"
          class="profile-panel__field-value"
        />
        <NFlex v-if="canWrite" :size="0" align="center" class="profile-panel__field-actions">
          <NButton
            quaternary
            size="tiny"
            title="上移"
            aria-label="上移"
            :disabled="idx === 0"
            @click="moveField(idx, -1)"
          >
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:up" />
              </NIcon>
            </template>
          </NButton>
          <NButton
            quaternary
            size="tiny"
            title="下移"
            aria-label="下移"
            :disabled="idx === profile.fields.length - 1"
            @click="moveField(idx, 1)"
          >
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:down" />
              </NIcon>
            </template>
          </NButton>
          <NButton
            quaternary
            type="error"
            size="tiny"
            title="删除"
            aria-label="删除"
            @click="removeField(idx)"
          >
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:delete" />
              </NIcon>
            </template>
          </NButton>
        </NFlex>
      </div>
      <NButton v-if="canWrite" dashed block class="profile-panel__add" @click="addField">
        添加字段
      </NButton>
    </div>
  </NForm>
</template>

<style scoped>
.profile-panel {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  overflow-x: hidden;
}

.profile-panel :deep(.n-form-item) {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  /* 禁止 top 布局里 blank 行用 1fr 撑高，避免分区叠层 */
  grid-template-rows: auto auto auto !important;
  margin-bottom: 12px;
}

.profile-panel :deep(.n-form-item-blank) {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  justify-content: flex-start;
  overflow: hidden;
}

.profile-panel :deep(.n-form-item-blank > .n-input) {
  width: 100%;
  max-width: 100%;
  min-width: 0;
}

.profile-panel__section {
  margin-top: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color, #efeff5);
}

.profile-panel__heading {
  margin: 0 0 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-1, #333);
  line-height: 1.4;
}

.profile-panel__hint {
  display: block;
  margin-bottom: 10px;
  font-size: 12px;
  line-height: 1.5;
}

.profile-panel__field {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 6px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  margin-bottom: 8px;
  overflow: hidden;
}

.profile-panel__field-label {
  flex: 0 0 56px;
  width: 56px;
  min-width: 0;
}

.profile-panel__field-value {
  flex: 1 1 0%;
  min-width: 0;
}

.profile-panel__field-actions {
  flex: none;
}

.profile-panel__add {
  margin-top: 4px;
  width: 100%;
  max-width: 100%;
}
</style>
