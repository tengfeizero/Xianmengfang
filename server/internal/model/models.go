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

// Admin 后台管理员
type Admin struct {
	Model
	Username     string `gorm:"uniqueIndex;size:32" json:"username"`
	PasswordHash string `json:"-"`
	Status       int    `json:"status"`
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
	CategoryID  int64  `gorm:"index" json:"category_id"`
	Title       string `gorm:"size:64" json:"title"`
	Description string `json:"description"`
	Images      string `gorm:"type:text" json:"images"` // json 数组
	Status      int    `json:"status"`                  // 1 上架 0 下架
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

// Order 订单
type Order struct {
	Model
	OrderNo      string     `gorm:"uniqueIndex;size:32" json:"order_no"`
	UserID       int64      `gorm:"index" json:"user_id"`
	Status       string     `gorm:"size:20;index" json:"status"`
	PayType      string     `gorm:"size:10" json:"pay_type"` // mock / wxpay
	TotalCents   int64      `json:"total_cents"`
	RefundFlag   int        `json:"refund_flag"` // 仅标记，MVP 无退款资金流
	CancelReason string     `json:"cancel_reason"`
	PaidAt       *time.Time `json:"paid_at"`
	ShippedAt    *time.Time `json:"shipped_at"`
	DoneAt       *time.Time `json:"done_at"`
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

// AllMigrations 全部迁移模型，AutoMigrate 幂等
func AllMigrations() []any {
	return []any{
		&User{}, &Admin{}, &Config{},
		&ProductCategory{}, &Product{}, &SKU{}, &Work{},
		&Board{}, &Element{},
		&Order{}, &OrderItem{}, &OrderEvent{}, &Shipment{}, &AITask{},
	}
}
