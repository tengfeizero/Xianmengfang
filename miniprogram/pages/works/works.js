const api = require('../../utils/request')
const { img } = require('../../config')

Page({
  data: {
    works: [],
    loading: true,
  },

  onShow() {
    this.load()
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
          })),
          loading: false,
        })
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
  },

  preview(e) {
    const work = this.data.works.find((w) => w.id === e.currentTarget.dataset.id)
    if (work) wx.previewImage({ urls: [work.cover] })
  },

  del(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '删除作品',
      content: '删除后不可恢复，确定吗？',
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

  goStudio() {
    wx.switchTab({ url: '/pages/studio/studio' })
  },
})
