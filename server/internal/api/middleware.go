package api

import (
	"net/http"
	"strings"

	"xmf/internal/model"
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

// AdminAuthRequired 管理端二次鉴权：每请求从 DB 取账号+角色，
// 使「禁用账号 / 改角色权限」立即生效（JWT 里只有 id，不带权限快照）。
func AdminAuthRequired(h *Handlers) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims").(*service.Claims)
		a, role, err := h.RBAC.GetAdminWithRole(claims.UserID)
		if err != nil || a == nil || a.Status != 1 || role == nil || role.Status != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "账号已被禁用或角色已停用，请重新登录"})
			return
		}
		c.Set("admin", a)
		c.Set("role", role)
		c.Next()
	}
}

// PermAny 任一权限码满足即放行（如上传接口同时被商品与素材编辑复用）
func PermAny(h *Handlers, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		a := c.MustGet("admin").(*model.Admin)
		role := c.MustGet("role").(*model.Role)
		for _, code := range codes {
			if service.HasPerm(role, code) {
				c.Next()
				return
			}
		}
		h.Audit.Log(service.AuditEntry{
			AdminID: a.ID, AdminName: a.Username, Action: "deny",
			TargetType: "permission", TargetID: strings.Join(codes, "|"),
			Method: c.Request.Method, Path: c.Request.URL.Path,
			IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
			Detail: "越权访问：需要权限 " + strings.Join(codes, " 或 "), Result: "deny",
			StatusCode: http.StatusForbidden,
		})
		c.Set("audit_skip", true) // 已记一条 deny，避免审计中间件重复记一次
		c.Set("audit_skip", true)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权限执行该操作"})
	}
}

// PermRequired 权限码校验。超级管理员直接放行；越权请求写审计日志（deny）。
func PermRequired(h *Handlers, code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		a := c.MustGet("admin").(*model.Admin)
		role := c.MustGet("role").(*model.Role)
		if service.HasPerm(role, code) {
			c.Next()
			return
		}
		h.Audit.Log(service.AuditEntry{
			AdminID: a.ID, AdminName: a.Username, Action: "deny",
			TargetType: "permission", TargetID: code,
			Method: c.Request.Method, Path: c.Request.URL.Path,
			IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
			Detail: "越权访问：需要权限 " + code, Result: "deny",
			StatusCode: http.StatusForbidden,
		})
		c.Set("audit_skip", true)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权限执行该操作"})
	}
}
