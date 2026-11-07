package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"math/rand"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"xmf/internal/imageutil"
	"xmf/internal/model"
	"xmf/internal/service"

	"github.com/gin-gonic/gin"
)

// ---- 上传 ----

var allowExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

func (h *Handlers) upload(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		fail(c, 400, "请选择要上传的文件")
		return
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowExt[ext] {
		fail(c, 400, "仅支持 jpg/png/webp 图片")
		return
	}
	if fh.Size > 10*1024*1024 {
		fail(c, 400, "图片不能超过 10MB")
		return
	}
	name := fmt.Sprintf("up_%s_%04d%s", time.Now().Format("20060102150405"), rand.Intn(10000), ext)
	if err := c.SaveUploadedFile(fh, filepath.Join(h.UploadsDir, name)); err != nil {
		fail(c, 500, "上传失败，请稍后再试")
		return
	}
	ok(c, gin.H{"url": "/static/" + name})
}

// ---- 拼装素材 ----

func (h *Handlers) collageBoards(c *gin.Context) {
	var boards []model.Board
	if err := h.DB.Where("status = 1").Order("id").Find(&boards).Error; err != nil {
		fail(c, 500, "获取底板失败，请稍后再试")
		return
	}
	ok(c, boards)
}

// collageElements 用户端元素列表：按底板取（一对多）。
// board_id 必填——元素是挂在具体底板画布上的素材，跨底板混用会错位/串风格。
func (h *Handlers) collageElements(c *gin.Context) {
	boardID, err := strconv.ParseInt(c.Query("board_id"), 10, 64)
	if err != nil || boardID <= 0 {
		fail(c, 400, "请先选择底板")
		return
	}
	q := h.DB.Where("status = 1 AND board_id = ?", boardID).Order("category, sort, id")
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	var els []model.Element
	if err := q.Find(&els).Error; err != nil {
		fail(c, 500, "获取素材失败，请稍后再试")
		return
	}
	ok(c, els)
}

// ---- 拼装合成 ----

type placementReq struct {
	ElementID int64   `json:"element_id" binding:"required"`
	X         float64 `json:"x" binding:"required"`
	Y         float64 `json:"y" binding:"required"`
	Scale     float64 `json:"scale"`
}

type composeReq struct {
	BoardID    int64          `json:"board_id" binding:"required"`
	Placements []placementReq `json:"placements" binding:"required,min=1"`
}

func (h *Handlers) collageCompose(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req composeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if len(req.Placements) > 12 {
		fail(c, 400, "元素数量太多啦（最多 12 个）")
		return
	}
	var board model.Board
	if err := h.DB.Where("status = 1").First(&board, req.BoardID).Error; err != nil {
		fail(c, 404, "底板不存在或已下架")
		return
	}
	boardImg, err := h.loadStaticImage(board.URL)
	if err != nil {
		fail(c, 500, "底板素材缺失，请联系管理员")
		return
	}
	placements := make([]imageutil.Placement, 0, len(req.Placements))
	for _, p := range req.Placements {
		var el model.Element
		if err := h.DB.Where("status = 1").First(&el, p.ElementID).Error; err != nil {
			fail(c, 404, fmt.Sprintf("素材不存在或已下架（id=%d）", p.ElementID))
			return
		}
		// 一对多归属校验：元素必须是该底板下的素材，防止跨底板拼出风格错乱的图
		if el.BoardID != board.ID {
			fail(c, 400, fmt.Sprintf("「%s」不属于当前底板，请重新选择", el.Name))
			return
		}
		img, err := h.loadStaticImage(el.ImageURL)
		if err != nil {
			fail(c, 500, "素材文件缺失，请联系管理员")
			return
		}
		scale := p.Scale
		if scale <= 0 {
			scale = 1
		}
		if scale > 4 {
			fail(c, 400, "元素放缩超出范围")
			return
		}
		placements = append(placements, imageutil.Placement{
			Img: img, RelX: p.X, RelY: p.Y, Scale: scale,
			Z: categoryZ(el.Category),
		})
	}
	// 图层顺序按分类固定（与游戏换装一致）：服装 < 头发 < 表情 < 睫毛 < 帽子 < 配饰，
	// 用户添加顺序不再影响遮挡关系
	sort.SliceStable(placements, func(i, j int) bool { return placements[i].Z < placements[j].Z })
	out, err := imageutil.Compose(boardImg, placements)
	if err != nil {
		fail(c, 500, "作品合成失败，请稍后再试")
		return
	}
	name := fmt.Sprintf("collage_%s_%04d.png", time.Now().Format("20060102150405"), rand.Intn(10000))
	if err := imageutil.SavePNG(out, filepath.Join(h.UploadsDir, name)); err != nil {
		fail(c, 500, "作品保存失败，请稍后再试")
		return
	}
	meta, _ := json.Marshal(req)
	work := &model.Work{
		UserID: claims.UserID, Type: "collage",
		ImageURL: "/static/" + name, ThumbnailURL: "/static/" + name,
		Meta: string(meta), Status: 1,
	}
	if err := h.DB.Create(work).Error; err != nil {
		fail(c, 500, "作品保存失败，请稍后再试")
		return
	}
	ok(c, work)
}

// loadStaticImage 把 /static/xxx 映射到本地 uploads 并解码
func (h *Handlers) loadStaticImage(url string) (image.Image, error) {
	if !strings.HasPrefix(url, "/static/") || strings.Contains(url, "..") {
		return nil, errors.New("素材地址不合法")
	}
	return imageutil.LoadPNG(filepath.Join(h.UploadsDir, strings.TrimPrefix(url, "/static/")))
}

// ---- AI 生成 ----

type aiTaskReq struct {
	Type     string `json:"type" binding:"required"`
	Prompt   string `json:"prompt"`
	PhotoURL string `json:"photo_url"`
}

func (h *Handlers) aiCreateTask(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var req aiTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	task, work, err := h.AI.CreateTask(claims.UserID, req.Type, req.Prompt, req.PhotoURL)
	if err != nil {
		if errors.Is(err, service.ErrQuotaExceeded) {
			fail(c, 429, err.Error())
			return
		}
		fail(c, 400, err.Error())
		return
	}
	ok(c, gin.H{"task": task, "work": work})
}

// ---- 作品库 ----

// workList 我的作品库；scope=liked 时返回我点赞过的公开作品（商品页选图用）
func (h *Handlers) workList(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	if c.Query("scope") == "liked" {
		works, err := h.Community.LikedWorks(claims.UserID)
		if err != nil {
			fail(c, 500, "获取点赞作品失败，请稍后再试")
			return
		}
		ok(c, works)
		return
	}
	var works []model.Work
	if err := h.DB.Where("user_id = ? AND status = 1", claims.UserID).Order("id desc").Find(&works).Error; err != nil {
		fail(c, 500, "获取作品失败，请稍后再试")
		return
	}
	ok(c, works)
}

func (h *Handlers) workDetail(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var w model.Work
	// 自己的作品，或已发布到广场的他人作品（v1.7）
	if err := h.DB.First(&w, c.Param("id")).Error; err != nil || w.Status != 1 ||
		(w.UserID != claims.UserID && w.IsPublic != 1) {
		fail(c, 404, "作品不存在")
		return
	}
	ok(c, w)
}

func (h *Handlers) workDelete(c *gin.Context) {
	claims := c.MustGet("claims").(*service.Claims)
	var w model.Work
	if err := h.DB.First(&w, c.Param("id")).Error; err != nil {
		fail(c, 404, "作品不存在")
		return
	}
	if w.UserID != claims.UserID {
		fail(c, 403, "无权操作该作品")
		return
	}
	if err := h.DB.Model(&w).Update("status", 0).Error; err != nil {
		fail(c, 500, "删除失败，请稍后再试")
		return
	}
	ok(c, nil)
}

// ---- 后台：素材与底板管理 ----

type boardReq struct {
	Name   string `json:"name" binding:"required"`
	URL    string `json:"url" binding:"required"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func (h *Handlers) adminBoardCreate(c *gin.Context) {
	var req boardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	b := model.Board{Name: req.Name, URL: req.URL, Width: req.Width, Height: req.Height, Status: 1}
	if err := h.DB.Create(&b).Error; err != nil {
		fail(c, 500, "创建失败，请稍后再试")
		return
	}
	ok(c, b)
}

func (h *Handlers) adminBoardUpdate(c *gin.Context) {
	var req boardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	res := h.DB.Model(&model.Board{}).Where("id = ?", c.Param("id")).Updates(map[string]any{
		"name": req.Name, "url": req.URL, "width": req.Width, "height": req.Height,
	})
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "底板不存在或更新失败")
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminBoardDelete(c *gin.Context) {
	id := paramID(c)
	// 一对多保护：底板下还有启用元素时不允许下架，否则会出现「底板不可选、元素仍挂在启用列表」
	var n int64
	if err := h.DB.Model(&model.Element{}).Where("board_id = ? AND status = 1", id).Count(&n).Error; err != nil {
		fail(c, 500, "下架失败，请稍后再试")
		return
	}
	if n > 0 {
		fail(c, 409, fmt.Sprintf("该底板下还有 %d 个启用元素，请先下架元素", n))
		return
	}
	res := h.DB.Model(&model.Board{}).Where("id = ?", id).Update("status", 0)
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "底板不存在或下架失败")
		return
	}
	ok(c, nil)
}

type elementReq struct {
	BoardID  int64  `json:"board_id" binding:"required"`
	Category string `json:"category" binding:"required"`
	Name     string `json:"name" binding:"required"`
	ImageURL string `json:"image_url" binding:"required"`
	Anchor   string `json:"anchor"`
	Sort     int    `json:"sort"`
}

// checkElementBoard 校验元素归属底板存在且上架；返回 false 时已写好响应
func (h *Handlers) checkElementBoard(c *gin.Context, boardID int64) bool {
	var n int64
	if err := h.DB.Model(&model.Board{}).Where("id = ? AND status = 1", boardID).Count(&n).Error; err != nil {
		fail(c, 500, "校验底板失败，请稍后再试")
		return false
	}
	if n == 0 {
		fail(c, 400, "底板不存在或已下架")
		return false
	}
	return true
}

func (h *Handlers) adminElementCreate(c *gin.Context) {
	var req elementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if !validElementCategory(req.Category) {
		fail(c, 400, "元素分类不合法")
		return
	}
	if !h.checkElementBoard(c, req.BoardID) {
		return
	}
	if req.Anchor == "" {
		req.Anchor = `{"x":0.5,"y":0.5,"scale":1}`
	}
	el := model.Element{
		BoardID: req.BoardID, Category: req.Category, Name: req.Name,
		ImageURL: req.ImageURL, Anchor: req.Anchor, Sort: req.Sort, Status: 1,
	}
	if err := h.DB.Create(&el).Error; err != nil {
		fail(c, 500, "创建失败，请稍后再试")
		return
	}
	ok(c, el)
}

func (h *Handlers) adminElementUpdate(c *gin.Context) {
	var req elementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if !validElementCategory(req.Category) {
		fail(c, 400, "元素分类不合法")
		return
	}
	if !h.checkElementBoard(c, req.BoardID) {
		return
	}
	res := h.DB.Model(&model.Element{}).Where("id = ?", c.Param("id")).Updates(map[string]any{
		"board_id": req.BoardID, "category": req.Category, "name": req.Name, "image_url": req.ImageURL,
		"anchor": req.Anchor, "sort": req.Sort,
	})
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "元素不存在或更新失败")
		return
	}
	ok(c, nil)
}

func (h *Handlers) adminElementDelete(c *gin.Context) {
	res := h.DB.Model(&model.Element{}).Where("id = ?", c.Param("id")).Update("status", 0)
	if res.Error != nil || res.RowsAffected == 0 {
		fail(c, 404, "元素不存在或下架失败")
		return
	}
	ok(c, nil)
}

func validElementCategory(cat string) bool {
	switch cat {
	case "hair", "hat", "lash", "expression", "outfit", "accessory":
		return true
	}
	return false
}

// categoryZ 拼装图层顺序（小者底层）：服装 < 头发 < 表情 < 睫毛 < 帽子 < 配饰
func categoryZ(cat string) int {
	switch cat {
	case "outfit":
		return 0
	case "hair":
		return 1
	case "expression":
		return 2
	case "lash":
		return 3
	case "hat":
		return 4
	case "accessory":
		return 5
	}
	return 3
}
