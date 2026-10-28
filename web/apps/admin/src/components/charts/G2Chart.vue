<!-- Author: go-blog — AntV G2 通用容器：数据变化重绘，卸载销毁 -->

<script setup lang="ts">
import { Chart } from '@antv/g2'
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 图表高度 */
    height?: number
    /** 由外部声明式配置图表 */
    spec: (chart: Chart) => void
    /** 依赖变化时重绘（通常传入序列化后的数据） */
    deps?: unknown
  }>(),
  {
    height: 280,
    deps: undefined,
  },
)

const hostRef = ref<HTMLDivElement | null>(null)
let chart: Chart | null = null

function render() {
  const el = hostRef.value
  if (!el) return
  chart?.destroy()
  chart = new Chart({
    container: el,
    autoFit: true,
    height: props.height,
  })
  props.spec(chart)
  chart.render()
}

onMounted(render)
watch(
  () => [props.height, props.deps] as const,
  () => render(),
  { deep: true },
)
onBeforeUnmount(() => {
  chart?.destroy()
  chart = null
})
</script>

<template>
  <div ref="hostRef" class="g2-chart-host w-full min-w-0" />
</template>
