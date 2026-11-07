const api = require('../../utils/request')

Page({
  data: {
    notify: true,
    cacheText: '',
    version: 'v2.0.0',
  },

  onLoad() {
    this.setData({ notify: wx.getStorageSync('xmf_notify') !== 0 })
    this.calcCache()
  },

  calcCache() {
    try {
      const info = wx.getStorageInfoSync()
      this.setData({ cacheText: info.currentSize + ' KB' })
    } catch (e) {
      this.setData({ cacheText: '0 KB' })
    }
  },

  onNotify(e) {
    this.setData({ notify: e.detail.value })
    wx.setStorageSync('xmf_notify', e.detail.value ? 1 : 0)
  },

  clearCache() {
    wx.showModal({
      title: '清除本地缓存',
      content: '搜索历史等本地数据将被清空，登录状态保留。',
      confirmColor: '#6c5ce7',
      success: (r) => {
        if (!r.confirm) return
        try {
          wx.removeStorageSync('xmf_search_history')
        } catch (e) {
          // 存储为空时忽略
        }
        wx.showToast({ title: '已清除', icon: 'success' })
        this.calcCache()
      },
    })
  },

  goAbout() {
    wx.navigateTo({ url: '/pages/about/about' })
  },

  goHelp() {
    wx.navigateTo({ url: '/pages/help/help' })
  },

  logout() {
    wx.showModal({
      title: '退出登录',
      content: '退出后需要重新登录才能继续使用',
      confirmColor: '#f04452',
      success: (r) => {
        if (!r.confirm) return
        api.clearToken()
        wx.reLaunch({ url: '/pages/login/login' })
      },
    })
  },
})
