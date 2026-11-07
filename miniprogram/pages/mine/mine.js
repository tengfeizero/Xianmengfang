const api = require('../../utils/request')
const { setTab } = require('../../utils/tabbar')

// 订单入口：进行中=生产/打包/待发货合并，减少用户理解成本
const ORDER_ENTRIES = [
  { key: 'pending_accept', label: '待接单', icon: 'clock' },
  { key: 'producing', label: '进行中', icon: 'sparkles' },
  { key: 'shipped', label: '已发货', icon: 'truck' },
  { key: 'done', label: '已完成', icon: 'check' },
]

const CELLS = [
  { key: 'cart', label: '购物车', icon: 'cart', color: '#ff6fa5', bg: '#ffeef4' },
  { key: 'address', label: '收货地址', icon: 'pin', color: '#14b8a6', bg: '#e8f8f4' },
  { key: 'points', label: '积分明细', icon: 'bolt', color: '#f5a623', bg: '#fff6e4' },
  { key: 'help', label: '帮助与反馈', icon: 'help', color: '#3b82f6', bg: '#eaf1ff' },
  { key: 'settings', label: '设置', icon: 'settings', color: '#6c5ce7', bg: '#f1eeff' },
  { key: 'about', label: '关于现梦坊', icon: 'info', color: '#5d576e', bg: '#f3f1f7' },
]

Page({
  data: {
    user: { id: 0, nickname: '', avatar: '', points: 0 },
    stats: { works: 0, public_works: 0, liked: 0, cart_count: 0, orders: {} },
    orderEntries: ORDER_ENTRIES,
    cells: CELLS,
    version: 'v2.0.0',
  },

  onShow() {
    setTab(this, 4)
    if (!api.getToken()) {
      wx.navigateTo({ url: '/pages/login/login' })
      return
    }
    this.load()
  },

  onPullDownRefresh() {
    this.load()
  },

  load() {
    api
      .request('GET', '/me')
      .then((d) => {
        const stats = d.stats || {}
        const o = stats.orders || {}
        // 进行中 = 生产中 + 打包中 + 待发货
        const producing = (o.producing || 0) + (o.packing || 0) + (o.pending_ship || 0)
        const orderEntries = ORDER_ENTRIES.map((it) =>
          Object.assign({}, it, {
            count: it.key === 'producing' ? producing : o[it.key] || 0,
          })
        )
        this.setData({ user: d.user || {}, stats, orderEntries })
        wx.setStorageSync('xmf_cart_count', stats.cart_count || 0)
      })
      .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
      .then(() => wx.stopPullDownRefresh())
  },

  goOrderEntry(e) {
    wx.navigateTo({ url: '/pages/order-list/order-list?status=' + e.currentTarget.dataset.key })
  },

  goOrders() {
    wx.navigateTo({ url: '/pages/order-list/order-list' })
  },

  goPointsPage() {
    wx.navigateTo({ url: '/pages/points/points' })
  },

  goPlaza() {
    wx.switchTab({ url: '/pages/plaza/plaza' })
  },

  tapCell(e) {
    const key = e.currentTarget.dataset.key
    const map = {
      cart: '/pages/cart/cart',
      address: '/pages/address/list',
      points: '/pages/points/points',
      help: '/pages/help/help',
      settings: '/pages/settings/settings',
      about: '/pages/about/about',
    }
    if (map[key]) wx.navigateTo({ url: map[key] })
  },

  goWorks() {
    wx.switchTab({ url: '/pages/works/works' })
  },

  goLiked() {
    wx.navigateTo({ url: '/pages/search/search?tab=liked' })
  },

  // 编辑昵称：showModal editable（基础库 2.17+）
  editNickname() {
    wx.showModal({
      title: '修改昵称',
      editable: true,
      placeholderText: '1~16 个字',
      content: this.data.user.nickname || '',
      confirmColor: '#6c5ce7',
      success: (r) => {
        if (!r.confirm || !r.content || !r.content.trim()) return
        api
          .request('PUT', '/me', { nickname: r.content.trim() })
          .then(() => {
            wx.showToast({ title: '已保存', icon: 'success' })
            this.load()
          })
          .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
      },
    })
  },
})
