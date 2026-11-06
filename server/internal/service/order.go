package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"xmf/internal/model"

	"gorm.io/gorm"
)

// 订单状态机（唯一合法流转，见 docs/DATA-MODEL.md）
const (
	StatusPendingAccept = "pending_accept" // 待接单（mock 支付下单直达）
	StatusPendingShip   = "pending_ship"   // 待发货
	StatusShipped       = "shipped"        // 已发货
	StatusDone          = "done"           // 已完成（确认收货）
	StatusCanceled      = "canceled"       // 已取消
)

var (
	ErrNotFound          = errors.New("订单不存在")
	ErrStockInsufficient = errors.New("商品库存不足")
	ErrIllegalTransition = errors.New("订单当前状态不允许该操作")
	ErrForbidden         = errors.New("无权操作该订单")
	ErrEmptyItems        = errors.New("下单明细不能为空")
)

type OrderService struct {
	DB *gorm.DB
}

// ItemInput 下单项
type ItemInput struct {
	SKUID  int64
	WorkID *int64
	Qty    int
}

// CreateResult 下单结果
type CreateResult struct {
	Order      *model.Order
	TotalCents int64
}

// Create 下单（mock 支付：直接进入待接单）。库存原子扣减，任何一步失败整体回滚。
func (s *OrderService) Create(userID int64, items []ItemInput) (*CreateResult, error) {
	if len(items) == 0 {
		return nil, ErrEmptyItems
	}
	var total int64
	now := time.Now()
	order := &model.Order{
		OrderNo:  genOrderNo(now),
		UserID:   userID,
		Status:   StatusPendingAccept,
		PayType:  "mock",
		PaidAt:   &now,
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		total = 0
		var snapshots []model.OrderItem
		for _, it := range items {
			if it.Qty <= 0 {
				return errors.New("购买数量必须大于 0")
			}
			// 库存原子扣减：不允许"读出来算好再写回"
			res := tx.Exec(
				"UPDATE skus SET stock = stock - ? WHERE id = ? AND stock >= ?",
				it.Qty, it.SKUID, it.Qty)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return ErrStockInsufficient
			}
			var sku model.SKU
			if err := tx.First(&sku, it.SKUID).Error; err != nil {
				return errors.New("商品规格不存在")
			}
			var p model.Product
			if err := tx.First(&p, sku.ProductID).Error; err != nil {
				return errors.New("商品不存在")
			}
			if p.Status != 1 {
				return errors.New("商品已下架")
			}
			total += sku.PriceCents * int64(it.Qty)
			snapshots = append(snapshots, model.OrderItem{
				SKUID:      sku.ID,
				WorkID:     it.WorkID,
				Title:      p.Title,
				Specs:      sku.Specs,
				PriceCents: sku.PriceCents,
				Qty:        it.Qty,
			})
		}
		order.TotalCents = total
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for i := range snapshots {
			snapshots[i].OrderID = order.ID
		}
		if err := tx.Create(&snapshots).Error; err != nil {
			return err
		}
		return appendEvent(tx, order.ID, "", StatusPendingAccept, fmt.Sprintf("user:%d", userID), "下单，模拟支付成功")
	})
	if err != nil {
		return nil, err
	}
	return &CreateResult{Order: order, TotalCents: total}, nil
}

// Cancel 取消订单。用户只能取消待接单；管理员可取消待接单和待发货。取消恢复库存。
func (s *OrderService) Cancel(orderID, userID int64, byAdmin bool, reason string) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var o model.Order
		if err := tx.First(&o, orderID).Error; err != nil {
			return ErrNotFound
		}
		if !byAdmin && o.UserID != userID {
			return ErrForbidden
		}
		allowed := map[string]bool{StatusPendingAccept: true}
		if byAdmin {
			allowed[StatusPendingShip] = true
		}
		if !allowed[o.Status] {
			return ErrIllegalTransition
		}
		return s.cancelTx(tx, &o, userID, byAdmin, reason)
	})
}

func (s *OrderService) cancelTx(tx *gorm.DB, o *model.Order, userID int64, byAdmin bool, reason string) error {
	// 恢复库存
	var items []model.OrderItem
	if err := tx.Where("order_id = ?", o.ID).Find(&items).Error; err != nil {
		return err
	}
	for _, it := range items {
		if err := tx.Exec("UPDATE skus SET stock = stock + ? WHERE id = ?", it.Qty, it.SKUID).Error; err != nil {
			return err
		}
	}
	operator := fmt.Sprintf("user:%d", userID)
	if byAdmin {
		operator = fmt.Sprintf("admin:%d", userID)
	}
	o.Status = StatusCanceled
	o.CancelReason = reason
	if err := tx.Model(o).Updates(map[string]any{"status": o.Status, "cancel_reason": o.CancelReason}).Error; err != nil {
		return err
	}
	return appendEvent(tx, o.ID, fromOf(o), StatusCanceled, operator, "取消订单："+reason)
}

// Accept 接单：pending_accept → pending_ship
func (s *OrderService) Accept(orderID, adminID int64) error {
	return s.transition(orderID, adminID, StatusPendingAccept, StatusPendingShip, "接单")
}

// Ship 发货：pending_ship → shipped
func (s *OrderService) Ship(orderID, adminID int64, company, trackingNo string) error {
	if company == "" || trackingNo == "" {
		return errors.New("物流公司和单号不能为空")
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var o model.Order
		if err := tx.First(&o, orderID).Error; err != nil {
			return ErrNotFound
		}
		if o.Status != StatusPendingShip {
			return ErrIllegalTransition
		}
		now := time.Now()
		if err := tx.Model(o).Updates(map[string]any{"status": StatusShipped, "shipped_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Shipment{OrderID: o.ID, Company: company, TrackingNo: trackingNo}).Error; err != nil {
			return err
		}
		return appendEvent(tx, o.ID, StatusPendingShip, StatusShipped, fmt.Sprintf("admin:%d", adminID),
			fmt.Sprintf("发货：%s %s", company, trackingNo))
	})
}

// Confirm 确认收货：shipped → done（仅订单本人）
func (s *OrderService) Confirm(orderID, userID int64) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var o model.Order
		if err := tx.First(&o, orderID).Error; err != nil {
			return ErrNotFound
		}
		if o.UserID != userID {
			return ErrForbidden
		}
		if o.Status != StatusShipped {
			return ErrIllegalTransition
		}
		now := time.Now()
		if err := tx.Model(o).Updates(map[string]any{"status": StatusDone, "done_at": now}).Error; err != nil {
			return err
		}
		return appendEvent(tx, o.ID, StatusShipped, StatusDone, fmt.Sprintf("user:%d", userID), "确认收货")
	})
}

// MarkRefund 后台退款标记（仅标记，MVP 无资金流）
func (s *OrderService) MarkRefund(orderID, adminID int64) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var o model.Order
		if err := tx.First(&o, orderID).Error; err != nil {
			return ErrNotFound
		}
		if o.Status != StatusShipped && o.Status != StatusDone {
			return ErrIllegalTransition
		}
		if err := tx.Model(o).Update("refund_flag", 1).Error; err != nil {
			return err
		}
		return appendEvent(tx, o.ID, o.Status, o.Status, fmt.Sprintf("admin:%d", adminID), "标记退款")
	})
}

func (s *OrderService) transition(orderID, adminID int64, from, to, note string) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var o model.Order
		if err := tx.First(&o, orderID).Error; err != nil {
			return ErrNotFound
		}
		if o.Status != from {
			return ErrIllegalTransition
		}
		if err := tx.Model(o).Update("status", to).Error; err != nil {
			return err
		}
		return appendEvent(tx, o.ID, from, to, fmt.Sprintf("admin:%d", adminID), note)
	})
}

func appendEvent(tx *gorm.DB, orderID int64, from, to, operator, note string) error {
	return tx.Create(&model.OrderEvent{OrderID: orderID, From: from, To: to, Operator: operator, Note: note}).Error
}

func fromOf(o *model.Order) string { return o.Status }

func genOrderNo(now time.Time) string {
	return fmt.Sprintf("XMF%s%04d", now.Format("20060102150405"), rand.Intn(10000))
}
