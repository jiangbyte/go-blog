<!-- Author: go-blog — Admin 内简历全屏预览（后端 HTML iframe，不跳 Portal） -->

<script setup lang="ts">
import { Icon } from '@iconify/vue/offline'
import { resumeApi } from '@/api'
import {
  NButton,
  NEmpty,
  NFlex,
  NIcon,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  NSpin,
  NText,
} from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ResumeBackendPreview from './ResumeBackendPreview.vue'
import { downloadResumeDocument } from './resumePrint'

const route = useRoute()
const router = useRouter()
const resumeId = computed(() => String(route.query.id || ''))

const loading = ref(false)
const downloading = ref(false)
const doc = ref<Record<string, any> | null>(null)

/** 用管理端 detail 拉完整纸面数据。 */
async function load() {
  if (!resumeId.value) {
    window.$message.error('缺少简历 ID')
    return
  }
  loading.value = true
  try {
    const res = await resumeApi.detail(resumeId.value)
    doc.value = res.data || null
  } finally {
    loading.value = false
  }
}

function goBack() {
  void router.push('/resume/list')
}

function goEdit() {
  void router.push({ path: '/resume/edit', query: { id: resumeId.value } })
}

/** 下载与纸面一致的 PDF（export-pdf）。 */
async function onDownload() {
  if (!resumeId.value || downloading.value) return
  downloading.value = true
  try {
    await downloadResumeDocument(
      resumeId.value,
      doc.value?.profile_json?.name || doc.value?.title || '简历',
    )
    window.$message.success('简历已下载')
  } catch (err: any) {
    window.$message.error(err?.message || '下载失败')
  } finally {
    downloading.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <NLayout class="resume-preview-page" style="height: 100vh; overflow: hidden">
    <NLayoutHeader
      bordered
      position="absolute"
      class="h-52px px-3"
      style="left: 0; right: 0; top: 0; z-index: 10"
    >
      <NFlex class="h-full" align="center" justify="space-between" :size="12">
        <NFlex align="center" :size="10">
          <NButton quaternary size="small" @click="goBack">
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:left" />
              </NIcon>
            </template>
            返回列表
          </NButton>
          <NText strong>
            {{ doc?.title || '简历预览' }}
          </NText>
        </NFlex>
        <NFlex :size="8">
          <NButton size="small" quaternary :loading="downloading" @click="onDownload">
            <template #icon>
              <NIcon>
                <Icon icon="lucide:download" />
              </NIcon>
            </template>
            下载
          </NButton>
          <NButton size="small" type="primary" @click="goEdit"> 编辑 </NButton>
        </NFlex>
      </NFlex>
    </NLayoutHeader>

    <NLayoutContent
      position="absolute"
      :native-scrollbar="false"
      style="top: 52px; bottom: 0; left: 0; right: 0"
      content-style="padding: 24px 16px 48px; background: var(--body-color); height: 100%; box-sizing: border-box"
    >
      <NSpin :show="loading">
        <ResumeBackendPreview v-if="doc" :doc="doc" :debounce-ms="0" />
        <NEmpty v-else-if="!loading" description="简历不存在" class="py-24" />
      </NSpin>
    </NLayoutContent>
  </NLayout>
</template>
