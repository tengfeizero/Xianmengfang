// 骨架屏：首屏用形态相同的灰块占位，避免「加载中...」造成的空白跳动
Component({
  properties: {
    type: { type: String, value: 'list' }, // list | grid | card
    count: { type: Number, value: 3 },
  },
  data: { items: [] },
  observers: {
    count: function (n) {
      this.setData({ items: new Array(n > 0 ? n : 1).fill(0) })
    },
  },
  lifetimes: {
    attached() {
      const n = this.data.count > 0 ? this.data.count : 1
      this.setData({ items: new Array(n).fill(0) })
    },
  },
})
