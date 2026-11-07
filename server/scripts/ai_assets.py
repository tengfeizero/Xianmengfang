# -*- coding: utf-8 -*-
"""AI 素材处理流水线（配合 ImageGen 使用）：
1) key_white: 白底图 -> 透明背景 PNG（从四边泛洪填充，只抠与边连通的白，
   不会误伤角色内部的白色高光），裁到内容包围盒，按目标宽高比补边后缩放输出。
2) extract_diff: 底板图与"穿上道具"变体图做像素差分，抠出道部件 PNG，
   并输出道具中心在底板上的相对坐标（即拼装锚点 x/y）。
用法示例见文件底部 __main__。
"""
import os
import sys
from collections import deque

from PIL import Image

WHITE_T = 236  # 近白阈值


def _flood_key_white(im):
    """从边界泛洪把近白背景抠成透明，返回 RGBA"""
    im = im.convert("RGB")
    w, h = im.size
    px = im.load()
    alpha = [[255] * w for _ in range(h)]
    q = deque()

    def near_white(x, y):
        r, g, b = px[x, y]
        return r >= WHITE_T and g >= WHITE_T and b >= WHITE_T

    for x in range(w):
        for y in (0, h - 1):
            if near_white(x, y) and alpha[y][x]:
                alpha[y][x] = 0
                q.append((x, y))
    for y in range(h):
        for x in (0, w - 1):
            if near_white(x, y) and alpha[y][x]:
                alpha[y][x] = 0
                q.append((x, y))
    while q:
        x, y = q.popleft()
        for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
            nx, ny = x + dx, y + dy
            if 0 <= nx < w and 0 <= ny < h and alpha[ny][nx] and near_white(nx, ny):
                alpha[ny][nx] = 0
                q.append((nx, ny))

    out = Image.new("RGBA", (w, h))
    op = out.load()
    for y in range(h):
        for x in range(w):
            r, g, b = px[x, y]
            op[x, y] = (r, g, b, alpha[y][x])
    return out


def _fit_canvas(im, tw, th):
    """内容缩放到 tw x th 内（保持比例），居中放在透明画布上"""
    im = im.crop(im.getbbox())
    scale = min(tw / im.width, th / im.height)
    nw, nh = max(1, round(im.width * scale)), max(1, round(im.height * scale))
    im = im.resize((nw, nh), Image.LANCZOS)
    canvas = Image.new("RGBA", (tw, th), (0, 0, 0, 0))
    canvas.paste(im, ((tw - nw) // 2, (th - nh) // 2), im)
    return canvas


def key_white(src, dst, tw=400, th=500, erase_regions=()):
    """白底转透明并输出固定尺寸素材；erase_regions=(l,t,r,b,...) 先抹白再抠"""
    im = Image.open(src).convert("RGB")
    for (l, t, r, b) in erase_regions:
        im.paste((255, 255, 255), (l, t, r, b))
    out = _fit_canvas(_flood_key_white(im), tw, th)
    os.makedirs(os.path.dirname(dst), exist_ok=True)
    out.save(dst)
    return dst


def _morph(mask, erode=True):
    """4 邻域腐蚀/膨胀。mask: set of (x,y)"""
    out = set()
    if erode:
        for (x, y) in mask:
            if (x + 1, y) in mask and (x - 1, y) in mask and (x, y + 1) in mask and (x, y - 1) in mask:
                out.add((x, y))
    else:
        for (x, y) in mask:
            for dx, dy in ((0, 0), (1, 0), (-1, 0), (0, 1), (0, -1)):
                out.add((x + dx, y + dy))
    return out


def _clean_diff(diff, keep_ratio=0.15, keep_min=300):
    """开运算去毛刺 + 连通域过滤去重绘噪点"""
    w, h = diff.size
    dp = diff.load()
    mask = {(x, y) for y in range(h) for x in range(w) if dp[x, y][3] != 0}
    mask = _morph(_morph(mask, erode=True), erode=False)  # 开运算：腐蚀+膨胀

    # 连通域
    seen = set()
    comps = []
    for p in mask:
        if p in seen:
            continue
        q = deque([p])
        seen.add(p)
        comp = []
        while q:
            x, y = q.popleft()
            comp.append((x, y))
            for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                n = (x + dx, y + dy)
                if n in mask and n not in seen:
                    seen.add(n)
                    q.append(n)
        comps.append(comp)
    if not comps:
        diff.putalpha(Image.new("L", (w, h), 0))
        return
    mx = max(len(c) for c in comps)
    keep = set()
    for c in comps:
        if len(c) >= max(keep_ratio * mx, keep_min):
            keep.update(c)
    for y in range(h):
        for x in range(w):
            if (x, y) not in keep:
                dp[x, y] = (0, 0, 0, 0)


def extract_diff(base_src, variant_src, dst, tw=400, th=500, erase_regions=(), diff_t=60, keep_min=300):
    """差分抠部件：返回 (dst, anchor_x, anchor_y)（相对 400x500 画布）"""
    base = _flood_key_white(_open_erased(base_src, erase_regions))
    vari = _flood_key_white(_open_erased(variant_src, erase_regions))
    if base.size != vari.size:
        vari = vari.resize(base.size, Image.LANCZOS)
    bp, vp = base.load(), vari.load()
    w, h = base.size
    diff = Image.new("RGBA", (w, h), (0, 0, 0, 0))
    dp = diff.load()
    for y in range(h):
        for x in range(w):
            br, bg, bb, ba = bp[x, y]
            vr, vg, vb, va = vp[x, y]
            # 变体不透明且与底板（或其透明处）颜色差明显 → 属于新部件
            if va > 128:
                if ba < 128:
                    dp[x, y] = (vr, vg, vb, 255)
                else:
                    d = abs(br - vr) + abs(bg - vg) + abs(bb - vb)
                    if d > diff_t:
                        dp[x, y] = (vr, vg, vb, 255)
    _clean_diff(diff, keep_min=keep_min)
    bbox = diff.getbbox()
    if not bbox:
        raise RuntimeError("差分为空: " + dst)
    cx = (bbox[0] + bbox[2]) / 2 / w
    cy = (bbox[1] + bbox[3]) / 2 / h
    item = diff.crop(bbox)
    # 统一放进 tw x th 画布（不缩放超过 1 倍，保持清晰）
    scale = min(1.0, tw / item.width, th / item.height)
    if scale < 1.0:
        item = item.resize((round(item.width * scale), round(item.height * scale)), Image.LANCZOS)
    canvas = Image.new("RGBA", (tw, th), (0, 0, 0, 0))
    canvas.paste(item, ((tw - item.width) // 2, (th - item.height) // 2), item)
    os.makedirs(os.path.dirname(dst), exist_ok=True)
    canvas.save(dst)
    return dst, round(cx, 3), round(cy, 3), round(scale, 3)


def _open_erased(src, erase_regions):
    im = Image.open(src).convert("RGB")
    for (l, t, r, b) in erase_regions:
        im.paste((255, 255, 255), (l, t, r, b))
    return im


if __name__ == "__main__":
    # 相对仓库根解析：generated-images/ 是不入库的本地素材目录（AI 原图放这里）
    ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
    G = os.path.join(ROOT, "generated-images")
    UP = os.path.join(ROOT, "server", "data", "uploads")
    ERASE = [(810, 950, 1024, 1024)]  # 右下角 AI 生成水印
    jobs = [
        ("Q版可爱小女孩全身立绘_手机游戏换装系统角色底板风格_扁平风_2026-11-06T07-13-30.png", "board_girl.png"),
        ("把这个角色换成Q版小男孩_棕色短发换成深棕色短发碎发_粉色连_2026-11-06T07-19-59.png", "board_boy.png"),
        ("把这个角色换成Q版小兔子_头上加一对长长白色兔耳朵_带粉色内_2026-11-06T07-20-21.png", "board_bunny.png"),
        ("把这个角色换成Q版小熊猫_头上加圆圆的小熊猫耳朵_白心深棕边_2026-11-06T07-20-45.png", "board_panda.png"),
    ]
    for src, name in jobs:
        key_white(os.path.join(G, src), os.path.join(UP, name), 400, 500, ERASE)
        print("OK", name)
