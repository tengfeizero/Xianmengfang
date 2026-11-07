package api

import (
	"image"
	"image/color"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"xmf/internal/imageutil"
	"xmf/internal/model"
	"xmf/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 一对多归属：元素必须挂在具体底板下，用户端只能取本底板元素，合成时也不能跨底板混搭。
func newCollageTestEnv(t *testing.T) (*gorm.DB, *gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(model.AllMigrations()...); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	// 底板与元素的占位 PNG（合成接口会真的读文件）
	mk := func(name string) {
		img := image.NewRGBA(image.Rect(0, 0, 20, 20))
		img.Set(1, 1, color.RGBA{R: 200, A: 255})
		if err := imageutil.SavePNG(img, filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []string{"board_a.png", "board_b.png", "el_a.png", "el_b.png"} {
		mk(n)
	}

	db.Create(&model.Board{Name: "A", URL: "/static/board_a.png", Width: 400, Height: 500, Status: 1})
	db.Create(&model.Board{Name: "B", URL: "/static/board_b.png", Width: 400, Height: 500, Status: 1})
	db.Create(&model.Element{BoardID: 1, Category: "hat", Name: "A帽", ImageURL: "/static/el_a.png", Anchor: `{"x":0.5,"y":0.5,"scale":1}`, Status: 1})
	db.Create(&model.Element{BoardID: 2, Category: "hat", Name: "B帽", ImageURL: "/static/el_b.png", Anchor: `{"x":0.5,"y":0.5,"scale":1}`, Status: 1})

	h := &Handlers{DB: db, UploadsDir: dir}
	r := gin.New()
	ug := r.Group("/user")
	ug.GET("/collage/elements", h.collageElements)
	ug.POST("/collage/compose", func(c *gin.Context) {
		c.Set("claims", &service.Claims{UserID: 1})
		c.Next()
	}, h.collageCompose)

	ag := r.Group("/admin")
	ag.GET("/boards", h.adminBoardList)
	ag.DELETE("/boards/:id", h.adminBoardDelete)
	ag.POST("/elements", h.adminElementCreate)
	ag.DELETE("/elements/:id", h.adminElementDelete)
	return db, r, dir
}

func TestCollageElementsRequiresBoard(t *testing.T) {
	_, r, _ := newCollageTestEnv(t)

	if w := doReq(r, http.MethodGet, "/user/collage/elements", ""); w.Code != 400 {
		t.Fatalf("缺 board_id 应 400，实际 %d：%s", w.Code, w.Body.String())
	}

	w := doReq(r, http.MethodGet, "/user/collage/elements?board_id=1", "")
	if w.Code != 200 {
		t.Fatalf("取底板 1 元素应 200，实际 %d：%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "A帽") || strings.Contains(body, "B帽") {
		t.Fatalf("只应返回底板 1 的元素，实际：%s", body)
	}
}

func TestComposeRejectsElementOfAnotherBoard(t *testing.T) {
	_, r, _ := newCollageTestEnv(t)

	// 底板 1 + 底板 2 的元素（id=2）→ 应 400
	cross := `{"board_id":1,"placements":[{"element_id":2,"x":0.5,"y":0.5,"scale":1}]}`
	w := doReq(r, http.MethodPost, "/user/collage/compose", cross)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "不属于当前底板") {
		t.Fatalf("跨底板元素应 400，实际 %d：%s", w.Code, w.Body.String())
	}

	// 同底板 → 200 并落作品
	ok := `{"board_id":1,"placements":[{"element_id":1,"x":0.5,"y":0.5,"scale":1}]}`
	if w := doReq(r, http.MethodPost, "/user/collage/compose", ok); w.Code != 200 {
		t.Fatalf("同底板合成应 200，实际 %d：%s", w.Code, w.Body.String())
	}
}

func TestBoardDeleteBlockedByActiveElements(t *testing.T) {
	db, r, _ := newCollageTestEnv(t)

	// 底板 1 下仍有启用元素 → 409
	w := doReq(r, http.MethodDelete, "/admin/boards/1", "")
	if w.Code != 409 || !strings.Contains(w.Body.String(), "先下架元素") {
		t.Fatalf("底板下还有启用元素应 409，实际 %d：%s", w.Code, w.Body.String())
	}
	// 先下架元素，再下架底板 → 200，底板 status 变 0
	if w := doReq(r, http.MethodDelete, "/admin/elements/1", ""); w.Code != 200 {
		t.Fatalf("下架元素应 200，实际 %d：%s", w.Code, w.Body.String())
	}
	if w := doReq(r, http.MethodDelete, "/admin/boards/1", ""); w.Code != 200 {
		t.Fatalf("元素下架后底板应可下架，实际 %d：%s", w.Code, w.Body.String())
	}
	var b model.Board
	db.First(&b, 1)
	if b.Status != 0 {
		t.Fatalf("底板 1 应变为下架，实际 status=%d", b.Status)
	}
}

func TestAdminBoardListCountsElements(t *testing.T) {
	_, r, _ := newCollageTestEnv(t)
	w := doReq(r, http.MethodGet, "/admin/boards", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"element_count":1`) {
		t.Fatalf("底板列表应带 element_count，实际 %d：%s", w.Code, w.Body.String())
	}
}

func TestAdminElementCreateRequiresBoard(t *testing.T) {
	db, r, _ := newCollageTestEnv(t)

	if w := doReq(r, http.MethodPost, "/admin/elements", `{"category":"hat","name":"无主帽子","image_url":"/static/el_a.png"}`); w.Code != 400 {
		t.Fatalf("缺 board_id 应 400，实际 %d：%s", w.Code, w.Body.String())
	}
	if w := doReq(r, http.MethodPost, "/admin/elements", `{"board_id":999,"category":"hat","name":"x","image_url":"/static/el_a.png"}`); w.Code != 400 {
		t.Fatalf("底板不存在应 400，实际 %d：%s", w.Code, w.Body.String())
	}
	if w := doReq(r, http.MethodPost, "/admin/elements", `{"board_id":2,"category":"hat","name":"B新帽","image_url":"/static/el_b.png"}`); w.Code != 200 {
		t.Fatalf("正常创建应 200，实际 %d：%s", w.Code, w.Body.String())
	}
	var n int64
	db.Model(&model.Element{}).Where("board_id = 2 AND name = ?", "B新帽").Count(&n)
	if n != 1 {
		t.Fatalf("底板 2 应新增 1 个元素，实际 %d", n)
	}
}
