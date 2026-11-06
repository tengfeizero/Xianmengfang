import { http } from '@/api/http'
import type { AdminLoginResult } from '@/types'

export function adminLogin(username: string, password: string) {
  return http.post<AdminLoginResult>('/admin/auth/login', { username, password })
}
