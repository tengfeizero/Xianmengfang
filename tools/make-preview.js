#!/usr/bin/env node
// 生成 README 预览图：把多张手机截图拼成一张横排展示图（浏览器渲染，无需图像库），
// 后台大图则缩放到固定宽度。输出到 docs/images/（随仓库提交，README 引用）。
//
// 用法：node tools/make-preview.js
// 前置：shots/mp/*.png（tools/mp-preview/shot-all.js 产出）、shots/web/*.png（tools/web-shot.js 产出）
const path = require('path')
const fs = require('fs')
const { chromium } = require('./pw')

const ROOT = path.resolve(__dirname, '..')
const MP = path.join(ROOT, 'shots', 'mp')
const WEB = path.join(ROOT, 'shots', 'web')
const OUT = path.join(ROOT, 'docs', 'images')

// 小程序拼图分组：[源文件, 标题]
const PHONE_GROUPS = [
  {
    file: 'mp-browse.jpg',
    alt: '小程序 · 浏览与创作',
    phones: [
      ['pages_index_index.png', '首页'],
      ['pages_ai_ai.png', 'AI 生成'],
      ['pages_plaza_plaza.png', '灵感广场'],
    ],
  },
  {
    file: 'mp-trade.jpg',
    alt: '小程序 · 商品与订单',
    phones: [
      ['pages_product_product.png', '商品详情'],
      ['pages_order-detail_order-detail.png', '订单详情'],
      ['pages_mine_mine.png', '我的'],
    ],
  },
]

// 后台单图：[源文件, 输出文件, 目标宽度]
const WEB_SHOTS = [
  ['dashboard.png', 'web-dashboard.png', 1100],
  ['work-usages.png', 'web-usages.png', 1100],
]

const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;')

;(async () => {
  for (const g of PHONE_GROUPS) {
    for (const [f] of g.phones) {
      if (!fs.existsSync(path.join(MP, f))) throw new Error(`缺少截图 ${f}，先跑 tools/mp-preview/shot-all.js`)
    }
  }
  for (const [f] of WEB_SHOTS) {
    if (!fs.existsSync(path.join(WEB, f))) throw new Error(`缺少截图 ${f}，先跑 tools/web-shot.js`)
  }
  fs.mkdirSync(OUT, { recursive: true })

  const browser = await chromium.launch({ channel: 'msedge', headless: true })

  // —— 小程序拼图 ——
  for (const g of PHONE_GROUPS) {
    const page = await browser.newPage({ viewport: { width: 1400, height: 1000 }, deviceScaleFactor: 2 })
    const imgs = g.phones
      .map(
        ([f, label]) => `
      <figure>
        <div class="phone"><img src="file:///${path.join(MP, f).replace(/\\/g, '/')}"></div>
        <figcaption>${esc(label)}</figcaption>
      </figure>`,
      )
      .join('')
    const html = `<!doctype html><meta charset="utf8"><style>
      * { margin: 0; box-sizing: border-box; }
      body { font-family: "Microsoft YaHei", "PingFang SC", sans-serif; }
      #board {
        display: inline-flex; gap: 22px; padding: 26px 30px;
        background: linear-gradient(180deg, #faf9fb, #f0eff4);
      }
      figure { display: flex; flex-direction: column; align-items: center; }
      .phone {
        width: 320px; border-radius: 22px; overflow: hidden;
        box-shadow: 0 10px 30px rgba(31, 27, 45, .16), 0 2px 6px rgba(31, 27, 45, .08);
        background: #fff;
      }
      .phone img { display: block; width: 100%; }
      figcaption { margin-top: 12px; font-size: 14px; color: #7a7980; }
    </style><div id="board">${imgs}</div>`
    const tmp = path.join(OUT, '_tmp.html')
    fs.writeFileSync(tmp, html)
    await page.goto('file:///' + tmp.replace(/\\/g, '/'))
    await page.waitForTimeout(400)
    const board = await page.$('#board')
    // JPEG：大面积纯色 UI 压缩率远高于 PNG，GitHub 缩放显示后无可见差别
    await board.screenshot({ path: path.join(OUT, g.file), type: 'jpeg', quality: 90 })
    fs.unlinkSync(tmp)
    await page.close()
    console.log('made', g.file)
  }

  // —— 后台大图缩放 ——
  for (const [src, dest, width] of WEB_SHOTS) {
    const page = await browser.newPage({ viewport: { width: width + 40, height: 900 }, deviceScaleFactor: 1 })
    const html = `<!doctype html><meta charset="utf8"><style>*{margin:0}img{display:block}</style>
      <img id="shot" width="${width}" src="file:///${path.join(WEB, src).replace(/\\/g, '/')}">`
    const tmp = path.join(OUT, '_tmp.html')
    fs.writeFileSync(tmp, html)
    await page.goto('file:///' + tmp.replace(/\\/g, '/'))
    await page.waitForTimeout(300)
    const el = await page.$('#shot')
    await el.screenshot({ path: path.join(OUT, dest) })
    fs.unlinkSync(tmp)
    await page.close()
    console.log('made', dest)
  }

  await browser.close()
  for (const f of fs.readdirSync(OUT)) {
    const kb = Math.round(fs.statSync(path.join(OUT, f)).size / 1024)
    console.log(`  ${f}  ${kb} KB`)
  }
})().catch((e) => {
  console.error(e)
  process.exit(1)
})
