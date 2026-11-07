package service

import (
	"testing"

	"xmf/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newProfileTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(model.AllMigrations()...); err != nil {
		t.Fatal(err)
	}
	return db
}

// 注销：清作品/购物车/地址、打上注销标记、释放 openid 让同一微信能重新注册
func TestProfile_Deactivate(t *testing.T) {
	db := newProfileTestDB(t)
	auth := NewAuthService(db, "", "")
	u, err := auth.DevLogin("openid_todelete")
	if err != nil {
		t.Fatal(err)
	}
	uid := u.ID

	// 铺数据：1 个已发布作品、1 条购物车、1 个地址、1 笔订单
	if err := db.Create(&model.Work{UserID: uid, Type: "collage", Status: 1, IsPublic: 1, ImageURL: "/static/a.png"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.SKU{ID: 1}).Error; err != nil {
		// SKU 表若有必填字段，直接建购物车行（外键未强制）
		t.Logf("create sku: %v", err)
	}
	if err := db.Create(&model.CartItem{UserID: uid, SKUID: 1, WorkID: 0, Qty: 1, Checked: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Address{UserID: uid, Name: "张三", Phone: "13800001111", Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	order := model.Order{UserID: uid, OrderNo: "XMFTEST0001", Status: "done", TotalCents: 100}
	if err := db.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	svc := &ProfileService{DB: db}
	if err := svc.Deactivate(uid); err != nil {
		t.Fatalf("注销失败: %v", err)
	}

	// 1. 用户：禁用 + 打标记 + 资料清空 + openid 已换成 tombstone
	var got model.User
	if err := db.First(&got, uid).Error; err != nil {
		t.Fatal(err)
	}
	if got.Status != 0 || got.DeactivatedAt == nil {
		t.Fatalf("应禁用并写入注销时间: status=%d at=%v", got.Status, got.DeactivatedAt)
	}
	if got.Nickname != "已注销用户" || got.Avatar != "" {
		t.Fatalf("资料应清空: %q %q", got.Nickname, got.Avatar)
	}
	if got.OpenID == "openid_todelete" || got.OpenID == "" {
		t.Fatalf("openid 应换成 tombstone，实际 %q", got.OpenID)
	}

	// 2. 作品：下架 + 软删
	var w model.Work
	db.Where("user_id = ?", uid).First(&w)
	if w.Status != 0 || w.IsPublic != 0 {
		t.Fatalf("作品应下架软删: status=%d public=%d", w.Status, w.IsPublic)
	}

	// 3. 购物车清空
	var cartN int64
	db.Model(&model.CartItem{}).Where("user_id = ?", uid).Count(&cartN)
	if cartN != 0 {
		t.Fatalf("购物车应清空，剩 %d 条", cartN)
	}

	// 4. 地址软删
	var addr model.Address
	db.Where("user_id = ?", uid).First(&addr)
	if addr.Status != 0 {
		t.Fatalf("地址应软删: status=%d", addr.Status)
	}

	// 5. 订单保留（交易凭证）
	var orderN int64
	db.Model(&model.Order{}).Where("user_id = ?", uid).Count(&orderN)
	if orderN != 1 {
		t.Fatalf("订单应保留，实际 %d 笔", orderN)
	}

	// 6. 同一微信再登录 → 全新账号（不等于原 id）
	nu, err := auth.DevLogin("openid_todelete")
	if err != nil {
		t.Fatalf("注销后重新登录应成功: %v", err)
	}
	if nu.ID == uid {
		t.Fatal("应创建新账号而不是复用已注销账号")
	}

	// 7. 重复注销 → 409 语义
	if err := svc.Deactivate(uid); err != ErrAlreadyDeactivated {
		t.Fatalf("重复注销应返回 ErrAlreadyDeactivated，实际 %v", err)
	}
}

// 注销不存在的用户
func TestProfile_DeactivateNotFound(t *testing.T) {
	db := newProfileTestDB(t)
	svc := &ProfileService{DB: db}
	if err := svc.Deactivate(99999); err != ErrProfileNotFound {
		t.Fatalf("应返回 ErrProfileNotFound，实际 %v", err)
	}
}
