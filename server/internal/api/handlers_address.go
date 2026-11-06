package api

import (
	"xmf/internal/service"

	"github.com/gin-gonic/gin"
)

// ---- 收货地址（用户端） ----

type addressReq struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Detail    string `json:"detail"`
	IsDefault bool   `json:"is_default"`
}

func (r addressReq) toInput() service.AddressInput {
	return service.AddressInput{
		Name: r.Name, Phone: r.Phone,
		Province: r.Province, City: r.City, District: r.District, Detail: r.Detail,
		IsDefault: r.IsDefault,
	}
}

func (h *Handlers) addressList(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	list, err := h.Addresses.List(claims.UserID)
	if err != nil {
		fail(c, 500, "获取地址失败，请稍后再试")
		return
	}
	ok(c, list)
}

func (h *Handlers) addressCreate(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req addressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	a, err := h.Addresses.Create(claims.UserID, req.toInput())
	if err != nil {
		failAddressErr(c, err)
		return
	}
	ok(c, a)
}

func (h *Handlers) addressUpdate(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req addressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if err := h.Addresses.Update(claims.UserID, paramID(c), req.toInput()); err != nil {
		failAddressErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) addressDelete(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Addresses.Delete(claims.UserID, paramID(c)); err != nil {
		failAddressErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) addressSetDefault(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Addresses.SetDefault(claims.UserID, paramID(c)); err != nil {
		failAddressErr(c, err)
		return
	}
	ok(c, nil)
}

func failAddressErr(c *gin.Context, err error) {
	switch err {
	case service.ErrAddressNotFound:
		fail(c, 404, err.Error())
	case service.ErrAddressForbidden:
		fail(c, 403, err.Error())
	case service.ErrAddressLimit:
		fail(c, 400, err.Error())
	case service.ErrAddressInvalid:
		fail(c, 400, err.Error())
	default:
		fail(c, 400, err.Error())
	}
}
