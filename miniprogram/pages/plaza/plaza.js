const api = require('../../utils/request')
const { img } = require('../../config')
const { setTab } = require('../../utils/tabbar')

const TABS = [
  { key: 'latest', label: '最新' },
  { key: 'uses', label: '最常用' },
  { key: 'likes', label: '最多赞' },
]

Page({
  data: {
    tabs: TABS,
    tabKey: 'latest',
    items: [],
    page: 1,
    size: 10,
    total: 0,
    loading: true,
    loadingMore: false,
    hasMore: true,
    showTop: false,
  },

  onPageScroll(e) {
    const show = e.scrollTop > 300
    if (show !== this.data.showTop) this.setData({ showTop: show })
  },

  goTop() {
    wx.pageScrollTo({ scrollTop: 0, duration: 300 })
  },

  onLoad() {
    this.load(true)
  },

  onShow() {
    setTab(this, 1)
  },

  onPullDownRefresh() {
    this.load(true)
  },

  onReachBottom() {
    if (!this.data.hasMore || this.data.loadingMore) return
    this.load(false)
  },

  switchTab(e) {
    const key = e.currentTarget.dataset.key
    if (key === this.data.tabKey) return
    this.setData({ tabKey: key, items: [], page: 1, hasMore: true, loading: true })
    this.load(true)
  },

  // reset=true 重新拉第一页（切榜 / 下拉刷新）
  load(reset) {
    const page = reset ? 1 : this.data.page
    this.setData(reset ? { loading: true, loadingMore: false } : { loadingMore: true })
    api
      .request('GET', '/plaza?sort=' + this.data.tabKey + '&page=' + page + '&size=' + this.data.size)
      .then((d) => {
        const list = (d.items || []).map((w) => ({
          id: w.id,
          cover: img(w.thumbnail_url || w.image_url),
          prompt: w.prompt || '',
          typeText: w.type === 'collage' ? '拼装作品' : w.type === 'ai_text' ? 'AI 文生图' : '照片转风格',
          isAI: w.type !== 'collage',
          likeCount: w.like_count || 0,
          useCount: w.use_count || 0,
          liked: !!w.liked,
          isMine: !!w.is_mine,
          author: (w.author && w.author.nickname) || '匿名用户',
          avatar: img((w.author && w.author.avatar) || ''),
          pop: false,
        }))
        const items = reset ? list : this.data.items.concat(list)
        this.setData({
          items,
          total: d.total || 0,
          page: page + 1,
          loading: false,
          loadingMore: false,
          hasMore: items.length < (d.total || 0),
        })
      })
      .catch((e) => {
        this.setData({ loading: false, loadingMore: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
      .then(() => wx.stopPullDownRefresh())
  },

  goWork(e) {
    wx.navigateTo({ url: '/pages/work-detail/work-detail?id=' + e.currentTarget.dataset.id })
  },

  goSearch() {
    wx.navigateTo({ url: '/pages/search/search?tab=works' })
  },

  goWorks() {
    wx.switchTab({ url: '/pages/works/works' })
  },

  // 点赞/取消点赞：原地更新 + 心形弹一下，不整页重载（避免列表跳动）
  toggleLike(e) {
    const id = e.currentTarget.dataset.id
    const idx = this.data.items.findIndex((w) => w.id === id)
    if (idx < 0) return
    const item = this.data.items[idx]
    if (item.isMine) {
      wx.showToast({ title: '不能给自己的作品点赞', icon: 'none' })
      return
    }
    const method = item.liked ? 'DELETE' : 'POST'
    api
      .request(method, '/works/' + id + '/like')
      .then((d) => {
        const patch = {}
        patch['items[' + idx + '].liked'] = !item.liked
        patch['items[' + idx + '].likeCount'] = d.like_count
        if (!item.liked) patch['items[' + idx + '].pop'] = true
        this.setData(patch)
        if (!item.liked) {
          setTimeout(() => this.setData({ ['items[' + idx + '].pop']: false }), 450)
        }
      })
      .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
  },
})
