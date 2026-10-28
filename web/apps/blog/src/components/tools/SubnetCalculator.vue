<!-- Author: go-blog — IPv4 子网划分计算（左右分栏） -->

<script setup lang="ts">
import { computed, ref } from 'vue'
import { calculateSubnet, parseSubnetInput, type SubnetResult } from '@/utils/subnet'

const cidr = ref('192.168.1.0/24')
const mask = ref('')
const error = ref('')
const result = ref<SubnetResult | null>(null)

const rows = computed(() => {
  const r = result.value
  if (!r) return []
  return [
    { label: 'CIDR', value: r.input },
    { label: 'IP 地址', value: r.ip },
    { label: '前缀长度', value: String(r.prefix) },
    { label: '子网掩码', value: r.netmask },
    { label: '通配符掩码', value: r.wildcard },
    { label: '网络地址', value: r.network },
    { label: '广播地址', value: r.broadcast },
    { label: '可用主机范围', value: `${r.firstHost} – ${r.lastHost}` },
    { label: '地址总数', value: String(r.totalHosts) },
    { label: '可用主机数', value: String(r.usableHosts) },
    { label: '主机位', value: String(r.hostBits) },
  ]
})

function run() {
  error.value = ''
  result.value = null
  const parsed = parseSubnetInput(cidr.value, mask.value)
  if ('error' in parsed) {
    error.value = parsed.error
    return
  }
  const out = calculateSubnet(parsed.ip, parsed.prefix)
  if ('error' in out) {
    error.value = out.error
    return
  }
  result.value = out
}

function fillExample(example: string) {
  cidr.value = example
  mask.value = ''
  run()
}

run()
</script>

<template>
  <div class="tool-panel glass-card">
    <div class="tool-split">
      <section class="tool-pane">
        <header class="tool-pane__head">输入</header>
        <div class="tool-fields">
          <label class="tool-field">
            <span class="tool-label">IP / CIDR</span>
            <input
              v-model="cidr"
              class="tool-input"
              type="text"
              spellcheck="false"
              placeholder="192.168.1.10/24"
              @keyup.enter="run"
            />
          </label>
          <label class="tool-field">
            <span class="tool-label">掩码或前缀（可选）</span>
            <input
              v-model="mask"
              class="tool-input"
              type="text"
              spellcheck="false"
              placeholder="255.255.255.0 或 24"
              @keyup.enter="run"
            />
          </label>
          <div class="tool-actions">
            <button type="button" class="tool-btn tool-btn--primary" @click="run">计算</button>
            <button type="button" class="tool-btn" @click="fillExample('10.0.0.1/8')">/8 示例</button>
            <button type="button" class="tool-btn" @click="fillExample('172.16.5.20/16')">
              /16 示例
            </button>
            <button type="button" class="tool-btn" @click="fillExample('192.168.1.100/24')">
              /24 示例
            </button>
          </div>
          <p v-if="error" class="tool-feedback tool-feedback--error">{{ error }}</p>
          <p class="tool-hint">支持 CIDR，或 IP + 点分掩码 / 前缀长度。</p>
        </div>
      </section>

      <section class="tool-pane">
        <header class="tool-pane__head">结果</header>
        <dl v-if="rows.length" class="tool-result">
          <div v-for="row in rows" :key="row.label" class="tool-result__row">
            <dt>{{ row.label }}</dt>
            <dd>{{ row.value }}</dd>
          </div>
        </dl>
        <p v-else class="tool-empty">计算结果将显示在这里</p>
      </section>
    </div>
  </div>
</template>

<style scoped>
.tool-panel {
  padding: 1rem 1.1rem 1.15rem;
}

.tool-split {
  display: grid;
  grid-template-columns: minmax(16rem, 0.9fr) minmax(18rem, 1.1fr);
  gap: 1rem;
  align-items: start;
}

.tool-pane {
  min-width: 0;
}

.tool-pane__head {
  margin-bottom: 0.55rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--milk-ink);
}

.tool-fields {
  display: grid;
  gap: 0.75rem;
}

.tool-field {
  display: grid;
  gap: 0.35rem;
}

.tool-label {
  font-size: 0.8rem;
  color: var(--efu-secondtext);
}

.tool-input {
  width: 100%;
  box-sizing: border-box;
  min-height: 2.4rem;
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--efu-card-border);
  border-radius: 8px;
  background: var(--efu-secondbg);
  color: var(--milk-ink);
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 0.9rem;
}

.tool-input:focus {
  outline: 2px solid color-mix(in srgb, var(--milk-ink) 22%, transparent);
  outline-offset: 1px;
}

.tool-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.tool-btn {
  min-height: 2rem;
  padding: 0.3rem 0.85rem;
  border: 1px solid var(--efu-card-border);
  border-radius: 8px;
  background: var(--efu-secondbg);
  color: var(--milk-ink);
  font-size: 0.85rem;
  cursor: pointer;
}

.tool-btn:hover {
  background: var(--efu-ahoverbg);
}

.tool-btn--primary {
  border-color: transparent;
  background: var(--milk-ink);
  color: #fff;
}

.tool-btn--primary:hover {
  opacity: 0.9;
  background: var(--milk-ink);
  color: #fff;
}

.tool-feedback {
  margin: 0;
  font-size: 0.85rem;
}

.tool-feedback--error {
  color: #d4483a;
}

.tool-hint {
  margin: 0;
  font-size: 0.78rem;
  color: var(--efu-secondtext);
  line-height: 1.45;
}

.tool-empty {
  margin: 0;
  padding: 1.25rem;
  border: 1px dashed var(--efu-card-border);
  border-radius: 10px;
  color: var(--efu-secondtext);
  font-size: 0.85rem;
  text-align: center;
}

.tool-result {
  margin: 0;
  border: 1px solid var(--efu-card-border);
  border-radius: 10px;
  overflow: hidden;
}

.tool-result__row {
  display: grid;
  grid-template-columns: 8.5rem 1fr;
  gap: 0.75rem;
  padding: 0.65rem 0.9rem;
  border-top: 1px solid var(--efu-card-border);
  font-size: 0.9rem;
}

.tool-result__row:first-child {
  border-top: none;
}

.tool-result__row:nth-child(even) {
  background: var(--efu-secondbg);
}

.tool-result__row dt {
  margin: 0;
  color: var(--efu-secondtext);
}

.tool-result__row dd {
  margin: 0;
  font-family: var(--font-mono, ui-monospace, monospace);
  word-break: break-all;
}

@media (max-width: 839px) {
  .tool-split {
    grid-template-columns: 1fr;
  }

  .tool-result__row {
    grid-template-columns: 1fr;
    gap: 0.2rem;
  }
}
</style>
