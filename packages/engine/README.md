# @rombik/engine

Перетворення структурованого коду (Python / C++) на блок-схему алгоритму за
**ДСТУ 19.701-90**. Чистий TypeScript, **без залежностей від DOM чи фреймворку** —
працює у браузері, Node і будь-де.

```
код ──parser──► AST-JSON ──► IR ──layout──► Diagram ──render──► SVG / Typst / …
```

## Структура

```
src/
  index.ts          публічний API
  diagram.ts        модель геометрії (фігури, ребра) + підпис
  ir.ts             IR — мова-агностик дерево керування            (TODO)
  parser/           відображення tree-sitter → AST-JSON            (TODO: порт з web/parser.js)
  layout/           розкладка IR → Diagram (рекурсивна геометрія)  (TODO: порт з pkg/layout)
  render/
    svg.ts          SVG-рендерер (готово, byte-парність із Go)
    typst.ts        Typst                                          (TODO)
    excalidraw.ts   Excalidraw                                     (TODO)
test/
  svg.test.ts       golden-парність (node:test)
  golden/*.json     еталони, зняті з Go-рушія (оракул міграції)
  corpus/           вхідні приклади (Python + C++)
  capture-golden.mjs  тимчасовий: регенерує оракул зі старого Go-рушія
```

## Статус міграції Go → TS

Рушій історично на Go (компілювався у WASM для вебу). Триває порт на TS, щоб
прибрати кордон Go↔wasm і двомовність. **Інваріант:** кожен портований модуль
має давати БАЙТ-У-БАЙТ той самий вивід, що Go — це стереже `test/golden`.

Готово: модель + SVG-рендерер. Далі: `parser` → `ir` → `layout`.

## Тести

```sh
npm test            # node:test, golden-парність SVG (потрібен Node 22+)
npm run typecheck   # tsc --noEmit (strict)
```
