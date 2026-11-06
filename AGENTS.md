# 项目规则（AGENTS.md）— 现梦坊

> AI 每次会话自动读取。改任何代码前先读 docs/PRD.md 和 docs/CONVENTIONS.md。

- 技术栈：Go 1.22+（Gin + GORM + SQLite，后期迁 PostgreSQL，禁止提前引入 MQ/Redis/ES）；前端为微信小程序原生 + Vue 3 + TS + Element Plus
- 分层：api 只做参数绑定和鉴权，repo 只做 CRUD，业务规则（金额、库存、状态机、配额）全部在 service 层
- 金额一律 int（分），禁止 float 参与金额计算；数量 int
- 每次改动必须：`go vet ./... && go test ./...` 通过，Web 后端实测后才能报告"完成"；必须给出证据（命令输出/截图/断言），不接受"我改好了"
- 订单状态机见 docs/DATA-MODEL.md，状态流转只允许在 service 层发生
- 改完代码必须同步 docs/（API.md / DATA-MODEL.md / CHANGELOG.md / TODO.md）；一个功能一个 commit，提交信息写"为什么"
- 密钥（AI 生图 API key、JWT secret、小程序 secret）只走环境变量或 data/ 下的本地文件，绝不进仓库；data/、*.db 不提交
- 用户上传内容与 AI 生成内容必须过微信内容安全校验后才可作为作品使用
