import { http } from '@/api/http'
import type { AdminMe, AdminRow, AuditPage, PermGroup, RoleRow } from '@/types'

// ---- 登录态 ----

export function getMe() {
  return http.get<AdminMe>('/admin/me')
}

export function getPermCatalog() {
  return http.get<PermGroup[]>('/admin/permissions')
}

// ---- 角色 ----

export function listRoles() {
  return http.get<RoleRow[]>('/admin/roles')
}

export function createRole(body: { name: string; remark: string; perms: string[] }) {
  return http.post<RoleRow>('/admin/roles', body)
}

export function updateRole(id: number, body: { name: string; remark: string; perms: string[]; status: number }) {
  return http.put(`/admin/roles/${id}`, body)
}

export function deleteRole(id: number) {
  return http.del(`/admin/roles/${id}`)
}

// ---- 管理员账号 ----

export function listAdmins() {
  return http.get<AdminRow[]>('/admin/admins')
}

export function createAdmin(body: { username: string; password: string; nickname: string; role_id: number }) {
  return http.post<{ id: number }>('/admin/admins', body)
}

export function updateAdmin(
  id: number,
  body: { nickname: string; role_id: number; status: number; password?: string },
) {
  return http.put(`/admin/admins/${id}`, body)
}

export function deleteAdmin(id: number) {
  return http.del(`/admin/admins/${id}`)
}

// ---- 审计日志 ----

export function listAuditLogs(params: {
  admin_id?: number | ''
  action?: string
  result?: string
  keyword?: string
  page: number
  size: number
}) {
  const q = new URLSearchParams()
  if (params.admin_id) q.set('admin_id', String(params.admin_id))
  if (params.action) q.set('action', params.action)
  if (params.result) q.set('result', params.result)
  if (params.keyword) q.set('keyword', params.keyword)
  q.set('page', String(params.page))
  q.set('size', String(params.size))
  return http.get<AuditPage>('/admin/audit-logs?' + q.toString())
}
