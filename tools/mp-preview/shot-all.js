#!/usr/bin/env node
// 用 playwright-core(Edge 通道) 对所有预览页截图，比裸 Edge --screenshot 稳定
const path = require('path')
const fs = require('fs')
const { chromium } = require('../pw')

const OUT = path.join(__dirname, 'out')
const SHOTS = path.resolve(__dirname, '../../shots/mp')
const BASE = 'http://127.0.0.1:8123'

;(async () => {
  fs.mkdirSync(SHOTS, { recursive: true })
  const browser = await chromium.launch({ channel: 'msedge', headless: true })
  const page = await browser.newPage({ viewport: { width: 375, height: 812 }, deviceScaleFactor: 2 })
  const files = fs.readdirSync(OUT).filter((f) => f.startsWith('pages_') && f.endsWith('.html'))
  // 支持只截指定页：node shot-all.js pages_help_help pages_settings_settings
  const only = process.argv.slice(2)
  const picked = only.length ? files.filter((f) => only.some((k) => f.includes(k))) : files
  for (const f of picked) {
    await page.goto(BASE + '/' + f, { waitUntil: 'networkidle' })
    await page.waitForTimeout(350)
    await page.screenshot({ path: path.join(SHOTS, f.replace('.html', '.png')) })
    // 长页面：展开 .phone/.screen 后截整页，便于检查底部内容
    await page.addStyleTag({
      content:
        '.phone{height:auto !important;overflow:visible !important}.screen{position:static !important;height:auto !important;overflow:visible !important;padding-top:88px !important}',
    })
    await page.waitForTimeout(120)
    await page.screenshot({ path: path.join(SHOTS, f.replace('.html', '.full.png')), fullPage: true })
    console.log('shot', f)
  }
  await browser.close()
})().catch((e) => {
  console.error(e)
  process.exit(1)
})
