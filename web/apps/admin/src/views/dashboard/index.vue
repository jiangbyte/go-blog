<!-- Author: go-blog — 博客运营台：内容库存 + 流量 + 草稿待办 + Top -->

<script setup lang="ts">
import type { Chart } from '@antv/g2'
import type { StatsOverview } from '@/api/sys/stats'
import { statsApi } from '@/api'
import G2Chart from '@/components/charts/G2Chart.vue'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils'
import {
  NButton,
  NCard,
  NEmpty,
  NFlex,
  NGi,
  NGrid,
  NRadioButton,
  NRadioGroup,
  NSkeleton,
  NStatistic,
  NTag,
} from 'naive-ui'
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const appStore = useAppStore()
const days = ref<7 | 14 | 30>(7)
const loading = ref(false)
const overview = ref<StatsOverview | null>(null)

/** 本站主题色（跟随外观设置） */
const themeColor = computed(() => appStore.primaryColor || '#1677FF')

async function load() {
  loading.value = true
  try {
    const res = await statsApi.overview({ days: days.value })
    overview.value = res.data
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(days, load)

const counts = computed(() => overview.value?.counts)
const period = computed(() => overview.value?.period)
const total = computed(() => overview.value?.total)
const drafts = computed(() => overview.value?.drafts ?? [])

const trafficFlat = computed(() => {
  const rows = overview.value?.traffic ?? []
  return rows.flatMap((r) => [
    { date: r.date, type: 'PV', value: r.site_pv },
    { date: r.date, type: 'UV', value: r.site_uv },
  ])
})

const articlePvSeries = computed(() =>
  (overview.value?.article.pv_series ?? []).map((d) => ({
    date: formatAxisDate(d.date),
    count: d.count,
  })),
)

const topBars = computed(() =>
  (overview.value?.article.top ?? []).map((t) => ({
    title: t.title.length > 16 ? `${t.title.slice(0, 16)}…` : t.title,
    view_count: t.view_count,
    uid: t.uid || t.id,
  })),
)

/** ISO 日期 → YY-MM-DD，对齐趋势图刻度样式 */
function formatAxisDate(iso: string): string {
  const parts = iso.split('-')
  if (parts.length !== 3) return iso
  return `${parts[0].slice(-2)}-${parts[1]}-${parts[2]}`
}

function trafficSpec(chart: Chart) {
  chart.data(trafficFlat.value)
  chart.encode('x', 'date')
  chart.encode('y', 'value')
  chart.encode('color', 'type')
  chart.scale('y', { nice: true })
  chart.axis('x', { title: false, labelAutoRotate: true })
  chart.axis('y', { title: false })
  chart.legend({ position: 'top' })
  chart.line().encode('shape', 'smooth').tooltip({ channel: 'y' })
  chart.point().encode('size', 2).tooltip(false)
}

function articlePvSpec(chart: Chart) {
  const color = themeColor.value
  chart.data(articlePvSeries.value)
  chart.encode('x', 'date')
  chart.encode('y', 'count')
  chart.scale('x', { range: [0, 1] })
  chart.scale('y', { nice: true, zero: true })
  chart.axis('x', { title: false, labelAutoRotate: false })
  chart.axis('y', {
    title: false,
    grid: true,
    gridStroke: '#e8e8e8',
    gridStrokeOpacity: 1,
    gridLineDash: [4, 4],
  })
  chart.legend(false)
  chart
    .area()
    .encode('shape', 'smooth')
    .style('fill', `l(270) 0:#ffffff 1:${color}`)
    .style('fillOpacity', 0.28)
    .tooltip(false)
  chart
    .line()
    .encode('shape', 'smooth')
    .style('stroke', color)
    .style('lineWidth', 2)
    .tooltip({ channel: 'y', name: '阅读' })
}

function topSpec(chart: Chart) {
  chart.coordinate({ transform: [{ type: 'transpose' }] })
  chart.data(topBars.value)
  chart.encode('x', 'title')
  chart.encode('y', 'view_count')
  chart.scale('y', { nice: true, zero: true })
  chart.axis('x', { title: false })
  chart.axis('y', { title: false })
  chart.legend(false)
  chart.interval().tooltip({ channel: 'y', name: '阅读' })
}

/** 内容库存卡片（无用户） */
const contentCards = computed(() => [
  { key: 'post_published', label: '已发布文章', value: counts.value?.post_published ?? 0 },
  { key: 'post_draft', label: '文章草稿', value: counts.value?.post_draft ?? 0 },
  { key: 'project_published', label: '已发布项目', value: counts.value?.project_published ?? 0 },
  { key: 'category', label: '分类', value: counts.value?.category ?? 0 },
  { key: 'tag', label: '标签', value: counts.value?.tag ?? 0 },
  { key: 'project_tag', label: '项目标签', value: counts.value?.project_tag ?? 0 },
])

/** 运营指标：近 N 天 + 累计，与内容库存同密度排布 */
const opsCards = computed(() => [
  { key: 'period_pv', label: `近 ${days.value} 天 PV`, value: period.value?.site_pv ?? 0 },
  { key: 'period_uv', label: `近 ${days.value} 天 UV`, value: period.value?.site_uv ?? 0 },
  { key: 'period_read', label: `近 ${days.value} 天阅读`, value: period.value?.article_pv ?? 0 },
  { key: 'total_pv', label: '累计 PV', value: total.value?.site_pv ?? 0 },
  { key: 'total_uv', label: '累计 UV', value: total.value?.site_uv ?? 0 },
  { key: 'total_read', label: '累计文章阅读', value: total.value?.article_pv ?? 0 },
])

function goPosts(status?: string) {
  void router.push({ path: '/blog/posts', query: status ? { status } : undefined })
}

function goCreatePost() {
  void router.push('/blog/posts/create')
}

function goProjects() {
  void router.push('/blog/projects')
}

function goEditPost(uid: string) {
  void router.push({ path: '/blog/posts/edit', query: { uid } })
}
</script>

<template>
  <div class="h-full min-h-0 overflow-auto p-8px flex flex-col gap-12px">
    <div class="flex flex-wrap items-center justify-between gap-8px">
      <div class="text-16px font-medium">博客运营台</div>
      <NFlex align="center" :size="8" :wrap="true">
        <NButton size="small" type="primary" secondary @click="goCreatePost">写文章</NButton>
        <NButton size="small" secondary @click="goPosts('draft')">草稿箱</NButton>
        <NButton size="small" secondary @click="goPosts('published')">已发布</NButton>
        <NButton size="small" secondary @click="goProjects">项目</NButton>
        <n-radio-group v-model:value="days" size="small">
          <n-radio-button :value="7">近 7 天</n-radio-button>
          <n-radio-button :value="14">近 14 天</n-radio-button>
          <n-radio-button :value="30">近 30 天</n-radio-button>
        </n-radio-group>
      </NFlex>
    </div>

    <!-- 左：内容库存 + 运营数据；右：待处理草稿 -->
    <n-grid cols="1 m:4" responsive="screen" :x-gap="12" :y-gap="12" item-responsive>
      <n-gi span="4 m:3">
        <div class="flex flex-col gap-12px h-full">
          <div>
            <div class="text-13px font-medium text-[var(--n-text-color-2)] mb-8px">内容库存</div>
            <n-skeleton v-if="loading && !overview" height="88px" :sharp="false" />
            <n-grid v-else cols="2 s:3 m:6" responsive="screen" :x-gap="12" :y-gap="12">
              <n-gi v-for="card in contentCards" :key="card.key">
                <n-card size="small" :bordered="false">
                  <n-statistic :label="card.label" :value="card.value" />
                </n-card>
              </n-gi>
            </n-grid>
          </div>
          <div>
            <div class="text-13px font-medium text-[var(--n-text-color-2)] mb-8px">运营数据</div>
            <n-grid cols="2 s:3 m:6" responsive="screen" :x-gap="12" :y-gap="12">
              <n-gi v-for="card in opsCards" :key="card.key">
                <n-card size="small" :bordered="false">
                  <n-statistic :label="card.label" :value="card.value" />
                </n-card>
              </n-gi>
            </n-grid>
          </div>
        </div>
      </n-gi>
      <n-gi span="4 m:1">
        <n-card size="small" :bordered="false" title="待处理草稿" class="h-full drafts-card">
          <n-empty v-if="!drafts.length" description="暂无草稿，去写一篇吧" class="py-16px">
            <template #extra>
              <NButton size="small" type="primary" @click="goCreatePost">写文章</NButton>
            </template>
          </n-empty>
          <div v-else class="flex flex-col gap-8px">
            <button
              v-for="item in drafts"
              :key="item.uid"
              type="button"
              class="draft-row"
              @click="goEditPost(item.uid)"
            >
              <div class="draft-row__title">{{ item.title || '未命名草稿' }}</div>
              <div class="draft-row__meta">
                <NTag size="tiny" :bordered="false">草稿</NTag>
                <span>{{ formatDateTime(item.updated_at) }}</span>
              </div>
            </button>
            <NButton text type="primary" size="small" class="self-start" @click="goPosts('draft')">
              查看全部草稿
            </NButton>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 图表同一行：全站流量 / 阅读趋势 / 阅读 Top -->
    <n-grid cols="1 m:3" responsive="screen" :x-gap="12" :y-gap="12">
      <n-gi>
        <n-card size="small" :bordered="false" title="全站流量">
          <G2Chart
            v-if="trafficFlat.length"
            :height="240"
            :deps="trafficFlat"
            :spec="trafficSpec"
          />
          <n-empty v-else description="暂无流量数据" class="py-24px" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small" :bordered="false" title="阅读趋势">
          <G2Chart
            v-if="articlePvSeries.length"
            :height="240"
            :deps="[articlePvSeries, themeColor]"
            :spec="articlePvSpec"
          />
          <n-empty v-else description="暂无阅读数据" class="py-24px" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small" :bordered="false" title="阅读 Top">
          <G2Chart v-if="topBars.length" :height="240" :deps="topBars" :spec="topSpec" />
          <n-empty v-else description="暂无阅读排行" class="py-24px" />
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<style scoped>
.drafts-card :deep(.n-card__content) {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 160px;
}

.draft-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  background: var(--n-color);
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.15s,
    background 0.15s;
}

.draft-row:hover {
  border-color: var(--primary-color, #18a058);
  background: var(--n-color-hover, #f7f9fa);
}

.draft-row__title {
  font-size: 13px;
  font-weight: 500;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.draft-row__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--n-text-color-3);
}
</style>
