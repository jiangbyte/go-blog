<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { getHomePosts, getSite } from '@/api'
import { coverOf } from '@/utils/markdown'

const site = getSite()
const items = ref<any[]>([])
const active = ref(0)
let timer: number | undefined

const current = computed(() => items.value[active.value])

onMounted(async () => {
  const { list } = await getHomePosts({})
  items.value = list.slice(0, 5)
  if (!items.value.length) return
  timer = window.setInterval(() => {
    active.value = (active.value + 1) % items.value.length
  }, 5000)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <section v-if="items.length" class="glass-card relative mb-5 h-[280px] overflow-hidden md:h-[320px]">
    <div class="absolute inset-0">
      <div
        v-for="(item, i) in items"
        :key="item.uid"
        class="absolute inset-0 transition-opacity duration-500"
        :class="i === active ? 'opacity-100' : 'opacity-0'"
      >
        <img
          :src="coverOf(item.cover, site.default_covers, i)"
          alt=""
          class="h-full w-full object-cover"
          :loading="i === 0 ? 'eager' : 'lazy'"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-black/55 via-black/15 to-transparent" />
      </div>
    </div>
    <div class="relative z-10 flex h-full flex-col justify-end p-5 text-white md:p-7">
      <div class="mb-2 flex flex-wrap items-center gap-2">
        <span
          v-if="current?.sticky"
          class="rounded-full bg-gradient-to-br from-[#e85d4c] to-[#d4483a] px-2.5 py-0.5 text-xs font-bold shadow-[0_2px_8px_rgba(212,72,58,0.35)]"
        >
          置顶
        </span>
        <div class="w-fit rounded-full bg-white/20 px-2.5 py-0.5 text-xs backdrop-blur">
          {{ current?.categories?.[0]?.name ?? current?.categories?.[0] ?? '文章' }}
        </div>
      </div>
      <RouterLink
        :to="`/p/${current?.uid}`"
        class="line-clamp-2 text-xl font-bold md:text-2xl"
      >
        {{ current?.title }}
      </RouterLink>
      <p
        v-if="current?.description"
        class="mt-2 max-w-2xl line-clamp-2 text-sm text-white/88 md:text-base"
      >
        {{ current.description }}
      </p>
      <div class="mt-4 flex gap-1.5">
        <button
          v-for="(_, i) in items"
          :key="i"
          type="button"
          class="h-1.5 w-1.5 rounded-full transition"
          :class="i === active ? 'bg-white' : 'bg-white/40'"
          @click="active = i"
        />
      </div>
    </div>
  </section>
</template>
