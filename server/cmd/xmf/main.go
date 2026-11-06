package main

import (
	"log"
	"path/filepath"

	"xmf/internal/api"
	"xmf/internal/config"
	"xmf/internal/repo"
	"xmf/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}
	db, err := repo.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	uploadsDir := filepath.Join(cfg.DataDir, "uploads")
	if err := repo.Seed(db, uploadsDir); err != nil {
		log.Fatalf("写入演示数据失败: %v", err)
	}

	h := &api.Handlers{
		DB:         db,
		Secret:     cfg.JWTSecret,
		Auth:       service.NewAuthService(db, cfg.WxAppID, cfg.WxSecret),
		Orders:     &service.OrderService{DB: db},
		AI:         &service.AIService{DB: db, Provider: service.MockProvider{}, UploadsDir: uploadsDir},
		UploadsDir: uploadsDir,
		Dev:        cfg.DevLogin,
	}
	r := api.Router(h)
	log.Printf("现梦坊服务启动: %s (db=%s, dev_login=%v, ai=%s)", cfg.Addr, cfg.DBPath, cfg.DevLogin, h.AI.Provider.Name())
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
