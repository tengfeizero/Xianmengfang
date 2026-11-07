#!/bin/bash
# 对所有预览页截图（依赖旧版 headless Edge，见 render.js 头注释）
set -e
cd "$(dirname "$0")/../.."
EDGE="/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe"
OUT="$(pwd)/shots/mp"
mkdir -p "$OUT"
for f in tools/mp-preview/out/*.html; do
  name=$(basename "$f" .html)
  "$EDGE" --headless --disable-gpu --no-sandbox \
    --user-data-dir="$(pwd)/.tmp-edge" \
    --window-size=375,812 --force-device-scale-factor=2 --hide-scrollbars \
    --screenshot="$(pwd)/shots/mp/$name.png" \
    "file:///$(pwd)/tools/mp-preview/out/$name.html" >/dev/null 2>&1 || echo "shot fail: $name"
done
ls "$OUT"
