package repo

import (
	"fmt"

	"xmf/internal/model"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open 打开数据库并执行幂等迁移
func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path+"?_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	if err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;").Error; err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(model.AllMigrations()...); err != nil {
		return nil, fmt.Errorf("迁移失败: %w", err)
	}
	return db, nil
}

// Seed 演示数据：仅当为空时写入，可重复执行
func Seed(db *gorm.DB, uploadsDir string) error {
	// 演示素材 PNG（缺文件才生成）
	if err := SeedAssets(uploadsDir); err != nil {
		return err
	}
	// 管理员 admin / admin123（上线前必须改密）
	var adminCount int64
	if err := db.Model(&model.Admin{}).Count(&adminCount).Error; err != nil {
		return err
	}
	if adminCount == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if err := db.Create(&model.Admin{Username: "admin", PasswordHash: string(hash), Status: 1}).Error; err != nil {
			return err
		}
	}

	// 系统配置
	defaults := map[string]string{
		"ai_daily_free_limit": "3",
		"work_review_enabled": "false",
	}
	for k, v := range defaults {
		var n int64
		if err := db.Model(&model.Config{}).Where("key = ?", k).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if err := db.Create(&model.Config{Key: k, Value: v}).Error; err != nil {
				return err
			}
		}
	}

	// 商品分类 + 演示商品
	var catCount int64
	if err := db.Model(&model.ProductCategory{}).Count(&catCount).Error; err != nil {
		return err
	}
	if catCount > 0 {
		return nil
	}
	cats := []model.ProductCategory{{Name: "服饰", Sort: 1, Status: 1}, {Name: "周边", Sort: 2, Status: 1}, {Name: "玩偶", Sort: 3, Status: 1}}
	if err := db.Create(&cats).Error; err != nil {
		return err
	}
	products := []model.Product{
		{CategoryID: cats[0].ID, Title: "定制图案T恤", Description: "纯棉 220g，把你的作品穿在身上", Images: "[]", Status: 1},
		{CategoryID: cats[1].ID, Title: "定制图案马克杯", Description: "陶瓷白杯 350ml，作品图案热转印", Images: "[]", Status: 1},
		{CategoryID: cats[1].ID, Title: "定制帆布袋", Description: "加厚帆布，单面印制", Images: "[]", Status: 1},
		{CategoryID: cats[2].ID, Title: "定制小姑娘玩偶", Description: "按作品形象定制毛绒玩偶，20cm", Images: "[]", Status: 1},
	}
	if err := db.Create(&products).Error; err != nil {
		return err
	}
	skus := []model.SKU{
		{ProductID: products[0].ID, Specs: `{"尺码":"M","颜色":"白"}`, PriceCents: 9900, Stock: 100, Code: "TEE-W-M"},
		{ProductID: products[0].ID, Specs: `{"尺码":"L","颜色":"白"}`, PriceCents: 9900, Stock: 100, Code: "TEE-W-L"},
		{ProductID: products[0].ID, Specs: `{"尺码":"M","颜色":"黑"}`, PriceCents: 9900, Stock: 80, Code: "TEE-B-M"},
		{ProductID: products[1].ID, Specs: `{"规格":"标准"}`, PriceCents: 5900, Stock: 200, Code: "MUG-STD"},
		{ProductID: products[2].ID, Specs: `{"规格":"标准"}`, PriceCents: 6900, Stock: 150, Code: "BAG-STD"},
		{ProductID: products[3].ID, Specs: `{"规格":"20cm"}`, PriceCents: 19900, Stock: 50, Code: "DOLL-20"},
	}
	if err := db.Create(&skus).Error; err != nil {
		return err
	}

	// 拼装底板与元素
	return SeedCatalogAssets(db)
}
