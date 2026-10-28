export type ProjectListItem = {
  id: string
  name: string
  slug: string
  tech_stack: string[]
  role: string
  summary: string
  cover_object_name: string
  cover_url: string
  category_id: string | null
  category_name: string
  category_slug: string
  start_date: string | null
  end_date: string | null
  updated_at: string | null
}

export type ProjectDetail = ProjectListItem & {
  description: string
  images: string[]
  image_urls: string[]
}

export type ProjectPage = {
  records: ProjectListItem[]
  total: string
  current: string
  size: string
}

export type ProjectCategoryNode = {
  id: string
  parent_id: string | null
  name: string
  slug: string
  sort_code: number
  project_count: number
  children: ProjectCategoryNode[]
}
