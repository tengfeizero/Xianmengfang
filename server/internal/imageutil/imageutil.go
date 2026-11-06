// Package imageutil 提供图像合成与演示素材生成（仅用标准库 + x/image/draw）
package imageutil

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
)

// RGBA 是标准库 image.RGBA 的别名，便于包外使用
type RGBA = image.RGBA

// LoadPNG 读 PNG
func LoadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

// SavePNG 写 PNG（父目录不存在则创建）
func SavePNG(img image.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// NewCanvas 建一块带透明背景的画布
func NewCanvas(w, h int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

// FillCircle 实心圆
func FillCircle(dst *image.RGBA, cx, cy, r float64, c color.Color) {
	for y := int(cy - r); y <= int(cy+r); y++ {
		for x := int(cx - r); x <= int(cx+r); x++ {
			if x < 0 || y < 0 || x >= dst.Bounds().Dx() || y >= dst.Bounds().Dy() {
				continue
			}
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= r*r {
				dst.Set(x, y, c)
			}
		}
	}
}

// FillRing 圆环（描边圆）
func FillRing(dst *image.RGBA, cx, cy, r, width float64, c color.Color) {
	FillCircle(dst, cx, cy, r, c)
	inner := r - width
	for y := int(cy - r); y <= int(cy+r); y++ {
		for x := int(cx - r); x <= int(cx+r); x++ {
			if x < 0 || y < 0 || x >= dst.Bounds().Dx() || y >= dst.Bounds().Dy() {
				continue
			}
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= inner*inner {
				dst.Set(x, y, color.Transparent)
			}
		}
	}
}

// FillArc 圆弧描边（角度制，0°=正右方，顺时针增加；只画 [start,end] 区间）
func FillArc(dst *image.RGBA, cx, cy, r, width float64, startDeg, endDeg float64, c color.Color) {
	step := 1.0
	if endDeg < startDeg {
		step = -step
	}
	for deg := startDeg; (step > 0 && deg <= endDeg) || (step < 0 && deg >= endDeg); deg += step {
		rad := deg * math.Pi / 180
		FillCircle(dst, cx+r*math.Cos(rad), cy+r*math.Sin(rad), width/2, c)
	}
}

// FillTriangle 实心三角形
func FillTriangle(dst *image.RGBA, p1, p2, p3 [2]float64, c color.Color) {
	minX := math.Min(p1[0], math.Min(p2[0], p3[0]))
	maxX := math.Max(p1[0], math.Max(p2[0], p3[0]))
	minY := math.Min(p1[1], math.Min(p2[1], p3[1]))
	maxY := math.Max(p1[1], math.Max(p2[1], p3[1]))
	area := triArea(p1, p2, p3)
	if area == 0 {
		return
	}
	for y := int(minY); y <= int(maxY); y++ {
		for x := int(minX); x <= int(maxX); x++ {
			if x < 0 || y < 0 || x >= dst.Bounds().Dx() || y >= dst.Bounds().Dy() {
				continue
			}
			p := [2]float64{float64(x) + 0.5, float64(y) + 0.5}
			a1 := triArea(p, p2, p3)
			a2 := triArea(p1, p, p3)
			a3 := triArea(p1, p2, p)
			if math.Abs(a1+a2+a3-area) < 1e-6 {
				dst.Set(x, y, c)
			}
		}
	}
}

func triArea(a, b, c [2]float64) float64 {
	return math.Abs((a[0]*(b[1]-c[1]) + b[0]*(c[1]-a[1]) + c[0]*(a[1]-b[1])) / 2)
}

// FillRoundRect 实心圆角矩形
func FillRoundRect(dst *image.RGBA, x0, y0, w, h, r float64, c color.Color) {
	FillCircle(dst, x0+r, y0+r, r, c)
	FillCircle(dst, x0+w-r, y0+r, r, c)
	FillCircle(dst, x0+r, y0+h-r, r, c)
	FillCircle(dst, x0+w-r, y0+h-r, r, c)
	FillRect(dst, int(x0+r), int(y0), int(w-2*r), int(h), c)
	FillRect(dst, int(x0), int(y0+r), int(w), int(h-2*r), c)
}

func FillRect(dst *image.RGBA, x, y, w, h int, c color.Color) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			px, py := x+dx, y+dy
			if px < 0 || py < 0 || px >= dst.Bounds().Dx() || py >= dst.Bounds().Dy() {
				continue
			}
			dst.Set(px, py, c)
		}
	}
}

// Scale 等比缩放（CatmullRom 高质量）
func Scale(src image.Image, factor float64) image.Image {
	if factor == 1 {
		return src
	}
	w := max(1, int(float64(src.Bounds().Dx())*factor))
	h := max(1, int(float64(src.Bounds().Dy())*factor))
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

// PasteOver 把 src 以 (左上角) 贴到 dst，保留透明度
func PasteOver(dst *image.RGBA, src image.Image, x, y int) {
	r := image.Rect(x, y, x+src.Bounds().Dx(), y+src.Bounds().Dy())
	draw.Draw(dst, r, src, src.Bounds().Min, draw.Over)
}

// Compose 合成：底板 + 放置列表。placement 坐标为相对底板的比例 (0~1)，表示元素中心位置；
// scale 为相对元素原始尺寸的倍数。
type Placement struct {
	Img       image.Image
	RelX      float64
	RelY      float64
	Scale     float64
	AlphaOver bool
	Z         int // 图层顺序，小者先画（在底层）
}

func Compose(board image.Image, placements []Placement) (*image.RGBA, error) {
	b := board.Bounds()
	canvas := NewCanvas(b.Dx(), b.Dy())
	PasteOver(canvas, board, 0, 0)
	for i, p := range placements {
		if p.Img == nil {
			return nil, fmt.Errorf("第 %d 个元素图像为空", i+1)
		}
		img := Scale(p.Img, p.Scale)
		cx := int(p.RelX*float64(b.Dx())) - img.Bounds().Dx()/2
		cy := int(p.RelY*float64(b.Dy())) - img.Bounds().Dy()/2
		PasteOver(canvas, img, cx, cy)
	}
	return canvas, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
