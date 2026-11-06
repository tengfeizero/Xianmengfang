const { API } = require('../config')

const TOKEN_KEY = 'xmf_token'

function getToken() {
  return wx.getStorageSync(TOKEN_KEY) || ''
}

function setToken(t) {
  wx.setStorageSync(TOKEN_KEY, t)
}

function clearToken() {
  wx.removeStorageSync(TOKEN_KEY)
}

// 统一请求封装：Promise + JWT + 业务码处理
// 返回 resolve(data)；业务错误 reject(Error(msg))，http 层错误同样 reject
function request(method, path, data, opts) {
  opts = opts || {}
  return new Promise((resolve, reject) => {
    wx.request({
      url: API + path,
      method,
      data,
      timeout: 15000,
      header: Object.assign(
        { 'Content-Type': 'application/json' },
        getToken() ? { Authorization: 'Bearer ' + getToken() } : {}
      ),
      success(res) {
        // 上传等接口直接拿 body
        const body = res.data || {}
        if (res.statusCode === 401) {
          clearToken()
          wx.navigateTo({ url: '/pages/login/login' })
          reject(new Error('请先登录'))
          return
        }
        if (res.statusCode >= 200 && res.statusCode < 300 && body.code === 0) {
          resolve(body.data)
          return
        }
        reject(new Error(body.msg || '请求失败(' + res.statusCode + ')'))
      },
      fail() {
        reject(new Error('网络异常，请检查服务是否启动'))
      },
    })
  })
}

// multipart 上传（wx.uploadFile 返回的是字符串 body）
function uploadFile(filePath) {
  return new Promise((resolve, reject) => {
    wx.uploadFile({
      url: API + '/upload',
      filePath,
      name: 'file',
      header: getToken() ? { Authorization: 'Bearer ' + getToken() } : {},
      success(res) {
        let body
        try {
          body = JSON.parse(res.data)
        } catch (e) {
          reject(new Error('上传响应解析失败'))
          return
        }
        if (res.statusCode >= 200 && res.statusCode < 300 && body.code === 0) {
          resolve(body.data)
        } else {
          reject(new Error(body.msg || '上传失败'))
        }
      },
      fail() {
        reject(new Error('网络异常'))
      },
    })
  })
}

module.exports = { request, uploadFile, getToken, setToken, clearToken }
