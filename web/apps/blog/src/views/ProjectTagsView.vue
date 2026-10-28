<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listProjectTags } from '@/api'
import TagCloudPage from '@/components/taxonomy/TagCloudPage.vue'

const terms = ref<any[]>([])

onMounted(async () => {
  const { list } = await listProjectTags({})
  terms.value = list
})
</script>

<template>
  <TagCloudPage title="项目标签" desc="按标签浏览项目" :terms="terms" base-path="/project-tags/" />
</template>
