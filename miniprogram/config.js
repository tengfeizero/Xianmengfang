// 服务端地址：本地开发用 localhost；上线前改为 https 域名
// 微信开发者工具需勾选「不校验合法域名」才能访问 http://localhost
const BASE = 'http://127.0.0.1:8080'

// 用户端接口前缀：后端按端分命名空间（小程序 /api/v1/user，后台 /api/v1/admin），
// 页面内只写 /orders 这类资源路径，前缀统一在这里加，换域名/换前缀只改这一处
const API = BASE + '/api/v1/user'

module.exports = {
  BASE,
  API,
  // 把 /static/xxx 这类相对地址转成完整图片地址
  img(url) {
    if (!url) return ''
    return url.startsWith('http') ? url : BASE + url
  },
}
