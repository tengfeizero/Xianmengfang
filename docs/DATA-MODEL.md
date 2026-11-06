# DATA-MODEL — 现梦坊

> AI 起草，人审核金额/库存/状态机口径。状态：v0.1 草稿待审。

## 通用约定

- 所有表含 `id INTEGER PRIMARY KEY`（SQLite 自增）、`created_at`、`updated_at`
- 金额单位分（int），字段后缀 `_cents`
- 软删除仅用于 `works`、`products`（`status` 字段控制），订单不删除

## 表

### users（小程序用户）
| 字段 | 类型 | 说明 |
|---|---|---|
| openid | text unique | 微信 openid |
| nickname / avatar | text | 可后补 |
| status | int | 1 正常 / 0 禁用 |

### addresses（收货地址，M1.5）
| 字段 | 说明 |
|---|---|
| user_id | 归属（上限 20 条/人） |
| name / phone | 收货人（phone 校验 `^1\d{10}$`） |
| province / city / district / detail | 地区与详址 |
| is_default | 1 默认（每用户唯一；第一个自动默认；删除默认后最新一条自动补位） |
| status | 1 正常 / 0 已删除（软删） |
- 下单时快照 receiver_name / receiver_phone / receiver_address 到订单，地址后续修改不影响历史订单。

### works（作品）
| 字段 | 类型 | 说明 |
|---|---|---|
| user_id | int | 归属 |
| type | text | `collage` 拼装 / `ai_text` 文生图 / `ai_photo` 照片转风格 |
| image_url / thumbnail_url | text | 成品图与缩略图 |
| meta | json | 拼装记录（元素 id、坐标缩放）或生成参数，便于复刻 |
| status | int | 1 正常（可用于商品）/ 0 已删除 / 2 被下架 |

### elements（拼装元素）· boards（底板）
| 字段 | 说明 |
|---|---|
| category | 元素分类：hat/lash/hair/face/expression/outfit/accessory 等 |
| name / image_url / sort / status | PNG 素材、排序、上下架 |
| anchor | json，槽位锚点（x/y/scale），拼装时按锚点定位 |

### ai_tasks（AI 生成任务）
| 字段 | 说明 |
|---|---|
| user_id / type / params | 类型 ai_text / ai_photo；prompt 或原图引用 |
| status | pending / done / failed |
| provider / cost_points | 供应商与费用估算（分） |
| result_url / error | 产物或错误信息 |
| quota_date | 用于每日配额统计（每用户每日免费 N 次，N 后台可配，默认 3） |

### product_categories（商品分类）· products（商品 SPU）· skus
- categories：name / sort / status
- products（v0.9 起）：category_id / title / **subtitle 副标题** / description（兼容一句话简介）/
  **images JSON 数组（主图，首张封面）** / **original_price_cents 划线价（0 不展示，不参与结算）** /
  **unit 计价单位** / **detail 文字详情（空行分段）** / **detail_images JSON 数组（详情长图）** /
  **service_notes 服务承诺 + 售后例外** / **sort 排序权重** / status（上下架）
- skus：product_id / specs json（如 `{"尺码":"M","颜色":"白"}`）/ **price_cents int** / **stock int** / code
- 字段取舍与 P1/P2 规划（attrs、skus.image_url、tags、起订量、运费模板等）见 PRD 3.2.1

### orders（订单）· order_items（明细）· shipments（发货）
| 字段 | 说明 |
|---|---|
| order_no | 唯一业务单号，`XMF + yyyyMMdd + 序列` |
| status | 状态机见下 |
| pay_type | `mock` / `wxpay`（预留） |
| total_cents | int 分 = Σ(明细 price_cents × qty) |
| order_items | sku 快照（名称/规格/单价）+ **work_id（绑定的作品）** + qty |
| shipments | 物流公司 / 单号 / 发货时间（一单可多条，MVP 一单一条） |

## 订单状态机（唯一合法流转）

```
pending_pay ──支付──→ pending_accept ──接单──→ pending_ship ──发货──→ shipped ──确认收货──→ done
     │                    │
     └────── cancel ──────┴──（用户取消 / 后台取消，均记 cancel_reason）
```

- mock 支付：下单直接进入 `pending_accept`（跳过 pending_pay），`pay_type=mock`
- `done` 之前允许后台标记退款标记（仅标记字段，MVP 无退款资金流）
- 状态变更写 `order_events` 流水表（order_id / from / to / operator / note），只追加

## admins（后台管理员）

username / password_hash（bcrypt）/ status。登录用 JWT，密钥从环境变量读取。

## configs（系统配置，KV）

- `ai_daily_free_limit`（默认 3）
- `work_review_enabled`（默认 false）
- AI provider 的模型/端点选择等

## 迁移与演示数据

- 迁移幂等，可重复执行
- 内置演示数据：1 个底板 + 每分类 3 个元素 + 2 个分类 4 个商品（含 2 组规格 SKU）+ 1 个演示订单全链路状态，供回归验证
