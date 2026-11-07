# CHANGELOG — 现梦坊

## v1.4 (2026-11-07) — RBAC 多用户权限 + 敏感操作审计日志

原来只有一个 admin 账号、全权限、无留痕，多人协作既不安全也无法追责。补齐两层能力：

**权限（RBAC）**
- 新增 `roles` 表（code/name/perms/builtin/status）与 `admins.role_id`；权限码按模块细分 19 个（订单 5 / 商品 2 / 素材 2 / 用户 2 / 作品 2 / 配置 2 / 权限与日志 5）
- 内置角色：超级管理员（`["*"]`，不可改权限）、运营专员、只读账号；支持自建角色
- 鉴权链改为 JWT → 端校验 → **每请求查库载入账号与角色** → 权限码拦截：禁用账号、调整角色权限**立即生效**（不再等 token 过期）
- 保护规则：不能禁用/降级/删除自己；系统必须保留至少一个启用超管；内置角色不可删、在用角色不可删
- 新增接口：`/admin/me`、`/admin/permissions`、`/admin/roles*`、`/admin/admins*`；每条管理端路由按权限码挂中间件
- Web：菜单与操作按钮按权限过滤，路由守卫拦截无权限页面；新增「角色权限」「管理员账号」两页

**审计日志**
- 新增 `audit_logs` 表；`AuditWrite` 中间件统一记录管理端写操作（动作映射 order.ship / config.update / role.delete 等），另记登录成功/失败、越权拦截（deny）
- 每条含操作人、动作、目标、IP、UA、结果、状态码、失败原因、请求摘要；**password/secret/token/api_key 字段自动脱敏为 `***`**，日志不会成为泄密面
- 新增 `GET /admin/audit-logs`（按操作人/动作前缀/结果/关键字筛选 + 分页）；Web 新增「操作日志」页，展开行看请求摘要与 UA
- 实测（`scripts/rbac_e2e.py`）：发货员仅能看单+发货，接单/配置/账号管理均 403；禁用后旧 token 立即 401；收回权限即时 403；内置/在用角色与自我禁用均 409；日志含 9 类动作、密码已脱敏

## v1.3.5 (2026-11-07) — 订单状态补「生产中/打包中」+ 后台列表去掉折叠行

工厂是定制生产流程，接单后到发货前的「生产、打包」两段此前无状态可标，用户也看不到进度。调整：

- 状态机扩展：`pending_accept → producing（生产中） → packing（打包中） → pending_ship（待发货） → shipped → done`；历史停在 `pending_ship` 的订单可直接发货，无需补动作
- 新增接口 `POST /admin/orders/:id/pack`（生产完成）、`POST /admin/orders/:id/ready_ship`（打包完成）；`accept` 语义改为「接单 → 生产中」
- 取消权限：管理员可取消至待发货（生产中/打包中也允许，库存自动回补）；用户仍只能取消待接单
- Web 订单页：**去掉列表展开行**（明细统一进详情弹窗），状态筛选补两项，操作按钮按状态给出「接单 / 生产完成 / 打包完成 / 发货」
- StatusTag 改为分状态独立配色（生产蓝、打包紫、待发货青），生产与打包不再与待接单同色
- 小程序：订单列表 tab 增加「生产中/打包中」，详情页与流水文案同步（如「工厂正在制作你的作品」）
- e2e 脚本（m0/m1/mp）流转步骤补 pack + ready_ship；实测 8080 全链路：接单→生产→打包→待发货→发货，流水 5 条完整，非法重复流转 409

## v1.3.4 (2026-11-07) — 管理端订单详情增强

后台订单列表此前只有展开行（明细文字 + 物流 + 收货），看不到买家是谁、状态流转过程、金额构成和商品图。补齐：

- 后端新增 `GET /admin/orders/:id`：复用用户端 `orderDetailView`（明细含商品首图/绑定作品图、物流、状态流水、金额明细），附加买家信息（昵称/头像）
- 订单列表增强：行内附买家昵称头像（批量查询防 N+1）；支持 `order_no` 模糊搜索与 `user_id` 过滤；补「买家」「下单时间」列
- Web 订单页新增「详情」弹窗：基础信息（单号/状态/支付方式/下单支付时间/收货快照/取消原因）、金额明细（商品合计/运费/优惠/实付）、商品明细表（图片可点击放大）、物流、状态流水时间线（操作者自动识别用户/管理员）
- 实测 8080：详情字段齐全（amounts 6900 分、events 2 条、商品图与作品图均返回）、单号搜索命中正常

## v1.3.3 (2026-11-06) — 管理端作品管理

后台此前看不到用户作品，无内容治理手段。补齐：

- 后端：`GET /admin/works`（全站作品 + 作者昵称，type/status 过滤）、`PUT /admin/works/:id/status`（上架 1 / 下架 2）
- 下架后用户端作品列表/详情/加购绑作品均不可见（复用 status=1 过滤，含 cart checkout 校验）；用户自删（status=0）的作品后台不可变更状态（409 保护）
- Web 后台新增「作品管理」页：图片预览（点击放大）、类型、作者、生成参数（prompt）、状态、下架/恢复操作
- 实测 8080：13 作品列出、过滤正常、新拼作品下架后用户端消失、自删后改状态被 409 拒绝

## v1.3.2 (2026-11-06) — 管理端用户管理

后台此前完全看不到用户（无接口也无页面）。补齐：

- 后端：`GET /admin/users`（列表含订单数/作品数统计，支持昵称模糊搜索与状态过滤）、`PUT /admin/users/:id/status`（启用/禁用）
- 禁用即拒绝登录（`auth.findOrCreateUser` 已有 Status 校验）；注意已签发的 JWT 在过期前仍有效，如需强制下线需引入 token 黑名单（后续项）
- Web 后台新增「用户管理」页：头像/昵称/订单数/作品数/注册时间/状态，禁用需二次确认
- 实测 8080：列表 6 用户、过滤搜索、禁用后登录 400 被拒、再启用，全通

## v1.3.1 (2026-11-06) — AI 配置补 Secret ID（ak）

TokenHub 密钥是 AK（Key ID）+ SK（SecretKey）一对，v1.3 配置页只放了一个 API Key。补齐：

- `configs` 新增 `ai_secret_id`：后台系统配置页可填（脱敏展示、留空不修改），随「清空密钥」一并清空
- `HunyuanProvider` 增 `SecretID` 字段；当前 TokenHub 端点 Bearer 鉴权只需 sk（SecretKey），ak 当前用不到但留档，为将来签名类端点预留
- 环境变量对应 `XMF_AI_SECRET_ID`；老库重启自动回填该配置行
- 实测 8080：写入/脱敏（ak-20****2ed0）/清空/写回全通

## v1.3 (2026-11-06) — 接入真实 AI 生图（后台可配置）

「文字生成图」「照片转风格」此前为 MockProvider 占位实现，产品核心卖点不可真实演示。选定**腾讯混元 Hy Image 3.5 preview**（一个模型同时覆盖文生图 + 参考图指令编辑，约 0.11 元/张，只对输出收费）。

- `service.HunyuanProvider` 实现 Provider 接口（业务代码零改动）：文生图直发 prompt；照片转风格 = 本地照片转 data URL 参考图 + 固定风格化指令（保持构图只改风格）；结果 URL 立即下载并统一转码 PNG 落盘（临时 URL 12 小时过期）；支持 messages（TokenHub 官方）/ images（OpenAI 风格）两种请求格式
- **后台可配置**（用户要求）：`configs` 新增 `ai_api_key` / `ai_base_url` / `ai_model` / `ai_style`，后台系统配置页可直接改，立即生效；密钥脱敏展示（前 5 后 4），**留空 = 停用真实生图回落 Mock**；密钥只存服务器 DB，不进代码仓库
- 兜底链：DB 配置 → 环境变量注入 Provider（`XMF_AI_*`）→ Mock，未配置密钥时现有开发流程完全不受影响
- 联调实测（8080 主库）：文生图与照片转风格全通，效果远超占位图；四套 e2e 在隔离 Mock 实例全过（回归不走真实计费）
- 踩坑：GORM 复用已填充结构体做 `First` 会把上次主键带进 WHERE（连续读多行配置必须用新变量）；TokenHub 国内端点是 `tokenhub.tencentmaas.com`，Bearer 鉴权只认控制台 SecretKey（sk- 开头），Key ID（ak-）不参与请求头；清空已为空的配置值时 GORM `Update` 影响行数为 0 会误判 404，改查后 `Save`

## v1.2 (2026-11-06) — 拼装素材 AI 生成换代 + 图层顺序

v1.1 扩了数量但素材仍是程序几何图，用户反馈"拼出来很丑"。参照游戏换装系统（分层纸娃娃）的做法整体换代：**素材好看的关键不是架构而是资产生产方式**——游戏是美术先画好整套统一风格的立绘再拆层，任意组合都协调。

- **差分抠图管线**（`scripts/ai_assets.py` + `scripts/extract_elements.py`）：以小姑娘立绘为基准 → img2img"给角色戴上道具"（input_fidelity=high 保持构图）→ 与底板像素差分 + 形态学开运算 + 连通域过滤 → 部件透明 PNG；**锚点由差分包围盒自动算出**，不再手工猜坐标
- 4 张底板 img2img 换角色生成（小姑娘为基准，小男孩/小兔子/小熊猫构图一致）；白底→透明用边界泛洪（不误伤角色内部白色），右下角水印先抹除
- 素材 4 底板 + 24 部件（hair×5、hat×5、lash×1、expression×4、outfit×4、accessory×5），内嵌 `go:embed internal/repo/aipack`（1.8MB），新库启动自动落盘；波波头/长直发/可爱睫毛为旧遗留记录，种子时删除，演示元素同名记录直接更新文件与锚点
- **合成 z-order**：服务端按分类固定图层（服装<头发<表情<睫毛<帽子<配饰，`categoryZ`），小程序预览同序绘制，用户添加顺序不再影响遮挡
- 发型类素材全部为"叠加式"生成（在底板发型上加马尾/发髻），规避差分只能抠"新增像素"的限制
- 返工 7 件（爱心眼/嘟嘟嘴/侧马尾/星星睫毛/围裙/披风/气球）：小部件差分阈值下调，提示词避免细线（气球去绳）与大面积替换
- 踩坑：img2img 会微重绘整图产生差分噪点（白边碎屑），仅连通域过滤不够，需加开运算；`FillRing` 打穿下层问题见 v1.1
- e2e 更新：24 元素断言 + 小姑娘整套换装乱序合成（验证 z-order）全过

## v1.1 (2026-11-06) — 拼装素材扩充

规划见 PRD 3.1.1。此前仅 1 底板 + 6 元素，小程序「表情」「服装」两个分类 tab 为空。

- 底板 1 → 4：新增小男孩（短发蓝衣）、小兔子（长耳白裙）、小熊猫（圆耳黑眼圈），脸部几何与底板·小姑娘统一（头心/眼/嘴同坐标），元素锚点跨底板通用
- 元素 6 → 26：头发 +4（长直发/丸子头/卷发/双丸子）、帽子 +3（草帽/皇冠/圣诞帽）、睫毛 +1（星星睫毛）、表情 +4（笑眼弯弯/爱心眼/眨眼/嘟嘟嘴）、服装 +4（围裙/披风/围巾/领结）、配饰 +4（圆框眼镜/气球/棒棒糖/星星发夹）
- 表情元素带肤色底垫，可自然盖住底板原眼睛/嘴，实现"换表情"
- 种子逻辑从「元素表非空即跳过」改为**按名字幂等补插**（底板同）：老库启动自动获得新素材，不覆盖后台已建/已改内容；新素材 PNG 缺失时程序生成，已存在不覆盖
- 踩坑：`imageutil.FillRing` 会把环内擦成透明——叠在实心图形上画环（如草帽帽带）会打穿下层，需改用双圆叠加
- 新增 `server/scripts/collage_e2e.py`：老库回填数量与分类断言 → 空分类拉取 → 小熊猫底板 + 新元素合成 → 作品入库 → 种子幂等
- 修复：创作页无法切换底板——底板行 `wx:if="boards.length > 1"` 此前只有 1 个底板从未渲染，点击事件实际是 `bindtap` + `data-idx`，`pickBoard` 却按 picker 组件读 `e.detail.value`（恒 undefined）；改读 `e.currentTarget.dataset.idx` 并加越界/重复选择保护

## v1.0 (2026-11-06) — 购物车

规划与决策记录见 PRD 3.2.2（tabBar 第 5 入口 / 服务端存储 / 勾选部分结算 / 每项可绑作品，四项均已确认）。

- 新表 `cart_items`：唯一键 (user_id, sku_id, work_id)，`work_id=0` 表示未绑作品（SQLite 唯一索引对 NULL 互不相等，禁用 NULL）；`checked` 存服务端跨设备一致；不存快照、不锁库存（列表实时 join，失效项 `valid=false` 由用户手动清理）
- 接口 8 个（`/api/v1/user/cart`）：列表（含勾选汇总）/ 加购（同款同图累加数量，单件 ≤99、≤50 条）/ 单项改数量勾选 / 批量勾选（空 item_ids = 全选）/ 删单项 / 清空（?checked=1 只清勾选）/ **checkout**（复用 `OrderService.Create` 一单多明细，成功后移除对应项；删除失败仅残留不影响订单）
- `binding:"required"` 坑：对 int 0 也判缺失，勾选 0 被拒——改用 `*int`（cartCheckReq.Checked）
- GORM 坑重现：`SKUID` 被推成 `skuid` 列（OrderItem 踩过）——`CartItem.SKUID` 显式 `gorm:"column:sku_id"`
- 小程序：tabBar 新增「购物车」（第 2 位，程序生成图标）；购物车页 = 地址选择 + 商品项（勾选/数量步进/删除/失效标灰）+ 全选 + 结算栏，tabBar 角标显示勾选件数；商品页底部改双按钮（加入购物车 / 立即购买）；列表以服务端为唯一事实源，任何变更后整单重拉
- 新增 `server/scripts/cart_e2e.py`：加购累加 → 作品绑定 → 批量勾选 → 结算一单 2 明细 35600 分对账 → 结算缺货 409（两用户先后购买制造"加购时够、结算时缺货"）→ 越权 403 → 清空
- 回归：go vet/test、m0/m1/mp 三套 e2e、小程序全量 JS/JSON 校验全过；8080 已换新二进制

## v0.9.1 (2026-11-06) — 商品页默认选中首个规格

- 商品详情页加载后默认选中第一个有货 SKU（全部缺货则选第一个并提示），不再要求用户先点一次；单规格商品此前虽自动选中，但合计一直显示 `--`（`computeTotal` 只在点选/改数量时触发）
- 顶部价格从"最低售价 + 起"改为随选中规格变化——默认选中后两者若不联动，会出现"标价 59 起、应付 99"的误导；划线价改为与当前规格价比
- 合计改用 `price_cents` 计算（原来 `parseFloat(priceText)*100`，9.9 这类价格有浮点误差）
- 后端 SKU 查询统一 `ORDER BY id`（用户端列表/详情、后台列表）：让"第一个规格"等于后台录入顺序，不再依赖 SQLite 的返回顺序
- 无库存规格：点击与提交均提示"该规格暂时缺货"，不进入下单流程

## v0.9 (2026-11-06) — 商品信息补全（P0）

规划与 P1/P2 后续批次见 PRD 3.2.1。此前 `products` 只有分类/标题/描述/images/上下架 5 个字段，后台弹窗只有 3 个输入框且没有图片上传入口，演示数据 `images` 全空——小程序商品页实际是"一张空封面 + 一句话"。

- Product 模型新增 7 个字段：`subtitle`（卖点副标题）、`original_price_cents`（划线价，0 不展示不结算）、`unit`（计价单位）、`detail`（文字详情，空行分段）、`detail_images`（详情长图 JSON 数组）、`service_notes`（服务承诺 + 售后例外）、`sort`（排序权重）
- `images` 字段补齐使用链路：后台上传入口 → 用户端轮播。AutoMigrate 幂等加列，SQLite 不强制 varchar 长度，长度校验在 `productReq.normalize()` 兜底
- 演示商品整条重写（材质/工艺/尺码/注意事项四段式文案 + 划线价 + 单位 + 服务承诺），4 张商品封面与 4 张特写图由程序生成（`repo/product_assets.go`，几何占位，真实图后台传）；**老库启动时按标题回填空字段**，不覆盖后台已改过的值
- 后台商品弹窗从 3 个输入扩成完整表单：多图上传（首张封面、可排序删除）、详情长图上传、划线价、单位/排序、文字详情、服务承诺；列表加封面缩略图与副标题。上下架接口是全量覆盖语义，前端改走 `productInput()` 带全字段，避免把新字段清空
- 用户端商品列表/后台列表改按 `sort desc` 排序；小程序商品页重做：多图轮播（点图预览大图）、副标题、划线价（高于最低售价才显示）、单位（共 N 件）、服务承诺条、图文详情（分段 + 长图）、底部固定提示"定制商品不支持七天无理由退货"（消法 25 条定作品除外，必须下单前告知）；首页列表显示副标题与划线价
- 纪律说明：划线价放 SPU 不放 SKU（当前每款同价，够用；下沉 SKU 已列入 P1）

## v0.8.1 (2026-11-06) — 订单详情补全信息

- 后端 `GET /user/orders/:id` 响应扩展（不新增接口，兼容旧调用）：
  - `items[]` 增加 `image_url`（商品首图，取 `product.images[0]`）与 `work_image_url`（绑定作品图，thumbnail 优先）
  - 新增 `events`：`order_events` 流水按发生顺序返回，可直接渲染时间轴（含 operator `user:1`/`admin:1` 与 note）
  - 新增 `amounts`：`items_cents / shipping_cents / discount_cents / total_cents`，运费与优惠由合计差额反推（订单表暂无独立字段，当前均为 0）
- 归属校验从 api 层移到 `OrderService.GetForUser`（他人订单 403、不存在 404），api 只做参数与响应，回到"业务规则只在 service"的纪律
- 小程序订单详情页重做（原来只有状态、单号、下单时间、明细文字、收货、物流六项）：
  - 状态头：状态大字 + 一句话说明（等待商家接单/正在打包/已发出/已完成/已取消）+ 取消原因 + 退款标记
  - 订单信息：单号（可复制）、下单时间、支付方式、支付时间
  - 订单进度：时间轴（倒序、最新一条高亮），含商家接单/发货单号等 note
  - 商品明细：商品图（缺图时用作品图兜底，可点开放大）、规格、单价、数量、小计、定制作品图
  - 金额明细：商品金额 / 运费（免运费）/ 优惠 / 实付（大字）
  - 收货信息：姓名、电话（可复制）、地址（可复制）
  - 物流信息：快递公司、运单号（可复制）、发货时间、无轨迹说明
  - 底部操作栏：待接单可取消、已发货可确认收货；支持下拉刷新
- e2e mp 增加详情断言：金额明细对账、流水 4 条（下单→接单→发货→收货）、作品图非空

## v0.8 (2026-11-06) — API 命名空间按端拆分

- 用户端（小程序）接口统一收进 `/api/v1/user/*`，含公开的 `/user/auth/login`、`/user/auth/dev_login`、`/user/products`（商品浏览不需要登录，但仍属用户端）；管理后台保持 `/api/v1/admin/*`。`/api/v1` 根下不再挂业务接口，静态资源 `/static` 不变
- 路由按 group 组织：`v1.Group("/user")` 放公开接口，其下再挂 `AuthRequired + RoleRequired("user")` 的需登录 group；后台仅 `admin/auth/login` 公开
- 小程序：`config.js` 新增 `API` 常量（`BASE + '/api/v1/user'`），`request.js` / 上传统一用它拼 URL，8 个页面里的 `/orders`、`/addresses` 等资源路径零改动；顺手删掉 login.js 未使用的 `BASE` 引用
- web 后台无需改动（`api/modules/*` 本就全是 `/admin/...`）
- 三套 e2e 脚本路径同步为 `/user/...`；`go vet` + `go test` 全过，m0/m1/mp 全链路通过
- 实测（:8080 新版二进制）：`/api/v1/user/products` 200、`/api/v1/admin/auth/login` 200、`/api/v1/products` 404、`/ui/` 200
- 破坏性变更：旧的无端前缀路径立即失效（项目未上线，不做兼容层）；新增接口一律归入 user/admin 其一（CONVENTIONS 接口约定 20-21）

## v0.7 (2026-11-06) — 前端工程规范化重构

**web/（按社区标准补齐骨架）**
- 补 ESLint9（扁平配置 + eslint-plugin-vue + typescript-eslint）与 Prettier；`npm run lint / format`
- 新增 `.env` / `.env.example`（VITE_API_BASE），API 基址不再硬编码
- 目录分层：`layouts/`（Layout 移出 views）、`api/http.ts + api/modules/*`（按域分模块）、`router/`（守卫 + document.title）、`stores/`（Pinia auth，组件不直接碰 localStorage）、`components/`（StatusTag/AmountText）、`types/`（集中 DTO）、`styles/`（主题变量）、`utils/`（auth/money）
- Element Plus 改按需引入（unplugin-auto-import + unplugin-vue-components + ElConfigProvider 中文 locale），dist 从全量降至 754K，视图按域拆 chunk
- 视图全部改用模块化 API + 复用组件；路由守卫增强（已登录访问 /login 回 orders）
- 401 统一处理：http 层清 token 回登录页

**miniprogram/（规范补齐）**
- 新增 `components/`：empty-state（空态/加载/行动按钮）、status-tag（状态文案颜色，externalClasses 定制）；四页接入并移除页面内重复实现
- tabBar 补 4 组图标（81×81 PNG，Go 程序生成、超采样抗锯齿；临时工具已删）
- 页面重命名：pages/orders→order-list、pages/order→order-detail（目录与文件同步，杜绝单复数混淆）
- project.private.config.json 移出版本库（.gitignore）

**验收**：vue-tsc + vite build + eslint 全过；小程序 node --check + JSON 全过；m0/m1/mp 三套 e2e 全过（金额/快照对账一致）；Edge 实测登录→四页截图，按需引入样式无异常。

**文档：目录规范入档**
- `docs/CONVENTIONS.md` 原「目录结构」重写为规范章节：仓库总览（根目录只允许 AGENTS.md + docs/ server/ web/ miniprogram/）→ server/ web/ miniprogram/ 三端完整目录树（每层注明职责与禁止事项）→ **新增文件落位决策表**（新接口/业务规则/表字段/配置项/管理页/跨页组件/小程序页/脚本/文档 → 落点 + 需同步的文档）→ 目录纪律 15-19
- 纪律编号重排为 硬性 1-9 → 前端 10-14 → 目录 15-19，去掉「提交」重复条；`handlers_<域>.go` 定为新增接口的命名方式（handlers_m1/m2 为历史切法，不再沿用）
- `AGENTS.md` 增一条硬规则：新增文件必须按落位表放，找不到落点先问，不在根目录或 `web/src/` 平铺
- `docs/AI协作开发流程.md`：文档体系表标注 CONVENTIONS 含目录规范与落位表，并补一段"为什么目录规范也要写进 CONVENTIONS"（v0.7 前根目录出现 login.yaml/orders.yaml 的根因是没有落点规则，不是马虎）

## v0.6 (2026-11-06) — 收货地址管理
- 后端：addresses 表 + AddressService（上限 20 条、默认地址唯一、第一个自动默认、删默认自动补位最新一条、手机号 `^1\d{10}$` 校验、归属校验）；用户端 CRUD + 设默认 5 个接口
- 下单改造：POST /orders 必传 address_id，收货信息**快照**进订单（receiver_name/phone/address），地址后续修改不影响历史订单；无地址/他人地址/非法地址下单分别 400/403/400
- 小程序：地址列表页（默认标记/设默认/编辑/删除、选地址模式）、地址编辑页（省市区 picker、表单校验）；商品页顶部地址卡（默认预选、点击换地址、无地址引导）；订单列表/详情展示收货人信息；订单页「收货地址」入口
- 回归：go vet/test 全过（新增 7 组地址测试：首地址默认、默认唯一、越权 403、删默认补位、上限 20、非法入参、下单快照+地址修改不影响快照）；m0/m1/mp 三套 e2e 更新为带地址下单，全部通过
- 本地服务已重启（:8080，XMF_WEB 托管 /ui/）

## v0.5 (2026-11-06) — M1.5 小程序端（原生）
- miniprogram/ 原生工程：8 页面（登录/商店首页/商品详情/创作工坊/AI 生成/作品库/订单列表/订单详情），4 tab（商店/创作/作品/订单）
- utils/request.js：Promise 封装 wx.request，JWT 自动携带、401 清 token 跳登录、业务码统一处理；config.js 集中管理 BASE 地址与图片地址拼接
- 登录：wx.login→/auth/login 优先（配 appid 后即生效），失败自动露出开发者模拟登录入口（本地开发）
- 创作工坊：canvas 2d 渲染底板+元素图层，触摸拖动改相对坐标（0~1）、滑块缩放（0.5~2）、虚线框选中、12 个上限，生成时 POST /collage/compose 由服务端合成（与前端预览同一套参数，所见即所得）
- AI 生成：文生图（textarea）+ 照片转风格（chooseMedia→/upload→/ai/tasks），配额超限 429 直接展示后端文案
- 下单：SKU 单选（含库存置灰）、数量、可选绑定作品，模拟支付；订单：状态筛选、取消（仅待接单）、确认收货、物流展示
- 校验：JS 语法/JSON 全过；mp_e2e.py 模拟小程序完整调用链（登录→商品→素材→合成→作品→下单绑作品→接单发货收货）全通过，金额对账一致
- WorkBuddy 应用预览卡已创建（wbapp_ropiy3qTft1ZvSw04r1E6t，未发布）

## v0.4 (2026-11-06) — M1.4 Web 管理后台
- 后端新增：商品/分类/SKU 管理（删除保护：分类下有商品、SKU 有订单引用均拒绝）、系统配置白名单读写（ai_daily_free_limit 0~1000）、boards/elements 列表接口、后台上传
- **修复真实 bug**：GORM 把 OrderItem.SKUID 列名转成 `skuid` 而非 `sku_id`，导致"有订单引用的 SKU 禁止删除"检查静默失效——单测抓到，加 `gorm:"column:sku_id"` 显式列名修复（旧开发库需删除重建）
- web/ 前端：Vite + Vue3 + TS + Element Plus（hash 路由），页面：登录 / 订单管理（筛选/接单/发货弹窗/取消/退款标记）/ 商品管理（分类+SPU+SKU，价格元分换算）/ 素材管理（底板+元素，上传+锚点编辑）/ 系统配置（AI 每日配额）
- Go 单二进制托管前端：XMF_WEB=../web/dist 时服务 /ui/（SPA 回落 index.html）
- 质量门禁：vue-tsc 严格类型 + vite build 通过；go vet/test 全过（新增 SKU 删除保护/配置白名单测试）
- **浏览器实测通过**（playwright + Edge）：登录→订单接单→发货弹窗→状态变已发货（库存 100→99 对账）→商品页（4 商品 6 SKU 正确）→素材页（底板+6 元素预览正常）→配置页，截图存 shots/

## v0.3 (2026-11-06) — M1 后端：拼装工坊 + AI 框架 + 作品库
- imageutil 包：合成（底板+元素按相对坐标/倍率叠加）、PNG 读写、圆/弧/三角/圆角矩形绘制
- 演示素材程序化生成：底板小姑娘 + 双马尾/波波头/尖尖帽/贝雷帽/睫毛/蝴蝶结（几何占位，后台上传可替换）
- 拼装合成接口 POST /collage/compose：服务端合成 PNG 存 data/uploads，/static 静态服务，作品自动入库（meta 记录放置参数可复刻）
- AI 框架：Provider 接口 + MockProvider（文生图=占位图、照片转风格=复制原图），换真厂商只改 Provider；每日配额读 configs.ai_daily_free_limit（默认3，后台可改），超限 429；生成完成自动建作品
- 上传 POST /upload（白名单扩展名+10MB 限制+路径越界校验）；作品库 GET/DELETE（软删，校验归属）
- 后台：boards/elements CRUD（POST/PUT/DELETE）
- 单测新增：AI 配额（用尽 429、配额0 拒绝、他人不受影响）、照片路径越界拒绝、合成位置与倍率正确性
- 全链路实测：素材→拼装合成（400x500 PNG 验证）→上传→AI照片/文生图→配额429→作品库→下单绑作品→发货 全部通过；合成图人工目检位置正确
- 坑：boards 表无 sort 列导致查询 500（已修）；素材生成路径与 image_url 不一致导致 404（已修）——改代码后必须重启服务才生效

## v0.2 (2026-11-06) — M0 最小闭环竖切（后端）
- Go 后端骨架：Gin + GORM + SQLite（glebarez 纯 Go 驱动，免 CGO），幂等迁移 + 演示数据（管理员 admin/admin123、3 分类 4 商品 6 SKU、系统配置）
- 认证：微信 code2session 登录（待配置 appid）、开发者模拟登录（XMF_DEV_LOGIN）、管理员 JWT 登录（bcrypt）
- 订单核心：mock 支付下单直达待接单；库存原子扣减（`UPDATE ... WHERE stock >= ?`）；状态机 pending_accept→pending_ship→shipped→done；用户取消仅限待接单、后台可取消到待发货，取消恢复库存；order_events 流水只追加
- 单元测试 6 组：金额对账、库存不足回滚、下架商品拒单、全生命周期、取消恢复库存、退款标记
- 全链路实测通过（临时端口 :18080 + 临时库）：登录→下单→接单→发货→确认收货，金额库存对账一致；权限(403)/鉴权(401)/库存不足(400)/非法流转(409) 均正确返回
- 坑：GBK 终端 curl 发中文物流名乱码是终端问题，Python UTF-8 脚本验证服务端正常；实测脚本保留在 server/scripts/m0_e2e.py 作回归
- 环境变量：XMF_ADDR / XMF_DB / XMF_JWT_SECRET（缺省随机生成存 data/.jwt_secret 0600）/ XMF_WX_APPID / XMF_WX_SECRET / XMF_DEV_LOGIN

## v0.1 (2026-11-06)
- 需求澄清完成，产出 PRD v0.2：小程序创作（拼装/AI文生图/照片转风格）→ 选品下单 → 工厂后台接单发货 → 确认收货
- 定名「现梦坊」；技术栈定为 Go（Gin+GORM）+ SQLite（后期迁 PostgreSQL）
- 冷启动地基：AGENTS.md、docs/CONVENTIONS.md、docs/DATA-MODEL.md、docs/TODO.md、本文件
- MVP 范围确定为最小闭环，支付先用模拟（用户暂无企业主体与商户号）
