<script setup lang="ts">

defineProps<{
  title: string
  desc?: string
  terms: any[]
  basePath: string
}>()
</script>

<template>
  <section class="glass-card p-6">
    <h1 class="text-xl font-bold">{{ title }}</h1>
    <p v-if="desc" class="page-desc mt-1 text-center text-sm text-muted md:text-left">{{ desc }}</p>
    <div class="mt-5 flex flex-wrap gap-2">
      <RouterLink
        v-for="t in terms"
        :key="t.uid"
        :to="`${basePath}${t.uid}`"
        class="milk-hover inline-flex items-center gap-1.5 rounded-full border border-card-border bg-card px-3 py-1.5 text-sm text-ink"
      >
        <span>{{ t.name }}</span>
        <sup class="text-xs text-muted">{{ t.count }}</sup>
      </RouterLink>
    </div>
  </section>
</template>
