#!/usr/bin/env node
// 一次性生成所有页面的预览数据文件 data/pages_*_*.js
const fs = require('fs')
const path = require('path')
const { art } = require('./mock')

const DIR = path.join(__dirname, 'data')
fs.mkdirSync(DIR, { recursive: true })

const w = (name, obj) =>
  fs.writeFileSync(
    path.join(DIR, name + '.js'),
    'const { art } = require("../mock")\nmodule.exports = ' +
      JSON.stringify(obj, null, 2).replace(/"__ART_(\d+)_(\d+)_(\d+)__"/g, (_, a, b, c) => `art(${a}, ${b}, ${c})`) +
      '\n'
  )
const A = (a, b, c) => `__ART_${a}_${b}_${c}__`

// 从真实页面源码里取顶层 `const XXX = [...]` 数组，避免预览数据与页面写两份而脱节
// （例如「我的」页的入口列表，改了页面却忘了改假数据，截图就看不出问题）
const fromPage = (relPath, name) => {
  const src = fs.readFileSync(path.join(__dirname, '../../miniprogram', relPath), 'utf8')
  const m = src.match(new RegExp('const ' + name + ' = (\\[[\\s\\S]*?\\n\\])'))
  if (!m) throw new Error(`未在 ${relPath} 中找到 const ${name} = [...]`)
  return new Function('return ' + m[1])()
}

const works = (n, off) =>
  Array.from({ length: n }, (_, i) => ({
    id: off + i,
    cover: A(off + i, 400, 500),
    prompt: i % 2 ? '扎双马尾的小姑娘，粉色背景，卡通风格，柔和光线' : '',
    typeText: i % 3 === 0 ? '拼装作品' : 'AI 文生图',
    isAI: i % 3 !== 0,
    likeCount: 3 + i * 5,
    useCount: i,
    liked: i % 4 === 0,
    isMine: false,
    author: ['桃桃', '阿岚', 'Momo', '小圆'][i % 4],
    avatar: '',
    pop: false,
  }))

// 广场
w('pages_plaza_plaza', {
  tabs: [
    { key: 'latest', label: '最新' },
    { key: 'uses', label: '最常用' },
    { key: 'likes', label: '最多赞' },
  ],
  tabKey: 'latest',
  items: works(6, 1),
  loading: false,
  hasMore: false,
  loadingMore: false,
})

// 作品
w('pages_works_works', {
  profile: { nickname: '桃桃', avatar: '' },
  points: 12,
  stats: { total: 8, published: 3, likes: 26 },
  filters: [
    { key: 'all', label: '全部' },
    { key: 'public', label: '已公开' },
    { key: 'private', label: '私密' },
  ],
  filterKey: 'all',
  shown: works(4, 20).map((x, i) => ({ ...x, isMine: true, isPublic: i % 2 === 0 })),
  loading: false,
})

// 创作
w('pages_studio_studio', {
  boards: [
    { id: 1, name: '小姑娘', thumb: A(31, 200, 250) },
    { id: 2, name: '小男孩', thumb: A(32, 200, 250) },
    { id: 3, name: '小兔子', thumb: A(33, 200, 250) },
  ],
  boardIdx: 0,
  cats: [
    { key: 'hair', label: '头发' },
    { key: 'hat', label: '帽子' },
    { key: 'outfit', label: '服装' },
    { key: 'expression', label: '表情' },
  ],
  catIdx: 0,
  elements: Array.from({ length: 8 }, (_, i) => ({ id: i + 1, name: '素材 ' + (i + 1), full: A(40 + i, 200, 200) })),
  placements: [],
  selectedIdx: -1,
  composing: false,
})

// 我的：入口与订单宫格直接取自页面源码，页面改了这里自动跟随
const mineOrderEntries = fromPage('pages/mine/mine.js', 'ORDER_ENTRIES')
const mineCells = fromPage('pages/mine/mine.js', 'CELLS')
const mineOrders = { pending_accept: 1, producing: 2, packing: 0, pending_ship: 0, shipped: 1, done: 4 }
w('pages_mine_mine', {
  user: { id: 3, nickname: '桃桃', avatar: '', points: 12 },
  stats: {
    works: 8,
    public_works: 3,
    liked: 5,
    cart_count: 2,
    orders: mineOrders,
  },
  orderEntries: mineOrderEntries.map((it) =>
    Object.assign({}, it, {
      count:
        it.key === 'producing'
          ? mineOrders.producing + mineOrders.packing + mineOrders.pending_ship
          : mineOrders[it.key] || 0,
    })
  ),
  cells: mineCells,
  version: 'v2.0.0',
})

// 商品详情
w('pages_product_product', {
  loading: false,
  galleryIdx: 0,
  cartCount: 2,
  product: {
    id: 1,
    title: '小梦女孩亚克力立牌 · 定制款',
    subtitle: '双面高清印刷，附透明底座',
    images: [A(51, 600, 600), A(52, 600, 600)],
    priceText: '39.00',
    originalText: '49.00',
    unit: '件',
    detailParagraphs: ['采用 5mm 高透明亚克力，UV 打印双面成像，边缘打磨圆润。', '支持任意作品图案定制，下单后 48 小时内发出。'],
    detailImages: [A(53, 700, 400)],
    serviceNotes: '48 小时内发货',
  },
  skus: [
    { id: 1, specsText: '10cm 单面', priceText: '39.00', stock: 10 },
    { id: 2, specsText: '15cm 双面', priceText: '59.00', stock: 5 },
  ],
  selectedSku: { id: 2, specsText: '15cm 双面', priceText: '59.00', stock: 5 },
  qty: 1,
  myWorks: [
    { id: 21, cover: A(54, 300, 380) },
    { id: 22, cover: A(55, 300, 380) },
    { id: 23, cover: A(56, 300, 380) },
  ],
  likedWorks: [
    { id: 31, cover: A(57, 300, 380) },
    { id: 32, cover: A(58, 300, 380) },
  ],
  selectedWork: { id: 22, cover: A(55, 300, 380) },
  addresses: [],
  selectedAddress: { id: 1, name: '桃桃', phone: '13800001111', full: '广东省 深圳市 南山区 科技园 1 号', isDefault: true },
  submitting: false,
  totalText: '59.00',
})

// 购物车
w('pages_cart_cart', {
  loading: false,
  items: [
    { id: 1, title: '小梦女孩亚克力立牌 · 定制款', specsText: '15cm 双面', qty: 2, checked: 1, valid: 1, invalidReason: '', priceText: '59.00', subtotalText: '118.00', cover: '', workThumb: A(61, 300, 300) },
    { id: 2, title: '定制帆布包 大容量单肩', specsText: '米白 / 12oz', qty: 1, checked: 1, valid: 1, invalidReason: '', priceText: '59.00', subtotalText: '59.00', cover: A(62, 300, 300), workThumb: '' },
    { id: 3, title: '已下架商品', specsText: '', qty: 1, checked: 0, valid: 0, invalidReason: '商品已下架，无法结算', priceText: '29.00', subtotalText: '29.00', cover: A(63, 300, 300), workThumb: '' },
  ],
  allChecked: false,
  checkedCount: 3,
  totalText: '177.00',
  selectedAddress: { id: 1, name: '桃桃', phone: '13800001111', full: '广东省 深圳市 南山区 科技园 1 号' },
  submitting: false,
})

// 订单列表
w('pages_order-list_order-list', {
  tabs: [
    { key: '', label: '全部' },
    { key: 'pending_accept', label: '待接单' },
    { key: 'active', label: '进行中' },
    { key: 'shipped', label: '已发货' },
    { key: 'done', label: '已完成' },
    { key: 'canceled', label: '已取消' },
  ],
  tabIdx: 0,
  loading: false,
  orders: [
    { id: 1, orderNo: 'XM20261107001', status: 'shipped', totalText: '177.00', itemCount: 3, thumbs: [A(71, 300, 300), A(72, 300, 300)], firstTitle: '小梦女孩亚克力立牌 · 15cm 双面', shipment: { company: '顺丰速运', trackingNo: 'SF20261107' } },
    { id: 2, orderNo: 'XM20261106002', status: 'producing', totalText: '59.00', itemCount: 1, thumbs: [A(73, 300, 300)], firstTitle: '定制帆布包 · 米白', shipment: null },
    { id: 3, orderNo: 'XM20261105003', status: 'done', totalText: '39.00', itemCount: 1, thumbs: [A(74, 300, 300)], firstTitle: '小梦女孩亚克力立牌 · 10cm', shipment: null },
  ],
})

// 订单详情
w('pages_order-detail_order-detail', {
  loading: false,
  steps: [
    { key: 'pending_accept', label: '待接单' },
    { key: 'producing', label: '生产中' },
    { key: 'packing', label: '打包中' },
    { key: 'pending_ship', label: '待发货' },
    { key: 'shipped', label: '已发货' },
    { key: 'done', label: '已完成' },
  ],
  order: {
    id: 1,
    orderNo: 'XM20261107001',
    status: 'shipped',
    statusText: '已发货',
    statusDesc: '商品已发出，请注意查收',
    stepIdx: 4,
    payType: '模拟支付（演示）',
    payTime: '2026-11-06 12:30',
    createdAt: '2026-11-06 12:28',
    cancelReason: '',
    refundFlag: 0,
    receiverName: '桃桃',
    receiverPhone: '13800001111',
    receiverAddress: '广东省深圳市南山区科技园 1 号 A 座 2201',
  },
  items: [
    { id: 1, title: '小梦女孩亚克力立牌 · 定制款', specsText: '15cm 双面', qty: 2, priceText: '59.00', subtotalText: '118.00', imageUrl: A(75, 300, 300), workImageUrl: A(76, 300, 300) },
    { id: 2, title: '定制帆布包 大容量单肩', specsText: '米白 / 12oz', qty: 1, priceText: '59.00', subtotalText: '59.00', imageUrl: A(77, 300, 300), workImageUrl: '' },
  ],
  events: [
    { id: 1, action: '商家发货', time: '2026-11-07 09:10', note: '顺丰速运 SF20261107' },
    { id: 2, action: '打包完成，等待发货', time: '2026-11-06 18:40', note: '' },
    { id: 3, action: '商家接单，开始生产', time: '2026-11-06 13:02', note: '' },
    { id: 4, action: '提交订单', time: '2026-11-06 12:28', note: '' },
  ],
  amounts: { itemsText: '177.00', shippingText: '0.00', discountText: '0.00', hasDiscount: false, totalText: '177.00' },
  shipment: { company: '顺丰速运', trackingNo: 'SF20261107', time: '2026-11-07 09:10' },
  canCancel: false,
  canConfirm: true,
})

// 积分
w('pages_points_points', {
  points: 12,
  total: 3,
  loading: false,
  logs: [
    { id: 1, delta: 2, deltaText: '+2', balance: 12, title: '作品被使用', icon: 'sparkles', remark: '', time: '2026-11-07 10:20' },
    { id: 2, delta: -1, deltaText: '-1', balance: 10, title: 'AI 生成抵扣', icon: 'wand', remark: '', time: '2026-11-06 21:02' },
    { id: 3, delta: 1, deltaText: '+1', balance: 11, title: '作品被使用', icon: 'sparkles', remark: '', time: '2026-11-05 09:14' },
  ],
  hasMore: false,
  loadingMore: false,
})

// 作品详情
w('pages_work-detail_work-detail', {
  loading: false,
  work: { id: 9, is_public: 1, type: 'ai_text' },
  cover: A(81, 700, 880),
  author: { nickname: '阿岚', avatar: '' },
  liked: false,
  isMine: false,
  likeCount: 36,
  useCount: 5,
  typeText: 'AI 文生图',
  prompt: '扎双马尾的小姑娘，站在粉色气球中间，柔和逆光，宫崎骏风格',
  pop: false,
})

// 搜索
w('pages_search_search', {
  kw: '',
  tab: 'product',
  hotWords: ['双马尾', '赛博朋克', '帆布包', '立牌', '水彩', '小兔子'],
  history: ['亚克力立牌', '小熊猫'],
  products: [
    { id: 1, title: '小梦女孩亚克力立牌 · 定制款', subtitle: '双面高清印刷，附透明底座', cover: A(91, 400, 400), priceText: '39.00' },
    { id: 2, title: '定制帆布包 大容量单肩', subtitle: '12oz 加厚 / 可水洗', cover: A(92, 400, 400), priceText: '59.00' },
  ],
  allProducts: [],
  works: [],
  searching: false,
  searched: false,
  likedLoading: false,
})

// AI
w('pages_ai_ai', {
  mode: 'ai_text',
  suggestions: ['扎双马尾的小姑娘', '水彩风格小猫', '赛博朋克城市夜景', '复古邮票插画'],
  prompt: '扎双马尾的小姑娘，粉色背景，卡通风格',
  photoPath: '',
  generating: false,
  result: { cover: A(95, 700, 700), type: 'ai_text' },
  quota: { freeLimit: 5, freeLeft: 2, points: 12 },
  btnText: '开始生成（今日还剩 2 张）',
})

// 登录 / 设置 / 帮助 / 关于 / 地址
w('pages_login_login', { openid: 'wx_dev_48321', loading: false, realLoginFailed: true })
w('pages_settings_settings', { notify: true, cacheText: '128 KB', version: 'v2.0.0' })
w('pages_help_help', {
  openIdx: 0,
  faqs: [
    { q: '怎么把作品变成实物？', a: '在首页选一件商品 → 在商品页「使用我的作品」里选一张图 → 选规格下单。' },
    { q: '积分怎么获得、怎么用？', a: '你的作品被订单使用一次（按件数）就 +1 积分；积分可以抵扣 AI 生成。' },
    { q: '为什么要点赞别人的作品？', a: '点赞是「借用」的前置条件：点赞过的公开作品，下单时可选作定制图案。' },
  ],
})
w('pages_about_about', { version: 'v2.0.0' })
w('pages_address_list', {
  loading: false,
  selectMode: false,
  addresses: [
    { id: 1, name: '桃桃', phone: '13800001111', full: '广东省 深圳市 南山区 科技园 1 号 A 座 2201', isDefault: true },
    { id: 2, name: '小圆', phone: '13900002222', full: '广东省 广州市 天河区 体育西路 8 号', isDefault: false },
  ],
})

console.log('data files written to', DIR)
