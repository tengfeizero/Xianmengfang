// 金额分 → 元字符串
function fen2yuan(cents) {
  return (Number(cents || 0) / 100).toFixed(2)
}

// 订单状态文案与颜色
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
  pending_accept: '#ff9800',
  producing: '#409eff',
  packing: '#8b5cf6',
  pending_ship: '#14b8a6',
  shipped: '#8bc34a',
  done: '#999999',
  canceled: '#bbbbbb',
}

// 拼装元素分类文案
const ELEMENT_CATS = [
  { key: 'hair', label: '头发' },
  { key: 'hat', label: '帽子' },
  { key: 'lash', label: '睫毛' },
  { key: 'expression', label: '表情' },
  { key: 'outfit', label: '服装' },
  { key: 'accessory', label: '配饰' },
]

function statusText(s) {
  return STATUS_TEXT[s] || s
}
function statusColor(s) {
  return STATUS_COLOR[s] || '#999999'
}

module.exports = { fen2yuan, statusText, statusColor, ELEMENT_CATS }
