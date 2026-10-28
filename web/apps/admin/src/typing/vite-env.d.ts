/// <reference types="vite/client" />

declare module '@jiangbyte/shared/resume/styles'

interface ImportMetaEnv {
  readonly VITE_API_URL?: string
  readonly VITE_BASE_URL?: string
  readonly VITE_HOME_PATH: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
