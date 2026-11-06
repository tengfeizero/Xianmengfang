package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func paramID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	return id
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
		ag.PUT("/boards/:id", h.adminBoardUpdate)
		ag.DELETE("/boards/:id", h.adminBoardDelete)
		ag.POST("/elements", h.adminElementCreate)
		ag.PUT("/elements/:id", h.adminElementUpdate)
		ag.DELETE("/elements/:id", h.adminElementDelete)
	}

	return r
}
