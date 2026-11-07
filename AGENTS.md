# 项目规则（AGENTS.md）— 现梦坊

> AI 每次会话自动读取。改任何代码前先读 docs/PRD.md 和 docs/CONVENTIONS.md。人看的上手说明在 README.md。

## 新机器上手（三条命令）

```bash
cd server && go run ./cmd/xmf          # :8080，首次自动建库 + 种子（admin/admin123）
cd web && npm i && npm run dev         # :5173/ui/ ，/api 代理到 8080
# 小程序：微信开发者工具导入 miniprogram/，改 project.config.json 的 appid，勾「不校验合法域名」
```

改了 Go 代码必须重编译再重启（Windows 下运行中的 exe 被锁）。开发登录：`POST /api/v1/user/auth/dev_login {"openid":"dev_x", "nickname":"随便"}`。

## 文档索引（写之前先查，改完同步）

| 想知道 | 看 |
|---|---|
| 需求、信息架构、视觉规范 | docs/PRD.md（小程序部分见 3.5） |
| 接口、权限码 | docs/API.md |
| 表结构、订单状态机、积分口径 | docs/DATA-MODEL.md |
| 目录落位、编码纪律、接口命名 | docs/CONVENTIONS.md |
| 为什么这么改 / 有哪些坑 | docs/CHANGELOG.md、docs/TODO.md |

- 技术栈：Go 1.22+（Gin + GORM + SQLite，后期迁 PostgreSQL，禁止提前引入 MQ/Redis/ES）；前端为微信小程序原生 + Vue 3 + TS + Element Plus
- 分层：api 只做参数绑定和鉴权，repo 只做 CRUD，业务规则（金额、库存、状态机、配额）全部在 service 层
- 目录：新增文件必须按 docs/CONVENTIONS.md「目录结构 › 落位决策表」放；仓库根只允许 README.md / LICENSE / AGENTS.md + docs/ server/ web/ miniprogram/ tools/，找不到落点先问，不要新建散文件或平铺
- 金额一律 int（分），禁止 float 参与金额计算；数量 int
- 每次改动必须：`go vet ./... && go test ./...` 通过，Web 后端实测后才能报告"完成"；必须给出证据（命令输出/截图/断言），不接受"我改好了"
- 订单状态机见 docs/DATA-MODEL.md，状态流转只允许在 service 层发生
- 改完代码必须同步 docs/（API.md / DATA-MODEL.md / CHANGELOG.md / TODO.md）；一个功能一个 commit，提交信息写"为什么"
- 密钥（AI 生图 API key、JWT secret、小程序 secret）只走环境变量或 data/ 下的本地文件，绝不进仓库；data/、*.db 不提交。变量清单见 server/.env.example
- 仓库要能被别人 clone 就跑：**不写死本机绝对路径**（脚本依赖统一走 tools/pw.js 这类解析器）
- 用户上传内容与 AI 生成内容必须过微信内容安全校验后才可作为作品使用
