# -*- coding: utf-8 -*-
"""批量差分抠部件：底板(小姑娘) vs 变体图 → 元素 PNG + 锚点清单。
输出: server/data/uploads/elements/*.png + anchors.json（名字→{x,y,scale}）"""
import json
import os
import sys

sys.path.insert(0, os.path.dirname(__file__))
from ai_assets import extract_diff

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
G = os.path.join(ROOT, "generated-images")          # AI 原图（不入库）
UP = os.path.join(ROOT, "server", "data", "uploads")
ERASE = [(810, 950, 1024, 1024)]  # 右下角水印

BASE = os.path.join(G, "Q版可爱小女孩全身立绘_手机游戏换装系统角色底板风格_扁平风_2026-11-06T07-13-30.png")

JOBS = [
    # (元素文件, 中文名, 变体图)
    ("hat_santa.png", "圣诞帽", "给这个角色戴上一顶红色圣诞帽_白色绒毛帽边_帽顶白色绒球_帽_2026-11-06T07-23-08.png"),
    ("hat_pointy.png", "红色尖尖帽", "给这个角色戴上一顶红色尖尖魔法帽_帽顶微弯_帽顶有个白色小绒_2026-11-06T07-27-51.png"),
    ("hat_beret.png", "蓝色贝雷帽", "给这个角色头上戴一顶蓝色贝雷帽_歪戴_帽顶有个黄色小帽扣___2026-11-06T07-28-08.png"),
    ("hat_straw.png", "黄色草帽", "给这个角色戴上一顶黄色草帽_宽帽檐_帽身有红色缎带__其余一_2026-11-06T07-28-20.png"),
    ("hat_crown.png", "金色皇冠", "给这个角色头上戴一个金色小皇冠_三个尖角_镶红色宝石__其余_2026-11-06T07-28-36.png"),
    ("hair_twin.png", "双马尾", "在她现有发型两侧各加一条翘起的棕色双马尾_用粉色发圈扎住_其_2026-11-06T07-26-19.png"),
    ("hair_side.png", "侧马尾", "在她头部左侧加一条粗粗蓬松的棕色侧马尾辫_用粉色大蝴蝶结扎住_2026-11-06T07-46-24.png"),
    ("hair_bun.png", "丸子头", "在她头顶正中加一个棕色丸子发髻_用粉色发圈装饰_其余一切_脸_2026-11-06T07-26-48.png"),
    ("hair_buns.png", "双丸子", "在她头顶两侧各加一个棕色丸子发髻_双丸子头__其余一切_脸__2026-11-06T07-27-02.png"),
    ("hair_curly.png", "卷发", "在她头发两侧各加一大缕棕色卷曲发卷垂到肩部_其余一切_脸_表_2026-11-06T07-27-18.png"),
    ("exp_smile.png", "笑眼弯弯", "把她的大眼睛完全换成两条弯弯的开心眯眯眼_眼睑向下弯成月牙形_2026-11-06T07-29-08.png"),
    ("exp_heart.png", "爱心眼", "把她的两只眼睛瞳孔中央各换成一颗小小的粉色爱心_爱心只占瞳孔_2026-11-06T07-45-58.png"),
    ("exp_wink.png", "眨眼", "把她的左眼_画面右侧那只_换成弯弯的闭眼笑线条_右眼保持原来_2026-11-06T07-29-34.png"),
    ("exp_pout.png", "嘟嘟嘴", "把她的微笑嘴巴换成一个明显更大的粉红色嘟嘟嘴_噘起的小圆唇__2026-11-06T07-46-11.png"),
    ("lash_star.png", "星星睫毛", "给她两只眼睛的上眼睑各加上三根又长又粗又翘的黑色可爱睫毛_非_2026-11-06T07-46-37.png"),
    ("outfit_apron.png", "小围裙", "给她系上一条白色小围裙_细肩带_围裙主体覆盖裙子的下半段和腰_2026-11-06T07-47-09.png"),
    ("outfit_cape.png", "红色小披风", "给她披上一件红色连帽小披风_从头顶两侧肩膀完整地垂下来_整块_2026-11-06T07-47-24.png"),
    ("outfit_scarf.png", "蓝色围巾", "给这个角色脖子上围一条蓝色针织围巾_绕脖子一圈_一侧垂下一截_2026-11-06T07-31-01.png"),
    ("outfit_bowtie.png", "红色领结", "给这个角色领口加一个红色大领结_系在裙子蝴蝶结上方正中__其_2026-11-06T07-31-12.png"),
    ("acc_bow.png", "蝴蝶结", "在她头发左侧_画面左侧_加一个粉色大蝴蝶结发饰_其余一切_脸_2026-11-06T07-31-43.png"),
    ("acc_glasses.png", "圆框眼镜", "给这个角色戴上一副棕色圆框眼镜_两个圆镜片加细镜腿_镜片完全_2026-11-06T07-31-56.png"),
    ("acc_balloon.png", "粉色气球", "在她头部右上方的空白处加一只圆润饱满的粉色气球_没有绳子_气_2026-11-06T07-47-38.png"),
    ("acc_lollipop.png", "棒棒糖", "在她右手_画面左侧的手_里加一支粉色螺旋纹大棒棒糖_糖在头部_2026-11-06T07-32-22.png"),
    ("acc_starclip.png", "星星发夹", "在她头发右侧_画面右侧_加一个金色星星小发夹_闪闪发光的四角_2026-11-06T07-32-34.png"),
]

# 小面积部件（嘴等）用更低阈值
LOW_T = {"exp_pout.png": (45, 60), "lash_star.png": (45, 40), "acc_starclip.png": (45, 40)}

anchors = {}
for fname, name, variant in JOBS:
    out = os.path.join(UP, "elements", fname)
    if os.path.exists(out):
        print("SKIP", name)
        continue
    diff_t, keep_min = LOW_T.get(fname, (60, 300))
    dst, ax, ay, sc = extract_diff(BASE, os.path.join(G, variant),
                                   out, 400, 500, ERASE, diff_t=diff_t, keep_min=keep_min)
    anchors[name] = {"file": fname, "x": ax, "y": ay, "scale": sc}
    print("OK %-20s anchor=(%.3f,%.3f) scale=%.3f" % (name, ax, ay, sc))

# 已提取过的也并入 anchors.json（上次运行结果）
prev = {"圣诞帽": ("hat_santa.png", 0.454, 0.194, 0.753), "红色尖尖帽": ("hat_pointy.png", 0.505, 0.172, 0.839),
        "蓝色贝雷帽": ("hat_beret.png", 0.562, 0.161, 1.0), "黄色草帽": ("hat_straw.png", 0.504, 0.218, 0.628),
        "金色皇冠": ("hat_crown.png", 0.5, 0.086, 1.0), "双马尾": ("hair_twin.png", 0.5, 0.269, 0.613),
        "侧马尾": ("hair_side.png", 0.347, 0.495, 1.0), "丸子头": ("hair_bun.png", 0.503, 0.05, 1.0),
        "双丸子": ("hair_buns.png", 0.5, 0.143, 0.748), "卷发": ("hair_curly.png", 0.501, 0.446, 0.774),
        "笑眼弯弯": ("exp_smile.png", 0.501, 0.335, 1.0), "爱心眼": ("exp_heart.png", 0.5, 0.334, 1.0),
        "眨眼": ("exp_wink.png", 0.587, 0.334, 1.0)}
for name, (f, ax, ay, sc) in prev.items():
    anchors.setdefault(name, {"file": f, "x": ax, "y": ay, "scale": sc})

with open(os.path.join(UP, "elements", "anchors.json"), "w", encoding="utf-8") as f:
    json.dump(anchors, f, ensure_ascii=False, indent=2)
print("anchors.json written,", len(anchors), "items")
