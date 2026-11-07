import { http } from '@/api/http'

// 明细类接口（v2.3）：积分流水 / 点赞 / 使用。
// 后端一次返回全量（上限 2000 条）+ 汇总，前端用 usePager 做本地分页，
// 与其他后台列表（用户、作品）保持同一套交互。

export interface PointLogRow {
  id: number
  user_id: number
  nickname: string
  delta: number // +获得 / -消耗
  balance: number // 变动后余额
  reason: string // work_used / ai_spend / admin_adjust
  reason_text: string
  ref_type: string // order / work / ai_task
  ref_id: number
  remark: string
  created_at: string
}

export interface PointLogResp {
  items: PointLogRow[]
  total: number
  summary: { in: number; out: number }
}

export function listPointLogs(params?: {
  user_id?: string | number
  reason?: string
  keyword?: string
}) {
  const q = new URLSearchParams()
  if (params?.user_id) q.set('user_id', String(params.user_id))
  if (params?.reason) q.set('reason', params.reason)
  if (params?.keyword) q.set('keyword', params.keyword)
  const qs = q.toString()
  return http.get<PointLogResp>('/admin/point-logs' + (qs ? `?${qs}` : ''))
}

export interface WorkLikeRow {
  id: number
  user_id: number
  nickname: string
  created_at: string
  work: {
    id: number
    type: string
    image_url: string
    thumbnail_url: string
    prompt: string
    status: number
    is_public: number
    author_id: number
    author_name: string
  }
}

export interface WorkLikeResp {
  items: WorkLikeRow[]
  total: number
}

export function listWorkLikes(params?: {
  work_id?: string | number
  user_id?: string | number
  keyword?: string
}) {
  const q = new URLSearchParams()
  if (params?.work_id) q.set('work_id', String(params.work_id))
  if (params?.user_id) q.set('user_id', String(params.user_id))
  if (params?.keyword) q.set('keyword', params.keyword)
  const qs = q.toString()
  return http.get<WorkLikeResp>('/admin/work-likes' + (qs ? `?${qs}` : ''))
}

export interface WorkUsageRow {
  id: number
  order_id: number
  order_no: string
  work_id: number
  work_image: string
  work_image_url: string
  work_type: string
  work_prompt: string
  author_id: number
  title: string
  specs: string
  qty: number
  user_id: number
  nickname: string
  status: string
  created_at: string
}

export interface WorkUsageResp {
  items: WorkUsageRow[]
  total: number
  summary: { qty: number }
}

export function listWorkUsages(params?: {
  work_id?: string | number
  user_id?: string | number
  status?: string
  keyword?: string
}) {
  const q = new URLSearchParams()
  if (params?.work_id) q.set('work_id', String(params.work_id))
  if (params?.user_id) q.set('user_id', String(params.user_id))
  if (params?.status) q.set('status', params.status)
  if (params?.keyword) q.set('keyword', params.keyword)
  const qs = q.toString()
  return http.get<WorkUsageResp>('/admin/work-usages' + (qs ? `?${qs}` : ''))
}
