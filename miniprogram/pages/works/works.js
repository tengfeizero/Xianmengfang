const api = require('../../utils/request')
const { img } = require('../../config')

Page({
  data: {
    works: [],
    points: 0,
    loading: true,
  },

  onShow() {
    this.load()
    this.loadPoints()
  },

  load() {
    api
      .request('GET', '/works')
      .then((list) => {
        this.setData({
          works: (list || []).map((w) => ({
            id: w.id,
            type: w.type,
            cover: img(w.thumbnail_url || w.image_url),
            isPublic: w.is_public === 1,
            likeCount: w.like_count || 0,
            useCount: w.use_count || 0,
            prompt: w.prompt || '',
            // 拼装作品没有提示词，公开后展示用类型文案
            typeText: w.type === 'collage' ? '拼装作品' : w.type === 'ai_text' ? 'AI 文生图' : '照片转风格',
          })),
          loading: false,
        })
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
  },

  loadPoints() {
    api
      .request('GET', '/points')
      .then((d) => this.setData({ points: d.points || 0 }))
      .catch(() => {})
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
      ? '发布后作品会出现在作品广场，' +
        (work.prompt ? '提示词「' + work.prompt + '」也会公开，' : '') +
        '其他人可以点赞并借用它下单。'
      : '取消发布后作品从广场移除，已有的点赞和使用次数保留。'
    wx.showModal({
      title: toPublic ? '发布到广场' : '取消发布',
      content: tips,
      confirmColor: '#ff5c7a',
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
      confirmColor: '#ff5c7a',
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

  goPlaza() {
    wx.navigateTo({ url: '/pages/plaza/plaza' })
  },

  goStudio() {
    wx.switchTab({ url: '/pages/studio/studio' })
  },
})
