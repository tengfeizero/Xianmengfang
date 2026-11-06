package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"xmf/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newM2TestEnv(t *testing.T) (*gorm.DB, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(model.AllMigrations()...); err != nil {
		t.Fatal(err)
	}
	seed := []any{
		&model.ProductCategory{Name: "服饰", Sort: 1, Status: 1},
		&model.Product{CategoryID: 1, Title: "T恤", Status: 1},
		&model.SKU{ProductID: 1, Specs: `{"尺码":"M"}`, PriceCents: 9900, Stock: 5, Code: "T1"},
		&model.SKU{ProductID: 1, Specs: `{"尺码":"L"}`, PriceCents: 9900, Stock: 5, Code: "T2"},
		&model.Config{Key: "ai_daily_free_limit", Value: "3"},
		&model.Config{Key: "work_review_enabled", Value: "false"},
		&model.Config{Key: "ai_api_key", Value: ""},
		&model.Config{Key: "ai_base_url", Value: "https://tokenhub.tencentmaas.com"},
		&model.Config{Key: "ai_model", Value: "hy-image-v3.5-preview"},
		&model.Config{Key: "ai_style", Value: "messages"},
	}
	for _, s := range seed {
		if err := db.Create(s).Error; err != nil {
			t.Fatal(err)
		}
	}
	h := &Handlers{DB: db}
	r := gin.New()
	ag := r.Group("/admin")
	ag.POST("/skus", h.adminSKUCreate)
	ag.DELETE("/skus/:id", h.adminSKUDelete)
	ag.PUT("/configs", h.adminConfigUpdate)
	return db, r
}

func doReq(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestAdminSKUDelete_ProtectOrdered(t *testing.T) {
	db, r := newM2TestEnv(t)
	// SKU1 已被订单引用
	if err := db.Create(&model.OrderItem{OrderID: 1, SKUID: 1, Title: "T恤", Specs: `{}`, PriceCents: 9900, Qty: 1}).Error; err != nil {
		t.Fatal(err)
	}
	w := doReq(r, "DELETE", "/admin/skus/1", "")
	if w.Code != 400 {
		t.Fatalf("有订单引用的 SKU 删除应返回 400: %d %s", w.Code, w.Body.String())
	}
	// 无引用的 SKU2 可删
	w = doReq(r, "DELETE", "/admin/skus/2", "")
	if w.Code != 200 {
		t.Fatalf("无引用 SKU 应可删除: %d %s", w.Code, w.Body.String())
	}
	var n int64
	db.Model(&model.SKU{}).Where("id = ?", 2).Count(&n)
	if n != 0 {
		t.Fatal("SKU2 未被删除")
	}
}

func TestAdminSKUCreate_NegativeStock(t *testing.T) {
	_, r := newM2TestEnv(t)
	w := doReq(r, "POST", "/admin/skus", `{"product_id":1,"specs":"{}","price_cents":100,"stock":-1}`)
	if w.Code != 400 {
		t.Fatalf("负库存应拒绝: %d", w.Code)
	}
	w = doReq(r, "POST", "/admin/skus", `{"product_id":1,"specs":"{}","price_cents":100,"stock":3}`)
	if w.Code != 200 {
		t.Fatalf("正常创建应通过: %d %s", w.Code, w.Body.String())
	}
}

func TestAdminConfigUpdate(t *testing.T) {
	db, r := newM2TestEnv(t)
	// 白名单外拒绝
	w := doReq(r, "PUT", "/admin/configs", `{"key":"evil","value":"1"}`)
	if w.Code != 400 {
		t.Fatalf("未知配置键应拒绝: %d", w.Code)
	}
	// 非法值拒绝
	w = doReq(r, "PUT", "/admin/configs", `{"key":"ai_daily_free_limit","value":"abc"}`)
	if w.Code != 400 {
		t.Fatalf("非法值应拒绝: %d", w.Code)
	}
	// 合法更新
	w = doReq(r, "PUT", "/admin/configs", `{"key":"ai_daily_free_limit","value":"5"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("合法更新应通过: %d %s", w.Code, w.Body.String())
	}
	var cfg model.Config
	db.Where("key = ?", "ai_daily_free_limit").First(&cfg)
	if cfg.Value != "5" {
		t.Fatalf("配置未生效: %s", cfg.Value)
	}
}

func TestAdminConfig_AIKeys(t *testing.T) {
	db, r := newM2TestEnv(t)
	// ai_style 非法值拒绝
	w := doReq(r, "PUT", "/admin/configs", `{"key":"ai_style","value":"graphql"}`)
	if w.Code != 400 {
		t.Fatalf("ai_style 非法值应拒绝: %d", w.Code)
	}
	// ai_base_url 非 http 拒绝
	w = doReq(r, "PUT", "/admin/configs", `{"key":"ai_base_url","value":"ftp://x"}`)
	if w.Code != 400 {
		t.Fatalf("ai_base_url 非 http 应拒绝: %d", w.Code)
	}
	// ai_api_key 允许空值（清空 = 停用）
	w = doReq(r, "PUT", "/admin/configs", `{"key":"ai_api_key","value":""}`)
	if w.Code != http.StatusOK {
		t.Fatalf("清空 ai_api_key 应通过: %d %s", w.Code, w.Body.String())
	}
	// 写入密钥（自动去首尾空白）
	w = doReq(r, "PUT", "/admin/configs", `{"key":"ai_api_key","value":"  sk-test1234567890abcd  "}`)
	if w.Code != http.StatusOK {
		t.Fatalf("写入 ai_api_key 应通过: %d %s", w.Code, w.Body.String())
	}
	var cfg model.Config
	db.Where("key = ?", "ai_api_key").First(&cfg)
	if cfg.Value != "sk-test1234567890abcd" {
		t.Fatalf("密钥未正确保存: %q", cfg.Value)
	}
	// 列表脱敏：不含完整密钥
	h := &Handlers{DB: db}
	gin.SetMode(gin.TestMode)
	r2 := gin.New()
	r2.GET("/admin/configs", h.adminConfigList)
	w = doReq(r2, "GET", "/admin/configs", "")
	if strings.Contains(w.Body.String(), "sk-test1234567890abcd") {
		t.Fatalf("列表泄露完整密钥: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "sk-te****abcd") {
		t.Fatalf("列表未返回脱敏形式: %s", w.Body.String())
	}
}
