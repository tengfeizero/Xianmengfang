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
st, d = req("POST", "/user/auth/dev_login", body={"openid": "mp_user_001"})
assert st == 200 and d["code"] == 0, (st, d)
ut = d["data"]["token"]
steps.append("登录 OK")

# 2. 商店首页（pages/index）
st, d = req("GET", "/user/products", ut)
assert st == 200 and len(d["data"]) > 0
pid = d["data"][0]["id"]
sku = d["data"][0]["skus"][0]
steps.append("商品列表 OK（%d 个商品，%d 个 SKU）" % (len(d["data"]), len(d["data"][0]["skus"])))

# 3. 商品详情（pages/product）
st, d = req("GET", "/user/products/%d" % pid, ut)
assert st == 200 and d["data"]["product"]["id"] == pid
steps.append("商品详情 OK")

# 4. 创作工坊：底板+元素（pages/studio）
st, d = req("GET", "/user/collage/boards", ut)
assert st == 200 and d["data"], (st, d)
board = d["data"][0]
st, els = req("GET", "/user/collage/elements?category=hair", ut)
assert st == 200
steps.append("底板 %d 个 / 头发素材 %d 个 OK" % (1, len(els["data"])))

# 5. 合成作品（studio compose）
placements = [{"element_id": els["data"][0]["id"], "x": 0.5, "y": 0.2, "scale": 1}]
st, d = req("POST", "/user/collage/compose", ut, {"board_id": board["id"], "placements": placements})
assert st == 200 and d["data"]["id"], (st, d)
work_id = d["data"]["id"]
image_url = d["data"]["image_url"]
assert image_url.startswith("/static/"), image_url
steps.append("拼装合成 OK（作品 %d，图 %s）" % (work_id, image_url))

# 6. 作品库（pages/works）
st, d = req("GET", "/user/works", ut)
assert st == 200 and any(w["id"] == work_id for w in d["data"])
steps.append("作品库 OK（%d 件）" % len(d["data"]))

# 6.5 地址管理（pages/address/list、pages/address/edit）
addr_body = {"name": "李雷", "phone": "13800138000", "province": "广东省",
             "city": "深圳市", "district": "南山区", "detail": "科技园 8 栋 501", "is_default": True}
st, d = req("POST", "/user/addresses", ut, addr_body)
assert st == 200 and d["data"]["is_default"] == 1, (st, d)
ADDR = d["data"]["id"]
st, d = req("POST", "/user/addresses", ut, dict(addr_body, name="韩梅梅", is_default=False))
assert st == 200 and d["data"]["is_default"] == 0
ADDR2 = d["data"]["id"]
st, d = req("POST", "/user/addresses/%d/default" % ADDR2, ut)
assert st == 200
st, d = req("GET", "/user/addresses", ut)
assert len(d["data"]) == 2 and sum(1 for a in d["data"] if a["is_default"] == 1) == 1
st, d = req("PUT", "/user/addresses/%d" % ADDR, ut, dict(addr_body, name="李雷改"))
assert st == 200
st, d = req("DELETE", "/user/addresses/%d" % ADDR2, ut)  # 删默认 → 剩余一条自动补默认
assert st == 200
st, d = req("GET", "/user/addresses", ut)
assert len(d["data"]) == 1 and d["data"][0]["id"] == ADDR and d["data"][0]["is_default"] == 1
steps.append("地址管理 OK（增/设默认/唯一默认/编辑/删默认自动补位）")

# 7. 下单绑定作品+地址（pages/product submit）
st, d = req("POST", "/user/orders", ut, {"address_id": ADDR, "items": [{"sku_id": sku["id"], "work_id": work_id, "qty": 1}]})
assert st == 200, (st, d)
oid = d["data"]["order"]["id"]
total = d["data"]["total_cents"]
assert total == sku["price_cents"]
o = d["data"]["order"]
assert o["receiver_name"] == "李雷改" and o["receiver_phone"] == "13800138000"
assert o["receiver_address"] == "广东省深圳市南山区科技园 8 栋 501"
steps.append("下单绑作品+地址 OK（订单 %d，金额 %d 分对账一致，收货快照正确）" % (oid, total))

# 7.5 无地址/他人地址下单应拒绝
st, d = req("POST", "/user/orders", ut, {"items": [{"sku_id": sku["id"], "qty": 1}]})
assert st == 400, (st, d)
st, d = req("POST", "/user/orders", ut, {"address_id": 99999, "items": [{"sku_id": sku["id"], "qty": 1}]})
assert st == 400, (st, d)
steps.append("无地址/非法地址下单被拒 OK")

# 8. 订单列表/详情/确认收货（pages/orders、pages/order）
st, d = req("GET", "/user/orders", ut)
assert st == 200 and d["data"][0]["order"]["id"] == oid and d["data"][0]["items"][0]["work_id"] == work_id
st, d = req("POST", "/admin/auth/login", body={"username": "admin", "password": "admin123"})
at = d["data"]["token"]
req("POST", "/admin/orders/%d/accept" % oid, at)
req("POST", "/admin/orders/%d/pack" % oid, at)
req("POST", "/admin/orders/%d/ready_ship" % oid, at)
req("POST", "/admin/orders/%d/ship" % oid, at, {"company": "顺丰速运", "tracking_no": "SF777"})
st, d = req("POST", "/user/orders/%d/confirm" % oid, ut)
assert st == 200
st, d = req("GET", "/user/orders/%d" % oid, ut)
assert d["data"]["order"]["status"] == "done" and d["data"]["shipment"]["tracking_no"] == "SF777"
detail = d["data"]
assert detail["amounts"]["total_cents"] == total == detail["amounts"]["items_cents"]
assert len(detail["events"]) == 6 and detail["events"][-1]["to"] == "done", detail["events"]
assert detail["items"][0]["work_image_url"], "订单详情缺少作品图"
steps.append("订单列表/详情/接单/生产/打包/发货/确认收货 OK（金额明细、状态流水 6 条、作品图，含中文物流名）")

print("小程序调用链全部通过：")
for s in steps:
    print("  ✔ " + s)
