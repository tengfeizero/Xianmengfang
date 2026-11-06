let token = localStorage.getItem('xmf_token') || ''

export function getToken(): string {
  return token
}

export function setToken(t: string) {
  token = t
  localStorage.setItem('xmf_token', t)
}

export function clearToken() {
  token = ''
  localStorage.removeItem('xmf_token')
}

export interface Resp<T = unknown> {
  code: number
  msg: string
  data: T
}

export class ApiError extends Error {}

export async function api<T = unknown>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = 'Bearer ' + token
  const resp = await fetch('/api/v1' + path, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  const json = (await resp.json()) as Resp<T>
  if (json.code !== 0) throw new ApiError(json.msg)
  return json.data
}

export async function uploadImage(file: File): Promise<string> {
  const form = new FormData()
  form.append('file', file)
  const resp = await fetch('/api/v1/admin/upload', {
    method: 'POST',
    headers: { Authorization: 'Bearer ' + token },
    body: form,
  })
  const json = (await resp.json()) as Resp<{ url: string }>
  if (json.code !== 0) throw new ApiError(json.msg)
  return json.data.url
}

// 金额：分 → 元（展示）
export function yuan(cents: number): string {
  return (cents / 100).toFixed(2)
}

// 金额：元（输入）→ 分
export function cents(y: number): number {
  return Math.round(y * 100)
}
