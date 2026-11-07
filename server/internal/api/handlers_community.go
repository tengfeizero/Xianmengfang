package api

import (
	"errors"
	"strconv"

	"xmf/internal/service"

	"github.com/gin-gonic/gin"
)

// ---- 作品社区（v1.7）：发布 / 广场 / 点赞 / 积分 ----

// failCommunity 把社区相关 service 错误映射成用户可读响应
func failCommunity(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWorkNotFound):
		fail(c, 404, "作品不存在")
	case errors.Is(err, service.ErrWorkForbidden):
		fail(c, 403, "无权操作该作品")
	case errors.Is(err, service.ErrWorkNotPublic):
		fail(c, 404, "作品未发布或已下架")
	case errors.Is(err, service.ErrSelfLike):
		fail(c, 400, err.Error())
	case errors.Is(err, service.ErrWorkUnavailable):
		fail(c, 400, err.Error())
	case errors.Is(err, service.ErrPointsNotEnough):
		fail(c, 400, "积分不足")
	default:
		fail(c, 400, err.Error())
	}
}

type publishReq struct {
	IsPublic *int `json:"is_public" binding:"required"`
}

// workPublish 发布/取消发布到广场（发布后 AI 提示词一并公开）
func (h *Handlers) workPublish(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req publishReq
	if err := c.ShouldBindJSON(&req); err != nil || req.IsPublic == nil {
		fail(c, 400, "参数错误：is_public 只能为 0 或 1")
		return
	}
	if *req.IsPublic != 0 && *req.IsPublic != 1 {
		fail(c, 400, "参数错误：is_public 只能为 0 或 1")
		return
	}
	w, err := h.Community.SetPublic(claims.UserID, paramID(c), *req.IsPublic == 1)
	if err != nil {
		failCommunity(c, err)
		return
	}
	ok(c, w)
}

// workPublicList 作品广场 / 排行榜：sort=latest|uses|likes。
// 路径用 /plaza 而不是 /works/public：gin 同段静态路径与 /works/:id 的通配符冲突。
func (h *Handlers) workPublicList(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	sort := c.Query("sort")
	total, items, err := h.Community.PublicList(claims.UserID, sort, c.Query("keyword"), page, size)
	if err != nil {
		fail(c, 500, "获取作品广场失败，请稍后再试")
		return
	}
	ok(c, gin.H{"total": total, "items": items})
}

// workLike 点赞（幂等）
func (h *Handlers) workLike(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	n, err := h.Community.Like(claims.UserID, paramID(c))
	if err != nil {
		failCommunity(c, err)
		return
	}
	ok(c, gin.H{"like_count": n, "liked": true})
}

// workUnlike 取消点赞（幂等）
func (h *Handlers) workUnlike(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	n, err := h.Community.Unlike(claims.UserID, paramID(c))
	if err != nil {
		failCommunity(c, err)
		return
	}
	ok(c, gin.H{"like_count": n, "liked": false})
}

// userPoints 我的积分与流水。默认返回最近 50 条（作品页角标用），
// 带 page 参数时按页返回（积分明细页），并附 total。
func (h *Handlers) userPoints(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	balance, err := h.Community.Balance(claims.UserID)
	if err != nil {
		fail(c, 500, "获取积分失败，请稍后再试")
		return
	}
	if c.Query("page") == "" {
		logs, err := h.Community.Logs(claims.UserID, 50)
		if err != nil {
			fail(c, 500, "获取积分失败，请稍后再试")
			return
		}
		ok(c, gin.H{"points": balance, "logs": logs})
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	total, logs, err := h.Community.LogsPage(claims.UserID, page, size)
	if err != nil {
		fail(c, 500, "获取积分失败，请稍后再试")
		return
	}
	ok(c, gin.H{"points": balance, "total": total, "logs": logs})
}

// userAIQuota 今日免费次数与积分余额
func (h *Handlers) userAIQuota(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	limit, used, points, err := h.AI.Quota(claims.UserID)
	if err != nil {
		fail(c, 500, "获取生成额度失败，请稍后再试")
		return
	}
	left := limit - used
	if left < 0 {
		left = 0
	}
	ok(c, gin.H{
		"free_limit": limit, "free_used": used, "free_left": left,
		"points": points, "can_generate": left > 0 || points > 0,
	})
}
