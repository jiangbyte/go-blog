<!-- Author: go-blog — 内容分类编辑 -->

<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { contentCategoryApi } from '@/api'
import { createRequiredRule } from '@/utils'
import { computed, reactive, ref } from 'vue'

const emit = defineEmits<{ saved: [] }>()

const formRef = ref<FormInst | null>(null)
const treeOptions = ref<any[]>([])

const defaultFormData = {
  name: '',
  slug: '',
  parent_id: null as string | null,
  sort_code: 0,
}

const state = reactive({
  showModal: false,
  loading: false,
  submitLoading: false,
  dataId: null as string | null,
  formModel: { ...defaultFormData },
})

const modalTitle = computed(() => (state.dataId ? '编辑分类' : '新增分类'))

const rules = computed<FormRules>(() => ({
  name: [createRequiredRule('名称', 'input')],
}))

function mapTreeOptions(nodes: any[], excludeId?: string | null): any[] {
  return (nodes || [])
    .filter((n) => n.id !== excludeId)
    .map((n) => ({
      key: n.id,
      label: n.name,
      children: n.children?.length ? mapTreeOptions(n.children, excludeId) : undefined,
    }))
}

async function openModal(id?: string, row?: any, parentId?: string | null) {
  state.dataId = id ?? null
  state.formModel = { ...defaultFormData, parent_id: parentId ?? null }
  state.showModal = true
  state.loading = true
  try {
    const treeRes = await contentCategoryApi.tree()
    treeOptions.value = mapTreeOptions(treeRes.data || [], id)
    if (id) {
      try {
        const res = await contentCategoryApi.detail(id)
        fillFromRow(res.data ?? row)
      } catch {
        fillFromRow(row)
      }
    }
  } finally {
    state.loading = false
  }
}

function fillFromRow(row: any) {
  if (!row) return
  state.formModel = {
    name: String(row.name ?? ''),
    slug: String(row.slug ?? ''),
    parent_id: row.parent_id ?? null,
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
      parent_id: state.formModel.parent_id,
      sort_code: state.formModel.sort_code,
    }
    if (state.dataId) {
      await contentCategoryApi.update({ id: state.dataId, ...payload })
      window.$message.success('更新成功')
    } else {
      await contentCategoryApi.create(payload)
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
    :width="520"
    :mask-closable="false"
  >
    <NSpin :show="state.loading">
      <NForm
        ref="formRef"
        :model="state.formModel"
        :rules="rules"
        label-placement="left"
        label-width="80"
        :disabled="state.loading || state.submitLoading"
      >
        <NFormItem label="父分类" path="parent_id">
          <NTreeSelect
            v-model:value="state.formModel.parent_id"
            :options="treeOptions"
            clearable
            placeholder="不选则为根分类"
            key-field="key"
            label-field="label"
            default-expand-all
          />
        </NFormItem>
        <NFormItem label="名称" path="name">
          <NInput v-model:value="state.formModel.name" maxlength="128" show-count />
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
