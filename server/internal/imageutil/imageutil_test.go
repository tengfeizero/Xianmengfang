package imageutil

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestCompose_PlacementCenterAndScale(t *testing.T) {
	// 底板 100x100 纯红
	board := NewCanvas(100, 100)
	FillCircle(board, 50, 50, 1, color.RGBA{255, 0, 0, 255})
	FillRect(board, 0, 0, 100, 100, color.RGBA{255, 0, 0, 255})
	// 元素 20x20 纯蓝
	el := NewCanvas(20, 20)
	FillRect(el, 0, 0, 20, 20, color.RGBA{0, 0, 255, 255})

	out, err := Compose(board, []Placement{{Img: el, RelX: 0.5, RelY: 0.5, Scale: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if out.Bounds().Dx() != 100 || out.Bounds().Dy() != 100 {
		t.Fatalf("合成尺寸应等于底板: %+v", out.Bounds())
	}
	// 中心点应是元素（蓝），角落应是底板（红）
	if got := rgbaAt(out, 50, 50); got.B != 255 || got.R != 0 {
		t.Fatalf("中心应为元素颜色: %+v", got)
	}
	if got := rgbaAt(out, 2, 2); got.R != 255 || got.B != 0 {
		t.Fatalf("角落应为底板颜色: %+v", got)
	}
}

func TestCompose_EmptyElement(t *testing.T) {
	board := NewCanvas(10, 10)
	if _, err := Compose(board, []Placement{{RelX: 0.5, RelY: 0.5}}); err == nil {
		t.Fatal("空元素应报错")
	}
}

func TestSaveLoadPNG(t *testing.T) {
	img := NewCanvas(8, 8)
	FillCircle(img, 4, 4, 3, color.RGBA{9, 9, 9, 255})
	p := filepath.Join(t.TempDir(), "a", "b.png")
	if err := SavePNG(img, p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("文件未写入")
	}
	back, err := LoadPNG(p)
	if err != nil {
		t.Fatal(err)
	}
	if back.Bounds().Dx() != 8 {
		t.Fatal("读回尺寸不符")
	}
}

func rgbaAt(img *RGBA, x, y int) color.RGBA {
	return img.RGBAAt(x, y)
}
