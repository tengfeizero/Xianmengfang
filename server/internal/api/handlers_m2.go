package api

import (
	"fmt"

	"xmf/internal/model"

	"github.com/gin-gonic/gin"
)

// ---- 后台：商品分类 ----

func (h *Handlers) adminCategoryList(c *gin.Context) {
	var cats []model.ProductCategory
	if err := h.DB.Order("sort, id").Find(&cats).Error; err != nil {
		fail(c, 500, "获取分类失败，请稍后再试")
		return
	}
	ok(c, cats)
}

type categoryReq struct {
	Name string `json:"name" binding:"required"`
	Sort int    `json:"sort"`
}

func (h *Handlers) adminCategoryCreate(c *gin.Context) {
	var req categoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	cat := model.ProductCategory{Name: req.Name, Sort: req.Sort, Status: 1}
	if err := h.DB.Create(&cat).Error; err != nil {
		fail(c, 500, "创建失败，请稍后再试")
		return
	}
	ok(c, cat)
}

func (h *Handlers) adminCategoryUpdate(c *gin.Context) {
	var req categoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	res := h.DB.Model(&model.ProductCategory{}).Where("id = ?", c.Param("id")).
		Updates(map[string]any{"name": req.Name, "sort": req.Sort})
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "分类不存在或更新失败")
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminCategoryDelete(c *gin.Context) {
	// 分类下有商品时不允许删除
	var n int64
	h.DB.Model(&model.Product{}).Where("category_id = ?", c.Param("id")).Count(&n)
	if n > 0 {
		fail(c, 400, "该分类下还有商品，先移走商品再删除")
		return
	}
	res := h.DB.Model(&model.ProductCategory{}).Where("id = ?", c.Param("id")).Update("status", 0)
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "分类不存在或删除失败")
		return
	}
	ok(c, nil)
}

// ---- 后台：商品与 SKU ----

type productReq struct {
	CategoryID  int64  `json:"category_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Images      string `json:"images"`
	Status      *int   `json:"status"`
}

func (h *Handlers) adminProductList(c *gin.Context) {
	var products []model.Product
	q := h.DB.Order("id desc")
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if err := q.Find(&products).Error; err != nil {
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

func (h *Handlers) adminProductCreate(c *gin.Context) {
	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	var cat model.ProductCategory
	if err := h.DB.Where("status = 1").First(&cat, req.CategoryID).Error; err != nil {
		fail(c, 400, "分类不存在")
		return
	}
	if req.Images == "" {
		req.Images = "[]"
	}
	p := model.Product{CategoryID: req.CategoryID, Title: req.Title, Description: req.Description, Images: req.Images, Status: 1}
	if err := h.DB.Create(&p).Error; err != nil {
		fail(c, 500, "创建失败，请稍后再试")
		return
	}
	ok(c, p)
}

func (h *Handlers) adminProductUpdate(c *gin.Context) {
	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	updates := map[string]any{
		"category_id": req.CategoryID, "title": req.Title,
		"description": req.Description, "images": req.Images,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	res := h.DB.Model(&model.Product{}).Where("id = ?", c.Param("id")).Updates(updates)
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "商品不存在或更新失败")
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminProductDelete(c *gin.Context) {
	res := h.DB.Model(&model.Product{}).Where("id = ?", c.Param("id")).Update("status", 0)
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "商品不存在或下架失败")
		return
	}
	ok(c, nil)
}

type skuReq struct {
	ProductID  int64  `json:"product_id"`
	Specs      string `json:"specs" binding:"required"`
	PriceCents int64  `json:"price_cents" binding:"required,min=1"`
	Stock      int64  `json:"stock"`
	Code       string `json:"code"`
}

func (h *Handlers) adminSKUGetByID(id string) (*model.SKU, bool) {
	var sku model.SKU
	if err := h.DB.First(&sku, id).Error; err != nil {
		return nil, false
	}
	return &sku, true
}

func (h *Handlers) adminSKUCreate(c *gin.Context) {
	var req skuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if req.ProductID == 0 {
		fail(c, 400, "缺少商品 ID")
		return
	}
	var p model.Product
	if err := h.DB.First(&p, req.ProductID).Error; err != nil {
		fail(c, 400, "商品不存在")
		return
	}
	if req.Stock < 0 {
		fail(c, 400, "库存不能为负数")
		return
	}
	sku := model.SKU{ProductID: req.ProductID, Specs: req.Specs, PriceCents: req.PriceCents, Stock: req.Stock, Code: req.Code}
	if err := h.DB.Create(&sku).Error; err != nil {
		fail(c, 500, "创建失败，请稍后再试")
		return
	}
	ok(c, sku)
}

func (h *Handlers) adminSKUUpdate(c *gin.Context) {
	var req skuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if req.Stock < 0 {
		fail(c, 400, "库存不能为负数")
		return
	}
	updates := map[string]any{"specs": req.Specs, "price_cents": req.PriceCents, "stock": req.Stock, "code": req.Code}
	res := h.DB.Model(&model.SKU{}).Where("id = ?", c.Param("id")).Updates(updates)
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "规格不存在或更新失败")
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminSKUDelete(c *gin.Context) {
	sku, exists := h.adminSKUGetByID(c.Param("id"))
	if !exists {
		fail(c, 404, "规格不存在")
		return
	}
	// 有订单引用的 SKU 不允许删除，只能把库存清零下架
	var n int64
	h.DB.Model(&model.OrderItem{}).Where("sku_id = ?", sku.ID).Count(&n)
	if n > 0 {
		fail(c, 400, "该规格已有订单记录，不能删除（可把库存改为 0）")
		return
	}
	if err := h.DB.Delete(sku).Error; err != nil {
		fail(c, 500, "删除失败，请稍后再试")
		return
	}
	ok(c, nil)
}

// ---- 后台：素材与底板列表（含下架） ----

func (h *Handlers) adminBoardList(c *gin.Context) {
	var boards []model.Board
	if err := h.DB.Order("id").Find(&boards).Error; err != nil {
		fail(c, 500, "获取底板失败，请稍后再试")
		return
	}
	ok(c, boards)
}

func (h *Handlers) adminElementList(c *gin.Context) {
	q := h.DB.Order("category, sort, id")
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	var els []model.Element
	if err := q.Find(&els).Error; err != nil {
		fail(c, 500, "获取素材失败，请稍后再试")
		return
	}
	ok(c, els)
}

// ---- 后台：系统配置 ----

var configWhitelist = map[string]bool{
	"ai_daily_free_limit": true,
	"work_review_enabled": true,
}

func (h *Handlers) adminConfigList(c *gin.Context) {
	var cfgs []model.Config
	if err := h.DB.Order("key").Find(&cfgs).Error; err != nil {
		fail(c, 500, "获取配置失败，请稍后再试")
		return
	}
	ok(c, cfgs)
}

func (h *Handlers) adminConfigUpdate(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if !configWhitelist[req.Key] {
		fail(c, 400, "不支持的配置项")
		return
	}
	if req.Key == "ai_daily_free_limit" {
		var n int
		if _, err := fmt.Sscanf(req.Value, "%d", &n); err != nil || n < 0 || n > 1000 {
			fail(c, 400, "每日免费次数必须是 0~1000 的整数")
			return
		}
	}
	res := h.DB.Model(&model.Config{}).Where("key = ?", req.Key).Update("value", req.Value)
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "配置项不存在或更新失败")
		return
	}
	ok(c, nil)
}
