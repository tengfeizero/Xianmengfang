package service

import (
	"errors"
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xmf/internal/imageutil"
	"xmf/internal/model"

	"gorm.io/gorm"
)

// ErrQuotaExceeded 每日免费配额用尽且积分不足（v1.7 起可用 1 积分换 1 张）
var ErrQuotaExceeded = errors.New("今日免费次数已用完，积分不足；作品被他人使用可获得积分")

// Provider AI 生图供应商接口——换厂商不动业务代码
type Provider interface {
	Name() string
	// Text2Image 文生图，输出 PNG 到 outPath
	Text2Image(prompt, outPath string) error
	// PhotoStyle 照片转风格，srcPath 为已上传的本地文件
	PhotoStyle(srcPath, outPath string) error
}

// MockProvider 本地占位实现：开发期跑通流程，拿到真实 API 密钥后替换
type MockProvider struct{}

func (MockProvider) Name() string { return "mock" }

func (MockProvider) Text2Image(prompt, outPath string) error {
	c := imageutil.NewCanvas(512, 512)
	// 竖向渐变底
	top := color.RGBA{255, 209, 102, 255}
	bottom := color.RGBA{244, 143, 177, 255}
	for y := 0; y < 512; y++ {
		t := float64(y) / 511
		c2 := color.RGBA{
			uint8(float64(top.R)*(1-t) + float64(bottom.R)*t),
			uint8(float64(top.G)*(1-t) + float64(bottom.G)*t),
			uint8(float64(top.B)*(1-t) + float64(bottom.B)*t),
			255,
		}
		for x := 0; x < 512; x++ {
			c.Set(x, y, c2)
		}
	}
	imageutil.FillCircle(c, 256, 256, 120, color.RGBA{255, 255, 255, 230})
	imageutil.FillArc(c, 256, 300, 60, 10, 20, 160, color.RGBA{77, 60, 50, 255})
	imageutil.FillCircle(c, 216, 240, 10, color.RGBA{77, 60, 50, 255})
	imageutil.FillCircle(c, 296, 240, 10, color.RGBA{77, 60, 50, 255})
	return imageutil.SavePNG(c, outPath)
}

func (MockProvider) PhotoStyle(srcPath, outPath string) error {
	// 占位实现：直接复制原图作为"转风格"结果
	src, err := os.Open(srcPath)
	if err != nil {
		return errors.New("找不到上传的图片，请重新上传")
	}
	defer src.Close()
	dst, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = dst.ReadFrom(src)
	return err
}

// AIService AI 生成任务
type AIService struct {
	DB         *gorm.DB
	Provider   Provider // 兜底供应商（无后台密钥配置时使用，通常是 Mock 或环境变量注入的实现）
	UploadsDir string
}

// resolveProvider 后台配置优先：configs.ai_api_key 有值即按腾讯混元调用（base_url/model/style 可配），
// 未配置则回落注入的 Provider（环境变量注入的真实实现或 Mock）。
func (s *AIService) resolveProvider() Provider {
	var cfg model.Config
	if err := s.DB.Where("key = ?", "ai_api_key").First(&cfg).Error; err != nil || strings.TrimSpace(cfg.Value) == "" {
		return s.Provider
	}
	p := &HunyuanProvider{
		APIKey:  strings.TrimSpace(cfg.Value),
		BaseURL: "https://tokenhub.tencentmaas.com",
		Model:   "hy-image-v3.5-preview",
		Style:   "messages",
	}
	// 注意：GORM 复用已填充的结构体做 First 会带上一次主键条件，必须每次用新变量
	var c model.Config
	if err := s.DB.Where("key = ?", "ai_base_url").First(&c).Error; err == nil && strings.TrimSpace(c.Value) != "" {
		p.BaseURL = strings.TrimSpace(c.Value)
	}
	var c3 model.Config
	if err := s.DB.Where("key = ?", "ai_model").First(&c3).Error; err == nil && strings.TrimSpace(c3.Value) != "" {
		p.Model = strings.TrimSpace(c3.Value)
	}
	var c4 model.Config
	if err := s.DB.Where("key = ?", "ai_style").First(&c4).Error; err == nil && strings.TrimSpace(c4.Value) != "" {
		p.Style = strings.TrimSpace(c4.Value)
	}
	var c5 model.Config
	if err := s.DB.Where("key = ?", "ai_secret_id").First(&c5).Error; err == nil && strings.TrimSpace(c5.Value) != "" {
		p.SecretID = strings.TrimSpace(c5.Value)
	}
	return p
}

// CreateTask 创建并同步执行生成任务（MVP 同步；量大后再改异步队列）
func (s *AIService) CreateTask(userID int64, taskType, prompt, photoURL string) (*model.AITask, *model.Work, error) {
	if taskType != "ai_text" && taskType != "ai_photo" {
		return nil, nil, errors.New("不支持的生成类型")
	}
	if taskType == "ai_text" && strings.TrimSpace(prompt) == "" {
		return nil, nil, errors.New("请输入作品描述")
	}
	if taskType == "ai_photo" && photoURL == "" {
		return nil, nil, errors.New("请先上传照片")
	}
	// 每日配额（configs.ai_daily_free_limit，后台可改）
	var cfg model.Config
	if err := s.DB.Where("key = ?", "ai_daily_free_limit").First(&cfg).Error; err != nil {
		return nil, nil, errors.New("系统配置缺失")
	}
	var limit int
	if _, err := fmt.Sscanf(cfg.Value, "%d", &limit); err != nil || limit < 0 {
		return nil, nil, errors.New("系统配置不合法")
	}
	community := &CommunityService{DB: s.DB}
	today := time.Now().Format("2006-01-02")
	used, err := community.TodayQuota(userID)
	if err != nil {
		return nil, nil, err
	}
	// 免费额度用完后改扣积分：1 积分 = 1 张；余额不足直接拒绝，避免白跑一次生成
	payType := "free"
	if used >= limit {
		payType = "point"
		balance, err := community.Balance(userID)
		if err != nil {
			return nil, nil, err
		}
		if balance < 1 {
			return nil, nil, ErrQuotaExceeded
		}
	}

	outName := fmt.Sprintf("ai_%s_%04d.png", time.Now().Format("20060102150405"), rand.Intn(10000))
	outPath := filepath.Join(s.UploadsDir, outName)

	provider := s.resolveProvider()
	task := &model.AITask{
		UserID: userID, Type: taskType, Status: "pending",
		Provider: provider.Name(), QuotaDate: today, PayType: payType,
		Params: prompt,
	}
	if taskType == "ai_photo" {
		srcPath, err := safeUploadPath(s.UploadsDir, photoURL)
		if err != nil {
			return nil, nil, err
		}
		task.Params = photoURL
		if err := provider.PhotoStyle(srcPath, outPath); err != nil {
			return nil, nil, err
		}
	} else {
		if err := provider.Text2Image(prompt, outPath); err != nil {
			return nil, nil, err
		}
	}
	task.Status = "done"
	task.ResultURL = "/static/" + outName
	// 落库与扣积分同一事务：积分扣不动（并发用光）就整笔失败，不白送一张图
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		if payType == "point" {
			return AddPoints(tx, userID, -1, PointReasonAISpend, PointRefAITask, task.ID,
				"AI 生成消耗 1 积分")
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	work := &model.Work{
		UserID: userID, Type: taskType,
		ImageURL: task.ResultURL, ThumbnailURL: task.ResultURL,
		Meta:   fmt.Sprintf(`{"task_id":%d,"prompt":%q}`, task.ID, prompt),
		Prompt: strings.TrimSpace(prompt), // 发布到广场时作为公开提示词
		Status: 1,
	}
	if err := s.DB.Create(work).Error; err != nil {
		return nil, nil, err
	}
	return task, work, nil
}

// Quota 今日免费额度、已用次数与积分余额，供小程序 AI 页展示「还能生成几张」
func (s *AIService) Quota(userID int64) (limit, used, points int, err error) {
	var cfg model.Config
	if err = s.DB.Where("key = ?", "ai_daily_free_limit").First(&cfg).Error; err != nil {
		return 0, 0, 0, errors.New("系统配置缺失")
	}
	if _, err = fmt.Sscanf(cfg.Value, "%d", &limit); err != nil || limit < 0 {
		return 0, 0, 0, errors.New("系统配置不合法")
	}
	community := &CommunityService{DB: s.DB}
	if used, err = community.TodayQuota(userID); err != nil {
		return 0, 0, 0, err
	}
	if points, err = community.Balance(userID); err != nil {
		return 0, 0, 0, err
	}
	return limit, used, points, nil
}

// safeUploadPath 把 /static/xxx 映射回本地 uploads 路径，拒绝越界
func safeUploadPath(uploadsDir, url string) (string, error) {
	if !strings.HasPrefix(url, "/static/") {
		return "", errors.New("图片地址不合法")
	}
	rel := strings.TrimPrefix(url, "/static/")
	if rel == "" || strings.Contains(rel, "..") || strings.Contains(rel, "/") || strings.Contains(rel, "\\") {
		return "", errors.New("图片地址不合法")
	}
	p := filepath.Join(uploadsDir, filepath.FromSlash(rel))
	if !strings.HasPrefix(p, uploadsDir) {
		return "", errors.New("图片地址不合法")
	}
	if _, err := os.Stat(p); err != nil {
		return "", errors.New("找不到上传的图片，请重新上传")
	}
	return p, nil
}
