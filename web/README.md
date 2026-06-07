# rombik web

Фронтенд: SvelteKit (Svelte 5) + Tailwind v4, повністю статичний (`adapter-static`) —
без сервера. Лендинг + гайд + редактор (`/app`) з CodeMirror і темною темою.

## Розробка

Спершу згенеруй WASM-артефакти (з кореня репозиторію), потім запускай фронт:

```bash
./build-wasm.sh        # або з кореня: ./web/build-wasm.sh
npm install
npm rebuild esbuild    # npm 11 блокує install-скрипти; esbuild потребує бінарник
npm run dev            # http://localhost:5173
npm run build          # статика у build/
```

## Двигун у браузері (без сервера)

Усе працює клієнтсько:

- **Pyodide** (CPython у WASM) виконує `parser.py` → AST-JSON;
- **rombik.wasm** (легкий Go-двигун) бере AST-JSON + опції → `{functions:[{name, svg,
  typst, diagram}]}` (`rombikGenerate`); має ще `rombikRenderOne` (живий ре-рендер
  підпису) і `rombikTypstAll`;
- **rombik-raster.wasm** (важкий, ~16 МБ) — нативні PNG/PDF (`rombikPng`/`rombikPdf`/
  `rombikPdfAll`); підвантажується **ліниво**, лише на першому експорті PNG/PDF.

Клей — `src/lib/engine.js` (`warmup`/`generate`/`renderCaption`/`renderPng`/`renderPdf`…).

Артефакти у `static/` готує `build-wasm.sh`:

```bash
GOOS=js GOARCH=wasm go build -o web/static/rombik.wasm        ./cmd/wasm
GOOS=js GOARCH=wasm go build -o web/static/rombik-raster.wasm ./cmd/wasmraster
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/
cp pkg/parser/python/parser.py web/static/
```

Деталі — у `brain/04-Веб/` ([[Браузерний-двигун]], [[WASM-міст]], [[Фронтенд-SvelteKit]]).
