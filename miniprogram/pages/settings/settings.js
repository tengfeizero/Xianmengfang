const api = require('../../utils/request')

Page({
  data: {
    notify: true,
    cacheText: '',
    version: 'v2.1.0',
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
      confirmColor: '#f2455f',
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

  // 注销账号：不可逆，二次确认后清空数据并回登录页
  deactivate() {
    wx.showModal({
      title: '注销账号',
      content: '注销后作品、点赞、购物车与收货地址将被清空且无法恢复；订单与积分记录会保留。确定注销？',
      confirmText: '注销',
      confirmColor: '#e0362f',
      success: (r) => {
        if (!r.confirm) return
        wx.showLoading({ title: '正在注销', mask: true })
        api
          .request('DELETE', '/me')
          .then(() => {
            wx.hideLoading()
            api.clearToken()
            wx.removeStorageSync('xmf_cart_count')
            wx.reLaunch({ url: '/pages/login/login' })
          })
          .catch((e) => {
            wx.hideLoading()
            wx.showToast({ title: e.message, icon: 'none' })
          })
      },
    })
  },

  logout() {
    wx.showModal({
      title: '退出登录',
      content: '退出后需要重新登录才能继续使用',
      confirmColor: '#e0362f',
      success: (r) => {
        if (!r.confirm) return
        api.clearToken()
        wx.reLaunch({ url: '/pages/login/login' })
      },
    })
  },
})
