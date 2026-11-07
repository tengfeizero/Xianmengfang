// 金额分 → 元字符串
function fen2yuan(cents) {
  return (Number(cents || 0) / 100).toFixed(2)
}

// 订单状态文案与配色（v2.0：语义色统一走设计系统）
const STATUS_TEXT = {
  pending_accept: '待接单',
  producing: '生产中',
  packing: '打包中',
  pending_ship: '待发货',
  shipped: '已发货',
  done: '已完成',
  canceled: '已取消',
}
const STATUS_COLOR = {
  pending_accept: '#f59f00',
  producing: '#6c5ce7',
  packing: '#7a5cff',
  pending_ship: '#14b8a6',
  shipped: '#3b82f6',
  done: '#9a93a8',
  canceled: '#b8b1c4',
}
const STATUS_BG = {
  pending_accept: '#fff6e4',
  producing: '#f1eeff',
  packing: '#f1eeff',
  pending_ship: '#e8f8f4',
  shipped: '#eaf1ff',
  done: '#f3f1f7',
  canceled: '#f6f4f9',
}
// 订单详情页顶部进度条节点（canceled 不在链路里）
const STATUS_STEPS = ['pending_accept', 'producing', 'packing', 'pending_ship', 'shipped', 'done']

// 拼装元素分类文案
const ELEMENT_CATS = [
  { key: 'hair', label: '头发' },
  { key: 'hat', label: '帽子' },
  { key: 'lash', label: '睫毛' },
  { key: 'expression', label: '表情' },
  { key: 'outfit', label: '服装' },
  { key: 'accessory', label: '配饰' },
]

// 积分流水原因（与后端 point_logs.reason 对应）
const POINT_REASON = {
  work_used: { text: '作品被使用', icon: 'sparkles' },
  ai_spend: { text: 'AI 生成抵扣', icon: 'wand' },
  admin_adjust: { text: '后台调整', icon: 'settings' },
}

function statusText(s) {
  return STATUS_TEXT[s] || s
}
function statusColor(s) {
  return STATUS_COLOR[s] || '#9a93a8'
}
function statusBg(s) {
  return STATUS_BG[s] || '#f3f1f7'
}
function statusStep(s) {
  return STATUS_STEPS.indexOf(s)
}

function fmtTime(s) {
  if (!s) return ''
  return String(s).replace('T', ' ').slice(0, 16)
}

module.exports = {
  fen2yuan,
  statusText,
  statusColor,
  statusBg,
  statusStep,
  fmtTime,
  ELEMENT_CATS,
  POINT_REASON,
}
