# CONVENTIONS — 现梦坊代码纪律

> 人定一次，AI 每次遵守。改动本文件必须经过用户确认。

## 目录结构

> 落位先于编码：新增文件先在下表找到位置。**找不到位置就是设计没想清楚**——先问，不要随手在根目录或 `src/` 下平铺。

### 0. 仓库总览

```
zxdz/   （现梦坊，单仓库三端）
├─ AGENTS.md        # AI 会话入口：只放红线，不写教程
├─ docs/            # 唯一事实来源（见 AI协作开发流程.md 文档体系表）
├─ server/          # Go 后端（Gin + GORM + SQLite）
├─ web/             # Vue 3 + TS 管理后台
└─ miniprogram/     # 微信小程序原生
```

`.workbuddy/`（本地 AI 日记）、`shots/`（实测截图）、`.playwright-cli/`、`*.genie`、`*.db`、`server/data/` 均**不进 git**（见 .gitignore）。

### 1. server/

```
server/
  cmd/xmf/main.go        # 唯一入口：config → db → repo → service → api 装配，不写业务
  internal/
    api/                 # 薄层：路由 + 参数绑定 + 鉴权 + 统一响应
      router.go          # 路由注册表（新接口必须在此登记）
      handlers.go        # 通用接口
      handlers_<域>.go   # 新增一律按域命名（如 handlers_address.go）
      handlers_m1.go     # 历史按里程碑分组，不再沿用此切法
      handlers_m2.go
      middleware.go      # JWT / 日志 / recover
      respond.go         # {code,msg,data} 统一出口
    service/             # 业务规则唯一归属：订单状态机、库存、AI 配额、地址、鉴权
    repo/                # 只做 CRUD 与事务内数据库操作，不含 if/else 业务判断
      db.go              # 连接 + AutoMigrate
      assets.go          # 素材/作品读写
    model/               # GORM 模型与迁移（models.go 按域分段）
    config/              # 环境变量与默认值，统一 XMF_ 前缀
    imageutil/           # 纯函数图像处理（拼装合成、tabBar 图标生成），不依赖业务
  scripts/               # 联调/回归脚本：m0_e2e.py、m1_e2e.py、mp_e2e.py；不进编译产物
  data/                  # 运行时：.db、上传文件、密钥（0600），不进 git
```

### 2. web/（Vue 3 + TS + Element Plus）

```
web/
  index.html / vite.config.ts / tsconfig.json / package.json
  .env / .env.example    # VITE_API_BASE 等非敏感配置；密钥绝不进前端
  eslint.config.js       # ESLint9 扁平配置；.prettierrc.json + .prettierignore
  src/
    main.ts              # 只做挂载：router + pinia + ElConfigProvider(中文)
    App.vue
    api/
      http.ts            # fetch 封装：baseURL、token 注入、401 清态、ApiError
      modules/           # 按域一文件：auth / orders / products / assets / configs
    components/          # 跨页复用组件（StatusTag、AmountText），不发请求
    layouts/             # 布局壳（Layout.vue：侧边菜单 + header）
    router/index.ts      # 路由表 + 登录守卫 + document.title
    stores/              # Pinia（auth）：token 的唯一内存持有者
    styles/index.css     # 全局样式与主题变量
    types/               # index.ts 手写 DTO；auto-imports.d.ts、components.d.ts 为生成物
    utils/               # 无状态工具：auth.ts(token)、money.ts(分↔元)
    views/               # 业务页面，一页一文件，文件名 = 路由名（PascalCase）
```

### 3. miniprogram/（原生）

```
miniprogram/
  app.js / app.json / app.wxss
  config.js              # 基址等集中配置，页面内禁止硬编码 URL
  components/            # 公共组件：empty-state、status-tag（externalClasses 定制外观）
  images/tabbar/         # tabBar 图标 10 张（5 项 × 普通选中，81×81，程序生成）
  pages/                 # 每页四件套 js/json/wxml/wxss，目录名 = 路由名
    index/ product/ studio/ ai/ works/ order-list/ order-detail/ address/ login/
  utils/                 # request.js（JWT/401/统一报错）、util.js（金额与状态文案）
  project.config.json / sitemap.json
  project.private.config.json   # 个人本地配置，不进 git
```

### 4. 新增文件落位决策表

| 要加的东西 | 落在哪 | 同步 |
|---|---|---|
| 新接口 | `api/handlers_<域>.go` + `api/router.go` 的对应 group（user/admin）注册 | API.md |
| 业务规则 | `service/<域>.go`（先写 `_test.go`） | CONVENTIONS/文档口径 |
| 新表/字段 | `model/models.go` + `repo/<域>.go` | DATA-MODEL.md |
| 新配置项 | `internal/config/config.go`（`XMF_` 前缀）+ `.env.example` | TODO/CHANGELOG |
| 新管理页 | `web/src/views/Xxx.vue` + `router/index.ts` + `api/modules/xxx.ts` + `types/index.ts` | — |
| 跨页组件 | `web/src/components/` 或 `miniprogram/components/`（**两个以上页面用到才抽**） | — |
| 新小程序页 | `pages/<name>/` 四件套 + `app.json` 注册 | 全量搜索旧路径 |
| 联调/一次性脚本 | `server/scripts/`（不进 main 包、不进前端） | — |
| 文档 | `docs/` 对应文件，与代码同一次提交 | — |

## 硬性纪律

1. **金额**：`int`，单位分。字段名 `price_cents / total_cents`。任何地方禁止 float/money 类型。
2. **库存**：扣减必须用带条件的原子更新（`UPDATE ... WHERE stock >= n`），禁止"读出来算好再写回"。库存流水只追加不修改。
3. **状态机**：订单状态只允许合法流转（见 DATA-MODEL.md），非法流转返回 409，不静默处理。
4. **业务规则**只写在 service 层；api 层发现 if/else 业务判断即为坏味道。
5. **AI 配额**：每用户每日免费次数从配置读取（默认 3），计数持久化，不做内存计数。
6. **错误返回**：统一 `{code, msg, data}`；msg 必须是人话，禁止把 500 internal error 直接抛给用户。
7. **密钥**：环境变量优先；本地开发用 `data/` 下文件（0600），两者都不进 git。
8. **测试**：业务规则先写测试再实现；演示数据要自洽（订单、库存、金额对得上）。
9. **提交**：一个功能一个 commit；`feat/fix/docs` 前缀；信息写"为什么"。

## 前端纪律

10. **web 类型**：表格插槽的行类型用 `row as Type` 断言（Element Plus 插槽行是 DefaultRow）；DTO 一律放 `types/`，禁止在视图里重复声明。
11. **web 请求**：视图只 import `api/modules/*`，不直接写 fetch；金额换算只经 `utils/money`。
12. **Element Plus 按需引入**：新增组件直接写 `<el-*>` 即可（unplugin 自动注册），ElMessage/ElMessageBox 无需 import；不要退回全量 `app.use(ElementPlus)`。
13. **小程序组件**：两个以上页面重复的 UI 块必须抽到 `components/`；页面路径重命名时全量搜索旧路径引用。
14. **门禁**：web 提交前 `npm run lint && npm run build`；小程序提交前全量 `node --check` + JSON 校验。

## 目录纪律

15. **不许在仓库根新建散文件**：临时脚本、测试 yaml、截图、日志一律进 `server/scripts/` 或 `.gitignore` 覆盖的目录（`shots/`、`.workbuddy/`）。根目录只允许 AGENTS.md + docs/ server/ web/ miniprogram/。
16. **不许在 `web/src/` 平铺**：新页面必进 `views/`，新请求必进 `api/modules/`，新类型必进 `types/`；`views/` 里出现裸 fetch、内联金额换算都算违规。
17. **生成物与依赖不手改不提交**：`web/dist/`、`web/node_modules/`、`src/types/auto-imports.d.ts`、`src/types/components.d.ts`（unplugin 生成；要跑 vue-tsc 须先 build）。
18. **`internal/` 不新增无业务包**：纯工具先做对应域包的私有函数；确需独立包（如图像处理）必须无业务依赖且有单测。
19. **目录名与路由/接口名一致**：小程序页面目录名 = `app.json` 路由 = 后端 resource 名（`order-list` ↔ `/orders`）；不用易混缩写（故不用 `order`，用 `order-detail`）。

## 接口约定

20. **一端一个命名空间**：用户端 `/api/v1/user/*`（含登录与商品浏览等公开接口），管理后台 `/api/v1/admin/*`；`/api/v1` 根下不挂业务接口。新增接口必须归入其一，路由只在 `api/router.go` 的对应 group 里注册。
21. **路径形态**：资源用复数（`/orders`、`/addresses`），状态变动用子路径 `POST`（`/orders/:id/confirm`），禁止 `/confirmOrder` 这类动词根路径；前端各端只认自己的前缀（小程序 `config.js` 的 `API`、web 的 `VITE_API_BASE`），页面/module 内只写资源段。
22. **管理端接口必须挂权限码**：`/admin/*` 下每条路由都要 `PermRequired("<模块>:<动作>")`，无权限码的路由等于对所有人开放。新功能先加权限码（`service.PermGroups()`）再挂路由，并同步 API.md「权限码」表；**权限码只增不改**——改名会让已存在角色的 `perms` 静默失效。前端按钮/.菜单同步用 `auth.hasPerm()` 过滤，但**权限判定只认后端**，前端过滤仅作展示。
23. **敏感操作必须留痕**：管理端写操作（`POST/PUT/DELETE`）由 `AuditWrite` 中间件统一记录，新增动作需在 `service/audit.go` 的 action 映射里登记中文动作名与 `target_type`；越权拦截（403）也写 `deny` 日志。**日志里禁止出现明文密钥**：新增字段若含 password/secret/token/api_key 语义，必须走 `MaskDetail` 脱敏（有单测锁定）。
