#!/usr/bin/env node
// 渲染全部页面并输出每页的 HTML 路径（配合 shot.sh 截图）
const { execSync } = require('child_process')
const path = require('path')

const PAGES = [
  'pages/index/index',
  'pages/plaza/plaza',
  'pages/studio/studio',
  'pages/works/works',
  'pages/mine/mine',
  'pages/product/product',
  'pages/cart/cart',
  'pages/search/search',
  'pages/ai/ai',
  'pages/order-list/order-list',
  'pages/order-detail/order-detail',
  'pages/points/points',
  'pages/work-detail/work-detail',
  'pages/address/list',
  'pages/address/edit',
  'pages/settings/settings',
  'pages/help/help',
  'pages/about/about',
  'pages/login/login',
]

const here = __dirname
for (const p of PAGES) {
  try {
    const out = execSync(`node "${path.join(here, 'render.js')}" ${p}`, { encoding: 'utf8' })
    console.log(path.basename(out.trim()))
  } catch (e) {
    console.error('FAIL', p, e.message)
  }
}
