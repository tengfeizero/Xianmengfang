const api = require('../../utils/request')
const { fen2yuan, statusText, statusColor } = require('../../utils/util')

const TABS = [
  { key: '', label: '全部' },
  { key: 'pending_accept', label: '待接单' },
  { key: 'pending_ship', label: '待发货' },
  { key: 'shipped', label: '已发货' },
  { key: 'done', label: '已完成' },
]

Page({
  data: {
    tabs: TABS,
    tabIdx: 0,
    orders: [],
    loading: true,
  },

  onShow() {
    this.load()
  },

  switchTab(e) {
    this.setData({ tabIdx: e.currentTarget.dataset.idx })
    this.load()
  },

  load() {
    const key = TABS[this.data.tabIdx].key
    const qs = key ? '?status=' + key : ''
    api
      .request('GET', '/orders' + qs)
      .then((list) => {
        const orders = (list || []).map((d) => ({
          id: d.order.id,
          orderNo: d.order.order_no,
          status: d.order.status,
          statusLabel: statusText(d.order.status),
          statusColor: statusColor(d.order.status),
          totalText: fen2yuan(d.order.total_cents),
          receiver: d.order.receiver_name
            ? d.order.receiver_name + ' ' + d.order.receiver_phone
            : '',
          receiverAddress: d.order.receiver_address || '',
          items: (d.items || []).map((it) => ({
            title: it.title,
            specsText: it.specs,
            qty: it.qty,
            priceText: fen2yuan(it.price_cents),
          })),
          shipment: d.shipment
            ? { company: d.shipment.company, trackingNo: d.shipment.tracking_no }
            : null,
        }))
        this.setData({ orders, loading: false })
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
  },

  confirm(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '确认收货',
      content: '请确认已收到商品',
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
      confirmColor: '#ff5c7a',
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
    wx.navigateTo({ url: '/pages/order/order?id=' + e.currentTarget.dataset.id })
  },

  manageAddress() {
    wx.navigateTo({ url: '/pages/address/list' })
  },
})
