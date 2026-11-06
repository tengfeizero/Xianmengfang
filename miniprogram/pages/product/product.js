const api = require('../../utils/request')
const { fen2yuan } = require('../../utils/util')
const { img } = require('../../config')

Page({
  data: {
    product: null,
    skus: [],
    selectedSku: null,
    qty: 1,
    works: [], // 可绑定的作品
    selectedWork: null,
    addresses: [],
    selectedAddress: null,
    submitting: false,
    totalText: '--',
  },

  onLoad(query) {
    this.productId = Number(query.id)
    this.load()
    this.loadWorks()
  },

  // 从地址列表/编辑页返回时刷新地址
  onShow() {
    if (this.productId) this.loadAddresses()
  },

  load() {
    api
      .request('GET', '/products/' + this.productId)
      .then((d) => {
        const skus = (d.skus || []).map((s) => ({
          id: s.id,
          specsText: this.specsText(s.specs),
          priceText: fen2yuan(s.price_cents),
          stock: s.stock,
        }))
        this.setData({
          product: {
            id: d.product.id,
            title: d.product.title,
            description: d.product.description,
            cover: img(this.firstImg(d.product.images)),
          },
          skus,
          selectedSku: skus.length === 1 ? skus[0] : null,
        })
      })
      .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
  },

  firstImg(imagesJson) {
    try {
      return JSON.parse(imagesJson || '[]')[0] || ''
    } catch (e) {
      return ''
    }
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

  loadWorks() {
    api
      .request('GET', '/works')
      .then((list) => {
        this.setData({
          works: (list || []).map((w) => ({
            id: w.id,
            cover: img(w.thumbnail_url || w.image_url),
          })),
        })
      })
      .catch(() => {})
  },

  computeTotal() {
    const s = this.data.selectedSku
    if (!s) {
      this.setData({ totalText: '--' })
      return
    }
    this.setData({ totalText: fen2yuan(parseFloat(s.priceText) * 100 * this.data.qty) })
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
        // 默认选中默认地址；已选地址若已被删除则回退到默认地址
        const keep =
          this.data.selectedAddress && addresses.find((a) => a.id === this.data.selectedAddress.id)
        this.setData({
          addresses,
          selectedAddress: keep || addresses.find((a) => a.isDefault) || null,
        })
      })
      .catch(() => {})
  },

  chooseAddress() {
    wx.navigateTo({
      url: '/pages/address/list?select=1',
      events: {
        select: (addr) => this.setData({ selectedAddress: addr }),
      },
    })
  },

  manageAddress() {
    wx.navigateTo({ url: '/pages/address/list' })
  },

  pickSku(e) {
    const sku = this.data.skus.find((s) => s.id === e.currentTarget.dataset.id)
    this.setData({ selectedSku: sku })
    this.computeTotal()
  },

  pickWork(e) {
    const id = e.currentTarget.dataset.id
    const cur = this.data.selectedWork
    this.setData({ selectedWork: cur && cur.id === id ? null : this.data.works.find((w) => w.id === id) })
  },

  onQty(e) {
    const qty = Math.max(1, Math.min(99, Number(e.detail.value) || 1))
    this.setData({ qty })
    this.computeTotal()
  },

  submit() {
    if (!this.data.selectedAddress) {
      wx.showToast({ title: '请选择收货地址', icon: 'none' })
      return
    }
    if (!this.data.selectedSku) {
      wx.showToast({ title: '请先选择规格', icon: 'none' })
      return
    }
    this.setData({ submitting: true })
    const item = { sku_id: this.data.selectedSku.id, qty: this.data.qty }
    if (this.data.selectedWork) item.work_id = this.data.selectedWork.id
    api
      .request('POST', '/orders', { address_id: this.data.selectedAddress.id, items: [item] })
      .then((d) => {
        wx.showToast({ title: '下单成功', icon: 'success' })
        setTimeout(() => wx.switchTab({ url: '/pages/order-list/order-list' }), 800)
      })
      .catch((e) => wx.showModal({ title: '下单失败', content: e.message, showCancel: false }))
      .then(() => this.setData({ submitting: false }))
  },
})
