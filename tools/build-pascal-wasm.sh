#!/usr/bin/env bash
# Збирає tree-sitter-pascal.wasm для web/static (потрібен docker для емскриптена).
# Граматика — npm-пакет tree-sitter-pascal; регенеруємо парсер під поточний
# tree-sitter і збираємо wasm. Запуск із кореня репо.
set -euo pipefail
TMP="$(mktemp -d)"
TS="$(pwd)/web/node_modules/.bin/tree-sitter"
( cd "$TMP" && npm pack tree-sitter-pascal && tar xzf tree-sitter-pascal-*.tgz )
( cd "$TMP/package" && "$TS" generate && "$TS" build --wasm )
cp "$TMP/package/tree-sitter-pascal.wasm" web/static/
echo "Готово: web/static/tree-sitter-pascal.wasm"
rm -rf "$TMP"
