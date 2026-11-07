# -*- coding: utf-8 -*-
"""作品社区与积分 e2e（v1.7）

覆盖：发布/取消发布 → 广场可见性与提示词公开、点赞（幂等/自赞拦截）、
「自己的或已点赞的公开作品」才能下单、订单完成才累计使用次数并给作者积分（按件数）、
免费用完后 1 积分换 1 张、积分不足 429。
"""
import json
import os
import urllib.error
import urllib.request

BASE = os.environ.get("WORKS_BASE", "http://127.0.0.1:18086") + "/api/v1"
ADMIN_USER = os.environ.get("XMF_ADMIN_USER", "admin")
ADMIN_PASS = os.environ.get("XMF_ADMIN_PASS", "admin123")


def req(method, path, token=None, body=None):
    h = {"Content-Type": "application/json"}
    if token:
        h["Authorization"] = "Bearer " + token
    r = urllib.request.Request(
        BASE + path,
        method=method,
        data=json.dumps(body).encode("utf-8") if body is not None else None,
        headers=h,
    )
    try:
        res = urllib.request.urlopen(r)
    except urllib.error.HTTPError as e:
        res = e
    return res.status, json.loads(res.read().decode("utf-8"))


def login(openid):
    st, d = req("POST", "/user/auth/dev_login", body={"openid": openid})
    assert st == 200 and d["code"] == 0, (st, d)
    return d["data"]["token"]


steps = []

# ---------- 0. 三个用户 + 管理员 ----------
author = login("works_author_001")
buyer = login("works_buyer_001")
third = login("works_third_001")
st, d = req("POST", "/admin/auth/login", body={"username": ADMIN_USER, "password": ADMIN_PASS})
assert st == 200 and d["code"] == 0, (st, d)
admin = d["data"]["token"]
steps.append("登录：作者 / 买家 / 路人 / 管理员 OK")

# ---------- 1. 作者生成作品 ----------
st, d = req("POST", "/user/ai/tasks", author, {"type": "ai_text", "prompt": "粉色背景的双马尾小姑娘"})
assert st == 200, (st, d)
work = d["data"]["work"]
wid = work["id"]
assert work["prompt"] == "粉色背景的双马尾小姑娘", work
assert work["is_public"] == 0, work
steps.append("AI 生成作品 #%d，默认私密，prompt 已落库" % wid)

# ---------- 2. 私密作品不进广场 ----------
st, d = req("GET", "/user/plaza?sort=latest", buyer)
assert st == 200, (st, d)
ids = [i["id"] for i in d["data"]["items"]]
assert wid not in ids, ids
steps.append("私密作品不进广场 ✔")

# ---------- 3. 发布 → 广场可见 + 提示词公开 ----------
st, d = req("PUT", "/user/works/%d/publish" % wid, author, {"is_public": 1})
assert st == 200 and d["data"]["is_public"] == 1, (st, d)
st, d = req("GET", "/user/plaza?sort=latest", buyer)
item = next(i for i in d["data"]["items"] if i["id"] == wid)
assert item["prompt"] == "粉色背景的双马尾小姑娘", item
assert item["is_mine"] is False and item["liked"] is False, item
assert item["author"]["nickname"], item
steps.append("发布后广场可见，提示词一并公开，作者信息完整 ✔")

# ---------- 4. 点赞规则 ----------
st, d = req("POST", "/user/works/%d/like" % wid, author)
assert st == 400 and "自己" in d["msg"], (st, d)
steps.append("给自己的作品点赞 → 400 ✔")

st, d = req("POST", "/user/works/%d/like" % wid, buyer)
assert st == 200 and d["data"]["like_count"] == 1, (st, d)
st, d = req("POST", "/user/works/%d/like" % wid, buyer)
assert st == 200 and d["data"]["like_count"] == 1, (st, d)
steps.append("点赞幂等：重复点赞计数仍为 1 ✔")

st, d = req("GET", "/user/plaza?sort=likes", buyer)
item = next(i for i in d["data"]["items"] if i["id"] == wid)
assert item["liked"] is True and item["like_count"] == 1, item
steps.append("点赞榜 + liked 字段正确 ✔")

# ---------- 5. 下单权限：未点赞的路人不行 ----------
st, d = req("GET", "/user/products", third)
sku = d["data"][0]["skus"][0]
st, d = req("POST", "/user/cart", third, {"sku_id": sku["id"], "qty": 1, "work_id": wid})
assert st == 400, (st, d)
steps.append("未点赞的他人作品不能加购 → 400 ✔")

# ---------- 6. 买家下单（已点赞） ----------
st, d = req("POST", "/user/addresses", buyer, {
    "name": "买家", "phone": "13800002222", "province": "广东省", "city": "深圳市",
    "district": "南山区", "detail": "科技园 1 号", "is_default": True,
})
assert st == 200, (st, d)
addr_id = d["data"]["id"]
st, d = req("POST", "/user/orders", buyer, {
    "address_id": addr_id, "items": [{"sku_id": sku["id"], "work_id": wid, "qty": 2}],
})
assert st == 200, (st, d)
oid = d["data"]["order"]["id"]
steps.append("买家点赞后用该作品下单（2 件）→ 订单 #%d" % oid)

# 未完成前不结算
st, d = req("GET", "/user/works/%d" % wid, author)
assert d["data"]["use_count"] == 0, d
st, d = req("GET", "/user/points", author)
assert d["data"]["points"] == 0, d
steps.append("订单未完成：使用次数/积分都不动 ✔")

# ---------- 7. 管理员推进到已发货，买家确认收货 ----------
for act, body in (("accept", None), ("pack", None), ("ready_ship", None),
                  ("ship", {"company": "顺丰速运", "tracking_no": "SF20261107"})):
    st, d = req("POST", "/admin/orders/%d/%s" % (oid, act), admin, body)
    assert st == 200, (act, st, d)
st, d = req("POST", "/user/orders/%d/confirm" % oid, buyer)
assert st == 200, (st, d)
steps.append("订单走完 接单→生产→打包→发货→确认收货 ✔")

# ---------- 8. 完成后按件数结算 ----------
st, d = req("GET", "/user/works/%d" % wid, author)
assert d["data"]["use_count"] == 2, d
st, d = req("GET", "/user/points", author)
assert d["data"]["points"] == 2, d
assert d["data"]["logs"][0]["reason"] == "work_used", d
steps.append("完成结算：使用次数 +2、作者积分 +2（按件数）✔")

# 重复确认不应重复结算
st, d = req("POST", "/user/orders/%d/confirm" % oid, buyer)
assert st != 200, (st, d)
st, d = req("GET", "/user/points", author)
assert d["data"]["points"] == 2, d
steps.append("重复确认收货被状态机拦截，不重复结算 ✔")

# ---------- 9. 使用量榜 ----------
st, d = req("GET", "/user/plaza?sort=uses", buyer)
item = next(i for i in d["data"]["items"] if i["id"] == wid)
assert item["use_count"] == 2, item
steps.append("使用量榜读到 use_count=2 ✔")

# ---------- 10. 取消发布：广场消失，计数保留 ----------
st, d = req("PUT", "/user/works/%d/publish" % wid, author, {"is_public": 0})
assert st == 200, (st, d)
st, d = req("GET", "/user/plaza?sort=latest", buyer)
assert wid not in [i["id"] for i in d["data"]["items"]], d
st, d = req("GET", "/user/works/%d" % wid, author)
assert d["data"]["like_count"] == 1 and d["data"]["use_count"] == 2, d
steps.append("取消发布后广场消失，点赞/使用计数保留 ✔")

# ---------- 11. 免费用完 → 1 积分换 1 张 ----------
st, d = req("GET", "/user/ai/quota", author)
q = d["data"]
free_left, points = q["free_left"], q["points"]
assert points == 2 and free_left >= 0, q
steps.append("AI 配额：免费剩 %d 张、积分 %d" % (free_left, points))

for i in range(free_left):
    st, d = req("POST", "/user/ai/tasks", author, {"type": "ai_text", "prompt": "免费额度 %d" % i})
    assert st == 200, (st, d)
st, d = req("GET", "/user/ai/quota", author)
assert d["data"]["free_left"] == 0 and d["data"]["points"] == 2, d
steps.append("免费额度用尽，积分不动 ✔")

st, d = req("POST", "/user/ai/tasks", author, {"type": "ai_text", "prompt": "积分换图 1"})
assert st == 200, (st, d)
assert d["data"]["task"]["pay_type"] == "point", d
st, d = req("GET", "/user/ai/quota", author)
assert d["data"]["points"] == 1 and d["data"]["free_left"] == 0, d
assert d["data"]["free_used"] == q["free_limit"], d
steps.append("免费用完后生成 → 扣 1 积分，且不占免费额度 ✔")

# 用掉最后 1 积分后再生成 → 429
st, d = req("POST", "/user/ai/tasks", author, {"type": "ai_text", "prompt": "积分换图 2"})
assert st == 200, (st, d)
st, d = req("POST", "/user/ai/tasks", author, {"type": "ai_text", "prompt": "积分换图 3"})
assert st == 429 and "积分不足" in d["msg"], (st, d)
steps.append("积分用尽 → 429 且提示积分不足 ✔")

st, d = req("GET", "/user/points", author)
spend = [l for l in d["data"]["logs"] if l["reason"] == "ai_spend"]
assert len(spend) == 2 and spend[0]["balance"] == 0, spend
steps.append("积分流水：2 条 ai_spend，余额归零 ✔")

print("\n".join("· " + s for s in steps))
print("\n作品社区与积分 e2e 全部通过")
