/** Vite 客户端类型：styles/index.ts 使用 import.meta.glob 自动装载皮肤 CSS */
interface ImportMeta {
  readonly glob: (
    pattern: string,
    options?: { eager?: boolean; import?: string; query?: string },
  ) => Record<string, unknown>
}
