package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// Config 运行配置。密钥只走环境变量；JWT 密钥无环境变量时生成随机值持久化到 data/。
type Config struct {
	Addr      string // 监听地址
	DBPath    string // SQLite 文件路径
	JWTSecret string
	WxAppID   string
	WxSecret  string
	DevLogin  bool // 本地开发模拟登录开关
	DataDir   string

	// AI 生图：AIKey 有值即启用真实供应商（腾讯混元），否则回落 Mock
	AIKey      string
	AISecretID string
	AIBaseURL  string
	AIModel    string
	AIStyle    string
}

func Load() (*Config, error) {
	c := &Config{
		Addr:     envOr("XMF_ADDR", ":8080"),
		DBPath:   envOr("XMF_DB", "data/xmf.db"),
		WxAppID:  os.Getenv("XMF_WX_APPID"),
		WxSecret: os.Getenv("XMF_WX_SECRET"),
		DevLogin: envOr("XMF_DEV_LOGIN", "1") == "1",

		AIKey:      os.Getenv("XMF_AI_API_KEY"),
		AISecretID: os.Getenv("XMF_AI_SECRET_ID"),
		AIBaseURL:  os.Getenv("XMF_AI_BASE_URL"),
		AIModel:    envOr("XMF_AI_MODEL", "hy-image-v3.5-preview"),
		AIStyle:    envOr("XMF_AI_STYLE", "messages"),
	}
	c.DataDir = filepath.Dir(c.DBPath)
	if err := os.MkdirAll(c.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}
	secret := os.Getenv("XMF_JWT_SECRET")
	if secret == "" {
		persisted, err := loadOrCreateSecret(filepath.Join(c.DataDir, ".jwt_secret"))
		if err != nil {
			return nil, err
		}
		secret = persisted
	}
	c.JWTSecret = secret
	return c, nil
}

func loadOrCreateSecret(path string) (string, error) {
	if b, err := os.ReadFile(path); err == nil && len(b) >= 32 {
		return string(b), nil
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	s := hex.EncodeToString(buf)
	if err := os.WriteFile(path, []byte(s), 0o600); err != nil {
		return "", err
	}
	return s, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
