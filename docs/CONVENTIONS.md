# CONVENTIONS — 现梦坊代码纪律

> 人定一次，AI 每次遵守。改动本文件必须经过用户确认。

## 目录结构（Go 后端）

```
server/
  cmd/xmf/          # 入口 main.go
  internal/
    api/            # 路由 + 参数绑定（薄层），不含业务规则
    service/        # 业务规则：金额、库存、订单状态机、AI 配额
    repo/           # CRUD，只操作数据库
    model/          # GORM 模型与迁移
    ai/             # AI 生图 provider 封装（可换厂商，不碰业务）
  data/             # 运行时数据（.db、密钥文件），不进 git
  docs/             # 项目文档（在仓库根 docs/）
web/                # Vue 3 + TS 管理后台
miniprogram/        # 微信小程序原生
```

## 硬性纪律

1. **金额**：`int`，单位分。字段名 `price_cents / total_cents`。任何地方禁止 float/money 类型。
2. **库存**：扣减必须用带条件的原子更新（`UPDATE ... WHERE stock >= n`），禁止"读出来算好再写回"。库存流水只追加不修改。
3. **状态机**：订单状态只允许合法流转（见 DATA-MODEL.md），非法流转返回 409，不静默处理。
4. **业务规则**只写在 service 层；api 层发现 if/else 业务判断即为坏味道。
5. **AI 配额**：每用户每日免费次数从配置读取（默认 3），计数持久化，不做内存计数。
6. **错误返回**：统一 `{code, msg, data}`；msg 必须是人话，禁止把 500 internal error 直接抛给用户。
7. **密钥**：环境变量优先；本地开发用 `data/` 下文件（0600），两者都不进 git。
8. **测试**：业务规则先写测试再实现；演示数据要自洽（订单、库存、金额对得上）。
9. **提交**：一个功能一个 commit；`feat/fix/docs` 前缀；信息写"为什么"。
