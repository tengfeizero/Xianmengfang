package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

func paramID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	return id
}

// ServeWebDir 托管 Web 后台静态文件（SPA，hash 路由；未匹配文件回落 index.html）
func ServeWebDir(r *gin.Engine, dir string) {
	r.GET("/ui", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/ui/") })
	r.GET("/ui/*filepath", func(c *gin.Context) {
		p := filepath.Join(dir, filepath.Clean(c.Param("filepath")))
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			c.File(p)
			return
		}
		c.File(filepath.Join(dir, "index.html"))
	})
}

// Router 组装全部路由
func Router(h *Handlers) *gin.Engine {
	r := gin.Default()

	// 静态资源（上传图片与生成物）
	r.Static("/static", h.UploadsDir)

	v1 := r.Group("/api/v1")

	// 用户端（小程序）：全部收在 /user 下，与 /admin 对称。
	// 公开接口（登录、商品浏览）也在 /user 内，保证"一个端一个前缀"。
	user := v1.Group("/user")
	{
		// 登录
		user.POST("/auth/login", h.wxLogin)
		user.POST("/auth/dev_login", h.devLogin)

		// 商品（公开）
		user.GET("/products", h.productList)
		user.GET("/products/:id", h.productDetail)
	}

	// 用户端（需登录）
	ug := user.Group("", AuthRequired(h.Secret), RoleRequired("user"))
	{
		ug.POST("/orders", h.orderCreate)
		ug.GET("/orders", h.orderList)
		ug.GET("/orders/:id", h.orderDetail)
		ug.POST("/orders/:id/confirm", h.orderConfirm)
		ug.POST("/orders/:id/cancel", h.orderCancel)

		// 上传与创作（M1）
		ug.POST("/upload", h.upload)
		ug.GET("/collage/boards", h.collageBoards)
		ug.GET("/collage/elements", h.collageElements)
		ug.POST("/collage/compose", h.collageCompose)
		ug.POST("/ai/tasks", h.aiCreateTask)

		// 作品库（M1）
		ug.GET("/works", h.workList)
		ug.GET("/works/:id", h.workDetail)
		ug.GET("/works/:id/view", h.workDetailView)
		ug.DELETE("/works/:id", h.workDelete)

		// 个人中心（v2.0）：「我的」页资料/资产/订单角标一次取全
		ug.GET("/me", h.userMe)
		ug.PUT("/me", h.userMeUpdate)
		ug.DELETE("/me", h.userMeDeactivate) // v2.1 账号注销

		// 作品社区与积分（v1.7）
		ug.PUT("/works/:id/publish", h.workPublish)
		ug.POST("/works/:id/like", h.workLike)
		ug.DELETE("/works/:id/like", h.workUnlike)
		ug.GET("/plaza", h.workPublicList)
		ug.GET("/points", h.userPoints)
		ug.GET("/ai/quota", h.userAIQuota)

		// 收货地址（M1.5）
		ug.GET("/addresses", h.addressList)
		ug.POST("/addresses", h.addressCreate)
		ug.PUT("/addresses/:id", h.addressUpdate)
		ug.DELETE("/addresses/:id", h.addressDelete)
		ug.POST("/addresses/:id/default", h.addressSetDefault)

		// 购物车（v1.0）：PUT /cart 是批量勾选，PUT /cart/:id 是改数量/勾选
		ug.GET("/cart", h.cartList)
		ug.POST("/cart", h.cartAdd)
		ug.PUT("/cart", h.cartSetChecked)
		ug.PUT("/cart/:id", h.cartUpdate)
		ug.DELETE("/cart/:id", h.cartRemove)
		ug.DELETE("/cart", h.cartClear)
		ug.POST("/cart/checkout", h.cartCheckout)
	}

	// 管理后台：登录公开，其余需管理员
	v1.POST("/admin/auth/login", h.adminLogin)

	// 管理后台（需管理员）：AuthRequired 验 JWT → RoleRequired 验端 → AdminAuthRequired 载入账号+角色
	// → AuditWrite 对写操作留痕；每条路由再按权限码精确拦截。
	ag := v1.Group("/admin",
		AuthRequired(h.Secret), RoleRequired("admin"), AdminAuthRequired(h), AuditWrite(h))
	{
		// 登录态自检与权限清单
		ag.GET("/me", h.adminMe)
		ag.GET("/permissions", h.adminPermCatalog)

		ag.GET("/orders", PermRequired(h, "order:view"), h.adminOrderList)
		ag.GET("/orders/:id", PermRequired(h, "order:view"), h.adminOrderDetail)
		ag.POST("/orders/:id/accept", PermRequired(h, "order:process"), h.adminOrderAccept)
		ag.POST("/orders/:id/pack", PermRequired(h, "order:process"), h.adminOrderPack)
		ag.POST("/orders/:id/ready_ship", PermRequired(h, "order:process"), h.adminOrderReadyShip)
		ag.POST("/orders/:id/ship", PermRequired(h, "order:ship"), h.adminOrderShip)
		ag.POST("/orders/:id/cancel", PermRequired(h, "order:cancel"), h.adminOrderCancel)
		ag.POST("/orders/:id/refund_mark", PermRequired(h, "order:refund"), h.adminOrderRefundMark)

		// 素材与底板管理（M1）
		ag.GET("/boards", PermRequired(h, "asset:view"), h.adminBoardList)
		ag.POST("/boards", PermRequired(h, "asset:edit"), h.adminBoardCreate)
		ag.PUT("/boards/:id", PermRequired(h, "asset:edit"), h.adminBoardUpdate)
		ag.DELETE("/boards/:id", PermRequired(h, "asset:edit"), h.adminBoardDelete)
		ag.GET("/elements", PermRequired(h, "asset:view"), h.adminElementList)
		ag.POST("/elements", PermRequired(h, "asset:edit"), h.adminElementCreate)
		ag.PUT("/elements/:id", PermRequired(h, "asset:edit"), h.adminElementUpdate)
		ag.DELETE("/elements/:id", PermRequired(h, "asset:edit"), h.adminElementDelete)

		// 商品/分类/SKU 管理 + 系统配置（M1.4）
		ag.POST("/upload", PermAny(h, "product:edit", "asset:edit"), h.upload)
		ag.GET("/categories", PermRequired(h, "product:view"), h.adminCategoryList)
		ag.POST("/categories", PermRequired(h, "product:edit"), h.adminCategoryCreate)
		ag.PUT("/categories/:id", PermRequired(h, "product:edit"), h.adminCategoryUpdate)
		ag.DELETE("/categories/:id", PermRequired(h, "product:edit"), h.adminCategoryDelete)
		ag.GET("/products", PermRequired(h, "product:view"), h.adminProductList)
		ag.POST("/products", PermRequired(h, "product:edit"), h.adminProductCreate)
		ag.PUT("/products/:id", PermRequired(h, "product:edit"), h.adminProductUpdate)
		ag.DELETE("/products/:id", PermRequired(h, "product:edit"), h.adminProductDelete)
		ag.POST("/skus", PermRequired(h, "product:edit"), h.adminSKUCreate)
		ag.PUT("/skus/:id", PermRequired(h, "product:edit"), h.adminSKUUpdate)
		ag.DELETE("/skus/:id", PermRequired(h, "product:edit"), h.adminSKUDelete)

		// 配置、用户、作品
		ag.GET("/configs", PermRequired(h, "config:view"), h.adminConfigList)
		ag.PUT("/configs", PermRequired(h, "config:edit"), h.adminConfigUpdate)
		ag.GET("/users", PermRequired(h, "user:view"), h.adminUserList)
		ag.PUT("/users/:id/status", PermRequired(h, "user:edit"), h.adminUserSetStatus)
		ag.GET("/works", PermRequired(h, "work:view"), h.adminWorkList)
		ag.PUT("/works/:id/status", PermRequired(h, "work:edit"), h.adminWorkSetStatus)

		// 明细（v2.3）：积分流水 / 点赞 / 使用，权限复用用户与作品的查看权限
		ag.GET("/point-logs", PermRequired(h, "user:view"), h.adminPointLogs)
		ag.GET("/work-likes", PermRequired(h, "work:view"), h.adminWorkLikes)
		ag.GET("/work-usages", PermRequired(h, "work:view"), h.adminWorkUsages)

		// RBAC：角色 / 管理员 / 审计日志
		ag.GET("/roles", PermRequired(h, "role:view"), h.adminRoleList)
		ag.POST("/roles", PermRequired(h, "role:edit"), h.adminRoleCreate)
		ag.PUT("/roles/:id", PermRequired(h, "role:edit"), h.adminRoleUpdate)
		ag.DELETE("/roles/:id", PermRequired(h, "role:edit"), h.adminRoleDelete)
		ag.GET("/admins", PermRequired(h, "admin:view"), h.adminAdminList)
		ag.POST("/admins", PermRequired(h, "admin:edit"), h.adminAdminCreate)
		ag.PUT("/admins/:id", PermRequired(h, "admin:edit"), h.adminAdminUpdate)
		ag.DELETE("/admins/:id", PermRequired(h, "admin:edit"), h.adminAdminDelete)
		ag.GET("/audit-logs", PermRequired(h, "audit:view"), h.adminAuditLogs)
	}

	return r
}
