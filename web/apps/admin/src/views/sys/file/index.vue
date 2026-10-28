<!-- Author: go-blog — 对象存储文件管理 -->

<script setup lang="tsx">
import type { PaginationProps, UploadFileInfo } from 'naive-ui'
import type { ProDataTableColumns, ProSearchFormColumns } from 'pro-naive-ui'
import { Icon } from '@iconify/vue/offline'
import { fileApi } from '@/api'
import {
  formatDateTime,
  formatFileSize,
  hasPermission,
  normalizeSearchValues,
  renderButtonIcon,
} from '@/utils'
import {
  NButton,
  NFlex,
  NIcon,
  NImage,
  NSelect,
  NTag,
  NText,
  NUpload,
} from 'naive-ui'
import { createProSearchForm, ProCard, ProDataTable, ProSearchForm } from 'pro-naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'
import ModalDetail from './components/ModalDetail.vue'
import ModalStorageStatus from './components/ModalStorageStatus.vue'
import type { EngineInfo } from './components/ModalStorageStatus.vue'

const state = reactive({
  records: [] as any[],
  total: 0,
  loading: false,
  uploading: false,
  searchValues: {} as { keyword?: string; storage_provider?: string },
  page: 1,
  pageSize: 20,
  checkedKeys: [] as string[],
  storageProvider: null as string | null,
})

const status = ref<{
  default_engine?: string
  upload_max_bytes?: number
  presign_expire_seconds?: number
  engines?: EngineInfo[]
  hint?: string
} | null>(null)

const statusModalRef = ref<InstanceType<typeof ModalStorageStatus> | null>(null)
const detailModalRef = ref<InstanceType<typeof ModalDetail> | null>(null)

const searchForm = createProSearchForm<any>({
  defaultCollapsed: false,
  onSubmit(values) {
    state.searchValues = normalizeSearchValues(values, {
      keyword: (value) => String(value).trim(),
      storage_provider: (value) => String(value).trim(),
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

const engineOptions = computed(() =>
  (status.value?.engines || [])
    .filter((e) => e.configured)
    .map((e) => ({
      label: `${e.display_name}${e.is_default ? '（默认）' : ''}`,
      value: e.provider,
    })),
)

const searchColumns = computed<ProSearchFormColumns<any>>(() => [
  { title: '关键词', path: 'keyword', field: 'input', proFieldProps: { placeholder: '文件名 / 对象键' } },
  {
    title: '存储',
    path: 'storage_provider',
    field: 'select',
    fieldProps: {
      options: [
        { label: '全部', value: '' },
        ...(status.value?.engines || []).map((e) => ({
          label: e.display_name,
          value: e.provider,
        })),
      ],
      clearable: true,
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

async function fetchStatus() {
  try {
    const res = await fileApi.storageStatus()
    status.value = res.data || null
  } catch {
    status.value = null
  }
}

async function fetchPage() {
  state.loading = true
  try {
    const res = await fileApi.page({
      current: state.page,
      size: state.pageSize,
      keyword: state.searchValues.keyword || undefined,
      storage_provider: state.searchValues.storage_provider || undefined,
    })
    state.records = res.data?.records || []
    state.total = Number(res.data?.total || 0)
    state.checkedKeys = []
  } finally {
    state.loading = false
  }
}

function isImage(row: any) {
  return String(row.content_type || '').startsWith('image/')
}

function providerLabel(provider: string) {
  const hit = (status.value?.engines || []).find((e) => e.provider === provider)
  return hit?.display_name || provider || '—'
}

function openDetail(row: any) {
  if (!row?.id) return
  detailModalRef.value?.openModal(String(row.id), row)
}

async function copyText(text: string) {
  const value = String(text || '')
  if (!value) return
  try {
    await navigator.clipboard.writeText(value)
    window.$message.success('已复制')
  } catch {
    window.$message.error('复制失败')
  }
}

async function onDownload(row: any) {
  try {
    await fileApi.downloadFile(row)
    window.$message.success('开始下载')
  } catch {
    window.$message.error('下载失败')
  }
}

async function onRemove(ids: string[]) {
  if (!ids.length) return
  window.$dialog.warning({
    title: '删除文件',
    content: `确定删除选中的 ${ids.length} 个文件？将同时删除对象存储中的对象。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await fileApi.remove({ ids })
      window.$message.success('已删除')
      await fetchPage()
    },
  })
}

async function customUpload(options: {
  file: UploadFileInfo
  onFinish: () => void
  onError: () => void
}) {
  const raw = options.file.file
  if (!raw) {
    options.onError()
    return
  }
  state.uploading = true
  try {
    await fileApi.upload(raw, {
      storage_provider: state.storageProvider || undefined,
    })
    window.$message.success('上传成功')
    options.onFinish()
    await fetchPage()
  } catch (e: any) {
    window.$message.error(e?.message || '上传失败')
    options.onError()
  } finally {
    state.uploading = false
  }
}

const tableColumns = computed<ProDataTableColumns<any>>(() => [
  { type: 'selection', width: 48 },
  {
    title: '预览',
    key: 'preview',
    width: 72,
    render: (row) =>
      isImage(row) && row.url ? (
        <NImage src={row.url} width={40} height={40} objectFit="cover" />
      ) : (
        <NText depth={3}>—</NText>
      ),
  },
  {
    title: '文件名',
    path: 'original_name',
    minWidth: 160,
    ellipsis: { tooltip: true },
    render: (row) => (
      <NButton text type="primary" onClick={() => openDetail(row)}>
        {row.original_name || '—'}
      </NButton>
    ),
  },
  {
    title: '对象键',
    path: 'object_name',
    minWidth: 200,
    ellipsis: { tooltip: true },
  },
  {
    title: '存储',
    path: 'storage_provider',
    width: 100,
    render: (row) => <NTag size="small">{providerLabel(row.storage_provider)}</NTag>,
  },
  {
    title: '桶',
    path: 'bucket',
    width: 110,
    ellipsis: { tooltip: true },
    render: (row) => row.bucket || '—',
  },
  {
    title: '大小',
    path: 'size',
    width: 90,
    render: (row) => formatFileSize(row.size),
  },
  {
    title: '类型',
    path: 'content_type',
    width: 130,
    ellipsis: { tooltip: true },
  },
  {
    title: '创建人',
    path: 'created_by',
    width: 100,
    ellipsis: { tooltip: true },
    render: (row) => row.created_by || '—',
  },
  {
    title: '上传时间',
    path: 'created_at',
    width: 170,
    render: (row) => formatDateTime(row.created_at),
  },
  {
    title: '操作',
    key: 'actions',
    width: 168,
    fixed: 'right',
    render: (row) => (
      <NFlex size={4}>
        <NButton text title="详情" aria-label="详情" onClick={() => openDetail(row)}>
          {renderButtonIcon('icon-park-outline:info')}
        </NButton>
        <NButton
          text
          title="复制链接"
          aria-label="复制链接"
          onClick={() => copyText(row.url || row.object_name)}
        >
          {renderButtonIcon('icon-park-outline:copy')}
        </NButton>
        <NButton text title="下载" aria-label="下载" onClick={() => onDownload(row)}>
          {renderButtonIcon('icon-park-outline:download')}
        </NButton>
        {hasPermission('admin:file:delete') ? (
          <NButton text title="删除" aria-label="删除" onClick={() => onRemove([row.id])}>
            {renderButtonIcon('icon-park-outline:delete')}
          </NButton>
        ) : null}
      </NFlex>
    ),
  },
])

onMounted(async () => {
  await fetchStatus()
  await fetchPage()
})
</script>

<template>
  <NFlex class="h-full min-h-0" vertical :size="12">
    <ProCard>
      <ProSearchForm
        :form="searchForm"
        :columns="searchColumns"
        :reset-button-props="{ content: '重置' }"
        :search-button-props="{ content: '查询' }"
        :collapse-button-props="{ content: searchForm.collapsed.value ? '展开' : '收起' }"
      />
    </ProCard>

    <ProDataTable
      class="min-h-0 flex-1"
      remote
      title="文件"
      row-key="id"
      :scroll-x="1280"
      :columns="tableColumns"
      :data="state.records"
      :loading="state.loading"
      :pagination="pagination"
      :checked-row-keys="state.checkedKeys"
      :on-update-checked-row-keys="(keys) => (state.checkedKeys = keys as string[])"
    >
      <template #toolbar>
        <div class="file-toolbar">
          <NSelect
            v-model:value="state.storageProvider"
            class="file-toolbar__select"
            size="small"
            clearable
            placeholder="上传引擎"
            :options="engineOptions"
          />
          <NUpload
            class="file-toolbar__upload"
            :show-file-list="false"
            :custom-request="customUpload"
            :disabled="state.uploading"
            multiple
          >
            <NButton type="primary" text title="上传" aria-label="上传" :loading="state.uploading">
              <template #icon>
                <NIcon>
                  <Icon icon="icon-park-outline:upload" />
                </NIcon>
              </template>
            </NButton>
          </NUpload>
          <NButton
            v-if="hasPermission('admin:file:delete')"
            text
            title="批量删除"
            aria-label="批量删除"
            :disabled="!state.checkedKeys.length"
            @click="onRemove(state.checkedKeys)"
          >
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:delete" />
              </NIcon>
            </template>
          </NButton>
          <NButton text title="存储引擎" aria-label="存储引擎" @click="statusModalRef?.open()">
            <template #icon>
              <NIcon>
                <Icon icon="icon-park-outline:hard-disk" />
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
        </div>
      </template>
    </ProDataTable>

    <ModalDetail
      ref="detailModalRef"
      :provider-label="providerLabel"
      @deleted="fetchPage"
    />
    <ModalStorageStatus
      ref="statusModalRef"
      :engines="status?.engines || []"
      :hint="status?.hint"
      :default-engine="status?.default_engine"
      :upload-max-bytes="status?.upload_max_bytes"
      :presign-expire-seconds="status?.presign_expire_seconds"
    />
  </NFlex>
</template>

<style scoped>
.file-toolbar {
  display: inline-flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 8px;
  max-width: 100%;
}
.file-toolbar__select {
  width: 140px;
  flex: 0 0 140px;
}
.file-toolbar__upload {
  display: inline-flex !important;
  width: auto !important;
}
.file-toolbar__upload :deep(.n-upload-trigger) {
  display: inline-flex;
}
</style>
