# -*- coding: utf-8 -*-
"""M1 全链路实测：素材→拼装合成→AI生成(mock)→作品库→下单绑作品。临时端口 18080"""
import json
import struct
import urllib.request
import uuid

BASE = "http://127.0.0.1:18080"
API = BASE + "/api/v1"


def req(method, path, token=None, body=None):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    r = urllib.request.Request(API + path, method=method,
        data=json.dumps(body).encode("utf-8") if body else None, headers=headers)
    try:
        resp = urllib.request.urlopen(r)
    except urllib.error.HTTPError as e:
        resp = e
    return json.loads(resp.read().decode("utf-8"))


def upload(token, filename, data):
    boundary = uuid.uuid4().hex
    body = (("--%s\r\nContent-Disposition: form-data; name=\"file\"; filename=\"%s\"\r\n"
             "Content-Type: image/png\r\n\r\n") % (boundary, filename)).encode() + data + \
           ("\r\n--%s--\r\n" % boundary).encode()
    r = urllib.request.Request(API + "/user/upload", data=body, method="POST", headers={
        "Authorization": "Bearer " + token,
        "Content-Type": "multipart/form-data; boundary=" + boundary})
    try:
        resp = urllib.request.urlopen(r)
    except urllib.error.HTTPError as e:
        resp = e
    return json.loads(resp.read().decode("utf-8"))


ut = req("POST", "/user/auth/dev_login", body={"openid": "m1user"})["data"]["token"]
ADDR = req("POST", "/user/addresses", ut, {"name": "张三", "phone": "13800001111", "province": "广东省", "city": "深圳市", "district": "南山区", "detail": "科技园 1 号楼"})["data"]["id"]
at = req("POST", "/admin/auth/login", body={"username": "admin", "password": "admin123"})["data"]["token"]

# 1. 素材
boards = req("GET", "/user/collage/boards", ut)["data"]
els = req("GET", "/user/collage/elements", ut)["data"]
print("底板:", len(boards), "个 | 元素:", len(els), "个")
assert len(boards) >= 1 and len(els) >= 5, "演示素材缺失"
board = boards[0]
by_cat = {}
for e in els:
    by_cat.setdefault(e["category"], []).append(e)

# 2. 拼装合成：头发+帽子+睫毛+蝴蝶结（用各自锚点）
placements = []
for cat in ("hair", "hat", "lash", "accessory"):
    el = by_cat[cat][0]
    anchor = json.loads(el["anchor"])
    placements.append({"element_id": el["id"], "x": anchor["x"], "y": anchor["y"], "scale": anchor["scale"]})
r = req("POST", "/user/collage/compose", ut, {"board_id": board["id"], "placements": placements})
assert r["code"] == 0, "拼装合成失败: %s" % r
work = r["data"]
print("拼装作品: id=%s url=%s" % (work["id"], work["image_url"]))

# 3. 下载合成图验证是有效 PNG 且尺寸=底板 400x500
png = urllib.request.urlopen(BASE + work["image_url"]).read()
w, h = struct.unpack(">II", png[16:24])
print("合成图: %d 字节, %dx%d, PNG签名=%s" % (len(png), w, h, png[:8] == b"\x89PNG\r\n\x1a\n"))
assert png[:8] == b"\x89PNG\r\n\x1a\n" and (w, h) == (400, 500), "合成图不符合预期"

# 4. 上传照片 + AI 照片转风格（mock）
tiny_png = bytes.fromhex("89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c4890000000d4944415478da63f8ffff3f0300050201f4a92a610000000049454e44ae426082")
up = upload(ut, "photo.png", tiny_png)
assert up["code"] == 0, "上传失败: %s" % up
print("上传成功:", up["data"]["url"])

r = req("POST", "/user/ai/tasks", ut, {"type": "ai_photo", "photo_url": up["data"]["url"]})
assert r["code"] == 0, "AI照片转风格失败: %s" % r
print("AI照片转风格( mock ) → 作品 id=%s" % r["data"]["work"]["id"])

# 5. AI 文生图 + 配额（默认每日 3 次，已用 1 次 photo，再用 2 次，第 4 次应 429）
for i in range(2):
    r = req("POST", "/user/ai/tasks", ut, {"type": "ai_text", "prompt": "扎双马尾的可爱小姑娘"})
    assert r["code"] == 0, "AI文生图失败: %s" % r
r = req("POST", "/user/ai/tasks", ut, {"type": "ai_text", "prompt": "第4次应被拒"})
print("配额第4次返回:", r["code"], r["msg"])
assert r["code"] == 429, "配额应已用尽"

# 6. 作品库
works = req("GET", "/user/works", ut)["data"]
print("作品库:", len(works), "个作品")
assert len(works) == 4, "作品数量应为4"

# 7. 用拼装作品下单（绑定作品）
sku = req("GET", "/user/products", ut)["data"][0]["skus"][0]
r = req("POST", "/user/orders", ut, {"address_id": ADDR, "items": [{"sku_id": sku["id"], "work_id": work["id"], "qty": 1}]})
assert r["code"] == 0, "下单失败: %s" % r
oid = r["data"]["order"]["id"]
req("POST", "/admin/orders/%d/accept" % oid, at)
req("POST", "/admin/orders/%d/pack" % oid, at)
req("POST", "/admin/orders/%d/ready_ship" % oid, at)
req("POST", "/admin/orders/%d/ship" % oid, at, {"company": "中通快递", "tracking_no": "ZN8866"})
d = req("GET", "/user/orders/%d" % oid, ut)["data"]
assert d["items"][0]["work_id"] == work["id"], "订单未绑定作品"
print("下单绑作品成功: 订单=%s 状态=%s 绑定作品=%s" % (d["order"]["order_no"], d["order"]["status"], d["items"][0]["work_id"]))

# 8. 删除作品（软删）后列表应减少
r = req("DELETE", "/user/works/%d" % works[-1]["id"], ut)
assert r["code"] == 0
works2 = req("GET", "/user/works", ut)["data"]
print("删除后作品库:", len(works2), "个作品")
assert len(works2) == 3

print("\n=== M1 全链路实测全部通过 ===")
