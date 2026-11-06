const api = require('../../utils/request')

Page({
  data: {
    id: 0, // >0 编辑模式
    selectMode: false,
    name: '',
    phone: '',
    region: ['', '', ''],
    detail: '',
    isDefault: false,
    saving: false,
  },

  onLoad(query) {
    this.selectMode = query.select === '1'
    this.setData({ selectMode: this.selectMode })
    if (query.id) {
      this.data.id = Number(query.id)
      this.loadOne(this.data.id)
    }
  },

  loadOne(id) {
    api
      .request('GET', '/addresses')
      .then((list) => {
        const a = (list || []).find((x) => x.id === id)
        if (!a) {
          wx.showToast({ title: '地址不存在', icon: 'none' })
          return
        }
        this.setData({
          name: a.name,
          phone: a.phone,
          region: [a.province, a.city, a.district],
          detail: a.detail,
          isDefault: a.is_default === 1,
        })
      })
      .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
  },

  onInput(e) {
    this.setData({ [e.currentTarget.dataset.field]: e.detail.value })
  },

  onRegion(e) {
    this.setData({ region: e.detail.value })
  },

  onDefault(e) {
    this.setData({ isDefault: e.detail.value })
  },

  save() {
    const d = this.data
    if (!d.name.trim()) return this.toast('请填写收货人')
    if (!/^1\d{10}$/.test(d.phone)) return this.toast('手机号格式不正确')
    if (!d.region[0]) return this.toast('请选择所在地区')
    if (!d.detail.trim()) return this.toast('请填写详细地址')
    this.setData({ saving: true })
    const body = {
      name: d.name.trim(),
      phone: d.phone,
      province: d.region[0],
      city: d.region[1],
      district: d.region[2],
      detail: d.detail.trim(),
      is_default: d.isDefault,
    }
    const done = (addr) => {
      wx.showToast({ title: '已保存', icon: 'success' })
      if (this.selectMode && this.eventChannelOf()) {
        this.eventChannelOf().emit('select', {
          id: addr.id,
          name: addr.name,
          phone: addr.phone,
          full: addr.province + addr.city + addr.district + addr.detail,
          isDefault: addr.is_default === 1,
        })
      }
      setTimeout(() => wx.navigateBack(), 500)
    }
    const p = d.id
      ? api.request('PUT', '/addresses/' + d.id, body)
      : api.request('POST', '/addresses', body)
    p.then(done)
      .catch((e) => wx.showModal({ title: '保存失败', content: e.message, showCancel: false }))
      .then(() => this.setData({ saving: false }))
  },

  eventChannelOf() {
    const ec = this.getOpenerEventChannel && this.getOpenerEventChannel()
    return ec && ec.emit ? ec : null
  },

  toast(msg) {
    wx.showToast({ title: msg, icon: 'none' })
  },
})
