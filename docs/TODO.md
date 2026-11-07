# TODO — 现梦坊

## RBAC 多用户权限 + 敏感操作审计日志 ✅ 已完成（2026-11-07，v1.4）
起因：全站只有 `admin/admin123` 一个全权账号、操作无留痕，多人协作时客服能改 AI 密钥、运营能删账号，出事无法追溯。
- [x] 权限模型：新增 `roles` 表（code/name/perms JSON/builtin/status），`admins` 加 `role_id`/`nickname`/`last_login_at`；权限码按模块细分 20 个（订单 5 / 商品 2 / 素材 2 / 用户 2 / 作品 2 / 配置 2 / 权限与日志 5，见 API.md「权限码」）
- [x] 鉴权链重构：JWT 只带 admin_id，`AdminAuthRequired` 每请求查库载入账号与角色 → `PermRequired(code)` 按权限码拦截；**禁用账号、改权限立即生效**，不等 token 过期
- [x] 内置角色：super（`["*"]`，权限不可改）、ops、viewer；支持自建（如只给订单查看+发货的仓库角色）
- [x] 保护规则：不能禁用/降级/删除自己；必须保留至少一个启用超管；内置角色不可删、在用角色不可删（均 409）
- [x] 审计：新增 `audit_logs` 表 + `AuditWrite` 中间件，记录管理端全部写操作、登录成功/失败、越权拦截（deny）；password/secret/token/api_key 及 `{key,value}` 形态的敏感配置值自动脱敏
- [x] 接口：`/admin/me`、`/admin/permissions`、`/admin/roles*`、`/admin/admins*`、`/admin/audit-logs`
- [x] Web：`config/menu.ts` 单一菜单源 + 路由 `meta.perm` + `auth.hasPerm` 控制按钮；新增「角色权限」「管理员账号」「操作日志」三页
- [x] 验证：`scripts/rbac_e2e.py` 覆盖越权/即时生效/保护规则/日志脱敏；`go test ./...` 全绿（修好 v1.3.5 遗留的 order_test）
- 遗留：token 黑名单（目前禁用靠每请求查库已失效，无需黑名单）；上线前必须改掉 admin123 并按下岗位拆分账号

## 接入真实 AI 生图（腾讯混元 Hy Image 3.5 preview）✅ 已完成（2026-11-06，v1.3）
选定模型：腾讯混元 `hy-image-v3.5-preview`（一个模型同时支持文生图 + 参考图指令编辑，正好对应 ai_text / ai_photo 两种任务；按张计费约 0.11 元/张，只对输出收费）。
- [x] `service.HunyuanProvider`：实现 Provider 接口两方法；messages 格式（腾讯云 TokenHub）为主，预留 images 格式（OpenAI 风格端点）；结果 URL 立即下载落盘（返回的 URL 12 小时过期），统一转码 PNG
- [x] **后台可配置**（用户要求，密钥不进代码仓库）：`configs` 表新增 `ai_api_key` / `ai_base_url` / `ai_model` / `ai_style` 四键，后台系统配置页可改；密钥列表脱敏展示（前5后4），留空 = 停用真实生图回落 Mock；AIService 每次任务前 resolveProvider（DB 配置 → 注入兜底）
- [x] 单测：httptest 假服务验证请求结构与图片落盘；resolveProvider 解析与回落；配置校验/脱敏；配额逻辑不变（ai_daily_free_limit）
- [x] 真实 key 联调（8080 主库实测）：文生图（水彩柴犬咖啡店）+ 照片转风格（简笔画→卡通插画，构图保持）全通；m0/m1/mp/cart 四套 e2e 在隔离 Mock 实例全过
- 备注：国内站端点为 `https://tokenhub.tencentmaas.com`（Bearer 用控制台 sk- 开头的 SecretKey，ak- 是 Key ID 不参与鉴权）；回归测试须用隔离实例（无 key → Mock），避免真实计费

## 购物车 ✅ 已完成（2026-11-06，v1.0，规划见 PRD 3.2.2）

- [x] 后端：cart_items 模型（唯一键 user+sku+work、checked 默认 1）+ CartService + 8 个接口（列表/加购/单项改/批量勾选/删单项/清空/checkout）
- [x] checkout 复用 OrderService.Create 一单多明细，成功后移除对应项；失效项拦截
- [x] 小程序：tabBar 第 5 个「购物车」页（勾选/全选/数量步进/失效标灰/删除/地址选择/结算 + tabBar 角标）；商品页加「加入购物车」双按钮；tabBar 图标程序生成
- [x] e2e：cart_e2e.py（加购累加 → 勾选结算一单多明细 → 金额对账 → 项移除 → 结算缺货 409 → 越权 403 → 清空）
- [x] 回归：go vet/test、m0/m1/mp 三套 e2e、小程序全量校验；CHANGELOG/CONVENTIONS 同步

## 已完成
- [x] 需求澄清，PRD v0.2（docs/PRD.md）
- [x] 定名「现梦坊」；技术栈 Go + SQLite（后期 pgsql）
- [x] 冷启动地基：AGENTS.md / CONVENTIONS.md / DATA-MODEL.md / API.md / TODO / CHANGELOG

## 当前迭代：M0 地基 + 最小闭环竖切 ✅ 已完成（2026-11-06，v0.2）
- [x] Go 工程初始化（cmd/xmf + internal 分层 + AutoMigrate 幂等迁移 + 演示数据）
- [x] 用户微信登录（code2session 已实现，待配置 appid）+ 开发者模拟登录 + JWT
- [x] 商品与 SKU 表 + 演示数据
- [x] 下单接口（service 层金额/库存校验，mock 支付直接进 pending_accept）
- [x] 订单状态机 + order_events 流水
- [x] 后台管理员登录（JWT）
- [x] 后台订单列表/接单/发货/取消/退款标记接口
- [x] 确认收货接口
- [x] **竖切验收**：go vet/test 通过；临时库实测 下单→接单→发货→确认收货 全链路，金额库存对账一致（回归脚本 server/scripts/m0_e2e.py）

## M1 后端切片 ✅ 已完成（2026-11-06，v0.3）
- [x] imageutil 合成引擎 + 演示素材程序化生成
- [x] 拼装合成接口（服务端合成，作品入库）
- [x] AI Provider 框架 + MockProvider + 每日配额（ai_daily_free_limit 后台可改）
- [x] 上传 + 作品库（软删、归属校验）
- [x] 后台 boards/elements CRUD
- [x] 回归脚本 server/scripts/m1_e2e.py 全链路通过

## M1.4 Web 管理后台 ✅ 已完成（2026-11-06，v0.4）
- [x] 后端：商品/分类/SKU CRUD（删除保护）、configs 白名单读写、素材列表、admin 上传
- [x] Vite 工程初始化 + 登录页 + 布局
- [x] 订单管理页（列表/筛选/接单/发货/取消/退款标记）
- [x] 商品管理页（分类/SPU/SKU/库存，元分换算）
- [x] 素材管理页（底板/元素上传与锚点编辑）
- [x] 系统配置页（AI 每日配额）
- [x] vue-tsc + npm run build 通过；playwright(Edge) 浏览器实测全流程通过（截图 shots/）
- [x] Go 单二进制托管 /ui/（XMF_WEB=../web/dist）

## M1.5 小程序端（原生） ✅ 已完成（2026-11-06，v0.5）
- [x] 原生工程 miniprogram/（官方脚手架工具模板哈希错误×2，降级手工搭建基础文件）
- [x] 登录页（wx.login 真登录优先，失败降级开发者模拟登录）
- [x] 商店首页/商品详情（SKU 选择、数量、绑定作品下单）
- [x] 创作工坊（canvas 2d 拼装：选底板/分类选元素/拖动/缩放/移除，提交 placements 服务端合成）
- [x] AI 生成页（文生图 + 照片转风格 wx.chooseMedia→upload→ai/tasks）
- [x] 作品库（网格预览、大图预览、删除）
- [x] 订单列表（状态筛选/取消/确认收货）+ 订单详情（物流信息）
- [x] 校验：全部 JS node --check 通过、JSON 格式通过；接口契约联调脚本 server/scripts/mp_e2e.py 全链路通过
- [x] WorkBuddy 应用预览卡已创建（未发布，发布从预览面板「分享」进入）

## 收货地址管理 ✅ 已完成（2026-11-06，v0.6）
- [x] 后端 addresses CRUD + 设默认（唯一默认/上限 20/软删补位/手机号校验）
- [x] 下单必传 address_id，收货信息快照进订单
- [x] 小程序地址列表/编辑页、商品页选地址、订单展示收货人
- [x] 7 组地址单测 + 三套 e2e 回归全过

## 前端工程规范化 ✅ 已完成（2026-11-06，v0.7）
- [x] web：ESLint9/Prettier、.env、layouts/api/components/stores/types/utils 分层、Pinia、Element Plus 按需引入
- [x] 小程序：components 抽取、tabBar 图标、order-list/order-detail 重命名、私有配置出库
- [x] 三套 e2e + 浏览器截图回归全过
- [x] 目录规范入档：CONVENTIONS 三端目录树 + 落位决策表 + 目录纪律 15-19；AGENTS.md 增落位硬规则

## API 命名空间按端拆分 ✅ 已完成（2026-11-06，v0.8）
- [x] 用户端统一 `/api/v1/user/*`（含登录、商品浏览等公开接口），后台保持 `/api/v1/admin/*`
- [x] 小程序前缀集中到 `config.js` 的 `API`，页面资源路径零改动
- [x] 三套 e2e 路径同步 + go vet/test + m0/m1/mp 全过；8080 已重启验证新路径

## 订单详情补全 ✅ 已完成（2026-11-06，v0.8.1）
- [x] 详情接口返回商品图/作品图、状态流水 events、金额明细 amounts；归属校验移入 service
- [x] 小程序详情页：状态头/时间轴/商品图/金额明细/收货与物流可复制/底部操作栏/下拉刷新
- [x] mp e2e 增加详情字段断言；8080 已换新二进制

## 商品信息补全（P0）✅ 已完成（2026-11-06，v0.9）
规划见 PRD 3.2.1。分三批，本批只做"看得见"的展示字段。
- [x] 模型：Product 增 subtitle / original_price_cents / unit / detail / detail_images / service_notes / sort
- [x] 演示数据：4 款商品补真实文案、卖点、单位、服务承诺；程序生成 8 张封面/特写图；老库按标题回填
- [x] 接口：admin 商品 create/update 支持新字段；用户端列表按 sort desc, id 排；详情返回新字段
- [x] 后台：商品弹窗从 3 个输入扩成完整表单（多图上传 + 详情图上传 + 划线价 + 排序）
- [x] 小程序：商品页加多图轮播、副标题、划线价、单位、详情图文、服务承诺；列表页显示副标题与划线价
- [x] 合规：下单前固定展示"定制商品不支持七天无理由退货"，不可关闭
- [x] 回归：go vet/test、web lint+build、小程序语法校验、三套 e2e

## 商品信息增强（P1）· 待排期
- [ ] `attrs` 属性组（材质/尺寸/重量/工艺/产地/洗涤）→ 详情页参数表
- [ ] `skus.image_url` SKU 图（换色换图）；划线价下沉到 SKU
- [ ] `tags` 标签、`min_qty`/`max_qty` 起订与限购、`lead_time_days` 生产周期、`share_cover` 分享封面

## 商品信息增强（P2）· 待排期
- [ ] 运费模板（满 X 包邮 / 按重量 / 按地区），需要独立表并改下单金额计算
- [ ] 评价体系（新表 + 晒图 + 审核）
- [ ] 销量冗余字段与汇总任务、商品视频、富文本详情编辑器

## Backlog
- [ ] 真实微信支付（等主体+商户号资质）
- [ ] 多租户改造
- [ ] 退款售后流、物流跟踪、评价
- [ ] PostgreSQL 迁移
- [ ] 服务器采购 + 域名 ICP 备案 + HTTPS（用户负责，尽早启动）
