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
          priceCents: s.price_cents,
          stock: s.stock,
        }))
        const p = d.product
        // 默认选中第一个有货规格（全部无货时选第一个，提交会提示缺货）
        const firstSku = skus.find((s) => s.stock > 0) || skus[0] || null
        // 顶部价格跟随选中规格，避免"起价"与实际应付金额对不上
        const curCents = firstSku ? firstSku.priceCents : 0
        this.originalCents = p.original_price_cents || 0
        this.setData(
          {
            product: {
              id: p.id,
              title: p.title,
              subtitle: p.subtitle,
              images: this.parseList(p.images).map(img),
              priceText: fen2yuan(curCents),
              // 划线价只有高于当前售价才有"优惠"含义
              originalText: this.originalCents > curCents ? fen2yuan(this.originalCents) : '',
              unit: p.unit || '件',
              detailParagraphs: (p.detail || '')
                .split(/\n\s*\n/)
                .map((s) => s.trim())
                .filter(Boolean),
              detailImages: this.parseList(p.detail_images).map(img),
              serviceNotes: p.service_notes,
            },
            skus,
            selectedSku: firstSku,
          },
          () => this.computeTotal()
        )
      })
      .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
  },

  parseList(imagesJson) {
    try {
      return JSON.parse(imagesJson || '[]') || []
    } catch (e) {
      return []
    }
  },

  // 主图/详情图点开大图，可左右滑动
  previewImage(e) {
    const url = e.currentTarget.dataset.url
    if (!url) return
    const urls = this.data.product.images.concat(this.data.product.detailImages)
    wx.previewImage({ current: url, urls: urls.length ? urls : [url] })
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
    // 用分计算再换算，避免 9.9*100 这类浮点误差
    this.setData({ totalText: fen2yuan(s.priceCents * this.data.qty) })
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
    if (!sku) return
    if (sku.stock <= 0) {
      wx.showToast({ title: '该规格暂时缺货', icon: 'none' })
      return
    }
    // 顶部价格随选中规格变化
    const product = Object.assign({}, this.data.product)
    product.priceText = fen2yuan(sku.priceCents)
    product.originalText = this.originalCents > sku.priceCents ? fen2yuan(this.originalCents) : ''
    this.setData({ selectedSku: sku, product }, () => this.computeTotal())
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
    if (this.data.selectedSku.stock <= 0) {
      wx.showToast({ title: '该规格暂时缺货', icon: 'none' })
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
