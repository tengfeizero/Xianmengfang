# -*- coding: utf-8 -*-
"""购物车全链路：加购累加→勾选→结算一单多明细→金额对账→失效/越权拦截（v1.0）"""
import json
import urllib.request

BASE = "http://127.0.0.1:18084/api/v1"


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

# 1. 登录与商品
st, d = req("POST", "/user/auth/dev_login", body={"openid": "cart_user_001"})
assert st == 200 and d["code"] == 0, (st, d)
ut = d["data"]["token"]
st, d = req("GET", "/user/products", ut)
assert st == 200 and len(d["data"]) >= 2, (st, d)
sku_tee = d["data"][0]["skus"][0]   # 9900
sku_mug = d["data"][1]["skus"][0]   # 5900
steps.append("登录与商品列表 OK")

# 2. 加购累加：同 sku 同作品两次 = 一条 qty 2
st, d = req("POST", "/user/cart", ut, {"sku_id": sku_tee["id"], "qty": 1})
assert st == 200 and d["data"]["qty"] == 1 and d["data"]["checked"] == 1, (st, d)
st, d = req("POST", "/user/cart", ut, {"sku_id": sku_tee["id"], "qty": 1})
assert st == 200 and d["data"]["qty"] == 2, (st, d)
st, d = req("GET", "/user/cart", ut)
assert len(d["data"]["items"]) == 1, d["data"]
steps.append("加购累加 OK（同款两次 = 一条 qty 2，默认勾选）")

# 3. 作品绑定：同一 sku 配不同作品 = 不同项
st, d = req("GET", "/user/collage/boards", ut)
assert st == 200 and d["data"], (st, d)
st, d = req("GET", "/user/collage/elements?category=hair", ut)
els = d["data"]
assert els, (st, d)
st, d = req("POST", "/user/collage/compose", ut, {"board_id": req("GET", "/user/collage/boards", ut)[1]["data"][0]["id"],
                                                  "placements": [{"element_id": els[0]["id"], "x": 0.5, "y": 0.2, "scale": 1}]})
assert st == 200, (st, d)
work_id = d["data"]["id"]
st, d = req("POST", "/user/cart", ut, {"sku_id": sku_mug["id"], "work_id": work_id, "qty": 1})
assert st == 200 and d["data"]["work_id"] == work_id, (st, d)
st, d = req("POST", "/user/cart", ut, {"sku_id": sku_mug["id"], "qty": 1})
assert st == 200 and d["data"]["work_id"] == 0, (st, d)
st, d = req("GET", "/user/cart", ut)
items = d["data"]["items"]
assert len(items) == 3, d["data"]
mug_work = next(i for i in items if i["work_id"] == work_id)
assert mug_work["work_thumb"], "带作品的项缺缩略图"
steps.append("作品绑定 OK（同 sku 配不同作品 = 不同项，缩略图返回）")

# 4. 改数量与勾选；汇总只算勾选且有效项
tee = next(i for i in items if i["sku_id"] == sku_tee["id"])
st, d = req("PUT", "/user/cart/%d" % tee["id"], ut, {"qty": 3})
assert st == 200, (st, d)
mug_plain = next(i for i in items if i["sku_id"] == sku_mug["id"] and i["work_id"] == 0)
st, d = req("PUT", "/user/cart", ut, {"checked": 0, "item_ids": [mug_plain["id"]]})
assert st == 200, (st, d)
st, d = req("GET", "/user/cart", ut)
assert d["data"]["summary"] == {"checked_count": 4, "checked_cents": 29700 + 5900}, d["data"]["summary"]
steps.append("改数量/批量勾选 OK（汇总只算勾选项：3×9900+5900=35600）")

# 5. 结算：一单多明细 + 购物车项移除
st, d = req("POST", "/user/addresses", ut, {"name": "车测", "phone": "13800002222",
                                            "province": "广东省", "city": "深圳市", "district": "南山区", "detail": "购物车 1 号"})
assert st == 200, (st, d)
addr_id = d["data"]["id"]
st, d = req("POST", "/user/cart/checkout", ut, {"address_id": addr_id})
assert st == 200, (st, d)
order = d["data"]["order"]
assert d["data"]["total_cents"] == 35600, d["data"]
st, d = req("GET", "/user/orders/%d" % order["id"], ut)
assert st == 200 and len(d["data"]["items"]) == 2, d["data"]
assert d["data"]["amounts"]["items_cents"] == 35600
st, d = req("GET", "/user/cart", ut)
left = d["data"]["items"]
assert len(left) == 1 and left[0]["work_id"] == 0 and left[0]["sku_id"] == sku_mug["id"], left
steps.append("结算 OK（一单 2 明细 35600 分对账一致，已结算项从购物车移除）")

# 6. 边界：数量上限、加购校验、结算缺货（并发语义）、越权、清空
mug_cart = left[0]
st, d = req("PUT", "/user/cart/%d" % mug_cart["id"], ut, {"qty": 500})
assert st == 400 and "99" in d["msg"], (st, d)
st, d = req("PUT", "/user/cart/%d" % mug_cart["id"], ut, {"qty": 1})
assert st == 200

# 造缺货：库存 199。A 先买 99，B 再买 99，剩 1；两人各加购 1（都不锁库存），
# B 先结算成功后 A 结算应 409
st, d = req("POST", "/user/cart/checkout", ut, {"address_id": addr_id, "item_ids": [mug_cart["id"]]})
assert st == 200 and d["data"]["total_cents"] == 5900, (st, d)
st, d = req("POST", "/user/auth/dev_login", body={"openid": "cart_other_001"})
assert st == 200
ut2 = d["data"]["token"]
st, d = req("POST", "/user/cart", ut2, {"sku_id": sku_mug["id"], "qty": 99})
assert st == 200, (st, d)
st, d = req("POST", "/user/addresses", ut2, {"name": "车测二", "phone": "13800003333",
                                             "province": "广东省", "city": "深圳市", "district": "南山区", "detail": "购物车 2 号"})
assert st == 200
addr2 = d["data"]["id"]
st, d = req("POST", "/user/cart/checkout", ut2, {"address_id": addr2})
assert st == 200, (st, d)
# 再买 98，库存剩 1
st, d = req("POST", "/user/cart", ut2, {"sku_id": sku_mug["id"], "qty": 98})
assert st == 200, (st, d)
st, d = req("POST", "/user/cart/checkout", ut2, {"address_id": addr2})
assert st == 200, (st, d)
st, d = req("GET", "/user/products", ut)
mug_stock = d["data"][1]["skus"][0]["stock"]
assert mug_stock == 1, mug_stock
st, d = req("POST", "/user/cart", ut, {"sku_id": sku_mug["id"], "qty": 1})
assert st == 200, (st, d)
st, d = req("POST", "/user/cart", ut2, {"sku_id": sku_mug["id"], "qty": 1})
assert st == 200, (st, d)  # 加购不锁库存，仍成功
st, d = req("POST", "/user/cart/checkout", ut2, {"address_id": addr2})
assert st == 200, (st, d)  # B 先结算，库存归 0
st, d = req("POST", "/user/cart/checkout", ut, {"address_id": addr_id})
assert st == 409 and "库存" in d["msg"], (st, d)

# 越权与清空
st, d = req("GET", "/user/cart", ut)
other_item = d["data"]["items"][0]["id"]
st, d = req("PUT", "/user/cart/%d" % other_item, ut2, {"qty": 1})
assert st == 403, (st, d)
st, d = req("DELETE", "/user/cart/%d" % other_item, ut2)
assert st == 403, (st, d)
st, d = req("DELETE", "/user/cart", ut)
assert st == 200
st, d = req("GET", "/user/cart", ut)
assert len(d["data"]["items"]) == 0, d["data"]
steps.append("边界 OK（数量上限/结算缺货 409/越权 403/清空）")

print("购物车链路全部通过：")
for s in steps:
    print("  ✔ " + s)
