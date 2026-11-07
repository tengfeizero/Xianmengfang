const api = require('../../utils/request')
const { img } = require('../../config')
const { fen2yuan } = require('../../utils/util')

const HISTORY_KEY = 'xmf_search_history'
const HOT_WORDS = ['双马尾', '赛博朋克', '帆布包', '立牌', '水彩', '小兔子']

Page({
  data: {
    kw: '',
    tab: 'product', // product | works | liked
    hotWords: HOT_WORDS,
    history: [],
    products: [],
    allProducts: [],
    works: [],
    searching: false,
    searched: false,
    likedLoading: true,
  },

  onLoad(query) {
    const history = wx.getStorageSync(HISTORY_KEY) || []
    this.setData({ history, tab: query.tab === 'liked' ? 'liked' : query.tab === 'works' ? 'works' : 'product' })
    this.loadProducts()
    if (this.data.tab === 'liked') this.loadLiked()
  },

  loadProducts() {
    api
      .request('GET', '/products')
      .then((list) => {
        this.allProducts = (list || []).map((p) => ({
          id: p.id,
          title: p.title,
          subtitle: p.subtitle || p.description,
          cover: img(this.firstImg(p.images)),
          priceText: fen2yuan(this.minPrice(p.skus)),
        }))
        this.filterProducts()
      })
      .catch(() => {})
  },

  loadLiked() {
    this.setData({ likedLoading: true })
    api
      .request('GET', '/works?scope=liked')
      .then((list) => {
        this.setData({
          works: (list || []).map((w) => ({
            id: w.id,
            cover: img(w.thumbnail_url || w.image_url),
            prompt: w.prompt || '',
            likeCount: w.like_count || 0,
            useCount: w.use_count || 0,
            author: (w.author && w.author.nickname) || '匿名用户',
          })),
          likedLoading: false,
        })
      })
      .catch(() => this.setData({ likedLoading: false }))
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

  onInput(e) {
    const kw = e.detail.value
    this.setData({ kw })
    if (!kw.trim()) {
      this.setData({ products: this.data.allProducts, works: [], searched: false })
      return
    }
    if (this.data.tab === 'product') this.filterProducts()
  },

  onConfirm() {
    this.doSearch(this.data.kw)
  },

  tapWord(e) {
    const kw = e.currentTarget.dataset.kw
    this.setData({ kw })
    this.doSearch(kw)
  },

  clearHistory() {
    wx.removeStorageSync(HISTORY_KEY)
    this.setData({ history: [] })
  },

  tapClear() {
    this.setData({ kw: '', searched: false })
    this.filterProducts()
  },

  goBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: '/pages/index/index' }) })
  },

  goPlaza() {
    wx.switchTab({ url: '/pages/plaza/plaza' })
  },

  switchTab(e) {
    const tab = e.currentTarget.dataset.tab
    if (tab === this.data.tab) return
    this.setData({ tab, works: [], searched: false })
    if (tab === 'liked') this.loadLiked()
    else if (tab === 'product') this.filterProducts()
  },

  doSearch(kw) {
    kw = (kw || '').trim()
    if (!kw) return
    const history = [kw].concat((wx.getStorageSync(HISTORY_KEY) || []).filter((h) => h !== kw)).slice(0, 8)
    wx.setStorageSync(HISTORY_KEY, history)
    this.setData({ kw, history, searching: true, searched: true })
    if (this.data.tab === 'product') {
      this.filterProducts()
      this.setData({ searching: false })
      return
    }
    api
      .request('GET', '/plaza?sort=latest&keyword=' + encodeURIComponent(kw) + '&size=30')
      .then((d) => {
        this.setData({
          works: (d.items || []).map((w) => ({
            id: w.id,
            cover: img(w.thumbnail_url || w.image_url),
            prompt: w.prompt || '',
            likeCount: w.like_count || 0,
            useCount: w.use_count || 0,
            author: (w.author && w.author.nickname) || '匿名用户',
          })),
          searching: false,
        })
      })
      .catch(() => this.setData({ searching: false }))
  },

  filterProducts() {
    const kw = this.data.kw.trim()
    const all = this.allProducts || []
    this.setData({
      products: kw ? all.filter((p) => (p.title + (p.subtitle || '')).toLowerCase().includes(kw.toLowerCase())) : all,
    })
  },

  goProduct(e) {
    wx.navigateTo({ url: '/pages/product/product?id=' + e.currentTarget.dataset.id })
  },

  goWork(e) {
    wx.navigateTo({ url: '/pages/work-detail/work-detail?id=' + e.currentTarget.dataset.id })
  },
})
