/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** 接口前缀，默认 /api/v1（与 Go 后端一致） */
  readonly VITE_API_BASE: string
  readonly VITE_APP_TITLE: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
