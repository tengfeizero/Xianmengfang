package service

import (
	
	"testing"

	"xmf/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newAITestDB(t *testing.T, uploadsDir string) (*gorm.DB, *AIService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(model.AllMigrations()...); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Config{Key: "ai_daily_free_limit", Value: "2"}).Error; err != nil {
		t.Fatal(err)
	}
	svc := &AIService{DB: db, Provider: MockProvider{}, UploadsDir: uploadsDir}
	return db, svc
}

func TestAI_QuotaLimit(t *testing.T) {
	uploads := t.TempDir()
	db, svc := newAITestDB(t, uploads)
	// 配额 2：前两次成功
	for i := 0; i < 2; i++ {
		_, _, err := svc.CreateTask(1, "ai_text", "测试作品描述", "")
		if err != nil {
			t.Fatalf("第 %d 次生成不应失败: %v", i+1, err)
		}
	}
	// 第三次应 429
	_, _, err := svc.CreateTask(1, "ai_text", "第三次", "")
	if err != ErrQuotaExceeded {
		t.Fatalf("应报配额用尽，得到: %v", err)
	}
	// 其他用户不受影响
	if _, _, err := svc.CreateTask(2, "ai_text", "别人还有额度", ""); err != nil {
		t.Fatalf("其他用户不应受配额影响: %v", err)
	}
	// 生成物落盘且作品入库
	var works []model.Work
	db.Where("user_id = ?", 1).Find(&works)
	if len(works) != 2 || works[0].Type != "ai_text" {
		t.Fatalf("作品未正确入库: %+v", works)
	}
	var tasks []model.AITask
	db.Where("user_id = ?", 1).Find(&tasks)
	if len(tasks) != 2 || tasks[0].Status != "done" || tasks[0].Provider != "mock" {
		t.Fatalf("任务记录不正确: %+v", tasks)
	}
}

func TestAI_PhotoPathValidation(t *testing.T) {
	uploads := t.TempDir()
	_, svc := newAITestDB(t, uploads)
	// 越界路径拒绝
	_, _, err := svc.CreateTask(1, "ai_photo", "", "/static/../secret.png")
	if err == nil || err.Error() != "图片地址不合法" {
		t.Fatalf("越界路径应被拒绝: %v", err)
	}
	// 静态目录内但文件不存在
	_, _, err = svc.CreateTask(1, "ai_photo", "", "/static/up_notexist.png")
	if err == nil {
		t.Fatalf("不存在的图片应报错")
	}
}

func TestAI_QuotaDisabledWhenZero(t *testing.T) {
	uploads := t.TempDir()
	db, svc := newAITestDB(t, uploads)
	db.Model(&model.Config{}).Where("key = ?", "ai_daily_free_limit").Update("value", "0")
	_, _, err := svc.CreateTask(1, "ai_text", "配额为0", "")
	if err != ErrQuotaExceeded {
		t.Fatalf("配额为 0 应直接拒绝: %v", err)
	}
}
