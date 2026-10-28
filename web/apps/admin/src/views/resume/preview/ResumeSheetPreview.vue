<!-- Author: go-blog — Admin 简历纸面预览（sheet-v1 活 DOM，与 Portal 同源样式） -->

<script setup lang="ts">
import { Icon } from '@iconify/vue/offline'
import {
  balanceProfileRows,
  fontStylesheetUrl,
  resolveContactHref,
  resolveResumeIcon,
  themeCssVars,
} from '@jiangbyte/shared'
import ResumeHtmlBody from '../edit/ResumeHtmlBody.vue'
import { computed, onBeforeUnmount, watch } from 'vue'
import '@jiangbyte/shared/resume/styles'
import './admin-shell.css'

export type ResumePreviewDoc = {
  id?: string
  template_id?: string
  theme_json?: Record<string, any>
  custom_css?: string
  profile_json?: {
    name?: string
    tagline?: string
    avatar_url?: string
    github_url?: string
    fields?: { key: string; label: string; value: string; icon?: string }[]
  }
  sections?: {
    id: string
    type: string
    title: string
    icon?: string
    sort_code?: string | number
    meta_json?: Record<string, unknown>
    body_html?: string
    entries_json?: {
      title?: string
      role?: string
      period?: string
      body_html?: string
      tags?: string[]
    }[]
  }[]
}

const props = withDefaults(
  defineProps<{
    doc: ResumePreviewDoc
    /** 内嵌模式：去掉全页背景，供工作台/缩略图使用 */
    embedded?: boolean
  }>(),
  { embedded: true },
)

const theme = computed(() => props.doc?.theme_json || {})
const layout = computed(() => theme.value?.layout || 'left')
const templateId = computed(() => props.doc?.template_id || 'classic')
const profile = computed(
  () =>
    props.doc?.profile_json || {
      name: '',
      tagline: '',
      avatar_url: '',
      github_url: '',
      fields: [] as { key: string; label: string; value: string }[],
    },
)
const sections = computed(() =>
  [...(props.doc?.sections || [])]
    .filter((s) => s.meta_json?.visible !== false)
    .sort((a, b) => Number(a.sort_code || 0) - Number(b.sort_code || 0)),
)
const cssVars = computed(() => themeCssVars(theme.value))
const infoDisplay = computed(() => theme.value?.infoDisplay || 'text')
const pageId = computed(() => props.doc?.id || 'preview')
const customCss = computed(() => (props.doc?.custom_css || '').trim())
const injectedCss = computed(() => {
  if (!customCss.value) return ''
  return `.resume-page[data-resume="${pageId.value}"] { }\n${customCss.value}`
})

/** 按需挂载字体样式表。 */
let fontLinkEl: HTMLLinkElement | null = null
watch(
  () => theme.value?.fontFamily,
  (family) => {
    const href = fontStylesheetUrl(family)
    if (fontLinkEl) {
      fontLinkEl.remove()
      fontLinkEl = null
    }
    if (!href || typeof document === 'undefined') return
    const link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = href
    link.dataset.resumeFont = '1'
    document.head.appendChild(link)
    fontLinkEl = link
  },
  { immediate: true },
)
onBeforeUnmount(() => {
  fontLinkEl?.remove()
  fontLinkEl = null
})

const contactRows = computed(() => {
  const fields = (profile.value.fields || []).filter((f) => (f.value || '').trim())
  return balanceProfileRows(fields, String(infoDisplay.value || 'text'))
})
</script>

<template>
  <div
    class="resume-page"
    :class="[
      `resume-layout--${layout}`,
      `resume-template--${templateId}`,
      { 'resume-page--embedded': embedded },
    ]"
    :data-resume="pageId"
    :style="cssVars"
  >
    <component :is="'style'" v-if="injectedCss">{{ injectedCss }}</component>

    <article class="resume-sheet">
      <div class="resume-body">
        <header class="resume-header">
          <div class="resume-header__main">
            <h1 class="resume-name">
              {{ profile.name || '未命名' }}
            </h1>
            <div
              v-if="contactRows.length"
              class="resume-contacts"
              :class="{
                'resume-contacts--plain': infoDisplay === 'plain',
                'resume-contacts--icon': infoDisplay === 'icon',
              }"
            >
              <div v-for="(row, ri) in contactRows" :key="ri" class="resume-contacts__row">
                <template v-for="(field, fi) in row" :key="field.key">
                  <span v-if="fi > 0" class="sep">|</span>
                  <span class="resume-contacts__item">
                    <a
                      v-if="resolveContactHref(field.value)"
                      :href="resolveContactHref(field.value)"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <Icon
                        v-if="infoDisplay === 'icon'"
                        class="resume-contacts__icon"
                        :icon="resolveResumeIcon(field.icon)"
                        width="13"
                        height="13"
                      />
                      <span v-if="infoDisplay === 'text'" class="label">{{ field.label }}：</span>
                      <span>{{ field.value }}</span>
                    </a>
                    <template v-else>
                      <Icon
                        v-if="infoDisplay === 'icon'"
                        class="resume-contacts__icon"
                        :icon="resolveResumeIcon(field.icon)"
                        width="13"
                        height="13"
                      />
                      <span v-if="infoDisplay === 'text'" class="label">{{ field.label }}：</span>
                      <span>{{ field.value }}</span>
                    </template>
                  </span>
                </template>
              </div>
            </div>
          </div>
          <img
            v-if="profile.avatar_url"
            class="resume-avatar"
            :src="profile.avatar_url"
            alt=""
            width="92"
            height="118"
          />
        </header>

        <section v-for="section in sections" :key="section.id" class="resume-section">
          <div class="section-header">
            <span class="section-header__icon" aria-hidden="true">
              <Icon :icon="resolveResumeIcon(section.icon)" width="11" height="11" />
            </span>
            <h2 class="section-header__title">
              {{ section.title }}
            </h2>
            <span class="section-header__line" aria-hidden="true" />
          </div>

          <div class="resume-section__content resume-md">
            <div
              v-if="
                section.type === 'education' &&
                (section.meta_json?.school || section.meta_json?.major || section.meta_json?.period)
              "
              class="entry-header"
            >
              <span class="entry-header__title">{{ section.meta_json?.school || '' }}</span>
              <span v-if="section.meta_json?.major" class="entry-header__role">{{
                section.meta_json?.major
              }}</span>
              <span v-if="section.meta_json?.period" class="entry-header__period">{{
                section.meta_json?.period
              }}</span>
            </div>

            <div
              v-for="(entry, idx) in section.entries_json || []"
              :key="`${section.id}-${idx}`"
              class="resume-entry"
            >
              <div class="entry-header">
                <span class="entry-header__title">{{ entry.title }}</span>
                <span v-if="entry.role" class="entry-header__role">{{ entry.role }}</span>
                <span v-if="entry.period" class="entry-header__period">{{ entry.period }}</span>
              </div>
              <div
                v-if="entry.tags?.length"
                class="resume-entry__tags"
                style="margin-bottom: 4px; opacity: 0.75"
              >
                {{ entry.tags.join(' · ') }}
              </div>
              <div v-if="entry.body_html?.trim()" class="resume-entry__body">
                <ResumeHtmlBody :value="entry.body_html" />
              </div>
            </div>

            <ResumeHtmlBody v-if="section.body_html?.trim()" :value="section.body_html" />
          </div>
        </section>
      </div>
    </article>
  </div>
</template>
