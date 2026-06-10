---
tags: [component, api, facade]
---

# Публічний API (rombik-engine)

**Пакет:** `rombik-engine` · **Модуль:** `index.ts` (`packages/engine/src/index.ts`)
**Оркестрація:** `build.ts`

Публічний бар'єр рушія: `index.ts` ре-експортує типи й функції всіх внутрішніх модулів,
а `build.ts` склеює парсер → розкладку → рендери в кілька зручних викликів. Це точка
входу для споживачів (CLI, веб, інші пакети). Рушій — **чистий TS без DOM**.

## Що експортує index.ts

```ts
// модель геометрії
export type { Diagram, Shape, Edge, Point, Kind } from './diagram.ts';
export { captionLine, labelAnchor } from './diagram.ts';
// IR
export type { Node, Func, Block, If, For, While } from './ir.ts';
// AST-контракт
export type { AstNode, AstFunc } from './astjson.ts';
export { fromJson } from './astjson.ts';
// опції + розкладка
export type { Options } from './layout/options.ts';
export { layoutProgram } from './layout/place.ts';
// оркестрація
export { fromAst, fromTree, splitFromAst, splitByHeight, type Result } from './build.ts';
// парсер
export { parseTree, type TSNode, type TSTree, type Lang } from './parser/treesitter.ts';
// рендери
export { render as renderSvg, renderAll as renderSvgAll } from './render/svg.ts';
export { render as renderTypst, renderAll as renderTypstAll,
         fragment as renderTypstFragment, fragmentAll as renderTypstFragmentAll } from './render/typst.ts';
export { render as renderExcalidraw, renderAll as renderExcalidrawAll } from './render/excalidraw.ts';
```

## Тип Result

```ts
export interface Result { name: string; diagram: Diagram; }
```

## Конструктори (код/AST → схеми) — build.ts

```ts
function fromTree(tree: TSTree, lang: Lang, opts?: Options): Result[]   // повний шлях код→схема в TS
function fromAst(astJSON: string | AstFunc[], opts?: Options): Result[] // з готового AST-JSON
function splitFromAst(astJSON, opts, name, maxH): Result[]              // розбити довгу схему на частини
function splitByHeight(f: Func, opts, maxH): Result[]                   // те саме на рівні IR-Func
```

- `fromTree` = `parseTree` → `fromJson` → `buildResults`. Це наскрізний шлях
  «дерево tree-sitter (Python/C/C++/Pascal) → схеми» цілком у TS.
- `fromAst` стартує з уже розпарсеного [[Чому-AST-JSON-як-контракт|AST-JSON]].
- Усі йдуть через внутрішній `buildResults`, який на кожну функцію кличе
  [[Layout-рушій-розкладки|layoutProgram]] і **засіває підпис**: `caption` = ім'я
  функції, `figNum` = порядковий номер (1..N), `capWord` = опція. → [[Diagram-модель-геометрії]].

### Розбиття довгих схем

`splitByHeight(f, opts, maxH)` жадібно ріже тіло функції на частини висотою ≤ `maxH`;
на стиках вставляє пару конекторів-`Connector` (вихід попередньої частини → вхід
наступної, літери `А`,`Б`,…) і виставляє `noStart`/`noEnd`, щоб термінатори були лише
на краях. Якщо схема вміщається — одна частина. → [[IR-проміжне-представлення]].

## Рендери на результаті

`Result.diagram` віддають у будь-який рендер з `index.ts`:

```ts
renderSvg(r.diagram)         // SVG-рядок
renderTypst(r.diagram)       // .typ (CeTZ)
renderExcalidraw(r.diagram)  // .excalidraw
```

PNG/PDF — **не в рушії**: їх дає браузер (SVG→canvas+jsPDF). → [[Растровий-рендер-PNG-PDF]].

## Парність із Go

Алгоритм layout/рендерів портовано з Go **байт-у-байт**; парність стережуть
**golden-тести** (`packages/engine/test/golden`, запуск через `node:test`).

## Чому такий бар'єр

- **Зручність:** споживачу не треба знати про `astjson`, `layout`, окремі рендери —
  лише `from*` + рендер-функції на `diagram`.
- **Єдине засівання підпису:** логіка «ім'я функції → caption, номер → figNum» в одному
  місці (`buildResults`), а не дублюється у CLI/вебі.
- **Чисті межі:** `index.ts` нічого не додає до ядра — лише ре-експортує. → [[Розділення-відповідальностей]].

## Пов'язане

- [[Layout-рушій-розкладки]] · [[Diagram-модель-геометрії]]
- [[SVG-рендерер]] · [[Typst-рендер]] · [[Растровий-рендер-PNG-PDF]]
- [[CLI-довідник]] · [[Браузерний-рушій]]
