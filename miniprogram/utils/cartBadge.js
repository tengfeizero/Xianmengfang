const api = require('./request')

// 刷新 tabBar 购物车角标（index:1）。未登录/网络失败静默。
function refreshBadge() {
  return api
    .request('GET', '/cart')
    .then((d) => {
      const n = (d.summary && d.summary.checked_count) || 0
      if (n > 0) {
        wx.setTabBarBadge({ index: 1, text: n > 99 ? '99+' : String(n), fail: () => {} })
      } else {
        wx.removeTabBarBadge({ index: 1, fail: () => {} })
      }
      return n
    })
    .catch(() => {})
}

module.exports = { refreshBadge }
