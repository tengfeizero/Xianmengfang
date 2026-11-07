const api = require('../../utils/request')
const { fen2yuan } = require('../../utils/util')
const { img } = require('../../config')

// 「进行中」= 生产中+打包中+待发货（对用户是同一件事：都在商家手里）
const TABS = [
  { key: '', label: '全部' },
  { key: 'pending_accept', label: '待接单' },
  { key: 'active', label: '进行中' },
  { key: 'shipped', label: '已发货' },
  { key: 'done', label: '已完成' },
  { key: 'canceled', label: '已取消' },
]

Page({
  data: {
    tabs: TABS,
    tabIdx: 0,
    orders: [],
    loading: true,
  },

  onLoad(query) {
    if (query.status) {
      const idx = TABS.findIndex((t) => t.key === query.status)
      if (idx >= 0) this.setData({ tabIdx: idx })
    }
  },

  onShow() {
    this.load()
  },

  onPullDownRefresh() {
    this.load()
  },

  switchTab(e) {
    this.setData({ tabIdx: e.currentTarget.dataset.idx, loading: true, orders: [] })
    this.load()
  },

  load() {
    const key = TABS[this.data.tabIdx].key
    const qs = key && key !== 'active' ? '?status=' + key : ''
    api
      .request('GET', '/orders' + qs)
      .then((list) => {
        let orders = (list || []).map((d) => {
          const thumbs = (d.items || [])
            .map((it) => it.work_image_url || it.image_url)
            .filter(Boolean)
            .slice(0, 2)
            .map(img)
          return {
            id: d.order.id,
            orderNo: d.order.order_no,
            status: d.order.status,
            totalText: fen2yuan(d.order.total_cents),
            itemCount: (d.items || []).reduce((s, it) => s + it.qty, 0),
            thumbs,
            firstTitle:
              d.items && d.items.length
                ? d.items[0].title + (d.items[0].specs ? ' · ' + d.items[0].specs : '')
                : '',
            shipment: d.shipment
              ? { company: d.shipment.company, trackingNo: d.shipment.tracking_no }
              : null,
          }
        })
        if (key === 'active') {
          orders = orders.filter((o) => ['producing', 'packing', 'pending_ship'].includes(o.status))
        }
        this.setData({ orders, loading: false })
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
      .then(() => wx.stopPullDownRefresh())
  },

  confirm(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '确认收货',
      content: '请确认已收到商品',
      confirmColor: '#6c5ce7',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('POST', '/orders/' + id + '/confirm')
          .then(() => {
            wx.showToast({ title: '交易完成', icon: 'success' })
            this.load()
          })
          .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
      },
    })
  },

  cancel(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '取消订单',
      content: '仅待接单状态可取消，库存将自动恢复',
      confirmColor: '#f04452',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('POST', '/orders/' + id + '/cancel')
          .then(() => {
            wx.showToast({ title: '已取消', icon: 'success' })
            this.load()
          })
          .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
      },
    })
  },

  detail(e) {
    wx.navigateTo({ url: '/pages/order-detail/order-detail?id=' + e.currentTarget.dataset.id })
  },

  manageAddress() {
    wx.navigateTo({ url: '/pages/address/list' })
  },
})
