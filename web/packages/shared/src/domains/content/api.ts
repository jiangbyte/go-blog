import type { AxiosInstance } from 'axios'
import type { ApiEnvelope } from '../user/account/types'
import type {
  ContentArticleDetail,
  ContentArticleListItem,
  ContentArticleNeighbors,
  ContentArticlePage,
  ContentArticleSort,
  ContentCategoryNode,
  ContentTagOption,
} from './types'

/** Portal 公开内容 API */
export function createContentApi(http: AxiosInstance) {
  return {
    /** 有界列表：首页最新/最热/推荐，最多 20，非分页 */
    list(params?: { sort?: ContentArticleSort; limit?: number }) {
      return http.get<ApiEnvelope<ContentArticleListItem[]>>('/portal/content/article/list', {
        params,
        public: true,
      })
    },
    page(params?: {
      current?: number
      size?: number
      sort?: ContentArticleSort
      title?: string
      category_id?: string
      tag_id?: string
    }) {
      return http.get<ApiEnvelope<ContentArticlePage>>('/portal/content/article/page', {
        params,
        public: true,
      })
    },
    detail(params: { id?: string; slug?: string }) {
      return http.get<ApiEnvelope<ContentArticleDetail>>('/portal/content/article/detail', {
        params,
        public: true,
      })
    },
    neighbors(params: { id?: string; slug?: string }) {
      return http.get<ApiEnvelope<ContentArticleNeighbors>>('/portal/content/article/neighbors', {
        params,
        public: true,
      })
    },
    recordView(body: { id?: string; slug?: string }) {
      return http.post<ApiEnvelope<{ counted: boolean; view_count: number }>>(
        '/portal/content/article/record-view',
        body,
        { public: true },
      )
    },
    categoryTree() {
      return http.get<ApiEnvelope<ContentCategoryNode[]>>('/portal/content/category/tree', {
        public: true,
      })
    },
    tagOptions() {
      return http.get<ApiEnvelope<ContentTagOption[]>>('/portal/content/tag/options', {
        public: true,
      })
    },
    recommendList(limit = 10) {
      return http.get<ApiEnvelope<ContentArticleListItem[]>>('/portal/content/recommend/list', {
        params: { limit },
        public: true,
      })
    },
  }
}

export type ContentApi = ReturnType<typeof createContentApi>
