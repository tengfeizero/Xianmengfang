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
  subtitle: string
  description: string
  /** JSON 数组字符串：商品主图，首张为封面 */
  images: string
  /** 划线价（分），0 不展示，不参与结算 */
  original_price_cents: number
  unit: string
  /** 文字详情，空行分段 */
  detail: string
  /** JSON 数组字符串：详情长图 */
  detail_images: string
  service_notes: string
  sort: number
  status: number
  skus: Sku[]
}

export type OrderStatus =
  | 'pending_accept'
  | 'producing'
  | 'packing'
  | 'pending_ship'
  | 'shipped'
  | 'done'
  | 'canceled'

export interface OrderInfo {
  id: number
  order_no: string
  status: OrderStatus
  total_cents: number
  user_id: number
  pay_type: string
  refund_flag: number
  cancel_reason: string
  receiver_name: string
  receiver_phone: string
  receiver_address: string
  created_at: string
  paid_at: string | null
  shipped_at: string | null
  done_at: string | null
}

export interface OrderItem {
  id: number
  order_id: number
  sku_id: number
  title: string
  specs: string
  qty: number
  price_cents: number
  work_id: number | null
  /** 管理端详情接口返回：商品首图 */
  image_url?: string
  /** 管理端详情接口返回：绑定作品图 */
  work_image_url?: string
}

export interface Shipment {
  id: number
  order_id: number
  company: string
  tracking_no: string
  created_at: string
}

export interface OrderEvent {
  id: number
  order_id: number
  from: string
  to: string
  operator: string
  note: string
  created_at: string
}

export interface OrderAmounts {
  items_cents: number
  shipping_cents: number
  discount_cents: number
  total_cents: number
}

export interface OrderBuyer {
  id: number
  nickname: string
  avatar: string
}

export interface AdminOrderRow {
  order: OrderInfo
  items: OrderItem[]
  shipment: Shipment | null
  user?: OrderBuyer
}

/** 管理端订单详情（GET /admin/orders/:id） */
export interface AdminOrderDetail {
  order: OrderInfo
  items: OrderItem[]
  shipment: Shipment | null
  events: OrderEvent[]
  amounts: OrderAmounts
  user?: OrderBuyer
}

export interface Board {
  id: number
  name: string
  url: string
  width: number
  height: number
  status: number
  /** 该底板下启用元素数（后台列表返回） */
  element_count?: number
}

export interface ElementItem {
  id: number
  /** 所属底板（一对多），元素的画布与锚点按该底板计算 */
  board_id: number
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

// ---- RBAC 与审计日志（v1.4） ----

/** 登录态自检返回：权限以后端为准，前端只做展示层过滤 */
export interface AdminMe {
  id: number
  username: string
  nickname: string
  is_super: boolean
  role: { id: number; code: string; name: string }
  /** 超级管理员为 ["*"] */
  perms: string[]
}

export interface PermItem {
  code: string
  name: string
}

export interface PermGroup {
  key: string
  name: string
  items: PermItem[]
}

export interface RoleRow {
  id: number
  code: string
  name: string
  remark: string
  perms: string[]
  builtin: number
  status: number
  admin_count: number
}

export interface AdminRow {
  id: number
  username: string
  nickname: string
  role_id: number
  role_name: string
  role_code: string
  perms: string[]
  status: number
  last_login_at: string | null
  created_at: string
}

export interface AuditLogRow {
  id: number
  admin_id: number
  admin_name: string
  action: string
  target_type: string
  target_id: string
  method: string
  path: string
  ip: string
  user_agent: string
  detail: string
  result: 'ok' | 'fail' | 'deny'
  status_code: number
  err_msg: string
  created_at: string
}

export interface AuditPage {
  list: AuditLogRow[]
  total: number
  page: number
  size: number
}
