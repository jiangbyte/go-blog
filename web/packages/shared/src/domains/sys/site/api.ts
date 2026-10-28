import type { AxiosInstance } from 'axios'
import type { ApiEnvelope } from '../../user/account/types'
import type { LegalContentData, SiteFooterData } from './types'

/** 公开站点信息 API */
export function createSiteApi(http: AxiosInstance) {
  return {
    getSiteFooter() {
      return http.get<ApiEnvelope<SiteFooterData>>('/portal/sys/site/footer', {
        public: true,
      })
    },
    getLegal() {
      return http.get<ApiEnvelope<LegalContentData>>('/portal/sys/site/legal', {
        public: true,
      })
    },
  }
}

export type SiteApi = ReturnType<typeof createSiteApi>
