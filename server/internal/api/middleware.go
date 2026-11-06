package api

import (
	"net/http"
	"strings"

	"xmf/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthRequired 用户/管理员 JWT 鉴权中间件
func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "请先登录"})
			return
		}
		claims, err := service.ParseToken(secret, strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": err.Error()})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

func RoleRequired(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims").(*service.Claims)
		if claims.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权限"})
			return
		}
		c.Next()
	}
}
