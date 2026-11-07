const api = require('../../utils/request')
const { img } = require('../../config')
const { setTab } = require('../../utils/tabbar')

const FILTERS = [
  { key: 'all', label: '全部' },
  { key: 'public', label: '已公开' },
  { key: 'private', label: '私密' },
]

Page({
  data: {
    works: [],
    shown: [],
    filters: FILTERS,
    filterKey: 'all',
    profile: { nickname: '', avatar: '' },
    points: 0,
    stats: { total: 0, published: 0, likes: 0 },
    loading: true,
  },

  onShow() {
    setTab(this, 3)
    this.load()
    this.loadMe()
  },

  onPullDownRefresh() {
    this.load()
    this.loadMe()
  },

  loadMe() {
    api
      .request('GET', '/me')
      .then((d) => {
        this.setData({
          profile: d.user || {},
          points: (d.user && d.user.points) || 0,
        })
      })
      .catch(() => {})
  },

  load() {
    api
      .request('GET', '/works')
      .then((list) => {
        const works = (list || []).map((w) => ({
          id: w.id,
          type: w.type,
          cover: img(w.thumbnail_url || w.image_url),
          isPublic: w.is_public === 1,
          likeCount: w.like_count || 0,
          useCount: w.use_count || 0,
          prompt: w.prompt || '',
          // 拼装作品没有提示词，公开后展示用类型文案
          typeText: w.type === 'collage' ? '拼装作品' : w.type === 'ai_text' ? 'AI 文生图' : '照片转风格',
        }))
        this.setData({
          works,
          loading: false,
          stats: {
            total: works.length,
            published: works.filter((w) => w.isPublic).length,
            likes: works.reduce((s, w) => s + w.likeCount, 0),
          },
        })
        this.applyFilter()
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
  },

  pickFilter(e) {
    const key = e.currentTarget.dataset.key
    if (key === this.data.filterKey) return
    this.setData({ filterKey: key })
    this.applyFilter()
  },

  applyFilter() {
    const k = this.data.filterKey
    const shown = this.data.works.filter((w) =>
      k === 'all' ? true : k === 'public' ? w.isPublic : !w.isPublic
    )
    this.setData({ shown })
  },

  goWork(e) {
    wx.navigateTo({ url: '/pages/work-detail/work-detail?id=' + e.currentTarget.dataset.id })
  },

  goPoints() {
    wx.navigateTo({ url: '/pages/points/points' })
  },

  preview(e) {
    const work = this.data.works.find((w) => w.id === e.currentTarget.dataset.id)
    if (work) wx.previewImage({ urls: [work.cover] })
  },

  // 发布/取消发布：发布后作品进广场，AI 提示词一并公开
  togglePublish(e) {
    const id = e.currentTarget.dataset.id
    const work = this.data.works.find((w) => w.id === id)
    if (!work) return
    const toPublic = !work.isPublic
    const tips = toPublic
      ? '发布后作品会出现在灵感广场，' +
        (work.prompt ? '提示词「' + work.prompt + '」也会公开，' : '') +
        '其他人可以点赞并借用它下单。'
      : '取消发布后作品从广场移除，已有的点赞和使用次数保留。'
    wx.showModal({
      title: toPublic ? '发布到广场' : '取消发布',
      content: tips,
      confirmColor: '#6c5ce7',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('PUT', '/works/' + id + '/publish', { is_public: toPublic ? 1 : 0 })
          .then(() => {
            wx.showToast({ title: toPublic ? '已发布' : '已取消', icon: 'success' })
            this.load()
          })
          .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
      },
    })
  },

  del(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '删除作品',
      content: '删除后不可恢复，广场上也会一并移除，确定吗？',
      confirmColor: '#f04452',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('DELETE', '/works/' + id)
          .then(() => {
            wx.showToast({ title: '已删除', icon: 'success' })
            this.load()
          })
          .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
      },
    })
  },

  goStudio() {
    wx.switchTab({ url: '/pages/studio/studio' })
  },
})
