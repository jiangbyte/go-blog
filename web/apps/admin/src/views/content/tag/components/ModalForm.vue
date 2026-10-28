<!-- Author: go-blog — 内容标签编辑 -->

<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { contentTagApi } from '@/api'
import { createRequiredRule } from '@/utils'
import { computed, reactive, ref } from 'vue'

const emit = defineEmits<{ saved: [] }>()
const formRef = ref<FormInst | null>(null)
const defaultFormData = { name: '', slug: '', sort_code: 0 }

const state = reactive({
  showModal: false,
  loading: false,
  submitLoading: false,
  dataId: null as string | null,
  formModel: { ...defaultFormData },
})

const modalTitle = computed(() => (state.dataId ? '编辑标签' : '新增标签'))
const rules = computed<FormRules>(() => ({
  name: [createRequiredRule('名称', 'input')],
}))

async function openModal(id?: string, row?: any) {
  state.dataId = id ?? null
  state.formModel = { ...defaultFormData }
  state.showModal = true
  if (id) {
    state.loading = true
    try {
      try {
        const res = await contentTagApi.detail(id)
        fillFromRow(res.data ?? row)
      } catch {
        fillFromRow(row)
      }
    } finally {
      state.loading = false
    }
  }
}

function fillFromRow(row: any) {
  if (!row) return
  state.formModel = {
    name: String(row.name ?? ''),
    slug: String(row.slug ?? ''),
    sort_code: Number(row.sort_code ?? 0) || 0,
  }
}

function closeModal() {
  state.showModal = false
  state.submitLoading = false
}

async function submitForm() {
  await formRef.value?.validate()
  state.submitLoading = true
  try {
    const payload = {
      name: state.formModel.name.trim(),
      sort_code: state.formModel.sort_code,
    }
    if (state.dataId) {
      await contentTagApi.update({ id: state.dataId, ...payload })
      window.$message.success('更新成功')
    } else {
      await contentTagApi.create(payload)
      window.$message.success('创建成功')
    }
    closeModal()
    emit('saved')
  } finally {
    state.submitLoading = false
  }
}

defineExpose({ openModal })
</script>

<template>
  <FormContainer
    v-model:show="state.showModal"
    :title="modalTitle"
    :width="480"
    :mask-closable="false"
  >
    <NSpin :show="state.loading">
      <NForm
        ref="formRef"
        :model="state.formModel"
        :rules="rules"
        label-placement="left"
        label-width="70"
        :disabled="state.loading || state.submitLoading"
      >
        <NFormItem label="名称" path="name">
          <NInput v-model:value="state.formModel.name" maxlength="64" show-count />
        </NFormItem>
        <NFormItem v-if="state.dataId" label="标识" path="slug">
          <NInput :value="state.formModel.slug" disabled />
        </NFormItem>
        <NFormItem label="排序" path="sort_code">
          <NInputNumber
            v-model:value="state.formModel.sort_code"
            :show-button="false"
            class="w-full"
          />
        </NFormItem>
      </NForm>
    </NSpin>
    <template #action>
      <NSpace justify="end">
        <NButton @click="closeModal"> 取消 </NButton>
        <NButton type="primary" :loading="state.submitLoading" @click="submitForm"> 确认 </NButton>
      </NSpace>
    </template>
  </FormContainer>
</template>
