# TODO — 现梦坊

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

## Backlog
- [ ] 真实微信支付（等主体+商户号资质）
- [ ] 多租户改造
- [ ] 退款售后流、物流跟踪、评价
- [ ] PostgreSQL 迁移
- [ ] 服务器采购 + 域名 ICP 备案 + HTTPS（用户负责，尽早启动）
