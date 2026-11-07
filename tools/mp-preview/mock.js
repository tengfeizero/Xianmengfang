// 预览用占位图：生成柔和的渐变 SVG，避免设计走查时全是灰块
const PALS = [
  ['#cabfff', '#ffd3e4'],
  ['#bfe7e1', '#d6ecff'],
  ['#ffd7c2', '#ffc7d9'],
  ['#d6cfff', '#c6f0e7'],
  ['#ffe4a8', '#ffcdd7'],
  ['#c9d8ff', '#e6d5ff'],
]

function art(seed, w, h) {
  const p = PALS[Math.abs(seed) % PALS.length]
  const s = Math.abs(seed)
  const svg =
    `<svg xmlns="http://www.w3.org/2000/svg" width="${w || 400}" height="${h || 400}" viewBox="0 0 ${w || 400} ${h || 400}">` +
    `<defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1">` +
    `<stop offset="0" stop-color="${p[0]}"/><stop offset="1" stop-color="${p[1]}"/>` +
    `</linearGradient></defs>` +
    `<rect width="100%" height="100%" fill="url(#g)"/>` +
    `<circle cx="${(w || 400) * 0.72}" cy="${(h || 400) * 0.26}" r="${(w || 400) * 0.18}" fill="#ffffff" opacity="0.45"/>` +
    `<circle cx="${(w || 400) * 0.28}" cy="${(h || 400) * 0.72}" r="${(w || 400) * 0.26}" fill="#ffffff" opacity="0.3"/>` +
    `<rect x="${(w || 400) * 0.18}" y="${(h || 400) * 0.34}" width="${(w || 400) * 0.64}" height="${(h || 400) * 0.34}" rx="${(w || 400) * 0.06}" fill="#ffffff" opacity="0.32"/>` +
    (s % 3 === 0
      ? `<text x="50%" y="56%" text-anchor="middle" font-size="${(w || 400) * 0.13}" font-family="PingFang SC" fill="#ffffff" opacity="0.9">作品 ${s}</text>`
      : '') +
    `</svg>`
  return 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(svg)
}

module.exports = { art }
