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
	cRedD   = color.RGBA{198, 60, 60, 255}
	cBlue   = color.RGBA{100, 149, 237, 255}
	cPink   = color.RGBA{244, 143, 177, 255}
	cDark   = color.RGBA{77, 60, 50, 255}
	cWhite  = color.RGBA{255, 255, 255, 255}
	cCheek  = color.RGBA{250, 170, 160, 255}
	cYellow = color.RGBA{255, 209, 102, 255}
	cCream  = color.RGBA{252, 243, 232, 255}
	cGray   = color.RGBA{240, 238, 236, 255}
	cStraw  = color.RGBA{226, 190, 132, 255}
	cStrawD = color.RGBA{204, 164, 106, 255}
	cWood   = color.RGBA{214, 178, 128, 255}
	cString = color.RGBA{186, 180, 174, 255}
)

// SeedAssets 程序化生成演示素材（底板 + 元素 PNG），已存在则跳过，可重复执行。
// 返回生成的素材目录。演示素材为几何图形占位，真实素材由后台上传替换。
func SeedAssets(uploadsDir string) error {
	dir := filepath.Join(uploadsDir, "elements")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	gens := map[string]func() *imageutil.RGBA{
		// 底板（uploads 根目录）
		"board_girl.png":  genBoardGirl,
		"board_boy.png":   genBoardBoy,
		"board_bunny.png": genBoardBunny,
		"board_panda.png": genBoardPanda,
		// 头发
		"hair_twin.png":  genHairTwin,
		"hair_bob.png":   genHairBob,
		"hair_long.png":  genHairLong,
		"hair_bun.png":   genHairBun,
		"hair_curly.png": genHairCurly,
		"hair_buns.png":  genHairBuns,
		// 帽子
		"hat_pointy.png": genHatPointy,
		"hat_beret.png":  genHatBeret,
		"hat_straw.png":  genHatStraw,
		"hat_crown.png":  genHatCrown,
		"hat_santa.png":  genHatSanta,
		// 睫毛
		"lash_cute.png": genLashCute,
		"lash_star.png": genLashStar,
		// 表情（带肤色底垫，可覆盖底板原表情）
		"exp_smile.png": genExpSmile,
		"exp_heart.png": genExpHeart,
		"exp_wink.png":  genExpWink,
		"exp_pout.png":  genExpPout,
		// 服装
		"outfit_apron.png":   genOutfitApron,
		"outfit_cape.png":    genOutfitCape,
		"outfit_scarf.png":   genOutfitScarf,
		"outfit_bowtie.png":  genOutfitBowtie,
		// 配饰
		"acc_bow.png":      genBow,
		"acc_glasses.png":  genAccGlasses,
		"acc_balloon.png":  genAccBalloon,
		"acc_lollipop.png": genAccLollipop,
		"acc_starclip.png": genAccStarclip,
	}
	for name, gen := range gens {
		p := filepath.Join(uploadsDir, name)
		if name != "board_girl.png" && name != "board_boy.png" &&
			name != "board_bunny.png" && name != "board_panda.png" { // 元素素材在 elements/ 子目录，与 image_url 一致
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

// 底板统一脸部几何（元素锚点跨底板通用）：
// 头心 (200,180) r105，眼 (162,190)/(238,190)，腮红 (138,220)/(262,220)，嘴弧 (200,216)。
func drawFace(c *imageutil.RGBA) {
	imageutil.FillCircle(c, 162, 190, 10, cDark)
	imageutil.FillCircle(c, 238, 190, 10, cDark)
	imageutil.FillCircle(c, 165, 186, 3, cWhite)
	imageutil.FillCircle(c, 241, 186, 3, cWhite)
	imageutil.FillArc(c, 200, 216, 22, 6, 20, 160, cDark)
}

func drawCheeks(c *imageutil.RGBA) {
	imageutil.FillCircle(c, 138, 220, 14, cCheek)
	imageutil.FillCircle(c, 262, 220, 14, cCheek)
}

// SeedCatalogAssets 写入底板与元素的数据库记录。
// 按「名字匹配、只补缺失」幂等回填：老库也能自动获得新素材，不覆盖后台已建/已改内容。
func SeedCatalogAssets(db *gorm.DB) error {
	boards := []model.Board{
		{Name: "底板·小姑娘", URL: "/static/board_girl.png", Width: 400, Height: 500, Status: 1},
		{Name: "底板·小男孩", URL: "/static/board_boy.png", Width: 400, Height: 500, Status: 1},
		{Name: "底板·小兔子", URL: "/static/board_bunny.png", Width: 400, Height: 500, Status: 1},
		{Name: "底板·小熊猫", URL: "/static/board_panda.png", Width: 400, Height: 500, Status: 1},
	}
	for i := range boards {
		var n int64
		if err := db.Model(&model.Board{}).Where("name = ?", boards[i].Name).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if err := db.Create(&boards[i]).Error; err != nil {
				return err
			}
		}
	}

	els := []model.Element{
		// hair 头发
		{Category: "hair", Name: "双马尾", ImageURL: "/static/elements/hair_twin.png", Anchor: `{"x":0.5,"y":0.2,"scale":1}`, Sort: 1, Status: 1},
		{Category: "hair", Name: "波波头", ImageURL: "/static/elements/hair_bob.png", Anchor: `{"x":0.5,"y":0.2,"scale":1}`, Sort: 2, Status: 1},
		{Category: "hair", Name: "长直发", ImageURL: "/static/elements/hair_long.png", Anchor: `{"x":0.5,"y":0.2,"scale":1}`, Sort: 3, Status: 1},
		{Category: "hair", Name: "丸子头", ImageURL: "/static/elements/hair_bun.png", Anchor: `{"x":0.5,"y":0.19,"scale":1}`, Sort: 4, Status: 1},
		{Category: "hair", Name: "卷发", ImageURL: "/static/elements/hair_curly.png", Anchor: `{"x":0.5,"y":0.2,"scale":1}`, Sort: 5, Status: 1},
		{Category: "hair", Name: "双丸子", ImageURL: "/static/elements/hair_buns.png", Anchor: `{"x":0.5,"y":0.18,"scale":1}`, Sort: 6, Status: 1},
		// hat 帽子
		{Category: "hat", Name: "红色尖尖帽", ImageURL: "/static/elements/hat_pointy.png", Anchor: `{"x":0.5,"y":0.14,"scale":1}`, Sort: 1, Status: 1},
		{Category: "hat", Name: "蓝色贝雷帽", ImageURL: "/static/elements/hat_beret.png", Anchor: `{"x":0.5,"y":0.15,"scale":1}`, Sort: 2, Status: 1},
		{Category: "hat", Name: "黄色草帽", ImageURL: "/static/elements/hat_straw.png", Anchor: `{"x":0.5,"y":0.15,"scale":1}`, Sort: 3, Status: 1},
		{Category: "hat", Name: "金色皇冠", ImageURL: "/static/elements/hat_crown.png", Anchor: `{"x":0.5,"y":0.12,"scale":1}`, Sort: 4, Status: 1},
		{Category: "hat", Name: "圣诞帽", ImageURL: "/static/elements/hat_santa.png", Anchor: `{"x":0.5,"y":0.14,"scale":1}`, Sort: 5, Status: 1},
		// lash 睫毛
		{Category: "lash", Name: "可爱睫毛", ImageURL: "/static/elements/lash_cute.png", Anchor: `{"x":0.5,"y":0.37,"scale":1}`, Sort: 1, Status: 1},
		{Category: "lash", Name: "星星睫毛", ImageURL: "/static/elements/lash_star.png", Anchor: `{"x":0.5,"y":0.37,"scale":1}`, Sort: 2, Status: 1},
		// expression 表情（带肤色底垫，可覆盖底板原表情）
		{Category: "expression", Name: "笑眼弯弯", ImageURL: "/static/elements/exp_smile.png", Anchor: `{"x":0.5,"y":0.38,"scale":1}`, Sort: 1, Status: 1},
		{Category: "expression", Name: "爱心眼", ImageURL: "/static/elements/exp_heart.png", Anchor: `{"x":0.5,"y":0.38,"scale":1}`, Sort: 2, Status: 1},
		{Category: "expression", Name: "眨眼", ImageURL: "/static/elements/exp_wink.png", Anchor: `{"x":0.5,"y":0.38,"scale":1}`, Sort: 3, Status: 1},
		{Category: "expression", Name: "嘟嘟嘴", ImageURL: "/static/elements/exp_pout.png", Anchor: `{"x":0.5,"y":0.436,"scale":1}`, Sort: 4, Status: 1},
		// outfit 服装
		{Category: "outfit", Name: "小围裙", ImageURL: "/static/elements/outfit_apron.png", Anchor: `{"x":0.5,"y":0.75,"scale":1}`, Sort: 1, Status: 1},
		{Category: "outfit", Name: "红色小披风", ImageURL: "/static/elements/outfit_cape.png", Anchor: `{"x":0.5,"y":0.73,"scale":1}`, Sort: 2, Status: 1},
		{Category: "outfit", Name: "蓝色围巾", ImageURL: "/static/elements/outfit_scarf.png", Anchor: `{"x":0.5,"y":0.63,"scale":1}`, Sort: 3, Status: 1},
		{Category: "outfit", Name: "红色领结", ImageURL: "/static/elements/outfit_bowtie.png", Anchor: `{"x":0.5,"y":0.58,"scale":0.8}`, Sort: 4, Status: 1},
		// accessory 配饰
		{Category: "accessory", Name: "蝴蝶结", ImageURL: "/static/elements/acc_bow.png", Anchor: `{"x":0.33,"y":0.1,"scale":0.6}`, Sort: 1, Status: 1},
		{Category: "accessory", Name: "圆框眼镜", ImageURL: "/static/elements/acc_glasses.png", Anchor: `{"x":0.5,"y":0.38,"scale":0.95}`, Sort: 2, Status: 1},
		{Category: "accessory", Name: "粉色气球", ImageURL: "/static/elements/acc_balloon.png", Anchor: `{"x":0.85,"y":0.3,"scale":0.9}`, Sort: 3, Status: 1},
		{Category: "accessory", Name: "棒棒糖", ImageURL: "/static/elements/acc_lollipop.png", Anchor: `{"x":0.85,"y":0.6,"scale":0.9}`, Sort: 4, Status: 1},
		{Category: "accessory", Name: "星星发夹", ImageURL: "/static/elements/acc_starclip.png", Anchor: `{"x":0.27,"y":0.13,"scale":0.6}`, Sort: 5, Status: 1},
	}
	for i := range els {
		var n int64
		if err := db.Model(&model.Element{}).Where("name = ?", els[i].Name).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if err := db.Create(&els[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// ---- 底板：400x500，脸部几何统一，见 drawFace ----

func genBoardGirl() *imageutil.RGBA {
	c := imageutil.NewCanvas(400, 500)
	// 裙身（三角）
	imageutil.FillTriangle(c, [2]float64{200, 260}, [2]float64{110, 470}, [2]float64{290, 470}, cPink)
	// 头
	imageutil.FillCircle(c, 200, 180, 105, cSkin)
	drawFace(c)
	drawCheeks(c)
	return c
}

func genBoardBoy() *imageutil.RGBA {
	c := imageutil.NewCanvas(400, 500)
	// 上衣 + 手臂
	imageutil.FillRoundRect(c, 140, 285, 120, 150, 26, cBlue)
	imageutil.FillRoundRect(c, 104, 295, 40, 105, 18, cBlue)
	imageutil.FillRoundRect(c, 256, 295, 40, 105, 18, cBlue)
	imageutil.FillCircle(c, 124, 402, 14, cSkin)
	imageutil.FillCircle(c, 276, 402, 14, cSkin)
	// 短裤
	imageutil.FillRoundRect(c, 132, 435, 136, 50, 18, cDark)
	// 头：先深色整体，再肤色圆盖出"上深下肤"的短发刘海
	imageutil.FillCircle(c, 200, 160, 100, cHair2)
	imageutil.FillCircle(c, 200, 190, 97, cSkin)
	// 鬓角
	imageutil.FillRect(c, 105, 165, 16, 42, cHair2)
	imageutil.FillRect(c, 279, 165, 16, 42, cHair2)
	drawFace(c)
	drawCheeks(c)
	return c
}

func genBoardBunny() *imageutil.RGBA {
	c := imageutil.NewCanvas(400, 500)
	// 长耳朵（先画，头盖住耳根）
	imageutil.FillRoundRect(c, 148, 18, 44, 140, 22, cCream)
	imageutil.FillRoundRect(c, 208, 18, 44, 140, 22, cCream)
	imageutil.FillRoundRect(c, 158, 30, 24, 112, 12, cPink)
	imageutil.FillRoundRect(c, 218, 30, 24, 112, 12, cPink)
	// 头 + 白裙
	imageutil.FillCircle(c, 200, 180, 105, cCream)
	imageutil.FillTriangle(c, [2]float64{200, 265}, [2]float64{112, 470}, [2]float64{288, 470}, cWhite)
	// 裙身扣子
	imageutil.FillCircle(c, 200, 330, 7, cPink)
	imageutil.FillCircle(c, 200, 370, 7, cPink)
	drawFace(c)
	drawCheeks(c)
	// 兔鼻
	imageutil.FillTriangle(c, [2]float64{190, 208}, [2]float64{210, 208}, [2]float64{200, 220}, cPink)
	return c
}

func genBoardPanda() *imageutil.RGBA {
	c := imageutil.NewCanvas(400, 500)
	// 圆耳朵
	imageutil.FillCircle(c, 118, 100, 42, cDark)
	imageutil.FillCircle(c, 282, 100, 42, cDark)
	imageutil.FillCircle(c, 118, 100, 20, cGray)
	imageutil.FillCircle(c, 282, 100, 20, cGray)
	// 头 + 身体
	imageutil.FillCircle(c, 200, 180, 105, cWhite)
	imageutil.FillTriangle(c, [2]float64{200, 265}, [2]float64{115, 470}, [2]float64{285, 470}, cGray)
	imageutil.FillCircle(c, 200, 400, 58, cWhite)
	// 黑眼圈 + 眼
	imageutil.FillCircle(c, 160, 192, 26, cDark)
	imageutil.FillCircle(c, 240, 192, 26, cDark)
	imageutil.FillCircle(c, 162, 190, 9, cWhite)
	imageutil.FillCircle(c, 238, 190, 9, cWhite)
	imageutil.FillCircle(c, 162, 191, 4.5, cDark)
	imageutil.FillCircle(c, 238, 191, 4.5, cDark)
	drawCheeks(c)
	// 鼻 + 嘴
	imageutil.FillCircle(c, 200, 222, 8, cDark)
	imageutil.FillArc(c, 200, 236, 14, 5, 20, 160, cDark)
	return c
}

// ---- 元素素材（透明背景；尺寸相对底板 400x500 设计，合成按原始像素贴放） ----

// punch 露脸挖孔
func punch(c *imageutil.RGBA, cx, cy, r float64) {
	imageutil.FillCircle(c, cx, cy, r, color.Transparent)
}

func genHairTwin() *imageutil.RGBA { // 双马尾
	c := imageutil.NewCanvas(200, 160)
	imageutil.FillCircle(c, 100, 70, 68, cHair)
	punch(c, 100, 95, 52)
	imageutil.FillCircle(c, 26, 105, 26, cHair)
	imageutil.FillCircle(c, 174, 105, 26, cHair)
	imageutil.FillCircle(c, 26, 130, 18, cHair)
	imageutil.FillCircle(c, 174, 130, 18, cHair)
	return c
}

func genHairBob() *imageutil.RGBA { // 波波头
	c := imageutil.NewCanvas(200, 150)
	imageutil.FillCircle(c, 100, 65, 70, cHair2)
	punch(c, 100, 90, 55)
	imageutil.FillRect(c, 30, 60, 24, 80, cHair2)
	imageutil.FillRect(c, 146, 60, 24, 80, cHair2)
	return c
}

func genHairLong() *imageutil.RGBA { // 长直发
	c := imageutil.NewCanvas(200, 210)
	imageutil.FillRoundRect(c, 12, 70, 34, 135, 16, cHair)
	imageutil.FillRoundRect(c, 154, 70, 34, 135, 16, cHair)
	imageutil.FillCircle(c, 100, 75, 70, cHair)
	punch(c, 100, 100, 52)
	return c
}

func genHairBun() *imageutil.RGBA { // 丸子头
	c := imageutil.NewCanvas(200, 170)
	imageutil.FillCircle(c, 100, 22, 30, cHair)
	imageutil.FillCircle(c, 100, 80, 68, cHair)
	punch(c, 100, 105, 52)
	return c
}

func genHairCurly() *imageutil.RGBA { // 卷发（一圈云朵状圆）
	c := imageutil.NewCanvas(200, 160)
	for _, p := range [][2]float64{{40, 85}, {66, 45}, {100, 32}, {134, 45}, {160, 85}} {
		imageutil.FillCircle(c, p[0], p[1], 30, cHair)
	}
	imageutil.FillCircle(c, 100, 60, 55, cHair)
	punch(c, 100, 95, 52)
	return c
}

func genHairBuns() *imageutil.RGBA { // 双丸子
	c := imageutil.NewCanvas(200, 160)
	imageutil.FillCircle(c, 100, 70, 68, cHair)
	punch(c, 100, 95, 52)
	imageutil.FillCircle(c, 18, 28, 26, cHair)
	imageutil.FillCircle(c, 182, 28, 26, cHair)
	imageutil.FillCircle(c, 36, 50, 8, cPink)
	imageutil.FillCircle(c, 164, 50, 8, cPink)
	return c
}

func genHatPointy() *imageutil.RGBA { // 红色尖尖帽
	c := imageutil.NewCanvas(200, 200)
	imageutil.FillTriangle(c, [2]float64{100, 10}, [2]float64{38, 150}, [2]float64{162, 150}, cRed)
	imageutil.FillCircle(c, 100, 158, 62, cRed)
	punch(c, 100, 165, 55)
	imageutil.FillCircle(c, 100, 10, 10, cWhite) // 帽顶球
	return c
}

func genHatBeret() *imageutil.RGBA { // 蓝色贝雷帽
	c := imageutil.NewCanvas(200, 130)
	imageutil.FillCircle(c, 100, 68, 62, cBlue)
	punch(c, 100, 95, 52)
	imageutil.FillCircle(c, 100, 14, 9, cYellow) // 帽扣
	return c
}

func genHatStraw() *imageutil.RGBA { // 草帽：宽檐 + 帽顶 + 红带（FillRing 会打穿内圈，帽带用双圆叠加）
	c := imageutil.NewCanvas(200, 160)
	imageutil.FillCircle(c, 100, 80, 76, cStraw)
	punch(c, 100, 80, 56)
	imageutil.FillCircle(c, 100, 72, 62, cStrawD)
	imageutil.FillCircle(c, 100, 72, 48, cRed)
	imageutil.FillCircle(c, 100, 72, 38, cStrawD)
	return c
}

func genHatCrown() *imageutil.RGBA { // 金色皇冠
	c := imageutil.NewCanvas(200, 110)
	imageutil.FillRect(c, 36, 62, 128, 32, cYellow)
	imageutil.FillTriangle(c, [2]float64{36, 62}, [2]float64{68, 14}, [2]float64{100, 62}, cYellow)
	imageutil.FillTriangle(c, [2]float64{68, 62}, [2]float64{100, 10}, [2]float64{132, 62}, cYellow)
	imageutil.FillTriangle(c, [2]float64{100, 62}, [2]float64{132, 14}, [2]float64{164, 62}, cYellow)
	imageutil.FillCircle(c, 68, 18, 6, cRed)
	imageutil.FillCircle(c, 100, 13, 6, cRed)
	imageutil.FillCircle(c, 132, 18, 6, cRed)
	imageutil.FillCircle(c, 100, 78, 8, cRed)
	return c
}

func genHatSanta() *imageutil.RGBA { // 圣诞帽
	c := imageutil.NewCanvas(200, 190)
	imageutil.FillTriangle(c, [2]float64{58, 14}, [2]float64{44, 132}, [2]float64{158, 124}, cRed)
	imageutil.FillRoundRect(c, 28, 112, 132, 34, 17, cWhite)
	imageutil.FillCircle(c, 56, 16, 15, cWhite)
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

func genLashStar() *imageutil.RGBA { // 星星睫毛：睫毛 + 金色星光
	c := imageutil.NewCanvas(220, 60)
	for i := 0; i < 4; i++ {
		x := 24 + i*12
		imageutil.FillRect(c, x, 26, 4, 14, cDark)
		imageutil.FillCircle(c, float64(x)+2, 40, 3, cDark)
		x2 := 130 + i*12
		imageutil.FillRect(c, x2, 26, 4, 14, cDark)
		imageutil.FillCircle(c, float64(x2)+2, 40, 3, cDark)
	}
	imageutil.FillCircle(c, 14, 10, 6, cYellow)
	imageutil.FillCircle(c, 206, 10, 6, cYellow)
	imageutil.FillCircle(c, 14, 10, 2.5, cWhite)
	imageutil.FillCircle(c, 206, 10, 2.5, cWhite)
	return c
}

// expFace 肤色底垫：盖住底板原眼睛，让表情元素可叠加
func expFace(c *imageutil.RGBA) {
	imageutil.FillCircle(c, 50, 35, 30, cSkin)
	imageutil.FillCircle(c, 130, 35, 30, cSkin)
}

// heartAt 简易心形（两个圆 + 一个三角）
func heartAt(c *imageutil.RGBA, cx, cy, s float64, col color.Color) {
	imageutil.FillCircle(c, cx-s*0.36, cy-s*0.22, s*0.42, col)
	imageutil.FillCircle(c, cx+s*0.36, cy-s*0.22, s*0.42, col)
	imageutil.FillTriangle(c,
		[2]float64{cx - s*0.76, cy - s*0.1},
		[2]float64{cx + s*0.76, cy - s*0.1},
		[2]float64{cx, cy + s*0.82}, col)
}

func genExpSmile() *imageutil.RGBA { // 笑眼弯弯 ∩∩
	c := imageutil.NewCanvas(180, 70)
	expFace(c)
	imageutil.FillArc(c, 50, 38, 16, 6, 180, 360, cDark)
	imageutil.FillArc(c, 130, 38, 16, 6, 180, 360, cDark)
	return c
}

func genExpHeart() *imageutil.RGBA { // 爱心眼
	c := imageutil.NewCanvas(180, 70)
	expFace(c)
	heartAt(c, 50, 33, 24, cPink)
	heartAt(c, 130, 33, 24, cPink)
	return c
}

func genExpWink() *imageutil.RGBA { // 眨眼：左圆眼 + 右 ∩
	c := imageutil.NewCanvas(180, 70)
	expFace(c)
	imageutil.FillCircle(c, 50, 33, 10, cDark)
	imageutil.FillCircle(c, 53, 30, 3, cWhite)
	imageutil.FillArc(c, 130, 38, 16, 6, 180, 360, cDark)
	return c
}

func genExpPout() *imageutil.RGBA { // 嘟嘟嘴（盖住底板原嘴）
	c := imageutil.NewCanvas(80, 80)
	imageutil.FillCircle(c, 40, 42, 27, cSkin)
	imageutil.FillCircle(c, 40, 44, 8, cDark)
	imageutil.FillCircle(c, 40, 48, 4.5, cPink)
	return c
}

func genOutfitApron() *imageutil.RGBA { // 小围裙
	c := imageutil.NewCanvas(160, 230)
	imageutil.FillRect(c, 32, 6, 16, 58, cPink)
	imageutil.FillRect(c, 112, 6, 16, 58, cPink)
	imageutil.FillRoundRect(c, 20, 58, 120, 150, 16, cWhite)
	imageutil.FillRoundRect(c, 52, 120, 56, 48, 10, cPink)
	return c
}

func genOutfitCape() *imageutil.RGBA { // 红色小披风（梯形）
	c := imageutil.NewCanvas(200, 200)
	imageutil.FillTriangle(c, [2]float64{28, 16}, [2]float64{172, 16}, [2]float64{188, 192}, cRed)
	imageutil.FillTriangle(c, [2]float64{28, 16}, [2]float64{188, 192}, [2]float64{12, 192}, cRed)
	imageutil.FillCircle(c, 100, 14, 12, cWhite)
	return c
}

func genOutfitScarf() *imageutil.RGBA { // 蓝色围巾 + 垂尾
	c := imageutil.NewCanvas(170, 130)
	imageutil.FillRoundRect(c, 14, 18, 142, 46, 22, cBlue)
	imageutil.FillRoundRect(c, 96, 58, 40, 58, 12, cBlue)
	imageutil.FillRect(c, 100, 114, 7, 14, cBlue)
	imageutil.FillRect(c, 112, 114, 7, 14, cBlue)
	imageutil.FillRect(c, 124, 114, 7, 14, cBlue)
	return c
}

func genOutfitBowtie() *imageutil.RGBA { // 红色领结
	c := imageutil.NewCanvas(120, 70)
	imageutil.FillTriangle(c, [2]float64{56, 35}, [2]float64{6, 8}, [2]float64{6, 62}, cRed)
	imageutil.FillTriangle(c, [2]float64{64, 35}, [2]float64{114, 8}, [2]float64{114, 62}, cRed)
	imageutil.FillCircle(c, 60, 35, 13, cRedD)
	return c
}

func genBow() *imageutil.RGBA { // 蝴蝶结
	c := imageutil.NewCanvas(120, 80)
	imageutil.FillTriangle(c, [2]float64{58, 40}, [2]float64{10, 12}, [2]float64{10, 68}, cPink)
	imageutil.FillTriangle(c, [2]float64{62, 40}, [2]float64{110, 12}, [2]float64{110, 68}, cPink)
	imageutil.FillCircle(c, 60, 40, 12, cRed)
	return c
}

func genAccGlasses() *imageutil.RGBA { // 圆框眼镜（镜距对齐底板双眼 76px）
	c := imageutil.NewCanvas(190, 90)
	imageutil.FillRing(c, 55, 45, 32, 7, cDark)
	imageutil.FillRing(c, 135, 45, 32, 7, cDark)
	imageutil.FillRect(c, 86, 40, 18, 7, cDark)
	imageutil.FillRect(c, 0, 42, 24, 7, cDark)
	imageutil.FillRect(c, 166, 42, 24, 7, cDark)
	return c
}

func genAccBalloon() *imageutil.RGBA { // 粉色气球 + 线
	c := imageutil.NewCanvas(120, 240)
	imageutil.FillCircle(c, 60, 52, 46, cPink)
	imageutil.FillCircle(c, 44, 38, 10, cWhite)
	imageutil.FillTriangle(c, [2]float64{50, 102}, [2]float64{70, 102}, [2]float64{60, 114}, cPink)
	imageutil.FillRect(c, 58, 114, 3, 118, cString)
	return c
}

func genAccLollipop() *imageutil.RGBA { // 棒棒糖（同心环）
	c := imageutil.NewCanvas(120, 220)
	imageutil.FillRing(c, 60, 58, 46, 11, cPink)
	imageutil.FillRing(c, 60, 58, 30, 11, cWhite)
	imageutil.FillRing(c, 60, 58, 15, 9, cPink)
	imageutil.FillCircle(c, 60, 58, 6, cWhite)
	imageutil.FillRect(c, 57, 102, 6, 112, cWood)
	return c
}

func genAccStarclip() *imageutil.RGBA { // 星星发夹（四芒星光）
	c := imageutil.NewCanvas(90, 90)
	imageutil.FillTriangle(c, [2]float64{45, 6}, [2]float64{53, 38}, [2]float64{37, 38}, cYellow)
	imageutil.FillTriangle(c, [2]float64{45, 84}, [2]float64{53, 52}, [2]float64{37, 52}, cYellow)
	imageutil.FillTriangle(c, [2]float64{6, 45}, [2]float64{38, 37}, [2]float64{38, 53}, cYellow)
	imageutil.FillTriangle(c, [2]float64{84, 45}, [2]float64{52, 37}, [2]float64{52, 53}, cYellow)
	imageutil.FillCircle(c, 45, 45, 10, cWhite)
	return c
}
