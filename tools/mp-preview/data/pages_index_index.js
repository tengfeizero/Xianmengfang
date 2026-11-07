const { art } = require('../mock')

module.exports = {
  cartCount: 2,
  quick: [
    { icon: 'wand', label: 'AI 生图', color: '#6c5ce7', bg: '#f1eeff' },
    { icon: 'palette', label: '拼装工坊', color: '#ff6fa5', bg: '#ffeef4' },
    { icon: 'compass', label: '灵感广场', color: '#14b8a6', bg: '#e8f8f4' },
    { icon: 'grid', label: '我的作品', color: '#f5a623', bg: '#fff6e4' },
  ],
  plaza: [1, 2, 3, 4].map((i) => ({ id: i, cover: art(i, 400, 500), likeCount: 12 * i })),
  products: [
    { id: 1, title: '小梦女孩亚克力立牌 · 定制款', subtitle: '双面印刷 / 附底座', cover: art(11, 500, 500), priceText: '39.00' },
    { id: 2, title: '定制帆布包 大容量单肩', subtitle: '12oz 加厚 / 可水洗', cover: art(12, 500, 500), priceText: '59.00' },
    { id: 3, title: '照片转风格 明信片套组', subtitle: '一套 6 张 / 300g 卡纸', cover: art(13, 500, 500), priceText: '29.00' },
    { id: 4, title: '拼贴手机壳 磨砂手感', subtitle: '全包边 / 防摔', cover: art(14, 500, 500), priceText: '49.00' },
  ],
  loading: false,
}
