<!-- Author: go-blog — 博客关于页（整页一份 Markdown；左编右预） -->

<script setup lang="ts">
import { MdSplitWorkspace } from '@/components/editor'
import { blogAboutApi } from '@/api'
import { hasPermission } from '@/utils'
import { ProCard } from 'pro-naive-ui'
import { onMounted, reactive } from 'vue'

const state = reactive({
  loading: false,
  saving: false,
  title: '关于',
  content: '',
})

function fillFromDoc(doc: Record<string, any>) {
  state.title = typeof doc?.title === 'string' && doc.title.trim() ? doc.title.trim() : '关于'
  state.content = typeof doc?.content === 'string' ? doc.content : ''
}

function buildPayload() {
  return {
    title: state.title.trim() || '关于',
    content: state.content,
  }
}

async function reload() {
  state.loading = true
  try {
    const res = await blogAboutApi.getAbout()
    fillFromDoc(res.data || {})
  } finally {
    state.loading = false
  }
}

async function save() {
  state.saving = true
  try {
    const res = await blogAboutApi.updateAbout(buildPayload())
    fillFromDoc(res.data || buildPayload())
    window.$message.success('已保存')
  } finally {
    state.saving = false
  }
}

onMounted(() => {
  void reload()
})
</script>

<template>
  <div class="min-h-full">
    <ProCard title="关于页" :show-collapse="false" :trigger-areas="[]">
      <NSpin :show="state.loading">
        <NForm label-placement="top" :disabled="state.saving">
          <NFormItem label="正文（Markdown）">
            <MdSplitWorkspace
              v-model:value="state.content"
              :height="560"
              placeholder="关于页全文 Markdown..."
              :disabled="state.loading || state.saving"
            />
          </NFormItem>
        </NForm>

        <NFlex class="mt-4" justify="end">
          <NButton :loading="state.loading" :disabled="state.saving" @click="reload">重新加载</NButton>
          <NButton
            v-if="hasPermission('admin:blog:about:update')"
            type="primary"
            :loading="state.saving"
            :disabled="state.loading"
            @click="save"
          >
            保存
          </NButton>
        </NFlex>
      </NSpin>
    </ProCard>
  </div>
</template>
