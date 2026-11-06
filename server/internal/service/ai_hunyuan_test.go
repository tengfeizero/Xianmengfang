package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"xmf/internal/model"
)

// tinyPNG 生成 1x1 PNG 字节。
func tinyPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestHunyuanProviderMessages(t *testing.T) {
	imgBytes := tinyPNG(t)
	// 图片下载源
	imgSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(imgBytes)
	}))
	defer imgSrv.Close()

	var gotPath, gotAuth, gotModel string
	var hasText bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Content []struct {
					Type     string `json:"type"`
					Text     string `json:"text"`
					ImageURL *struct {
						URL string `json:"url"`
					} `json:"image_url"`
				} `json:"content"`
			} `json:"messages"`
		}
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("bad json: %v", err)
		}
		gotModel = body.Model
		for _, m := range body.Messages {
			for _, c := range m.Content {
				if c.Type == "text" {
					hasText = true
				}
				if c.Type == "image_url" && c.ImageURL != nil {
					if !strings.HasPrefix(c.ImageURL.URL, "data:image/png;base64,") {
						t.Errorf("参考图不是 data URL")
					}
				}
			}
		}
		fmt.Fprintf(w, `{"choices":[{"delta":{"image":{"url":"%s/x.png"}}}]}`, imgSrv.URL)
	}))
	defer srv.Close()

	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	os.WriteFile(src, imgBytes, 0o644)

	p := &HunyuanProvider{APIKey: "sk-test", BaseURL: srv.URL, Style: "messages", HTTP: srv.Client()}
	out := filepath.Join(dir, "out.png")
	if err := p.PhotoStyle(src, out); err != nil {
		t.Fatalf("PhotoStyle: %v", err)
	}
	if gotPath != "/v1/wand/hunyuan-image/v35-generation" {
		t.Errorf("endpoint = %s", gotPath)
	}
	if gotAuth != "Bearer sk-test" || gotModel != "hy-image-v3.5-preview" || !hasText {
		t.Errorf("auth=%s model=%s hasText=%v", gotAuth, gotModel, hasText)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, format, err := image.DecodeConfig(f); err != nil || format != "png" {
		t.Errorf("落盘不是 PNG: %v %s", err, format)
	}

	// 文生图：无参考图
	out2 := filepath.Join(dir, "out2.png")
	if err := p.Text2Image("一只猫", out2); err != nil {
		t.Fatalf("Text2Image: %v", err)
	}
}

func TestAI_ResolveProviderFromConfig(t *testing.T) {
	uploads := t.TempDir()
	db, svc := newAITestDB(t, uploads)
	// 未配置密钥 → 回落注入 Provider
	if _, ok := svc.resolveProvider().(MockProvider); !ok {
		t.Fatalf("未配置密钥应回落 Mock")
	}
	// 配置密钥 → 混元，默认端点/模型/格式
	db.Create(&model.Config{Key: "ai_api_key", Value: " sk-abc "})
	p, ok := svc.resolveProvider().(*HunyuanProvider)
	if !ok {
		t.Fatalf("配置密钥后应为 HunyuanProvider")
	}
	if p.APIKey != "sk-abc" || p.BaseURL != "https://tokenhub.tencentmaas.com" ||
		p.Model != "hy-image-v3.5-preview" || p.Style != "messages" {
		t.Fatalf("默认配置不正确: %+v", p)
	}
	// 覆盖 base_url / model / style
	db.Create(&model.Config{Key: "ai_base_url", Value: "https://api.example.com"})
	db.Create(&model.Config{Key: "ai_model", Value: "m2"})
	db.Create(&model.Config{Key: "ai_style", Value: "images"})
	p = svc.resolveProvider().(*HunyuanProvider)
	if p.BaseURL != "https://api.example.com" || p.Model != "m2" || p.Style != "images" {
		t.Fatalf("覆盖配置未生效: %+v", p)
	}
	// 清空密钥 → 回落 Mock
	db.Where("key = ?", "ai_api_key").Delete(&model.Config{})
	if _, ok := svc.resolveProvider().(MockProvider); !ok {
		t.Fatalf("清空密钥后应回落 Mock")
	}
}

func TestHunyuanProviderAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer srv.Close()
	p := &HunyuanProvider{APIKey: "bad", BaseURL: srv.URL, HTTP: srv.Client()}
	err := p.Text2Image("x", filepath.Join(t.TempDir(), "o.png"))
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("期望 401 错误, got %v", err)
	}
}
