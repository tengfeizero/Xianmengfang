package api

import (
	"errors"

	"xmf/internal/service"

	"github.com/gin-gonic/gin"
)

// ---- 购物车（v1.0，PRD 3.2.2）----

type cartAddReq struct {
	SKUID  int64 `json:"sku_id" binding:"required"`
	WorkID int64 `json:"work_id"` // 0 = 不绑作品
	Qty    int   `json:"qty"`
}

func (h *Handlers) cartList(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	items, summary, err := h.Carts.List(claims.UserID)
	if err != nil {
		fail(c, 500, "获取购物车失败，请稍后再试")
		return
	}
	ok(c, gin.H{"items": items, "summary": summary})
}

func (h *Handlers) cartAdd(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req cartAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	item, err := h.Carts.Add(claims.UserID, req.SKUID, req.WorkID, req.Qty)
	if err != nil {
		failCartErr(c, err)
		return
	}
	ok(c, item)
}

type cartUpdateReq struct {
	Qty     *int `json:"qty"`
	Checked *int `json:"checked"`
}

func (h *Handlers) cartUpdate(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req cartUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if err := h.Carts.Update(claims.UserID, paramID(c), req.Qty, req.Checked); err != nil {
		failCartErr(c, err)
		return
	}
	ok(c, nil)
}

type cartCheckReq struct {
	Checked *int    `json:"checked" binding:"required"` // 指针：0 也是合法值
	ItemIDs []int64 `json:"item_ids"`                   // 空 = 全选/全不选
}

func (h *Handlers) cartSetChecked(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req cartCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if err := h.Carts.SetChecked(claims.UserID, *req.Checked, req.ItemIDs); err != nil {
		failCartErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) cartRemove(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Carts.Remove(claims.UserID, paramID(c)); err != nil {
		failCartErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) cartClear(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	onlyChecked := c.Query("checked") == "1"
	if err := h.Carts.Clear(claims.UserID, onlyChecked); err != nil {
		fail(c, 500, "清空失败，请稍后再试")
		return
	}
	ok(c, nil)
}

type cartCheckoutReq struct {
	AddressID int64   `json:"address_id"`
	ItemIDs   []int64 `json:"item_ids"` // 空 = 结算所有勾选项
}

func (h *Handlers) cartCheckout(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req cartCheckoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if req.AddressID <= 0 {
		fail(c, 400, "请选择收货地址")
		return
	}
	addr, err := h.Addresses.GetForOrder(claims.UserID, req.AddressID)
	if err != nil {
		if errors.Is(err, service.ErrAddressForbidden) {
			fail(c, 403, "无权使用该地址")
		} else {
			fail(c, 400, "收货地址不存在，请重新选择")
		}
		return
	}
	res, err := h.Carts.Checkout(claims.UserID, req.ItemIDs, addr)
	if err != nil {
		failCartErr(c, err)
		return
	}
	ok(c, gin.H{"order": res.Order, "total_cents": res.TotalCents})
}

// failCartErr 购物车错误 → HTTP。错误消息本身面向用户可读。
func failCartErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCartNotFound), errors.Is(err, service.ErrNotFound):
		fail(c, 404, err.Error())
	case errors.Is(err, service.ErrCartForbidden):
		fail(c, 403, err.Error())
	case errors.Is(err, service.ErrStockInsufficient), errors.Is(err, service.ErrCartInvalidItem):
		fail(c, 409, err.Error())
	case errors.Is(err, service.ErrCartEmpty), errors.Is(err, service.ErrCartQtyMax),
		errors.Is(err, service.ErrCartFull), errors.Is(err, service.ErrCartBadWork):
		fail(c, 400, err.Error())
	default:
		fail(c, 500, "操作失败，请稍后再试")
	}
}
