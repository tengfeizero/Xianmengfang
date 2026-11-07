package api

import (
	"encoding/json"
	"errors"

	"xmf/internal/model"
	"xmf/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	DB         *gorm.DB
	Secret     string
	Auth       *service.AuthService
	Orders     *service.OrderService
	Addresses  *service.AddressService
	Carts      *service.CartService
	AI         *service.AIService
	RBAC       *service.RBACService
	Audit      *service.AuditService
	Community  *service.CommunityService
	UploadsDir string
	Dev        bool // 是否允许开发者模拟登录
}

// ---- 登录 ----

type loginReq struct {
	Code string `json:"code"`
}

func (h *Handlers) wxLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	u, err := h.Auth.WxLogin(req.Code)
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	h.issueUserToken(c, u)
}

type devLoginReq struct {
	OpenID string `json:"openid"`
}

func (h *Handlers) devLogin(c *gin.Context) {
	if !h.Dev {
		fail(c, 403, "开发者模拟登录未开启")
		return
	}
	var req devLoginReq
	if err := c.ShouldBindJSON(&req); err != nil || req.OpenID == "" {
		fail(c, 400, "openid 不能为空")
		return
	}
	u, err := h.Auth.DevLogin(req.OpenID)
	if err != nil {
		fail(c, 400, err.Error())
		return
	}
	h.issueUserToken(c, u)
}

func (h *Handlers) issueUserToken(c *gin.Context, u *model.User) {
	token, err := service.GenerateToken(h.Secret, "user", u.ID)
	if err != nil {
		fail(c, 500, "登录失败，请稍后再试")
		return
	}
	ok(c, gin.H{"token": token, "user": u})
}

type adminLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handlers) adminLogin(c *gin.Context) {
	var req adminLoginReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		fail(c, 400, "请输入用户名和密码")
		return
	}
	a, err := h.Auth.AdminLogin(req.Username, req.Password)
	if err != nil {
		h.auditLogin(c, 0, req.Username, "fail", err.Error(), 401)
		fail(c, 401, err.Error())
		return
	}
	token, err := service.GenerateToken(h.Secret, "admin", a.ID)
	if err != nil {
		fail(c, 500, "登录失败，请稍后再试")
		return
	}
	h.RBAC.TouchLogin(a.ID)
	h.auditLogin(c, a.ID, a.Username, "ok", "登录成功", 200)
	ok(c, gin.H{
		"token": token,
		"admin": gin.H{"id": a.ID, "username": a.Username, "nickname": a.Nickname},
	})
}

// ---- 商品 ----

func (h *Handlers) productList(c *gin.Context) {
	var products []model.Product
	if err := h.DB.Where("status = 1").Order("sort desc, id").Find(&products).Error; err != nil {
		fail(c, 500, "获取商品失败，请稍后再试")
		return
	}
	type vo struct {
		model.Product
		SKUs []model.SKU `json:"skus"`
	}
	list := make([]vo, 0, len(products))
	for _, p := range products {
		var skus []model.SKU
		h.DB.Where("product_id = ?", p.ID).Order("id").Find(&skus)
		list = append(list, vo{Product: p, SKUs: skus})
	}
	ok(c, list)
}

func (h *Handlers) productDetail(c *gin.Context) {
	var p model.Product
	if err := h.DB.First(&p, c.Param("id")).Error; err != nil || p.Status != 1 {
		fail(c, 404, "商品不存在或已下架")
		return
	}
	var skus []model.SKU
	// 按 id 升序：小程序默认选中第一个规格，顺序必须稳定（= 后台录入顺序）
	h.DB.Where("product_id = ?", p.ID).Order("id").Find(&skus)
	ok(c, gin.H{"product": p, "skus": skus})
}

// ---- 订单（用户端） ----

type createOrderReq struct {
	AddressID int64 `json:"address_id"` // 收货地址（M1.5 起必传，快照进订单）
	Items     []struct {
		SKUID  int64  `json:"sku_id"`
		WorkID *int64 `json:"work_id"`
		Qty    int    `json:"qty"`
	} `json:"items" binding:"required,min=1"`
}

func (h *Handlers) orderCreate(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	var addr *model.Address
	if req.AddressID > 0 {
		a, err := h.Addresses.GetForOrder(claims.UserID, req.AddressID)
		if err != nil {
			if errors.Is(err, service.ErrAddressForbidden) {
				fail(c, 403, "无权使用该地址")
			} else {
				fail(c, 400, "收货地址不存在，请重新选择")
			}
			return
		}
		addr = a
	} else {
		fail(c, 400, "请选择收货地址")
		return
	}
	items := make([]service.ItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, service.ItemInput{SKUID: it.SKUID, WorkID: it.WorkID, Qty: it.Qty})
	}
	res, err := h.Orders.Create(claims.UserID, items, addr)
	if err != nil {
		failErr(c, err)
		return
	}
	ok(c, gin.H{"order": res.Order, "total_cents": res.TotalCents})
}

func (h *Handlers) orderList(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var orders []model.Order
	q := h.DB.Where("user_id = ?", claims.UserID).Order("id desc")
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if err := q.Find(&orders).Error; err != nil {
		fail(c, 500, "获取订单失败，请稍后再试")
		return
	}
	ok(c, withDetails(h.DB, orders))
}

func (h *Handlers) orderDetail(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	o, err := h.Orders.GetForUser(paramID(c), claims.UserID)
	if err != nil {
		failErr(c, err)
		return
	}
	ok(c, orderDetailView(h.DB, *o))
}

func (h *Handlers) orderConfirm(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Orders.Confirm(paramID(c), claims.UserID); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) orderCancel(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.Orders.Cancel(paramID(c), claims.UserID, false, req.Reason); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

// ---- 订单（管理后台） ----

func (h *Handlers) adminOrderList(c *gin.Context) {
	var orders []model.Order
	q := h.DB.Order("id desc")
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if no := c.Query("order_no"); no != "" {
		q = q.Where("order_no LIKE ?", "%"+no+"%")
	}
	if uid := c.Query("user_id"); uid != "" {
		q = q.Where("user_id = ?", uid)
	}
	if err := q.Find(&orders).Error; err != nil {
		fail(c, 500, "获取订单失败，请稍后再试")
		return
	}
	list := withDetails(h.DB, orders)

	// 批量补买家昵称（列表页直接可见，免得只看到 user_id）
	ids := make([]int64, 0, len(orders))
	seen := map[int64]bool{}
	for _, o := range orders {
		if !seen[o.UserID] {
			seen[o.UserID] = true
			ids = append(ids, o.UserID)
		}
	}
	users := map[int64]model.User{}
	if len(ids) > 0 {
		var us []model.User
		if err := h.DB.Where("id IN ?", ids).Find(&us).Error; err == nil {
			for _, u := range us {
				users[u.ID] = u
			}
		}
	}
	for i := range list {
		o := list[i]["order"].(model.Order)
		if u, ok := users[o.UserID]; ok {
			list[i]["user"] = gin.H{"id": u.ID, "nickname": u.Nickname, "avatar": u.Avatar}
		}
	}
	ok(c, list)
}

// adminOrderDetail 管理端单条订单完整详情：明细（含图）、物流、状态流水、金额明细、买家信息
func (h *Handlers) adminOrderDetail(c *gin.Context) {
	var o model.Order
	if err := h.DB.First(&o, paramID(c)).Error; err != nil {
		fail(c, 404, "订单不存在")
		return
	}
	view := orderDetailView(h.DB, o)
	var u model.User
	if err := h.DB.First(&u, o.UserID).Error; err == nil {
		view["user"] = gin.H{"id": u.ID, "nickname": u.Nickname, "avatar": u.Avatar}
	}
	ok(c, view)
}

func (h *Handlers) adminOrderAccept(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Orders.Accept(paramID(c), claims.UserID); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminOrderPack(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Orders.Pack(paramID(c), claims.UserID); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminOrderReadyShip(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Orders.ReadyShip(paramID(c), claims.UserID); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminOrderShip(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req struct {
		Company    string `json:"company"`
		TrackingNo string `json:"tracking_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if err := h.Orders.Ship(paramID(c), claims.UserID, req.Company, req.TrackingNo); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminOrderCancel(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.Orders.Cancel(paramID(c), claims.UserID, true, req.Reason); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminOrderRefundMark(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Orders.MarkRefund(paramID(c), claims.UserID); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

// ---- 辅助 ----

// withDetails 列表用：明细 + 物流（不查图片与流水，列表不需要）
func withDetails(db *gorm.DB, orders []model.Order) []gin.H {
	list := make([]gin.H, 0, len(orders))
	for _, o := range orders {
		var items []model.OrderItem
		db.Where("order_id = ?", o.ID).Find(&items)
		var shipment *model.Shipment
		var sh model.Shipment
		if err := db.Where("order_id = ?", o.ID).Order("id desc").First(&sh).Error; err == nil {
			shipment = &sh
		}
		list = append(list, gin.H{"order": o, "items": items, "shipment": shipment})
	}
	return list
}

// orderItemView 订单明细展示视图：订单快照字段 + 商品首图 + 绑定作品图。
// 图不在快照里，取当前商品/作品的图（商品换图会改变历史订单展示，MVP 接受）。
type orderItemView struct {
	model.OrderItem
	ImageURL     string `json:"image_url"`
	WorkImageURL string `json:"work_image_url"`
}

// orderAmounts 金额明细（分）。订单表无运费/优惠独立字段，由合计差额反推。
type orderAmounts struct {
	ItemsCents    int64 `json:"items_cents"`
	ShippingCents int64 `json:"shipping_cents"`
	DiscountCents int64 `json:"discount_cents"`
	TotalCents    int64 `json:"total_cents"`
}

// orderDetailView 单条订单的完整详情：明细（含图）、物流、状态流水、金额明细
func orderDetailView(db *gorm.DB, o model.Order) gin.H {
	var items []model.OrderItem
	db.Where("order_id = ?", o.ID).Order("id asc").Find(&items)

	views := make([]orderItemView, 0, len(items))
	var itemsTotal int64
	for _, it := range items {
		v := orderItemView{OrderItem: it}
		var sku model.SKU
		if err := db.First(&sku, it.SKUID).Error; err == nil {
			var p model.Product
			if err := db.First(&p, sku.ProductID).Error; err == nil {
				v.ImageURL = firstImage(p.Images)
			}
		}
		if it.WorkID != nil {
			var w model.Work
			if err := db.First(&w, *it.WorkID).Error; err == nil {
				v.WorkImageURL = w.ThumbnailURL
				if v.WorkImageURL == "" {
					v.WorkImageURL = w.ImageURL
				}
			}
		}
		itemsTotal += it.PriceCents * int64(it.Qty)
		views = append(views, v)
	}

	var shipment *model.Shipment
	var sh model.Shipment
	if err := db.Where("order_id = ?", o.ID).Order("id desc").First(&sh).Error; err == nil {
		shipment = &sh
	}

	events := make([]model.OrderEvent, 0, 4)
	db.Where("order_id = ?", o.ID).Order("id asc").Find(&events)

	shipping := o.TotalCents - itemsTotal
	if shipping < 0 {
		shipping = 0
	}
	discount := itemsTotal + shipping - o.TotalCents
	if discount < 0 {
		discount = 0
	}

	return gin.H{
		"order":    o,
		"items":    views,
		"shipment": shipment,
		"events":   events,
		"amounts":  orderAmounts{ItemsCents: itemsTotal, ShippingCents: shipping, DiscountCents: discount, TotalCents: o.TotalCents},
	}
}

// firstImage 从商品 images 的 JSON 数组里取首图
func firstImage(raw string) string {
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil || len(arr) == 0 {
		return ""
	}
	return arr[0]
}
