/** Author: go-blog */

import { http } from '@/utils'

/** 站点公开信息：品牌 + 版权备案（Admin / Portal 共用）。 */
export function siteFooter() {
  return http.get<{
    app_name?: string
    app_subtitle?: string
    copyright_text?: string
    copyright_url?: string
    icp_number?: string
    icp_url?: string
    psb_number?: string
    psb_url?: string
  }>('/portal/sys/site/footer', {
    public: true,
  })
}

/** 服务条款 / 隐私政策 HTML。 */
export function legal() {
  return http.get<{
    terms_of_service?: string
    privacy_policy?: string
  }>('/portal/sys/site/legal', {
    public: true,
  })
}
