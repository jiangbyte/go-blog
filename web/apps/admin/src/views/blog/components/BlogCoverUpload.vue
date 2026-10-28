<!-- Author: go-blog — 博客文章/项目封面：仅绑定自定义封面，空=未上传 -->

<script setup lang="ts">
import ImageUpload from '@/components/upload/ImageUpload.vue'
import { blogSiteApi } from '@/api'
import { computed, onMounted, ref } from 'vue'

const props = defineProps<{
  value?: string | null
}>()

const emit = defineEmits<{
  'update:value': [value: string]
}>()

/** 站点是否配置了默认封面池（仅用于提示文案，不参与上传回显） */
const hasDefaultCovers = ref(false)

const coverModel = computed({
  get: () => String(props.value || '').trim(),
  set: (next: string) => {
    emit('update:value', String(next || '').trim())
  },
})

const hint = computed(() =>
  hasDefaultCovers.value
    ? '未上传时前台使用站点默认封面；清除后保持为空'
    : '未上传时前台将使用站点默认封面（当前站点未配置 default_covers）',
)

onMounted(async () => {
  try {
    const res = await blogSiteApi.getSite()
    const raw = res.data?.default_covers
    hasDefaultCovers.value = Array.isArray(raw) && raw.some((item) => String(item || '').trim())
  } catch {
    // 提示失败不影响上传
  }
})
</script>

<template>
  <NFlex vertical :size="6">
    <ImageUpload
      v-model:value="coverModel"
      variant="card"
      value-type="object_name"
      button-text="上传封面"
    />
    <NText depth="3" style="font-size: 12px">
      {{ hint }}
    </NText>
  </NFlex>
</template>
