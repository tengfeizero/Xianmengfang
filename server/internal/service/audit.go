package service

import (
	"encoding/json"
	"strings"

	"xmf/internal/model"

	"gorm.io/gorm"
)

// AuditService 敏感操作留痕。写日志失败不影响业务（吞掉错误）。
type AuditService struct {
	DB *gorm.DB
}

type AuditEntry struct {
	AdminID    int64
	AdminName  string
	Action     string
	TargetType string
	TargetID   string
	Method     string
	Path       string
	IP         string
	UserAgent  string
	Detail     string
	Result     string // ok / fail / deny
	StatusCode int
	ErrMsg     string
}

const auditDetailLimit = 800

// Log 写入一条审计日志
func (s *AuditService) Log(e AuditEntry) {
	if s == nil || s.DB == nil {
		return
	}
	ua := e.UserAgent
	if len(ua) > 255 {
		ua = ua[:255]
	}
	if e.Result == "" {
		e.Result = "ok"
	}
	_ = s.DB.Create(&model.AuditLog{
		AdminID:    e.AdminID,
		AdminName:  e.AdminName,
		Action:     e.Action,
		TargetType: e.TargetType,
		TargetID:   e.TargetID,
		Method:     e.Method,
		Path:       e.Path,
		IP:         e.IP,
		UserAgent:  ua,
		Detail:     e.Detail,
		Result:     e.Result,
		StatusCode: e.StatusCode,
		ErrMsg:     truncate(e.ErrMsg, 255),
	}).Error
}

// AuditQuery 日志查询条件
type AuditQuery struct {
	AdminID int64
	Action  string // 前缀匹配，如 order
	Result  string
	Keyword string // 匹配路径 / 目标 / 明细
	Page    int
	Size    int
}

type AuditPage struct {
	List  []model.AuditLog `json:"list"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

func (s *AuditService) Query(q AuditQuery) (*AuditPage, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 100 {
		q.Size = 20
	}
	db := s.DB.Model(&model.AuditLog{})
	if q.AdminID > 0 {
		db = db.Where("admin_id = ?", q.AdminID)
	}
	if q.Action != "" {
		db = db.Where("action LIKE ?", q.Action+"%")
	}
	if q.Result != "" {
		db = db.Where("result = ?", q.Result)
	}
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		db = db.Where("path LIKE ? OR target_id LIKE ? OR detail LIKE ? OR admin_name LIKE ?", like, like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.AuditLog
	if err := db.Order("id desc").Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&list).Error; err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.AuditLog{}
	}
	return &AuditPage{List: list, Total: total, Page: q.Page, Size: q.Size}, nil
}

// MaskDetail 请求体脱敏：密码/密钥/令牌类字段一律替换为 ***，并截断长度。
// 审计日志可查但不该成为泄密面。
func MaskDetail(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		// 非 JSON（如 multipart）只留长度信息
		return truncate("<非JSON请求体>", auditDetailLimit)
	}
	maskValue(v)
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return truncate(string(b), auditDetailLimit)
}

func maskValue(v any) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			lk := strings.ToLower(k)
			if strings.Contains(lk, "password") || strings.Contains(lk, "secret") ||
				strings.Contains(lk, "token") || strings.Contains(lk, "api_key") ||
				strings.Contains(lk, "apikey") || lk == "key" || strings.Contains(lk, "pwd") {
				t[k] = "***"
				continue
			}
			maskValue(val)
		}
	case []any:
		for _, item := range t {
			maskValue(item)
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	// 按 rune 截断，避免切坏中文
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
