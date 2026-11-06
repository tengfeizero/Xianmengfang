# -*- coding: utf-8 -*-
"""模拟小程序端完整调用链：登录→商店→创作→作品→下单→订单（M1.5 联调验证）"""
import json
import urllib.request

BASE = "http://127.0.0.1:18081/api/v1"


def req(method, path, token=None, body=None):
    h = {"Content-Type": "application/json"}
    if token:
        h["Authorization"] = "Bearer " + token
    r = urllib.request.Request(BASE + path, method=method,
                               data=json.dumps(body).encode("utf-8") if body else None, headers=h)
    try:
        res = urllib.request.urlopen(r)
    except urllib.error.HTTPError as e:
        res = e
    data = json.loads(res.read().decode("utf-8"))
    return res.status, data


steps = []

# 1. 登录（对应 pages/login）
st, d = req("POST", "/auth/dev_login", body={"openid": "mp_user_001"})
assert st == 200 and d["code"] == 0, (st, d)
ut = d["data"]["token"]
steps.append("登录 OK")

# 2. 商店首页（pages/index）
st, d = req("GET", "/products", ut)
assert st == 200 and len(d["data"]) > 0
pid = d["data"][0]["id"]
sku = d["data"][0]["skus"][0]
steps.append("商品列表 OK（%d 个商品，%d 个 SKU）" % (len(d["data"]), len(d["data"][0]["skus"])))

# 3. 商品详情（pages/product）
st, d = req("GET", "/products/%d" % pid, ut)
assert st == 200 and d["data"]["product"]["id"] == pid
steps.append("商品详情 OK")

# 4. 创作工坊：底板+元素（pages/studio）
st, d = req("GET", "/collage/boards", ut)
assert st == 200 and d["data"], (st, d)
board = d["data"][0]
st, els = req("GET", "/collage/elements?category=hair", ut)
assert st == 200
steps.append("底板 %d 个 / 头发素材 %d 个 OK" % (1, len(els["data"])))

# 5. 合成作品（studio compose）
placements = [{"element_id": els["data"][0]["id"], "x": 0.5, "y": 0.2, "scale": 1}]
st, d = req("POST", "/collage/compose", ut, {"board_id": board["id"], "placements": placements})
assert st == 200 and d["data"]["id"], (st, d)
work_id = d["data"]["id"]
image_url = d["data"]["image_url"]
assert image_url.startswith("/static/"), image_url
steps.append("拼装合成 OK（作品 %d，图 %s）" % (work_id, image_url))

# 6. 作品库（pages/works）
st, d = req("GET", "/works", ut)
assert st == 200 and any(w["id"] == work_id for w in d["data"])
steps.append("作品库 OK（%d 件）" % len(d["data"]))

# 7. 下单绑定作品（pages/product submit）
st, d = req("POST", "/orders", ut, {"items": [{"sku_id": sku["id"], "work_id": work_id, "qty": 1}]})
assert st == 200, (st, d)
oid = d["data"]["order"]["id"]
total = d["data"]["total_cents"]
assert total == sku["price_cents"]
steps.append("下单绑作品 OK（订单 %d，金额 %d 分对账一致）" % (oid, total))

# 8. 订单列表/详情/确认收货（pages/orders、pages/order）
st, d = req("GET", "/orders", ut)
assert st == 200 and d["data"][0]["order"]["id"] == oid and d["data"][0]["items"][0]["work_id"] == work_id
st, d = req("POST", "/admin/auth/login", body={"username": "admin", "password": "admin123"})
at = d["data"]["token"]
req("POST", "/admin/orders/%d/accept" % oid, at)
req("POST", "/admin/orders/%d/ship" % oid, at, {"company": "顺丰速运", "tracking_no": "SF777"})
st, d = req("POST", "/orders/%d/confirm" % oid, ut)
assert st == 200
st, d = req("GET", "/orders/%d" % oid, ut)
assert d["data"]["order"]["status"] == "done" and d["data"]["shipment"]["tracking_no"] == "SF777"
steps.append("订单列表/详情/接单/发货/确认收货 OK（含中文物流名）")

print("小程序调用链全部通过：")
for s in steps:
    print("  ✔ " + s)
