<script setup lang="ts">
import {
  AllApplication,
  ApplicationMenu,
  BookOpen,
  Calendar,
  Close,
  Code,
  Copy,
  DocDetail,
  Down,
  Up,
  FolderClose,
  Github,
  HamburgerButton,
  Home,
  Link,
  Local,
  PreviewOpen,
  Refresh,
  Rss,
  Search,
  TagOne,
  Time,
  User,
  Word,
} from '@icon-park/vue-next'
import type { Icon as IconParkComp } from '@icon-park/vue-next/lib/runtime'
import { Icon as IconifyIcon } from '@iconify/vue/offline'
import { computed } from 'vue'

export type AppIconName =
  | 'folder'
  | 'clone'
  | 'tags'
  | 'github'
  | 'user'
  | 'file'
  | 'code'
  | 'rss'
  | 'menu'
  | 'close'
  | 'down'
  | 'up'
  | 'apps'
  | 'home'
  | 'calendar'
  | 'time'
  | 'local'
  | 'refresh'
  | 'book'
  | 'link'
  | 'word'
  | 'search'
  | 'view'

/** 历史短名 → Iconify，与后台 IconSelect（icon-park-outline）对齐 */
const LEGACY_TO_ICONIFY: Record<AppIconName, string> = {
  folder: 'icon-park-outline:folder-close',
  clone: 'icon-park-outline:copy',
  tags: 'icon-park-outline:tag-one',
  github: 'icon-park-outline:github',
  user: 'icon-park-outline:user',
  file: 'icon-park-outline:doc-detail',
  code: 'icon-park-outline:code',
  rss: 'icon-park-outline:rss',
  menu: 'icon-park-outline:hamburger-button',
  close: 'icon-park-outline:close',
  down: 'icon-park-outline:down',
  up: 'icon-park-outline:up',
  apps: 'icon-park-outline:application-menu',
  home: 'icon-park-outline:home',
  calendar: 'icon-park-outline:calendar',
  time: 'icon-park-outline:time',
  local: 'icon-park-outline:local',
  refresh: 'icon-park-outline:refresh',
  book: 'icon-park-outline:book-open',
  link: 'icon-park-outline:link',
  word: 'icon-park-outline:word',
  search: 'icon-park-outline:search',
  view: 'icon-park-outline:preview-open',
}

const icons: Record<AppIconName, IconParkComp> = {
  folder: FolderClose,
  clone: Copy,
  tags: TagOne,
  github: Github,
  user: User,
  file: DocDetail,
  code: Code,
  rss: Rss,
  menu: HamburgerButton,
  close: Close,
  down: Down,
  up: Up,
  apps: ApplicationMenu,
  home: Home,
  calendar: Calendar,
  time: Time,
  local: Local,
  refresh: Refresh,
  book: BookOpen,
  link: Link,
  word: Word,
  search: Search,
  view: PreviewOpen,
}

const props = withDefaults(
  defineProps<{
    name: AppIconName | string
    size?: number | string
    theme?: 'outline' | 'filled' | 'two-tone' | 'multi-color'
  }>(),
  {
    size: 18,
    theme: 'outline',
  },
)

const iconifyName = computed(() => {
  const raw = String(props.name || '').trim()
  if (!raw) return ''
  if (raw.includes(':')) return raw
  return LEGACY_TO_ICONIFY[raw as AppIconName] || ''
})

const parkComp = computed(() => icons[props.name as AppIconName] ?? AllApplication)
</script>

<template>
  <!-- 优先 Iconify（后台选择器写入的完整名 / 短名映射） -->
  <IconifyIcon
    v-if="iconifyName"
    class="app-icon inline-flex shrink-0 align-middle"
    :icon="iconifyName"
    :width="size"
    :height="size"
  />
  <component
    v-else
    :is="parkComp"
    class="app-icon inline-flex shrink-0 align-middle"
    :theme="theme"
    :size="size"
    fill="currentColor"
    :stroke-width="3"
  />
</template>
