<!-- Author: go-blog -->

<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import CropAvatarUploadModal from '@/components/upload/CropAvatarUploadModal.vue'
import StorageImage from '@/components/upload/StorageImage.vue'
import ConfigSectionLayout from './ConfigSectionLayout.vue'
import { loadByCategory, parseTags, saveByKeys, tagsToJson } from '../composables/useConfigForm'

const CATEGORY = 'SYS'
const MAX_TAGS = 8
const MAX_SIGNATURE = 120

const state = reactive({
  loading: false,
  saving: false,
  avatarModalShow: false,
  appName: 'Demo Blog',
  appSubtitle: '',
  copyrightText: '',
  copyrightUrl: '',
  icpNumber: '',
  icpUrl: 'https://beian.miit.gov.cn/',
  psbNumber: '',
  psbUrl: '',
  homeProfileAvatar: '',
  homeProfileName: '',
  homeProfileTags: [] as string[],
  homeProfileSignature: '',
  snapshot: '',
})

onMounted(() => {
  void reload()
})

/** 规范化标签：去空白、去重空项、截断上限。 */
function normalizeTags(tags: string[]): string[] {
  return (tags || [])
    .map((t) => String(t).trim())
    .filter(Boolean)
    .slice(0, MAX_TAGS)
}

function snapshotOf() {
  return JSON.stringify({
    appName: state.appName,
    appSubtitle: state.appSubtitle,
    copyrightText: state.copyrightText,
    copyrightUrl: state.copyrightUrl,
    icpNumber: state.icpNumber,
    icpUrl: state.icpUrl,
    psbNumber: state.psbNumber,
    psbUrl: state.psbUrl,
    homeProfileAvatar: state.homeProfileAvatar,
    homeProfileName: state.homeProfileName,
    homeProfileTags: state.homeProfileTags,
    homeProfileSignature: state.homeProfileSignature,
  })
}

async function reload() {
  state.loading = true
  try {
    const map = await loadByCategory(CATEGORY)
    state.appName = map.APP_NAME || 'Demo Blog'
    state.appSubtitle = map.APP_SUBTITLE || ''
    state.copyrightText = map.COPYRIGHT_TEXT || ''
    state.copyrightUrl = map.COPYRIGHT_URL || ''
    state.icpNumber = map.SITE_ICP_NUMBER || ''
    state.icpUrl = map.SITE_ICP_URL || 'https://beian.miit.gov.cn/'
    state.psbNumber = map.SITE_PSB_NUMBER || ''
    state.psbUrl = map.SITE_PSB_URL || ''
    state.homeProfileAvatar = map.HOME_PROFILE_AVATAR || ''
    state.homeProfileName = map.HOME_PROFILE_NAME || ''
    state.homeProfileTags = normalizeTags(parseTags(map.HOME_PROFILE_TAGS || ''))
    state.homeProfileSignature = map.HOME_PROFILE_SIGNATURE || ''
    state.snapshot = snapshotOf()
  } finally {
    state.loading = false
  }
}

function reset() {
  if (!state.snapshot) return
  Object.assign(state, JSON.parse(state.snapshot))
}

function onTagsUpdate(tags: string[]) {
  state.homeProfileTags = normalizeTags(tags)
}

function onAvatarUploaded(url: string) {
  state.homeProfileAvatar = (url || '').trim()
}

function clearAvatar() {
  state.homeProfileAvatar = ''
}

async function save() {
  state.saving = true
  try {
    const tags = normalizeTags(state.homeProfileTags)
    await saveByKeys([
      {
        config_key: 'APP_NAME',
        config_value: state.appName.trim(),
        category: CATEGORY,
      },
      {
        config_key: 'APP_SUBTITLE',
        config_value: state.appSubtitle.trim(),
        category: CATEGORY,
      },
      {
        config_key: 'COPYRIGHT_TEXT',
        config_value: state.copyrightText,
        category: CATEGORY,
      },
      {
        config_key: 'COPYRIGHT_URL',
        config_value: state.copyrightUrl,
        category: CATEGORY,
      },
      {
        config_key: 'SITE_ICP_NUMBER',
        config_value: state.icpNumber.trim(),
        category: CATEGORY,
      },
      {
        config_key: 'SITE_ICP_URL',
        config_value: state.icpUrl.trim(),
        category: CATEGORY,
      },
      {
        config_key: 'SITE_PSB_NUMBER',
        config_value: state.psbNumber.trim(),
        category: CATEGORY,
      },
      {
        config_key: 'SITE_PSB_URL',
        config_value: state.psbUrl.trim(),
        category: CATEGORY,
      },
      {
        config_key: 'HOME_PROFILE_AVATAR',
        config_value: (state.homeProfileAvatar || '').trim(),
        category: CATEGORY,
      },
      {
        config_key: 'HOME_PROFILE_NAME',
        config_value: state.homeProfileName.trim(),
        category: CATEGORY,
      },
      {
        config_key: 'HOME_PROFILE_TAGS',
        config_value: tagsToJson(tags),
        category: CATEGORY,
        value_type: 'JSON',
      },
      {
        config_key: 'HOME_PROFILE_SIGNATURE',
        config_value: state.homeProfileSignature.trim().slice(0, MAX_SIGNATURE),
        category: CATEGORY,
      },
    ])
    state.homeProfileTags = tags
    state.homeProfileSignature = state.homeProfileSignature.trim().slice(0, MAX_SIGNATURE)
    window.$message.success('保存成功')
    state.snapshot = snapshotOf()
  } finally {
    state.saving = false
  }
}
</script>

<template>
  <NSpin :show="state.loading">
    <ConfigSectionLayout
      description="配置应用名称、副标题、首页个人卡片、版权文案与 ICP/公安备案信息。Admin、Portal 等端通过公开接口统一获取。"
      :saving="state.saving"
      @save="save"
      @reset="reset"
    >
      <NForm class="sys-config-form sys-config-form--fluid" label-placement="top">
        <NGrid :cols="24" :x-gap="24" :y-gap="4">
          <NGi :span="12">
            <NFormItem label="应用名称">
              <NInput v-model:value="state.appName" placeholder="Demo Blog" />
            </NFormItem>
          </NGi>
          <NGi :span="12">
            <NFormItem label="应用副标题">
              <NInput v-model:value="state.appSubtitle" placeholder="探索～发现～" />
            </NFormItem>
          </NGi>

          <NGi :span="24">
            <NDivider title-placement="left"> 首页个人卡片 </NDivider>
          </NGi>
          <NGi :span="12">
            <NFormItem label="头像">
              <NSpace align="center" :size="16">
                <button
                  class="home-profile-avatar-btn"
                  type="button"
                  title="更换头像"
                  @click="state.avatarModalShow = true"
                >
                  <div v-if="state.homeProfileAvatar" class="home-profile-avatar-btn__img">
                    <StorageImage
                      :object-name="state.homeProfileAvatar"
                      :src="state.homeProfileAvatar"
                      alt="头像"
                      :width="72"
                      :height="72"
                      object-fit="cover"
                    />
                  </div>
                  <NAvatar v-else round :size="72">
                    <NovaIcon icon="icon-park-outline:user" :size="32" />
                  </NAvatar>
                  <span class="home-profile-avatar-btn__badge">编辑</span>
                </button>
                <NSpace vertical :size="8">
                  <NButton secondary type="primary" @click="state.avatarModalShow = true">
                    裁剪上传
                  </NButton>
                  <NButton v-if="state.homeProfileAvatar" quaternary @click="clearAvatar">
                    清除头像
                  </NButton>
                </NSpace>
              </NSpace>
            </NFormItem>
          </NGi>
          <NGi :span="12">
            <NFormItem label="姓名">
              <NInput
                v-model:value="state.homeProfileName"
                placeholder="展示在首页关于卡片"
                :maxlength="32"
                show-count
              />
            </NFormItem>
            <NFormItem label="个人标签" :feedback="`最多 ${MAX_TAGS} 个`">
              <NDynamicTags
                :value="state.homeProfileTags"
                :max="MAX_TAGS"
                @update:value="onTagsUpdate"
              />
            </NFormItem>
          </NGi>
          <NGi :span="24">
            <NFormItem label="个人签名">
              <NInput
                v-model:value="state.homeProfileSignature"
                type="textarea"
                placeholder="一句话介绍自己"
                :maxlength="MAX_SIGNATURE"
                show-count
                :autosize="{ minRows: 2, maxRows: 4 }"
              />
            </NFormItem>
          </NGi>

          <NGi :span="24">
            <NDivider title-placement="left"> 版权信息 </NDivider>
          </NGi>
          <NGi :span="12">
            <NFormItem label="版权文案">
              <NInput v-model:value="state.copyrightText" />
            </NFormItem>
          </NGi>
          <NGi :span="12">
            <NFormItem label="版权链接">
              <NInput v-model:value="state.copyrightUrl" placeholder="https://" />
            </NFormItem>
          </NGi>

          <NGi :span="24">
            <NDivider title-placement="left"> 备案信息 </NDivider>
          </NGi>
          <NGi :span="12">
            <NFormItem label="ICP 备案号">
              <NInput v-model:value="state.icpNumber" placeholder="京ICP备xxxxxxxx号" />
            </NFormItem>
          </NGi>
          <NGi :span="12">
            <NFormItem label="ICP 备案链接">
              <NInput v-model:value="state.icpUrl" placeholder="https://beian.miit.gov.cn/" />
            </NFormItem>
          </NGi>
          <NGi :span="12">
            <NFormItem label="公安备案号">
              <NInput v-model:value="state.psbNumber" placeholder="京公网安备 xxxxxxxx号" />
            </NFormItem>
          </NGi>
          <NGi :span="12">
            <NFormItem label="公安备案链接">
              <NInput v-model:value="state.psbUrl" placeholder="https://www.beian.gov.cn/..." />
            </NFormItem>
          </NGi>
        </NGrid>
      </NForm>
    </ConfigSectionLayout>

    <CropAvatarUploadModal
      v-model:show="state.avatarModalShow"
      :avatar="state.homeProfileAvatar || null"
      @uploaded="onAvatarUploaded"
    />
  </NSpin>
</template>

<style scoped>
.home-profile-avatar-btn {
  position: relative;
  display: inline-flex;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
  border-radius: 999px;
}

.home-profile-avatar-btn__img {
  width: 72px;
  height: 72px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--body-color);
}

.home-profile-avatar-btn__badge {
  position: absolute;
  right: 0;
  bottom: 0;
  padding: 2px 8px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--primary-color) 92%, #000);
  color: #fff;
  font-size: 12px;
  line-height: 1.4;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.16);
}
</style>
