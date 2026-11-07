import { http } from '@/api/http'

export interface AdminUserRow {
  id: number
  nickname: string
  avatar: string
  status: number // 1 正常 0 禁用
  created_at: string
  order_count: number
  work_count: number
  points: number // 积分余额（v1.7）
}

export function listUsers(params?: { keyword?: string; status?: string }) {
  const q = new URLSearchParams()
  if (params?.keyword) q.set('keyword', params.keyword)
  if (params?.status !== undefined && params.status !== '') q.set('status', params.status)
  const qs = q.toString()
  return http.get<AdminUserRow[]>('/admin/users' + (qs ? `?${qs}` : ''))
}

export function setUserStatus(id: number, status: 0 | 1) {
  return http.put(`/admin/users/${id}/status`, { status })
}
