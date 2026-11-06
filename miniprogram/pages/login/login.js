const api = require('../../utils/request')

Page({
  data: {
    openid: 'wx_dev_' + Math.floor(Math.random() * 100000),
    loading: false,
    realLoginFailed: false,
  },

  onLoad() {
    // 已登录直接走
    if (api.getToken()) {
      wx.switchTab({ url: '/pages/index/index' })
    }
  },

  // 微信一键登录：code 换 JWT（需后端配置 appid/secret）
  wxLogin() {
    this.setData({ loading: true })
    wx.login({
      success: (res) => {
        if (!res.code) {
          this.setData({ loading: false, realLoginFailed: true })
          return
        }
        api
          .request('POST', '/auth/login', { code: res.code })
          .then((d) => {
            api.setToken(d.token)
            wx.switchTab({ url: '/pages/index/index' })
          })
          .catch((e) => {
            // 本地开发未配 appid 时会失败，展示开发者登录入口
            this.setData({ loading: false, realLoginFailed: true })
            wx.showToast({ title: e.message, icon: 'none' })
          })
      },
      fail: () => {
        this.setData({ loading: false, realLoginFailed: true })
      },
    })
  },

  onOpenIDInput(e) {
    this.setData({ openid: e.detail.value })
  },

  // 开发者模拟登录（后端 XMF_DEV_LOGIN=1 时可用）
  devLogin() {
    if (!this.data.openid) {
      wx.showToast({ title: '请输入 openid', icon: 'none' })
      return
    }
    this.setData({ loading: true })
    api
      .request('POST', '/auth/dev_login', { openid: this.data.openid })
      .then((d) => {
        api.setToken(d.token)
        wx.switchTab({ url: '/pages/index/index' })
      })
      .catch((e) => {
        wx.showToast({ title: e.message, icon: 'none' })
      })
      .then(() => this.setData({ loading: false }))
  },
})
