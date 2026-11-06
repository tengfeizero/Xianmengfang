package service

import (
	"errors"
	"regexp"

	"xmf/internal/model"

	"gorm.io/gorm"
)

// 地址业务规则：上限 20 条、默认地址唯一、只能操作自己的地址、删除默认地址后自动补位
const (
	maxAddressesPerUser = 20
)

var phoneRe = regexp.MustCompile(`^1\d{10}$`)

var (
	ErrAddressNotFound  = errors.New("地址不存在")
	ErrAddressForbidden = errors.New("无权操作该地址")
	ErrAddressLimit     = errors.New("收货地址最多 20 条")
	ErrAddressInvalid   = errors.New("地址信息不完整")
)

// AddressInput 新增/编辑地址入参
type AddressInput struct {
	Name      string
	Phone     string
	Province  string
	City      string
	District  string
	Detail    string
	IsDefault bool
}

type AddressService struct {
	DB *gorm.DB
}

func validateAddress(in AddressInput) error {
	if in.Name == "" || in.Phone == "" || in.Province == "" || in.City == "" || in.Detail == "" {
		return ErrAddressInvalid
	}
	if !phoneRe.MatchString(in.Phone) {
		return errors.New("手机号格式不正确")
	}
	return nil
}

func (s *AddressService) fullAddress(a *model.Address) string {
	return a.Province + a.City + a.District + a.Detail
}

// List 用户的地址列表（不含已删除）
func (s *AddressService) List(userID int64) ([]model.Address, error) {
	var list []model.Address
	err := s.DB.Where("user_id = ? AND status = 1", userID).
		Order("is_default DESC, id DESC").Find(&list).Error
	return list, err
}

// Create 新增地址；第一个地址自动设为默认
func (s *AddressService) Create(userID int64, in AddressInput) (*model.Address, error) {
	if err := validateAddress(in); err != nil {
		return nil, err
	}
	var count int64
	if err := s.DB.Model(&model.Address{}).Where("user_id = ? AND status = 1", userID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count >= maxAddressesPerUser {
		return nil, ErrAddressLimit
	}
	a := &model.Address{
		UserID: userID, Name: in.Name, Phone: in.Phone,
		Province: in.Province, City: in.City, District: in.District, Detail: in.Detail,
		IsDefault: boolToInt(in.IsDefault || count == 0), Status: 1,
	}
	if a.IsDefault == 1 {
		s.clearDefault(s.DB, userID, 0)
	}
	if err := s.DB.Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

// Update 编辑地址（也可调整是否默认）
func (s *AddressService) Update(userID, addressID int64, in AddressInput) error {
	if err := validateAddress(in); err != nil {
		return err
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := s.owned(tx, userID, addressID)
		if err != nil {
			return err
		}
		updates := map[string]any{
			"name": in.Name, "phone": in.Phone,
			"province": in.Province, "city": in.City, "district": in.District, "detail": in.Detail,
		}
		if in.IsDefault && a.IsDefault != 1 {
			s.clearDefault(tx, userID, addressID)
			updates["is_default"] = 1
		} else if !in.IsDefault && a.IsDefault == 1 {
			// 允许取消默认（保持用户自由），但列表仍按默认优先展示
			updates["is_default"] = 0
		}
		return tx.Model(a).Updates(updates).Error
	})
}

// SetDefault 设为默认（唯一默认）
func (s *AddressService) SetDefault(userID, addressID int64) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.owned(tx, userID, addressID); err != nil {
			return err
		}
		s.clearDefault(tx, userID, addressID)
		return tx.Model(&model.Address{}).Where("id = ?", addressID).Update("is_default", 1).Error
	})
}

// Delete 软删；若删的是默认地址，自动把最新一条设为默认
func (s *AddressService) Delete(userID, addressID int64) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := s.owned(tx, userID, addressID)
		if err != nil {
			return err
		}
		if err := tx.Model(a).Update("status", 0).Error; err != nil {
			return err
		}
		if a.IsDefault == 1 {
			var next model.Address
			if err := tx.Where("user_id = ? AND status = 1", userID).Order("id DESC").First(&next).Error; err == nil {
				if err := tx.Model(&next).Update("is_default", 1).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// GetForOrder 下单取地址：校验归属与状态，返回快照所需信息
func (s *AddressService) GetForOrder(userID, addressID int64) (*model.Address, error) {
	return s.owned(s.DB, userID, addressID)
}

func (s *AddressService) owned(tx *gorm.DB, userID, addressID int64) (*model.Address, error) {
	var a model.Address
	if err := tx.First(&a, addressID).Error; err != nil || a.Status != 1 {
		return nil, ErrAddressNotFound
	}
	if a.UserID != userID {
		return nil, ErrAddressForbidden
	}
	return &a, nil
}

func (s *AddressService) clearDefault(tx *gorm.DB, userID, excludeID int64) {
	q := tx.Model(&model.Address{}).Where("user_id = ? AND is_default = 1", userID)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	q.Update("is_default", 0)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
