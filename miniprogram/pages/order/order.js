const api = require('../../utils/request')
const { fen2yuan, statusText, statusColor } = require('../../utils/util')

Page({
  data: {
    order: null,
  },

  onLoad(query) {
    api
      .request('GET', '/orders/' + query.id)
      .then((d) => {
        this.setData({
          order: {
            id: d.order.id,
            orderNo: d.order.order_no,
            status: d.order.status,
            statusLabel: statusText(d.order.status),
            statusColor: statusColor(d.order.status),
            totalText: fen2yuan(d.order.total_cents),
            payType: d.order.pay_type === 'mock' ? '模拟支付' : '微信支付',
            cancelReason: d.order.cancel_reason,
            createdAt: (d.order.created_at || '').replace('T', ' ').slice(0, 19),
            items: (d.items || []).map((it) => ({
              title: it.title,
              specsText: it.specs,
              qty: it.qty,
              priceText: fen2yuan(it.price_cents),
            })),
            shipment: d.shipment
              ? { company: d.shipment.company, trackingNo: d.shipment.tracking_no }
              : null,
          },
        })
      })
      .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
  },

  confirm() {
    wx.showModal({
      title: '确认收货',
      content: '请确认已收到商品',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('POST', '/orders/' + this.data.order.id + '/confirm')
          .then(() => {
            wx.showToast({ title: '交易完成', icon: 'success' })
            this.onLoad({ id: String(this.data.order.id) })
          })
          .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
      },
    })
  },
})
