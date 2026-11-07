// 预览静态服务器：serve out/ 目录，供 playwright 走 http 访问（file: 被禁）
const http = require('http')
const fs = require('fs')
const path = require('path')

const DIR = path.join(__dirname, 'out')
const PORT = 8123

http
  .createServer((req, res) => {
    const name = decodeURIComponent(req.url.slice(1)) || 'index.html'
    const file = path.join(DIR, path.normalize(name).replace(/^(\.\.[/\\])+/, ''))
    fs.readFile(file, (err, buf) => {
      if (err) {
        res.writeHead(404)
        res.end('not found')
        return
      }
      res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' })
      res.end(buf)
    })
  })
  .listen(PORT, () => console.log('serving ' + DIR + ' at http://127.0.0.1:' + PORT))
