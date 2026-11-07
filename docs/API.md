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
| GET | /user/cart | 购物车列表（实时 join 商品/SKU/作品；含勾选件数与金额汇总；失效项 `valid=false`） | ✅ 已实现 |
| POST | /user/cart | 加购 `{sku_id, work_id?, qty}`（同款同图累加数量；单件 ≤99、≤50 条；校验上架与库存 ≥1） | ✅ 已实现 |
| PUT | /user/cart/:id | 改数量 `{qty}` 或改勾选 `{checked}`（二选一或同传；归属校验 403） | ✅ 已实现 |
| PUT | /user/cart | 批量勾选 `{checked, item_ids?}`（不传 item_ids 即全选/全不选） | ✅ 已实现 |
| DELETE | /user/cart/:id | 删除单项 | ✅ 已实现 |
| DELETE | /user/cart | 清空（`?checked=1` 只清已勾选） | ✅ 已实现 |
| POST | /user/cart/checkout | 结算 `{address_id, item_ids}` → 一单多明细订单（复用下单校验与库存原子扣减），成功后移除对应购物车项 | ✅ 已实现 |
| GET | /user/addresses | 收货地址列表（默认优先） | ✅ 已实现 |
| POST | /user/addresses | 新增地址（第一个自动设默认；上限 20；手机号 ^1\d{10}$） | ✅ 已实现 |
| PUT | /user/addresses/:id | 编辑地址（可改默认） | ✅ 已实现 |
| DELETE | /user/addresses/:id | 删除地址（软删；删默认自动补位最新一条） | ✅ 已实现 |
| POST | /user/addresses/:id/default | 设为默认（唯一默认） | ✅ 已实现 |
| POST | /user/upload | 上传图片（multipart，jpg/png/webp ≤10MB），返回 /static/ 地址 | ✅ 已实现 |
| GET | /user/collage/boards | 底板列表（仅上架） | ✅ 已实现 |
| GET | /user/collage/elements?board_id=1&category=hair | 元素列表；**board_id 必填**，只返回该底板下的上架元素（缺 board_id 返回 400） | ✅ 已实现 |
| POST | /user/collage/compose | 拼装合成（board_id + placements[x,y,scale]），服务端合成 PNG 并存为作品；**元素必须属于该底板**，否则 400 | ✅ 已实现 |
| POST | /user/ai/tasks | 发起 AI 生成（ai_text/ai_photo）；**每日免费额度用完后自动扣 1 积分（1 积分 = 1 张），积分不足返回 429** | ✅ 已实现 |
| GET | /user/ai/quota | 今日免费额度与积分余额（`{free_limit, free_used, free_left, points, can_generate}`） | ✅ 已实现 |
| GET | /user/works | 我的作品库（含 `is_public/like_count/use_count/prompt`）；`?scope=liked` 返回我点赞过的公开作品（下单选图用） | ✅ 已实现 |
| GET | /user/works/:id | 作品详情（自己的，或已发布的他人作品） | ✅ 已实现 |
| DELETE | /user/works/:id | 删除作品（软删；同时从广场下架） | ✅ 已实现 |
| PUT | /user/works/:id/publish | 发布/取消发布 `{is_public: 1\|0}`；发布后进入广场且公开 AI 提示词 | ✅ 已实现 |
| GET | /user/plaza | **作品广场 / 排行榜**：`?sort=latest\|uses\|likes&page=1&size=20&keyword=`；只返回 `is_public=1 且 status=1`，含作者昵称头像、提示词、点赞/使用数、`liked`（我是否点过赞）。路径不用 `/works/public` 是因为与 `/works/:id` 通配段冲突 | ✅ 已实现 |
| POST | /user/works/:id/like | 点赞（幂等；自己的作品 400；返回最新 `like_count`） | ✅ 已实现 |
| DELETE | /user/works/:id/like | 取消点赞（幂等；返回最新 `like_count`） | ✅ 已实现 |
| GET | /user/points | 我的积分（最近 50 条）：`{points, logs:[{delta,balance,reason,remark,created_at}]}`；积分明细页传 `?page=1&size=20` 走分页，此时额外返回 `total`（size 上限 50） | ✅ 已实现（v2.0 加分页） |
| GET | /user/me | **「我的」页聚合**（v2.0）：用户资料 + 资产统计（作品/已发布/点赞/购物车）+ 各状态订单数，首屏一次取全（结构见下文） | ✅ 已实现 |
| PUT | /user/me | 改昵称/头像 `{nickname?, avatar?}`；空串表示不改（避免误清空），昵称 1~16 字否则 400；头像先走 `/upload` 拿地址再回填 | ✅ 已实现 |
| GET | /user/works/:id/view | **作品详情页**（v2.0）：`{work, author, liked, is_mine}`；自己的作品或已发布的他人作品，其余 404。与 `/works/:id` 并存，后者保留老的扁平结构 | ✅ 已实现 |

**下单/加购的作品规则（v1.7）**：`work_id` 可选，允许两种作品——**自己的**（status=1），或**我点赞过的公开作品**（is_public=1 且 status=1）。他人未公开/已下架/已删除的作品一律 400；加购时放行的作品，若之后被作者取消发布，结算时按「作品已失效」拦截（不影响已生成的订单）。

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

### 购物车响应约定（GET /user/cart）

```json
{
  "items": [{
    "id": 5, "sku_id": 1, "work_id": 3, "qty": 2, "checked": 1, "valid": true,
    "title": "定制图案T恤", "specs": "{\"尺码\":\"M\",\"颜色\":\"白\"}",
    "price_cents": 9900, "image_url": "/static/products/tee_main.png",
    "work_thumb": "/static/collage_xxx.png", "stock": 99,
    "invalid_reason": "", "created_at": "..."
  }],
  "summary": { "checked_count": 3, "checked_cents": 29700 }
}
```

- `valid=false` 时 `invalid_reason` 为人话（商品已下架 / 商品规格已失效 / 作品已删除），前端标灰禁用勾选
- `summary` 只统计勾选且有效项；价格是参考，结算以 SKU 现价为准
- checkout 错误：`404 无可结算商品`、`400 请选择收货地址`、`409 商品库存不足`（下单原子扣减失败）、`400 含失效商品，请先移除`

### 作品广场响应（GET /user/plaza）

```json
{
  "total": 37,
  "items": [{
    "id": 12, "type": "ai_text", "image_url": "/static/ai_xxx.png", "thumbnail_url": "/static/ai_xxx.png",
    "prompt": "扎双马尾的小姑娘，粉色背景，卡通风格",
    "like_count": 8, "use_count": 21, "liked": true, "published_at": "2026-11-07T10:20:00+08:00",
    "author": { "id": 3, "nickname": "小梦", "avatar": "/static/avatar_3.png" },
    "is_mine": false
  }]
}
```

- `sort`：`latest` 最新榜（published_at desc）/ `uses` 使用量榜（use_count desc，同分比 published_at）/ `likes` 点赞榜（like_count desc，同分比 published_at）
- 只返回 `is_public=1 AND status=1`；`prompt` 仅 AI 作品有值（照片转风格为空），拼装作品为空
- `liked` / `is_mine` 依赖登录态；未登录（401）时前端走登录页
- `keyword` 对提示词做模糊匹配（AI 作品可被搜到）

### 积分响应（GET /user/points）

```json
{ "points": 12, "logs": [{ "delta": 2, "balance": 12, "reason": "work_used", "remark": "作品《小熊猫》被使用 2 件", "created_at": "..." }] }
```

- 不带 `page/size` 时取最近 50 条、不返回 `total`（兼容 v1.7 老调用）
- 带 `page/size` 时分页，响应额外带 `total`；`size` 缺省 20、上限 50

### 「我的」页响应（GET /user/me）

```json
{
  "user": {
    "id": 3, "nickname": "阿岚", "avatar": "/static/avatar_3.png", "status": 1, "points": 12,
    "created_at": "...", "updated_at": "..."
  },
  "stats": {
    "works": 8, "public_works": 5, "liked": 36, "cart_count": 2,
    "orders": { "pending_accept": 0, "producing": 1, "packing": 0, "pending_ship": 1, "shipped": 2, "done": 7 }
  }
}
```

- `stats.works` 只算 `status=1`（软删不计）；`public_works` 在其基础上要求 `is_public=1`
- `stats.orders` 由一次 `GROUP BY status` 得出，`cancelled` 不进统计（「我的」入口只展示在途与已完成）
- 小程序把 `producing + packing + pending_ship` 拍平展示为「进行中」角标——对用户而言这三步都在商家手里
- 该接口替代了 v1.x 里「我的」页分头请求 user / works / cart / orders 的做法，首屏只发一个请求

### 作品详情响应（GET /user/works/:id/view）

```json
{
  "work": {
    "id": 12, "user_id": 3, "type": "ai_text", "image_url": "/static/ai_xxx.png",
    "thumbnail_url": "/static/ai_xxx.png", "prompt": "扎双马尾的小姑娘，粉色背景，卡通风格",
    "is_public": 1, "status": 1, "like_count": 8, "use_count": 21, "created_at": "..."
  },
  "author": { "id": 3, "nickname": "阿岚", "avatar": "/static/avatar_3.png" },
  "liked": false,
  "is_mine": true
}
```

- 未发布/已下架/已删除/他人私有 → 404（不区分原因，避免枚举他人作品）
- `liked` 对自己的作品恒为 `false`（自己不能给自己点赞）
- 与 `GET /user/works/:id` 的差异：多带 `author` / `liked` / `is_mine`，少了归属校验的 403（不可见一律 404）

## 管理后台 — 前缀 `/api/v1/admin`

| 方法 | 路径 | 说明 | 状态 |
|---|---|---|---|
| POST | /admin/auth/login | 管理员登录 | ✅ 已实现 |
| GET | /admin/orders | 订单列表（筛选状态；order_no 模糊 / user_id 过滤；行含买家昵称头像） | ✅ 已实现 |
| GET | /admin/orders/:id | 订单详情（明细含商品图/作品图、物流、状态流水、金额明细、买家信息） | ✅ 已实现 |
| POST | /admin/orders/:id/accept | 接单（待接单 → 生产中） | ✅ 已实现 |
| POST | /admin/orders/:id/pack | 生产完成（生产中 → 打包中） | ✅ 已实现 |
| POST | /admin/orders/:id/ready_ship | 打包完成（打包中 → 待发货） | ✅ 已实现 |
| POST | /admin/orders/:id/ship | 发货（物流公司+单号） | ✅ 已实现 |
| POST | /admin/orders/:id/cancel | 后台取消（待接单/生产中/打包中/待发货，恢复库存） | ✅ 已实现 |
| POST | /admin/orders/:id/refund_mark | 退款标记（仅标记，已发货/已完成） | ✅ 已实现 |
| GET/POST/PUT/DELETE | /admin/products… | 商品管理（列表含下架，删除=下架）。v0.9 起支持 subtitle / images / original_price_cents / unit / detail / detail_images / service_notes / sort；images 与 detail_images 为 JSON 数组字符串；update 为全量覆盖语义，前端上下架时必须带全字段 | ✅ 已实现 |
| POST/PUT/DELETE | /admin/skus… | SKU 管理（有订单引用的 SKU 禁止删除） | ✅ 已实现 |
| GET/POST/PUT/DELETE | /admin/categories… | 商品分类管理（分类下有商品禁止删除） | ✅ 已实现 |
| GET | /admin/boards | 底板列表（含 `element_count` 启用元素数，供底板管理页展示） | ✅ 已实现 |
| POST/PUT | /admin/boards… | 新增/编辑底板（name/url/width/height） | ✅ 已实现 |
| DELETE | /admin/boards/:id | 下架底板（软删）；**该底板仍有启用元素时返回 409**，需先下架元素 | ✅ 已实现 |
| GET | /admin/elements?board_id=&category= | 元素列表；**board_id 必填**（后台元素页按底板展示），category 可选 | ✅ 已实现 |
| POST/PUT | /admin/elements… | 新增/编辑元素；**board_id 必填**（底板不存在返回 400），可改归属底板；anchor 为空时用默认 `{"x":0.5,"y":0.5,"scale":1}` | ✅ 已实现 |
| DELETE | /admin/elements/:id | 下架元素（软删） | ✅ 已实现 |
| GET/PUT | /admin/configs | 系统配置（白名单键，ai_daily_free_limit 0~1000） | ✅ 已实现 |
| GET | /admin/users | 用户列表（含订单数/作品数统计、积分余额；keyword 按昵称模糊、status=0/1 过滤） | ✅ 已实现 |
| PUT | /admin/users/:id/status | 启用/禁用用户（{status:0\|1}）。禁用后无法登录；已签发 JWT 在过期前仍有效 | ✅ 已实现 |
| GET | /admin/works | 全站作品列表（含作者昵称、公开状态、提示词、点赞数、使用次数；type=collage/ai_text/ai_photo、status=0/1/2、public=0/1 过滤） | ✅ 已实现 |
| PUT | /admin/works/:id/status | 上架(1)/下架(2)。用户已删除(status=0)的作品不可变更，返回 409 | ✅ 已实现 |
| POST | /admin/upload | 后台上传图片（与用户端 /upload 同逻辑） | ✅ 已实现 |
| GET | /admin/me | 登录态自检：账号、角色、权限码列表（超管为 `["*"]`） | ✅ 已实现 |
| GET | /admin/permissions | 权限清单（按模块分组的权限码 + 中文名，前端渲染勾选） | ✅ 已实现 |
| GET/POST/PUT/DELETE | /admin/roles… | 角色管理：列表含账号数；内置角色不可删、超管角色权限不可改；在用角色不可删 | ✅ 已实现 |
| GET/POST/PUT/DELETE | /admin/admins… | 管理员账号：新增（用户名唯一/密码≥6位）、改角色/昵称/状态、重置密码、删除 | ✅ 已实现 |
| GET | /admin/audit-logs | 审计日志分页查询：`admin_id` / `action`（前缀）/ `result`(ok\|fail\|deny) / `keyword` / `page` / `size` | ✅ 已实现 |

### 权限码（v1.4）

| 模块 | 权限码 |
|---|---|
| 订单 | `order:view` `order:process`（接单/生产完成/打包完成） `order:ship` `order:cancel` `order:refund` |
| 商品 | `product:view` `product:edit`（含分类/SKU/上传） |
| 素材 | `asset:view` `asset:edit` |
| 用户 | `user:view` `user:edit` |
| 作品 | `work:view` `work:edit` |
| 配置 | `config:view` `config:edit` |
| 权限与日志 | `admin:view` `admin:edit` `role:view` `role:edit` `audit:view` |

- 超级管理员（`super` 角色，perms=`["*"]`）绕过全部权限校验
- 鉴权链路：`AuthRequired`（JWT）→ `RoleRequired("admin")` → `AdminAuthRequired`（每请求查库，禁用/改权限立即生效）→ `PermRequired`（权限码）→ `AuditWrite`（写操作留痕）
- 越权返回 403 并写一条 `action=deny` 的审计日志（target_id 为所需权限码）

## 状态码约定

- 400 参数错 / 401 未登录 / 403 无权限 / 404 不存在 / 409 状态机非法流转 / 429 配额用尽 / 500 系统错误
