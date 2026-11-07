const api = require('../../utils/request')
const { fen2yuan } = require('../../utils/util')
const { img } = require('../../config')
const { getCount, refreshBadge } = require('../../utils/cartBadge')
const { setTab } = require('../../utils/tabbar')

const QUICK = [
  { icon: 'wand', label: 'AI 生图', color: '#f2455f', bg: '#fff1f3', url: '/pages/ai/ai' },
  { icon: 'palette', label: '拼装工坊', color: '#ff6fa5', bg: '#ffeef4', url: '/pages/studio/studio' },
  { icon: 'compass', label: '灵感广场', color: '#14b8a6', bg: '#e8f8f4', url: '/pages/plaza/plaza' },
  { icon: 'grid', label: '我的作品', color: '#f5a623', bg: '#fff6e4', url: '/pages/works/works' },
]

const TAB_PATHS = ['/pages/studio/studio', '/pages/plaza/plaza', '/pages/works/works']

const SORTS = [
  { key: 'hot', label: '最热' },
  { key: 'latest', label: '最新' },
  { key: 'price', label: '价格' },
]

Page({
  data: {
    products: [],
    quick: QUICK,
    sorts: SORTS,
    sortKey: 'hot',
    cartCount: 0,
    loading: true,
    showTop: false,
  },

  onPageScroll(e) {
    const show = e.scrollTop > 300
    if (show !== this.data.showTop) this.setData({ showTop: show })
  },

  goTop() {
    wx.pageScrollTo({ scrollTop: 0, duration: 300 })
  },

  onShow() {
    setTab(this, 0)
    if (!api.getToken()) {
      wx.navigateTo({ url: '/pages/login/login' })
      return
    }
    this.setData({ cartCount: getCount() })
    this.load()
    refreshBadge().then((n) => this.setData({ cartCount: n }))
  },

  onPullDownRefresh() {
    this.load()
    refreshBadge().then((n) => this.setData({ cartCount: n }))
  },

  load() {
    api
      .request('GET', '/products?sort=' + this.data.sortKey)
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

  pickSort(e) {
    const key = e.currentTarget.dataset.key
    if (key === this.data.sortKey) return
    this.setData({ sortKey: key, loading: true, products: [] })
    this.load()
  },

  goProduct(e) {
    wx.navigateTo({ url: '/pages/product/product?id=' + e.currentTarget.dataset.id })
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
