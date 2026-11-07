// 购物车角标：v2.0 起购物车不再是 tab，首页右上角展示数量，
// 数量落 storage，各页读取即可，不需要跨页通信。
const api = require('./request')

const KEY = 'xmf_cart_count'

function refreshBadge() {
  return api
    .request('GET', '/cart')
    .then((d) => {
      const n = (d.items || []).length
      wx.setStorageSync(KEY, n)
      return n
    })
    .catch(() => getCount())
}

function getCount() {
  return wx.getStorageSync(KEY) || 0
}

module.exports = { refreshBadge, getCount }
