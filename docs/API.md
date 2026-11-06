# API — 现梦坊

> AI 维护，接口变动必须同步本文件。状态：v0.1 骨架（随 M0 实现逐条补充参数与错误码）。

## 通用

- Base：`https://<域名>/api/v1`
- **命名空间**：用户端（小程序）全部在 `/api/v1/user/*`，管理后台全部在 `/api/v1/admin/*`；v1 根下不再挂业务接口（只有 `/static` 静态资源）。新增接口必须归入其一，禁止在 v1 根上新增
- 鉴权：`Authorization: Bearer <JWT>`
- 响应：`{ "code": 0, "msg": "ok", "data": ... }`；非 0 为业务错误，msg 面向用户可读
- 金额单位分

## 用户端（小程序）— 前缀 `/api/v1/user`

> 公开接口（登录、商品浏览）也在 `/user` 下，保证"一个端一个前缀"；小程序侧前缀集中在 `miniprogram/config.js` 的 `API` 常量，页面内只写 `/orders` 这类资源路径。

| 方法 | 路径 | 说明 | 状态 |
|---|---|---|---|
| POST | /user/auth/login | wx code 换 JWT（需配置 XMF_WX_APPID/SECRET） | ✅ 已实现 |
| POST | /user/auth/dev_login | 开发者模拟登录（仅 XMF_DEV_LOGIN=1，默认开） | ✅ 已实现 |
| GET | /user/products | 商品列表（含 SKU） | ✅ 已实现 |
| GET | /user/products/:id | 商品详情 | ✅ 已实现 |
| POST | /user/orders | 下单（address_id 必传、items:[{sku_id、work_id 可空、qty}]；mock 支付直达待接单；收货信息快照进订单） | ✅ 已实现 |
| GET | /user/orders | 我的订单（?status= 筛选） | ✅ 已实现 |
| GET | /user/orders/:id | 订单详情（明细含商品图/作品图 + 物流 + 状态流水 + 金额明细，结构见下文） | ✅ 已实现 |
| POST | /user/orders/:id/confirm | 确认收货 | ✅ 已实现 |
| POST | /user/orders/:id/cancel | 取消订单（仅待接单，恢复库存） | ✅ 已实现 |
| GET | /user/addresses | 收货地址列表（默认优先） | ✅ 已实现 |
| POST | /user/addresses | 新增地址（第一个自动设默认；上限 20；手机号 ^1\d{10}$） | ✅ 已实现 |
| PUT | /user/addresses/:id | 编辑地址（可改默认） | ✅ 已实现 |
| DELETE | /user/addresses/:id | 删除地址（软删；删默认自动补位最新一条） | ✅ 已实现 |
| POST | /user/addresses/:id/default | 设为默认（唯一默认） | ✅ 已实现 |
| POST | /user/upload | 上传图片（multipart，jpg/png/webp ≤10MB），返回 /static/ 地址 | ✅ 已实现 |
| GET | /user/collage/boards | 底板列表 | ✅ 已实现 |
| GET | /user/collage/elements?category= | 元素列表 | ✅ 已实现 |
| POST | /user/collage/compose | 拼装合成（board_id + placements[x,y,scale]），服务端合成 PNG 并存为作品 | ✅ 已实现 |
| POST | /user/ai/tasks | 发起 AI 生成（ai_text/ai_photo，每日配额超限返回 429；当前 mock 供应商） | ✅ 已实现 |
| GET | /user/works | 我的作品库 | ✅ 已实现 |
| GET | /user/works/:id | 作品详情 | ✅ 已实现 |
| DELETE | /user/works/:id | 删除作品（软删） | ✅ 已实现 |

### 订单详情响应（GET /user/orders/:id）

```json
{
  "order": {
    "id": 1, "order_no": "XMF20261106131900", "status": "shipped", "pay_type": "mock",
    "total_cents": 19800, "refund_flag": 0, "cancel_reason": "",
    "receiver_name": "张三", "receiver_phone": "13800001111",
    "receiver_address": "广东省深圳市南山区科技园 1 号楼 502",
    "paid_at": "2026-11-06T13:19:00+08:00", "shipped_at": "2026-11-06T13:19:00+08:00", "done_at": null,
    "created_at": "2026-11-06T13:19:00+08:00"
  },
  "items": [{
    "id": 1, "sku_id": 1, "work_id": 1, "title": "定制图案T恤",
    "specs": "{\"尺码\":\"M\",\"颜色\":\"白\"}", "price_cents": 9900, "qty": 2,
    "image_url": "", "work_image_url": "/static/collage_20261106131900_2421.png"
  }],
  "shipment": { "company": "顺丰速运", "tracking_no": "SF123456789", "created_at": "2026-11-06T13:19:00+08:00" },
  "events": [
    { "from": "", "to": "pending_accept", "operator": "user:1", "note": "下单，模拟支付成功", "created_at": "..." },
    { "from": "pending_accept", "to": "pending_ship", "operator": "admin:1", "note": "接单", "created_at": "..." },
    { "from": "pending_ship", "to": "shipped", "operator": "admin:1", "note": "发货：顺丰速运 SF123456789", "created_at": "..." }
  ],
  "amounts": { "items_cents": 19800, "shipping_cents": 0, "discount_cents": 0, "total_cents": 19800 }
}
```

- `items[].image_url`：商品首图（`product.images` 的第一张，可能为空）；`work_image_url`：绑定作品的图（无绑定为 null）。两者都不进订单快照，取的是当前商品/作品图——商品换图会改变历史订单展示（MVP 接受）
- `events`：`order_events` 流水，按发生顺序返回，可直接渲染时间轴；`operator` 形如 `user:12` / `admin:1`
- `amounts`：订单表暂无运费/优惠独立字段，`shipping_cents` 与 `discount_cents` 由 `total_cents - items_cents` 反推（当前均为 0）
- 他人订单 403、不存在 404（归属校验在 service 层）

## 管理后台 — 前缀 `/api/v1/admin`

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
