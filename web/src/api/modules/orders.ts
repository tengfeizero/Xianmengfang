import { http } from '@/api/http'
import type { AdminOrderRow } from '@/types'

export function listOrders(status?: string) {
  return http.get<AdminOrderRow[]>('/admin/orders' + (status ? `?status=${status}` : ''))
}

export type OrderAction = 'accept' | 'cancel' | 'refund_mark'

export function orderAction(id: number, action: OrderAction, body?: unknown) {
  return http.post(`/admin/orders/${id}/${action}`, body)
}

export function shipOrder(id: number, company: string, trackingNo: string) {
  return http.post(`/admin/orders/${id}/ship`, { company, tracking_no: trackingNo })
}
