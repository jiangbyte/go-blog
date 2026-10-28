<!-- Author: go-blog — 用户管理 -->

<script setup lang="tsx">
import type { PaginationProps } from 'naive-ui'
import type { ProDataTableColumns, ProSearchFormColumns } from 'pro-naive-ui'
import { Icon } from '@iconify/vue/offline'
import { accountApi } from '@/api'
import { formatDateTime, hasPermission, normalizeSearchValues, renderButtonIcon } from '@/utils'
import { NButton, NFlex, NIcon, NSwitch, NTag } from 'naive-ui'
import { createProSearchForm, ProCard, ProDataTable, ProSearchForm } from 'pro-naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import { wireBool } from '@/utils/wire'
import ModalDetail from './components/ModalDetail.vue'
import ModalForm from './components/ModalForm.vue'
import ModalGrantRole from '../components/ModalGrantRole.vue'

const formRef = ref<any>(null)
const detailRef = ref<any>(null)
const grantRoleModalRef = ref<any>(null)

const state = reactive({
  accounts: [] as any[],
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
      username: (value) => String(value).trim(),
    })
    state.page = 1
    void fetchPage()
  },
  onReset() {
    state.searchValues = {}
    state.page = 1
    void fetchPage()
  },
})

const searchColumns = computed<ProSearchFormColumns<any>>(() => [
  { title: '账号', path: 'username', field: 'input' },
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
    void fetchPage()
  },
  onUpdatePageSize: (value) => {
    state.pageSize = value
    state.page = 1
    void fetchPage()
  },
}))

const tableColumns = computed<ProDataTableColumns<any>>(() => [
  { title: '账号', path: 'username', width: 140, ellipsis: { tooltip: true } },
  { title: '昵称', path: 'nickname', width: 130, ellipsis: { tooltip: true } },
  {
    title: '类型',
    path: 'user_type',
    width: 100,
    render: (row) => (
      <NTag bordered={false} type={row.user_type === 'ADMIN' ? 'info' : 'success'} size="small">
        {row.user_type === 'ADMIN' ? '管理员' : '前台'}
      </NTag>
    ),
  },
  {
    title: '状态',
    path: 'enabled',
    width: 100,
    render: (row) =>
      hasPermission('admin:user:changeenabled') ? (
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
  { title: '手机号', path: 'phone', width: 140 },
  { title: '邮箱', path: 'email', width: 200, ellipsis: { tooltip: true } },
  {
    title: '更新时间',
    path: 'update_time',
    width: 180,
    render: (row) => formatDateTime(row.update_time || row.updated_at),
  },
  {
    title: '操作',
    key: 'actions',
    width: 130,
    fixed: 'right',
    render: (row) => (
      <NFlex size={12}>
        {hasPermission('admin:user:detail') ? (
          <NButton type="info" text onClick={() => detailRef.value?.openModal(row.id, row)}>
            {renderButtonIcon('icon-park-outline:preview-open')}
          </NButton>
        ) : null}
        {hasPermission('admin:user:update') ? (
          <NButton type="primary" text onClick={() => formRef.value?.openModal(row.id, row)}>
            {renderButtonIcon('icon-park-outline:edit')}
          </NButton>
        ) : null}
        {hasPermission('admin:user:grantrole') ? (
          <NButton
            type="warning"
            text
            onClick={() =>
              grantRoleModalRef.value?.openModal({
                id: row.id,
                username: row.username,
                nickname: row.nickname || row.username,
                name: row.nickname || row.username,
              })
            }
          >
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
    const response = await accountApi.page({
      page: state.page,
      page_size: state.pageSize,
      username: state.searchValues.username,
      enabled: state.searchValues.enabled,
    })
    state.accounts = response.data?.records ?? []
    state.total = Number(response.data?.total ?? 0)
    state.page = Number(response.data?.page ?? state.page)
  } finally {
    state.loading = false
  }
}

async function onToggle(row: any, enabled: boolean) {
  await accountApi.changeEnabled({
    user_id: row.id,
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
      title="用户管理"
      row-key="id"
      :scroll-x="1100"
      :columns="tableColumns"
      :data="state.accounts"
      :loading="state.loading"
      :pagination="pagination"
    >
      <template #toolbar>
        <NFlex>
          <NButton
            v-if="hasPermission('admin:user:create')"
            type="primary"
            text
            title="新增"
            aria-label="新增"
            @click="formRef?.openModal()"
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

    <ModalForm ref="formRef" @saved="fetchPage" />
    <ModalDetail ref="detailRef" />
    <ModalGrantRole ref="grantRoleModalRef" @saved="fetchPage" />
  </NFlex>
</template>
