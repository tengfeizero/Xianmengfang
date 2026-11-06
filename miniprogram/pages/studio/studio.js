const api = require('../../utils/request')
const { ELEMENT_CATS } = require('../../utils/util')
const { img } = require('../../config')

Page({
  data: {
    boards: [],
    boardIdx: 0,
    cats: ELEMENT_CATS,
    catIdx: 0,
    elements: [], // 当前分类的元素
    placements: [], // {elementId, name, x, y, scale}（坐标为相对底板的 0~1）
    selectedIdx: -1,
    scaleVal: 1,
    composing: false,
  },

  onReady() {
    this.imgCache = {}
    this.initCanvas()
    this.loadBoards()
  },

  onShow() {
    // 离开再回来（如从 AI 页返回）重绘一次
    if (this.ctx && this.data.placements.length) this.draw()
  },

  // ---- canvas 初始化 ----
  initCanvas() {
    const query = wx.createSelectorQuery()
    query
      .select('#stage')
      .fields({ node: true, size: true })
      .exec((res) => {
        if (!res || !res[0]) return
        const canvas = res[0].node
        const dpr = wx.getWindowInfo().pixelRatio || 2
        canvas.width = res[0].width * dpr
        canvas.height = res[0].height * dpr
        const ctx = canvas.getContext('2d')
        ctx.scale(dpr, dpr)
        this.canvas = canvas
        this.ctx = ctx
        this.cssW = res[0].width
        this.cssH = res[0].height
      })
  },

  // ---- 数据加载 ----
  loadBoards() {
    api
      .request('GET', '/collage/boards')
      .then((boards) => {
        this.setData({
          boards: (boards || []).map((b) => ({
            id: b.id,
            name: b.name,
            url: b.url,
            width: b.width || 400,
            height: b.height || 500,
            thumb: img(b.url),
          })),
        })
        if (this.data.boards.length) this.pickBoard({ detail: { value: 0 } })
      })
      .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
  },

  loadElements(catKey) {
    api
      .request('GET', '/collage/elements?category=' + catKey)
      .then((list) => {
        this.setData({
          elements: (list || []).map((el) => {
            let anchor = { x: 0.5, y: 0.5, scale: 1 }
            try {
              anchor = JSON.parse(el.anchor || '{}')
            } catch (e) {}
            return {
              id: el.id,
              name: el.name,
              url: el.image_url,
              full: img(el.image_url),
              anchor,
            }
          }),
        })
      })
      .catch((e) => wx.showToast({ title: e.message, icon: 'none' }))
  },

  // ---- 交互 ----
  pickBoard(e) {
    const idx = Number(e.currentTarget.dataset.idx)
    if (!(idx >= 0) || idx >= this.data.boards.length || idx === this.data.boardIdx) return
    this.setData({ boardIdx: idx, placements: [], selectedIdx: -1 })
    this.board = this.data.boards[idx]
    this.loadElements(this.data.cats[this.data.catIdx].key)
    this.draw()
  },

  pickCat(e) {
    const idx = e.currentTarget.dataset.idx
    this.setData({ catIdx: idx })
    this.loadElements(this.data.cats[idx].key)
  },

  // 点选元素 → 加到底板上（用素材默认锚点）
  addElement(e) {
    const el = this.data.elements.find((x) => x.id === e.currentTarget.dataset.id)
    if (!el) return
    if (this.data.placements.length >= 12) {
      wx.showToast({ title: '最多添加 12 个元素', icon: 'none' })
      return
    }
    const placements = this.data.placements.concat([
      { elementId: el.id, name: el.name, url: el.url, x: el.anchor.x || 0.5, y: el.anchor.y || 0.5, scale: el.anchor.scale || 1 },
    ])
    this.setData({ placements, selectedIdx: placements.length - 1, scaleVal: el.anchor.scale || 1 })
    this.ensureImg(el.url, () => this.draw())
    this.draw()
  },

  removeSelected() {
    if (this.data.selectedIdx < 0) return
    const placements = this.data.placements.slice()
    placements.splice(this.data.selectedIdx, 1)
    this.setData({ placements, selectedIdx: -1 })
    this.draw()
  },

  onScale(e) {
    const v = Number(e.detail.value)
    this.setData({ scaleVal: v })
    if (this.data.selectedIdx >= 0) {
      const key = 'placements[' + this.data.selectedIdx + '].scale'
      this.setData({ [key]: v })
      this.draw()
    }
  },

  // 触摸（canvas 事件里 touch.x/y 已是相对画布坐标）
  onTouchStart(e) {
    if (!this.board) return
    const t = e.touches[0]
    this.dragIdx = this.hitTest(t.x, t.y)
    if (this.dragIdx !== this.data.selectedIdx) {
      this.setData({
        selectedIdx: this.dragIdx,
        scaleVal: this.dragIdx >= 0 ? this.data.placements[this.dragIdx].scale : 1,
      })
      this.draw()
    }
    this.dragMoved = false
  },

  onTouchMove(e) {
    if (this.dragIdx < 0 || !this.board) return
    const t = e.touches[0]
    const x = Math.min(1, Math.max(0, t.x / this.cssW))
    const y = Math.min(1, Math.max(0, t.y / this.cssH))
    this.dragMoved = true
    this.setData({
      ['placements[' + this.dragIdx + '].x']: Math.round(x * 1000) / 1000,
      ['placements[' + this.dragIdx + '].y']: Math.round(y * 1000) / 1000,
    })
    this.draw()
  },

  onTouchEnd() {
    this.dragIdx = -1
  },

  hitTest(px, py) {
    const list = this.data.placements
    for (let i = list.length - 1; i >= 0; i--) {
      const r = this.placementRect(list[i])
      if (px >= r.x && px <= r.x + r.w && py >= r.y && py <= r.y + r.h) return i
    }
    return -1
  },

  // 某个放置的绘制矩形（css 像素），元素高度基准为画布高 30%
  placementRect(p) {
    const im = this.imgCache[p.url]
    const h = this.cssH * 0.3 * p.scale
    const w = im ? (h * im.width) / im.height : h * 0.7
    return { x: p.x * this.cssW - w / 2, y: p.y * this.cssH - h / 2, w, h }
  },

  // ---- 绘制 ----
  ensureImg(url, cb) {
    if (this.imgCache[url]) {
      if (cb) cb()
      return
    }
    const im = this.canvas.createImage()
    im.onload = () => {
      this.imgCache[url] = im
      if (cb) cb()
    }
    im.onerror = () => {}
    im.src = img(url)
  },

  draw() {
    if (!this.ctx) return
    const ctx = this.ctx
    ctx.clearRect(0, 0, this.cssW, this.cssH)
    if (!this.board) return
    const bim = this.imgCache[this.board.url]
    if (bim) {
      ctx.drawImage(bim, 0, 0, this.cssW, this.cssH)
    } else {
      this.ensureImg(this.board.url, () => this.draw())
    }
    this.data.placements.forEach((p) => {
      const im = this.imgCache[p.url]
      if (!im) {
        this.ensureImg(p.url, () => this.draw())
        return
      }
      const r = this.placementRect(p)
      ctx.drawImage(im, r.x, r.y, r.w, r.h)
      if (this.data.placements.indexOf(p) === this.data.selectedIdx) {
        ctx.strokeStyle = '#ff7ea0'
        ctx.setLineDash([6, 4])
        ctx.strokeRect(r.x, r.y, r.w, r.h)
        ctx.setLineDash([])
      }
    })
  },

  // ---- 生成作品 ----
  compose() {
    if (!this.board || !this.data.placements.length) {
      wx.showToast({ title: '请先选择元素拼装', icon: 'none' })
      return
    }
    this.setData({ composing: true })
    api
      .request('POST', '/collage/compose', {
        board_id: this.board.id,
        placements: this.data.placements.map((p) => ({
          element_id: p.elementId,
          x: p.x,
          y: p.y,
          scale: p.scale,
        })),
      })
      .then((work) => {
        wx.showModal({
          title: '作品已生成',
          content: '已保存到「作品」，可去商品页选用',
          confirmText: '看作品',
          cancelText: '继续拼',
          success: (r) => {
            if (r.confirm) wx.switchTab({ url: '/pages/works/works' })
            else this.draw()
          },
        })
      })
      .catch((e) => wx.showModal({ title: '生成失败', content: e.message, showCancel: false }))
      .then(() => this.setData({ composing: false }))
  },

  goAI() {
    wx.navigateTo({ url: '/pages/ai/ai' })
  },
})
