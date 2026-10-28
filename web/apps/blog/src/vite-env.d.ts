/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** API 根路径，默认 `/api` */
  readonly VITE_API_BASE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
