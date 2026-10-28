<!-- Author: go-blog — 左侧模块轨（NButton，跟随系统主题） -->

<script setup lang="ts">
import { Icon } from '@iconify/vue/offline'
import { resolveResumeIcon } from '@jiangbyte/shared'
import { NButton, NFlex, NIcon, NTooltip } from 'naive-ui'

export type RailKey = 'profile' | 'theme' | `section:${string}`

const props = defineProps<{
  active: RailKey
  sections: { id: string; title: string; icon?: string; meta_json?: { visible?: boolean } }[]
  canWrite: boolean
}>()

const emit = defineEmits<{
  select: [key: RailKey]
  add: []
}>()

function isActive(key: RailKey) {
  return props.active === key
}

/** 分节是否在纸面展示（缺省 true）。 */
function isVisible(sec: { meta_json?: { visible?: boolean } }) {
  return sec.meta_json?.visible !== false
}
</script>

<template>
  <NFlex class="h-full py-2" vertical align="center" justify="space-between" :size="4">
    <NFlex vertical align="center" :size="4">
      <NTooltip placement="left">
        <template #trigger>
          <NButton
            :type="isActive('profile') ? 'primary' : 'default'"
            :quaternary="!isActive('profile')"
            circle
            aria-label="基本信息"
            @click="emit('select', 'profile')"
          >
            <template #icon>
              <NIcon :size="18">
                <Icon icon="lucide:user" />
              </NIcon>
            </template>
          </NButton>
        </template>
        基本信息
      </NTooltip>

      <NTooltip v-for="sec in sections" :key="sec.id" placement="left">
        <template #trigger>
          <NButton
            :type="isActive(`section:${sec.id}`) ? 'primary' : 'default'"
            :quaternary="!isActive(`section:${sec.id}`)"
            circle
            :aria-label="sec.title"
            :style="isVisible(sec) ? undefined : { opacity: 0.4 }"
            @click="emit('select', `section:${sec.id}`)"
          >
            <template #icon>
              <NIcon :size="18">
                <Icon :icon="resolveResumeIcon(sec.icon || 'lucide:file-text')" />
              </NIcon>
            </template>
          </NButton>
        </template>
        {{ sec.title }}{{ isVisible(sec) ? '' : '（已隐藏）' }}
      </NTooltip>

      <NTooltip placement="left">
        <template #trigger>
          <NButton
            :type="isActive('theme') ? 'primary' : 'default'"
            :quaternary="!isActive('theme')"
            circle
            aria-label="外观"
            @click="emit('select', 'theme')"
          >
            <template #icon>
              <NIcon :size="18">
                <Icon icon="lucide:palette" />
              </NIcon>
            </template>
          </NButton>
        </template>
        外观
      </NTooltip>
    </NFlex>

    <NTooltip v-if="canWrite" placement="left">
      <template #trigger>
        <NButton circle dashed aria-label="添加模块" @click="emit('add')">
          <template #icon>
            <NIcon :size="18">
              <Icon icon="lucide:plus" />
            </NIcon>
          </template>
        </NButton>
      </template>
      添加模块
    </NTooltip>
  </NFlex>
</template>
