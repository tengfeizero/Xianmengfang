const api = require('../../utils/request')
const { fen2yuan } = require('../../utils/util')
const { img } = require('../../config')
const { getCount, refreshBadge } = require('../../utils/cartBadge')
const { setTab } = require('../../utils/tabbar')

const QUICK = [
  { icon: 'wand', label: 'AI 生图', color: '#6c5ce7', bg: '#f1eeff', url: '/pages/ai/ai' },
  { icon: 'palette', label: '拼装工坊', color: '#ff6fa5', bg: '#ffeef4', url: '/pages/studio/studio' },
  { icon: 'compass', label: '灵感广场', color: '#14b8a6', bg: '#e8f8f4', url: '/pages/plaza/plaza' },
  { icon: 'grid', label: '我的作品', color: '#f5a623', bg: '#fff6e4', url: '/pages/works/works' },
]

const TAB_PATHS = ['/pages/studio/studio', '/pages/plaza/plaza', '/pages/works/works']

Page({
  data: {
    products: [],
    plaza: [],
    quick: QUICK,
    cartCount: 0,
    loading: true,
  },

  onShow() {
    setTab(this, 0)
    if (!api.getToken()) {
      wx.navigateTo({ url: '/pages/login/login' })
      return
    }
    this.setData({ cartCount: getCount() })
    this.load()
    this.loadPlaza()
    refreshBadge().then((n) => this.setData({ cartCount: n }))
  },

  onPullDownRefresh() {
    this.load()
    this.loadPlaza()
    refreshBadge().then((n) => this.setData({ cartCount: n }))
  },

  load() {
    api
      .request('GET', '/products')
      .then((list) => {
        const products = (list || []).map((p) => {
          const min = this.minPrice(p.skus)
          return {
            id: p.id,
            title: p.title,
            subtitle: p.subtitle || p.description,
            cover: img(this.firstImg(p.images)),
            priceText: fen2yuan(min),
            // 划线价只在高于最低售价时展示，否则没有「优惠」含义
            originalText: p.original_price_cents > min ? fen2yuan(p.original_price_cents) : '',
          }
        })
        this.setData({ products, loading: false })
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
  },

  // 首页只取广场前 8 件做灵感横滑，失败不影响主流程
  loadPlaza() {
    api
      .request('GET', '/plaza?sort=likes&page=1&size=8')
      .then((d) => {
        this.setData({
          plaza: (d.items || []).map((w) => ({
            id: w.id,
            cover: img(w.thumbnail_url || w.image_url),
            likeCount: w.like_count || 0,
          })),
        })
      })
      .catch(() => {})
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

  tapQuick(e) {
    const item = QUICK[e.currentTarget.dataset.idx]
    if (!item) return
    if (TAB_PATHS.indexOf(item.url) >= 0) wx.switchTab({ url: item.url })
    else wx.navigateTo({ url: item.url })
  },

  goProduct(e) {
    wx.navigateTo({ url: '/pages/product/product?id=' + e.currentTarget.dataset.id })
  },

  goWork(e) {
    wx.navigateTo({ url: '/pages/work-detail/work-detail?id=' + e.currentTarget.dataset.id })
  },

  goStudio() {
    wx.switchTab({ url: '/pages/studio/studio' })
  },

  goPlaza() {
    wx.switchTab({ url: '/pages/plaza/plaza' })
  },

  goSearch() {
    wx.navigateTo({ url: '/pages/search/search' })
  },

  goCart() {
    wx.navigateTo({ url: '/pages/cart/cart' })
  },
})
