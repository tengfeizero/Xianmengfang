const api = require('../../utils/request')
const { fmtTime, POINT_REASON } = require('../../utils/util')

const SIZE = 20

Page({
  data: {
    points: 0,
    logs: [],
    total: 0,
    page: 1,
    loading: true,
    loadingMore: false,
    hasMore: false,
  },

  onLoad() {
    this.load(true)
  },

  onPullDownRefresh() {
    this.load(true)
  },

  onReachBottom() {
    if (!this.data.hasMore || this.data.loadingMore) return
    this.load(false)
  },

  load(reset) {
    const page = reset ? 1 : this.data.page
    this.setData(reset ? { loading: true, loadingMore: false } : { loadingMore: true })
    api
      .request('GET', '/points?page=' + page + '&size=' + SIZE)
      .then((d) => {
        const list = (d.logs || []).map((l) => {
          const meta = POINT_REASON[l.reason] || { text: l.reason, icon: 'info' }
          return {
            id: l.id,
            delta: l.delta,
            deltaText: (l.delta > 0 ? '+' : '') + l.delta,
            balance: l.balance,
            title: meta.text,
            icon: meta.icon,
            remark: l.remark || '',
            time: fmtTime(l.created_at),
          }
        })
        const logs = reset ? list : this.data.logs.concat(list)
        this.setData({
          points: d.points || 0,
          total: d.total || logs.length,
          logs,
          page: page + 1,
          loading: false,
          loadingMore: false,
          hasMore: logs.length < (d.total || 0),
        })
      })
      .catch((e) => {
        this.setData({ loading: false, loadingMore: false })
        wx.showToast({ title: e.message, icon: 'none' })
      })
      .then(() => wx.stopPullDownRefresh())
  },
})
