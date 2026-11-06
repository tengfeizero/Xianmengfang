import { clearToken, getToken } from '@/utils/auth'

const BASE = import.meta.env.VITE_API_BASE || '/api/v1'

export interface Resp<T = unknown> {
  code: number
  msg: string
  data: T
}

export class ApiError extends Error {
  readonly status: number

  constructor(message: string, status = 0) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

/** 统一解析响应：code !== 0 或 401 都转成 ApiError，401 顺带清 token 回登录 */
async function parse<T>(resp: Response): Promise<T> {
  if (resp.status === 401) {
    clearToken()
    if (!location.hash.includes('/login')) location.hash = '#/login'
    throw new ApiError('登录已过期，请重新登录', 401)
  }
  let json: Resp<T>
  try {
    json = (await resp.json()) as Resp<T>
  } catch {
    throw new ApiError('服务返回格式异常', resp.status)
  }
  if (json.code !== 0) throw new ApiError(json.msg || '请求失败', resp.status)
  return json.data
}

function authHeaders(): Record<string, string> {
  const token = getToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export async function request<T = unknown>(method: string, path: string, body?: unknown): Promise<T> {
  const resp = await fetch(BASE + path, {
    method,
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  return parse<T>(resp)
}

export const http = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T = void>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T = void>(path: string, body?: unknown) => request<T>('PUT', path, body),
  del: <T = void>(path: string) => request<T>('DELETE', path),
}

/** 图片上传走 multipart，单独封装（不能带 Content-Type，交给浏览器生成 boundary） */
export async function uploadImage(file: File): Promise<string> {
  const form = new FormData()
  form.append('file', file)
  const resp = await fetch(BASE + '/admin/upload', { method: 'POST', headers: authHeaders(), body: form })
  const data = await parse<{ url: string }>(resp)
  return data.url
}
