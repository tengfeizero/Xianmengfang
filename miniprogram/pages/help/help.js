const FAQS = [
  {
    q: '怎么把作品变成实物？',
    a: '在首页选一件商品 → 在商品页「使用我的作品」或「使用我点赞的作品」里选一张图 → 选规格下单。工厂接单后会按你选的图制作并发货。',
  },
  {
    q: '积分怎么获得、怎么用？',
    a: '你的作品被订单使用一次（按件数）就 +1 积分；积分可以在 AI 生成免费额度用完后抵扣，1 积分 = 1 张。',
  },
  {
    q: '为什么要点赞别人的作品？',
    a: '点赞是「借用」的前置条件：你点赞过的公开作品，可以在下单时选作定制图案。给自己点不了赞。',
  },
  {
    q: '作品发布到广场后还能取消吗？',
    a: '可以。在「我的作品」或作品详情里随时取消发布，作品会立刻从广场移除；已有的点赞与使用次数保留。',
  },
  {
    q: '定制商品可以退货吗？',
    a: '定制类商品按订单页提示：不支持七天无理由退货。若出现质量问题，请联系商家处理。',
  },
]

Page({
  data: { faqs: [], openIdx: -1 },

  onLoad() {
    this.setData({ faqs: FAQS })
  },

  toggle(e) {
    const idx = e.currentTarget.dataset.idx
    this.setData({ openIdx: this.data.openIdx === idx ? -1 : idx })
  },

  contact() {
    wx.setClipboardData({
      data: 'xianmengfang-support',
      success: () =>
        wx.showModal({
          title: '联系我们',
          content: '客服微信号 xianmengfang-support 已复制，添加时备注「现梦坊」。',
          showCancel: false,
          confirmColor: '#f2455f',
        }),
    })
  },
})
