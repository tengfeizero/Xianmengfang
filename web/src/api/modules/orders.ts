import { http } from '@/api/http'
import type { AdminOrderDetail, AdminOrderRow } from '@/types'

export function listOrders(params?: { status?: string; order_no?: string; user_id?: string }) {
  const qs = new URLSearchParams()
  if (params?.status) qs.set('status', params.status)
  if (params?.order_no) qs.set('order_no', params.order_no)
  if (params?.user_id) qs.set('user_id', params.user_id)
  const s = qs.toString()
  return http.get<AdminOrderRow[]>('/admin/orders' + (s ? `?${s}` : ''))
}

export function getOrderDetail(id: number) {
  return http.get<AdminOrderDetail>(`/admin/orders/${id}`)
}

export type OrderAction = 'accept' | 'cancel' | 'refund_mark'

export function orderAction(id: number, action: OrderAction, body?: unknown) {
  return http.post(`/admin/orders/${id}/${action}`, body)
}

export function shipOrder(id: number, company: string, trackingNo: string) {
  return http.post(`/admin/orders/${id}/ship`, { company, tracking_no: trackingNo })
}
