const { icon } = require('../utils/icons')

const INK = '#9a99a2'
const ACTIVE = '#f2455f'

const TABS = [
  { path: '/pages/index/index', text: '首页', icon: 'home' },
  { path: '/pages/plaza/plaza', text: '广场', icon: 'compass' },
  { path: '/pages/studio/studio', text: '创作', icon: 'plus', center: true },
  { path: '/pages/works/works', text: '作品', icon: 'grid' },
  { path: '/pages/mine/mine', text: '我的', icon: 'user' },
]

// 自定义 tabBar：中间「创作」用凸起渐变按钮，其余为线性图标 + 文字。
// 图标是运行时生成的 SVG data URI（见 utils/icons.js），选中态换色不用换图。
Component({
  data: {
    selected: 0,
    list: [],
    midSrc: '',
  },
  lifetimes: {
    attached() {
      this.setData({
        list: TABS.map((t) =>
          Object.assign({}, t, {
            src: icon(t.icon, INK),
            srcOn: icon(t.icon, ACTIVE, { sw: 2.2 }),
          })
        ),
        midSrc: icon('plus', '#ffffff', { sw: 2.6 }),
      })
    },
  },
  methods: {
    tap(e) {
      const idx = Number(e.currentTarget.dataset.index)
      if (idx === this.data.selected) return
      wx.switchTab({ url: TABS[idx].path })
    },
  },
})
