# -*- coding: utf-8 -*-
"""M0 全链路实测脚本：临时端口 18080"""
import json
import urllib.request

BASE = "http://127.0.0.1:18080/api/v1"


def req(method, path, token=None, body=None):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    r = urllib.request.Request(
        BASE + path, method=method,
        data=json.dumps(body).encode("utf-8") if body else None,
        headers=headers)
    try:
        resp = urllib.request.urlopen(r)
    except urllib.error.HTTPError as e:  # 业务错误（400/401/403/409）也返回 JSON
        resp = e
    return json.loads(resp.read().decode("utf-8"))


ut = req("POST", "/user/auth/dev_login", body={"openid": "testuser001"})["data"]["token"]
ADDR = req("POST", "/user/addresses", ut, {"name": "张三", "phone": "13800001111", "province": "广东省", "city": "深圳市", "district": "南山区", "detail": "科技园 1 号楼"})["data"]["id"]
at = req("POST", "/admin/auth/login", body={"username": "admin", "password": "admin123"})["data"]["token"]

# 权限校验：用户 token 不能访问后台接口
r = req("GET", "/admin/orders", ut)
print("权限校验:", "通过(403)" if r["code"] == 403 else "!! 漏洞: " + str(r))

# 未登录不能下单
r = req("POST", "/user/orders", body={"address_id": ADDR, "items": [{"sku_id": 1, "qty": 1}]})
print("鉴权校验:", "通过(401)" if r["code"] == 401 else "!! 漏洞: " + str(r))

# 库存不足
r = req("POST", "/user/orders", ut, {"address_id": ADDR, "items": [{"sku_id": 1, "qty": 999}]})
print("库存不足返回:", r["code"], r["msg"])

# 正常下单 + UTF-8 中文物流名
o = req("POST", "/user/orders", ut, {"address_id": ADDR, "items": [{"sku_id": 2, "qty": 1}]})["data"]["order"]
oid = o["id"]
print("下单成功:", o["order_no"], "金额(分):", o["total_cents"], "状态:", o["status"])

# 重复接单 / 状态机非法流转
req("POST", "/admin/orders/%d/accept" % oid, at)
r = req("POST", "/admin/orders/%d/accept" % oid, at)
print("重复接单返回:", r["code"], r["msg"])

req("POST", "/admin/orders/%d/ship" % oid, at, {"company": "顺丰速运", "tracking_no": "SF8866"})
d = req("GET", "/user/orders/%d" % oid, ut)["data"]
print("发货后状态:", d["order"]["status"], "| 物流:", d["shipment"]["company"], d["shipment"]["tracking_no"])

req("POST", "/user/orders/%d/confirm" % oid, ut)
d = req("GET", "/user/orders/%d" % oid, ut)["data"]
print("确认收货后状态:", d["order"]["status"])

# 后台订单列表按状态筛选
done = req("GET", "/admin/orders?status=done", at)["data"]
print("后台已完成订单数:", len(done))
