// 与后端 models.go / handlers 对应的 DTO，只放类型不放逻辑

export interface AdminLoginResult {
  token: string
}

export interface Category {
  id: number
  name: string
  sort: number
  status: number
}

export interface Sku {
  id: number
  product_id: number
  specs: string
  price_cents: number
  stock: number
  code: string
}

export interface Product {
  id: number
  category_id: number
  title: string
  description: string
  status: number
  skus: Sku[]
}

export type OrderStatus = 'pending_accept' | 'pending_ship' | 'shipped' | 'done' | 'canceled'

export interface OrderInfo {
  id: number
  order_no: string
  status: OrderStatus
  total_cents: number
  user_id: number
  refund_flag: number
  cancel_reason: string
  receiver_name: string
  receiver_phone: string
  receiver_address: string
}

export interface OrderItem {
  id: number
  title: string
  specs: string
  qty: number
  price_cents: number
  work_id: number | null
}

export interface Shipment {
  company: string
  tracking_no: string
}

export interface AdminOrderRow {
  order: OrderInfo
  items: OrderItem[]
  shipment: Shipment | null
}

export interface Board {
  id: number
  name: string
  url: string
  width: number
  height: number
  status: number
}

export interface ElementItem {
  id: number
  category: string
  name: string
  image_url: string
  anchor: string
  sort: number
  status: number
}

export interface ConfigItem {
  key: string
  value: string
}
