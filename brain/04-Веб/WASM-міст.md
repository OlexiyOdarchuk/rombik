---
tags: [web, component]
---

# Без Go-WASM: Tree-sitter + прямий TS-рушій

> [!important] Що змінилось
> Раніше рушій у браузері був **Go**, скомпільований у два WASM-модулі (легкий
> `rombik.wasm` для SVG/Typst + важкий `rombik-raster.wasm` для PNG/PDF), і `engine.js`
> був мостом Go↔JS через `wasm_exec.js`. **Після міграції на TypeScript Go-WASM немає
> зовсім.** Рушій — це звичайний TS-пакет `rombik-engine`, що бандлиться vite разом із
> застосунком. ЄДИНИЙ WASM, що лишився, — **граматики Tree-sitter** (`web-tree-sitter`),
> і це **парсер**, а не рушій.

## Дві частини рантайму

| Частина | Що це | WASM? |
|---------|-------|-------|
| **Парсер** | `web-tree-sitter` + граматики `tree-sitter-python/cpp.wasm` | так (граматики) |
| **Рушій** | `rombik-engine` (TypeScript: layout + рендери) | ні, чистий TS |

Парсер дає синтаксичне дерево `Tree`; рушій бере його й видає
[[Diagram-модель-геометрії|Diagram]] та формати. Межа між ними — AST
([[Чому-AST-JSON-як-контракт]], [[Розділення-відповідальностей]]).

## Рушій `rombik-engine`

Пакет `packages/engine` (`exports` → `src/index.ts`). Публічні точки входу замість
колишніх глобальних `rombik*`-функцій:

| Функція | Вхід | Вихід |
|---------|------|-------|
| `fromTree(tree, lang, opts?)` | Tree-sitter `Tree` | `Result[]` (`{name, diagram}`) |
| `fromAst(astJSON, opts?)` | AST-JSON / масив `AstFunc` | `Result[]` |
| `splitFromAst(astJSON, opts, name, maxH)` | + ім'я, макс. висота | `Result[]` (частини через конектори) |
| `renderSvg` / `renderSvgAll` | `Diagram` | SVG |
| `renderTypst` / `renderTypstAll` (+ `fragment…`) | `Diagram` | Typst |
| `renderExcalidraw` / `renderExcalidrawAll` | `Diagram` | JSON `.excalidraw` |
| `parseTree(tree, lang)` | Tree-sitter `Tree` | AST-JSON |
| `layoutProgram(prog, o)` | IR-блок | `Diagram` (ядро розкладки) |

**Живе редагування підпису** тепер тривіальне: фронт міняє `caption/figNum/capWord`
у вже наявному `Diagram`-об'єкті (`{...diagram, ...cap}`) і кличе `renderSvg`/`renderTypst`
заново — **без повторного парсингу**, без серіалізації через WASM-межу.

## PNG/PDF — браузерний растр (без важкого WASM)

Колишній `rombik-raster.wasm` (tdewolff/canvas, ~16 МБ, lazy) **прибрано**. Тепер
растеризація — суто браузерна: `renderSvg` → `<img>` зі SVG → `<canvas>` (білий фон) →
PNG; PDF — той самий растр посторінково через **jsPDF** (лінивий `import('jspdf')`).
Деталі — [[Браузерний-рушій]].

## Vite-нюанс

`optimizeDeps.exclude: ['rombik-engine', 'web-tree-sitter']`:
- `rombik-engine` — workspace-джерело (TS), обробляється як source, не пребандлиться;
- `web-tree-sitter` hoisted у корінь монорепо (застосунок бере статичний
  `tree-sitter.js`), тож пребандл vite падав ENOENT у `web/node_modules`.

## Чому так краще, ніж два Go-WASM

- **Швидкий старт:** немає 4 МБ + 16 МБ WASM-завантажень; рушій — частина JS-бандла.
- **Без межі серіалізації:** `Diagram` — звичайний JS-об'єкт, не треба
  marshal/unmarshal через `js.FuncOf`, нема `select{}`-трюку, щоб тримати Go-`main` живим.
- **Та сама межа даних:** усі точки входу беруть `Tree`/AST/`Diagram` і кличуть спільне
  ядро (`layoutProgram` + рендери) — як і CLI на Node. → [[Конвеєр-обробки]].

## Пов'язане

- [[Браузерний-рушій]] · [[Фронтенд-SvelteKit]]
- [[astjson-конвертер]] · [[Layout-рушій-розкладки]]
- [[Typst-рендер]]
- [[Розділення-відповідальностей]]
</content>
