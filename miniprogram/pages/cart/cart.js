const api = require('../../utils/request')
const { fen2yuan } = require('../../utils/util')
const { img } = require('../../config')
const { refreshBadge } = require('../../utils/cartBadge')

Page({
  data: {
    loading: true,
    items: [], // {id,title,specsText,qty,checked,valid,invalidReason,priceText,subtotalText,cover,workThumb}
    allChecked: false,
    checkedCount: 0,
    totalText: '0.00',
    selectedAddress: null,
    submitting: false,
  },

  onShow() {
    this.load()
    this.loadAddresses()
  },

  // 列表以服务端为唯一事实源：任何变更后整单重拉，前端不算金额
  load() {
    api
      .request('GET', '/cart')
      .then((d) => {
        const items = (d.items || []).map((it) => ({
          id: it.id,
          title: it.valid ? it.title : '商品已失效',
          specsText: this.specsText(it.specs),
          qty: it.qty,
          checked: it.checked === 1 && it.valid,
          valid: it.valid,
          invalidReason: it.invalid_reason || '',
          priceText: fen2yuan(it.price_cents),
          subtotalText: fen2yuan(it.price_cents * it.qty),
          cover: it.image_url ? img(it.image_url) : '',
          workThumb: it.work_thumb ? img(it.work_thumb) : '',
        }))
        const valid = items.filter((i) => i.valid)
        const summary = d.summary || { checked_count: 0, checked_cents: 0 }
        this.setData({
          items,
          loading: false,
          allChecked: valid.length > 0 && valid.every((i) => i.checked),
          checkedCount: summary.checked_count,
          totalText: fen2yuan(summary.checked_cents),
        })
        refreshBadge()
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
  },

  specsText(specsJson) {
    try {
      const o = JSON.parse(specsJson || '{}')
      return Object.keys(o)
        .map((k) => k + ': ' + o[k])
        .join(' / ')
    } catch (e) {
      return specsJson || ''
    }
  },

  loadAddresses() {
    api
      .request('GET', '/addresses')
      .then((list) => {
        const addresses = (list || []).map((a) => ({
          id: a.id,
          name: a.name,
          phone: a.phone,
          full: a.province + a.city + a.district + a.detail,
          isDefault: a.is_default === 1,
        }))
        const keep = this.data.selectedAddress && addresses.find((a) => a.id === this.data.selectedAddress.id)
        this.setData({
          selectedAddress: keep || addresses.find((a) => a.isDefault) || addresses[0] || null,
        })
      })
      .catch(() => {})
  },

  toggleItem(e) {
    const item = this.data.items.find((i) => i.id === e.currentTarget.dataset.id)
    if (!item || !item.valid) return
    api
      .request('PUT', '/cart/' + item.id, { checked: item.checked ? 0 : 1 })
      .then(() => this.load())
      .catch((e2) => wx.showToast({ title: e2.message, icon: 'none' }))
  },

  toggleAll() {
    api
      .request('PUT', '/cart', { checked: this.data.allChecked ? 0 : 1 })
      .then(() => this.load())
      .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
  },

  changeQty(e) {
    const item = this.data.items.find((i) => i.id === e.currentTarget.dataset.id)
    if (!item) return
    const qty = item.qty + Number(e.currentTarget.dataset.delta)
    if (qty < 1 || qty > 99) return
    api
      .request('PUT', '/cart/' + item.id, { qty })
      .then(() => this.load())
      .catch((e2) => wx.showToast({ title: e2.message, icon: 'none' }))
  },

  removeItem(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '移除商品',
      content: '确定把这件商品移出购物车吗？',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('DELETE', '/cart/' + id)
          .then(() => this.load())
          .catch((e2) => wx.showToast({ title: e2.message, icon: 'none' }))
      },
    })
  },

  goShop() {
    wx.switchTab({ url: '/pages/index/index' })
  },

  goProduct(e) {
    wx.navigateTo({ url: '/pages/product/product?id=' + e.currentTarget.dataset.id })
  },

  chooseAddress() {
    wx.navigateTo({
      url: '/pages/address/list?select=1',
      events: {
        select: (addr) => this.setData({ selectedAddress: addr }),
      },
    })
  },

  checkout() {
    if (this.data.checkedCount <= 0) {
      wx.showToast({ title: '请先勾选要结算的商品', icon: 'none' })
      return
    }
    if (!this.data.selectedAddress) {
      wx.showToast({ title: '请选择收货地址', icon: 'none' })
      return
    }
    const ids = this.data.items.filter((i) => i.valid && i.checked).map((i) => i.id)
    this.setData({ submitting: true })
    api
      .request('POST', '/cart/checkout', { address_id: this.data.selectedAddress.id, item_ids: ids })
      .then((d) => {
        wx.showToast({ title: '下单成功', icon: 'success' })
        setTimeout(() => wx.navigateTo({ url: '/pages/order-detail/order-detail?id=' + d.order.id }), 600)
      })
      .catch((e) => wx.showModal({ title: '结算失败', content: e.message, showCancel: false }))
      .then(() => {
        this.setData({ submitting: false })
        this.load()
      })
  },
})
