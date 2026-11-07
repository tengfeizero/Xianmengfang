package api

import (
	"strconv"
	"strings"
	"time"

	"xmf/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 三个只读「明细」接口（v2.3）：用户积分明细 / 作品点赞明细 / 作品使用明细。
//
// 约定与后台其他列表一致：一次返回全量（上限 insightLimit 条，防呆保护）+ 前端分页，
// 另外返回 summary 供页面顶部做汇总展示。所有筛选条件都构造在 build() 里，
// 保证「列表 / 总数 / 汇总」三者口径完全一致。

const insightLimit = 2000

func queryID(c *gin.Context, key string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(c.Query(key)), 10, 64)
	return v
}

// pointReasonText 积分原因码 → 人话（与小程序 utils/util.js 的 POINT_REASON 对齐）
func pointReasonText(code string) string {
	switch code {
	case "work_used":
		return "作品被使用"
	case "ai_spend":
		return "AI 生成抵扣"
	case "admin_adjust":
		return "后台调整"
	}
	return code
}

// adminPointLogs 用户积分明细：GET /admin/point-logs
// 支持 user_id / reason / keyword(昵称) 筛选；返回每笔的变动、变动后余额、关联单据。
func (h *Handlers) adminPointLogs(c *gin.Context) {
	build := func() *gorm.DB {
		q := h.DB.Model(&model.PointLog{})
		if uid := queryID(c, "user_id"); uid > 0 {
			q = q.Where("user_id = ?", uid)
		}
		if r := strings.TrimSpace(c.Query("reason")); r != "" {
			q = q.Where("reason = ?", r)
		}
		if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
			q = q.Where("user_id IN (?)",
				h.DB.Model(&model.User{}).Select("id").Where("nickname LIKE ?", "%"+kw+"%"))
		}
		return q
	}

	var logs []model.PointLog
	if err := build().Order("id desc").Limit(insightLimit).Find(&logs).Error; err != nil {
		fail(c, 500, "获取积分明细失败，请稍后再试")
		return
	}

	var total int64
	if err := build().Count(&total).Error; err != nil {
		total = int64(len(logs))
	}
	var sum struct {
		InSum  int64
		OutSum int64
	}
	build().Select(`COALESCE(SUM(CASE WHEN delta > 0 THEN delta ELSE 0 END), 0) AS in_sum,
		COALESCE(SUM(CASE WHEN delta < 0 THEN -delta ELSE 0 END), 0) AS out_sum`).Scan(&sum)

	// 批量补昵称，避免前端只看到 user_id
	ids := make([]int64, 0, len(logs))
	seen := map[int64]bool{}
	for _, l := range logs {
		if !seen[l.UserID] {
			seen[l.UserID] = true
			ids = append(ids, l.UserID)
		}
	}
	nick := map[int64]string{}
	if len(ids) > 0 {
		var us []model.User
		if err := h.DB.Where("id IN ?", ids).Find(&us).Error; err == nil {
			for _, u := range us {
				nick[u.ID] = u.Nickname
			}
		}
	}

	rows := make([]gin.H, 0, len(logs))
	for _, l := range logs {
		rows = append(rows, gin.H{
			"id": l.ID, "user_id": l.UserID, "nickname": nick[l.UserID],
			"delta": l.Delta, "balance": l.Balance,
			"reason": l.Reason, "reason_text": pointReasonText(l.Reason),
			"ref_type": l.RefType, "ref_id": l.RefID,
			"remark":     l.Remark,
			"created_at": l.CreatedAt,
		})
	}
	ok(c, gin.H{
		"items":   rows,
		"total":   total,
		"summary": gin.H{"in": sum.InSum, "out": sum.OutSum},
	})
}

// adminWorkLikes 作品点赞明细：GET /admin/work-likes
// 支持 work_id / user_id / keyword(昵称或提示词) 筛选；返回点赞人 + 被赞作品 + 作者。
func (h *Handlers) adminWorkLikes(c *gin.Context) {
	build := func() *gorm.DB {
		q := h.DB.Model(&model.WorkLike{})
		if wid := queryID(c, "work_id"); wid > 0 {
			q = q.Where("work_id = ?", wid)
		}
		if uid := queryID(c, "user_id"); uid > 0 {
			q = q.Where("user_id = ?", uid)
		}
		if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
			like := "%" + kw + "%"
			q = q.Where("(user_id IN (?) OR work_id IN (?))",
				h.DB.Model(&model.User{}).Select("id").Where("nickname LIKE ?", like),
				h.DB.Model(&model.Work{}).Select("id").Where("prompt LIKE ?", like))
		}
		return q
	}

	var likes []model.WorkLike
	if err := build().Order("id desc").Limit(insightLimit).Find(&likes).Error; err != nil {
		fail(c, 500, "获取点赞明细失败，请稍后再试")
		return
	}
	var total int64
	if err := build().Count(&total).Error; err != nil {
		total = int64(len(likes))
	}

	// 作品与用户各查一次，避免 N+1
	workIDs := make([]int64, 0, len(likes))
	userIDs := make([]int64, 0, len(likes))
	seenW, seenU := map[int64]bool{}, map[int64]bool{}
	for _, l := range likes {
		if !seenW[l.WorkID] {
			seenW[l.WorkID] = true
			workIDs = append(workIDs, l.WorkID)
		}
		if !seenU[l.UserID] {
			seenU[l.UserID] = true
			userIDs = append(userIDs, l.UserID)
		}
	}
	workMap := map[int64]model.Work{}
	if len(workIDs) > 0 {
		var ws []model.Work
		if err := h.DB.Where("id IN ?", workIDs).Find(&ws).Error; err == nil {
			for _, w := range ws {
				workMap[w.ID] = w
			}
			for _, w := range ws {
				if !seenU[w.UserID] {
					seenU[w.UserID] = true
					userIDs = append(userIDs, w.UserID)
				}
			}
		}
	}
	nick := map[int64]string{}
	if len(userIDs) > 0 {
		var us []model.User
		if err := h.DB.Where("id IN ?", userIDs).Find(&us).Error; err == nil {
			for _, u := range us {
				nick[u.ID] = u.Nickname
			}
		}
	}

	rows := make([]gin.H, 0, len(likes))
	for _, l := range likes {
		w := workMap[l.WorkID]
		rows = append(rows, gin.H{
			"id": l.ID,
			"work": gin.H{
				"id": w.ID, "type": w.Type, "image_url": w.ImageURL,
				"thumbnail_url": w.ThumbnailURL, "prompt": w.Prompt,
				"status": w.Status, "is_public": w.IsPublic,
				"author_id": w.UserID, "author_name": nick[w.UserID],
			},
			"user_id": l.UserID, "nickname": nick[l.UserID],
			"created_at": l.CreatedAt,
		})
	}
	ok(c, gin.H{"items": rows, "total": total})
}

// adminWorkUsages 作品使用明细：GET /admin/work-usages
//
// 数据源是订单明细里绑定了作品的条目（OrderItem.WorkID），也就是「这件定制品用的是谁的作品」。
// 说明：Work.UseCount 只在订单完成时按件数结算（见 CommunityService），
// 而这里展示的是「下单即产生」的使用证据，落库时间、订单号、买家都能追。
func (h *Handlers) adminWorkUsages(c *gin.Context) {
	build := func() *gorm.DB {
		q := h.DB.Table("order_items").
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("order_items.work_id IS NOT NULL AND order_items.work_id > 0")
		if wid := queryID(c, "work_id"); wid > 0 {
			q = q.Where("order_items.work_id = ?", wid)
		}
		if uid := queryID(c, "user_id"); uid > 0 {
			q = q.Where("orders.user_id = ?", uid)
		}
		if st := strings.TrimSpace(c.Query("status")); st != "" {
			q = q.Where("orders.status = ?", st)
		}
		if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
			like := "%" + kw + "%"
			q = q.Where("(orders.order_no LIKE ? OR orders.user_id IN (?))", like,
				h.DB.Model(&model.User{}).Select("id").Where("nickname LIKE ?", like))
		}
		return q
	}

	type usageRow struct {
		ID        int64 `gorm:"column:id"`
		OrderID   int64
		OrderNo   string
		WorkID    int64
		Title     string
		Specs     string
		Qty       int
		UserID    int64
		Status    string
		CreatedAt time.Time
	}
	var rows []usageRow
	err := build().Select(`order_items.id AS id,
			order_items.order_id AS order_id,
			orders.order_no AS order_no,
			order_items.work_id AS work_id,
			order_items.title AS title,
			order_items.specs AS specs,
			order_items.qty AS qty,
			orders.user_id AS user_id,
			orders.status AS status,
			order_items.created_at AS created_at`).
		Order("order_items.id desc").Limit(insightLimit).Scan(&rows).Error
	if err != nil {
		fail(c, 500, "获取使用明细失败，请稍后再试")
		return
	}

	var agg struct {
		Cnt    int64
		QtySum int64
	}
	build().Select("COUNT(*) AS cnt, COALESCE(SUM(order_items.qty), 0) AS qty_sum").Scan(&agg)

	// 补作品图 + 买家昵称
	workIDs := make([]int64, 0, len(rows))
	userIDs := make([]int64, 0, len(rows))
	seenW, seenU := map[int64]bool{}, map[int64]bool{}
	for _, r := range rows {
		if !seenW[r.WorkID] {
			seenW[r.WorkID] = true
			workIDs = append(workIDs, r.WorkID)
		}
		if !seenU[r.UserID] {
			seenU[r.UserID] = true
			userIDs = append(userIDs, r.UserID)
		}
	}
	workMap := map[int64]model.Work{}
	if len(workIDs) > 0 {
		var ws []model.Work
		if err := h.DB.Where("id IN ?", workIDs).Find(&ws).Error; err == nil {
			for _, w := range ws {
				workMap[w.ID] = w
			}
		}
	}
	nick := map[int64]string{}
	if len(userIDs) > 0 {
		var us []model.User
		if err := h.DB.Where("id IN ?", userIDs).Find(&us).Error; err == nil {
			for _, u := range us {
				nick[u.ID] = u.Nickname
			}
		}
	}

	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		w := workMap[r.WorkID]
		list = append(list, gin.H{
			"id": r.ID, "order_id": r.OrderID, "order_no": r.OrderNo,
			"work_id": r.WorkID, "work_image": w.ThumbnailURL, "work_image_url": w.ImageURL,
			"work_type": w.Type, "work_prompt": w.Prompt, "author_id": w.UserID,
			"title": r.Title, "specs": r.Specs, "qty": r.Qty,
			"user_id": r.UserID, "nickname": nick[r.UserID],
			"status": r.Status, "created_at": r.CreatedAt,
		})
	}
	ok(c, gin.H{
		"items":   list,
		"total":   agg.Cnt,
		"summary": gin.H{"qty": agg.QtySum},
	})
}
