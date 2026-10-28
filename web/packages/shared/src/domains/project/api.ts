import type { AxiosInstance } from 'axios'
import type { ApiEnvelope } from '../user/account/types'
import type { ProjectCategoryNode, ProjectDetail, ProjectPage } from './types'

/** Portal 公开项目 API */
export function createProjectApi(http: AxiosInstance) {
  return {
    page(params?: { current?: number; size?: number; name?: string; category_id?: string }) {
      return http.get<ApiEnvelope<ProjectPage>>('/portal/project/project/page', {
        params,
        public: true,
      })
    },
    detail(params: { id?: string; slug?: string }) {
      return http.get<ApiEnvelope<ProjectDetail>>('/portal/project/project/detail', {
        params,
        public: true,
      })
    },
    categoryTree() {
      return http.get<ApiEnvelope<ProjectCategoryNode[]>>('/portal/project/category/tree', {
        public: true,
      })
    },
  }
}

export type ProjectApi = ReturnType<typeof createProjectApi>
