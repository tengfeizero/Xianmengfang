package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// HunyuanProvider 腾讯混元生图（hy-image-v3.5-preview）。
// 同一模型支持文生图与参考图指令编辑，对应 Provider 的两个方法：
//   - Text2Image  → 纯文本 prompt
//   - PhotoStyle  → 参考图（本地文件转 data URL）+ 风格化指令
//
// 结果以临时 URL 返回（约 12 小时过期），这里立即下载并统一转码 PNG 落盘。
// 支持两种请求风格：
//   - "messages"（默认）：腾讯云 TokenHub /v1/wand/hunyuan-image/v35-generation
//   - "images"：OpenAI 风格 /v1/images/generations（operation=text-to-image / image-edit）
type HunyuanProvider struct {
	APIKey  string
	BaseURL string // 缺省 https://tokenhub-intl.tencentcloudmaas.com
	Model   string // 缺省 hy-image-v3.5-preview
	Style   string // messages | images
	HTTP    *http.Client
}

const photoStylePrompt = "把这张照片转换成可爱扁平风格的卡通插画，Q版人物，柔和配色带轻微阴影，保持人物主体、姿势和构图完全不变，纯色简洁背景"

func (p *HunyuanProvider) Name() string { return "hy-image" }

func (p *HunyuanProvider) Text2Image(prompt, outPath string) error {
	return p.generate("", prompt, outPath)
}

func (p *HunyuanProvider) PhotoStyle(srcPath, outPath string) error {
	return p.generate(srcPath, photoStylePrompt, outPath)
}

func (p *HunyuanProvider) client() *http.Client {
	if p.HTTP != nil {
		return p.HTTP
	}
	return &http.Client{Timeout: 120 * time.Second}
}

func (p *HunyuanProvider) baseURL() string {
	if p.BaseURL != "" {
		return strings.TrimRight(p.BaseURL, "/")
	}
	return "https://tokenhub-intl.tencentcloudmaas.com"
}

func (p *HunyuanProvider) model() string {
	if p.Model != "" {
		return p.Model
	}
	return "hy-image-v3.5-preview"
}

func (p *HunyuanProvider) style() string {
	if p.Style == "images" {
		return "images"
	}
	return "messages"
}

// generate 调 API 并把结果图落盘。srcPath 非空时作为参考图（图生图）。
func (p *HunyuanProvider) generate(srcPath, prompt, outPath string) error {
	var body map[string]any
	var imageURLPath string
	if p.style() == "images" {
		operation := "text-to-image"
		payload := map[string]any{
			"n": 1, "operation": operation, "response_format": "url",
			"size": "1024x1024", "model": p.model(), "prompt": prompt,
		}
		if srcPath != "" {
			payload["operation"] = "image-edit"
			dataURL, err := fileToDataURL(srcPath)
			if err != nil {
				return err
			}
			payload["image"] = []string{dataURL}
		}
		body = payload
		imageURLPath = "data.0.url"
	} else {
		content := []any{map[string]any{"type": "text", "text": prompt}}
		if srcPath != "" {
			dataURL, err := fileToDataURL(srcPath)
			if err != nil {
				return err
			}
			content = []any{
				map[string]any{"type": "image_url", "image_url": map[string]string{"url": dataURL}},
				map[string]any{"type": "text", "text": prompt},
			}
		}
		body = map[string]any{
			"model": p.model(), "size": "1024x1024",
			"messages": []any{map[string]any{"role": "user", "content": content}},
		}
		imageURLPath = "choices.0.delta.image.url"
	}

	imageURL, err := p.requestImageURL(body, imageURLPath)
	if err != nil {
		return err
	}
	return downloadAsPNG(imageURL, outPath, p.client())
}

// requestImageURL 发请求并按 dotted path（如 choices.0.delta.image.url）取图片 URL。
func (p *HunyuanProvider) requestImageURL(body map[string]any, urlPath string) (string, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	endpoint := p.baseURL() + "/v1/wand/hunyuan-image/v35-generation"
	if p.style() == "images" {
		endpoint = p.baseURL() + "/v1/images/generations"
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	resp, err := p.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("AI 服务请求失败: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI 服务返回 %d: %s", resp.StatusCode, snippet(raw))
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", fmt.Errorf("AI 服务响应解析失败: %w", err)
	}
	v, ok := dig(decoded, strings.Split(urlPath, "."))
	if !ok {
		return "", fmt.Errorf("AI 服务响应缺少图片地址: %s", snippet(raw))
	}
	u, _ := v.(string)
	if u == "" {
		return "", fmt.Errorf("AI 服务返回空图片地址")
	}
	return u, nil
}

// downloadAsPNG 下载图片并统一转码为 PNG 落盘（服务端产物统一 .png 扩展名）。
func downloadAsPNG(url, outPath string, client *http.Client) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("下载生成结果失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载生成结果返回 %d", resp.StatusCode)
	}
	img, _, err := image.Decode(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return fmt.Errorf("生成结果不是有效图片: %w", err)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// fileToDataURL 本地图片读为 data URL（参考图上传格式）。
func fileToDataURL(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New("找不到上传的图片，请重新上传")
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw), nil
}

// dig 按路径逐层取 map/slice 字段。
func dig(v any, path []string) (any, bool) {
	for _, k := range path {
		switch node := v.(type) {
		case map[string]any:
			next, ok := node[k]
			if !ok {
				return nil, false
			}
			v = next
		case []any:
			var idx int
			if _, err := fmt.Sscanf(k, "%d", &idx); err != nil || idx >= len(node) {
				return nil, false
			}
			v = node[idx]
		default:
			return nil, false
		}
	}
	return v, true
}

func snippet(b []byte) string {
	s := string(b)
	if len(s) > 300 {
		s = s[:300] + "..."
	}
	return s
}
