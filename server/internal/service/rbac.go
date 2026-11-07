package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"xmf/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 权限码：后端中间件按码拦截，前端按码渲染菜单/按钮。
// 命名约定「资源:动作」，只增不改（改码会导致已有角色失效）。
type PermItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type PermGroup struct {
	Key   string     `json:"key"`
	Name  string     `json:"name"`
	Items []PermItem `json:"items"`
}

var PermGroups = []PermGroup{
	{Key: "order", Name: "订单", Items: []PermItem{
		{"order:view", "查看订单列表与详情"},
		{"order:process", "接单 / 生产完成 / 打包完成"},
		{"order:ship", "发货（填物流单号）"},
		{"order:cancel", "取消订单"},
		{"order:refund", "标记退款"},
	}},
	{Key: "product", Name: "商品", Items: []PermItem{
		{"product:view", "查看商品/分类/规格"},
		{"product:edit", "编辑商品/分类/规格"},
	}},
	{Key: "asset", Name: "素材", Items: []PermItem{
		{"asset:view", "查看底板与元素"},
		{"asset:edit", "编辑底板与元素"},
	}},
	{Key: "user", Name: "用户", Items: []PermItem{
		{"user:view", "查看用户"},
		{"user:edit", "启用/禁用用户"},
	}},
	{Key: "work", Name: "作品", Items: []PermItem{
		{"work:view", "查看作品"},
		{"work:edit", "上架/下架作品"},
	}},
	{Key: "config", Name: "系统配置", Items: []PermItem{
		{"config:view", "查看配置"},
		{"config:edit", "修改配置（含 AI 密钥）"},
	}},
	{Key: "system", Name: "权限与日志", Items: []PermItem{
		{"admin:view", "查看管理员账号"},
		{"admin:edit", "管理管理员账号（新增/改角色/禁用/重置密码）"},
		{"role:view", "查看角色"},
		{"role:edit", "管理角色（新增/改权限/删除）"},
		{"audit:view", "查看敏感操作日志"},
	}},
}

// PermAll 全部权限码集合（用于校验角色提交的权限）
func PermAll() map[string]bool {
	m := map[string]bool{}
	for _, g := range PermGroups {
		for _, it := range g.Items {
			m[it.Code] = true
		}
	}
	return m
}

// 角色相关错误
var (
	ErrRoleNotFound   = errors.New("角色不存在")
	ErrRoleBuiltin    = errors.New("内置角色不允许删除")
	ErrRoleInUse      = errors.New("该角色下还有管理员账号，请先调整账号角色")
	ErrRoleCodeDup    = errors.New("角色标识已存在")
	ErrRoleNameEmpty  = errors.New("角色名称不能为空")
	ErrRolePermEmpty  = errors.New("请至少选择一项权限")
	ErrRolePermBad    = errors.New("包含不识别的权限码")
	ErrAdminNotFound  = errors.New("管理员不存在")
	ErrAdminDup       = errors.New("用户名已存在")
	ErrAdminPassword  = errors.New("密码长度至少 6 位")
	ErrAdminRoleBad   = errors.New("所选角色不存在或已停用")
	ErrLastSuper      = errors.New("必须保留至少一个启用状态的超级管理员")
	ErrSelfDisable    = errors.New("不能禁用或降级当前登录账号")
	ErrSelfDelete     = errors.New("不能删除当前登录账号")
	ErrAdminRoleUnset = errors.New("请选择角色")
)

// RBACService 角色与管理员账号
type RBACService struct {
	DB *gorm.DB
}

// IsSuper 超级管理员：内置 super 角色或持有通配权限
func IsSuper(role *model.Role) bool {
	if role == nil {
		return false
	}
	if role.Code == "super" {
		return true
	}
	for _, p := range ParsePerms(role.Perms) {
		if p == "*" {
			return true
		}
	}
	return false
}

// ParsePerms 角色权限 JSON → 码切片（非法 JSON 返回空，不让脏数据放开权限）
func ParsePerms(raw string) []string {
	var codes []string
	if raw == "" {
		return codes
	}
	if err := json.Unmarshal([]byte(raw), &codes); err != nil {
		return nil
	}
	return codes
}

// HasPerm 判定角色是否有某权限
func HasPerm(role *model.Role, code string) bool {
	if IsSuper(role) {
		return true
	}
	for _, p := range ParsePerms(role.Perms) {
		if p == code {
			return true
		}
	}
	return false
}

// RoleWithPerms 角色 + 解析后的权限码，供接口返回
type RoleWithPerms struct {
	model.Role
	Perms []string `json:"perms"`
	// 使用该角色的管理员数（列表页展示，删除前可见影响面）
	AdminCount int64 `json:"admin_count"`
}

func (s *RBACService) ListRoles() ([]RoleWithPerms, error) {
	var roles []model.Role
	if err := s.DB.Order("builtin desc, id asc").Find(&roles).Error; err != nil {
		return nil, err
	}
	out := make([]RoleWithPerms, 0, len(roles))
	for _, r := range roles {
		var n int64
		_ = s.DB.Model(&model.Admin{}).Where("role_id = ?", r.ID).Count(&n).Error
		out = append(out, RoleWithPerms{Role: r, Perms: ParsePerms(r.Perms), AdminCount: n})
	}
	return out, nil
}

func (s *RBACService) GetRole(id int64) (*model.Role, error) {
	var r model.Role
	if err := s.DB.First(&r, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &r, nil
}

func normalizePerms(codes []string) (string, error) {
	all := PermAll()
	seen := map[string]bool{}
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		c = strings.TrimSpace(c)
		if c == "" || seen[c] {
			continue
		}
		if c != "*" && !all[c] {
			return "", fmt.Errorf("%w: %s", ErrRolePermBad, c)
		}
		seen[c] = true
		out = append(out, c)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return "", ErrRolePermEmpty
	}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func slugCode(name string) string {
	return fmt.Sprintf("role_%d", time.Now().UnixNano()/1e6)
}

// CreateRole 新建自定义角色（内置角色只能由种子创建）
func (s *RBACService) CreateRole(name, remark string, perms []string) (*model.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrRoleNameEmpty
	}
	raw, err := normalizePerms(perms)
	if err != nil {
		return nil, err
	}
	code := slugCode(name)
	for i := 0; i < 5; i++ {
		var n int64
		_ = s.DB.Model(&model.Role{}).Where("code = ?", code).Count(&n).Error
		if n == 0 {
			break
		}
		code = slugCode(name)
	}
	r := &model.Role{Code: code, Name: name, Remark: remark, Perms: raw, Status: 1}
	if err := s.DB.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

// UpdateRole 修改角色。内置超管角色不允许改权限（避免把自己锁死）；
// 内置普通角色可改权限与名称，但不可改标识。
func (s *RBACService) UpdateRole(id int64, name, remark string, perms []string, status int) error {
	r, err := s.GetRole(id)
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrRoleNameEmpty
	}
	if r.Code == "super" {
		// 超管角色只能改名/备注，权限恒为全量
		return s.DB.Model(r).Updates(map[string]any{"name": name, "remark": remark}).Error
	}
	raw, err := normalizePerms(perms)
	if err != nil {
		return err
	}
	if status != 0 {
		status = 1
	}
	return s.DB.Model(r).Updates(map[string]any{
		"name": name, "remark": remark, "perms": raw, "status": status,
	}).Error
}

// DeleteRole 删除角色：内置角色、仍有账号使用的角色禁止删除
func (s *RBACService) DeleteRole(id int64) error {
	r, err := s.GetRole(id)
	if err != nil {
		return err
	}
	if r.Builtin == 1 {
		return ErrRoleBuiltin
	}
	var n int64
	if err := s.DB.Model(&model.Admin{}).Where("role_id = ?", id).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return ErrRoleInUse
	}
	return s.DB.Delete(&model.Role{}, id).Error
}

// AdminView 管理员列表视图：带角色名与权限码，前端直接渲染
type AdminView struct {
	model.Admin
	RoleName string   `json:"role_name"`
	RoleCode string   `json:"role_code"`
	Perms    []string `json:"perms"`
}

func (s *RBACService) ListAdmins() ([]AdminView, error) {
	var admins []model.Admin
	if err := s.DB.Order("id asc").Find(&admins).Error; err != nil {
		return nil, err
	}
	roles := map[int64]model.Role{}
	var rs []model.Role
	if err := s.DB.Find(&rs).Error; err == nil {
		for _, r := range rs {
			roles[r.ID] = r
		}
	}
	out := make([]AdminView, 0, len(admins))
	for _, a := range admins {
		v := AdminView{Admin: a}
		if r, ok := roles[a.RoleID]; ok {
			v.RoleName, v.RoleCode = r.Name, r.Code
			v.Perms = ParsePerms(r.Perms)
		}
		out = append(out, v)
	}
	return out, nil
}

// GetAdminWithRole 取账号与其角色（鉴权中间件每请求调用，保证改权限/禁用即时生效）
func (s *RBACService) GetAdminWithRole(id int64) (*model.Admin, *model.Role, error) {
	var a model.Admin
	if err := s.DB.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrAdminNotFound
		}
		return nil, nil, err
	}
	if a.RoleID == 0 {
		return &a, nil, nil
	}
	var r model.Role
	if err := s.DB.First(&r, a.RoleID).Error; err != nil {
		return &a, nil, nil
	}
	return &a, &r, nil
}

func (s *RBACService) CreateAdmin(username, password, nickname string, roleID int64) (*model.Admin, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}
	if len(password) < 6 {
		return nil, ErrAdminPassword
	}
	if _, err := s.GetRole(roleID); err != nil {
		return nil, ErrAdminRoleBad
	}
	var n int64
	if err := s.DB.Model(&model.Admin{}).Where("username = ?", username).Count(&n).Error; err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, ErrAdminDup
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	a := &model.Admin{Username: username, PasswordHash: string(hash), Nickname: nickname, RoleID: roleID, Status: 1}
	if err := s.DB.Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

// UpdateAdmin 修改账号：昵称/角色/状态/密码（密码留空不改）。
// 保护规则：不能禁用或降级自己；系统必须保留至少一个启用的超管。
func (s *RBACService) UpdateAdmin(operatorID, id int64, nickname string, roleID int64, status int, password string) error {
	a, _, err := s.GetAdminWithRole(id)
	if err != nil {
		return err
	}
	role, err := s.GetRole(roleID)
	if err != nil {
		return ErrAdminRoleBad
	}
	if status != 0 {
		status = 1
	}
	oldRole, _ := s.GetRole(a.RoleID)

	loseSuper := oldRole != nil && IsSuper(oldRole) && (!IsSuper(role) || status == 0)
	if loseSuper {
		if id == operatorID {
			return ErrSelfDisable
		}
		var n int64
		if err := s.DB.Model(&model.Admin{}).
			Where("role_id = ? AND status = 1 AND id <> ?", a.RoleID, id).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			return ErrLastSuper
		}
	}

	updates := map[string]any{"nickname": nickname, "role_id": roleID, "status": status}
	if password != "" {
		if len(password) < 6 {
			return ErrAdminPassword
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password_hash"] = string(hash)
	}
	return s.DB.Model(a).Updates(updates).Error
}

// DeleteAdmin 删除账号（禁止删自己，禁止删最后一个启用的超管）
func (s *RBACService) DeleteAdmin(operatorID, id int64) error {
	if id == operatorID {
		return ErrSelfDelete
	}
	a, role, err := s.GetAdminWithRole(id)
	if err != nil {
		return err
	}
	if role != nil && IsSuper(role) {
		var n int64
		if err := s.DB.Model(&model.Admin{}).
			Where("role_id = ? AND status = 1 AND id <> ?", a.RoleID, id).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			return ErrLastSuper
		}
	}
	return s.DB.Delete(&model.Admin{}, id).Error
}

// TouchLogin 记录最近登录时间（登录成功后调用）
func (s *RBACService) TouchLogin(id int64) {
	now := time.Now()
	_ = s.DB.Model(&model.Admin{}).Where("id = ?", id).Update("last_login_at", now).Error
}
