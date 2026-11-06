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
| POST | /orders | 下单（address_id 必传、items:[{sku_id、work_id 可空、qty}]；mock 支付直达待接单；收货信息快照进订单） | ✅ 已实现 |
| GET | /orders | 我的订单（?status= 筛选） | ✅ 已实现 |
| GET | /orders/:id | 订单详情（含明细+物流+收货快照） | ✅ 已实现 |
| POST | /orders/:id/confirm | 确认收货 | ✅ 已实现 |
| POST | /orders/:id/cancel | 取消订单（仅待接单，恢复库存） | ✅ 已实现 |
| GET | /addresses | 收货地址列表（默认优先） | ✅ 已实现 |
| POST | /addresses | 新增地址（第一个自动设默认；上限 20；手机号 ^1\d{10}$） | ✅ 已实现 |
| PUT | /addresses/:id | 编辑地址（可改默认） | ✅ 已实现 |
| DELETE | /addresses/:id | 删除地址（软删；删默认自动补位最新一条） | ✅ 已实现 |
| POST | /addresses/:id/default | 设为默认（唯一默认） | ✅ 已实现 |
| POST | /upload | 上传图片（multipart，jpg/png/webp ≤10MB），返回 /static/ 地址 | ✅ 已实现 |
| GET | /collage/boards | 底板列表 | ✅ 已实现 |
| GET | /collage/elements?category= | 元素列表 | ✅ 已实现 |
| POST | /collage/compose | 拼装合成（board_id + placements[x,y,scale]），服务端合成 PNG 并存为作品 | ✅ 已实现 |
| POST | /ai/tasks | 发起 AI 生成（ai_text/ai_photo，每日配额超限返回 429；当前 mock 供应商） | ✅ 已实现 |
| GET | /works | 我的作品库 | ✅ 已实现 |
| GET | /works/:id | 作品详情 | ✅ 已实现 |
| DELETE | /works/:id | 删除作品（软删） | ✅ 已实现 |

## 管理后台

| 方法 | 路径 | 说明 | 状态 |
|---|---|---|---|
| POST | /admin/auth/login | 管理员登录 | ✅ 已实现 |
| GET | /admin/orders | 订单列表（筛选状态） | ✅ 已实现 |
| POST | /admin/orders/:id/accept | 接单 | ✅ 已实现 |
| POST | /admin/orders/:id/ship | 发货（物流公司+单号） | ✅ 已实现 |
| POST | /admin/orders/:id/cancel | 后台取消（待接单/待发货，恢复库存） | ✅ 已实现 |
| POST | /admin/orders/:id/refund_mark | 退款标记（仅标记，已发货/已完成） | ✅ 已实现 |
| GET/POST/PUT/DELETE | /admin/products… | 商品管理（列表含下架，删除=下架） | ✅ 已实现 |
| POST/PUT/DELETE | /admin/skus… | SKU 管理（有订单引用的 SKU 禁止删除） | ✅ 已实现 |
| GET/POST/PUT/DELETE | /admin/categories… | 商品分类管理（分类下有商品禁止删除） | ✅ 已实现 |
| GET/POST/PUT/DELETE | /admin/elements… /boards… | 素材与底板管理（POST/GET/PUT/DELETE） | ✅ 已实现 |
| GET/PUT | /admin/configs | 系统配置（白名单键，ai_daily_free_limit 0~1000） | ✅ 已实现 |
| POST | /admin/upload | 后台上传图片（与用户端 /upload 同逻辑） | ✅ 已实现 |

## 状态码约定

- 400 参数错 / 401 未登录 / 403 无权限 / 404 不存在 / 409 状态机非法流转 / 429 配额用尽 / 500 系统错误
