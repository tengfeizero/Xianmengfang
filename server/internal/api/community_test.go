package api

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"xmf/internal/model"
	"xmf/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 作品社区与积分（v1.7）：发布、点赞、广场排序、下单作品权限、订单完成结算、积分抵扣。
func newCommunityTestEnv(t *testing.T) (*gorm.DB, *gin.Engine, *service.OrderService) {
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
	db.Create(&model.User{Nickname: "作者", Status: 1})
	db.Create(&model.User{Nickname: "买家", Status: 1})
	db.Create(&model.User{Nickname: "路人", Status: 1})
	// 一个可下单的商品
	db.Create(&model.Product{Title: "定制T恤", Images: `["/static/p.png"]`, Status: 1})
	db.Create(&model.SKU{ProductID: 1, Specs: `{"尺码":"M"}`, PriceCents: 9900, Stock: 100})
	// 免费额度 1 张，方便验证「用完后扣积分」
	db.Create(&model.Config{Key: "ai_daily_free_limit", Value: "1"})

	orders := &service.OrderService{DB: db}
	community := &service.CommunityService{DB: db}
	h := &Handlers{
		DB: db, Secret: "test",
		Orders:     orders,
		Carts:      &service.CartService{DB: db, Orders: orders},
		AI:         &service.AIService{DB: db, Provider: service.MockProvider{}, UploadsDir: dir},
		Community:  community,
		UploadsDir: dir,
	}
	r := gin.New()
	ug := r.Group("/user")
	ug.Use(func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.GetHeader("X-Test-User"), 10, 64)
		c.Set("claims", &service.Claims{UserID: id})
		c.Next()
	})
	ug.GET("/works", h.workList)
	ug.PUT("/works/:id/publish", h.workPublish)
	ug.POST("/works/:id/like", h.workLike)
	ug.DELETE("/works/:id/like", h.workUnlike)
	ug.GET("/plaza", h.workPublicList)
	ug.GET("/points", h.userPoints)
	ug.GET("/ai/quota", h.userAIQuota)
	ug.POST("/ai/tasks", h.aiCreateTask)
	ug.POST("/cart", h.cartAdd)
	ug.POST("/cart/checkout", h.cartCheckout)
	ug.POST("/orders/:id/confirm", h.orderConfirm)
	return db, r, orders
}

func doAs(r *gin.Engine, uid int64, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", strconv.FormatInt(uid, 10))
	r.ServeHTTP(w, req)
	return w
}

// newWork 造一个作品（作者默认 user1）
func newWork(t *testing.T, db *gorm.DB, userID int64, prompt string) *model.Work {
	t.Helper()
	w := &model.Work{UserID: userID, Type: "ai_text", ImageURL: "/static/w.png", ThumbnailURL: "/static/w.png", Prompt: prompt, Status: 1}
	if err := db.Create(w).Error; err != nil {
		t.Fatal(err)
	}
	return w
}

func publish(t *testing.T, r *gin.Engine, uid, workID int64, isPublic int) {
	t.Helper()
	if w := doAs(r, uid, http.MethodPut, "/user/works/"+strconv.FormatInt(workID, 10)+"/publish",
		`{"is_public":`+strconv.Itoa(isPublic)+`}`); w.Code != 200 {
		t.Fatalf("发布操作应 200，实际 %d：%s", w.Code, w.Body.String())
	}
}

func TestPublishControlsPlazaVisibility(t *testing.T) {
	db, r, _ := newCommunityTestEnv(t)
	w := newWork(t, db, 1, "扎双马尾的小姑娘，粉色背景")

	// 默认私密 → 广场看不到，自己的作品库能看到 is_public=0
	if out := doAs(r, 2, http.MethodGet, "/user/plaza", ""); !strings.Contains(out.Body.String(), `"total":0`) {
		t.Fatalf("未发布的作品不应出现在广场：%s", out.Body.String())
	}
	publish(t, r, 1, w.ID, 1)
	out := doAs(r, 2, http.MethodGet, "/user/plaza", "")
	if out.Code != 200 || !strings.Contains(out.Body.String(), `"total":1`) {
		t.Fatalf("发布后应出现在广场：%d %s", out.Code, out.Body.String())
	}
	if !strings.Contains(out.Body.String(), "扎双马尾") {
		t.Fatalf("发布后应公开提示词：%s", out.Body.String())
	}
	// 别人的作品 is_mine=false
	if strings.Contains(out.Body.String(), `"is_mine":true`) {
		t.Fatalf("他人作品不应标记 is_mine：%s", out.Body.String())
	}

	// 取消发布 → 广场消失，但计数保留
	like := doAs(r, 2, http.MethodPost, "/user/works/"+strconv.FormatInt(w.ID, 10)+"/like", "")
	if like.Code != 200 {
		t.Fatalf("点赞应 200，实际 %d：%s", like.Code, like.Body.String())
	}
	publish(t, r, 1, w.ID, 0)
	if out := doAs(r, 2, http.MethodGet, "/user/plaza", ""); !strings.Contains(out.Body.String(), `"total":0`) {
		t.Fatalf("取消发布后不应出现在广场：%s", out.Body.String())
	}
	var got model.Work
	db.First(&got, w.ID)
	if got.LikeCount != 1 {
		t.Fatalf("取消发布不应清空点赞数，实际 %d", got.LikeCount)
	}
}

func TestLikeRulesAndIdempotency(t *testing.T) {
	db, r, _ := newCommunityTestEnv(t)
	w := newWork(t, db, 1, "")
	path := "/user/works/" + strconv.FormatInt(w.ID, 10) + "/like"

	// 未发布 → 404
	if out := doAs(r, 2, http.MethodPost, path, ""); out.Code != 404 {
		t.Fatalf("未发布作品点赞应 404，实际 %d：%s", out.Code, out.Body.String())
	}
	publish(t, r, 1, w.ID, 1)

	// 自赞 → 400
	if out := doAs(r, 1, http.MethodPost, path, ""); out.Code != 400 || !strings.Contains(out.Body.String(), "自己") {
		t.Fatalf("给自己点赞应 400，实际 %d：%s", out.Code, out.Body.String())
	}

	// 重复点赞幂等：计数只加 1
	for i := 0; i < 2; i++ {
		if out := doAs(r, 2, http.MethodPost, path, ""); out.Code != 200 {
			t.Fatalf("点赞应 200，实际 %d：%s", out.Code, out.Body.String())
		}
	}
	// 重复取消幂等：计数不会变负
	for i := 0; i < 2; i++ {
		if out := doAs(r, 2, http.MethodDelete, path, ""); out.Code != 200 {
			t.Fatalf("取消点赞应 200，实际 %d：%s", out.Code, out.Body.String())
		}
	}
	var got model.Work
	db.First(&got, w.ID)
	if got.LikeCount != 0 {
		t.Fatalf("取消后点赞数应为 0，实际 %d", got.LikeCount)
	}
	// 广场 liked 字段跟随
	publish(t, r, 1, w.ID, 1)
	doAs(r, 3, http.MethodPost, path, "")
	if out := doAs(r, 3, http.MethodGet, "/user/plaza", ""); !strings.Contains(out.Body.String(), `"liked":true`) {
		t.Fatalf("点赞后广场应返回 liked=true：%s", out.Body.String())
	}
}

func TestPlazaSortByUses(t *testing.T) {
	db, r, _ := newCommunityTestEnv(t)
	a := newWork(t, db, 1, "作品A")
	b := newWork(t, db, 2, "作品B")
	publish(t, r, 1, a.ID, 1)
	publish(t, r, 2, b.ID, 1)
	db.Model(&model.Work{}).Where("id = ?", a.ID).Update("use_count", 1)
	db.Model(&model.Work{}).Where("id = ?", b.ID).Update("use_count", 9)

	out := doAs(r, 3, http.MethodGet, "/user/plaza?sort=uses", "")
	if out.Code != 200 {
		t.Fatalf("使用量榜应 200，实际 %d", out.Code)
	}
	body := out.Body.String()
	if strings.Index(body, "作品B") > strings.Index(body, "作品A") {
		t.Fatalf("使用量高的应排前面：%s", body)
	}
}

func TestOrderWorkPermission(t *testing.T) {
	db, r, _ := newCommunityTestEnv(t)
	w := newWork(t, db, 1, "别人的图")
	workID := strconv.FormatInt(w.ID, 10)

	// 他人私密作品：不能加购
	if out := doAs(r, 2, http.MethodPost, "/user/cart", `{"sku_id":1,"qty":1,"work_id":`+workID+`}`); out.Code != 400 {
		t.Fatalf("他人私密作品不能加购，实际 %d：%s", out.Code, out.Body.String())
	}
	publish(t, r, 1, w.ID, 1)
	// 已公开但没点赞：仍不能加购（点赞是借用前置条件）
	if out := doAs(r, 2, http.MethodPost, "/user/cart", `{"sku_id":1,"qty":1,"work_id":`+workID+`}`); out.Code != 400 {
		t.Fatalf("未点赞的公开作品不能加购，实际 %d：%s", out.Code, out.Body.String())
	}
	doAs(r, 2, http.MethodPost, "/user/works/"+workID+"/like", "")
	if out := doAs(r, 2, http.MethodPost, "/user/cart", `{"sku_id":1,"qty":1,"work_id":`+workID+`}`); out.Code != 200 {
		t.Fatalf("点赞后应可加购，实际 %d：%s", out.Code, out.Body.String())
	}
	// 作者取消发布 → 结算时拦截（不静默换图）
	publish(t, r, 1, w.ID, 0)
	if out := doAs(r, 2, http.MethodPost, "/user/cart/checkout", `{"address_id":0}`); out.Code != 400 {
		t.Fatalf("作品取消发布后结算应 400，实际 %d：%s", out.Code, out.Body.String())
	}
}

func TestOrderDoneSettlesUseCountAndPoints(t *testing.T) {
	db, r, orders := newCommunityTestEnv(t)
	w := newWork(t, db, 1, "小熊猫")
	publish(t, r, 1, w.ID, 1)
	doAs(r, 2, http.MethodPost, "/user/works/"+strconv.FormatInt(w.ID, 10)+"/like", "")

	// 买家 2 用作者 1 的作品买 2 件
	res, err := orders.Create(2, []service.ItemInput{{SKUID: 1, WorkID: &w.ID, Qty: 2}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// 未完成前不结算
	var before model.Work
	db.First(&before, w.ID)
	if before.UseCount != 0 {
		t.Fatalf("订单未完成不应累计使用次数，实际 %d", before.UseCount)
	}
	// 走完整状态机到已发货，再由用户确认收货
	if err := orders.Accept(res.Order.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := orders.Pack(res.Order.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := orders.ReadyShip(res.Order.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := orders.Ship(res.Order.ID, 1, "顺丰速运", "SF1"); err != nil {
		t.Fatal(err)
	}
	if out := doAs(r, 2, http.MethodPost, "/user/orders/"+strconv.FormatInt(res.Order.ID, 10)+"/confirm", ""); out.Code != 200 {
		t.Fatalf("确认收货应 200，实际 %d：%s", out.Code, out.Body.String())
	}
	var after model.Work
	db.First(&after, w.ID)
	if after.UseCount != 2 {
		t.Fatalf("使用次数应按件数累加（2），实际 %d", after.UseCount)
	}
	var author model.User
	db.First(&author, 1)
	if author.Points != 2 {
		t.Fatalf("作者应得 2 积分，实际 %d", author.Points)
	}
	// 重复确认收货被状态机拦住，不会重复结算
	if out := doAs(r, 2, http.MethodPost, "/user/orders/"+strconv.FormatInt(res.Order.ID, 10)+"/confirm", ""); out.Code == 200 {
		t.Fatalf("重复确认收货不应成功")
	}
	db.First(&author, 1)
	if author.Points != 2 {
		t.Fatalf("重复确认不应重复加积分，实际 %d", author.Points)
	}

	// 自用也奖励（产品确认口径）
	self := newWork(t, db, 1, "")
	res2, err := orders.Create(1, []service.ItemInput{{SKUID: 1, WorkID: &self.ID, Qty: 3}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	orders.Accept(res2.Order.ID, 1)
	orders.Pack(res2.Order.ID, 1)
	orders.ReadyShip(res2.Order.ID, 1)
	orders.Ship(res2.Order.ID, 1, "顺丰速运", "SF2")
	if out := doAs(r, 1, http.MethodPost, "/user/orders/"+strconv.FormatInt(res2.Order.ID, 10)+"/confirm", ""); out.Code != 200 {
		t.Fatalf("自用订单确认收货应 200，实际 %d：%s", out.Code, out.Body.String())
	}
	db.First(&author, 1)
	if author.Points != 5 {
		t.Fatalf("自用 3 件应再得 3 积分（合计 5），实际 %d", author.Points)
	}
}

func TestAIPointsFallback(t *testing.T) {
	db, r, _ := newCommunityTestEnv(t)
	// 作者已有 2 积分：免费 1 张用完后，第 2、3 张扣积分，第 4 张 429
	db.Model(&model.User{}).Where("id = ?", 1).Update("points", 2)

	body := `{"type":"ai_text","prompt":"水彩柴犬"}`
	for i := 0; i < 3; i++ {
		if out := doAs(r, 1, http.MethodPost, "/user/ai/tasks", body); out.Code != 200 {
			t.Fatalf("第 %d 张应成功，实际 %d：%s", i+1, out.Code, out.Body.String())
		}
	}
	var u model.User
	db.First(&u, 1)
	if u.Points != 0 {
		t.Fatalf("3 张里 1 张免费 + 2 张扣积分，余额应为 0，实际 %d", u.Points)
	}
	if out := doAs(r, 1, http.MethodPost, "/user/ai/tasks", body); out.Code != 429 {
		t.Fatalf("积分用完后应 429，实际 %d：%s", out.Code, out.Body.String())
	}
	// 积分抵扣不占免费额度：free_used 仍为 1
	out := doAs(r, 1, http.MethodGet, "/user/ai/quota", "")
	if !strings.Contains(out.Body.String(), `"free_used":1`) || !strings.Contains(out.Body.String(), `"free_left":0`) {
		t.Fatalf("积分换的图不应占用免费额度：%s", out.Body.String())
	}
	// 流水：2 条扣减
	var logs []model.PointLog
	db.Where("user_id = ? AND reason = ?", 1, service.PointReasonAISpend).Find(&logs)
	if len(logs) != 2 {
		t.Fatalf("应有 2 条积分消耗流水，实际 %d", len(logs))
	}
}
