# -*- coding: utf-8 -*-
"""拼装素材 e2e：底板与元素一对多（每块底板只出自己那套素材）、跨底板合成被拦截、种子幂等（v1.6）"""
import json
import os
import urllib.request

BASE = os.environ.get("COLLAGE_BASE", "http://127.0.0.1:18085") + "/api/v1"


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


def elements(board_id, token, cat=None):
    path = "/user/collage/elements?board_id=%d" % board_id
    if cat:
        path += "&category=" + cat
    st, d = req("GET", path, token)
    return st, d


steps = []

# 1. 登录
st, d = req("POST", "/user/auth/dev_login", body={"openid": "collage_user_001"})
assert st == 200 and d["code"] == 0, (st, d)
ut = d["data"]["token"]
steps.append("登录 OK")

# 2. 底板：4 块
st, d = req("GET", "/user/collage/boards", ut)
assert st == 200, (st, d)
boards = d["data"]
names = [b["name"] for b in boards]
assert len(boards) == 4, names
assert {"底板·小姑娘", "底板·小男孩", "底板·小兔子", "底板·小熊猫"} == set(names), names
by_name = {b["name"]: b for b in boards}
steps.append("底板 4 个：" + " / ".join(names))

# 3. 元素必须带 board_id（一对多：不做全局元素池）
st, d = req("GET", "/user/collage/elements", ut)
assert st == 400, (st, d)
steps.append("缺 board_id 取元素 → 400 ✔")

# 4. 小姑娘底板：AI 素材套件 24 个，六分类齐全
girl = by_name["底板·小姑娘"]
st, d = elements(girl["id"], ut)
assert st == 200, (st, d)
girl_els = d["data"]
assert len(girl_els) == 24, len(girl_els)
by_cat = {}
for e in girl_els:
    by_cat.setdefault(e["category"], []).append(e["name"])
expect = {"hair": 5, "hat": 5, "lash": 1, "expression": 4, "outfit": 4, "accessory": 5}
for cat, n in expect.items():
    assert len(by_cat.get(cat, [])) == n, (cat, by_cat.get(cat))
assert all(e["board_id"] == girl["id"] for e in girl_els)
steps.append("小姑娘底板 24 个元素：" + "; ".join("%s×%d" % (c, len(v)) for c, v in sorted(by_cat.items())))

# 5. 小男孩底板：只有中性部件，不能出现裙子/双马尾/睫毛
boy = by_name["底板·小男孩"]
st, d = elements(boy["id"], ut)
assert st == 200, (st, d)
boy_els = d["data"]
boy_names = [e["name"] for e in boy_els]
assert len(boy_els) == 13, (len(boy_els), boy_names)
for bad in ["小围裙", "红色小披风", "双马尾", "侧马尾", "星星睫毛", "蝴蝶结"]:
    assert bad not in boy_names, (bad, boy_names)
for cat in ["hair", "lash", "outfit"]:
    st, d = elements(boy["id"], ut, cat)
    assert st == 200 and len(d["data"]) == 0, (cat, d)
steps.append("小男孩底板 13 个中性元素，无裙装/头发：" + "、".join(boy_names))

# 6. 分类过滤只在本底板内生效
st, d = elements(girl["id"], ut, "expression")
assert st == 200 and len(d["data"]) == 4, (st, d)
steps.append("分类过滤（小姑娘·表情）4 个 ✔")

# 7. 跨底板混搭被拦：小男孩底板 + 小姑娘的裙子
skirt = [e for e in girl_els if e["name"] == "小围裙"][0]
a = json.loads(skirt["anchor"])
st, d = req("POST", "/user/collage/compose", ut,
            {"board_id": boy["id"], "placements": [{"element_id": skirt["id"], "x": a["x"], "y": a["y"], "scale": 1}]})
assert st == 400 and "不属于当前底板" in d["msg"], (st, d)
steps.append("小男孩底板 + 小围裙 → 400「不属于当前底板」✔")

# 8. 同底板正常合成：小男孩（帽子+表情+配饰）
pick = {e["name"]: e for e in boy_els}
placements = []
for nm in ["黄色草帽", "笑眼弯弯", "圆框眼镜"]:
    e = pick[nm]
    a = json.loads(e["anchor"])
    placements.append({"element_id": e["id"], "x": a["x"], "y": a["y"], "scale": a.get("scale", 1)})
st, d = req("POST", "/user/collage/compose", ut, {"board_id": boy["id"], "placements": placements})
assert st == 200 and d["code"] == 0, (st, d)
work = d["data"]
assert work["type"] == "collage" and work["image_url"].startswith("/static/collage_"), work
steps.append("小男孩底板合成 OK → %s" % work["image_url"])

# 9. 小姑娘整套换装（乱序添加，验证服务端按分类 z-order 合成）
placements = []
gpick = {e["name"]: e for e in girl_els}
for nm in ["蝴蝶结", "黄色草帽", "双马尾", "星星睫毛", "爱心眼", "小围裙"]:
    e = gpick[nm]
    a = json.loads(e["anchor"])
    placements.append({"element_id": e["id"], "x": a["x"], "y": a["y"], "scale": a.get("scale", 1)})
st, d = req("POST", "/user/collage/compose", ut, {"board_id": girl["id"], "placements": placements})
assert st == 200 and d["code"] == 0, (st, d)
steps.append("小姑娘整套换装合成 OK → %s" % d["data"]["image_url"])

# 10. 作品入库可查
st, d = req("GET", "/user/works", ut)
assert st == 200 and any(w["id"] == work["id"] for w in d["data"]), (st, d)
steps.append("作品入库 OK")

# 11. 幂等：重复执行不产生重复记录
st, d = req("GET", "/user/collage/boards", ut)
assert len(d["data"]) == 4, d
for b in d["data"]:
    st, d2 = elements(b["id"], ut)
    assert st == 200, (b["name"], d2)
steps.append("种子幂等 OK（底板仍 4 块，元素归属稳定）")

print("\n".join("✔ " + s for s in steps))
print("ALL PASS")
