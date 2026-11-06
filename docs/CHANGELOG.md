# CHANGELOG — 现梦坊

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
