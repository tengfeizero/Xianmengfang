package api

import (
	"strconv"

	"xmf/internal/model"
	"xmf/internal/service"

	"github.com/gin-gonic/gin"
)

// ---- 当前登录管理员 ----

// adminMe 登录态自检：前端据此渲染菜单与按钮（权限以后端为准）
func (h *Handlers) adminMe(c *gin.Context) {
	a := c.MustGet("admin").(*model.Admin)
	role := c.MustGet("role").(*model.Role)
	ok(c, gin.H{
		"id":       a.ID,
		"username": a.Username,
		"nickname": a.Nickname,
		"is_super": service.IsSuper(role),
		"role":     gin.H{"id": role.ID, "code": role.Code, "name": role.Name},
		"perms":    rolePermsForFrontend(role),
	})
}

// rolePermsForFrontend 超级管理员下发通配符，前端 hasPerm 认 "*"
func rolePermsForFrontend(role *model.Role) []string {
	if service.IsSuper(role) {
		return []string{"*"}
	}
	return service.ParsePerms(role.Perms)
}

// ---- 权限清单 ----

func (h *Handlers) adminPermCatalog(c *gin.Context) {
	ok(c, service.PermGroups)
}

// ---- 角色 ----

func (h *Handlers) adminRoleList(c *gin.Context) {
	roles, err := h.RBAC.ListRoles()
	if err != nil {
		fail(c, 500, "获取角色失败，请稍后再试")
		return
	}
	for i := range roles {
		roles[i].Perms = rolePermsForFrontend(&roles[i].Role)
	}
	ok(c, roles)
}

type roleReq struct {
	Name   string   `json:"name"`
	Remark string   `json:"remark"`
	Perms  []string `json:"perms"`
	Status int      `json:"status"`
}

func (h *Handlers) adminRoleCreate(c *gin.Context) {
	var req roleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	r, err := h.RBAC.CreateRole(req.Name, req.Remark, req.Perms)
	if err != nil {
		failErr(c, err)
		return
	}
	ok(c, r)
}

func (h *Handlers) adminRoleUpdate(c *gin.Context) {
	var req roleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if err := h.RBAC.UpdateRole(paramID(c), req.Name, req.Remark, req.Perms, req.Status); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminRoleDelete(c *gin.Context) {
	if err := h.RBAC.DeleteRole(paramID(c)); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

// ---- 管理员账号 ----

func (h *Handlers) adminAdminList(c *gin.Context) {
	list, err := h.RBAC.ListAdmins()
	if err != nil {
		fail(c, 500, "获取管理员失败，请稍后再试")
		return
	}
	ok(c, list)
}

type adminCreateReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	RoleID   int64  `json:"role_id"`
}

func (h *Handlers) adminAdminCreate(c *gin.Context) {
	var req adminCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if req.RoleID == 0 {
		failErr(c, service.ErrAdminRoleUnset)
		return
	}
	a, err := h.RBAC.CreateAdmin(req.Username, req.Password, req.Nickname, req.RoleID)
	if err != nil {
		failErr(c, err)
		return
	}
	ok(c, gin.H{"id": a.ID})
}

type adminUpdateReq struct {
	Nickname string `json:"nickname"`
	RoleID   int64  `json:"role_id"`
	Status   int    `json:"status"`
	Password string `json:"password"`
}

func (h *Handlers) adminAdminUpdate(c *gin.Context) {
	me := c.MustGet("admin").(*model.Admin)
	var req adminUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if req.RoleID == 0 {
		failErr(c, service.ErrAdminRoleUnset)
		return
	}
	if err := h.RBAC.UpdateAdmin(me.ID, paramID(c), req.Nickname, req.RoleID, req.Status, req.Password); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminAdminDelete(c *gin.Context) {
	me := c.MustGet("admin").(*model.Admin)
	if err := h.RBAC.DeleteAdmin(me.ID, paramID(c)); err != nil {
		failErr(c, err)
		return
	}
	ok(c, nil)
}

// ---- 审计日志 ----

func (h *Handlers) adminAuditLogs(c *gin.Context) {
	q := service.AuditQuery{
		Action:  c.Query("action"),
		Result:  c.Query("result"),
		Keyword: c.Query("keyword"),
	}
	if v := c.Query("admin_id"); v != "" {
		q.AdminID, _ = strconv.ParseInt(v, 10, 64)
	}
	q.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	q.Size, _ = strconv.Atoi(c.DefaultQuery("size", "10"))
	page, err := h.Audit.Query(q)
	if err != nil {
		fail(c, 500, "获取日志失败，请稍后再试")
		return
	}
	ok(c, page)
}

// ---- 登录留痕 ----

func (h *Handlers) auditLogin(c *gin.Context, adminID int64, name, result, msg string, code int) {
	h.Audit.Log(service.AuditEntry{
		AdminID: adminID, AdminName: name, Action: "auth.login",
		TargetType: "admin", TargetID: strconv.FormatInt(adminID, 10),
		Method: "POST", Path: c.Request.URL.Path,
		IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
		Detail: msg, Result: result, StatusCode: code,
	})
}
