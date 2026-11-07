const api = require('../../utils/request')
const { img } = require('../../config')

Page({
  data: {
    id: 0,
    work: null,
    cover: '',
    author: { nickname: '', avatar: '' },
    liked: false,
    isMine: false,
    likeCount: 0,
    useCount: 0,
    typeText: '',
    prompt: '',
    loading: true,
    liking: false,
  },

  onLoad(query) {
    this.workId = Number(query.id)
    this.load()
  },

  load() {
    api
      .request('GET', '/works/' + this.workId + '/view')
      .then((d) => {
        const w = d.work || {}
        this.setData({
          work: w,
          cover: img(w.thumbnail_url || w.image_url),
          author: d.author || {},
          liked: !!d.liked,
          isMine: !!d.is_mine,
          likeCount: w.like_count || 0,
          useCount: w.use_count || 0,
          prompt: w.prompt || '',
          typeText: w.type === 'collage' ? '拼装作品' : w.type === 'ai_text' ? 'AI 文生图' : '照片转风格',
          loading: false,
        })
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
  },

  preview() {
    if (this.data.cover) wx.previewImage({ urls: [this.data.cover] })
  },

  // 点赞：借用下单的前置条件；原地更新 + 弹跳
  toggleLike() {
    if (this.data.isMine || this.data.liking) return
    this.setData({ liking: true })
    const method = this.data.liked ? 'DELETE' : 'POST'
    api
      .request(method, '/works/' + this.workId + '/like')
      .then((d) => {
        this.setData({ liked: d.liked, likeCount: d.like_count, pop: !this.data.liked })
        if (!this.data.liked) setTimeout(() => this.setData({ pop: false }), 450)
      })
      .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
      .then(() => this.setData({ liking: false }))
  },

  copyPrompt() {
    if (!this.data.prompt) return
    wx.setClipboardData({
      data: this.data.prompt,
      success: () => wx.showToast({ title: '提示词已复制', icon: 'none' }),
    })
  },

  // 自己的作品：发布/取消发布
  togglePublish() {
    const toPublic = !(this.data.work && this.data.work.is_public === 1)
    const tips = toPublic
      ? '发布到灵感广场后，提示词也会公开，其他人可以点赞并借用它下单。'
      : '取消发布后作品从广场移除，已有的点赞和使用次数保留。'
    wx.showModal({
      title: toPublic ? '发布到广场' : '取消发布',
      content: tips,
      confirmColor: '#f2455f',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('PUT', '/works/' + this.workId + '/publish', { is_public: toPublic ? 1 : 0 })
          .then(() => {
            wx.showToast({ title: toPublic ? '已发布' : '已取消', icon: 'success' })
            this.load()
          })
          .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
      },
    })
  },

  del() {
    wx.showModal({
      title: '删除作品',
      content: '删除后不可恢复，广场上也会一并移除，确定吗？',
      confirmColor: '#e0362f',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('DELETE', '/works/' + this.workId)
          .then(() => {
            wx.showToast({ title: '已删除', icon: 'success' })
            setTimeout(() => wx.navigateBack(), 600)
          })
          .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
      },
    })
  },

  // 用作品下单：回首页挑商品，商品页里可选「我的作品/点赞的作品」
  useIt() {
    wx.switchTab({ url: '/pages/index/index' })
    setTimeout(() => {
      wx.showToast({ title: '选好商品后，在商品页选这张图', icon: 'none', duration: 2200 })
    }, 400)
  },
})
