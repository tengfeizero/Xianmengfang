package repo

import (
	"image/color"
	"os"
	"path/filepath"

	"xmf/internal/imageutil"
)

// 商品演示封面：程序生成的几何占位图，真实商品图由后台「商品管理」上传替换。
var (
	pBgTee  = color.RGBA{247, 243, 238, 255}
	pBgMug  = color.RGBA{234, 242, 248, 255}
	pBgBag  = color.RGBA{245, 239, 227, 255}
	pBgDoll = color.RGBA{251, 238, 241, 255}
	pLine   = color.RGBA{212, 204, 195, 255}
	pWhite  = color.RGBA{255, 255, 255, 255}
	pInk    = color.RGBA{62, 54, 50, 255}
	pPink   = color.RGBA{226, 122, 143, 255}
	pPinkLt = color.RGBA{244, 196, 205, 255}
	pCanvas = color.RGBA{239, 227, 204, 255}
	pFur    = color.RGBA{246, 220, 226, 255}
	pCupIn  = color.RGBA{226, 233, 238, 255}
)

const pSize = 800

// SeedProductImages 生成 4 款演示商品的封面图与特写图，已存在则跳过，可重复执行。
func SeedProductImages(uploadsDir string) error {
	dir := filepath.Join(uploadsDir, "products")
	specs := []struct {
		name string
		draw func(*imageutil.RGBA, float64)
		bg   color.RGBA
	}{
		{"tee", drawTee, pBgTee},
		{"mug", drawMug, pBgMug},
		{"bag", drawBag, pBgBag},
		{"doll", drawDoll, pBgDoll},
	}
	for _, s := range specs {
		views := []struct {
			suffix string
			k      float64
			bg     color.RGBA
		}{
			{"main", 1, s.bg},
			{"detail", 1.7, shade(s.bg, -10)},
		}
		for _, v := range views {
			p := filepath.Join(dir, s.name+"_"+v.suffix+".png")
			if _, err := os.Stat(p); err == nil {
				continue
			}
			c := imageutil.NewCanvas(pSize, pSize)
			imageutil.FillRect(c, 0, 0, pSize, pSize, v.bg)
			s.draw(c, v.k)
			if err := imageutil.SavePNG(c, p); err != nil {
				return err
			}
		}
	}
	return nil
}

// pp 以画布中心 (400,400) 为原点缩放坐标，k=1 为正视图，>1 为特写
func pp(v, k float64) float64 { return 400 + (v-400)*k }

func prr(c *imageutil.RGBA, x, y, w, h, r, k float64, col color.Color) {
	imageutil.FillRoundRect(c, pp(x, k), pp(y, k), w*k, h*k, r*k, col)
}

func pci(c *imageutil.RGBA, cx, cy, r, k float64, col color.Color) {
	imageutil.FillCircle(c, pp(cx, k), pp(cy, k), r*k, col)
}

func parc(c *imageutil.RGBA, cx, cy, r, w, a0, a1, k float64, col color.Color) {
	imageutil.FillArc(c, pp(cx, k), pp(cy, k), r*k, w*k, a0, a1, col)
}

func pring(c *imageutil.RGBA, cx, cy, r, w, k float64, col color.Color) {
	imageutil.FillRing(c, pp(cx, k), pp(cy, k), r*k, w*k, col)
}

// pheart 心形印花（两个圆 + 一个三角）
func pheart(c *imageutil.RGBA, cx, cy, s, k float64, col color.Color) {
	pci(c, cx-s*0.36, cy-s*0.22, s*0.42, k, col)
	pci(c, cx+s*0.36, cy-s*0.22, s*0.42, k, col)
	imageutil.FillTriangle(c,
		[2]float64{pp(cx-s*0.76, k), pp(cy-s*0.1, k)},
		[2]float64{pp(cx+s*0.76, k), pp(cy-s*0.1, k)},
		[2]float64{pp(cx, k), pp(cy+s*0.82, k)}, col)
}

func shade(c color.RGBA, d int) color.RGBA {
	f := func(v uint8) uint8 {
		n := int(v) + d
		if n < 0 {
			n = 0
		}
		if n > 255 {
			n = 255
		}
		return uint8(n)
	}
	return color.RGBA{f(c.R), f(c.G), f(c.B), c.A}
}

// drawTee 正视图：短袖 T 恤剪影 + 胸口印花
func drawTee(c *imageutil.RGBA, k float64) {
	prr(c, 538, 226, 98, 142, 32, k, pLine)
	prr(c, 544, 232, 86, 130, 28, k, pWhite)
	prr(c, 164, 226, 98, 142, 32, k, pLine)
	prr(c, 170, 232, 86, 130, 28, k, pWhite)
	prr(c, 244, 216, 312, 372, 30, k, pLine)
	prr(c, 250, 222, 300, 360, 24, k, pWhite)
	// 领口：背景色覆盖出 U 形，再补上半圈描边
	pci(c, 400, 222, 56, k, pBgTeeDark(k))
	parc(c, 400, 222, 56, 7, 180, 360, k, pLine)
	pheart(c, 400, 372, 96, k, pPinkLt)
	pheart(c, 400, 372, 66, k, pPink)
}

// drawMug 正视图：陶瓷马克杯 + 心形印花
func drawMug(c *imageutil.RGBA, k float64) {
	pring(c, 520, 432, 76, 26, k, pLine)
	pring(c, 520, 432, 76, 14, k, pWhite)
	prr(c, 262, 286, 252, 274, 22, k, pLine)
	prr(c, 268, 292, 240, 262, 18, k, pWhite)
	prr(c, 268, 292, 240, 30, 14, k, pCupIn)
	pheart(c, 388, 412, 84, k, pPink)
}

// drawBag 正视图：帆布袋 + 提手 + 印刷区
func drawBag(c *imageutil.RGBA, k float64) {
	parc(c, 344, 306, 54, 17, 180, 360, k, pLine)
	parc(c, 456, 306, 54, 17, 180, 360, k, pLine)
	prr(c, 244, 300, 312, 336, 16, k, pLine)
	prr(c, 250, 306, 300, 324, 12, k, pCanvas)
	prr(c, 306, 372, 188, 172, 10, k, pPinkLt)
	pheart(c, 400, 430, 70, k, pPink)
}

// drawDoll 正视图：毛绒玩偶（耳朵 + 头 + 身体 + 围裙）
func drawDoll(c *imageutil.RGBA, k float64) {
	pci(c, 292, 214, 58, k, pLine)
	pci(c, 292, 214, 48, k, pFur)
	pci(c, 508, 214, 58, k, pLine)
	pci(c, 508, 214, 48, k, pFur)
	prr(c, 306, 420, 188, 186, 62, k, pLine)
	prr(c, 312, 426, 176, 174, 56, k, pFur)
	pci(c, 400, 318, 142, k, pLine)
	pci(c, 400, 318, 132, k, pFur)
	pci(c, 356, 302, 15, k, pInk)
	pci(c, 360, 297, 4, k, pWhite)
	pci(c, 444, 302, 15, k, pInk)
	pci(c, 448, 297, 4, k, pWhite)
	pci(c, 322, 352, 23, k, pPinkLt)
	pci(c, 478, 352, 23, k, pPinkLt)
	parc(c, 400, 346, 32, 8, 20, 160, k, pInk)
	prr(c, 326, 486, 148, 76, 20, k, pPink)
}

// 领口挖空用：正视图直接给背景色，特写时背景已加深
func pBgTeeDark(k float64) color.Color {
	if k > 1.2 {
		return shade(pBgTee, -10)
	}
	return pBgTee
}
