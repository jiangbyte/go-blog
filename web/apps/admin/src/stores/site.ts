/** Author: go-blog — 站点品牌与页脚：来自 /portal/sys/site/footer */

import { defineStore } from 'pinia'
import { siteApi } from '@/api'
import {
  type SiteFooterInfo,
  DEFAULT_APP_NAME,
  emptySiteFooter,
  parseSiteFooter,
  resolveAppName,
} from '@/composables/siteFooter'

type SiteState = {
  loaded: boolean
  footer: SiteFooterInfo
}

let loadPromise: Promise<void> | null = null

export const useSiteStore = defineStore('site-store', {
  state: (): SiteState => ({
    loaded: false,
    footer: emptySiteFooter(),
  }),
  getters: {
    appName(state): string {
      return resolveAppName(state.footer.appName)
    },
    appSubtitle(state): string {
      return state.footer.appSubtitle
    },
  },
  actions: {
    /** 拉取公开站点信息（幂等，并发共用同一 Promise）。 */
    ensureLoaded(): Promise<void> {
      if (this.loaded) return Promise.resolve()
      if (loadPromise) return loadPromise
      // 1. 请求公开接口
      // 2. 解析并写入 state；失败时保留空页脚 + 默认应用名
      // 3. 同步 document.title 后缀用的品牌名
      loadPromise = siteApi
        .siteFooter()
        .then((res) => {
          this.footer = parseSiteFooter((res.data ?? {}) as Record<string, unknown>)
          this.loaded = true
        })
        .catch(() => {
          this.footer = { ...emptySiteFooter(), appName: DEFAULT_APP_NAME }
          this.loaded = true
        })
        .finally(() => {
          loadPromise = null
        })
      return loadPromise
    },
  },
})
