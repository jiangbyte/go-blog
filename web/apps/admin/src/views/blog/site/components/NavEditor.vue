<!-- Author: go-blog — 站点导航 / 外链编辑 -->

<script setup lang="ts">
import IconSelect from '@/components/common/IconSelect.vue'
import { NButton, NCheckbox, NDynamicInput, NFlex, NFormItem, NInput, NText } from 'naive-ui'
import { computed } from 'vue'

export type NavChild = {
  name: string
  href?: string
  icon?: string
  external?: boolean
}

export type NavItem = {
  name: string
  href?: string
  icon?: string
  external?: boolean
  children?: NavChild[]
}

export type NavGroupItem = {
  name: string
  url: string
  icon?: string
}

const nav = defineModel<NavItem[]>('nav', { default: () => [] })
const navGroup = defineModel<NavGroupItem[]>('navGroup', { default: () => [] })

const props = defineProps<{ disabled?: boolean }>()

const navModel = computed({
  get: () => (Array.isArray(nav.value) ? nav.value : []),
  set: (v) => {
    nav.value = v
  },
})

const groupModel = computed({
  get: () => (Array.isArray(navGroup.value) ? navGroup.value : []),
  set: (v) => {
    navGroup.value = v
  },
})

function createNavItem(): NavItem {
  return { name: '新菜单', href: '', icon: '', external: false, children: [] }
}

function createChild(): NavChild {
  return { name: '子项', href: '/', icon: '', external: false }
}

function createGroupItem(): NavGroupItem {
  return { name: '链接', url: 'https://', icon: '' }
}

function ensureChildren(item: NavItem) {
  if (!Array.isArray(item.children)) item.children = []
  return item.children
}
</script>

<template>
  <NFlex vertical :size="20">
    <div>
      <NText strong>主导航</NText>
      <NText depth="3" class="ml-2 text-xs">前台顶栏与移动端抽屉共用</NText>
      <NDynamicInput
        v-model:value="navModel"
        class="mt-2"
        :disabled="disabled"
        :on-create="createNavItem"
      >
        <template #default="{ value }">
          <div class="nav-editor__item">
            <NFlex :size="8" wrap align="center" class="mb-2">
              <NInput
                v-model:value="value.name"
                class="nav-editor__field"
                placeholder="名称"
                :disabled="disabled"
              />
              <NInput
                v-model:value="value.href"
                class="nav-editor__field"
                placeholder="链接（有子菜单可空）"
                :disabled="disabled"
              />
              <IconSelect
                v-model:value="value.icon"
                size="small"
                compact
                :disabled="disabled"
                placeholder="图标"
              />
              <NCheckbox v-model:checked="value.external" :disabled="disabled">外链</NCheckbox>
            </NFlex>
            <NFormItem label="子菜单" :show-feedback="false" label-placement="left">
              <NDynamicInput
                v-model:value="value.children"
                :disabled="disabled"
                :on-create="createChild"
              >
                <template #default="{ value: child }">
                  <NFlex :size="8" wrap align="center">
                    <NInput
                      v-model:value="child.name"
                      class="nav-editor__field"
                      placeholder="名称"
                      :disabled="disabled"
                    />
                    <NInput
                      v-model:value="child.href"
                      class="nav-editor__field"
                      placeholder="/path 或 https://"
                      :disabled="disabled"
                    />
                    <IconSelect
                      v-model:value="child.icon"
                      size="small"
                      compact
                      :disabled="disabled"
                      placeholder="图标"
                    />
                    <NCheckbox v-model:checked="child.external" :disabled="disabled">外链</NCheckbox>
                  </NFlex>
                </template>
              </NDynamicInput>
            </NFormItem>
            <NButton
              v-if="!ensureChildren(value).length"
              size="tiny"
              quaternary
              :disabled="disabled"
              @click="value.children = [createChild()]"
            >
              添加子菜单
            </NButton>
          </div>
        </template>
      </NDynamicInput>
    </div>

    <div>
      <NText strong>快捷链接</NText>
      <NText depth="3" class="ml-2 text-xs">顶栏「链接」与移动端底部外链</NText>
      <NDynamicInput
        v-model:value="groupModel"
        class="mt-2"
        :disabled="disabled"
        :on-create="createGroupItem"
      >
        <template #default="{ value }">
          <NFlex :size="8" wrap align="center">
            <NInput
              v-model:value="value.name"
              class="nav-editor__field"
              placeholder="名称"
              :disabled="disabled"
            />
            <NInput
              v-model:value="value.url"
              class="nav-editor__field"
              placeholder="https://"
              :disabled="disabled"
            />
            <IconSelect
              v-model:value="value.icon"
              size="small"
              compact
              :disabled="disabled"
              placeholder="图标"
            />
          </NFlex>
        </template>
      </NDynamicInput>
    </div>
  </NFlex>
</template>

<style scoped>
.nav-editor__item {
  width: 100%;
  padding: 4px 0;
  background: transparent;
}
.nav-editor__field {
  width: 160px;
}
</style>
