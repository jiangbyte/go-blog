<!-- Author: go-blog — 外观：模板 / 文字 / 样式 / 间距（NTabs） -->

<script setup lang="ts">
import {
  NButton,
  NCollapse,
  NCollapseItem,
  NFlex,
  NGi,
  NGrid,
  NInput,
  NSlider,
  NTabPane,
  NTabs,
  NText,
} from 'naive-ui'
import { themeTokens, RESUME_TEMPLATE_MANIFEST, type ResumeTemplateMeta } from '@jiangbyte/shared'
import { computed, ref } from 'vue'

type ThemeModel = {
  themeColor: string
  layout: string
  infoDisplay: string
  language: string
  fontFamily: string
  fontSizeScale: number
  spacing: {
    moduleGap: number
    lineHeight: number
    padX: number
    padTop: number
    padBottom: number
  }
}

const theme = defineModel<ThemeModel>('theme', { required: true })
const customCss = defineModel<string>('customCss', { required: true })
const templateId = defineModel<string>('templateId', { required: true })

const props = defineProps<{
  canWrite: boolean
  templates: ResumeTemplateMeta[]
}>()

const activeTab = ref('template')

const COLOR_SWATCHES = Object.entries(themeTokens.colors).map(([value, hex]) => ({
  value,
  hex: String(hex),
}))

const FONT_LABELS: Record<string, string> = {
  default: '默认黑体',
  'source-sans': '思源黑体',
  'source-serif': '思源宋体',
  wenkai: '霞鹜文楷',
  'inter-sans': 'Inter + 思源黑体',
}
const FONT_OPTIONS = Object.keys(themeTokens.fontStacks).map((value) => ({
  value,
  label: FONT_LABELS[value] || value,
}))

const LAYOUT_OPTIONS = (themeTokens.layouts || ['left', 'center', 'tiled']).map((value) => ({
  value,
  label:
    ({ left: '居左', center: '居中', tiled: '平铺' } as Record<string, string>)[value] || value,
}))

const INFO_OPTIONS = (themeTokens.infoDisplays || ['text', 'icon', 'plain']).map((value) => ({
  value,
  label:
    ({ text: '文字', icon: '图标', plain: '纯内容' } as Record<string, string>)[value] || value,
}))

const DEFAULT_SPACING = {
  moduleGap: 12,
  lineHeight: 1.55,
  padX: 48,
  padTop: 48,
  padBottom: 60,
}

/** 模板列表只信 API；空列表时用 shared manifest，避免再硬编码一份。 */
const templateCards = computed(() => {
  const fromApi = props.templates || []
  if (fromApi.length) return fromApi
  return RESUME_TEMPLATE_MANIFEST
})

const currentTemplate = computed(
  () => templateCards.value.find((t) => t.id === templateId.value) || templateCards.value[0],
)

/** 缩略图主题色：预设名 → hex，供通用 thumb 使用。 */
function thumbColor(t: ResumeTemplateMeta): string {
  const key = (t.defaultColor || 'blue').trim()
  const presets = themeTokens.colors as Record<string, string>
  if (presets[key]) return String(presets[key])
  if (/^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/.test(key)) return key
  return String(presets.blue || '#3b82f6')
}

const fontScalePercent = computed({
  get: () => Math.round((theme.value.fontSizeScale || 1) * 100),
  set: (v: number) => {
    theme.value.fontSizeScale = Math.round(v) / 100
  },
})

/** 选择模板并套用默认主题色。 */
function selectTemplate(t: ResumeTemplateMeta) {
  if (!props.canWrite) return
  templateId.value = t.id
  if (t.defaultColor) {
    theme.value.themeColor = t.defaultColor
  }
}

function setFont(id: string) {
  if (!props.canWrite) return
  theme.value.fontFamily = id
}

function setFontPreset(scale: number) {
  if (!props.canWrite) return
  theme.value.fontSizeScale = scale
}

function setSegment(key: 'layout' | 'infoDisplay', value: string) {
  if (!props.canWrite) return
  theme.value[key] = value
}

function setColor(value: string) {
  if (!props.canWrite) return
  theme.value.themeColor = value
}

function resetSpacing() {
  if (!props.canWrite) return
  Object.assign(theme.value.spacing, { ...DEFAULT_SPACING })
}
</script>

<template>
  <div class="theme-panel">
    <NTabs v-model:value="activeTab" type="line" size="small" animated>
      <!-- 模板 -->
      <NTabPane name="template" tab="模板">
        <div class="theme-panel__grid">
          <button
            v-for="t in templateCards"
            :key="t.id"
            type="button"
            class="tpl-card"
            :class="{ 'is-active': templateId === t.id }"
            :disabled="!canWrite"
            @click="selectTemplate(t)"
          >
            <div class="tpl-card__thumb" :style="{ '--tpl-accent': thumbColor(t) }">
              <span class="tpl-card__bar" />
              <span class="tpl-card__line" />
              <span class="tpl-card__line tpl-card__line--short" />
              <span class="tpl-card__line" />
            </div>
            <span class="tpl-card__name">{{ t.name }}</span>
          </button>
        </div>
      </NTabPane>

      <!-- 文字 -->
      <NTabPane name="text" tab="文字">
        <NFlex vertical :size="16">
          <div>
            <NText depth="3" class="theme-panel__label"> 字体 </NText>
            <NGrid :cols="2" :x-gap="8" :y-gap="8">
              <NGi v-for="f in FONT_OPTIONS" :key="f.value">
                <NButton
                  block
                  size="small"
                  :type="theme.fontFamily === f.value ? 'primary' : 'default'"
                  :ghost="theme.fontFamily !== f.value"
                  :disabled="!canWrite"
                  @click="setFont(f.value)"
                >
                  {{ f.label }}
                </NButton>
              </NGi>
            </NGrid>
          </div>

          <div>
            <NFlex justify="space-between" align="center" class="mb-1">
              <NText depth="3" class="theme-panel__label"> 字号 </NText>
              <NText type="primary" strong> {{ fontScalePercent }}% </NText>
            </NFlex>
            <NSlider
              v-model:value="fontScalePercent"
              :min="80"
              :max="120"
              :step="1"
              :disabled="!canWrite"
            />
            <NFlex :size="8" class="mt-2">
              <NButton size="tiny" :disabled="!canWrite" @click="setFontPreset(0.85)"> 小 </NButton>
              <NButton size="tiny" :disabled="!canWrite" @click="setFontPreset(1)"> 标准 </NButton>
              <NButton size="tiny" :disabled="!canWrite" @click="setFontPreset(1.15)"> 大 </NButton>
            </NFlex>
          </div>
        </NFlex>
      </NTabPane>

      <!-- 样式 -->
      <NTabPane name="style" tab="样式">
        <NFlex vertical :size="16">
          <div>
            <NText depth="3" class="theme-panel__label"> 主题色 </NText>
            <NFlex :size="10" class="theme-panel__swatches">
              <button
                v-for="c in COLOR_SWATCHES"
                :key="c.value"
                type="button"
                class="color-dot"
                :class="{ 'is-active': theme.themeColor === c.value }"
                :style="{ background: c.hex }"
                :disabled="!canWrite"
                :aria-label="c.value"
                @click="setColor(c.value)"
              />
            </NFlex>
          </div>

          <div>
            <NText depth="3" class="theme-panel__label"> 布局调整 </NText>
            <NFlex :size="8">
              <NButton
                v-for="opt in LAYOUT_OPTIONS"
                :key="opt.value"
                size="small"
                :type="theme.layout === opt.value ? 'primary' : 'default'"
                :ghost="theme.layout !== opt.value"
                :disabled="!canWrite"
                @click="setSegment('layout', opt.value)"
              >
                {{ opt.label }}
              </NButton>
            </NFlex>
          </div>

          <div>
            <NText depth="3" class="theme-panel__label"> 信息展示 </NText>
            <NFlex :size="8">
              <NButton
                v-for="opt in INFO_OPTIONS"
                :key="opt.value"
                size="small"
                :type="theme.infoDisplay === opt.value ? 'primary' : 'default'"
                :ghost="theme.infoDisplay !== opt.value"
                :disabled="!canWrite"
                @click="setSegment('infoDisplay', opt.value)"
              >
                {{ opt.label }}
              </NButton>
            </NFlex>
          </div>

          <div v-if="currentTemplate" class="theme-panel__tpl-info">
            <NText depth="3" class="theme-panel__label"> 当前模板 </NText>
            <div class="theme-panel__tpl-card">
              <NText strong>
                {{ currentTemplate.name }}
              </NText>
              <NText depth="3" class="block mt-1">
                {{ currentTemplate.description }}
              </NText>
            </div>
          </div>

          <NCollapse>
            <NCollapseItem title="自定义 CSS" name="css">
              <NInput
                v-model:value="customCss"
                type="textarea"
                :disabled="!canWrite"
                :rows="8"
                placeholder=".resume-name { letter-spacing: 0.12em; }"
              />
            </NCollapseItem>
          </NCollapse>
        </NFlex>
      </NTabPane>

      <!-- 间距 -->
      <NTabPane name="spacing" tab="间距">
        <NFlex vertical :size="18">
          <div>
            <NFlex justify="space-between" class="mb-1">
              <NText depth="3"> 模块上下间距 </NText>
              <NText type="primary" strong>
                {{ theme.spacing.moduleGap }}
              </NText>
            </NFlex>
            <NSlider
              v-model:value="theme.spacing.moduleGap"
              :min="6"
              :max="24"
              :step="1"
              :disabled="!canWrite"
            />
            <NFlex justify="space-between" class="theme-panel__hint">
              <span>紧凑</span><span>宽松</span>
            </NFlex>
          </div>

          <div>
            <NFlex justify="space-between" class="mb-1">
              <NText depth="3"> 行间距 </NText>
              <NText type="primary" strong>
                {{ Number(theme.spacing.lineHeight).toFixed(2) }}
              </NText>
            </NFlex>
            <NSlider
              v-model:value="theme.spacing.lineHeight"
              :min="1.1"
              :max="2"
              :step="0.05"
              :disabled="!canWrite"
            />
            <NFlex justify="space-between" class="theme-panel__hint">
              <span>紧凑</span><span>宽松</span>
            </NFlex>
          </div>

          <div>
            <NFlex justify="space-between" class="mb-1">
              <NText depth="3"> 左右边距 </NText>
              <NText type="primary" strong>
                {{ theme.spacing.padX }}
              </NText>
            </NFlex>
            <NSlider
              v-model:value="theme.spacing.padX"
              :min="28"
              :max="72"
              :step="2"
              :disabled="!canWrite"
            />
            <NFlex justify="space-between" class="theme-panel__hint">
              <span>窄</span><span>宽</span>
            </NFlex>
          </div>

          <div>
            <NFlex justify="space-between" class="mb-1">
              <NText depth="3"> 上边距 </NText>
              <NText type="primary" strong>
                {{ theme.spacing.padTop }}
              </NText>
            </NFlex>
            <NSlider
              v-model:value="theme.spacing.padTop"
              :min="28"
              :max="72"
              :step="2"
              :disabled="!canWrite"
            />
            <NFlex justify="space-between" class="theme-panel__hint">
              <span>窄</span><span>宽</span>
            </NFlex>
          </div>

          <div>
            <NFlex justify="space-between" class="mb-1">
              <NText depth="3"> 下边距 </NText>
              <NText type="primary" strong>
                {{ theme.spacing.padBottom }}
              </NText>
            </NFlex>
            <NSlider
              v-model:value="theme.spacing.padBottom"
              :min="32"
              :max="90"
              :step="2"
              :disabled="!canWrite"
            />
            <NFlex justify="space-between" class="theme-panel__hint">
              <span>窄</span><span>宽</span>
            </NFlex>
          </div>

          <NButton block :disabled="!canWrite" @click="resetSpacing"> 恢复默认间距 </NButton>
        </NFlex>
      </NTabPane>
    </NTabs>
  </div>
</template>

<style scoped>
.theme-panel__label {
  display: block;
  margin-bottom: 8px;
  font-size: 12px;
}

.theme-panel__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  /* 给选中描边留出空间，避免贴边被滚动容器裁切 */
  padding: 2px;
}

.tpl-card {
  appearance: none;
  border: 2px solid var(--border-color);
  border-radius: 8px;
  background: var(--card-color, #fff);
  padding: 6px;
  cursor: pointer;
  text-align: left;
  min-width: 0;
  transition: border-color 0.15s;
}

.tpl-card:disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.tpl-card.is-active {
  border-color: var(--primary-color, #3b82f6);
  box-shadow: none;
}

.tpl-card__thumb {
  height: 72px;
  border-radius: 4px;
  background: #f3f4f6;
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  gap: 5px;
  margin-bottom: 6px;
  overflow: hidden;
  position: relative;
}

.tpl-card__bar {
  height: 4px;
  width: 40%;
  border-radius: 2px;
  background: var(--tpl-accent, var(--primary-color, #3b82f6));
}

.tpl-card__line {
  height: 3px;
  width: 100%;
  border-radius: 1px;
  background: #d1d5db;
}

.tpl-card__line--short {
  width: 70%;
}

.tpl-card__name {
  font-size: 12px;
  font-weight: 600;
}

.tpl-card.is-active .tpl-card__name {
  color: var(--primary-color, #3b82f6);
}

.color-dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid transparent;
  box-shadow: inset 0 0 0 2px #fff;
  cursor: pointer;
  padding: 0;
}

.color-dot.is-active {
  outline: 2px solid var(--primary-color, #3b82f6);
  outline-offset: 2px;
}

.color-dot:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.theme-panel__tpl-card {
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--app-muted-bg, #f5f7fb);
}

.theme-panel__hint {
  font-size: 11px;
  color: var(--text-color-3, #999);
  margin-top: 2px;
}

.mb-1 {
  margin-bottom: 4px;
}

.mt-1 {
  margin-top: 4px;
}

.mt-2 {
  margin-top: 8px;
}

.block {
  display: block;
}
</style>
