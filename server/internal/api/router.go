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
	{
		// 登录
		v1.POST("/auth/login", h.wxLogin)
		v1.POST("/auth/dev_login", h.devLogin)
		v1.POST("/admin/auth/login", h.adminLogin)

		// 商品（公开）
		v1.GET("/products", h.productList)
		v1.GET("/products/:id", h.productDetail)
	}

	// 用户端（需登录）
	ug := v1.Group("", AuthRequired(h.Secret), RoleRequired("user"))
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
	}

	// 管理后台（需管理员）
	ag := v1.Group("/admin", AuthRequired(h.Secret), RoleRequired("admin"))
	{
		ag.GET("/orders", h.adminOrderList)
		ag.POST("/orders/:id/accept", h.adminOrderAccept)
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
	}

	return r
}
