package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"xmf/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 明细接口测试：造一份最小数据，校验「列表 / 总数 / 汇总 / 筛选」四处口径一致。
func newInsightTestEnv(t *testing.T) (*gorm.DB, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(model.AllMigrations()...); err != nil {
		t.Fatal(err)
	}

	users := []model.User{
		{Nickname: "阿岚", OpenID: "o_a", Status: 1, Points: 12},
		{Nickname: "桃桃", OpenID: "o_b", Status: 1, Points: 5},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	// 阿岚的一件作品（公开）
	w := model.Work{UserID: users[0].ID, Type: "ai_text", ImageURL: "a.png", ThumbnailURL: "at.png",
		Prompt: "赛博熊猫", Status: 1, IsPublic: 1}
	if err := db.Create(&w).Error; err != nil {
		t.Fatal(err)
	}

	// 桃桃点赞
	if err := db.Create(&model.WorkLike{WorkID: w.ID, UserID: users[1].ID}).Error; err != nil {
		t.Fatal(err)
	}

	// 积分流水：+2（作品被使用）、-1（AI 抵扣）
	logs := []model.PointLog{
		{UserID: users[0].ID, Delta: 2, Balance: 12, Reason: "work_used", RefType: "order", RefID: 1, Remark: "作品《赛博熊猫》被使用 2 件"},
		{UserID: users[0].ID, Delta: -1, Balance: 10, Reason: "ai_spend", RefType: "ai_task", RefID: 3},
	}
	for i := range logs {
		if err := db.Create(&logs[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	// 订单 + 绑定作品的明细项（qty=2）
	o := model.Order{OrderNo: "XMF20261107001", UserID: users[1].ID, Status: "done", TotalCents: 19800,
		ReceiverName: "桃桃", ReceiverPhone: "13800000000", ReceiverAddress: "上海市"}
	if err := db.Create(&o).Error; err != nil {
		t.Fatal(err)
	}
	wid := w.ID
	item := model.OrderItem{OrderID: o.ID, SKUID: 1, WorkID: &wid, Title: "定制T恤", Specs: `{"尺码":"M"}`, PriceCents: 9900, Qty: 2}
	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	// 另一个未绑定作品的明细项，不应出现在使用明细里
	if err := db.Create(&model.OrderItem{OrderID: o.ID, SKUID: 1, Title: "普通T恤", PriceCents: 9900, Qty: 1}).Error; err != nil {
		t.Fatal(err)
	}

	h := &Handlers{DB: db}
	r := gin.New()
	ag := r.Group("/admin")
	ag.GET("/point-logs", h.adminPointLogs)
	ag.GET("/work-likes", h.adminWorkLikes)
	ag.GET("/work-usages", h.adminWorkUsages)
	return db, r
}

type apiResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func getJSON(t *testing.T, r *gin.Engine, path string) apiResp {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("%s 期望 200，实际 %d：%s", path, w.Code, w.Body.String())
	}
	var resp apiResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败：%v", err)
	}
	return resp
}

func TestAdminPointLogs(t *testing.T) {
	_, r := newInsightTestEnv(t)

	var data struct {
		Items []struct {
			Nickname   string `json:"nickname"`
			Delta      int    `json:"delta"`
			ReasonText string `json:"reason_text"`
		} `json:"items"`
		Total   int64 `json:"total"`
		Summary struct {
			In  int64 `json:"in"`
			Out int64 `json:"out"`
		} `json:"summary"`
	}
	resp := getJSON(t, r, "/admin/point-logs")
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Total != 2 || len(data.Items) != 2 {
		t.Fatalf("期望 2 条流水，实际 total=%d len=%d", data.Total, len(data.Items))
	}
	if data.Summary.In != 2 || data.Summary.Out != 1 {
		t.Fatalf("汇总应为 收入2 / 支出1，实际 %d / %d", data.Summary.In, data.Summary.Out)
	}
	// 最新一条在前：-1（AI 抵扣）
	if data.Items[0].Delta != -1 || data.Items[0].ReasonText != "AI 生成抵扣" {
		t.Fatalf("首条应为 AI 抵扣，实际 %+v", data.Items[0])
	}
	if data.Items[0].Nickname != "阿岚" {
		t.Fatalf("昵称未填充，实际 %q", data.Items[0].Nickname)
	}

	// 原因筛选
	var filtered struct {
		Total int64 `json:"total"`
	}
	f := getJSON(t, r, "/admin/point-logs?reason=work_used")
	if err := json.Unmarshal(f.Data, &filtered); err != nil {
		t.Fatal(err)
	}
	if filtered.Total != 1 {
		t.Fatalf("按 work_used 筛选应为 1 条，实际 %d", filtered.Total)
	}

	// 昵称关键字筛选（命中不存在的昵称应为 0）
	var none struct {
		Total int64 `json:"total"`
	}
	n := getJSON(t, r, "/admin/point-logs?keyword=不存在的人")
	if err := json.Unmarshal(n.Data, &none); err != nil {
		t.Fatal(err)
	}
	if none.Total != 0 {
		t.Fatalf("不存在的昵称应为 0 条，实际 %d", none.Total)
	}
}

func TestAdminWorkLikes(t *testing.T) {
	_, r := newInsightTestEnv(t)

	var data struct {
		Items []struct {
			Nickname string `json:"nickname"`
			Work     struct {
				ID         int64  `json:"id"`
				Prompt     string `json:"prompt"`
				AuthorName string `json:"author_name"`
			} `json:"work"`
		} `json:"items"`
		Total int64 `json:"total"`
	}
	resp := getJSON(t, r, "/admin/work-likes")
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Total != 1 || len(data.Items) != 1 {
		t.Fatalf("期望 1 条点赞，实际 total=%d", data.Total)
	}
	it := data.Items[0]
	if it.Nickname != "桃桃" {
		t.Fatalf("点赞人应为桃桃，实际 %q", it.Nickname)
	}
	if it.Work.Prompt != "赛博熊猫" || it.Work.AuthorName != "阿岚" {
		t.Fatalf("作品/作者信息未补全，实际 %+v", it.Work)
	}

	// 按作品筛选命中，按不存在的作品筛选为 0
	var byWork struct {
		Total int64 `json:"total"`
	}
	bw := getJSON(t, r, "/admin/work-likes?work_id=1")
	if err := json.Unmarshal(bw.Data, &byWork); err != nil {
		t.Fatal(err)
	}
	if byWork.Total != 1 {
		t.Fatalf("按 work_id=1 应为 1 条，实际 %d", byWork.Total)
	}
	var miss struct {
		Total int64 `json:"total"`
	}
	m := getJSON(t, r, "/admin/work-likes?work_id=999")
	if err := json.Unmarshal(m.Data, &miss); err != nil {
		t.Fatal(err)
	}
	if miss.Total != 0 {
		t.Fatalf("不存在的作品应为 0 条，实际 %d", miss.Total)
	}
}

func TestAdminWorkUsages(t *testing.T) {
	_, r := newInsightTestEnv(t)

	var data struct {
		Items []struct {
			OrderNo   string `json:"order_no"`
			WorkID    int64  `json:"work_id"`
			Qty       int    `json:"qty"`
			Nickname  string `json:"nickname"`
			Status    string `json:"status"`
			WorkImage string `json:"work_image"`
		} `json:"items"`
		Total   int64 `json:"total"`
		Summary struct {
			Qty int64 `json:"qty"`
		} `json:"summary"`
	}
	resp := getJSON(t, r, "/admin/work-usages")
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatal(err)
	}
	// 两条明细只有一条绑定了作品
	if data.Total != 1 || len(data.Items) != 1 {
		t.Fatalf("期望 1 条使用明细（只算绑定作品的），实际 total=%d len=%d", data.Total, len(data.Items))
	}
	it := data.Items[0]
	if it.OrderNo != "XMF20261107001" || it.Qty != 2 || it.Nickname != "桃桃" || it.Status != "done" {
		t.Fatalf("明细字段不对：%+v", it)
	}
	if data.Summary.Qty != 2 {
		t.Fatalf("使用件数汇总应为 2，实际 %d", data.Summary.Qty)
	}
	if it.WorkImage != "at.png" {
		t.Fatalf("作品缩略图未补全，实际 %q", it.WorkImage)
	}

	// 订单状态筛选：done 命中，pending_accept 为空
	var pending struct {
		Total int64 `json:"total"`
	}
	p := getJSON(t, r, "/admin/work-usages?status=pending_accept")
	if err := json.Unmarshal(p.Data, &pending); err != nil {
		t.Fatal(err)
	}
	if pending.Total != 0 {
		t.Fatalf("待接单状态应为 0 条，实际 %d", pending.Total)
	}
}
