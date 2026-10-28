<!-- Author: go-blog — 导航菜单 PC / 移动端可视化预览 -->

<script setup lang="ts">
import NovaIcon from '@/components/common/NovaIcon.vue'
import type { NavGroupItem, NavItem } from '@/views/blog/site/components/NavEditor.vue'
import { NRadioButton, NRadioGroup, NText } from 'naive-ui'
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    siteName?: string
    siteIcon?: string
    nav?: NavItem[]
    navGroup?: NavGroupItem[]
  }>(),
  {
    siteName: '',
    siteIcon: '',
    nav: () => [],
    navGroup: () => [],
  },
)

const mode = ref<'pc' | 'mobile'>('pc')

const brandName = computed(() => String(props.siteName || '').trim() || '站点名')
const brandIcon = computed(() => String(props.siteIcon || '').trim())

/** 短名 / 完整 Iconify 名 → 可渲染的 icon 字符串 */
function resolveIcon(raw?: string) {
  const name = String(raw || '').trim()
  if (!name) return ''
  if (name.includes(':')) return name
  const legacy: Record<string, string> = {
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
  return legacy[name] || `icon-park-outline:${name}`
}

const navList = computed(() => (Array.isArray(props.nav) ? props.nav : []))
const groupList = computed(() => (Array.isArray(props.navGroup) ? props.navGroup : []))

/** PC：悬停展开的菜单键；null 表示全部收起 */
const openGroup = ref<string | null>(null)
/** 移动端：点击展开的分组，默认全部收起 */
const mobileExpanded = ref<Record<string, boolean>>({})

watch(navList, () => {
  openGroup.value = null
})

function setOpenGroup(key: string | null) {
  openGroup.value = key
}

function toggleMobile(name: string) {
  mobileExpanded.value[name] = !mobileExpanded.value[name]
}

function visibleChildren(item: NavItem) {
  return (item.children || []).filter((c) => String(c.name || '').trim())
}
</script>

<template>
  <div class="nav-preview">
    <div class="nav-preview__head">
      <NText strong>菜单预览</NText>
      <NRadioGroup v-model:value="mode" size="small">
        <NRadioButton value="pc">PC</NRadioButton>
        <NRadioButton value="mobile">移动端</NRadioButton>
      </NRadioGroup>
    </div>

    <!-- PC 顶栏：悬停展开下拉，离开收起 -->
    <div v-if="mode === 'pc'" class="nav-preview__pc">
      <div class="nav-preview__stage">
        <div class="nav-preview__bar">
          <div class="nav-preview__bar-left">
            <div
              class="nav-preview__apps"
              @mouseenter="setOpenGroup('links')"
              @mouseleave="setOpenGroup(null)"
            >
              <span class="nav-preview__chip nav-preview__chip--icon" title="快捷链接">
                <NovaIcon icon="icon-park-outline:application-menu" :size="16" />
              </span>
              <div
                v-if="openGroup === 'links' && groupList.length"
                class="nav-preview__flyout nav-preview__flyout--links"
              >
                <div class="nav-preview__flyout-inner">
                  <p class="nav-preview__flyout-label">链接</p>
                  <div v-for="link in groupList" :key="link.name" class="nav-preview__flyout-item">
                    <NovaIcon
                      v-if="resolveIcon(link.icon)"
                      :icon="resolveIcon(link.icon)"
                      :size="14"
                    />
                    <span>{{ link.name || '未命名' }}</span>
                  </div>
                </div>
              </div>
            </div>
            <span class="nav-preview__brand">{{ brandName }}</span>
          </div>

          <div class="nav-preview__bar-nav">
            <div
              v-for="item in navList"
              :key="item.name"
              class="nav-preview__nav-item"
              @mouseenter="
                visibleChildren(item).length ? setOpenGroup(item.name) : setOpenGroup(null)
              "
              @mouseleave="setOpenGroup(null)"
            >
              <span v-if="visibleChildren(item).length" class="nav-preview__chip">
                {{ item.name || '未命名' }}
                <NovaIcon icon="icon-park-outline:down" :size="12" />
              </span>
              <span v-else class="nav-preview__chip">
                <NovaIcon
                  v-if="resolveIcon(item.icon)"
                  :icon="resolveIcon(item.icon)"
                  :size="14"
                />
                {{ item.name || '未命名' }}
              </span>
              <div
                v-if="visibleChildren(item).length && openGroup === item.name"
                class="nav-preview__flyout"
              >
                <div class="nav-preview__flyout-inner">
                  <div
                    v-for="child in visibleChildren(item)"
                    :key="child.name"
                    class="nav-preview__flyout-item"
                  >
                    <NovaIcon
                      v-if="resolveIcon(child.icon)"
                      :icon="resolveIcon(child.icon)"
                      :size="14"
                    />
                    <span>{{ child.name }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <span class="nav-preview__chip nav-preview__chip--icon" title="搜索">
            <NovaIcon icon="icon-park-outline:search" :size="16" />
          </span>
        </div>
      </div>
      <NText depth="3" class="nav-preview__hint">悬停菜单可展开下拉</NText>
    </div>

    <!-- 移动端抽屉 -->
    <div v-else class="nav-preview__phone">
      <div class="nav-preview__phone-frame">
        <header class="nav-preview__phone-head">
          <div class="nav-preview__phone-brand">
            <img
              v-if="brandIcon"
              :src="brandIcon"
              alt=""
              class="nav-preview__phone-avatar"
              referrerpolicy="no-referrer"
            />
            <span class="nav-preview__phone-name">{{ brandName }}</span>
          </div>
          <NovaIcon icon="icon-park-outline:close" :size="16" />
        </header>

        <div class="nav-preview__phone-body">
          <div class="nav-preview__phone-row">
            <NovaIcon icon="icon-park-outline:search" :size="14" />
            <span>搜索</span>
          </div>

          <template v-for="item in navList" :key="item.name">
            <div v-if="visibleChildren(item).length" class="nav-preview__phone-group">
              <button
                type="button"
                class="nav-preview__phone-row nav-preview__phone-row--btn"
                @click="toggleMobile(item.name)"
              >
                <span>{{ item.name || '未命名' }}</span>
                <NovaIcon
                  :icon="
                    mobileExpanded[item.name]
                      ? 'icon-park-outline:up'
                      : 'icon-park-outline:down'
                  "
                  :size="12"
                />
              </button>
              <div v-if="mobileExpanded[item.name]" class="nav-preview__phone-children">
                <div
                  v-for="child in visibleChildren(item)"
                  :key="child.name"
                  class="nav-preview__phone-row nav-preview__phone-row--child"
                >
                  <NovaIcon
                    v-if="resolveIcon(child.icon)"
                    :icon="resolveIcon(child.icon)"
                    :size="14"
                  />
                  <span>{{ child.name }}</span>
                </div>
              </div>
            </div>
            <div v-else class="nav-preview__phone-row">
              <NovaIcon
                v-if="resolveIcon(item.icon)"
                :icon="resolveIcon(item.icon)"
                :size="14"
              />
              <span>{{ item.name || '未命名' }}</span>
            </div>
          </template>

          <template v-if="groupList.length">
            <p class="nav-preview__phone-section">链接</p>
            <div v-for="link in groupList" :key="link.name" class="nav-preview__phone-row">
              <NovaIcon
                v-if="resolveIcon(link.icon)"
                :icon="resolveIcon(link.icon)"
                :size="14"
              />
              <span>{{ link.name || '未命名' }}</span>
            </div>
          </template>
        </div>
      </div>
      <NText depth="3" class="nav-preview__hint">点击分组可展开/收起子菜单</NText>
    </div>
  </div>
</template>

<style scoped>
.nav-preview {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
  overflow: visible;
}

.nav-preview__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.nav-preview__hint {
  display: block;
  margin-top: 8px;
  font-size: 12px;
}

.nav-preview__pc {
  min-width: 0;
  overflow: visible;
}

.nav-preview__stage {
  position: relative;
  /* 预留下拉面板高度，避免悬停菜单被裁切 */
  min-height: 220px;
  overflow: visible;
}

.nav-preview__bar {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 48px;
  padding: 8px 10px;
  overflow: visible;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: var(--card-color);
}

.nav-preview__bar-left {
  display: flex;
  flex: none;
  align-items: center;
  gap: 6px;
}

.nav-preview__apps {
  position: relative;
}

.nav-preview__brand {
  max-width: 96px;
  overflow: hidden;
  font-weight: 700;
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-preview__bar-nav {
  display: flex;
  flex: 1;
  flex-wrap: nowrap;
  align-items: center;
  gap: 4px;
  min-width: 0;
  /* 勿设 overflow，否则绝对定位下拉会被裁切 */
  overflow: visible;
}

.nav-preview__nav-item {
  position: relative;
  z-index: 1;
}

.nav-preview__nav-item:hover,
.nav-preview__apps:hover {
  z-index: 3;
}

.nav-preview__chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 120px;
  padding: 4px 8px;
  overflow: hidden;
  border-radius: 6px;
  background: var(--hover-color);
  color: var(--text-color-1);
  font-size: 12px;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: default;
}

.nav-preview__chip--icon {
  flex: none;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  cursor: pointer;
}

.nav-preview__flyout {
  position: absolute;
  top: 100%;
  left: 0;
  z-index: 5;
  min-width: 128px;
  /* pt 作悬停桥，避免移入下拉时闪断 */
  padding-top: 6px;
}

.nav-preview__flyout-inner {
  padding: 6px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--modal-color, var(--card-color));
  box-shadow: 0 4px 12px rgb(0 0 0 / 8%);
}

.nav-preview__flyout--links {
  width: 148px;
}

.nav-preview__flyout-label {
  margin: 0 0 4px;
  padding: 0 6px;
  color: var(--text-color-3);
  font-size: 11px;
}

.nav-preview__flyout-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  border-radius: 6px;
  font-size: 12px;
}

.nav-preview__flyout-item:hover {
  background: var(--hover-color);
}

.nav-preview__phone {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.nav-preview__phone-frame {
  width: 260px;
  max-width: 100%;
  max-height: 420px;
  overflow: hidden;
  border: 1px solid var(--border-color);
  border-radius: 16px;
  background: var(--card-color);
  box-shadow: 0 8px 24px rgb(0 0 0 / 6%);
}

.nav-preview__phone-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 12px 12px 8px;
  border-bottom: 1px solid var(--border-color);
}

.nav-preview__phone-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.nav-preview__phone-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  object-fit: cover;
}

.nav-preview__phone-name {
  overflow: hidden;
  font-weight: 700;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-preview__phone-body {
  max-height: 340px;
  padding: 8px;
  overflow-y: auto;
}

.nav-preview__phone-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  font-size: 13px;
}

.nav-preview__phone-row--btn {
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  text-align: left;
}

.nav-preview__phone-row--btn span {
  flex: 1;
}

.nav-preview__phone-row--child {
  padding-left: 22px;
  color: var(--text-color-2);
  font-size: 12px;
}

.nav-preview__phone-section {
  margin: 10px 10px 4px;
  color: var(--text-color-3);
  font-size: 11px;
}

.nav-preview__phone-group + .nav-preview__phone-row,
.nav-preview__phone-row + .nav-preview__phone-group {
  margin-top: 2px;
}
</style>
