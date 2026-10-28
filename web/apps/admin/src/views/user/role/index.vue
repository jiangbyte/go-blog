<!-- Author: go-blog — 角色管理 -->

<script setup lang="tsx">
import type { PaginationProps } from 'naive-ui'
import type { ProDataTableColumns, ProSearchFormColumns } from 'pro-naive-ui'
import { Icon } from '@iconify/vue/offline'
import { roleApi } from '@/api'
import { hasPermission, normalizeSearchValues, renderButtonIcon } from '@/utils'
import { NButton, NFlex, NIcon, NSwitch, NTag } from 'naive-ui'
import { createProSearchForm, ProCard, ProDataTable, ProSearchForm } from 'pro-naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import { wireBool } from '@/utils/wire'
import ModalDetail from './components/ModalDetail.vue'
import ModalForm from './components/ModalForm.vue'
import ModalGrantResource from './components/ModalGrantResource.vue'

const formModalRef = ref<any>(null)
const detailModalRef = ref<any>(null)
const grantResourceModalRef = ref<any>(null)

const state = reactive({
  roles: [] as any[],
  allRoles: [] as any[],
  total: 0,
  loading: false,
  searchValues: {} as any,
  page: 1,
  pageSize: 20,
})

const searchForm = createProSearchForm<any>({
  defaultCollapsed: true,
  onSubmit(values) {
    state.searchValues = normalizeSearchValues(values, {
      code: (value) => String(value).trim(),
      name: (value) => String(value).trim(),
    })
    state.page = 1
    applyFilter()
  },
  onReset() {
    state.searchValues = {}
    state.page = 1
    applyFilter()
  },
})

const searchColumns = computed<ProSearchFormColumns<any>>(() => [
  {
    title: '角色名称',
    path: 'name',
    field: 'input',
  },
  {
    title: '角色编码',
    path: 'code',
    field: 'input',
  },
  {
    title: '状态',
    path: 'enabled',
    field: 'select',
    fieldProps: {
      clearable: true,
      options: [
        { label: '启用', value: 'true' },
        { label: '禁用', value: 'false' },
      ],
    },
  },
])

const pagination = computed<PaginationProps>(() => ({
  page: state.page,
  pageSize: state.pageSize,
  itemCount: state.total,
  showSizePicker: true,
  pageSizes: [10, 20, 30, 50],
  prefix: ({ itemCount }) => `${itemCount} 条`,
  onUpdatePage: (value) => {
    state.page = value
    applyFilter()
  },
  onUpdatePageSize: (value) => {
    state.pageSize = value
    state.page = 1
    applyFilter()
  },
}))

const tableColumns = computed<ProDataTableColumns<any>>(() => [
  {
    title: '角色名称',
    path: 'name',
    width: 160,
    ellipsis: { tooltip: true },
  },
  {
    title: '角色编码',
    path: 'code',
    width: 150,
    ellipsis: { tooltip: true },
  },
  {
    title: '状态',
    path: 'enabled',
    width: 100,
    render: (row) =>
      hasPermission('admin:role:changeenabled') ? (
        <NSwitch
          size="small"
          value={wireBool(row.enabled)}
          onUpdateValue={(v: boolean) => void onToggle(row, v)}
        />
      ) : (
        <NTag bordered={false} type={wireBool(row.enabled) ? 'success' : 'default'} size="small">
          {wireBool(row.enabled) ? '启用' : '禁用'}
        </NTag>
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 130,
    fixed: 'right',
    render: (row) => (
      <NFlex size={12}>
        {hasPermission('admin:role:page') ? (
          <NButton type="info" text={true} onClick={() => openDetailModal(row)}>
            {renderButtonIcon('icon-park-outline:preview-open')}
          </NButton>
        ) : null}
        {hasPermission('admin:role:update') ? (
          <NButton type="primary" text={true} onClick={() => openEditModal(row)}>
            {renderButtonIcon('icon-park-outline:edit')}
          </NButton>
        ) : null}
        {hasPermission('admin:role:grantresource') ? (
          <NButton type="warning" text={true} onClick={() => openGrantModal(row)}>
            {renderButtonIcon('icon-park-outline:permissions')}
          </NButton>
        ) : null}
      </NFlex>
    ),
  },
])

onMounted(() => {
  void fetchPage()
})

async function fetchPage() {
  state.loading = true
  try {
    const response = await roleApi.list()
    state.allRoles = Array.isArray(response.data) ? response.data : []
    applyFilter()
  } finally {
    state.loading = false
  }
}

/** 前端过滤分页：mypage 角色为全量列表 */
function applyFilter() {
  let list = state.allRoles
  const { name, code, enabled } = state.searchValues
  if (name) {
    list = list.filter((r) => String(r.name ?? '').includes(name))
  }
  if (code) {
    list = list.filter((r) => String(r.code ?? '').includes(code))
  }
  if (enabled === 'true' || enabled === 'false') {
    list = list.filter((r) => String(r.enabled) === enabled)
  }
  state.total = list.length
  const start = (state.page - 1) * state.pageSize
  state.roles = list.slice(start, start + state.pageSize)
}

function openDetailModal(row: any) {
  detailModalRef.value?.openModal(row)
}

function openCreateModal() {
  formModalRef.value?.openModal()
}

function openEditModal(row: any) {
  formModalRef.value?.openModal(row.id, row)
}

function openGrantModal(row: any) {
  grantResourceModalRef.value?.openModal({
    id: row.id,
    name: row.name,
  })
}

async function onToggle(row: any, enabled: boolean) {
  await roleApi.changeEnabled({
    role_id: row.id,
    enabled: enabled ? 'true' : 'false',
  })
  window.$message.success(enabled ? '已启用' : '已禁用')
  await fetchPage()
}
</script>

<template>
  <NFlex class="h-full min-h-0" vertical>
    <ProCard>
      <ProSearchForm
        :form="searchForm"
        :columns="searchColumns"
        :reset-button-props="{ content: '重置' }"
        :search-button-props="{ content: '搜索' }"
        :collapse-button-props="{
          content: searchForm.collapsed.value ? '展开' : '收起',
        }"
      />
    </ProCard>

    <ProDataTable
      class="min-h-0 flex-1"
      remote
      title="角色管理"
      row-key="id"
      :scroll-x="720"
      :columns="tableColumns"
      :data="state.roles"
      :loading="state.loading"
      :pagination="pagination"
    >
      <template #toolbar>
        <NFlex>
          <NButton
            v-if="hasPermission('admin:role:create')"
            type="primary"
            text
            title="新增"
            aria-label="新增"
            @click="openCreateModal"
          >
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:plus" />
              </NIcon>
            </template>
          </NButton>
          <NButton text title="刷新" aria-label="刷新" :loading="state.loading" @click="fetchPage">
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:reload" />
              </NIcon>
            </template>
          </NButton>
        </NFlex>
      </template>
    </ProDataTable>

    <ModalForm ref="formModalRef" @saved="fetchPage" />
    <ModalDetail ref="detailModalRef" />
    <ModalGrantResource ref="grantResourceModalRef" @saved="fetchPage" />
  </NFlex>
</template>
