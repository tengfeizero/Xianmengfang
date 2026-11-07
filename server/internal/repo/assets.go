package repo

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"xmf/internal/model"

	"gorm.io/gorm"
)

// aipack AI 生成的整套拼装素材（4 底板 + 24 部件，透明背景 400x500 画布）。
// 由 scripts/ai_assets.py + scripts/extract_elements.py 从 ImageGen 产出差分抠图，见 docs/CHANGELOG v1.2。
//
//go:embed aipack
var aipackFS embed.FS

// SeedAssets 把内嵌素材落盘到 uploads（已存在则跳过，不覆盖后台上传的同名文件）。
func SeedAssets(uploadsDir string) error {
	return fs.WalkDir(aipackFS, "aipack", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := filepath.ToSlash(p[len("aipack/"):]) // xxx.png 或 elements/xxx.png
		dst := filepath.Join(uploadsDir, rel)
		if _, err := os.Stat(dst); err == nil {
			return nil
		}
		data, err := aipackFS.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
}

func strings_trim_aipack(p string) string {
	if len(p) > len("aipack/") && p[:len("aipack/")] == "aipack/" {
		return p[len("aipack/"):]
	}
	return p
}

// demoElement 演示部件（名字/文件/锚点均与 aipack 内素材一一对应）
type demoElement struct {
	cat  string
	name string
	file string
	x, y float64
	sort int
}

func demoElements() []demoElement {
	return []demoElement{
		// hair 头发（叠加式发型，底板自带波波头为基底）
		{"hair", "双马尾", "hair_twin.png", 0.5, 0.269, 1},
		{"hair", "侧马尾", "hair_side.png", 0.704, 0.297, 2},
		{"hair", "丸子头", "hair_bun.png", 0.503, 0.05, 3},
		{"hair", "双丸子", "hair_buns.png", 0.5, 0.143, 4},
		{"hair", "卷发", "hair_curly.png", 0.501, 0.446, 5},
		// hat 帽子
		{"hat", "红色尖尖帽", "hat_pointy.png", 0.505, 0.172, 1},
		{"hat", "蓝色贝雷帽", "hat_beret.png", 0.562, 0.161, 2},
		{"hat", "黄色草帽", "hat_straw.png", 0.504, 0.218, 3},
		{"hat", "金色皇冠", "hat_crown.png", 0.5, 0.086, 4},
		{"hat", "圣诞帽", "hat_santa.png", 0.454, 0.194, 5},
		// lash 睫毛
		{"lash", "星星睫毛", "lash_star.png", 0.502, 0.315, 1},
		// expression 表情
		{"expression", "笑眼弯弯", "exp_smile.png", 0.501, 0.335, 1},
		{"expression", "爱心眼", "exp_heart.png", 0.502, 0.342, 2},
		{"expression", "眨眼", "exp_wink.png", 0.587, 0.334, 3},
		{"expression", "嘟嘟嘴", "exp_pout.png", 0.501, 0.405, 4},
		// outfit 服装
		{"outfit", "小围裙", "outfit_apron.png", 0.5, 0.607, 1},
		{"outfit", "红色小披风", "outfit_cape.png", 0.503, 0.618, 2},
		{"outfit", "蓝色围巾", "outfit_scarf.png", 0.501, 0.544, 3},
		{"outfit", "红色领结", "outfit_bowtie.png", 0.502, 0.521, 4},
		// accessory 配饰
		{"accessory", "蝴蝶结", "acc_bow.png", 0.359, 0.161, 1},
		{"accessory", "圆框眼镜", "acc_glasses.png", 0.501, 0.339, 2},
		{"accessory", "粉色气球", "acc_balloon.png", 0.795, 0.195, 3},
		{"accessory", "棒棒糖", "acc_lollipop.png", 0.302, 0.532, 4},
		{"accessory", "星星发夹", "acc_starclip.png", 0.644, 0.23, 5},
	}
}

// 旧版几何占位素材的遗留记录名，种子时删除
var legacyElementNames = []string{"波波头", "长直发", "可爱睫毛"}

// demoBoardSet 每块底板可用的元素范围（一对多：元素挂在具体底板下）。
// 现有 aipack 部件是在小姑娘画布上产出的，因此：
//   - 小姑娘底板：全部 24 个部件
//   - 小男孩/小兔子/小熊猫：只放中性部件（帽子/表情/配饰），不放头发、裙子、睫毛等女装向素材
type demoBoardSet struct {
	name string
	cats []string
	skip map[string]bool
}

func demoBoardSets() []demoBoardSet {
	neutral := []string{"hat", "expression", "accessory"}
	return []demoBoardSet{
		{name: "底板·小姑娘", cats: []string{"hair", "hat", "lash", "expression", "outfit", "accessory"}},
		{name: "底板·小男孩", cats: neutral, skip: map[string]bool{"蝴蝶结": true}},
		{name: "底板·小兔子", cats: neutral},
		{name: "底板·小熊猫", cats: neutral},
	}
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// SeedCatalogAssets 写入底板与元素的数据库记录。
// 演示素材升级为 AI 生成套件（v1.2）：同名记录直接更新文件与锚点（素材已整体换代，
// 旧锚点不再适用）；旧版遗留记录删除；后台自建的非演示记录不受影响。
// v1.6 起元素归属底板：老库里 board_id=0 的元素统一归到「底板·小姑娘」（它们本就是小姑娘画布上抠出来的）。
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

	// 底板名 → id（后面按底板归置元素）
	var all []model.Board
	if err := db.Find(&all).Error; err != nil {
		return err
	}
	boardID := make(map[string]int64, len(all))
	for _, b := range all {
		boardID[b.Name] = b.ID
	}

	for _, legacy := range legacyElementNames {
		if err := db.Where("name = ?", legacy).Delete(&model.Element{}).Error; err != nil {
			return err
		}
	}

	// 老数据迁移：v1.6 之前元素没有归属，全部来自小姑娘画布 → 归到小姑娘底板
	girlID := boardID["底板·小姑娘"]
	if girlID > 0 {
		if err := db.Model(&model.Element{}).Where("board_id = 0").Update("board_id", girlID).Error; err != nil {
			return err
		}
	}

	for _, set := range demoBoardSets() {
		bid := boardID[set.name]
		if bid == 0 {
			continue
		}
		for _, e := range demoElements() {
			if !contains(set.cats, e.cat) || set.skip[e.name] {
				continue
			}
			anchor := fmt.Sprintf(`{"x":%s,"y":%s,"scale":1}`,
				strconv.FormatFloat(e.x, 'f', -1, 64), strconv.FormatFloat(e.y, 'f', -1, 64))
			var el model.Element
			err := db.Where("board_id = ? AND name = ?", bid, e.name).First(&el).Error
			if err == gorm.ErrRecordNotFound {
				el = model.Element{
					BoardID: bid, Category: e.cat, Name: e.name,
					ImageURL: "/static/elements/" + e.file, Anchor: anchor, Sort: e.sort, Status: 1,
				}
				if err := db.Create(&el).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
			if err := db.Model(&model.Element{}).Where("id = ?", el.ID).Updates(map[string]any{
				"category": e.cat, "image_url": "/static/elements/" + e.file, "anchor": anchor, "sort": e.sort,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
