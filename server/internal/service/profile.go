package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"xmf/internal/model"

	"gorm.io/gorm"
)

// 「我的」页聚合（v2.0）。设计目标：小程序首屏一个请求拿全，
// 避免头像卡、积分、订单角标、购物车角标各自打一个接口造成抖动。
type ProfileService struct {
	DB *gorm.DB
}

var (
	ErrProfileNotFound    = errors.New("用户不存在")
	ErrNicknameInvalid    = errors.New("昵称需 1~16 个字")
	ErrAlreadyDeactivated = errors.New("账号已注销")
)

// OrderStats 各状态订单数（只统计在途 + 已完成，取消不进「我的」入口）
type OrderStats struct {
	PendingAccept int `json:"pending_accept"`
	Producing     int `json:"producing"`
	Packing       int `json:"packing"`
	PendingShip   int `json:"pending_ship"`
	Shipped       int `json:"shipped"`
	Done          int `json:"done"`
}

// ProfileStats 用户资产与内容规模
type ProfileStats struct {
	Works       int        `json:"works"`        // 我的作品（正常状态）
	PublicWorks int        `json:"public_works"` // 其中已发布到广场
	Liked       int        `json:"liked"`        // 我点赞的作品数
	CartCount   int        `json:"cart_count"`   // 购物车条目数
	Orders      OrderStats `json:"orders"`
}

type Profile struct {
	User  model.User   `json:"user"`
	Stats ProfileStats `json:"stats"`
}

// Get 读取用户资料与统计
func (s *ProfileService) Get(userID int64) (*Profile, error) {
	var u model.User
	if err := s.DB.First(&u, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	var st ProfileStats
	var n int64
	s.DB.Model(&model.Work{}).
		Where("user_id = ? AND status = 1", userID).Count(&n)
	st.Works = int(n)
	s.DB.Model(&model.Work{}).
		Where("user_id = ? AND status = 1 AND is_public = 1", userID).Count(&n)
	st.PublicWorks = int(n)
	s.DB.Model(&model.WorkLike{}).Where("user_id = ?", userID).Count(&n)
	st.Liked = int(n)
	s.DB.Model(&model.CartItem{}).Where("user_id = ?", userID).Count(&n)
	st.CartCount = int(n)

	type row struct {
		Status string
		N      int
	}
	var rows []row
	s.DB.Model(&model.Order{}).
		Select("status, count(*) as n").
		Where("user_id = ?", userID).
		Group("status").Scan(&rows)
	for _, r := range rows {
		switch r.Status {
		case "pending_accept":
			st.Orders.PendingAccept = r.N
		case "producing":
			st.Orders.Producing = r.N
		case "packing":
			st.Orders.Packing = r.N
		case "pending_ship":
			st.Orders.PendingShip = r.N
		case "shipped":
			st.Orders.Shipped = r.N
		case "done":
			st.Orders.Done = r.N
		}
	}
	// 生产/打包/待发货对用户是同一件事：都在商家手里，进度条拍平到「进行中」
	return &Profile{User: u, Stats: st}, nil
}

// Update 修改昵称/头像。传空串表示不改（避免误清空）。
func (s *ProfileService) Update(userID int64, nickname, avatar string) (*model.User, error) {
	var u model.User
	if err := s.DB.First(&u, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	updates := map[string]any{}
	if nickname != "" {
		nickname = strings.TrimSpace(nickname)
		if n := len([]rune(nickname)); n < 1 || n > 16 {
			return nil, ErrNicknameInvalid
		}
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if len(updates) == 0 {
		return &u, nil
	}
	if err := s.DB.Model(&u).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.DB.First(&u, userID).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// LogsPage 积分流水分页（积分明细页用）
func (s *CommunityService) LogsPage(userID int64, page, size int) (int64, []model.PointLog, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 50 {
		size = 20
	}
	var total int64
	if err := s.DB.Model(&model.PointLog{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var logs []model.PointLog
	err := s.DB.Where("user_id = ?", userID).
		Order("id desc").Offset((page - 1) * size).Limit(size).Find(&logs).Error
	return total, logs, err
}

// Deactivate 账号注销（v2.1）。
//
// 语义：用户主动「删除账号」。行不物理删除（订单、积分流水要与财务/售后对得上），
// 而是打上 deactivated_at、禁用、清资料，并把 open_id 换成 tombstone —— 唯一键腾出来后，
// 同一微信再次登录会拿到一个全新的空账号，不用人工介入解绑。
//
// 清理范围：作品全部下架并软删（广场不可见）、购物车清空、收货地址软删。
// 保留：订单（交易凭证）、积分流水（与订单对应）。
func (s *ProfileService) Deactivate(userID int64) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var u model.User
		if err := tx.First(&u, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProfileNotFound
			}
			return err
		}
		if u.DeactivatedAt != nil {
			return ErrAlreadyDeactivated
		}
		now := time.Now()
		tomb := fmt.Sprintf("deleted:%d:%d", u.ID, now.UnixNano())
		if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
			"open_id":        tomb,
			"status":         0,
			"deactivated_at": now,
			"nickname":       "已注销用户",
			"avatar":         "",
		}).Error; err != nil {
			return err
		}
		// 作品：下架 + 软删（status=0），广场不再可见
		if err := tx.Model(&model.Work{}).Where("user_id = ?", userID).
			Updates(map[string]any{"status": 0, "is_public": 0}).Error; err != nil {
			return err
		}
		// 购物车：直接清空（无历史价值）
		if err := tx.Where("user_id = ?", userID).Delete(&model.CartItem{}).Error; err != nil {
			return err
		}
		// 地址：软删（订单快照里已有收货信息，地址本身不删行）
		if err := tx.Model(&model.Address{}).Where("user_id = ?", userID).
			Update("status", 0).Error; err != nil {
			return err
		}
		return nil
	})
}
