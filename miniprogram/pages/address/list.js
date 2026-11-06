const api = require('../../utils/request')

Page({
  data: {
    addresses: [],
    selectMode: false, // 从商品页跳来选地址
    loading: true,
  },

  onLoad(query) {
    this.selectMode = query.select === '1'
    this.setData({ selectMode: this.selectMode })
    this.eventChannel = this.getOpenerEventChannel()
    this.load()
  },

  load() {
    api
      .request('GET', '/addresses')
      .then((list) => {
        this.setData({
          addresses: (list || []).map((a) => ({
            id: a.id,
            name: a.name,
            phone: a.phone,
            full: a.province + a.city + a.district + a.detail,
            isDefault: a.is_default === 1,
          })),
          loading: false,
        })
      })
      .catch((e) => {
        this.setData({ loading: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
  },

  // 选地址模式：点击即选中返回
  pick(e) {
    if (!this.selectMode) return
    const addr = this.data.addresses.find((a) => a.id === e.currentTarget.dataset.id)
    if (addr && this.eventChannel) {
      this.eventChannel.emit('select', addr)
      wx.navigateBack()
    }
  },

  setDefault(e) {
    const id = e.currentTarget.dataset.id
    api
      .request('POST', '/addresses/' + id + '/default')
      .then(() => {
        wx.showToast({ title: '已设为默认', icon: 'success' })
        this.load()
      })
      .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
  },

  del(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '删除地址',
      content: '确定删除这个收货地址吗？',
      confirmColor: '#ff5c7a',
      success: (r) => {
        if (!r.confirm) return
        api
          .request('DELETE', '/addresses/' + id)
          .then(() => {
            wx.showToast({ title: '已删除', icon: 'success' })
            this.load()
          })
          .catch((err) => wx.showToast({ title: err.message, icon: 'none' }))
      },
    })
  },

  edit(e) {
    wx.navigateTo({ url: '/pages/address/edit?id=' + e.currentTarget.dataset.id })
  },

  add() {
    const url = this.selectMode ? '/pages/address/edit?select=1' : '/pages/address/edit'
    wx.navigateTo({ url })
  },
})
