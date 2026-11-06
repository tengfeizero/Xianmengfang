package repo

import (
	"errors"
	"fmt"

	"xmf/internal/model"

	"gorm.io/gorm"
)

// demoProduct 演示商品定义：新库整条写入，老库按标题回填空字段，两处共用同一份文案。
type demoProduct struct {
	catName     string
	title       string
	subtitle    string
	description string
	unit        string
	original    int64
	sort        int
	service     string
	detail      string
	image       string // 封面文件名前缀，对应 /static/products/<image>_main.png
	skus        []demoSKU
}

type demoSKU struct {
	specs string
	code  string
	price int64
	stock int64
}

func demoProducts() []demoProduct {
	return []demoProduct{
		{
			catName: "服饰", title: "定制图案T恤",
			subtitle:    "220g 精梳棉 · 不透不塌不变形",
			description: "纯棉 220g，把你的作品穿在身上",
			unit:        "件", original: 12900, sort: 100,
			service: "48 小时内发货 · 定制商品不支持七天无理由退货",
			detail: "【面料】220g 精梳棉，克重足、不透不塌，反复水洗不易变形。\n\n" +
				"【工艺】数码直喷印花，图案层次细腻、手感柔软，不易开裂掉色。\n\n" +
				"【尺码】M / L 两码，版型微宽松；165cm / 55kg 建议 M，175cm / 70kg 建议 L。\n\n" +
				"【注意】屏幕显示与成品存在轻微色差，以实际印制效果为准。",
			image: "tee",
			skus: []demoSKU{
				{specs: `{"尺码":"M","颜色":"白"}`, code: "TEE-W-M", price: 9900, stock: 100},
				{specs: `{"尺码":"L","颜色":"白"}`, code: "TEE-W-L", price: 9900, stock: 100},
				{specs: `{"尺码":"M","颜色":"黑"}`, code: "TEE-B-M", price: 9900, stock: 80},
			},
		},
		{
			catName: "周边", title: "定制图案马克杯",
			subtitle:    "350ml 陶瓷 · 图案 360° 包裹不脱落",
			description: "陶瓷白杯 350ml，作品图案热转印",
			unit:        "个", original: 7900, sort: 90,
			service: "48 小时内发货 · 定制商品不支持七天无理由退货",
			detail: "【材质】350ml 高温陶瓷，杯壁均匀，可进微波炉与洗碗机。\n\n" +
				"【工艺】热转印，图案 360° 包裹杯身，日常刷洗不掉色。\n\n" +
				"【使用】首次使用前用温水加少量洗涤剂清洗，避免钢丝球摩擦印花面。\n\n" +
				"【注意】屏幕显示与成品存在轻微色差，以实际印制效果为准。",
			image: "mug",
			skus:  []demoSKU{{specs: `{"规格":"标准"}`, code: "MUG-STD", price: 5900, stock: 200}},
		},
		{
			catName: "周边", title: "定制帆布袋",
			subtitle:    "12 安加厚帆布 · 装得下 14 寸笔记本",
			description: "加厚帆布，单面印制",
			unit:        "个", original: 8900, sort: 80,
			service: "48 小时内发货 · 定制商品不支持七天无理由退货",
			detail: "【面料】12 安加厚纯棉帆布，肩带加宽处理，负重不勒手。\n\n" +
				"【工艺】单面丝印，图案边缘清晰；另一面为纯色。\n\n" +
				"【容量】38 × 42cm，可放 14 寸笔记本 + A4 文件 + 水杯。\n\n" +
				"【注意】首次水洗可能轻微掉色，建议单独冷水手洗、阴干。",
			image: "bag",
			skus:  []demoSKU{{specs: `{"规格":"标准"}`, code: "BAG-STD", price: 6900, stock: 150}},
		},
		{
			catName: "玩偶", title: "定制小姑娘玩偶",
			subtitle:    "20cm 短毛绒 · 按你的作品形象打版",
			description: "按作品形象定制毛绒玩偶，20cm",
			unit:        "个", original: 25900, sort: 70,
			service: "3–5 个工作日制作后发货 · 定制商品不支持七天无理由退货",
			detail: "【材质】短毛绒面料 + PP 棉填充，柔软可抱，符合 GB 6675 玩具安全标准。\n\n" +
				"【尺寸】坐高 20cm，适合摆件或随身携带。\n\n" +
				"【工艺】按你提交的作品形象分件打版、人工缝制，五官与服饰尽量还原原图。\n\n" +
				"【周期】定制款需 3–5 个工作日制作，下单后不支持改图，请确认后再下单。",
			image: "doll",
			skus:  []demoSKU{{specs: `{"规格":"20cm"}`, code: "DOLL-20", price: 19900, stock: 50}},
		},
	}
}

func demoImages(name string) string {
	return fmt.Sprintf(`["/static/products/%s_main.png","/static/products/%s_detail.png"]`, name, name)
}

// SeedShop 商品分类与演示商品。
// 空库整条写入；非空库只对"标题匹配且字段为空"的演示商品做回填，
// 不覆盖后台已改过的内容，也不碰用户自建商品。
func SeedShop(db *gorm.DB) error {
	var catCount int64
	if err := db.Model(&model.ProductCategory{}).Count(&catCount).Error; err != nil {
		return err
	}
	var cats []model.ProductCategory
	if catCount == 0 {
		cats = []model.ProductCategory{
			{Name: "服饰", Sort: 1, Status: 1},
			{Name: "周边", Sort: 2, Status: 1},
			{Name: "玩偶", Sort: 3, Status: 1},
		}
		if err := db.Create(&cats).Error; err != nil {
			return err
		}
	} else if err := db.Order("sort, id").Find(&cats).Error; err != nil {
		return err
	}

	catID := func(name string) int64 {
		for _, c := range cats {
			if c.Name == name {
				return c.ID
			}
		}
		if len(cats) > 0 {
			return cats[0].ID
		}
		return 0
	}

	for _, d := range demoProducts() {
		var p model.Product
		err := db.Where("title = ?", d.title).First(&p).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			p = model.Product{
				CategoryID: catID(d.catName), Title: d.title, Subtitle: d.subtitle,
				Description: d.description, Images: demoImages(d.image),
				OriginalPriceCents: d.original, Unit: d.unit, Detail: d.detail,
				DetailImages: "[]", ServiceNotes: d.service, Sort: d.sort, Status: 1,
			}
			if err := db.Create(&p).Error; err != nil {
				return err
			}
			for _, s := range d.skus {
				sku := model.SKU{ProductID: p.ID, Specs: s.specs, PriceCents: s.price, Stock: s.stock, Code: s.code}
				if err := db.Create(&sku).Error; err != nil {
					return err
				}
			}
			continue
		}
		if err != nil {
			return err
		}

		// 老库回填：只补空字段，后台改过的值一律尊重
		updates := map[string]any{}
		if p.Subtitle == "" {
			updates["subtitle"] = d.subtitle
		}
		if p.Images == "" || p.Images == "[]" {
			updates["images"] = demoImages(d.image)
		}
		if p.OriginalPriceCents == 0 {
			updates["original_price_cents"] = d.original
		}
		if p.Unit == "" {
			updates["unit"] = d.unit
		}
		if p.Detail == "" {
			updates["detail"] = d.detail
		}
		if p.DetailImages == "" {
			updates["detail_images"] = "[]"
		}
		if p.ServiceNotes == "" {
			updates["service_notes"] = d.service
		}
		if p.Sort == 0 {
			updates["sort"] = d.sort
		}
		if len(updates) > 0 {
			if err := db.Model(&model.Product{}).Where("id = ?", p.ID).Updates(updates).Error; err != nil {
				return err
			}
		}

		var skuCount int64
		if err := db.Model(&model.SKU{}).Where("product_id = ?", p.ID).Count(&skuCount).Error; err != nil {
			return err
		}
		if skuCount == 0 {
			for _, s := range d.skus {
				sku := model.SKU{ProductID: p.ID, Specs: s.specs, PriceCents: s.price, Stock: s.stock, Code: s.code}
				if err := db.Create(&sku).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}
