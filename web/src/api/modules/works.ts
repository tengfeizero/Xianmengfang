import { http } from '@/api/http'

export interface AdminWorkRow {
  id: number
  user_id: number
  nickname: string
  type: 'collage' | 'ai_text' | 'ai_photo'
  image_url: string
  thumbnail_url: string
  meta: string // JSON 字符串，ai 类含 prompt
  prompt: string // 公开提示词（v1.7；发布到广场时展示）
  is_public: number // 1 已发布 0 私密
  like_count: number
  use_count: number
  status: number // 1 正常 0 已删除 2 被下架
  created_at: string
}

export function listWorks(params?: { type?: string; status?: string; public?: string }) {
  const q = new URLSearchParams()
  if (params?.type) q.set('type', params.type)
  if (params?.status) q.set('status', params.status)
  if (params?.public) q.set('public', params.public)
  const qs = q.toString()
  return http.get<AdminWorkRow[]>('/admin/works' + (qs ? `?${qs}` : ''))
}

export function setWorkStatus(id: number, status: 1 | 2) {
  return http.put(`/admin/works/${id}/status`, { status })
}
