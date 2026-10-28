export type StatsDayPoint = {
  date: string
  site_pv: number
  site_uv: number
  article_pv: number
}

export type StatsSummary = {
  pv: number
  uv: number
  article_pv?: number
  series: StatsDayPoint[]
}
