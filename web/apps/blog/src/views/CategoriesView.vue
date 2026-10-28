<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listCategories } from '@/api'
import TagCloudPage from '@/components/taxonomy/TagCloudPage.vue'

const terms = ref<any[]>([])

onMounted(async () => {
  const { list } = await listCategories({})
  terms.value = list
})
</script>

<template>
  <TagCloudPage title="分类" desc="按分类浏览文章" :terms="terms" base-path="/categories/" />
</template>
