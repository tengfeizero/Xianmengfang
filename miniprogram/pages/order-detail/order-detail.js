const api = require('../../utils/request')
const { fen2yuan, statusText } = require('../../utils/util')
const { img } = require('../../config')

// 状态一句话说明（给用户看的，不用状态机术语）
const STATUS_DESC = {
  pending_accept: '等待商家接单',
  pending_ship: '商家已接单，正在打包',
  shipped: '商品已发出，请注意查收',
  done: '交易已完成，感谢你的支持',
  canceled: '订单已取消，库存已恢复',
}

// 订单流水：to 状态 → 动作文案
const EVENT_TEXT = {
  pending_accept: '提交订单',
  pending_ship: '商家接单',
  shipped: '商家发货',
  done: '确认收货',
  canceled: '订单取消',
}

function fmtTime(s) {
  if (!s) return ''
  return String(s).replace('T', ' ').slice(0, 19)
}

Page({
  data: {
    order: null,
    items: [],
    events: [],
    amounts: {},
    shipment: null,
    loading: true,
    canCancel: false,
    canConfirm: false,
  },

  onLoad(query) {
    this.orderId = query.id
    this.load()
  },

  onPullDownRefresh() {
    this.load(true)
  },

  load(fromPull) {
    const done = () => {
      if (fromPull) wx.stopPullDownRefresh()
    }
    api
      .request('GET', '/orders/' + this.orderId)
      .then((d) => this.setData(this.build(d), done))
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
        done()
      })
  },

  // 接口返回 → 页面展示结构（金额转元、时间格式化、图片拼域名）
  build(d) {
    const o = d.order
    const a = d.amounts || {}
    const items = (d.items || []).map((it) => {
      const cover = it.image_url ? img(it.image_url) : ''
      const work = it.work_image_url ? img(it.work_image_url) : ''
      return {
        id: it.id,
        title: it.title,
        specsText: it.specs || '',
        qty: it.qty,
        priceText: fen2yuan(it.price_cents),
        subtotalText: fen2yuan(it.price_cents * it.qty),
        // 商品图缺失时用定制作品图当主图（演示商品没配图，绑作品的订单仍有图可看）
        imageUrl: cover || work,
        workImageUrl: cover && work ? work : '',
      }
    })
    // 流水倒序：最新在最上面
    const events = (d.events || [])
      .map((ev) => ({
        id: ev.id,
        action: EVENT_TEXT[ev.to] || statusText(ev.to),
        time: fmtTime(ev.created_at),
        note: ev.note || '',
      }))
      .reverse()
    const s = d.shipment
    const totalCents = a.total_cents === undefined ? o.total_cents : a.total_cents
    return {
      order: {
        id: o.id,
        orderNo: o.order_no,
        status: o.status,
        statusText: statusText(o.status),
        statusDesc: STATUS_DESC[o.status] || '',
        payType: o.pay_type === 'wxpay' ? '微信支付' : '模拟支付（演示）',
        payTime: fmtTime(o.paid_at),
        createdAt: fmtTime(o.created_at),
        cancelReason: o.cancel_reason || '',
        refundFlag: o.refund_flag === 1,
        receiverName: o.receiver_name || '',
        receiverPhone: o.receiver_phone || '',
        receiverAddress: o.receiver_address || '',
      },
      items,
      events,
      amounts: {
        itemsText: fen2yuan(a.items_cents),
        shippingText: fen2yuan(a.shipping_cents),
        discountText: fen2yuan(a.discount_cents),
        hasDiscount: (a.discount_cents || 0) > 0,
        totalText: fen2yuan(totalCents),
      },
      shipment: s ? { company: s.company, trackingNo: s.tracking_no, time: fmtTime(s.created_at) } : null,
      canCancel: o.status === 'pending_accept',
      canConfirm: o.status === 'shipped',
      loading: false,
    }
  },

  copy(e) {
    const text = e.currentTarget.dataset.text
    if (!text) return
    wx.setClipboardData({
      data: String(text),
      success: () => wx.showToast({ title: '已复制', icon: 'none' }),
    })
  },

  preview(e) {
    const src = e.currentTarget.dataset.src
    if (src) wx.previewImage({ urls: [src], current: src })
  },

  confirm() {
    wx.showModal({
      title: '确认收货',
      content: '请确认已收到商品',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('POST', '/orders/' + this.orderId + '/confirm')
          .then(() => {
            wx.showToast({ title: '交易完成', icon: 'success' })
            this.load()
          })
          .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
      },
    })
  },

  cancel() {
    wx.showModal({
      title: '取消订单',
      content: '仅待接单状态可取消，库存将自动恢复',
      confirmColor: '#ff5c7a',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('POST', '/orders/' + this.orderId + '/cancel')
          .then(() => {
            wx.showToast({ title: '已取消', icon: 'success' })
            this.load()
          })
          .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
      },
    })
  },
})
