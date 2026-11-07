# TODO — 现梦坊

## 小程序视觉与信息架构重做 ✅ 已完成（2026-11-07，v2.0）
起因：小程序「能用但不像一个 App」——tab 里放了购物车和订单（低频、有入口可进），高频的广场却藏在二级；没有「我的」，资产与账号无处安放；视觉粉红单色、元素堆叠无层次。
- [x] 信息架构：5 tab 改为 **首页/广场/创作（中间凸起）/作品/我的**；购物车与订单移出 tab（首页角标 + 「我的」宫格进入）；广场升为一级（业务口径见 PRD 3.5）
- [x] 新增 7 页：我的、作品详情、搜索、积分明细、设置、帮助与反馈（FAQ 手风琴）、关于
- [x] 自定义 tabBar：白色毛玻璃条 + 中间 92rpx 渐变凸起按钮 + 白描边，选中态换色加深
- [x] 设计系统：`app.wxss` 重写（品牌紫/玫粉/主渐变/浅灰紫画布/大圆角双层阴影/间距字号规范），原子类 + 入场 `.rise` 延迟序列 + `pop`/`shimmer`/`breathe`
- [x] 图标体系：`utils/icons.js` 运行时 SVG（60+ 线性图标，任意着色改粗细）替代两套 png；新增 `x-icon`/`x-empty`/`x-skeleton` 三组件
- [x] 19 页视觉全部重做（首页 Hero、广场双列卡、创作画布与抽屉、作品数据带、商品画廊、订单时间轴、AI 额度条、登录呼吸 logo 等）
- [x] 后端补口：`GET/PUT /user/me`（资料 + 资产 + 订单角标一次取全）、`GET /user/works/:id/view`（含作者与 liked）、`GET /user/points` 分页；订单列表明细补缩略图字段
- [x] 设计走查工具 `tools/mp-preview/`：WXML→HTML + 真实 wxss + 19 页假数据 + playwright-core 截图（375×812 @2x 与整页长图）
- [x] 验证：`go build ./...` 通过；19 页截图逐页目视（无重叠/错位/截断）；脚本校验图标引用 100% 命中
- [ ] **真机自测（待用户用微信开发者工具导入 `miniprogram/`）**：自定义 tabBar 凸起按钮位置与点击区、canvas 拼装画布性能、下拉刷新、页面切换动画、安全区适配（iPhone 全面屏 / 安卓虚拟键）
- 备注：`app.json` 的 `tabBar.list` 仍保留 iconPath（自定义 tabBar 下不渲染，仅用于 tab 页识别），后续可清理；搜索复用了 `/plaza?keyword=`，未加 `/works/search`（会与 `/works/:id` 通配段冲突）
- 后续可选：广场分页游标化（当前 offset 分页，作品量大后深翻页变慢）、作品详情页支持横滑切换下一张、消息中心（点赞/订单通知）

## 作品社区 + 积分体系 ✅ 已完成（2026-11-07，v1.7）
起因：作品只能自己看自己用，没有创作正反馈；AI 生成只有每日限额、用完就断。
- [x] 数据：`works` 加 `is_public/published_at/prompt/like_count/use_count`；新增 `work_likes`、`point_logs`；`users.points`、`ai_tasks.pay_type`
- [x] 后端：发布/取消发布；作品广场（最新/使用量/点赞三榜 + 提示词搜索）；点赞/取消；我点赞的作品；积分余额与流水；今日配额查询
- [x] 后端：下单可选作品放宽为「自己的 或 已点赞的公开作品」（加购与结算同口径）
- [x] 后端：订单 `done` 时按件数累加使用次数并给作者加积分（自用也计，事务内幂等）
- [x] 后端：AI 生成免费额度用完后自动扣 1 积分（1 积分=1 张），积分不足 429；用积分不计入免费次数
- [x] 小程序：作品页发布开关 + 点赞/使用数；新增作品广场页（三榜、点赞、提示词）；商品页选图拆「我的/我点赞的」；AI 页显示剩余次数与积分
- [x] Web 后台：作品列表加公开状态/提示词/点赞/使用列与筛选；用户列表加积分列
- [x] 验证：新增 Go 单测 6 项（发布/点赞/广场排序/下单作品权限/完成结算/积分抵扣）+ `works_e2e.py` 全链路 20 步；lint/typecheck/build；后台两页截图核对
- 备注：积分目前只有「作品被使用」一个来源、只有「AI 生成」一个出口；后续可加签到/邀请/后台手工调整（`point_logs.reason=admin_adjust` 已预留）

## 素材按底板归属（底板/元素一对多）✅ 已完成（2026-11-07，v1.6）
起因：原来元素是全局池，小男孩底板的拼装页也会出现裙子、双马尾（素材是在小姑娘画布上抠的，混用必错位）。
- [x] 数据：`elements.board_id`（一对多）；老库迁移——board_id=0 的元素统一归到「底板·小姑娘」
- [x] 后端：用户端取元素必带 board_id（缺省 400）；compose 校验元素归属（跨底板 400「不属于当前底板」）；后台元素按 board_id 过滤/创建校验底板存在且上架；底板列表带 element_count；下架底板前必须先下架其启用元素（409 保护）
- [x] 演示数据：小姑娘底板挂全部 24 个部件；小男孩/小兔子/小熊猫只挂中性部件（帽子/表情/配饰，男孩再去掉蝴蝶结）——13/14/14 个
- [x] Web：「拼装素材」拆成「拼装底板」（元素数 + 管理元素入口 + 新增/编辑/下架）与「拼装元素」（按底板切换维护，board_id 写进地址栏可直达）两页
- [x] 小程序：拼装页按当前底板取元素，切底板清空已放置元素；空分类自动隐藏，空底板提示去后台配置；顺带修了首次进入不加载元素的 bug（pickBoard 伪造事件 + boardIdx 初始 0 被判为已选中）
- [x] 验证：go test 新增 5 个拼装归属单测；collage_e2e.py 重写（跨底板 400/底板元素集合断言/幂等）；lint+vue-tsc+build 通过；后台两页截图核对
- 备注：中性部件复用小姑娘画布素材，其他底板如果要专属素材，走后台按底板新增即可；AI 素材管线（scripts/ai_assets.py）产出的部件锚点按 400×500 画布计算，新底板保持同尺寸即可直接用

## 管理后台 UI 改版（电商后台风格）✅ 已完成（2026-11-07，v1.5）
起因：界面是「裸工具栏 + 裸表格」，无页头/卡片/图标，菜单平铺无分组。
- [x] 设计系统：styles/index.css 重写（品牌玫红 token、Element Plus 变量覆盖、panel/page-head 骨架类）
- [x] 布局：深色侧栏 + 分组菜单（交易/商品/内容/客户/系统）+ 图标 + 面包屑顶栏 + 账号下拉
- [x] 新增工作台首页 /dashboard（概览卡 + 订单进度 + 快捷入口 + 最近订单，按权限显示，前端聚合无新接口）
- [x] 9 个页面统一 PageHeader + panel；订单页状态页签带计数；登录页左右分栏品牌区
- [x] 验证：lint/vue-tsc 通过；Edge 无头截图逐页检查
- 后续可选：仪表盘统计改后端接口（现前端聚合）；移动端适配

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
