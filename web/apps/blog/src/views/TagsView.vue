<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listTags } from '@/api'
import TagCloudPage from '@/components/taxonomy/TagCloudPage.vue'

const terms = ref<any[]>([])

onMounted(async () => {
  const { list } = await listTags({})
  terms.value = list
})
</script>

<template>
  <TagCloudPage title="标签" desc="按标签浏览文章" :terms="terms" base-path="/tags/" />
</template>
