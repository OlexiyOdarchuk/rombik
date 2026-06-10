#!/usr/bin/env bash
# Збирає растеризатор у WASM (PNG/PDF) у web/static/. Рушій (парсер→схеми→SVG/Typst/
# Excalidraw) тепер чистий TS (@rombik/engine) — легкий rombik.wasm більше не потрібен.
# TODO: викинути і растер-WASM, коли PNG/PDF переедемо на браузерний SVG→canvas.
set -euo pipefail
cd "$(dirname "$0")/.."   # корінь репозиторію (де go.mod)

echo "→ rombik-raster.wasm (важкий: нативний PNG/PDF, вантажиться лениво)"
GOOS=js GOARCH=wasm go build -o web/static/rombik-raster.wasm ./cmd/wasmraster

echo "→ wasm_exec.js"
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/

echo "Готово: web/static/{rombik-raster.wasm, wasm_exec.js}"
