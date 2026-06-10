---
tags: [web, frontend]
---

# Фронтенд SvelteKit

**Тека:** `web/` · SvelteKit (Svelte 5) + Tailwind v4 + CodeMirror, повністю статичний
(`adapter-static`) — без сервера. Лендинг + гайд + повноцінний редактор.

## Стек

- **SvelteKit 2 / Svelte 5** з `adapter-static` — чиста статика (`build/`), хоститься
  будь-де (GitHub Pages). Широко використовуються руни Svelte 5: `$state` (для коду та
  опцій), `$derived` (результати генерації), `$effect` (синхронізація тем та CodeMirror).
- **Tailwind v4** (`@tailwindcss/vite`) — стилі, з підтримкою темної теми.
- **CodeMirror 6** — редактор коду Python/C++ (підсвітка, нумерація, дужки).
- Рантайм-рушій — **Tree-sitter (парсер) + `rombik-engine` (прямий TS-імпорт)**
  ([[Браузерний-рушій]]).

## Структура

```
web/src/
├── app.html, app.css        Tailwind + токени теми (--color-paper тощо)
├── lib/
│   ├── engine.js            клей: Tree-sitter + rombik-engine (TS-рушій)
│   ├── CodeEditor.svelte     CodeMirror 6 (Python та C++); тема через Compartment
│   ├── DiagramEditor.svelte  повноекранний візуальний редактор (полотно, ручне ділення/об'єднання схем)
│   ├── ThemeToggle.svelte    перемикач світла/темна
│   ├── Nav.svelte            шапка: лого-ромбік, навігація, ThemeToggle
│   └── Footer.svelte
└── routes/
    ├── +page.svelte         лендинг (герой, можливості, 3 кроки)
    ├── guide/+page.svelte   «як це працює» + довідник 6 фігур ([[ДСТУ-19.701-90]])
    └── app/+page.svelte     редактор
```

Парсер тепер усередині `rombik-engine` (`parseTree`) — окремого `parser.js` у `web/`
немає. Граматики Tree-sitter (`tree-sitter*.wasm`) лежать у `web/static/`. Збірка
автоматизована й розгортається через **GitHub Actions**. → [[Збірка-і-запуск]].

> [!note] vite та workspace-рушій
> `vite.config` має `optimizeDeps.exclude: ['rombik-engine', 'web-tree-sitter']`:
> `rombik-engine` — workspace-джерело (TS), його треба обробляти як source, а не
> пребандлити; `web-tree-sitter` застосунок не імпортує напряму (бере статичний
> `tree-sitter.js`), а в npm-workspace він hoisted у корінь — пребандл падав ENOENT.

## Темна тема

- Стратегія — **клас `.dark` на `<html>`** (Tailwind `@custom-variant dark`).
- `ThemeToggle` перемикає клас і зберігає вибір у `localStorage['theme']`.
- **CodeMirror** слухає зміну класу (`MutationObserver`) і через `Compartment`
  перемикає тему `oneDark` ↔ світла.
- Прев'ю схеми **завжди на світлому тлі** (SVG-фон білий) — щоб роздруківка збігалася.

## Редактор `/app`

**Розкладка (50/50 на десктопі):** зліва `CodeEditor` (код), справа список згенерованих
схем (`funcs.map(f => ...)`).

**Тулбар:** «Побудувати», модальне вікно **Налаштування**, групова кнопка експорту
(PDF / Typst для всіх схем разом).

**Модальне вікно налаштувань** (UI з вкладками):
- **Структура**: `singleEnd`, `callAsProcess`, `stripTypes`, `returnAsIO`, опції
  розбиття великих схем через з'єднувачі;
- **Текст і підписи**: мітки гілок (Так/Ні · Yes/No · +/−), слова I/O (Ввід · Введення
  · Ввести), шаблон і слова для підпису (Рисунок/Рис./Figure), стартовий номер;
- **Експорт**: налаштування форматування та Excalidraw-опції.

Повний перелік перемикачів — [[Опції-рендера]].

**Картка кожної схеми** (інлайн у `routes/app/+page.svelte`): ім'я функції, кнопки
експорту **SVG / PNG / Typst / PDF / Excalidraw**, поля редагування підпису (номер +
текст), прев'ю SVG, кнопка **«✎ Редагувати»** → відкриває візуальний редактор.

**Візуальний редактор** (`DiagramEditor.svelte`) — повноекранне безмежне полотно
(пан+зум), тягання блоків/стрілок, правка тексту, undo/redo. **Ручне ділення:**
виділення ласо/Shift-кліком → «⊞ Окрема схема» / «⊟ Об'єднати»; крос-групові стрілки
стають парами конекторів. «Редагувати всі» — усі функції на одному полотні.

## Потік у редакторі

1. Вибір мови та код + опції → `engine.generate` → Tree-sitter парсить, `parseTree`
   конвертує, `rombik-engine` (`fromAst`) розкладає, `renderSvg` рендерить
   ([[Браузерний-рушій]]).
2. Результат `{functions:[{name, svg, diagram}], warning?}` → `svg` у DOM.
3. Правка підпису → `renderCaption` (дешевий ре-рендер, без парсингу).
4. Експорт: SVG/Typst/Excalidraw — напряму з рушія; **PNG/PDF — браузерний растр**
   (SVG→canvas + jsPDF) ([[WASM-міст]]).

## Пов'язане

- [[Браузерний-рушій]] · [[WASM-міст]]
- [[Опції-рендера]] · [[Збірка-і-запуск]]
</content>
