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

// ErrQuotaExceeded 每日配额用尽
var ErrQuotaExceeded = errors.New("今日 AI 生成次数已用完，明天再来吧")

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
	Provider   Provider
	UploadsDir string
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
	today := time.Now().Format("2006-01-02")
	var used int64
	if err := s.DB.Model(&model.AITask{}).
		Where("user_id = ? AND quota_date = ?", userID, today).Count(&used).Error; err != nil {
		return nil, nil, err
	}
	if int(used) >= limit {
		return nil, nil, ErrQuotaExceeded
	}

	outName := fmt.Sprintf("ai_%s_%04d.png", time.Now().Format("20060102150405"), rand.Intn(10000))
	outPath := filepath.Join(s.UploadsDir, outName)

	task := &model.AITask{
		UserID: userID, Type: taskType, Status: "pending",
		Provider: s.Provider.Name(), QuotaDate: today,
		Params: prompt,
	}
	if taskType == "ai_photo" {
		srcPath, err := safeUploadPath(s.UploadsDir, photoURL)
		if err != nil {
			return nil, nil, err
		}
		task.Params = photoURL
		if err := s.Provider.PhotoStyle(srcPath, outPath); err != nil {
			return nil, nil, err
		}
	} else {
		if err := s.Provider.Text2Image(prompt, outPath); err != nil {
			return nil, nil, err
		}
	}
	task.Status = "done"
	task.ResultURL = "/static/" + outName
	if err := s.DB.Create(task).Error; err != nil {
		return nil, nil, err
	}
	work := &model.Work{
		UserID: userID, Type: taskType,
		ImageURL: task.ResultURL, ThumbnailURL: task.ResultURL,
		Meta:    fmt.Sprintf(`{"task_id":%d,"prompt":%q}`, task.ID, prompt),
		Status:  1,
	}
	if err := s.DB.Create(work).Error; err != nil {
		return nil, nil, err
	}
	return task, work, nil
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
