package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"xmf/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 作品社区与积分口径（v1.7，见 docs/PRD.md 3.4）
//
//   - 作品默认私密，作者可发布到广场；发布后 AI 提示词一并公开
//   - 点赞只能点别人的公开作品；点赞是「借用下单」的前置条件
//   - 订单完成时按件数累计作品使用次数，并给作者加等量积分（自用也计）
//   - 积分唯一出口是 AI 生成超出每日免费额度后的抵扣（1 积分 = 1 张）
const (
	PointReasonWorkUsed = "work_used" // 作品被订单使用
	PointReasonAISpend  = "ai_spend"  // AI 生成抵扣

	PointRefOrder  = "order"
	PointRefAITask = "ai_task"

	PublicSizeDefault = 20
	PublicSizeMax     = 100
)

var (
	ErrWorkNotFound    = errors.New("作品不存在")
	ErrWorkForbidden   = errors.New("无权操作该作品")
	ErrWorkNotPublic   = errors.New("作品未发布或已下架")
	ErrWorkUnavailable = errors.New("作品已失效，请重新选择图案")
	ErrSelfLike        = errors.New("不能给自己的作品点赞")
	ErrPointsNotEnough = errors.New("积分不足")
)

// CommunityService 作品广场、点赞、积分
type CommunityService struct {
	DB *gorm.DB
}

// AuthorBrief 广场卡片的作者信息（不带 openid 等敏感字段）
type AuthorBrief struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// PublicWork 广场作品视图
type PublicWork struct {
	ID           int64       `json:"id"`
	Type         string      `json:"type"`
	ImageURL     string      `json:"image_url"`
	ThumbnailURL string      `json:"thumbnail_url"`
	Prompt       string      `json:"prompt"`
	LikeCount    int         `json:"like_count"`
	UseCount     int         `json:"use_count"`
	Liked        bool        `json:"liked"`
	IsMine       bool        `json:"is_mine"`
	PublishedAt  *time.Time  `json:"published_at"`
	Author       AuthorBrief `json:"author"`
}

// SetPublic 发布/取消发布。取消发布只影响广场可见性，点赞/使用计数与已下单不受影响。
func (s *CommunityService) SetPublic(userID, workID int64, isPublic bool) (*model.Work, error) {
	var w model.Work
	if err := s.DB.First(&w, workID).Error; err != nil {
		return nil, ErrWorkNotFound
	}
	if w.UserID != userID {
		return nil, ErrWorkForbidden
	}
	if w.Status != 1 {
		return nil, errors.New("作品已删除或被下架，不能发布")
	}
	updates := map[string]any{"is_public": 0}
	if isPublic {
		now := time.Now()
		updates["is_public"] = 1
		updates["published_at"] = &now
	}
	if err := s.DB.Model(&w).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.DB.First(&w, workID).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

// PublicList 广场列表（排行榜）。
// sortKey：latest 最新榜 / uses 使用量榜 / likes 点赞榜；keyword 匹配公开提示词。
func (s *CommunityService) PublicList(viewerID int64, sortKey, keyword string, page, size int) (int64, []PublicWork, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > PublicSizeMax {
		size = PublicSizeDefault
	}
	q := s.DB.Model(&model.Work{}).Where("is_public = 1 AND status = 1")
	if kw := strings.TrimSpace(keyword); kw != "" {
		q = q.Where("prompt LIKE ?", "%"+kw+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	// 三个榜单都要稳定分页，统一用「主指标 + 发布时间 + id」兜底排序
	order := "COALESCE(published_at, updated_at) DESC, id DESC"
	switch sortKey {
	case "uses":
		order = "use_count DESC, COALESCE(published_at, updated_at) DESC, id DESC"
	case "likes":
		order = "like_count DESC, COALESCE(published_at, updated_at) DESC, id DESC"
	}
	var works []model.Work
	if err := q.Order(order).Offset((page - 1) * size).Limit(size).Find(&works).Error; err != nil {
		return 0, nil, err
	}
	if len(works) == 0 {
		return total, []PublicWork{}, nil
	}

	// 批量取作者信息与「我是否点过赞」，避免逐条查（N+1）
	userIDs := make([]int64, 0, len(works))
	workIDs := make([]int64, 0, len(works))
	for _, w := range works {
		userIDs = append(userIDs, w.UserID)
		workIDs = append(workIDs, w.ID)
	}
	authors := map[int64]model.User{}
	var us []model.User
	if err := s.DB.Where("id IN ?", userIDs).Find(&us).Error; err != nil {
		return 0, nil, err
	}
	for _, u := range us {
		authors[u.ID] = u
	}
	liked := map[int64]bool{}
	var likes []model.WorkLike
	if err := s.DB.Where("user_id = ? AND work_id IN ?", viewerID, workIDs).Find(&likes).Error; err != nil {
		return 0, nil, err
	}
	for _, l := range likes {
		liked[l.WorkID] = true
	}

	items := make([]PublicWork, 0, len(works))
	for _, w := range works {
		a := authors[w.UserID]
		nick := a.Nickname
		if strings.TrimSpace(nick) == "" {
			nick = fmt.Sprintf("用户#%d", a.ID) // 未授权昵称时的兜底展示
		}
		items = append(items, PublicWork{
			ID: w.ID, Type: w.Type,
			ImageURL: w.ImageURL, ThumbnailURL: w.ThumbnailURL,
			Prompt: w.Prompt, LikeCount: w.LikeCount, UseCount: w.UseCount,
			Liked: liked[w.ID], IsMine: w.UserID == viewerID, PublishedAt: w.PublishedAt,
			Author: AuthorBrief{ID: a.ID, Nickname: nick, Avatar: a.Avatar},
		})
	}
	return total, items, nil
}

// LikedWorks 我点赞过的作品（仅仍公开且正常的），供商品页选图
func (s *CommunityService) LikedWorks(userID int64) ([]model.Work, error) {
	var works []model.Work
	err := s.DB.Model(&model.Work{}).
		Joins("JOIN work_likes l ON l.work_id = works.id").
		Where("l.user_id = ? AND works.status = 1 AND works.is_public = 1", userID).
		Order("l.id DESC").Limit(200).Find(&works).Error
	if err != nil {
		return nil, err
	}
	return works, nil
}

// Like 点赞（幂等）。只有公开且正常的作品可点赞，自己的作品不可点赞。
func (s *CommunityService) Like(userID, workID int64) (int, error) {
	var w model.Work
	if err := s.DB.First(&w, workID).Error; err != nil {
		return 0, ErrWorkNotFound
	}
	if w.Status != 1 || w.IsPublic != 1 {
		return 0, ErrWorkNotPublic
	}
	if w.UserID == userID {
		return 0, ErrSelfLike
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&model.WorkLike{WorkID: workID, UserID: userID})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 { // 已点过赞，保持幂等
			return nil
		}
		return tx.Model(&model.Work{}).Where("id = ?", workID).
			UpdateColumn("like_count", gorm.Expr("COALESCE(like_count, 0) + 1")).Error
	})
	if err != nil {
		return 0, err
	}
	return s.likeCount(workID), nil
}

// Unlike 取消点赞（幂等），计数不会减到负数
func (s *CommunityService) Unlike(userID, workID int64) (int, error) {
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("work_id = ? AND user_id = ?", workID, userID).Delete(&model.WorkLike{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&model.Work{}).Where("id = ?", workID).
			UpdateColumn("like_count", gorm.Expr("CASE WHEN COALESCE(like_count, 0) > 0 THEN COALESCE(like_count, 0) - 1 ELSE 0 END")).Error
	})
	if err != nil {
		return 0, err
	}
	return s.likeCount(workID), nil
}

func (s *CommunityService) likeCount(workID int64) int {
	var n int
	s.DB.Model(&model.Work{}).Where("id = ?", workID).Select("like_count").Scan(&n)
	return n
}

// UsableWork 下单可选的作品：自己的，或已点赞且仍公开的他人作品。
// 返回作品本身，调用方可继续判空/取作者。
func (s *CommunityService) UsableWork(userID, workID int64) (*model.Work, error) {
	var w model.Work
	if err := s.DB.First(&w, workID).Error; err != nil {
		return nil, ErrWorkUnavailable
	}
	if w.Status != 1 {
		return nil, ErrWorkUnavailable
	}
	if w.UserID == userID {
		return &w, nil
	}
	if w.IsPublic != 1 {
		return nil, ErrWorkUnavailable
	}
	var n int64
	if err := s.DB.Model(&model.WorkLike{}).Where("work_id = ? AND user_id = ?", workID, userID).Count(&n).Error; err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrWorkUnavailable
	}
	return &w, nil
}

// Balance 积分余额
func (s *CommunityService) Balance(userID int64) (int, error) {
	var n int
	err := s.DB.Model(&model.User{}).Where("id = ?", userID).Select("points").Scan(&n).Error
	return n, err
}

// Logs 最近积分流水
func (s *CommunityService) Logs(userID int64, limit int) ([]model.PointLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var logs []model.PointLog
	err := s.DB.Where("user_id = ?", userID).Order("id desc").Limit(limit).Find(&logs).Error
	return logs, err
}

// TodayQuota 今日免费额度已用次数。只统计 pay_type=free 的任务，
// 用积分换的图不占免费次数（否则「花积分」反而更亏）。
func (s *CommunityService) TodayQuota(userID int64) (used int, err error) {
	today := time.Now().Format("2006-01-02")
	var n int64
	err = s.DB.Model(&model.AITask{}).
		Where("user_id = ? AND quota_date = ? AND pay_type <> ?", userID, today, "point").
		Count(&n).Error
	return int(n), err
}

// AddPoints 加/减积分并写流水，必须在事务内调用，保证余额与流水一致。
// 扣减走条件更新（points >= 扣减量），并发下不会扣成负数。
func AddPoints(tx *gorm.DB, userID int64, delta int, reason, refType string, refID int64, remark string) error {
	if delta == 0 {
		return nil
	}
	var res *gorm.DB
	if delta > 0 {
		res = tx.Exec("UPDATE users SET points = COALESCE(points, 0) + ? WHERE id = ?", delta, userID)
	} else {
		res = tx.Exec("UPDATE users SET points = COALESCE(points, 0) + ? WHERE id = ? AND COALESCE(points, 0) >= ?", delta, userID, -delta)
	}
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrPointsNotEnough
	}
	var balance int
	if err := tx.Raw("SELECT points FROM users WHERE id = ?", userID).Scan(&balance).Error; err != nil {
		return err
	}
	return tx.Create(&model.PointLog{
		UserID: userID, Delta: delta, Balance: balance,
		Reason: reason, RefType: refType, RefID: refID, Remark: remark,
	}).Error
}

// SettleOrderUsage 订单完成结算：明细里的作品按件数累计使用次数，
// 并给作者加等量积分（自用也计，口径经产品确认）。由订单状态机在 shipped→done 事务内调用，
// 状态机保证同一订单只会结算一次。
func SettleOrderUsage(tx *gorm.DB, order *model.Order) error {
	var items []model.OrderItem
	if err := tx.Where("order_id = ? AND work_id > 0", order.ID).Find(&items).Error; err != nil {
		return err
	}
	for _, it := range items {
		if it.WorkID == nil || *it.WorkID <= 0 {
			continue
		}
		var w model.Work
		if err := tx.First(&w, *it.WorkID).Error; err != nil {
			continue // 作品已被物理清理时跳过，不影响订单完成
		}
		qty := it.Qty
		if qty <= 0 {
			qty = 1
		}
		if err := tx.Model(&model.Work{}).Where("id = ?", w.ID).
			UpdateColumn("use_count", gorm.Expr("COALESCE(use_count, 0) + ?", qty)).Error; err != nil {
			return err
		}
		remark := fmt.Sprintf("作品《%s》被使用 %d 件（订单 %s）", workLabel(&w), qty, order.OrderNo)
		if err := AddPoints(tx, w.UserID, qty, PointReasonWorkUsed, PointRefOrder, order.ID, remark); err != nil {
			return err
		}
	}
	return nil
}

// workLabel 作品可读名：优先取公开提示词，截断避免流水过长
func workLabel(w *model.Work) string {
	if s := strings.TrimSpace(w.Prompt); s != "" {
		r := []rune(s)
		if len(r) > 12 {
			return string(r[:12]) + "…"
		}
		return s
	}
	return fmt.Sprintf("#%d", w.ID)
}
