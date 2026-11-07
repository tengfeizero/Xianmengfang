package service

import (
	"strings"
	"testing"

	"xmf/internal/model"
)

func mkRole(code, perms string) *model.Role {
	return &model.Role{Code: code, Perms: perms}
}

func TestMaskDetail(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		mustHas []string
		mustNot []string
	}{
		{
			name:    "登录密码脱敏",
			raw:     `{"username":"admin","password":"admin123"}`,
			mustHas: []string{`"username":"admin"`, `"password":"***"`},
			mustNot: []string{"admin123"},
		},
		{
			name:    "配置密钥脱敏但保留键名",
			raw:     `{"key":"ai_api_key","value":"sk-abcdef123456"}`,
			mustHas: []string{`"key":"ai_api_key"`, `"value":"***"`},
			mustNot: []string{"sk-abcdef123456"},
		},
		{
			name:    "嵌套对象与数组",
			raw:     `{"items":[{"api_key":"k1","secret_id":"s1","name":"混元"}]}`,
			mustHas: []string{`"name":"混元"`},
			mustNot: []string{"k1", "s1"},
		},
		{
			name:    "非 JSON 只留占位",
			raw:     "--boundary\r\nContent-Disposition: form-data\r\n",
			mustHas: []string{"非JSON"},
		},
		{
			name: "超长截断",
			raw:  `{"v":"` + strings.Repeat("a", 2000) + `"}`,
		},
	}
	for _, c := range cases {
		got := MaskDetail([]byte(c.raw))
		for _, h := range c.mustHas {
			if !strings.Contains(got, h) {
				t.Fatalf("%s: 期望包含 %q，实际 %q", c.name, h, got)
			}
		}
		for _, n := range c.mustNot {
			if strings.Contains(got, n) {
				t.Fatalf("%s: 不应出现明文 %q，实际 %q", c.name, n, got)
			}
		}
	}
	if got := MaskDetail(nil); got != "" {
		t.Fatalf("空请求体应为空串，实际 %q", got)
	}
	maxRunes := len([]rune(MaskDetail([]byte(`{"v":"` + strings.Repeat("a", 2000) + `"}`))))
	if maxRunes > auditDetailLimit+1 {
		t.Fatalf("超长请求体未截断: %d", maxRunes)
	}
}

func TestHasPerm(t *testing.T) {
	super := mkRole("super", `["*"]`)
	ops := mkRole("ops", `["order:view","order:ship"]`)
	if !HasPerm(super, "config:edit") {
		t.Fatal("超管应拥有全部权限")
	}
	if !HasPerm(ops, "order:ship") || HasPerm(ops, "config:edit") {
		t.Fatal("运营角色权限判定错误")
	}
	if HasPerm(ops, "*") {
		t.Fatal("非超管不应命中通配")
	}
}
