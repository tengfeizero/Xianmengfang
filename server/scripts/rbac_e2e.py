# -*- coding: utf-8 -*-
"""RBAC + 审计日志实测：权限拦截、角色保护规则、日志留痕与请求体脱敏。"""
import json
import urllib.request

BASE = "http://127.0.0.1:8080/api/v1"


def req(method, path, token=None, body=None):
    h = {"Content-Type": "application/json"}
    if token:
        h["Authorization"] = "Bearer " + token
    r = urllib.request.Request(
        BASE + path, method=method,
        data=json.dumps(body).encode("utf-8") if body is not None else None,
        headers=h)
    try:
        res = urllib.request.urlopen(r)
    except urllib.error.HTTPError as e:
        res = e
    try:
        return res.status, json.loads(res.read().decode("utf-8"))
    except Exception:
        return res.status, {}


def ok(st, d, what):
    assert st == 200 and d.get("code") == 0, "%s 失败: %s %s" % (what, st, d)
    return d["data"]


# 1. 超管登录 + 权限自检
st, d = req("POST", "/admin/auth/login", body={"username": "admin", "password": "admin123"})
at = ok(st, d, "超管登录")["token"]
me = ok(*req("GET", "/admin/me", at), what="超管 me")
assert me["is_super"] and me["perms"] == ["*"], me
print("1 超管自检 OK:", me["username"], me["role"]["name"], me["perms"])

# 2. 新建「发货员」角色：订单查看 + 发货
st, d = req("POST", "/admin/roles", at, {"name": "发货员", "remark": "只看订单并发货", "perms": ["order:view", "order:ship"]})
role = ok(st, d, "新建角色")
print("2 新建角色 OK: id=%s %s" % (role["id"], role["name"]))

# 3. 新建账号并登录，验证菜单级权限
st, d = req("POST", "/admin/admins", at, {
    "username": "shipper", "password": "ship123456", "nickname": "仓库小林", "role_id": role["id"]})
sid = ok(st, d, "新建管理员")["id"]
st, d = req("POST", "/admin/auth/login", body={"username": "shipper", "password": "ship123456"})
stok = ok(st, d, "发货员登录")["token"]
sme = ok(*req("GET", "/admin/me", stok), what="发货员 me")
assert not sme["is_super"] and set(sme["perms"]) == {"order:view", "order:ship"}, sme
print("3 发货员权限 OK:", sme["perms"])

# 4. 允许的操作：订单列表可看
st, d = req("GET", "/admin/orders", stok)
ok(st, d, "发货员看订单")

# 4a. 禁用账号立即失效（不等 JWT 过期）
ok(*req("PUT", "/admin/admins/%d" % sid, at, {"nickname": "仓库小林", "role_id": role["id"], "status": 0}),
   what="禁用发货员")
st, _ = req("GET", "/admin/orders", stok)
assert st == 401, "禁用后应 401，实际 %s" % st
ok(*req("PUT", "/admin/admins/%d" % sid, at, {"nickname": "仓库小林", "role_id": role["id"], "status": 1}),
   what="恢复发货员")
st, d = req("GET", "/admin/orders", stok)
ok(st, d, "恢复后看订单")
print("4a 账号禁用即时生效 OK（旧 token 立刻 401，恢复后可用）")

# 4b. 改角色权限立即生效：收回 order:ship
ok(*req("PUT", "/admin/roles/%d" % role["id"], at,
        {"name": "发货员", "remark": "只看订单", "perms": ["order:view"], "status": 1}), what="收回发货权限")
st, _ = req("POST", "/admin/orders/1/ship", stok, {"company": "顺丰", "tracking_no": "X1"})
assert st == 403, "收回权限后发货应 403，实际 %s" % st
ok(*req("PUT", "/admin/roles/%d" % role["id"], at,
        {"name": "发货员", "remark": "只看订单并发货", "perms": ["order:view", "order:ship"], "status": 1}),
   what="恢复发货权限")
print("4b 角色权限变更即时生效 OK（收回后发货 403）")

# 5. 越权：看配置 / 接单 / 管账号，全部 403 且留痕
for method, path, body, what in [
    ("GET", "/admin/configs", None, "看系统配置"),
    ("POST", "/admin/orders/1/accept", None, "接单"),
    ("GET", "/admin/admins", None, "看管理员列表"),
]:
    st, d = req(method, path, stok, body)
    assert st == 403, "%s 应 403，实际 %s %s" % (what, st, d)
print("5 越权拦截 OK: 配置/接单/账号管理均 403")

# 6. 保护规则：内置角色不可删；在用角色不可删；不能禁用自己
st, _ = req("DELETE", "/admin/roles/1", at)
assert st == 409, st
st, _ = req("DELETE", "/admin/roles/%d" % role["id"], at)
assert st == 409, st
state, _ = req("PUT", "/admin/admins/1", at, {"nickname": "系统管理员", "role_id": 1, "status": 0})
assert state == 409, state
print("6 保护规则 OK: 内置角色/在用角色/自我禁用均 409")

# 7. 审计日志：登录、角色创建、越权都在，且密码已脱敏
def logs(**kw):
    q = "&".join("%s=%s" % (k, v) for k, v in kw.items())
    return ok(*req("GET", "/admin/audit-logs?" + q, at), what="日志查询")

page = logs(size=100)
actions = {x["action"] for x in page["list"]}
for need in ["auth.login", "role.create", "admin.create", "deny"]:
    assert need in actions, "缺少动作 %s，实际 %s" % (need, actions)
denies = [x for x in page["list"] if x["action"] == "deny"]
assert any(x["admin_name"] == "shipper" for x in denies), denies
create = [x for x in page["list"] if x["action"] == "admin.create"][0]
assert "ship123456" not in create["detail"] and "***" in create["detail"], create["detail"]
print("7 审计日志 OK: 动作=%s" % sorted(actions))
print("   脱敏示例:", create["detail"])
print("   越权示例:", denies[0]["detail"], "|", denies[0]["target_id"])

# 8. 日志筛选
only_deny = logs(action="deny", size=50)
assert all(x["action"].startswith("deny") for x in only_deny["list"])
print("8 日志筛选 OK: deny 共 %d 条" % only_deny["total"])

# 9. 清理测试账号（保留角色，便于后台演示；账号删除也要留痕）
st, _ = req("DELETE", "/admin/admins/%d" % sid, at)
assert st == 200, st
print("9 清理测试账号 OK")

print("\n=== RBAC + 审计日志实测全部通过 ===")
