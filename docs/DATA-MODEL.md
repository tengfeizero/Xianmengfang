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
| user_id | int | 归属（作者） |
| type | text | `collage` 拼装 / `ai_text` 文生图 / `ai_photo` 照片转风格 |
| image_url / thumbnail_url | text | 成品图与缩略图 |
| meta | json | 拼装记录（元素 id、坐标缩放）或生成参数，便于复刻 |
| status | int | 1 正常（可用于商品）/ 0 已删除 / 2 被下架 |
| **is_public** | int | **1 已发布 / 0 私密（默认）**。发布后进入作品广场，AI 提示词一并公开 |
| **published_at** | datetime | 最近一次发布时间（取消发布不清空，重发刷新；「最新榜」按此排序） |
| **prompt** | text | **公开提示词**（ai_text 取生成描述、ai_photo 为空；拼装为空）。仅 `is_public=1` 时对外返回 |
| **like_count** | int | 点赞数（冗余计数，与 `work_likes` 同事务维护） |
| **use_count** | int | 使用次数（订单**完成**时按件数累加，见下） |

- **发布/取消发布**：作者可随时切换。取消发布只影响广场可见性，**已产生的点赞/使用计数保留**，购物车里已绑定的作品也不失效（仍可下单）。
- **下架（后台 status=2）**：立即从广场消失，且不可被选择下单；已发布状态保留，恢复上架后仍在广场。
- **使用次数与积分口径**：订单进入 `done` 时，对每条明细（`work_id` 非空）把对应作品的 `use_count += qty`，并给作者加 **qty 点积分**（自用同样计，见 PRD 3.4）。

### work_likes（作品点赞，v1.7）
| 字段 | 说明 |
|---|---|
| work_id | 被点赞作品（索引） |
| user_id | 点赞人（索引） |
| 唯一键 | `(work_id, user_id)`，重复点赞幂等；同一用户对同一作品最多 1 条 |

- 不能点赞自己的作品（400）；点赞/取消点赞在同一事务内维护 `works.like_count`。
- 点赞记录在作品取消发布/下架后保留，重新发布时点赞数仍在；作品被作者删除（status=0）时点赞记录不清理（只读展示不用），避免误删他人数据。

### point_logs（积分流水，v1.7）
| 字段 | 说明 |
|---|---|
| user_id | 归属（索引） |
| delta | 变动值：+N 获得 / -N 消耗 |
| balance | 变动后余额（对账用，同时冗余在 `users.points`） |
| reason | `work_used` 作品被使用 / `ai_spend` 生成扣减 / `admin_adjust` 后台调整（预留） |
| ref_type / ref_id | 关联对象：`order` + 订单 id、`work` + 作品 id、`ai_task` + 任务 id |
| remark | 可读说明，如「作品《小熊猫》被使用 2 件」 |

- 只在同一事务内与余额变更一起写入，**流水只追加**，不提供删除接口。

### users.points（积分余额，v1.7）

- 单一钱包，整数，不允许负数：扣减用 `UPDATE users SET points = points - ? WHERE id = ? AND points >= ?` 的条件更新，避免并发超扣。
- 获得：订单完成时按件数给作品作者 +N（自用也计）。
- 消费：AI 生成每日免费额度用完后，1 积分 = 1 张图（`configs.ai_daily_free_limit` 管免费张数）。

### ai_tasks.pay_type（v1.7）

`free`（每日免费额度内）/ `point`（积分抵扣）。免费额度按「当天 pay_type=free 的任务数」判定，
积分抵扣不计入免费额度，避免「用积分反而占掉免费次数」。

### boards（底板）· elements（拼装元素）

**关系：一块底板 → 多个元素（一对多，`elements.board_id`）。** 元素只在所属底板的拼装页出现，
合成接口会校验放置的每个元素都属于请求里的底板，跨底板混搭直接 400。
理由：素材是在某张底板的画布（尺寸/人物位置/风格）上产出的，混用到别的底板会错位、串风格
（例如小男孩底板出现裙子和双马尾）。

boards
| 字段 | 说明 |
|---|---|
| name / url | 底板名称与 PNG（透明背景，画布基准，如 400×500） |
| width / height | 画布像素尺寸，锚点与相对坐标都以此为准 |
| status | 1 上架（小程序可选）/ 0 下架 |

elements
| 字段 | 说明 |
|---|---|
| **board_id** | **所属底板（必填，索引）；元素的尺寸与锚点按该底板画布计算** |
| category | 元素分类：hair/hat/lash/expression/outfit/accessory（决定图层叠放顺序） |
| name / image_url / sort / status | PNG 素材、排序、上下架 |
| anchor | json，槽位锚点（x/y/scale），拼装时按锚点定位 |

约束：下架底板前必须先下架该底板下所有元素（接口 409 拦截），避免「底板不可选、元素却仍挂在启用列表」；
删除底板/元素均为软删（status=0），历史作品里的合成图不受影响。

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

### cart_items（购物车，v1.0）
- user_id（index）/ sku_id / work_id（可空，NULL 视为"未绑作品"参与唯一键）
- **唯一键 (user_id, sku_id, work_id)**：同款同图重复加购 = 数量累加（单件 qty 上限 99，条数上限 50）
- qty int / checked int（1 勾选 0 不勾，默认 1；勾选态存服务端，跨设备一致）
- 不存快照：列表实时 join skus/products/works 渲染，金额一律按 SKU 现价；join 不上（商品下架/SKU 删除/作品软删）即渲染为"失效项"，标灰不可结算，由用户手动清理（SKU 删除有订单引用保护，物理删除罕见，不值得为它存快照）
- 不锁库存：加购只校验库存 ≥ 1，扣减仍在下单事务里原子完成

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
pending_pay ──支付──→ pending_accept ──接单──→ producing ──生产完成──→ packing ──打包完成──→ pending_ship ──发货──→ shipped ──确认收货──→ done
     │                    │                        │                      │                        │
     └────── cancel ──────┴────────────────────────┴──────────────────────┴────────────────────────┘（均记 cancel_reason）
```

- 状态含义：`pending_accept` 待接单 / `producing` 生产中 / `packing` 打包中 / `pending_ship` 待发货（等快递揽收）/ `shipped` 已发货 / `done` 已完成 / `canceled` 已取消
- mock 支付：下单直接进入 `pending_accept`（跳过 pending_pay），`pay_type=mock`
- 取消权限：用户仅可取消 `pending_accept`；管理员可取消 `pending_accept`/`producing`/`packing`/`pending_ship`，均恢复库存
- `done` 之前允许后台标记退款标记（仅标记字段，MVP 无退款资金流）
- 状态变更写 `order_events` 流水表（order_id / from / to / operator / note），只追加
- 兼容说明：v1.3.5 之前的历史订单若停在 `pending_ship`，可直接发货（无需补生产/打包动作）

## admins（后台管理员，v1.4 起归属角色）

username（唯一）/ password_hash（bcrypt）/ nickname / role_id（→ roles.id）/ status / last_login_at。

- 登录鉴权链：JWT 只带 admin_id，**权限每次请求现查 DB**（role → perms），因此禁用账号、改角色权限立即生效，不等 token 过期
- 保护规则：不能禁用/降级/删除自己；系统必须保留至少一个启用状态的超级管理员

## roles（角色，v1.4）

code（唯一，内置 super/ops/viewer）/ name / remark / perms（JSON 数组，权限码）/ builtin（内置不可删）/ status。

- `super` 角色 perms 恒为 `["*"]`，拥有全部权限且不可改权限
- 自定义角色 code 自动生成（`role_<ts>`），删除前校验无账号使用
- 权限码清单见 API.md「权限码」；新增功能时先加权限码再挂路由

## audit_logs（敏感操作日志，v1.4）

admin_id / admin_name / action（如 `order.ship`、`config.update`、`auth.login`、`deny`）/ target_type / target_id /
method / path / ip / user_agent / detail（请求体摘要，**密码与密钥字段脱敏为 `***`**）/ result（ok / fail / deny）/ status_code / err_msg。

- 采集方式：管理端写方法（POST/PUT/DELETE）由 `AuditWrite` 中间件统一记录；登录成功/失败、越权拦截单独记录
- 日志只追加，不提供删除接口；查询需 `audit:view` 权限

## configs（系统配置，KV）

- `ai_daily_free_limit`（默认 3）
- `work_review_enabled`（默认 false）
- AI provider 的模型/端点选择等

## 迁移与演示数据

- 迁移幂等，可重复执行
- 内置演示数据：4 个底板 + 24 个元素（hair×5、hat×5、lash×1、expression×4、outfit×4、accessory×5，AI 生成套件，内嵌 go:embed aipack）+ 4 个分类 8 个商品（含多规格 SKU）+ 1 个演示订单全链路状态，供回归验证。素材缺失时种子自动落盘；演示记录按「名字匹配」回填，不碰后台自建内容。
