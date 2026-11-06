/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** 接口前缀，默认 /api/v1（后台接口在 /admin 下，用户端在 /user 下） */
  readonly VITE_API_BASE: string
  readonly VITE_APP_TITLE: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
