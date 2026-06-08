---
tags: [web, frontend]
---

# Фронтенд SvelteKit

**Тека:** `web/` · SvelteKit (Svelte 5) + Tailwind v4, повністю статичний
(`adapter-static`) — без сервера. Лендинг + гайд + повноцінний редактор.

## Стек

- **SvelteKit 2 / Svelte 5** з `adapter-static` — чиста статика (`build/`), хоститься
  будь-де (GitHub Pages, CDN).
- **Tailwind v4** (`@tailwindcss/vite`) — стилі, з підтримкою темної теми.
- **CodeMirror 6** — редактор коду Python (підсвітка, нумерація, дужки).
- Рантайм-рушій — **Pyodide + два WASM** ([[Браузерний-рушій]]).

## Структура

```
web/src/
├── app.html, app.css        Tailwind + токени теми (--color-paper тощо)
├── lib/
│   ├── engine.js            клей Pyodide + два WASM-рушійи
│   ├── CodeEditor.svelte    CodeMirror 6 (Python); тема через Compartment
│   ├── ThemeToggle.svelte   перемикач світла/темна
│   ├── Nav.svelte           шапка: лого-ромбік, навігація, ThemeToggle
│   └── Footer.svelte
└── routes/
    ├── +page.svelte         лендинг (герой, можливості, 3 кроки)
    ├── guide/+page.svelte   «як це працює» + довідник 6 фігур ([[ДСТУ-19.701-90]])
    └── app/+page.svelte     редактор
```

Артефакти `rombik.wasm`, `rombik-raster.wasm`, `wasm_exec.js`, `parser.py` —
**генеровані** (`build-wasm.sh`), у `.gitignore` `*.wasm`. → [[Збірка-і-запуск]].

## Темна тема

- Стратегія — **клас `.dark` на `<html>`** (Tailwind `@custom-variant dark`).
- `ThemeToggle` перемикає клас і зберігає вибір у `localStorage['theme']`.
- **CodeMirror** слухає зміну класу (`MutationObserver`) і через `Compartment`
  перемикає тему `oneDark` ↔ світла.
- Прев'ю схеми **завжди на світлому тлі** (SVG-фон білий) — щоб роздруківка збігалася.

## Редактор `/app`

**Розкладка (50/50 на десктопі):** зліва `CodeEditor`, справа список схем.

**Тулбар:** «Побудувати», випадайна панель **Налаштування**, групова кнопка експорту
(PDF / Typst для всіх схем разом).

**Панель налаштувань** (галочки/списки) → серіалізуються в `optionsJSON`:
- структура: `singleEnd`, `callAsProcess`, `stripTypes`, `returnAsIO`;
- мітки гілок (Так/Ні · Yes/No · +/−), слова I/O (Ввід · Введення · Ввести);
- **підпис**: показувати «Рисунок N», слово (Рисунок/Рис./Figure), шаблон, стартовий номер.

Повний перелік перемикачів — [[Опції-рендера]].

**Картка кожної схеми:** ім'я функції, кнопки експорту **SVG / PNG / Typst / PDF**,
поля редагування підпису (номер + текст), прев'ю SVG.

## Потік у редакторі

1. Код + опції → `engine.generate` → Pyodide парсить, легкий WASM розкладає й рендерить
   ([[Браузерний-рушій]]).
2. Результат `{functions:[{name, svg, typst, diagram}]}` → `svg` у DOM.
3. Правка підпису → `renderCaption` (дешевий ре-рендер, без парсингу).
4. Експорт: SVG/Typst — напряму з легкого модуля; **PNG/PDF — ліниво** через
   растровий WASM ([[WASM-міст]]).

## Пов'язане

- [[Браузерний-рушій]] · [[WASM-міст]]
- [[Опції-рендера]] · [[Збірка-і-запуск]]
