package api

import (
	"errors"

	"xmf/internal/model"
	"xmf/internal/service"

	"github.com/gin-gonic/gin"
)

// ---- 「我的」页（v2.0）：资料 + 资产统计 + 订单角标 ----

func (h *Handlers) userMe(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	p, err := h.Profile.Get(claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			fail(c, 404, "用户不存在")
			return
		}
		fail(c, 500, "获取个人资料失败，请稍后再试")
		return
	}
	ok(c, p)
}

type meUpdateReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// userMeUpdate 改昵称/头像。头像先走 POST /upload 拿地址，再回填到这里。
func (h *Handlers) userMeUpdate(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req meUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	u, err := h.Profile.Update(claims.UserID, req.Nickname, req.Avatar)
	if err != nil {
		if errors.Is(err, service.ErrNicknameInvalid) {
			fail(c, 400, err.Error())
			return
		}
		fail(c, 500, "保存失败，请稍后再试")
		return
	}
	ok(c, u)
}

// userMeDeactivate 账号注销：清空作品/购物车/地址、禁用账号并释放 openid。
// 前端拿到成功响应后直接清 token 回登录页。
func (h *Handlers) userMeDeactivate(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if err := h.Profile.Deactivate(claims.UserID); err != nil {
		switch {
		case errors.Is(err, service.ErrAlreadyDeactivated):
			fail(c, 409, "账号已注销")
		case errors.Is(err, service.ErrProfileNotFound):
			fail(c, 404, "用户不存在")
		default:
			fail(c, 500, "注销失败，请稍后再试")
		}
		return
	}
	ok(c, gin.H{"deactivated": true})
}

// workDetailView 作品详情（作品详情页用）：作品 + 作者 + 我是否点过赞。
// 与 GET /works/:id 保持并存：后者是老的扁平结构，已有脚本按它断言。
func (h *Handlers) workDetailView(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	id := paramID(c)
	var w model.Work
	if err := h.DB.First(&w, id).Error; err != nil || w.Status != 1 ||
		(w.UserID != claims.UserID && w.IsPublic != 1) {
		fail(c, 404, "作品不存在")
		return
	}
	author := service.AuthorBrief{ID: w.UserID}
	var u model.User
	if err := h.DB.Select("id", "nickname", "avatar").First(&u, w.UserID).Error; err == nil {
		author.Nickname = u.Nickname
		author.Avatar = u.Avatar
	}
	liked := false
	if w.UserID != claims.UserID {
		var n int64
		h.DB.Model(&model.WorkLike{}).Where("work_id = ? AND user_id = ?", w.ID, claims.UserID).Count(&n)
		liked = n > 0
	}
	ok(c, gin.H{
		"work":    w,
		"author":  author,
		"liked":   liked,
		"is_mine": w.UserID == claims.UserID,
	})
}
