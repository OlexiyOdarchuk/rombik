---
tags: [architecture, decision]
---

# Чому AST-JSON як контракт

> [!abstract] Рішення
> Між **парсером** (специфічним для мови) і **ядром** (розкладка+рендер)
> стоїть єдиний проміжний формат — **спрощений AST у JSON**. Будь-який фронтенд
> видає його, єдина функція `fromJson` зводить його в [[IR-проміжне-представлення|IR]].

## Що це

Файл `packages/engine/src/astjson.ts` визначає вузол:

```ts
interface AstNode {
  kind: string;            // process|io|terminal|call|if|for|while|dowhile|infloop|break|continue|block
  text?: string;           // для process/io/call/terminal
  cond?: string;           // для if/for/while/dowhile  (для for — це spec)
  stmts?: AstNode[];       // для block
  then?: AstNode | null;
  else?: AstNode | null;
  body?: AstNode | null;
}
interface AstFunc { name: string; block: AstNode; }
```

Парсер видає `AstFunc[]` (рядок JSON або вже розпарсений масив); `fromJson` повертає
`Func[]` (IR).

## Навіщо саме так

### 1. Багатомовність майже безкоштовна

Доданий фронтенд має лише **видати цей JSON**. Уся розкладка, рендер, опції — спільні.
Зараз один парсер (tree-sitter) уже видає AST-JSON і для Python, і для C++ —
зведення в `parser/treesitter.ts` спільне для обох. → [[Як-додати-нову-мову]].

### 2. Ядро не залежить від парсера

Якби ядро напряму викликало tree-sitter, воно тягло б за собою WASM-граматики й
середовище. Натомість межа — рядок AST-JSON:

- **CLI:** `tools/rombik.mjs` парсить через web-tree-sitter і кличе `fromTree`/`fromAst`.
- **Браузер:** `web/src/lib/engine.js` парсить через web-tree-sitter (WASM) і кличе ті самі функції.

Обидва шляхи сходяться у `fromJson`. → [[WASM-міст]], [[Браузерний-рушій]].

### 3. Один парсер на дві мови й два середовища

`parser/treesitter.ts` — один модуль для Python, C, C++ і Pascal (`parseTree(tree, lang)`),
однаковий у вебі й у Node. Tree-sitter уміє обидві граматики, тож фронтенд один, а
не по парсеру на мову. Раніше AST-JSON давали ДВА парсери (Python `ast` і tree-sitter);
тепер — лише tree-sitter, бо він покриває обидві мови.

### 4. JSON — зрозумілий шов

AST-JSON легко роздрукувати, порівняти, замокати в тесті. Межа між «розбором мови» і
«геометрією» стає явною й інспектованою.

## Чому проміжний `AstNode`, а не одразу `ir`

Можна було б десеріалізувати JSON прямо в `ir`. Але:

- `ir` — дискримінований union `Node` зі строгими структурами на кожен `kind`
  (`then`/`else`/`body` як `Block`, без `null`) — підганяти сирий JSON під нього
  без проміжної нормалізації незручно й крихко.
- `AstNode` — пласка структура з усіма опційними полями; `kind` — рядок-дискримінатор.
  `toNode`/`toBlock` роблять явний `switch` і будують типобезпечний `ir` (вкладені
  `block` інлайняться).

Тобто `AstNode` — це «DTO на дроті», `ir` — «доменна модель». Розділення дає чистоту обом.

## Пов'язане

- [[astjson-конвертер]]
- [[IR-проміжне-представлення]]
- [[Розділення-відповідальностей]]
- [[Як-додати-нову-мову]]
