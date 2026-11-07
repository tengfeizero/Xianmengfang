package model

import "time"

// Model 公共字段
type Model struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User 小程序用户
type User struct {
	Model
	OpenID   string `gorm:"uniqueIndex;size:64" json:"-"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"` // 1 正常 0 禁用
}

// Admin 后台管理员（v1.4 起归属角色，权限由角色决定）
type Admin struct {
	Model
	Username     string     `gorm:"uniqueIndex;size:32" json:"username"`
	PasswordHash string     `json:"-"`
	Nickname     string     `gorm:"size:32" json:"nickname"`
	RoleID       int64      `gorm:"index" json:"role_id"`
	Status       int        `json:"status"` // 1 正常 0 禁用
	LastLoginAt  *time.Time `json:"last_login_at"`
}

// Role 后台角色：权限码集合存 JSON 数组，内置角色不可删除
type Role struct {
	Model
	Code    string `gorm:"uniqueIndex;size:32" json:"code"` // super / ops / viewer / 自定义
	Name    string `gorm:"size:32" json:"name"`
	Remark  string `gorm:"size:255" json:"remark"`
	Perms   string `gorm:"type:text" json:"perms"` // JSON 数组，如 ["order:view"]
	Builtin int    `json:"builtin"`                // 1 内置（不可删）
	Status  int    `json:"status"`                 // 1 启用 0 停用
}

// AuditLog 敏感操作审计日志：后台写操作、登录、越权拦截全量留痕
type AuditLog struct {
	Model
	AdminID    int64  `gorm:"index" json:"admin_id"`
	AdminName  string `gorm:"size:32;index" json:"admin_name"`
	Action     string `gorm:"size:48;index" json:"action"` // order.ship / config.update / login / deny
	TargetType string `gorm:"size:32" json:"target_type"`
	TargetID   string `gorm:"size:32" json:"target_id"`
	Method     string `gorm:"size:8" json:"method"`
	Path       string `gorm:"size:128" json:"path"`
	IP         string `gorm:"size:64" json:"ip"`
	UserAgent  string `gorm:"size:255" json:"user_agent"`
	Detail     string `gorm:"type:text" json:"detail"`       // 请求摘要（密码/密钥已脱敏）
	Result     string `gorm:"size:12;index" json:"result"`   // ok / fail / deny
	StatusCode int    `json:"status_code"`                   // HTTP 状态码
	ErrMsg     string `gorm:"size:255" json:"err_msg"`
}

// Config KV 系统配置
type Config struct {
	Key   string `gorm:"primaryKey;size:64" json:"key"`
	Value string `json:"value"`
}

// ProductCategory 商品分类
type ProductCategory struct {
	Model
	Name   string `gorm:"size:32" json:"name"`
	Sort   int    `json:"sort"`
	Status int    `json:"status"` // 1 上架 0 下架
}

// Product 商品 SPU
type Product struct {
	Model
	CategoryID         int64  `gorm:"index" json:"category_id"`
	Title              string `gorm:"size:64" json:"title"`
	Subtitle           string `gorm:"size:128" json:"subtitle"`       // 卖点副标题，列表第二行
	Description        string `json:"description"`                    // 兼容字段：一句话简介，详情页不用
	Images             string `gorm:"type:text" json:"images"`        // json 数组，商品主图（首张为封面）
	OriginalPriceCents int64  `json:"original_price_cents"`           // 划线价，0 表示不展示，不参与结算
	Unit               string `gorm:"size:8" json:"unit"`             // 计价单位：件/个/套
	Detail             string `gorm:"type:text" json:"detail"`        // 文字详情，空行分段
	DetailImages       string `gorm:"type:text" json:"detail_images"` // json 数组，详情长图
	ServiceNotes       string `gorm:"size:255" json:"service_notes"`  // 服务承诺 + 售后例外
	Sort               int    `json:"sort"`                           // 排序权重，大的靠前
	Status             int    `json:"status"`                         // 1 上架 0 下架
}

// SKU 商品规格
type SKU struct {
	Model
	ProductID  int64  `gorm:"index" json:"product_id"`
	Specs      string `gorm:"size:255" json:"specs"` // json，如 {"尺码":"M","颜色":"白"}
	PriceCents int64  `json:"price_cents"`           // 金额一律分
	Stock      int64  `json:"stock"`
	Code       string `gorm:"size:32" json:"code"`
}

// Board 拼装底板
type Board struct {
	Model
	Name   string `gorm:"size:32" json:"name"`
	URL    string `gorm:"size:255" json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Status int    `json:"status"` // 1 上架 0 下架
}

// Element 拼装元素（PNG 图层素材）
type Element struct {
	Model
	Category string `gorm:"size:16;index" json:"category"` // hair/hat/lash/expression/outfit/accessory
	Name     string `gorm:"size:32" json:"name"`
	ImageURL string `gorm:"size:255" json:"image_url"`
	Anchor   string `gorm:"size:128" json:"anchor"` // json 默认锚点 {"x":0.5,"y":0.2,"scale":1}
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

// Work 用户作品（M1 启用，先建表）
type Work struct {
	Model
	UserID       int64  `gorm:"index" json:"user_id"`
	Type         string `gorm:"size:16" json:"type"` // collage / ai_text / ai_photo
	ImageURL     string `json:"image_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Meta         string `gorm:"type:text" json:"meta"`
	Status       int    `json:"status"` // 1 正常 0 已删除 2 被下架
}

// Address 用户收货地址
type Address struct {
	Model
	UserID    int64  `gorm:"index" json:"user_id"`
	Name      string `gorm:"size:32" json:"name"`
	Phone     string `gorm:"size:16" json:"phone"`
	Province  string `gorm:"size:32" json:"province"`
	City      string `gorm:"size:32" json:"city"`
	District  string `gorm:"size:32" json:"district"`
	Detail    string `gorm:"size:255" json:"detail"`
	IsDefault int    `json:"is_default"` // 1 默认 0 普通
	Status    int    `json:"status"`     // 1 正常 0 已删除
}

// Order 订单
type Order struct {
	Model
	OrderNo      string `gorm:"uniqueIndex;size:32" json:"order_no"`
	UserID       int64  `gorm:"index" json:"user_id"`
	Status       string `gorm:"size:20;index" json:"status"`
	PayType      string `gorm:"size:10" json:"pay_type"` // mock / wxpay
	TotalCents   int64  `json:"total_cents"`
	RefundFlag   int    `json:"refund_flag"` // 仅标记，MVP 无退款资金流
	CancelReason string `json:"cancel_reason"`
	// 收货信息快照（下单时从 Address 复制，地址后续修改不影响历史订单）
	ReceiverName    string     `gorm:"size:32" json:"receiver_name"`
	ReceiverPhone   string     `gorm:"size:16" json:"receiver_phone"`
	ReceiverAddress string     `gorm:"size:255" json:"receiver_address"`
	PaidAt          *time.Time `json:"paid_at"`
	ShippedAt       *time.Time `json:"shipped_at"`
	DoneAt          *time.Time `json:"done_at"`
}

// OrderItem 订单明细（SKU 快照 + 绑定作品）
type OrderItem struct {
	Model
	OrderID    int64  `gorm:"index" json:"order_id"`
	SKUID      int64  `gorm:"column:sku_id;index" json:"sku_id"`
	WorkID     *int64 `json:"work_id"` // M0 可空，M1 绑定作品
	Title      string `gorm:"size:64" json:"title"`
	Specs      string `gorm:"size:255" json:"specs"`
	PriceCents int64  `json:"price_cents"`
	Qty        int    `json:"qty"`
}

// OrderEvent 订单状态流水，只追加
type OrderEvent struct {
	Model
	OrderID  int64  `gorm:"index" json:"order_id"`
	From     string `gorm:"size:20" json:"from"`
	To       string `gorm:"size:20" json:"to"`
	Operator string `gorm:"size:32" json:"operator"` // user:12 / admin:1
	Note     string `gorm:"size:255" json:"note"`
}

// Shipment 发货记录
type Shipment struct {
	Model
	OrderID    int64  `gorm:"index" json:"order_id"`
	Company    string `gorm:"size:32" json:"company"`
	TrackingNo string `gorm:"size:32" json:"tracking_no"`
}

// AITask AI 生成任务（M1 启用，先建表）
type AITask struct {
	Model
	UserID    int64  `gorm:"index" json:"user_id"`
	Type      string `gorm:"size:16" json:"type"` // ai_text / ai_photo
	Params    string `gorm:"type:text" json:"params"`
	Status    string `gorm:"size:12" json:"status"` // pending/done/failed
	Provider  string `gorm:"size:32" json:"provider"`
	CostCents int64  `json:"cost_cents"`
	ResultURL string `json:"result_url"`
	Err       string `gorm:"size:255" json:"err"`
	QuotaDate string `gorm:"size:10;index" json:"quota_date"`
}

// CartItem 购物车项（v1.0）。WorkID 用 0 表示未绑作品（而非 NULL），
// 否则 SQLite 唯一索引视 NULL 互不相等，同款无作品会重复插入。
type CartItem struct {
	Model
	UserID  int64 `gorm:"index;uniqueIndex:idx_cart_user_sku_work,priority:1" json:"user_id"`
	SKUID   int64 `gorm:"column:sku_id;uniqueIndex:idx_cart_user_sku_work,priority:2" json:"sku_id"`
	WorkID  int64 `gorm:"uniqueIndex:idx_cart_user_sku_work,priority:3" json:"work_id"` // 0 = 未绑作品
	Qty     int   `json:"qty"`                                                          // 单件上限 99
	Checked int   `json:"checked"`                                                      // 1 勾选 0 不勾，默认 1
}

// AllMigrations 全部迁移模型，AutoMigrate 幂等
func AllMigrations() []any {
	return []any{
		&User{}, &Admin{}, &Config{}, &Address{},
		&ProductCategory{}, &Product{}, &SKU{}, &Work{},
		&Board{}, &Element{},
		&Order{}, &OrderItem{}, &OrderEvent{}, &Shipment{}, &AITask{},
		&CartItem{},
		&Role{}, &AuditLog{},
	}
}
