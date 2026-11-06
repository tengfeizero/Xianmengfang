# API — 现梦坊

> AI 维护，接口变动必须同步本文件。状态：v0.1 骨架（随 M0 实现逐条补充参数与错误码）。

## 通用

- Base：`https://<域名>/api/v1`
- 鉴权：`Authorization: Bearer <JWT>`
- 响应：`{ "code": 0, "msg": "ok", "data": ... }`；非 0 为业务错误，msg 面向用户可读
- 金额单位分

## 小程序端

| 方法 | 路径 | 说明 | 状态 |
|---|---|---|---|
| POST | /auth/login | wx code 换 JWT（需配置 XMF_WX_APPID/SECRET） | ✅ 已实现 |
| POST | /auth/dev_login | 开发者模拟登录（仅 XMF_DEV_LOGIN=1，默认开） | ✅ 已实现 |
| GET | /products | 商品列表（含 SKU） | ✅ 已实现 |
| GET | /products/:id | 商品详情 | ✅ 已实现 |
| POST | /orders | 下单（sku_id、work_id 可空、qty；mock 支付直达待接单） | ✅ 已实现 |
| GET | /orders | 我的订单（?status= 筛选） | ✅ 已实现 |
| GET | /orders/:id | 订单详情（含明细+物流） | ✅ 已实现 |
| POST | /orders/:id/confirm | 确认收货 | ✅ 已实现 |
| POST | /orders/:id/cancel | 取消订单（仅待接单，恢复库存） | ✅ 已实现 |
| GET | /works | 我的作品库 | M1 |
| POST | /works | 保存作品（拼装 meta 或 AI 结果） | M1 |
| POST | /ai/tasks | 发起 AI 生成（文生图/照片转风格，受每日配额限制） | M1 |
| GET | /collage/boards | 底板列表 | M1 |
| GET | /collage/elements?category= | 元素列表 | M1 |

## 管理后台

| 方法 | 路径 | 说明 | 状态 |
|---|---|---|---|
| POST | /admin/auth/login | 管理员登录 | ✅ 已实现 |
| GET | /admin/orders | 订单列表（筛选状态） | ✅ 已实现 |
| POST | /admin/orders/:id/accept | 接单 | ✅ 已实现 |
| POST | /admin/orders/:id/ship | 发货（物流公司+单号） | ✅ 已实现 |
| POST | /admin/orders/:id/cancel | 后台取消（待接单/待发货，恢复库存） | ✅ 已实现 |
| POST | /admin/orders/:id/refund_mark | 退款标记（仅标记，已发货/已完成） | ✅ 已实现 |
| GET/POST/PUT | /admin/products… | 商品与 SKU 管理 | M1 |
| GET/POST/PUT | /admin/elements… /boards… | 素材与底板管理 | M1 |
| GET/PUT | /admin/configs | 系统配置（AI 配额等） | M1 |

## 状态码约定

- 400 参数错 / 401 未登录 / 403 无权限 / 404 不存在 / 409 状态机非法流转 / 429 配额用尽 / 500 系统错误
