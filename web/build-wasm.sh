#!/usr/bin/env bash
# Збирає рушій у WASM і кладе всі клієнтські артефакти у web/static/.
# Запускати з кореня репозиторію АБО з web/ (скрипт сам перейде в корінь).
set -euo pipefail
cd "$(dirname "$0")/.."   # корінь репозиторію (де go.mod)

echo "→ rombik.wasm (легкий: парсер→схеми, SVG/Typst)"
GOOS=js GOARCH=wasm go build -o web/static/rombik.wasm ./cmd/wasm

echo "→ rombik-raster.wasm (важкий: нативний PNG/PDF, вантажиться лениво)"
GOOS=js GOARCH=wasm go build -o web/static/rombik-raster.wasm ./cmd/wasmraster

echo "→ wasm_exec.js"
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/

echo "Готово: web/static/{rombik.wasm, rombik-raster.wasm, wasm_exec.js}"
