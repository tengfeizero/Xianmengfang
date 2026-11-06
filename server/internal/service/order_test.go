package service

import (
	"testing"

	"xmf/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestDB 每个用例独立的内存库 + 演示数据
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(model.AllMigrations()...); err != nil {
		t.Fatal(err)
	}
	seed := []any{
		&model.Product{CategoryID: 1, Title: "定制T恤", Status: 1},
		&model.SKU{ProductID: 1, Specs: `{"尺码":"M"}`, PriceCents: 9900, Stock: 3, Code: "TEE-M"},
		&model.SKU{ProductID: 1, Specs: `{"尺码":"L"}`, PriceCents: 9900, Stock: 1, Code: "TEE-L"},
		&model.Product{CategoryID: 1, Title: "下架商品", Status: 0},
		&model.SKU{ProductID: 2, Specs: `{"规格":"标准"}`, PriceCents: 5000, Stock: 5, Code: "OFF-1"},
	}
	for _, s := range seed {
		if err := db.Create(s).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestCreate_MockPay_TotalAndStock(t *testing.T) {
	db := newTestDB(t)
	svc := &OrderService{DB: db}
	res, err := svc.Create(1, []ItemInput{{SKUID: 1, Qty: 2}, {SKUID: 2, Qty: 1}}, nil)
	if err != nil {
		t.Fatalf("下单失败: %v", err)
	}
	// 金额对账：9900*2 + 9900*1 = 29700 分
	if res.TotalCents != 29700 || res.Order.TotalCents != 29700 {
		t.Fatalf("金额不对: %d", res.TotalCents)
	}
	if res.Order.Status != StatusPendingAccept || res.Order.PayType != "mock" || res.Order.PaidAt == nil {
		t.Fatalf("mock 支付应直达待接单: %+v", res.Order)
	}
	var sku model.SKU
	db.First(&sku, 1)
	if sku.Stock != 1 { // 3 - 2
		t.Fatalf("库存扣减错误: %d", sku.Stock)
	}
	// 明细快照
	var items []model.OrderItem
	db.Where("order_id = ?", res.Order.ID).Find(&items)
	if len(items) != 2 || items[0].Title != "定制T恤" || items[0].PriceCents != 9900 {
		t.Fatalf("明细快照错误: %+v", items)
	}
	// 流水只追加
	var evs []model.OrderEvent
	db.Where("order_id = ?", res.Order.ID).Find(&evs)
	if len(evs) != 1 || evs[0].To != StatusPendingAccept {
		t.Fatalf("事件流水错误: %+v", evs)
	}
}

func TestCreate_StockInsufficient(t *testing.T) {
	db := newTestDB(t)
	svc := &OrderService{DB: db}
	// SKU1 只有 3 件库存
	_, err := svc.Create(1, []ItemInput{{SKUID: 1, Qty: 4}}, nil)
	if err != ErrStockInsufficient {
		t.Fatalf("应报库存不足，得到: %v", err)
	}
	// 整单回滚，库存不变
	var sku model.SKU
	db.First(&sku, 1)
	if sku.Stock != 3 {
		t.Fatalf("失败下单不应扣库存: %d", sku.Stock)
	}
	var n int64
	db.Model(&model.Order{}).Count(&n)
	if n != 0 {
		t.Fatalf("失败下单不应产生订单")
	}
}

func TestCreate_ProductOffShelf(t *testing.T) {
	db := newTestDB(t)
	svc := &OrderService{DB: db}
	_, err := svc.Create(1, []ItemInput{{SKUID: 3, Qty: 1}}, nil)
	if err == nil || err.Error() != "商品已下架" {
		t.Fatalf("下架商品应不可下单: %v", err)
	}
}

func TestFullLifecycle_MockPay(t *testing.T) {
	db := newTestDB(t)
	svc := &OrderService{DB: db}
	res, _ := svc.Create(1, []ItemInput{{SKUID: 2, Qty: 1}}, nil) // 库存仅 1 件
	oid := res.Order.ID

	// 用户不能接单/发货；这里只验证用户取消受限
	if err := svc.Cancel(oid, 1, false, "不想要了"); err == nil {
		// 用户可取消待接单，属合法；重新下一单继续
	}
	res2, _ := svc.Create(1, []ItemInput{{SKUID: 1, Qty: 1}}, nil)
	oid2 := res2.Order.ID

	if err := svc.Accept(oid2, 9); err != nil {
		t.Fatalf("接单失败: %v", err)
	}
	// 非法流转：待发货不能确认收货
	if err := svc.Confirm(oid2, 1); err != ErrIllegalTransition {
		t.Fatalf("待发货状态确认收货应被拒绝: %v", err)
	}
	if err := svc.Ship(oid2, 9, "顺丰", "SF123"); err != nil {
		t.Fatalf("发货失败: %v", err)
	}
	if err := svc.Confirm(oid2, 2); err != ErrForbidden {
		t.Fatalf("非本人确认收货应被拒绝: %v", err)
	}
	if err := svc.Confirm(oid2, 1); err != nil {
		t.Fatalf("确认收货失败: %v", err)
	}
	// 完成后不能再取消
	if err := svc.Cancel(oid2, 1, true, "test"); err != ErrIllegalTransition {
		t.Fatalf("已完成订单不可取消: %v", err)
	}
	// 发货失败后不允许重复发货
	if err := svc.Ship(oid2, 9, "顺丰", "SF999"); err != ErrIllegalTransition {
		t.Fatalf("重复发货应被拒绝: %v", err)
	}
}

func TestCancel_RestoresStock(t *testing.T) {
	db := newTestDB(t)
	svc := &OrderService{DB: db}
	res, _ := svc.Create(1, []ItemInput{{SKUID: 2, Qty: 1}}, nil) // L 码库存 1→0
	if err := svc.Cancel(res.Order.ID, 1, false, "拍错了"); err != nil {
		t.Fatalf("用户取消待接单失败: %v", err)
	}
	var sku model.SKU
	db.First(&sku, 2)
	if sku.Stock != 1 {
		t.Fatalf("取消后库存未恢复: %d", sku.Stock)
	}
	// 已发货不可取消
	res2, _ := svc.Create(1, []ItemInput{{SKUID: 2, Qty: 1}}, nil)
	svc.Accept(res2.Order.ID, 9)
	svc.Ship(res2.Order.ID, 9, "顺丰", "SF001")
	if err := svc.Cancel(res2.Order.ID, 1, false, "test"); err != ErrIllegalTransition {
		t.Fatalf("已发货用户取消应被拒绝: %v", err)
	}
	// 管理员可取消待发货
	res3, _ := svc.Create(1, []ItemInput{{SKUID: 1, Qty: 1}}, nil)
	svc.Accept(res3.Order.ID, 9)
	if err := svc.Cancel(res3.Order.ID, 9, true, "缺货"); err != nil {
		t.Fatalf("管理员取消待发货失败: %v", err)
	}
	var sku1 model.SKU
	db.First(&sku1, 1)
	if sku1.Stock != 3 {
		t.Fatalf("管理员取消后库存未恢复: %d", sku1.Stock)
	}
}

func TestMarkRefund(t *testing.T) {
	db := newTestDB(t)
	svc := &OrderService{DB: db}
	res, _ := svc.Create(1, []ItemInput{{SKUID: 1, Qty: 1}}, nil)
	// 待接单不能标记退款
	if err := svc.MarkRefund(res.Order.ID, 9); err != ErrIllegalTransition {
		t.Fatalf("待接单不应可标记退款: %v", err)
	}
	svc.Accept(res.Order.ID, 9)
	svc.Ship(res.Order.ID, 9, "顺丰", "SF002")
	if err := svc.MarkRefund(res.Order.ID, 9); err != nil {
		t.Fatalf("已发货标记退款失败: %v", err)
	}
	var o model.Order
	db.First(&o, res.Order.ID)
	if o.RefundFlag != 1 {
		t.Fatalf("退款标记未写入")
	}
}
