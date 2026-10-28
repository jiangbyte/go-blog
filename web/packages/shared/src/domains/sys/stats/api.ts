import type { AxiosInstance } from 'axios'
import type { ApiEnvelope } from '../../user/account/types'
import type { StatsSummary } from './types'

/** 公开全站流量 API */
export function createStatsApi(http: AxiosInstance) {
  return {
    hit(visitorId: string) {
      return http.post<ApiEnvelope<{ ok: boolean; visitor_id: string }>>(
        '/portal/sys/stats/hit',
        { visitor_id: visitorId },
        { public: true },
      )
    },
    summary() {
      return http.get<ApiEnvelope<StatsSummary>>('/portal/sys/stats/summary', {
        public: true,
      })
    },
  }
}

export type StatsApi = ReturnType<typeof createStatsApi>
