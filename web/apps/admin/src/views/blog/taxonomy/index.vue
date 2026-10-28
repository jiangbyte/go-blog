<!-- Author: go-blog — 旧「分类标签」合并页已拆分，重定向到分类 -->

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
onMounted(() => {
  void router.replace('/blog/categories')
})
</script>

<template>
  <div />
</template>
