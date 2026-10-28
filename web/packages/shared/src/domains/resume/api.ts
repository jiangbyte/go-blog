import type { AxiosInstance } from 'axios'
import type { ApiEnvelope } from '../user/account/types'
import type { ResumeDocument } from './types'

/** 公开 / 预览简历 API */
export function createResumeApi(http: AxiosInstance) {
  return {
    getPublished() {
      return http.get<ApiEnvelope<ResumeDocument>>('/portal/resume/resume', {
        public: true,
      })
    },
    getPreview(id: string) {
      return http.get<ApiEnvelope<ResumeDocument>>('/portal/resume/resume/preview', {
        params: { id },
      })
    },
    /** 纸面 HTML（结构性模板 / 与 PDF 同源） */
    getPublishedHtml() {
      return http.get<ApiEnvelope<{ html: string }>>('/portal/resume/resume/html', {
        public: true,
      })
    },
    getPreviewHtml(id: string) {
      return http.get<ApiEnvelope<{ html: string }>>('/portal/resume/resume/preview/html', {
        params: { id },
      })
    },
  }
}

export type ResumeApi = ReturnType<typeof createResumeApi>
