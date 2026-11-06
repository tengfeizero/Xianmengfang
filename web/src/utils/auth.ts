// token 唯一读写入口：http 层与 store 都走这里，避免循环依赖
const KEY = 'xmf_token'

export function getToken(): string {
  return localStorage.getItem(KEY) || ''
}

export function setToken(token: string): void {
  localStorage.setItem(KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(KEY)
}
