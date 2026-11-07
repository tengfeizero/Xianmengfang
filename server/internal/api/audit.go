package api

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"xmf/internal/model"
	"xmf/internal/service"

	"github.com/gin-gonic/gin"
)

// auditMeta 路由 → 动作/目标类型的映射（写操作才需要登记）
type auditMeta struct {
	Action string
	Target string
}

// routeAudit 用 gin 的路由模板（c.FullPath()）作 key，避免手写 URL 解析
var routeAudit = map[string]auditMeta{
	"POST /api/v1/admin/orders/:id/accept":       {"order.accept", "order"},
	"POST /api/v1/admin/orders/:id/pack":         {"order.pack", "order"},
	"POST /api/v1/admin/orders/:id/ready_ship":   {"order.ready_ship", "order"},
	"POST /api/v1/admin/orders/:id/ship":         {"order.ship", "order"},
	"POST /api/v1/admin/orders/:id/cancel":       {"order.cancel", "order"},
	"POST /api/v1/admin/orders/:id/refund_mark":  {"order.refund", "order"},
	"POST /api/v1/admin/boards":                  {"asset.board.create", "board"},
	"PUT /api/v1/admin/boards/:id":               {"asset.board.update", "board"},
	"DELETE /api/v1/admin/boards/:id":            {"asset.board.delete", "board"},
	"POST /api/v1/admin/elements":                {"asset.element.create", "element"},
	"PUT /api/v1/admin/elements/:id":             {"asset.element.update", "element"},
	"DELETE /api/v1/admin/elements/:id":          {"asset.element.delete", "element"},
	"POST /api/v1/admin/upload":                  {"asset.upload", "file"},
	"POST /api/v1/admin/categories":              {"product.category.create", "category"},
	"PUT /api/v1/admin/categories/:id":           {"product.category.update", "category"},
	"DELETE /api/v1/admin/categories/:id":        {"product.category.delete", "category"},
	"POST /api/v1/admin/products":                {"product.create", "product"},
	"PUT /api/v1/admin/products/:id":             {"product.update", "product"},
	"DELETE /api/v1/admin/products/:id":          {"product.delete", "product"},
	"POST /api/v1/admin/skus":                    {"product.sku.create", "sku"},
	"PUT /api/v1/admin/skus/:id":                 {"product.sku.update", "sku"},
	"DELETE /api/v1/admin/skus/:id":              {"product.sku.delete", "sku"},
	"PUT /api/v1/admin/configs":                  {"config.update", "config"},
	"PUT /api/v1/admin/users/:id/status":         {"user.status", "user"},
	"PUT /api/v1/admin/works/:id/status":         {"work.status", "work"},
	"POST /api/v1/admin/admins":                  {"admin.create", "admin"},
	"PUT /api/v1/admin/admins/:id":               {"admin.update", "admin"},
	"DELETE /api/v1/admin/admins/:id":            {"admin.delete", "admin"},
	"POST /api/v1/admin/roles":                   {"role.create", "role"},
	"PUT /api/v1/admin/roles/:id":                {"role.update", "role"},
	"DELETE /api/v1/admin/roles/:id":             {"role.delete", "role"},
}

// AuditWrite 敏感操作审计中间件：管理端写方法全量留痕。
// 请求体先读后还原，脱敏交给 service.MaskDetail（密码/密钥不会落库）。
func AuditWrite(h *Handlers) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}

		var body []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength <= 64*1024 {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		c.Next()

		if skip, exists := c.Get("audit_skip"); exists && skip == true {
			return
		}
		a, okAdmin := c.Get("admin")
		if !okAdmin {
			return
		}
		admin := a.(*model.Admin)

		meta, found := routeAudit[method+" "+c.FullPath()]
		action, target := meta.Action, meta.Target
		if !found {
			// 兜底：未登记的写接口也要留痕
			action = "admin." + strings.ToLower(method)
			target = "unknown"
		}

		result := "ok"
		code := c.Writer.Status()
		if code >= 400 {
			result = "fail"
			if code == http.StatusForbidden {
				result = "deny"
			}
		}
		errMsg, _ := c.Get("err_msg")
		errStr, _ := errMsg.(string)

		detail := service.MaskDetail(body)
		if strings.HasPrefix(c.ContentType(), "multipart/") {
			detail = "<文件上传>"
		}

		h.Audit.Log(service.AuditEntry{
			AdminID:    admin.ID,
			AdminName:  admin.Username,
			Action:     action,
			TargetType: target,
			TargetID:   c.Param("id"),
			Method:     method,
			Path:       c.Request.URL.Path,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Detail:     detail,
			Result:     result,
			StatusCode: code,
			ErrMsg:     errStr,
		})
	}
}
