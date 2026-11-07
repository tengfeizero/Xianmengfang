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
		ug.DELETE("/works/:id", h.workDelete)

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

	// 管理后台（需管理员）
	ag := v1.Group("/admin", AuthRequired(h.Secret), RoleRequired("admin"))
	{
		ag.GET("/orders", h.adminOrderList)
		ag.GET("/orders/:id", h.adminOrderDetail)
		ag.POST("/orders/:id/accept", h.adminOrderAccept)
		ag.POST("/orders/:id/pack", h.adminOrderPack)
		ag.POST("/orders/:id/ready_ship", h.adminOrderReadyShip)
		ag.POST("/orders/:id/ship", h.adminOrderShip)
		ag.POST("/orders/:id/cancel", h.adminOrderCancel)
		ag.POST("/orders/:id/refund_mark", h.adminOrderRefundMark)

		// 素材与底板管理（M1）
		ag.POST("/boards", h.adminBoardCreate)
		ag.GET("/boards", h.adminBoardList)
		ag.PUT("/boards/:id", h.adminBoardUpdate)
		ag.DELETE("/boards/:id", h.adminBoardDelete)
		ag.POST("/elements", h.adminElementCreate)
		ag.GET("/elements", h.adminElementList)
		ag.PUT("/elements/:id", h.adminElementUpdate)
		ag.DELETE("/elements/:id", h.adminElementDelete)

		// 商品/分类/SKU 管理 + 系统配置（M1.4）
		ag.POST("/upload", h.upload)
		ag.GET("/categories", h.adminCategoryList)
		ag.POST("/categories", h.adminCategoryCreate)
		ag.PUT("/categories/:id", h.adminCategoryUpdate)
		ag.DELETE("/categories/:id", h.adminCategoryDelete)
		ag.GET("/products", h.adminProductList)
		ag.POST("/products", h.adminProductCreate)
		ag.PUT("/products/:id", h.adminProductUpdate)
		ag.DELETE("/products/:id", h.adminProductDelete)
		ag.POST("/skus", h.adminSKUCreate)
		ag.PUT("/skus/:id", h.adminSKUUpdate)
		ag.DELETE("/skus/:id", h.adminSKUDelete)
		ag.GET("/configs", h.adminConfigList)
		ag.PUT("/configs", h.adminConfigUpdate)
		ag.GET("/users", h.adminUserList)
		ag.PUT("/users/:id/status", h.adminUserSetStatus)
		ag.GET("/works", h.adminWorkList)
		ag.PUT("/works/:id/status", h.adminWorkSetStatus)
	}

	return r
}
