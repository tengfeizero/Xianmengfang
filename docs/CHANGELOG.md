# CHANGELOG — 现梦坊

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
