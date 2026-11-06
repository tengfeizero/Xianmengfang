package repo

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"xmf/internal/imageutil"
	"xmf/internal/model"

	"gorm.io/gorm"
)

var (
	cSkin   = color.RGBA{247, 216, 196, 255}
	cHair   = color.RGBA{92, 64, 51, 255}
	cHair2  = color.RGBA{58, 44, 38, 255}
	cRed    = color.RGBA{229, 87, 86, 255}
	cBlue   = color.RGBA{100, 149, 237, 255}
	cPink   = color.RGBA{244, 143, 177, 255}
	cDark   = color.RGBA{77, 60, 50, 255}
	cWhite  = color.RGBA{255, 255, 255, 255}
	cCheek  = color.RGBA{250, 170, 160, 255}
	cYellow = color.RGBA{255, 209, 102, 255}
)

// SeedAssets 程序化生成演示素材（底板 + 元素 PNG），已存在则跳过，可重复执行。
// 返回生成的素材目录。演示素材为几何图形占位，真实素材由后台上传替换。
func SeedAssets(uploadsDir string) error {
	dir := filepath.Join(uploadsDir, "elements")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	gens := map[string]func() *imageutil.RGBA{
		"board_girl.png":  genBoardGirl,
		"hair_twin.png":   genHairTwin,
		"hair_bob.png":    genHairBob,
		"hat_pointy.png":  genHatPointy,
		"hat_beret.png":   genHatBeret,
		"lash_cute.png":   genLashCute,
		"acc_bow.png":     genBow,
	}
	for name, gen := range gens {
		p := filepath.Join(uploadsDir, name)
		if name != "board_girl.png" { // 元素素材在 elements/ 子目录，与 image_url 一致
			p = filepath.Join(dir, name)
		}
		if _, err := os.Stat(p); err == nil {
			continue
		}
		if err := imageutil.SavePNG(gen(), p); err != nil {
			return fmt.Errorf("生成素材 %s 失败: %w", name, err)
		}
	}
	return nil
}

// SeedCatalogAssets 写入底板与元素的数据库记录（空表才写，可重复执行）
func SeedCatalogAssets(db *gorm.DB) error {
	var n int64
	if err := db.Model(&model.Element{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	board := model.Board{
		Name:   "底板·小姑娘",
		URL:    "/static/board_girl.png",
		Width:  400, Height: 500,
		Status: 1,
	}
	if err := db.Create(&board).Error; err != nil {
		return err
	}
	els := []model.Element{
		{Category: "hair", Name: "双马尾", ImageURL: "/static/elements/hair_twin.png", Anchor: `{"x":0.5,"y":0.2,"scale":1}`, Sort: 1, Status: 1},
		{Category: "hair", Name: "波波头", ImageURL: "/static/elements/hair_bob.png", Anchor: `{"x":0.5,"y":0.2,"scale":1}`, Sort: 2, Status: 1},
		{Category: "hat", Name: "红色尖尖帽", ImageURL: "/static/elements/hat_pointy.png", Anchor: `{"x":0.5,"y":0.14,"scale":1}`, Sort: 1, Status: 1},
		{Category: "hat", Name: "蓝色贝雷帽", ImageURL: "/static/elements/hat_beret.png", Anchor: `{"x":0.5,"y":0.15,"scale":1}`, Sort: 2, Status: 1},
		{Category: "lash", Name: "可爱睫毛", ImageURL: "/static/elements/lash_cute.png", Anchor: `{"x":0.5,"y":0.37,"scale":1}`, Sort: 1, Status: 1},
		{Category: "accessory", Name: "蝴蝶结", ImageURL: "/static/elements/acc_bow.png", Anchor: `{"x":0.33,"y":0.1,"scale":0.6}`, Sort: 1, Status: 1},
	}
	return db.Create(&els).Error
}

// ---- 底板：400x500 小姑娘（头 + 裙身 + 眼睛 + 腮红 + 微笑） ----

func genBoardGirl() *imageutil.RGBA {
	c := imageutil.NewCanvas(400, 500)
	// 裙身（三角）
	imageutil.FillTriangle(c, [2]float64{200, 260}, [2]float64{110, 470}, [2]float64{290, 470}, cPink)
	// 头
	imageutil.FillCircle(c, 200, 180, 105, cSkin)
	// 眼睛
	imageutil.FillCircle(c, 162, 190, 10, cDark)
	imageutil.FillCircle(c, 238, 190, 10, cDark)
	imageutil.FillCircle(c, 165, 186, 3, cWhite)
	imageutil.FillCircle(c, 241, 186, 3, cWhite)
	// 腮红
	imageutil.FillCircle(c, 138, 220, 14, cCheek)
	imageutil.FillCircle(c, 262, 220, 14, cCheek)
	// 微笑（下半圆弧）
	imageutil.FillArc(c, 200, 216, 22, 6, 20, 160, cDark)
	return c
}

// ---- 元素素材（200x200 或指定尺寸，透明背景） ----

func genHairTwin() *imageutil.RGBA { // 双马尾
	c := imageutil.NewCanvas(200, 160)
	imageutil.FillCircle(c, 100, 70, 68, cHair)
	imageutil.FillCircle(c, 100, 95, 52, color.Transparent) // 露出脸
	imageutil.FillCircle(c, 26, 105, 26, cHair)             // 左马尾
	imageutil.FillCircle(c, 174, 105, 26, cHair)            // 右马尾
	imageutil.FillCircle(c, 26, 130, 18, cHair)
	imageutil.FillCircle(c, 174, 130, 18, cHair)
	return c
}

func genHairBob() *imageutil.RGBA { // 波波头
	c := imageutil.NewCanvas(200, 150)
	imageutil.FillCircle(c, 100, 65, 70, cHair2)
	imageutil.FillCircle(c, 100, 90, 55, color.Transparent) // 露出脸
	imageutil.FillRect(c, 30, 60, 24, 80, cHair2)
	imageutil.FillRect(c, 146, 60, 24, 80, cHair2)
	return c
}

func genHatPointy() *imageutil.RGBA { // 红色尖尖帽
	c := imageutil.NewCanvas(200, 200)
	imageutil.FillTriangle(c, [2]float64{100, 10}, [2]float64{38, 150}, [2]float64{162, 150}, cRed)
	imageutil.FillCircle(c, 100, 158, 62, cRed)
	imageutil.FillCircle(c, 100, 165, 55, color.Transparent)
	imageutil.FillCircle(c, 100, 10, 10, cWhite) // 帽顶球
	return c
}

func genHatBeret() *imageutil.RGBA { // 蓝色贝雷帽
	c := imageutil.NewCanvas(200, 130)
	imageutil.FillCircle(c, 100, 68, 62, cBlue)
	imageutil.FillCircle(c, 100, 95, 52, color.Transparent)
	imageutil.FillCircle(c, 100, 14, 9, cYellow) // 帽扣
	return c
}

func genLashCute() *imageutil.RGBA { // 一对睫毛
	c := imageutil.NewCanvas(220, 44)
	for i := 0; i < 4; i++ {
		x := 24 + i*12
		imageutil.FillRect(c, x, 16, 4, 14, cDark)
		imageutil.FillCircle(c, float64(x)+2, 30, 3, cDark)
		x2 := 130 + i*12
		imageutil.FillRect(c, x2, 16, 4, 14, cDark)
		imageutil.FillCircle(c, float64(x2)+2, 30, 3, cDark)
	}
	return c
}

func genBow() *imageutil.RGBA { // 蝴蝶结
	c := imageutil.NewCanvas(120, 80)
	imageutil.FillTriangle(c, [2]float64{58, 40}, [2]float64{10, 12}, [2]float64{10, 68}, cPink)
	imageutil.FillTriangle(c, [2]float64{62, 40}, [2]float64{110, 12}, [2]float64{110, 68}, cPink)
	imageutil.FillCircle(c, 60, 40, 12, cRed)
	return c
}
