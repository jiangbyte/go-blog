<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getAbout } from '@/api'
import AboutPage from '@/components/about/AboutPage.vue'

const doc = ref<any>(null)

onMounted(async () => {
  const res = await getAbout({})
  doc.value = res.doc
})
</script>

<template>
  <AboutPage v-if="doc" :doc="doc" />
</template>
