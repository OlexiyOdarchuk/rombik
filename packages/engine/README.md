# rombik-engine

Рушій **rombik**: код (**Python / C / C++ / Pascal**) → блок-схема алгоритму за **ДСТУ 19.701‑90**.
Чистий **TypeScript**, нуль рантайм-залежностей, без DOM і фреймворків. Дає геометрію
схеми та рендери у **SVG / Typst (CeTZ) / Excalidraw**.

```
код ──parser──► AST-JSON ──► IR ──layout──► Diagram ──render──► SVG / Typst / Excalidraw
```

```bash
npm install rombik-engine web-tree-sitter
```

> Розбір коду — через `web-tree-sitter` (WASM-граматики `tree-sitter-python` /
> `tree-sitter-cpp`). Рушій приймає вже готове дерево, тож граматику обираєш сам.

## Швидкий старт

```js
import { Parser, Language } from 'web-tree-sitter';
import { fromTree, renderSvg } from 'rombik-engine';

await Parser.init();
const parser = new Parser();
parser.setLanguage(await Language.load('tree-sitter-python.wasm'));

const tree = parser.parse(`def grade(s):\n    if s >= 60:\n        print("ok")`);
for (const { name, diagram } of fromTree(tree, 'python', {})) {
  console.log(name, renderSvg(diagram)); // → <svg>…</svg>
}
```

## API

| Функція | Призначення |
| --- | --- |
| `fromTree(tree, lang, opts)` | tree-sitter-дерево → `[{ name, diagram }]` (одна на функцію) |
| `parseTree(tree, lang)` | дерево → проміжний AST-JSON (для подальшої обробки) |
| `fromAst(ast, opts)` / `fromJson(json)` | AST-JSON → діаграми |
| `splitFromAst(ast, opts, name, maxH)` | розбити схему функції на частини ≤ `maxH` |
| `layoutProgram(funcs, opts)` | розкладка без розбору (низькорівнево) |
| `renderSvg(d)` / `renderSvgAll(ds)` | SVG однієї / усіх схем |
| `renderTypst(d)` / `renderTypstAll(ds)` | Typst (CeTZ); є фрагмент-режим |
| `renderExcalidraw(d)` / `renderExcalidrawAll(ds)` | `.excalidraw` (доредагувати на excalidraw.com) |
| `captionLine`, `labelAnchor` | допоміжні (підпис «Рисунок N», якорі підписів) |

`lang` — `'python' | 'cpp' | 'pascal'` (C → 'cpp'). `opts` (`Options`, усі поля необов'язкові): `singleEnd`,
`callAsProcess`, `stripTypes`, `returnAsIO`, `yes`/`no` (підписи гілок),
`inWord`/`outWord` (слова вводу/виводу), `capWord`. Типи (`Diagram`, `Shape`, `Edge`,
`Point`, `Options`, …) йдуть у комплекті.

PNG/PDF тут немає (це робота растеризатора): рендер дає **SVG**, далі —
браузерним `<canvas>` або зовнішнім `rsvg-convert` / `cairosvg`.

## Мови та конструкції

- **Python** — if/elif/else, match/case, try/except/finally, with, for…in range, while,
  for/else, while/else, while True…break, break/continue/return/raise,
  list/set/dict‑comprehension → цикл, методи й вкладені класи (`Клас.метод`), декоратори.
- **C / C++** — if/else, `switch` (усі case+default, fallthrough → умова через `||`),
  for (класичний і range‑based), while, do‑while, break/continue, рекурсія,
  `cin`/`cout` і `printf`/`scanf`, try/catch, методи класів (`Клас::метод`), оператори,
  namespace, шаблони. C — підмножина C++ (`lang: 'cpp'`).
- **Pascal** — procedure/function, `:=`, writeln/readln, if..then..else,
  for..to/downto..do, while..do, repeat..until, case..of.

## Розробка

```sh
npm test            # node:test (потрібен Node 22+)
npm run coverage    # тести + звіт покриття
npm run typecheck   # tsc --noEmit (strict)
npm run build       # dist/ (JS + .d.ts) для публікації
```

У монорепо веб і CLI беруть **джерела** (умова експорту `rombik-source`), а
npm-споживач — зібраний `dist/`. Частина [rombik](https://github.com/OlexiyOdarchuk/rombik).

## Ліцензія

MIT.
