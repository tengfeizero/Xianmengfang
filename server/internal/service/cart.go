package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"xmf/internal/model"

	"gorm.io/gorm"
)

// 购物车业务规则（PRD 3.2.2）
const (
	CartQtyMax   = 99 // 单件数量上限
	CartItemsMax = 50 // 购物车条数上限
	CartWorkNone = 0  // WorkID=0 表示未绑作品（NULL 在 SQLite 唯一索引里互不相等，禁用）
)

// 购物车错误。msg 面向用户可读，handler 直接透传。
var (
	ErrCartNotFound    = errors.New("购物车项不存在")
	ErrCartForbidden   = errors.New("无权操作该购物车项")
	ErrCartFull        = fmt.Errorf("购物车最多放 %d 种商品", CartItemsMax)
	ErrCartQtyMax      = fmt.Errorf("单件商品最多 %d 件", CartQtyMax)
	ErrCartEmpty       = errors.New("无可结算的商品")
	ErrCartInvalidItem = errors.New("购物车含失效商品，请先移除后再结算")
	ErrCartBadWork     = errors.New("作品已失效，请重新选择图案")
)

type CartService struct {
	DB     *gorm.DB
	Orders *OrderService
}

// CartView 列表渲染项。实时 join，不存快照（见 DATA-MODEL cart_items）。
type CartView struct {
	model.CartItem
	Title         string `json:"title"`
	Specs         string `json:"specs"`
	PriceCents    int64  `json:"price_cents"`
	ImageURL      string `json:"image_url"`
	WorkThumb     string `json:"work_thumb"`
	Stock         int64  `json:"stock"`
	Valid         bool   `json:"valid"`
	InvalidReason string `json:"invalid_reason"`
}

type CartSummary struct {
	CheckedCount int   `json:"checked_count"`
	CheckedCents int64 `json:"checked_cents"`
}

// firstImage 商品图 JSON 数组取首张
func firstImage(imagesJSON string) string {
	var arr []string
	if err := json.Unmarshal([]byte(imagesJSON), &arr); err != nil || len(arr) == 0 {
		return ""
	}
	return arr[0]
}

// List 全量渲染 + 勾选汇总。失效项 valid=false，仍返回（前端标灰、可删）。
func (s *CartService) List(userID int64) ([]CartView, CartSummary, error) {
	var items []model.CartItem
	if err := s.DB.Where("user_id = ?", userID).Order("id desc").Find(&items).Error; err != nil {
		return nil, CartSummary{}, err
	}
	views := make([]CartView, 0, len(items))
	summary := CartSummary{}
	for _, it := range items {
		v := CartView{CartItem: it, Valid: true}
		var sku model.SKU
		if err := s.DB.First(&sku, it.SKUID).Error; err != nil {
			v.Valid, v.InvalidReason = false, "商品规格已失效"
		} else {
			var p model.Product
			if err := s.DB.First(&p, sku.ProductID).Error; err != nil || p.Status != 1 {
				v.Valid, v.InvalidReason = false, "商品已下架"
			}
			v.Title, v.Specs, v.PriceCents, v.Stock = p.Title, sku.Specs, sku.PriceCents, sku.Stock
			v.ImageURL = firstImage(p.Images)
			if it.WorkID > 0 {
				var w model.Work
				if err := s.DB.First(&w, it.WorkID).Error; err != nil || w.Status != 1 {
					v.Valid, v.InvalidReason = false, "作品已删除，请重新选择图案"
				} else if _, err := (&CommunityService{DB: s.DB}).UsableWork(userID, it.WorkID); err != nil {
					v.Valid, v.InvalidReason = false, "作品已取消发布，请重新选择图案"
				} else {
					v.WorkThumb = w.ThumbnailURL
					if v.WorkThumb == "" {
						v.WorkThumb = w.ImageURL
					}
				}
			}
			if sku.Stock <= 0 && v.Valid {
				v.Valid, v.InvalidReason = false, "库存不足"
			}
		}
		if it.Checked == 1 && v.Valid {
			summary.CheckedCount += it.Qty
			summary.CheckedCents += v.PriceCents * int64(it.Qty)
		}
		views = append(views, v)
	}
	return views, summary, nil
}

// Add 加购：同 user+sku+work 数量累加；否则新建。不锁库存，只做基本校验。
func (s *CartService) Add(userID, skuID, workID int64, qty int) (*model.CartItem, error) {
	if qty <= 0 {
		qty = 1
	}
	if qty > CartQtyMax {
		return nil, ErrCartQtyMax
	}
	if workID < 0 {
		workID = CartWorkNone
	}
	var sku model.SKU
	if err := s.DB.First(&sku, skuID).Error; err != nil {
		return nil, errors.New("商品规格不存在")
	}
	var p model.Product
	if err := s.DB.First(&p, sku.ProductID).Error; err != nil || p.Status != 1 {
		return nil, errors.New("商品已下架")
	}
	if sku.Stock < int64(qty) {
		return nil, ErrStockInsufficient
	}
	if workID > 0 {
		// v1.7：自己的作品，或已点赞的他人公开作品都可下单
		if _, err := (&CommunityService{DB: s.DB}).UsableWork(userID, workID); err != nil {
			return nil, ErrCartBadWork
		}
	}

	var item model.CartItem
	err := s.DB.Where("user_id = ? AND sku_id = ? AND work_id = ?", userID, skuID, workID).First(&item).Error
	if err == nil {
		item.Qty += qty
		if item.Qty > CartQtyMax {
			return nil, ErrCartQtyMax
		}
		item.Checked = 1
		if err := s.DB.Model(&model.CartItem{}).Where("id = ?", item.ID).
			Updates(map[string]any{"qty": item.Qty, "checked": 1}).Error; err != nil {
			return nil, err
		}
		return &item, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	var count int64
	if err := s.DB.Model(&model.CartItem{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count >= CartItemsMax {
		return nil, ErrCartFull
	}
	item = model.CartItem{UserID: userID, SKUID: skuID, WorkID: workID, Qty: qty, Checked: 1}
	if err := s.DB.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// owned 归属校验：他人 403、不存在 404
func (s *CartService) owned(userID, itemID int64) (*model.CartItem, error) {
	var it model.CartItem
	if err := s.DB.First(&it, itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCartNotFound
		}
		return nil, err
	}
	if it.UserID != userID {
		return nil, ErrCartForbidden
	}
	return &it, nil
}

// Update 改数量或勾选。qty/checked 用指针区分"没传"与传 0。
func (s *CartService) Update(userID, itemID int64, qty, checked *int) error {
	it, err := s.owned(userID, itemID)
	if err != nil {
		return err
	}
	updates := map[string]any{}
	if qty != nil {
		if *qty <= 0 || *qty > CartQtyMax {
			return ErrCartQtyMax
		}
		updates["qty"] = *qty
	}
	if checked != nil {
		if *checked != 0 && *checked != 1 {
			return errors.New("参数错误")
		}
		updates["checked"] = *checked
	}
	if len(updates) == 0 {
		return nil
	}
	return s.DB.Model(&model.CartItem{}).Where("id = ?", it.ID).Updates(updates).Error
}

// SetChecked 批量勾选；itemIDs 为空 = 全选/全不选
func (s *CartService) SetChecked(userID int64, checked int, itemIDs []int64) error {
	if checked != 0 && checked != 1 {
		return errors.New("参数错误")
	}
	q := s.DB.Model(&model.CartItem{}).Where("user_id = ?", userID)
	if len(itemIDs) > 0 {
		q = q.Where("id IN ?", itemIDs)
	}
	return q.Update("checked", checked).Error
}

// Remove 删除单项
func (s *CartService) Remove(userID, itemID int64) error {
	it, err := s.owned(userID, itemID)
	if err != nil {
		return err
	}
	return s.DB.Delete(&model.CartItem{}, it.ID).Error
}

// Clear 清空；onlyChecked=true 只清勾选项
func (s *CartService) Clear(userID int64, onlyChecked bool) error {
	q := s.DB.Where("user_id = ?", userID)
	if onlyChecked {
		q = q.Where("checked = 1")
	}
	return q.Delete(&model.CartItem{}).Error
}

// Checkout 结算：勾选项 → 一单多明细 → 成功后移除对应购物车项。
// 地址校验在 api 层完成（与直接下单同一路径）。删除失败仅残留，不影响订单。
func (s *CartService) Checkout(userID int64, itemIDs []int64, addr *model.Address) (*CreateResult, error) {
	q := s.DB.Where("user_id = ?", userID)
	if len(itemIDs) > 0 {
		q = q.Where("id IN ?", itemIDs)
	} else {
		q = q.Where("checked = 1")
	}
	var items []model.CartItem
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrCartEmpty
	}
	ids := make([]int64, 0, len(items))
	inputs := make([]ItemInput, 0, len(items))
	for _, it := range items {
		// 结算前逐项校验有效性：下架/删除/缺货在此时拦截，而不是等 Create 里报一半
		var sku model.SKU
		if err := s.DB.First(&sku, it.SKUID).Error; err != nil {
			return nil, ErrCartInvalidItem
		}
		var p model.Product
		if err := s.DB.First(&p, sku.ProductID).Error; err != nil || p.Status != 1 {
			return nil, ErrCartInvalidItem
		}
		if sku.Stock < int64(it.Qty) {
			return nil, ErrStockInsufficient
		}
		if it.WorkID > 0 {
			// 作者取消发布/下架后，已加购的项在结算时拦截（不静默换成别的图）
			if _, err := (&CommunityService{DB: s.DB}).UsableWork(userID, it.WorkID); err != nil {
				return nil, ErrCartInvalidItem
			}
		}
		ids = append(ids, it.ID)
		wid := it.WorkID
		var workPtr *int64
		if wid > 0 {
			workPtr = &wid
		}
		inputs = append(inputs, ItemInput{SKUID: it.SKUID, WorkID: workPtr, Qty: it.Qty})
	}
	res, err := s.Orders.Create(userID, inputs, addr)
	if err != nil {
		return nil, err
	}
	// 订单已成立，移除失败只是购物车残留（用户可手动删），不回滚订单
	if err := s.DB.Where("id IN ?", ids).Delete(&model.CartItem{}).Error; err != nil {
		return res, nil
	}
	return res, nil
}
