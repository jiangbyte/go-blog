<!-- Author: go-blog — 分类 / 标签编辑弹窗 -->

<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import type { TaxonomyKind } from '@/api/blog/taxonomy'
import { blogTaxonomyApi } from '@/api'
import { createRequiredRule } from '@/utils'
import { computed, reactive, ref } from 'vue'

const props = defineProps<{
  kind: TaxonomyKind
  kindLabel: string
}>()

const emit = defineEmits<{ saved: [] }>()
const formRef = ref<FormInst | null>(null)

const state = reactive({
  showModal: false,
  submitLoading: false,
  dataUid: null as string | null,
  formModel: { name: '' },
})

const modalTitle = computed(() =>
  state.dataUid ? `编辑${props.kindLabel}` : `新增${props.kindLabel}`,
)

const rules = computed<FormRules>(() => ({
  name: [createRequiredRule('名称', 'input')],
}))

function openModal(row?: { uid?: string; name?: string } | null) {
  state.dataUid = row?.uid ? String(row.uid) : null
  state.formModel = { name: String(row?.name ?? '') }
  state.showModal = true
}

function closeModal() {
  state.showModal = false
  state.submitLoading = false
}

async function submitForm() {
  await formRef.value?.validate()
  state.submitLoading = true
  try {
    const name = state.formModel.name.trim()
    if (state.dataUid) {
      await blogTaxonomyApi.update({ uid: state.dataUid, name })
      window.$message.success('更新成功')
    } else {
      await blogTaxonomyApi.create({ kind: props.kind, name })
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
    :width="420"
    :mask-closable="false"
  >
    <NForm
      ref="formRef"
      :model="state.formModel"
      :rules="rules"
      label-placement="left"
      label-width="70"
      :disabled="state.submitLoading"
    >
      <NFormItem label="名称" path="name">
        <NInput v-model:value="state.formModel.name" maxlength="128" show-count />
      </NFormItem>
    </NForm>
    <template #action>
      <NSpace justify="end">
        <NButton @click="closeModal">取消</NButton>
        <NButton type="primary" :loading="state.submitLoading" @click="submitForm">确认</NButton>
      </NSpace>
    </template>
  </FormContainer>
</template>
