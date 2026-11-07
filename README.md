# 现梦坊 · Xianmengfang

**让创意（梦）走进现实（现）** —— 用户在小程序里创作图案，印到衣服 / 周边 / 玩偶上下单；工厂在后台接单、生产、发货。

一套系统 = 微信小程序 + Vue 管理后台 + Go 单二进制服务。

![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white) ![Vue](https://img.shields.io/badge/Vue-3.5-42B883?logo=vue.js&logoColor=white) ![Element Plus](https://img.shields.io/badge/Element%20Plus-2.13-409EFF) ![小程序](https://img.shields.io/badge/微信小程序-原生-07C160?logo=wechat&logoColor=white) ![License](https://img.shields.io/badge/License-MIT-black)

---

## 这是什么

一个**单厂部署**的定制文创交易系统（一套系统服务一个工厂，架构预留多租户）。它把「用户创作 → 下单 → 工厂生产 → 发货 → 确认收货」这条链完整跑通，不是 demo 级的增删改查：

- 用户端是**微信小程序**：三种创作方式、作品社区、购物车与订单、积分体系
- 管理端是**Web 后台**：商品/素材/订单/用户/权限/审计，加三个经营明细页（积分、点赞、使用）
- 后端是**一个 Go 二进制**：API + 托管后台静态资源，SQLite 起步（`data/xmf.db` 单文件，删了即重置）

## 能做什么

**创作（三条路）**
| 方式 | 说明 |
|---|---|
| 元素拼装 | 4 块底板 × 24 个 AI 生成部件（头发/帽子/睫毛/表情/服装/配饰），固定图层顺序，可微调位置缩放后合成 |
| 文字生成图 | 输入描述调 AI 生图（默认腾讯混元 `hy-image-v3.5`，可换供应商） |
| 照片转风格 | 上传照片转插画风格 |

拼装素材是**真图不是占位块**：以小姑娘立绘为风格基准做 img2img 换角色，部件用「给底板戴上道具 → 与底板像素差分」抠出，锚点由差分包围盒自动算出，所以换底板不会错位（细节见 [PRD 3.1.1](docs/PRD.md)）。

**交易**
- 商品 SPU/SKU、多图画廊、划线价、服务承诺、定制图案绑定
- 购物车（勾选部分结算）、收货地址（默认唯一、上限 20 条、下单写**快照**）
- 订单六态状态机：待接单 → 生产中 → 打包中 → 待发货 → 已发货 → 已完成（+ 已取消），只允许合法流转
- 支付为**模拟支付**（无微信支付资质时可用，替换点已隔离在 service 层）

**社区与积分**
- 作品发布到广场（最新 / 最常用 / 最多赞三榜）、点赞、提示词搜索
- 「点赞即可用」：下单可选自己的作品或已点赞的公开作品
- 积分：作品被他人使用按件加积分；AI 免费额度用尽后 1 积分兑 1 张
- 注销账号：事务内清作品/购物车/地址，`open_id` 换 tombstone 释放唯一键，同一微信再登录是全新账号（订单与积分流水保留）

**管理后台**
- 账号—角色—权限三层 RBAC（内置超管 / 运营 / 只读），**每请求现查权限**，禁用或改权限立即生效
- 写操作全量审计（含越权拦截），密钥字段脱敏
- 经营明细：积分流水 / 点赞明细 / 使用明细，可从用户行、作品行下钻
- 工作台：概览卡（待处理、今日成交、在售商品、作品、用户）+ 快捷入口 + 最近订单

## 三端一览

| 端 | 技术 | 入口 | 说明 |
|---|---|---|---|
| 服务端 | Go 1.27 · Gin · GORM · SQLite | `:8080/api/v1/**` | 分层：`api` 薄层 → `service` 业务 → `repo` CRUD |
| 管理后台 | Vue 3.5 · TS · Element Plus · Vite | `:8080/ui/` | 构建产物由 Go 直接托管，无需单独部署 |
| 小程序 | 微信原生（无 npm 依赖） | `miniprogram/` | 19 页 / 5 tab / 自定义 tabBar |

```
┌─────────────────┐   ┌──────────────────┐
│ 微信小程序        │   │ 管理后台（Vue）   │
│ 5 tab · 19 页     │   │ Element Plus     │
└────────┬────────┘   └────────┬─────────┘
         │ /api/v1/user         │ /api/v1/admin
         └──────────┬───────────┘
                    ▼
        ┌───────────────────────────┐
        │ Go 单二进制（Gin）          │
        │ api → service → repo      │
        ├───────────────────────────┤
        │ SQLite · go:embed 素材      │
        └───────────────────────────┘
```

## 快速开始

### 0. 前置

| 依赖 | 版本 | 用途 |
|---|---|---|
| Go | ≥ 1.27（见 `server/go.mod`） | 后端 |
| Node.js | ≥ 20（开发用 22） | 管理后台构建 |
| 微信开发者工具 | 最新稳定版 | 小程序预览 |
| Edge / Chrome | 任意近期版本 | **可选**，截图走查脚本用 |

### 1. 起后端（首次运行自动建库 + 写入演示数据）

```bash
cd server
go run ./cmd/xmf
# 监听 :8080，数据库落在 server/data/xmf.db
```

首次启动会 `AutoMigrate` 建表并写种子数据（4 款商品、4 块底板、24 个部件、3 个内置角色、管理员 admin/admin123）。想重置：停服务，删 `server/data/` 即可。

健康检查：

```bash
curl -i http://127.0.0.1:8080/api/v1/user/products   # 200
curl -i http://127.0.0.1:8080/ui/                   # 200（后台）
```

### 2. 管理后台

**开发（热更新）**

```bash
cd web
npm install
npm run dev      # http://localhost:5173/ui/ ，/api 自动代理到 :8080
```

**生产（构建后由 Go 托管）**

```bash
cd web && npm run build                  # 产物在 web/dist
cd ../server
go build -o xmf ./cmd/xmf                # Windows：-o xmf.exe
XMF_WEB=../web/dist ./xmf                # 访问 http://127.0.0.1:8080/ui/
```

> `XMF_WEB` 是相对**工作目录**解析的，所以要在 `server/` 下启动（`XMF_DB=data/xmf.db` 同理）。

**部署（Linux 单机，可选）**

```ini
# /etc/systemd/system/xmf.service
[Unit]
Description=xianmengfang
After=network.target

[Service]
WorkingDirectory=/opt/xmf/server
EnvironmentFile=/opt/xmf/server/.env      # 变量清单见 server/.env.example
ExecStart=/opt/xmf/server/xmf
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable --now xmf
# 前面加一层 Nginx：443 + 证书，proxy_pass http://127.0.0.1:8080;
# 小程序要求 HTTPS，且后台配置 request 合法域名
```

默认管理员：`admin` / `admin123` —— **上线前必须改密码**（`users` 表 bcrypt）。

### 3. 微信小程序

1. 微信开发者工具 → 导入项目 → 选择 `miniprogram/` 目录
2. 把 `miniprogram/project.config.json` 的 `appid` 换成**你自己的 AppID**（仓库里是作者开发用号，别人导入会提示无权限）
3. 确认 `miniprogram/config.js` 的 `BASE` 指向后端（默认 `http://127.0.0.1:8080`）
4. 详情 → 本地设置 → 勾选「不校验合法域名」（本地是 http）

没有微信账号也能跑通接口，用开发登录：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/user/auth/dev_login \
  -H 'Content-Type: application/json' \
  -d '{"openid":"dev_me","nickname":"测试用户"}'
# → data.token，后续请求带 Authorization: Bearer <token>
```

`XMF_DEV_LOGIN=0` 可关闭该后门（生产必关）。

## 配置项

服务端全部以 `XMF_` 前缀走环境变量，**密钥只走环境变量或 `server/data/` 本地文件，绝不进仓库**。

| 变量 | 默认 | 说明 |
|---|---|---|
| `XMF_ADDR` | `:8080` | 监听地址 |
| `XMF_DB` | `data/xmf.db` | SQLite 路径（相对**工作目录**） |
| `XMF_WEB` | 空 | 管理后台静态目录；为空则不托管 `/ui/` |
| `XMF_JWT_SECRET` | 自动生成 | 缺省时生成随机值持久化到 `data/.jwt_secret`（0600） |
| `XMF_DEV_LOGIN` | `1` | 开发登录接口开关，生产设 `0` |
| `XMF_WX_APPID` / `XMF_WX_SECRET` | 空 | 微信小程序登录；未配置时用 dev_login |
| `XMF_AI_API_KEY` / `XMF_AI_SECRET_ID` | 空 | AI 生图密钥；为空则回落 Mock（生成占位图，链路可走通） |
| `XMF_AI_BASE_URL` / `XMF_AI_MODEL` / `XMF_AI_STYLE` | 混元默认 | 换供应商只改这三个 + 实现一个 `Provider` |

**AI 密钥建议走后台配**：后台「AI 配置」页保存的 `configs.ai_api_key` **优先于环境变量**，请求时实时解析，改完即生效、不必重启。每日免费额度同理（`configs.ai_daily_free_limit`，默认 3）。

前端只有非敏感配置：`web/.env`（`VITE_API_BASE`、`VITE_APP_TITLE`），`web/.env.example` 为模板。

## 目录结构

```
.
├─ AGENTS.md            # AI 协作红线（给 Claude/Codex 等读的入口）
├─ docs/                # 唯一事实来源：PRD / API / DATA-MODEL / CONVENTIONS / CHANGELOG / TODO
├─ server/              # Go 后端
│  ├─ cmd/xmf/          # 唯一入口：config → db → repo → service → api 装配
│  ├─ internal/
│  │  ├─ api/           # 路由 + 参数绑定 + 鉴权 + 统一响应（薄层）
│  │  ├─ service/       # 业务规则唯一归属（状态机、库存、配额、积分、RBAC）
│  │  ├─ repo/          # CRUD 与事务；aipack/ 为 go:embed 的拼装素材
│  │  ├─ model/         # GORM 模型与迁移
│  │  ├─ config/        # 环境变量与默认值
│  │  └─ imageutil/     # 纯函数图像处理（合成、图标生成）
│  ├─ scripts/          # 联调/回归脚本（Python，不进编译产物）
│  └─ data/             # 运行时：db、上传文件、密钥（.gitignore）
├─ web/                 # Vue 3 管理后台（src/api · components · layouts · stores · views）
├─ miniprogram/         # 微信小程序（19 页 · components · utils/icons.js · custom-tab-bar）
├─ tools/               # 开发辅助：mp-preview（小程序离线预览截图）、web-shot（后台上台截图）
└─ docs/ + README.md
```

## 常用命令

```bash
# 后端
cd server
go build ./... && go vet ./...
go test ./...                                   # 全量单测（service / api）
go build -o data/xmf.exe ./cmd/xmf              # 出二进制

# 后台
cd web
npm run lint && npm run typecheck && npm run build

# 小程序（无构建，仅静态检查）
node --check miniprogram/utils/*.js
```

> Windows 提示：运行中的 `xmf.exe` 会被锁，**改了 Go 代码必须先停进程再 `go build`**，否则编译报「另一个程序正在使用此文件」。

## 设计上几个刻意的选择

| 决定 | 原因 |
|---|---|
| 金额一律 `int`，单位分 | 浮点算钱迟早差一分；字段名强制带 `_cents` |
| 库存用条件原子更新 | `UPDATE ... WHERE stock >= n`，禁止「读出来算好再写回」 |
| 订单状态机只在 service 层 | 非法流转返回 409，不静默修正；api 层出现业务 if/else 即视为坏味道 |
| 地址写进订单快照 | 之后改地址不影响历史订单，发货以快照为准 |
| 注销用 `open_id` tombstone | 释放唯一键，同一微信再登录是全新账号，不用人工解绑 |
| 素材 `go:embed` 进二进制 | 部署只拷贝一个文件，素材不会丢；后台上传真图可覆盖同名文件 |
| 小程序图标是运行时 SVG | `utils/icons.js` 内置 60+ 线性图标，按 name+color 生成 data URI，选中态只换色，不必维护两套 png |
| 后台复用 Go 的静态托管 | 单机部署少一个 Nginx/容器；要 CDN 时加一层反代即可 |

## 小程序预览工具（无需开发者工具）

没有微信开发者工具也能做视觉走查：`tools/mp-preview/` 把 WXML 渲染成 HTML（读真实 wxss，rpx→px），再用 playwright 截图。

```bash
cd tools && npm install                  # 只装 playwright-core；也可用 PLAYWRIGHT_CORE 指向已有副本
cd ..

node tools/mp-preview/make-data.js      # 生成 19 页假数据
node tools/mp-preview/render-all.js     # WXML → out/*.html
node tools/mp-preview/serve.js          # http://127.0.0.1:8123
node tools/mp-preview/shot-all.js       # shots/mp/*.png（可传页名过滤，如 pages_mine_mine）
```

后台截图：`node tools/web-shot.js dashboard users works`（需后端已启动，会自动登录）。

## 文档

| 文档 | 内容 |
|---|---|
| [docs/PRD.md](docs/PRD.md) | 需求、角色、流程、信息架构与视觉规范 |
| [docs/API.md](docs/API.md) | 全部接口、鉴权、权限码表 |
| [docs/DATA-MODEL.md](docs/DATA-MODEL.md) | 表结构、订单状态机、积分口径 |
| [docs/CONVENTIONS.md](docs/CONVENTIONS.md) | 目录落位、编码纪律、接口约定 |
| [docs/CHANGELOG.md](docs/CHANGELOG.md) | 逐版本变更与「为什么这么改」 |
| [docs/TODO.md](docs/TODO.md) | 已完成 / 待办 / 已知坑 |
| [docs/AI协作开发流程.md](docs/AI协作开发流程.md) | 本项目如何用 AI 协作开发（可复用的方法论） |

## 已知限制

- **支付是模拟的**：下单即到「待接单」，真实微信支付待资质齐备后替换（接口已隔离）
- **小程序未做真机全量自测**：自定义 tabBar 凸起按钮的点击区、canvas 拼装性能、全面屏安全区还需要在真机确认
- **SQLite 单机**：并发写有上限，工厂单量大时按 PRD 计划迁 PostgreSQL
- **没做 CI**：仓库目前靠本地门禁（`go test` + `npm run build`）

## 路线图

- [ ] GitHub Actions：Go 测试 + 前端构建 + 小程序 JSON/语法校验
- [ ] Dockerfile / docker-compose 一键起
- [ ] 真实微信支付（含退款）
- [ ] 消息中心（订单进度、被点赞、积分变动）
- [ ] PostgreSQL 迁移脚本

## 贡献

欢迎 issue / PR。提 PR 前请先读 [docs/CONVENTIONS.md](docs/CONVENTIONS.md)（目录落位、金额与库存纪律、接口命名），并保证本地门禁通过：

```bash
cd server && go vet ./... && go test ./...
cd web && npm run lint && npm run build
```

提交信息用 `feat/fix/docs/style/refactor` 前缀，正文写**为什么**改（而不只是改了什么）。一个功能一个 commit。

## 合规提醒

- 定制商品属《消法》第 25 条「消费者定作的商品」，**不适用七天无理由退货**，但必须在下单前明确告知（本项目在下单页固定展示不可关闭的提示）
- 用户上传与 AI 生成内容需过微信内容安全校验后才可作为作品使用
- 上线需 HTTPS 域名 + ICP 备案；小程序后台配置 request 合法域名

## License

[MIT](LICENSE) © 现梦坊 contributors
