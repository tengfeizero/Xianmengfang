package api

import (
	"xmf/internal/model"
	"xmf/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	DB     *gorm.DB
	Secret string
	Auth   *service.AuthService
	Orders *service.OrderService
	Dev    bool // 是否允许开发者模拟登录
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
		fail(c, 401, err.Error())
		return
	}
	token, err := service.GenerateToken(h.Secret, "admin", a.ID)
	if err != nil {
		fail(c, 500, "登录失败，请稍后再试")
		return
	}
	ok(c, gin.H{"token": token, "admin": gin.H{"id": a.ID, "username": a.Username}})
}

// ---- 商品 ----

func (h *Handlers) productList(c *gin.Context) {
	var products []model.Product
	if err := h.DB.Where("status = 1").Order("id").Find(&products).Error; err != nil {
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
		h.DB.Where("product_id = ?", p.ID).Find(&skus)
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
	h.DB.Where("product_id = ?", p.ID).Find(&skus)
	ok(c, gin.H{"product": p, "skus": skus})
}

// ---- 订单（用户端） ----

type createOrderReq struct {
	Items []struct {
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
	items := make([]service.ItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, service.ItemInput{SKUID: it.SKUID, WorkID: it.WorkID, Qty: it.Qty})
	}
	res, err := h.Orders.Create(claims.UserID, items)
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
	var o model.Order
	if err := h.DB.First(&o, c.Param("id")).Error; err != nil {
		fail(c, 404, "订单不存在")
		return
	}
	if o.UserID != claims.UserID {
		fail(c, 403, "无权查看该订单")
		return
	}
	ok(c, withDetails(h.DB, []model.Order{o})[0])
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
	if err := q.Find(&orders).Error; err != nil {
		fail(c, 500, "获取订单失败，请稍后再试")
		return
	}
	ok(c, withDetails(h.DB, orders))
}

func (h *Handlers) adminOrderAccept(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Orders.Accept(paramID(c), claims.UserID); err != nil {
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
