#!/usr/bin/env node
/**
 * 小程序页面离线预览器（仅用于设计走查，不参与小程序构建）
 *
 * 背景：本机没装微信开发者工具，但设计稿必须看到实物才能判断对齐/配色/动效，
 * 所以这里做一个「够用」的 WXML → HTML 渲染器：支持 {{}} / wx:if / wx:for / block /
 * 常用内置组件，以及本项目的 x-icon、x-empty、x-skeleton、status-tag、custom-tab-bar。
 * 图标直接 require 真实 utils/icons.js，样式直接读真实 wxss（rpx 按 375 宽换算），
 * 因此看到的效果和真机基本一致。
 *
 * 用法：node tools/mp-preview/render.js pages/index/index
 */
const fs = require('fs')
const path = require('path')

const ROOT = path.resolve(__dirname, '../../miniprogram')
const OUT = path.resolve(__dirname, 'out')
const { icon } = require(path.join(ROOT, 'utils/icons.js'))
const util = require(path.join(ROOT, 'utils/util.js'))

// 小程序里变量挂在 <page> 元素上；HTML 里没有 page 元素，把 page 选择器映射到 :root
const fixVars = (css) => css.replace(/\bpage\s*\{/g, ':root,page{')
const rpx2px = (css) =>
  fixVars(css).replace(/(-?\d*\.?\d+)rpx/g, (_, n) => (Number(n) / 2).toFixed(2) + 'px')

function readWxss(rel) {
  const p = path.join(ROOT, rel)
  return fs.existsSync(p) ? fs.readFileSync(p, 'utf8') : ''
}

/* ---------------- WXML 解析 ---------------- */
function parseWxml(src) {
  const root = { tag: '#root', attrs: {}, children: [] }
  const stack = [root]
  const re = /<!--[\s\S]*?-->|<\/([a-zA-Z0-9_-]+)\s*>|<([a-zA-Z0-9_-]+)((?:"[^"]*"|'[^']*'|[^>"'])*?)(\/?)>|([^<]+)/g
  let m
  while ((m = re.exec(src))) {
    if (m[0].startsWith('<!--')) continue
    if (m[1]) {
      if (stack.length > 1) stack.pop()
      continue
    }
    if (m[2]) {
      const attrs = parseAttrs(m[3] || '')
      const node = { tag: m[2], attrs, children: [] }
      stack[stack.length - 1].children.push(node)
      if (!m[4]) stack.push(node)
      continue
    }
    const text = m[5]
    if (text) stack[stack.length - 1].children.push({ tag: '#text', text })
  }
  return root
}

function parseAttrs(s) {
  const out = {}
  const re = /([a-zA-Z_:][\w:.-]*)(?:\s*=\s*("([^"]*)"|'([^']*)'))?/g
  let m
  while ((m = re.exec(s))) {
    out[m[1]] = m[3] !== undefined ? m[3] : m[4] !== undefined ? m[4] : 'true'
  }
  return out
}

/* ---------------- 表达式求值 ---------------- */
const cache = new Map()
function evalExpr(expr, scope) {
  try {
    let fn = cache.get(expr)
    if (!fn) {
      fn = new Function('$s', 'with ($s) { return (' + expr + ') }')
      cache.set(expr, fn)
    }
    return fn(scope)
  } catch (e) {
    return ''
  }
}
const MUSTACHE = /\{\{([\s\S]+?)\}\}/g
function interp(str, scope) {
  if (str === undefined || str === null) return ''
  const s = String(str)
  if (!s.includes('{{')) return s
  const whole = s.match(/^\{\{([\s\S]+)\}\}$/)
  if (whole) return evalExpr(whole[1], scope)
  return s.replace(MUSTACHE, (_, e) => {
    const v = evalExpr(e, scope)
    return v === undefined || v === null || v === false ? '' : String(v)
  })
}

/* ---------------- 组件 ---------------- */
function loadComponent(dir, name) {
  const base = path.join(ROOT, 'components', dir, name)
  if (!fs.existsSync(base + '.wxml')) return null
  return {
    wxml: fs.readFileSync(base + '.wxml', 'utf8'),
    wxss: fs.existsSync(base + '.wxss') ? fs.readFileSync(base + '.wxss', 'utf8') : '',
    json: fs.existsSync(base + '.json') ? JSON.parse(fs.readFileSync(base + '.json', 'utf8')) : {},
  }
}

const COMPONENTS = {
  'x-icon': loadComponent('x-icon', 'x-icon'),
  'x-empty': loadComponent('x-empty', 'x-empty'),
  'x-skeleton': loadComponent('x-skeleton', 'x-skeleton'),
  'status-tag': loadComponent('status-tag', 'status-tag'),
}

const COMPONENT_WXSS = Object.values(COMPONENTS)
  .map((c) => (c ? c.wxss : ''))
  .join('\n')

/* ---------------- 渲染 ---------------- */
const VOID_TAGS = ['image', 'img', 'input', 'slider', 'canvas', 'switch', 'textarea']

function esc(s) {
  return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function renderChildren(children, scope) {
  let html = ''
  let chainDone = false
  let chainActive = false
  for (const node of children) {
    if (node.tag === '#text') {
      const t = interp(node.text, scope)
      if (String(t).trim()) html += esc(t)
      else if (String(t).includes('\n')) html += ' '
      continue
    }
    const a = node.attrs || {}
    if ('wx:if' in a) {
      chainDone = true
      chainActive = !!interp(a['wx:if'], scope)
      if (!chainActive) continue
    } else if ('wx:elif' in a) {
      if (!chainDone || chainActive) continue
      chainActive = !!interp(a['wx:elif'], scope)
      if (!chainActive) continue
    } else if ('wx:else' in a) {
      if (!chainDone || chainActive) continue
      chainActive = true
    } else {
      chainDone = false
      chainActive = false
    }

    if ('wx:for' in a) {
      const listRaw = interp(a['wx:for'], scope) || []
      const list = Array.isArray(listRaw) ? listRaw : []
      const itemName = a['wx:for-item'] || 'item'
      const indexName = a['wx:for-index'] || 'index'
      const rest = Object.assign({}, a)
      delete rest['wx:for']
      delete rest['wx:for-item']
      delete rest['wx:for-index']
      delete rest['wx:key']
      list.forEach((it, i) => {
        const s = Object.create(scope)
        s[itemName] = it
        s[indexName] = i
        html += renderNode(node, s, rest)
      })
      continue
    }
    html += renderNode(node, scope, a)
  }
  return html
}

const TAG_MAP = {
  view: 'div',
  text: 'span',
  block: null,
  image: 'img',
  'scroll-view': 'div',
  swiper: 'div',
  'swiper-item': 'div',
  button: 'button',
  navigator: 'a',
  canvas: 'div',
  input: 'div',
  textarea: 'div',
  slider: 'div',
  switch: 'div',
  form: 'div',
  label: 'label',
  'movable-area': 'div',
  'movable-view': 'div',
}

function renderNode(node, scope, attrs) {
  const tag = node.tag
  // 自定义组件
  if (COMPONENTS[tag]) return renderComponent(tag, attrs, scope)

  if (tag === 'block') return renderChildren(node.children, scope)

  // 内置组件的特殊处理
  if (tag === 'image') {
    const src = interp(attrs.src || '', scope)
    const cls = interp(attrs.class || '', scope)
    const style = interp(attrs.style || '', scope)
    const mode = attrs.mode || 'scaleToFill'
    const fit = mode === 'aspectFill' ? 'cover' : mode === 'aspectFit' ? 'contain' : 'fill'
    return `<img class="${cls}" src="${src}" style="object-fit:${fit};${style}" />`
  }
  if (tag === 'input' || tag === 'textarea') {
    const ph = interp(attrs.placeholder || '', scope)
    const cls = interp(attrs.class || '', scope)
    const style = interp(attrs.style || '', scope)
    const val = interp(attrs.value || '', scope)
    return `<div class="${cls}" style="${style}">${esc(val || ph)}</div>`
  }
  if (tag === 'slider') {
    const cls = interp(attrs.class || '', scope)
    return `<div class="${cls}" style="height:2px;background:#e6e1f2;border-radius:2px;margin:20px 0"><div style="width:50%;height:2px;background:#6c5ce7"></div></div>`
  }
  if (tag === 'swiper') {
    const cls = interp(attrs.class || '', scope)
    const style = interp(attrs.style || '', scope)
    // 只渲染第一帧：处理 swiper-item 上的 wx:for（画廊图集）
    const first = node.children.filter((c) => c.tag === 'swiper-item')[0]
    let inner = ''
    if (first) {
      const a = first.attrs || {}
      if ('wx:for' in a) {
        const list = [].concat(interp(a['wx:for'], scope) || [])
        const itemName = a['wx:for-item'] || 'item'
        const s = Object.create(scope)
        s[itemName] = list[0]
        s[a['wx:for-index'] || 'index'] = 0
        inner = renderChildren(first.children, s)
      } else {
        inner = renderChildren(first.children, scope)
      }
    }
    return `<div class="${cls}" style="overflow:hidden;${style}">${inner}</div>`
  }
  if (tag === 'canvas') {
    const cls = interp(attrs.class || '', scope)
    return `<div class="${cls}" id="mp-canvas"></div>`
  }

  const htmlTag = TAG_MAP[tag] || 'div'
  const cls = interp(attrs.class || '', scope)
  let style = interp(attrs.style || '', scope)
  if (tag === 'scroll-view') {
    if ('scroll-x' in attrs) style += ';overflow-x:auto;overflow-y:hidden'
    if ('scroll-y' in attrs) style += ';overflow-y:auto'
  }
  const inner = renderChildren(node.children || [], scope)
  return `<${htmlTag}${cls ? ` class="${cls}"` : ''}${style ? ` style="${style}"` : ''}>${inner}</${htmlTag}>`
}

function renderComponent(name, attrs, scope) {
  const def = COMPONENTS[name]
  const props = {}
  Object.keys(attrs).forEach((k) => {
    if (k.startsWith('bind') || k.startsWith('catch')) return
    props[k] = interp(attrs[k], scope)
  })

  if (name === 'x-icon') {
    const size = Number(props.size || 40) / 2
    const src = icon(props.name || '', props.color || '#191624', {
      fill: props.fill === true || props.fill === 'true',
      sw: Number(props.sw || 1.8),
    })
    return `<img class="ic" src="${src}" style="width:${size}px;height:${size}px;display:block" />`
  }
  if (name === 'x-skeleton') {
    const n = Number(props.count || 3)
    const tree = parseWxml(def.wxml)
    return renderChildren(tree.children, { type: props.type || 'list', items: new Array(n).fill(0) })
  }
  if (name === 'status-tag') {
    const st = props.status || ''
    return renderChildren(parseWxml(def.wxml).children, {
      text: util.statusText(st),
      color: util.statusColor(st),
      bg: util.statusBg(st),
      dot: props.dot !== 'false',
    })
  }
  if (name === 'x-empty') {
    return renderChildren(parseWxml(def.wxml).children, {
      loading: props.loading === true || props.loading === 'true',
      text: props.text || '这里还什么都没有',
      desc: props.desc || '',
      icon: props.icon || 'sparkles',
      actionText: props.actionText || '',
    })
  }
  return ''
}

/* ---------------- 组装 HTML ---------------- */
function build(pagePath) {
  const wxml = fs.readFileSync(path.join(ROOT, pagePath + '.wxml'), 'utf8')
  const pageJsonPath = path.join(ROOT, pagePath + '.json')
  const pageJson = fs.existsSync(pageJsonPath) ? JSON.parse(fs.readFileSync(pageJsonPath, 'utf8')) : {}
  const pageWxss = readWxss(pagePath + '.wxss')
  const appWxss = readWxss('app.wxss')
  const tabWxss = readWxss('../miniprogram/custom-tab-bar/index.wxss')

  const dataBase = path.join(__dirname, 'data', pagePath.replace(/\//g, '_'))
  let data = {}
  if (fs.existsSync(dataBase + '.js')) data = require(dataBase + '.js')
  else if (fs.existsSync(dataBase + '.json')) data = JSON.parse(fs.readFileSync(dataBase + '.json', 'utf8'))

  const body = renderChildren(parseWxml(wxml).children, data)

  const isTab = ['pages/index/index', 'pages/plaza/plaza', 'pages/studio/studio', 'pages/works/works', 'pages/mine/mine'].includes(
    pagePath
  )
  const title = pageJson.navigationBarTitleText || '现梦坊'

  const tabbar = isTab ? renderTabBar(pagePath) : ''
  const css = rpx2px(
    [
      `
      html,body{margin:0;padding:0;background:#e9e7ef;}
      .phone{position:relative;width:375px;height:812px;overflow:hidden;background:#f7f5fb;font-family:-apple-system,"PingFang SC",sans-serif;}
      .nav{position:fixed;top:0;left:0;width:375px;height:88px;z-index:300;background:#fff;display:flex;align-items:flex-end;justify-content:center;padding-bottom:10px;box-sizing:border-box;}
      .nav .t{font-size:17px;font-weight:600;color:#191624;}
      .nav .status{position:absolute;top:6px;left:16px;font-size:11px;color:#191624;opacity:.8;}
      .screen{position:fixed;top:88px;left:0;width:375px;height:724px;overflow-y:auto;overflow-x:hidden;-webkit-overflow-scrolling:touch;}
      button{border:none;background:transparent;font-family:inherit;padding:0;}
      /* 静态截图在 t=0 拍摄，入场动画会把元素停在 opacity:0，预览时直接关掉 */
      *{animation:none !important;}
      .mp-empty{color:#9a93a8;}
      `,
      appWxss,
      COMPONENT_WXSS,
      isTab ? tabWxss : '',
      pageWxss,
    ].join('\n')
  )

  const html = `<!DOCTYPE html>
<html><head><meta charset="utf-8" /><style>${css}</style></head>
<body><div class="phone">
  <div class="nav"><div class="status">9:41</div><div class="t">${title}</div></div>
  <div class="screen">${body}</div>
  ${tabbar}
</div></body></html>`

  fs.mkdirSync(OUT, { recursive: true })
  const outFile = path.join(OUT, pagePath.replace(/\//g, '_') + '.html')
  fs.writeFileSync(outFile, html)
  return outFile
}

function renderTabBar(pagePath) {
  const tabs = [
    { path: '/pages/index/index', text: '首页', icon: 'home' },
    { path: '/pages/plaza/plaza', text: '广场', icon: 'compass' },
    { path: '/pages/studio/studio', text: '创作', icon: 'plus', center: true },
    { path: '/pages/works/works', text: '作品', icon: 'grid' },
    { path: '/pages/mine/mine', text: '我的', icon: 'user' },
  ]
  const active = tabs.findIndex((t) => t.path === '/' + pagePath)
  let html = '<div class="tabbar"><div class="bar">'
  tabs.forEach((t, i) => {
    const on = i === active
    if (t.center) {
      html += `<div class="item mid ${on ? 'on' : ''}"><div class="mid-btn ${on ? 'mid-on' : ''}"><img class="mid-ic" src="${icon(
        'plus',
        '#ffffff',
        { sw: 2.6 }
      )}" /></div><span class="label mid-label">${t.text}</span></div>`
    } else {
      html += `<div class="item ${on ? 'on' : ''}"><img class="ic" src="${icon(t.icon, on ? '#6c5ce7' : '#9a93a8', {
        sw: on ? 2.2 : 1.8,
      })}" /><span class="label">${t.text}</span></div>`
    }
  })
  html += '</div></div>'
  return html
}

const target = process.argv[2]
if (!target) {
  console.error('usage: node render.js pages/index/index')
  process.exit(1)
}
console.log(build(target))
