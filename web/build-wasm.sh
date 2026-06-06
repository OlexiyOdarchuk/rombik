#!/usr/bin/env bash
# Збирає двигун у WASM і кладе всі клієнтські артефакти у web/static/.
# Запускати з кореня репозиторію АБО з web/ (скрипт сам перейде в корінь).
set -euo pipefail
cd "$(dirname "$0")/.."   # корінь репозиторію (де go.mod)

echo "→ rombik.wasm"
GOOS=js GOARCH=wasm go build -o web/static/rombik.wasm ./cmd/wasm

echo "→ wasm_exec.js"
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/

echo "→ parser.py"
cp pkg/parser/python/parser.py web/static/

echo "Готово: web/static/{rombik.wasm, wasm_exec.js, parser.py}"
