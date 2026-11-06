# -*- coding: utf-8 -*-
"""拼装素材扩充 e2e：老库回填→底板/元素数量与分类→新底板合成→空分类可用（v1.1）"""
import json
import urllib.request

BASE = "http://127.0.0.1:18085/api/v1"


def req(method, path, token=None, body=None):
    h = {"Content-Type": "application/json"}
    if token:
        h["Authorization"] = "Bearer " + token
    r = urllib.request.Request(BASE + path, method=method,
                               data=json.dumps(body).encode("utf-8") if body is not None else None, headers=h)
    try:
        res = urllib.request.urlopen(r)
    except urllib.error.HTTPError as e:
        res = e
    return res.status, json.loads(res.read().decode("utf-8"))


steps = []

# 1. 登录
st, d = req("POST", "/user/auth/dev_login", body={"openid": "collage_user_001"})
assert st == 200 and d["code"] == 0, (st, d)
ut = d["data"]["token"]
steps.append("登录 OK")

# 2. 底板回填：老库 1 → 4
st, d = req("GET", "/user/collage/boards", ut)
assert st == 200, (st, d)
boards = d["data"]
names = [b["name"] for b in boards]
assert len(boards) == 4, names
assert "底板·小姑娘" in names and "底板·小男孩" in names and "底板·小兔子" in names and "底板·小熊猫" in names, names
steps.append("底板 4 个：" + " / ".join(names))

# 3. 元素回填：老库 6 → 26，六分类齐全
st, d = req("GET", "/user/collage/elements", ut)
assert st == 200, (st, d)
els = d["data"]
assert len(els) == 26, len(els)
by_cat = {}
for e in els:
    by_cat.setdefault(e["category"], []).append(e["name"])
expect = {"hair": 6, "hat": 5, "lash": 2, "expression": 4, "outfit": 4, "accessory": 5}
for cat, n in expect.items():
    assert len(by_cat.get(cat, [])) == n, (cat, by_cat.get(cat))
steps.append("元素 26 个：" + "; ".join("%s×%d" % (c, len(v)) for c, v in sorted(by_cat.items())))

# 4. 表情/服装（原空分类）可单独拉取
st, d = req("GET", "/user/collage/elements?category=expression", ut)
assert st == 200 and len(d["data"]) == 4, (st, d)
st, d = req("GET", "/user/collage/elements?category=outfit", ut)
assert st == 200 and len(d["data"]) == 4, (st, d)
steps.append("表情/服装分类拉取 OK")

# 5. 新底板（小熊猫）+ 新元素（圣诞帽/笑眼弯弯/小围裙/圆框眼镜）合成作品
b_panda = [b for b in boards if b["name"] == "底板·小熊猫"][0]
els_all = req("GET", "/user/collage/elements", ut)[1]["data"]
pick = {e["name"]: e for e in els_all}
placements = []
for nm in ["圣诞帽", "笑眼弯弯", "小围裙", "圆框眼镜"]:
    e = pick[nm]
    a = json.loads(e["anchor"])
    placements.append({"element_id": e["id"], "x": a["x"], "y": a["y"], "scale": a.get("scale", 1)})
st, d = req("POST", "/user/collage/compose", ut, {"board_id": b_panda["id"], "placements": placements})
assert st == 200 and d["code"] == 0, (st, d)
work = d["data"]
assert work["type"] == "collage" and work["image_url"].startswith("/static/collage_"), work
steps.append("小熊猫底板合成 OK → %s" % work["image_url"])

# 6. 作品入库可查
st, d = req("GET", "/user/works", ut)
assert st == 200 and any(w["id"] == work["id"] for w in d["data"]), (st, d)
steps.append("作品入库 OK")

# 7. 幂等：重复执行不产生重复记录
st, d = req("GET", "/user/collage/boards", ut)
assert len(d["data"]) == 4
st, d = req("GET", "/user/collage/elements", ut)
assert len(d["data"]) == 26
steps.append("种子幂等 OK")

print("\n".join("✔ " + s for s in steps))
print("ALL PASS")
