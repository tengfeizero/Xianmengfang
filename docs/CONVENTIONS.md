# CONVENTIONS — 现梦坊代码纪律

> 人定一次，AI 每次遵守。改动本文件必须经过用户确认。

## 目录结构（Go 后端）

```
server/
  cmd/xmf/          # 入口 main.go
  internal/
    api/            # 路由 + 参数绑定（薄层），不含业务规则
    service/        # 业务规则：金额、库存、订单状态机、AI 配额
    repo/           # CRUD，只操作数据库
    model/          # GORM 模型与迁移
    ai/             # AI 生图 provider 封装（可换厂商，不碰业务）
  data/             # 运行时数据（.db、密钥文件），不进 git
  docs/             # 项目文档（在仓库根 docs/）
web/                # Vue 3 + TS 管理后台
  src/
    api/            # http.ts（fetch 封装/401/ApiError）+ modules/（auth/orders/products/assets/configs，按域分模块）
    components/     # 可复用组件（StatusTag、AmountText）
    layouts/        # 布局（Layout.vue），与业务页分离
    router/         # 路由 + 登录守卫 + document.title
    stores/         # Pinia（auth 登录态），组件不直接碰 localStorage
    styles/         # 全局样式与主题变量
    types/          # DTO 类型 + 自动生成的 d.ts
    utils/          # auth(token 读写)/money(分↔元)
    views/          # 业务页面
  .env / .env.example  # VITE_API_BASE 等非敏感配置；密钥不进前端
  eslint.config.js  # ESLint9 扁平配置 + Prettier；npm run lint / format
miniprogram/        # 微信小程序原生
  app.js|json|wxss
  config.js         # 基址等集中配置
  components/       # 公共组件（empty-state、status-tag），页面经 usingComponents 引用
  images/tabbar/    # tabBar 图标（81x81，程序生成）
  pages/            # 每页四件套；命名用完整词（order-list/order-detail，不用易混缩写）
  utils/            # request(JWT/401)、util(金额/状态文案)
  project.private.config.json  # 个人本地配置，不进 git
```

## 前端纪律

10. **web 类型**：表格插槽的行类型用 `row as Type` 断言（Element Plus 插槽行是 DefaultRow）；DTO 一律放 `types/`，禁止在视图里重复声明。
11. **web 请求**：视图只 import `api/modules/*`，不直接写 fetch；金额换算只经 `utils/money`。
12. **Element Plus 按需引入**：新增组件直接写 `<el-*>` 即可（unplugin 自动注册），ElMessage/ElMessageBox 无需 import；不要退回全量 `app.use(ElementPlus)`。
13. **小程序组件**：两个以上页面重复的 UI 块必须抽到 `components/`；页面路径重命名时全量搜索旧路径引用。
14. **门禁**：web 提交前 `npm run lint && npm run build`；小程序提交前全量 `node --check` + JSON 校验。
15. **提交**：一个功能一个 commit；`feat/fix/docs` 前缀；信息写"为什么"。

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
