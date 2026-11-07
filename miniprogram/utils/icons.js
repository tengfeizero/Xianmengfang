// 图标系统：线性 SVG（24×24）运行时着色，免去为每个尺寸/颜色准备图片。
// 用法：icon('home', '#6c5ce7', { size: 40 }) → data URI，直接喂给 <image>。
// 之所以走 data URI 而不是字体图标：小程序无需引入 ttf，也不受字体文件体积限制。

const B64 = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/'

// 极简 base64（图标 SVG 全是 ASCII，不需要处理多字节）
function b64(str) {
  let out = ''
  for (let i = 0; i < str.length; i += 3) {
    const c1 = str.charCodeAt(i)
    const c2 = i + 1 < str.length ? str.charCodeAt(i + 1) : NaN
    const c3 = i + 2 < str.length ? str.charCodeAt(i + 2) : NaN
    out += B64[c1 >> 2]
    out += B64[((c1 & 3) << 4) | (isNaN(c2) ? 0 : c2 >> 4)]
    out += isNaN(c2) ? '=' : B64[((c2 & 15) << 2) | (isNaN(c3) ? 0 : c3 >> 6)]
    out += isNaN(c3) ? '=' : B64[c3 & 63]
  }
  return out
}

// 线性图标（stroke 描边）；fill 图标另在 PATHS 里用闭合路径，靠 fill 参数填色
const PATHS = {
  home: '<path d="M3 10.6 12 3l9 7.6"/><path d="M5.2 9.8V19a2 2 0 0 0 2 2h3.3v-5.6h3v5.6h3.3a2 2 0 0 0 2-2V9.8"/>',
  compass:
    '<circle cx="12" cy="12" r="9"/><path d="m15.6 8.4-2.1 5.1-5.1 2.1 2.1-5.1z"/>',
  plus: '<path d="M12 5.5v13M5.5 12h13"/>',
  grid: '<rect x="3.5" y="3.5" width="7" height="7" rx="2.2"/><rect x="13.5" y="3.5" width="7" height="7" rx="2.2"/><rect x="3.5" y="13.5" width="7" height="7" rx="2.2"/><rect x="13.5" y="13.5" width="7" height="7" rx="2.2"/>',
  user: '<circle cx="12" cy="8" r="4"/><path d="M4.2 20.6c.6-4.1 3.8-6.4 7.8-6.4s7.2 2.3 7.8 6.4"/>',
  search: '<circle cx="11" cy="11" r="7"/><path d="m20.2 20.2-3.6-3.6"/>',
  cart: '<path d="M3 4.2h2.3l2.3 11a1.8 1.8 0 0 0 1.8 1.5h8.4a1.8 1.8 0 0 0 1.8-1.4L21.4 8H6.1"/><circle cx="10" cy="20" r="1.4"/><circle cx="17.8" cy="20" r="1.4"/>',
  heart:
    '<path d="M12 20.2S4.5 15.7 4.5 10.9A4.2 4.2 0 0 1 12 8.3a4.2 4.2 0 0 1 7.5 2.6c0 4.8-7.5 9.3-7.5 9.3z"/>',
  star: '<path d="m12 3.6 2.7 5.5 6 .9-4.3 4.2 1 6-5.4-2.9-5.4 2.9 1-6L3.3 10l6-.9z"/>',
  bolt: '<path d="M13.2 3 5.4 13.6h5.4L10.8 21l7.8-10.6h-5.4z"/>',
  back: '<path d="M15 5.5 7.6 12.5 15 19.5"/>',
  right: '<path d="m9.5 5.5 7.4 7-7.4 7"/>',
  down: '<path d="m5.5 9.5 7 7 7-7"/>',
  close: '<path d="M6.2 6.2l11.6 11.6M17.8 6.2 6.2 17.8"/>',
  check: '<path d="m4.5 12.5 5 5L19.5 6.8"/>',
  trash:
    '<path d="M4 7h16M9.2 7V4.8h5.6V7M6.6 7l.8 12.2A1.9 1.9 0 0 0 9.3 21h5.4a1.9 1.9 0 0 0 1.9-1.8L17.4 7"/>',
  edit: '<path d="M4 20h4.2L19 9.2a2.9 2.9 0 0 0-4.2-4.2L4 15.8z"/><path d="m13.6 6.6 3.8 3.8"/>',
  share:
    '<circle cx="18" cy="5.2" r="2.6"/><circle cx="6" cy="12" r="2.6"/><circle cx="18" cy="18.8" r="2.6"/><path d="m8.4 10.7 7.2-4M8.4 13.3l7.2 4"/>',
  download: '<path d="M12 4v10.6m0 0 4.2-4.2M12 14.6 7.8 10.4"/><path d="M4.5 19.5h15"/>',
  refresh:
    '<path d="M4.2 12a7.8 7.8 0 0 1 13.4-5.5l2.4 2.3"/><path d="M20 4.4v4.4h-4.4"/><path d="M19.8 12a7.8 7.8 0 0 1-13.4 5.5L4 15.2"/><path d="M4 19.6v-4.4h4.4"/>',
  undo: '<path d="M4.5 9.5h9.2a4.8 4.8 0 1 1 0 9.6H8.4"/><path d="M8.4 5.2 4.2 9.5l4.2 4.3"/>',
  sparkles:
    '<path d="M11.4 3.6 13 8l4.4 1.6L13 11.2 11.4 15.6 9.8 11.2 5.4 9.6 9.8 8z"/><path d="M18 14.8l.9 2.4 2.4.9-2.4.9-.9 2.4-.9-2.4-2.4-.9 2.4-.9z"/>',
  wand: '<path d="m4 20 9.4-9.4"/><path d="m14.6 5.4 1.4 1.4M18.4 3.4l.8 2.6 2.6.8-2.6.8-.8 2.6-.8-2.6-2.6-.8 2.6-.8zM5.6 4.4l.6 1.8 1.8.6-1.8.6-.6 1.8-.6-1.8L3.2 6.8l1.8-.6z"/>',
  palette:
    '<path d="M12 3.2a8.8 8.8 0 1 0 0 17.6c1.4 0 2.1-.9 1.7-2.1-.5-1.5.4-2.7 2-2.7h1.6a3.2 3.2 0 0 0 3.2-3.2A8.8 8.8 0 0 0 12 3.2z"/><circle cx="8.4" cy="9.2" r="1.1"/><circle cx="12" cy="7" r="1.1"/><circle cx="15.6" cy="9.2" r="1.1"/>',
  image:
    '<rect x="3.2" y="4.2" width="17.6" height="15.6" rx="3.4"/><circle cx="8.8" cy="9.6" r="1.5"/><path d="m4.2 17.8 4.6-4.2 3.6 3.2 3-2.6 3.8 3.6"/>',
  camera:
    '<path d="M4 8.4h3l1.5-2.2h7L17 8.4h3a1.2 1.2 0 0 1 1.2 1.2v8.4a1.2 1.2 0 0 1-1.2 1.2H4a1.2 1.2 0 0 1-1.2-1.2V9.6A1.2 1.2 0 0 1 4 8.4z"/><circle cx="12" cy="13.6" r="3.4"/>',
  package:
    '<path d="M12 3 3.6 7.5v9L12 21l8.4-4.5v-9z"/><path d="m3.6 7.5 8.4 4.6 8.4-4.6M12 12.1V21"/>',
  truck:
    '<path d="M2.4 7h10.8v9.4H2.4z"/><path d="M13.2 10.2h3.9l3.3 3.4v2.8h-7.2"/><circle cx="6.4" cy="18.4" r="1.6"/><circle cx="17.2" cy="18.4" r="1.6"/>',
  pin: '<path d="M12 21s7-5.6 7-11.2a7 7 0 1 0-14 0C5 15.4 12 21 12 21z"/><circle cx="12" cy="9.8" r="2.6"/>',
  ticket:
    '<path d="M4 8.2a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v1.6a2.2 2.2 0 0 0 0 4.4v1.6a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2v-1.6a2.2 2.2 0 0 0 0-4.4z"/>',
  clock: '<circle cx="12" cy="12" r="8.8"/><path d="M12 7.4V12l3.4 2"/>',
  settings:
    '<circle cx="12" cy="12" r="3.1"/><path d="M12 2.8v2.6M12 18.6v2.6M4.2 7.6l2.3 1.3M17.5 15.1l2.3 1.3M4.2 16.4l2.3-1.3M17.5 8.9l2.3-1.3"/>',
  help: '<circle cx="12" cy="12" r="8.8"/><path d="M9.4 9.4a2.7 2.7 0 1 1 4.4 2.2c-.9.7-1.8 1.3-1.8 2.5"/><path d="M12 17.2h.01"/>',
  info: '<circle cx="12" cy="12" r="8.8"/><path d="M12 11v5.6"/><path d="M12 7.8h.01"/>',
  logout:
    '<path d="M14.6 4.8h3.2a2 2 0 0 1 2 2v10.4a2 2 0 0 1-2 2h-3.2"/><path d="m9.6 8-4 4 4 4M5.6 12h9.6"/>',
  eye: '<path d="M2.6 12S6.2 6.4 12 6.4 21.4 12 21.4 12 17.8 17.6 12 17.6 2.6 12 2.6 12z"/><circle cx="12" cy="12" r="2.8"/>',
  gift: '<path d="M4.4 10.8h15.2V20H4.4z"/><path d="M3.2 7.4h17.6v3.4H3.2zM12 7.4V20"/><path d="M12 7.4S9.6 3 7.6 4.4s1.8 3 4.4 3 4.4-3 2.4-3S12 7.4 12 7.4z"/>',
  medal: '<circle cx="12" cy="9.2" r="5.4"/><path d="m8.6 13.6-1.8 7.2L12 18.4l5.2 2.4-1.8-7.2"/>',
  layers:
    '<path d="m12 3.2 8.4 4.6L12 12.4 3.6 7.8z"/><path d="m3.6 12.2 8.4 4.6 8.4-4.6M3.6 16.4l8.4 4.6 8.4-4.6"/>',
  eraser: '<path d="m4.4 15.4 7-7 8 8-4.2 4.2H8.6z"/><path d="M9.6 20.4h10.8"/>',
  bell: '<path d="M18 9.4a6 6 0 1 0-12 0c0 5-2 6.2-2 6.2h16s-2-1.2-2-6.2z"/><path d="M10.4 19a1.7 1.7 0 0 0 3.2 0"/>',
  lock: '<rect x="4.8" y="10.6" width="14.4" height="9.6" rx="2.4"/><path d="M8.4 10.6V8a3.6 3.6 0 0 1 7.2 0v2.6"/>',
  filter: '<path d="M4.4 7.4h15.2M4.4 16.6h15.2"/><circle cx="15" cy="7.4" r="2.1" fill="#fff"/><circle cx="9" cy="16.6" r="2.1" fill="#fff"/>',
  copy: '<rect x="8.6" y="8.6" width="11.4" height="11.4" rx="2.6"/><path d="M15.4 6.4a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7.4a2 2 0 0 0 2 2"/>',
  play: '<circle cx="12" cy="12" r="8.8"/><path d="m10.2 8.8 5 3.2-5 3.2z"/>',
  fire: '<path d="M12 3s5.2 4.2 5.2 9a5.2 5.2 0 0 1-10.4 0C6.8 9 9 7.4 9 7.4s.4 2.4 1.6 2.4S12 6.6 12 3z"/>',
  spark: '<path d="M12 4.2 13.4 9l4.8 1.4-4.8 1.4L12 16.6l-1.4-4.8L5.8 10.4 10.6 9z"/>',
  shirt:
    '<path d="M8.4 4 4.2 6.4 5.8 11l2.6-1.1V20h7.2v-10.1L18.2 11l1.6-4.6L15.6 4l-2.2 1.6h-3z"/>',
  message:
    '<path d="M20.4 11.6c0 4-3.8 7.2-8.4 7.2a9.9 9.9 0 0 1-2.6-.35L4.6 20.2l1.2-3.5A6.9 6.9 0 0 1 3.6 11.6c0-4 3.8-7.2 8.4-7.2s8.4 3.2 8.4 7.2z"/>',
}

function icon(name, color, opts) {
  const body = PATHS[name]
  if (!body) return ''
  const o = opts || {}
  const fill = o.fill ? color : 'none'
  const sw = o.sw || 1.8
  const svg =
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="' +
    fill +
    '" stroke="' +
    (o.fill ? 'none' : color) +
    '" stroke-width="' +
    sw +
    '" stroke-linecap="round" stroke-linejoin="round">' +
    body +
    '</svg>'
  return 'data:image/svg+xml;base64,' + b64(svg)
}

module.exports = { icon, PATHS }
