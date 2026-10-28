<script setup lang="ts">
import SubnetCalculator from '@/components/tools/SubnetCalculator.vue'
</script>

<template>
  <div>
    <div class="mb-4">
      <h1 class="text-2xl font-bold">子网划分</h1>
      <p class="mt-1 text-sm text-muted">IPv4 CIDR / 掩码计算：网络地址、广播、可用主机等</p>
    </div>
    <SubnetCalculator />
  </div>
</template>
