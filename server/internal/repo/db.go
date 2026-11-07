package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	// 角色种子（按 code 幂等；super 权限恒为通配，不可改）
	superID, err := ensureRole(db, "super", "超级管理员", "系统内置：拥有全部权限，不可删除、不可改权限", `["*"]`)
	if err != nil {
		return err
	}
	if _, err := ensureRole(db, "ops", "运营专员", "订单处理、发货、商品素材查看、作品上下架",
		`["asset:view","order:cancel","order:process","order:ship","order:view","product:view","user:view","work:edit","work:view"]`); err != nil {
		return err
	}
	if _, err := ensureRole(db, "viewer", "只读账号", "仅可查看，不能执行任何写操作",
		`["asset:view","order:view","product:view","user:view","work:view"]`); err != nil {
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
		if err := db.Create(&model.Admin{
			Username: "admin", PasswordHash: string(hash), Nickname: "系统管理员", RoleID: superID, Status: 1,
		}).Error; err != nil {
			return err
		}
	}
	// 老库升级：v1.4 之前的账号没有角色，统一绑到超级管理员
	if err := db.Model(&model.Admin{}).Where("role_id = 0 OR role_id IS NULL").
		Update("role_id", superID).Error; err != nil {
		return err
	}

	// 系统配置
	defaults := map[string]string{
		"ai_daily_free_limit": "3",
		"work_review_enabled": "false",
		// AI 生图（腾讯混元 hy-image）：api_key 留空 = 走 Mock；后台系统配置页可改
		"ai_api_key":   "",
		"ai_secret_id": "", // TokenHub Key ID（ak），预留
		"ai_base_url":  "https://tokenhub.tencentmaas.com",
		"ai_model":     "hy-image-v3.5-preview",
		"ai_style":     "messages",
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
	if err := SeedCatalogAssets(db); err != nil {
		return err
	}
	// v1.7：老库补齐社区字段（新增列在老行里是 NULL，直接参与 +1 运算会一直为 NULL）
	return MigrateWorkCommunity(db)
}

// MigrateWorkCommunity v1.7 数据迁移：
//  1. works / users 新增列的 NULL 归一（is_public/like_count/use_count/prompt/points）
//  2. 老库 AI 作品的提示词从 meta 回填到 works.prompt（发布到广场时要展示）
//
// 幂等，可重复执行；拼装作品没有提示词，跳过。
func MigrateWorkCommunity(db *gorm.DB) error {
	if err := db.Exec(`UPDATE works SET
		is_public = COALESCE(is_public, 0),
		like_count = COALESCE(like_count, 0),
		use_count = COALESCE(use_count, 0),
		prompt = COALESCE(prompt, '')`).Error; err != nil {
		return err
	}
	if err := db.Exec("UPDATE users SET points = COALESCE(points, 0)").Error; err != nil {
		return err
	}

	var works []model.Work
	if err := db.Where("prompt = '' AND type IN ?", []string{"ai_text", "ai_photo"}).Find(&works).Error; err != nil {
		return err
	}
	for _, w := range works {
		var meta struct {
			Prompt string `json:"prompt"`
		}
		if err := json.Unmarshal([]byte(w.Meta), &meta); err != nil {
			continue
		}
		p := strings.TrimSpace(meta.Prompt)
		if p == "" {
			continue
		}
		if err := db.Model(&model.Work{}).Where("id = ?", w.ID).Update("prompt", p).Error; err != nil {
			return err
		}
	}
	return nil
}

// ensureRole 按 code 幂等写入内置角色，返回角色 ID
func ensureRole(db *gorm.DB, code, name, remark, perms string) (int64, error) {
	var r model.Role
	err := db.Where("code = ?", code).First(&r).Error
	if err == nil {
		return r.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	r = model.Role{Code: code, Name: name, Remark: remark, Perms: perms, Builtin: 1, Status: 1}
	if err := db.Create(&r).Error; err != nil {
		return 0, err
	}
	return r.ID, nil
}
