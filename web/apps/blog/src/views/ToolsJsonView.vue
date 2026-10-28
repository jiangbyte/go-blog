<script setup lang="ts">
import JsonFormatter from '@/components/tools/JsonFormatter.vue'
</script>

<template>
  <div>
    <div class="mb-4">
      <h1 class="text-2xl font-bold">JSON 格式化</h1>
      <p class="mt-1 text-sm text-muted">本地解析与格式化，数据不会上传服务器</p>
    </div>
    <JsonFormatter />
  </div>
</template>
