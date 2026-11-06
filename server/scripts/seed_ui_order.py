# -*- coding: utf-8 -*-
"""为后台 UI 实测准备数据：创建一笔待接单订单"""
import json
import urllib.request

def req(method, path, token=None, body=None):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    r = urllib.request.Request(
        "http://127.0.0.1:18080/api/v1" + path, method=method,
        data=json.dumps(body).encode("utf-8") if body else None, headers=headers)
    return json.loads(urllib.request.urlopen(r).read())

ut = req("POST", "/auth/dev_login", body={"openid": "uibuyer"})["data"]["token"]
ADDR = req("POST", "/addresses", ut, {"name": "张三", "phone": "13800001111", "province": "广东省", "city": "深圳市", "district": "南山区", "detail": "科技园 1 号楼"})["data"]["id"]
r = req("POST", "/orders", ut, body={"address_id": ADDR, "items": [{"sku_id": 1, "qty": 1}]})
print("测试订单已创建:", r["data"]["order"]["order_no"], r["data"]["order"]["status"])
