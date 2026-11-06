package api

import (
	"errors"
	"net/http"

	"xmf/internal/service"

	"github.com/gin-gonic/gin"
)

// 统一响应：{code, msg, data}；msg 必须是人话
func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": data})
}

func fail(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, gin.H{"code": httpCode, "msg": msg, "data": nil})
}

// failErr 把 service 层错误映射为统一的用户可读响应
func failErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		fail(c, 404, "订单不存在")
	case errors.Is(err, service.ErrForbidden):
		fail(c, 403, "无权操作该订单")
	case errors.Is(err, service.ErrIllegalTransition):
		fail(c, 409, "订单当前状态不允许该操作")
	case errors.Is(err, service.ErrStockInsufficient):
		fail(c, 400, err.Error())
	case errors.Is(err, service.ErrEmptyItems):
		fail(c, 400, err.Error())
	default:
		fail(c, 400, err.Error())
	}
}
