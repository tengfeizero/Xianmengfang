#!/usr/bin/env node
// 后台页面截图走查：登录 → 逐页截图（Edge 通道，和 mp-preview 同一套）
const path = require('path')
const fs = require('fs')
const { chromium } = require('./pw')

const BASE = 'http://127.0.0.1:8080/ui'
const OUT = path.resolve(__dirname, '../shots/web')
const PAGES = process.argv.slice(2).length
  ? process.argv.slice(2)
  : ['point-logs', 'work-likes', 'work-usages', 'users', 'works']

;(async () => {
  fs.mkdirSync(OUT, { recursive: true })
  const browser = await chromium.launch({ channel: 'msedge', headless: true })
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 })
  const page = await ctx.newPage()

  // 登录
  await page.goto(`${BASE}/#/login`, { waitUntil: 'networkidle' })
  await page.fill('input[placeholder="用户名"]', 'admin')
  await page.fill('input[placeholder="密码"]', 'admin123')
  await page.click('button.submit')
  await page.waitForTimeout(1500)
  console.log('after login:', page.url())

  for (const p of PAGES) {
    await page.goto(`${BASE}/#/${p}`, { waitUntil: 'networkidle' })
    await page.waitForTimeout(900)
    const file = path.join(OUT, `${p}.png`)
    await page.screenshot({ path: file, fullPage: true })
    console.log('shot', file)
  }

  await browser.close()
})().catch((e) => {
  console.error(e.message)
  process.exit(1)
})
