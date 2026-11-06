const api = require('../../utils/request')
const { img } = require('../../config')

Page({
  data: {
    mode: 'ai_text', // ai_text | ai_photo
    prompt: '',
    photoPath: '',
    generating: false,
    result: null, // {cover, type}
  },

  switchMode(e) {
    this.setData({ mode: e.currentTarget.dataset.mode, result: null })
  },

  onPrompt(e) {
    this.setData({ prompt: e.detail.value })
  },

  choosePhoto() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        this.setData({ photoPath: res.tempFiles[0].tempFilePath, result: null })
      },
    })
  },

  generate() {
    if (this.data.mode === 'ai_text' && !this.data.prompt.trim()) {
      wx.showToast({ title: '请描述你想要的画面', icon: 'none' })
      return
    }
    if (this.data.mode === 'ai_photo' && !this.data.photoPath) {
      wx.showToast({ title: '请先选择照片', icon: 'none' })
      return
    }
    this.setData({ generating: true })
    const run = (extra) =>
      api
        .request('POST', '/ai/tasks', Object.assign({ type: this.data.mode }, extra))
        .then((d) => {
          this.setData({
            result: { cover: img(d.work.image_url), type: d.work.type },
          })
          wx.showToast({ title: '生成成功，已存入作品库', icon: 'success' })
        })
        .catch((e) => wx.showModal({ title: '生成失败', content: e.message, showCancel: false }))
        .then(() => this.setData({ generating: false }))

    if (this.data.mode === 'ai_text') {
      run({ prompt: this.data.prompt.trim() })
    } else {
      // 照片先上传，再发起转风格
      api
        .uploadFile(this.data.photoPath)
        .then((d) => run({ photo_url: d.url }))
        .catch((e) => {
          wx.showToast({ title: e.message, icon: 'none' })
          this.setData({ generating: false })
        })
    }
  },

  goWorks() {
    wx.switchTab({ url: '/pages/works/works' })
  },
})
