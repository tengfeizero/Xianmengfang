// 服务端地址：本地开发用 localhost；上线前改为 https 域名
// 微信开发者工具需勾选「不校验合法域名」才能访问 http://localhost
const BASE = 'http://127.0.0.1:8080'

module.exports = {
  BASE,
  // 把 /static/xxx 这类相对地址转成完整图片地址
  img(url) {
    if (!url) return ''
    return url.startsWith('http') ? url : BASE + url
  },
}
