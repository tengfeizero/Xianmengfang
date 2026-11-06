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

## 下一步候选（M1 起）
- [ ] 拼装工坊（底板 + 槽位元素 + 微调 + 合成图）
- [ ] AI 文生图 / 照片转风格（低价 API，配额控制）
- [ ] 作品库管理
- [ ] Web 后台商品管理（分类/SPU/SKU/库存）
- [ ] Web 后台素材管理
- [ ] 小程序端各页面

## Backlog
- [ ] 真实微信支付（等主体+商户号资质）
- [ ] 多租户改造
- [ ] 退款售后流、物流跟踪、评价
- [ ] PostgreSQL 迁移
- [ ] 服务器采购 + 域名 ICP 备案 + HTTPS（用户负责，尽早启动）
