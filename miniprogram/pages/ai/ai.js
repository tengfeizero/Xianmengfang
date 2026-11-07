const api = require('../../utils/request')
const { img } = require('../../config')

Page({
  data: {
    mode: 'ai_text', // ai_text | ai_photo
    prompt: '',
    photoPath: '',
    generating: false,
    result: null, // {cover, type}
    quota: null, // {freeLimit, freeLeft, points}
    btnText: '开始生成',
  },

  onLoad() {
    this.loadQuota()
  },

  // 今日免费剩余 + 积分余额：免费用完就要花积分（1 积分 = 1 张）
  loadQuota() {
    api
      .request('GET', '/ai/quota')
      .then((d) => {
        const freeLeft = d.free_left || 0
        const points = d.points || 0
        let btnText = '开始生成'
        if (freeLeft > 0) btnText = '开始生成（今日还剩 ' + freeLeft + ' 张）'
        else if (points > 0) btnText = '开始生成（消耗 1 积分）'
        else btnText = '今日额度已用完'
        this.setData({
          quota: { freeLimit: d.free_limit, freeLeft, points },
          btnText,
        })
      })
      .catch(() => {})
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
          this.loadQuota()
        })
        .catch((e) => {
          // 免费用完且积分不足：引导去广场（作品被使用能得积分）
          const noQuota = /积分不足/.test(e.message)
          wx.showModal({
            title: noQuota ? '次数已用完' : '生成失败',
            content: e.message,
            showCancel: noQuota,
            confirmText: noQuota ? '去广场' : '知道了',
            success: (r) => {
              if (noQuota && r.confirm) wx.navigateTo({ url: '/pages/plaza/plaza' })
            },
          })
        })
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
