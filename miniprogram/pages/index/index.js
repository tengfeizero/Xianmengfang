const api = require('../../utils/request')
const { fen2yuan } = require('../../utils/util')
const { img } = require('../../config')

Page({
  data: {
    products: [],
    loading: true,
  },

  onShow() {
    if (!api.getToken()) {
      wx.navigateTo({ url: '/pages/login/login' })
      return
    }
    this.load()
  },

  load() {
    api
      .request('GET', '/products')
      .then((list) => {
        const products = (list || []).map((p) => ({
          id: p.id,
          title: p.title,
          description: p.description,
          cover: img(this.firstImg(p.images)),
          priceText: fen2yuan(this.minPrice(p.skus)),
        }))
        this.setData({ products, loading: false })
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
  },

  firstImg(imagesJson) {
    try {
      const arr = JSON.parse(imagesJson || '[]')
      return arr[0] || ''
    } catch (e) {
      return ''
    }
  },

  minPrice(skus) {
    if (!skus || !skus.length) return 0
    return Math.min.apply(
      null,
      skus.map((s) => s.price_cents)
    )
  },

  goProduct(e) {
    wx.navigateTo({ url: '/pages/product/product?id=' + e.currentTarget.dataset.id })
  },

  goStudio() {
    wx.switchTab({ url: '/pages/studio/studio' })
  },
})
