---
tags: [web, frontend]
---

# Фронтенд SvelteKit

**Тека:** `web/` · SvelteKit + Tailwind v4, повністю статичний (`adapter-static`) — без
сервера. Лендинг + гайд + редактор.

## Стек

- **SvelteKit** з `adapter-static` — на виході чиста статика (`build/`), хоститься
  будь-де (GitHub Pages, CDN).
- **Tailwind v4** — стилі.
- **Vite** — збірка/дев-сервер.
- Рантайм-двигун — **Pyodide + rombik.wasm** ([[Браузерний-двигун]]).

## Структура

```
web/
├── build-wasm.sh         збірка WASM-артефактів у static/
├── src/
│   ├── app.html, app.css
│   ├── lib/
│   │   ├── engine.js      ← клей Pyodide+WASM (warmup/generate)
│   │   ├── Nav.svelte
│   │   └── Footer.svelte
│   └── routes/
│       ├── +layout.svelte / +layout.js
│       ├── +page.svelte         лендинг
│       ├── guide/+page.svelte   «як це працює» + довідник фігур
│       └── app/+page.svelte     редактор
└── static/
    ├── rombik.wasm      Go-двигун (артефакт)
    ├── wasm_exec.js      міст Go↔JS (артефакт)
    ├── parser.py         парсер для Pyodide (артефакт)
    └── favicon.svg
```

> `rombik.wasm`, `wasm_exec.js`, `parser.py` — **генеровані** артефакти
> (`build-wasm.sh`); у `.gitignore` `*.wasm` ігнорується. Перед `npm run dev` їх треба
> згенерувати. → [[Збірка-і-запуск]].

## Сторінки

- **`/`** — лендинг: що це, можливості, заклик спробувати.
- **`/guide`** — пояснення конвеєра + довідник ДСТУ-фігур ([[ДСТУ-19.701-90]]).
- **`/app`** — редактор: поле коду → `engine.generate(code, options)` → перегляд SVG +
  експорт (SVG/PNG/JSON). Опції рендера — галочки, що йдуть у `optionsJSON`
  ([[Опції-рендера]]).

## Потік у редакторі

1. На вхід — Python-код і обрані опції.
2. `engine.generate` → Pyodide парсить, rombik.wasm розкладає
   ([[Браузерний-двигун]]).
3. Результат `{functions:[{name, svg, diagram}]}` — вставляємо `svg` у DOM.
4. Експорт: SVG напряму; PNG — через canvas у браузері; JSON — поле `diagram`.

## Пов'язане

- [[Браузерний-двигун]]
- [[WASM-міст]]
- [[Збірка-і-запуск]]
- [[Опції-рендера]]
