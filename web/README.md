# flowgen web

Фронтенд: SvelteKit + Tailwind v4, повністю статичний (`adapter-static`) — без сервера.
Лендинг + гайд + редактор (`/app`).

## Розробка

```bash
npm install
npm rebuild esbuild   # npm 11 блокує install-скрипти; esbuild потребує бінарник
npm run dev           # http://localhost:5173
npm run build         # статика у build/
```

## Двигун у браузері (наступний крок)

Редактор працюватиме клієнтсько, без сервера:

- **Pyodide** (CPython у WASM) виконує `parser.py` → AST-JSON;
- **flowgen.wasm** (Go-двигун) бере AST-JSON + опції → `{functions:[{name, svg, diagram}]}`.

Артефакти у `static/` готує скрипт (з кореня репозиторію):

```bash
GOOS=js GOARCH=wasm go build -o web/static/flowgen.wasm ./cmd/wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/
cp internal/parser/python/parser.py web/static/
```
