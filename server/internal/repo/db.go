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
	// 商品演示封面（缺文件才生成）
	if err := SeedProductImages(uploadsDir); err != nil {
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

	// 商品分类 + 演示商品（空库写入，老库按标题回填空字段）
	if err := SeedShop(db); err != nil {
		return err
	}

	// 拼装底板与元素
	return SeedCatalogAssets(db)
}
