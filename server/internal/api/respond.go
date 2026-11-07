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
	// 供审计中间件取失败原因（不改变响应体）
	c.Set("err_msg", msg)
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
	case errors.Is(err, service.ErrRoleNotFound), errors.Is(err, service.ErrAdminNotFound):
		fail(c, 404, err.Error())
	case errors.Is(err, service.ErrRoleBuiltin), errors.Is(err, service.ErrRoleInUse),
		errors.Is(err, service.ErrLastSuper), errors.Is(err, service.ErrAdminDup),
		errors.Is(err, service.ErrRoleCodeDup), errors.Is(err, service.ErrSelfDisable),
		errors.Is(err, service.ErrSelfDelete):
		fail(c, 409, err.Error())
	case errors.Is(err, service.ErrAdminRoleBad), errors.Is(err, service.ErrRolePermBad),
		errors.Is(err, service.ErrRolePermEmpty), errors.Is(err, service.ErrRoleNameEmpty),
		errors.Is(err, service.ErrAdminPassword), errors.Is(err, service.ErrAdminRoleUnset):
		fail(c, 400, err.Error())
	default:
		fail(c, 400, err.Error())
	}
}
