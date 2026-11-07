# 现梦坊 — 项目长期记忆

## 本地运行（重要）
- 服务端：`cd server && XMF_WEB=../web/dist ./data/xmf.exe`（监听 :8080，顺带托管 Web 后台到 `/ui/`）
  - 二进制在 `server/data/xmf.exe`，CWD 必须是 `server/`（默认 `XMF_DB=data/xmf.db` 是相对路径）
  - **改了 Go 代码必须重新编译并重启**：Windows 下运行中的 exe 被锁，需先停进程再 `go build -o data/xmf.exe ./cmd/xmf`
  - 接口返回 `404 page not found`（gin 纯文本）= 路由不存在，通常就是**旧二进制**；接口返回 `{"code":404,...}` = 路由存在、业务判定未命中
- 环境变量：`XMF_ADDR`(默认 :8080)、`XMF_DB`、`XMF_DEV_LOGIN`(默认 1)、`XMF_JWT_SECRET`(缺省读 data/.jwt_secret)、`XMF_WX_APPID/SECRET`、`XMF_AI_*`
- AI 密钥**存数据库** `configs.ai_api_key`（后台可改，请求时解析），不依赖环境变量 → 重启丢环境变量不影响 AI
- 开发者登录：`POST /api/v1/user/auth/dev_login {"openid":"xxx"}`（nickname 可选），返回 `data.token`；后台 `admin/admin123`

## 小程序设计/预览
- 无微信开发者工具时用 `tools/mp-preview/`：`node render-all.js` 出 HTML（`out/`，已 gitignore）→ `node serve.js`(:8123) → `node shot-all.js [页名过滤...]` 截图到 `shots/mp/`
- 截图一律走 playwright-core + Edge 通道；headless Edge `--screenshot` 会假成功/写空白图，不要用
- 图标统一走 `miniprogram/utils/icons.js`（运行时 SVG，`x-icon` 组件），新增图标只加 `PATHS` 条目

## 换机器 / 新环境
- 代码与本目录的记忆都在仓库里（`.workbuddy/memory/` 已入库），clone 后 AI 能直接接着干
- **不入库、要手动带**：`server/data/`（`xmf.db`、`uploads/`、`.jwt_secret`）、`web/dist`、`shots/`、`generated-images/`
- 新机器 `go run ./cmd/xmf` 自动建库 + 种子（4 商品 / 4 底板 / 24 部件 / admin-admin123）；AI key 存在库里，需在后台重填，否则回落 Mock

## 约定
- 用户端接口前缀 `/api/v1/user`，后台 `/api/v1/admin`；gin 同段通配冲突：`/works/:id` 存在时不能再加 `/works/search`（搜索复用 `/plaza?keyword=`）
- 文档以 `docs/` 为准（PRD / API / CHANGELOG / TODO 均随功能同步更新）；人看的上手入口是根目录 `README.md`，AI 入口是 `AGENTS.md`
- 仓库要能 clone 就跑：脚本依赖统一走 `tools/pw.js`（解析 playwright-core），**禁止写死本机绝对路径**；密钥只走环境变量或 `server/data/`，变量清单见 `server/.env.example`
- 品牌色全站统一 `#f2455f`（后台与小程序同源），渐变 `#ff6b7d→#f2455f→#d92c47`
