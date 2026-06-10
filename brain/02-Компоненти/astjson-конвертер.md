---
tags: [component, parser]
---

# astjson — конвертер

**Модуль:** `astjson.ts` (`packages/engine/src/astjson.ts`)

Єдина точка перетворення **[[Чому-AST-JSON-як-контракт|AST-JSON]] → [[IR-проміжне-представлення|IR]]**.
Спільна для всіх мов: будь-який фронтенд парсера (нині — tree-sitter для Python, C, C++ і Pascal)
видає цей формат, а тут — єдина конвертація в IR.

## Типи на дроті

```ts
export interface AstNode {
  kind: string;
  text?: string;          // process/io/call/terminal
  cond?: string;          // if/for/while/dowhile (для for — це spec)
  stmts?: AstNode[];      // block
  then?: AstNode | null;
  else?: AstNode | null;
  body?: AstNode | null;
}
export interface AstFunc { name: string; block: AstNode; }
```

Пласка структура з усіма можливими полями; `kind` — рядок-дискримінатор. Чому не
будувати одразу `ir` — пояснено в [[Чому-AST-JSON-як-контракт]].

## Дві функції

### `fromJson(data: string | AstFunc[]): Func[]`

Точка входу. Приймає JSON-рядок або вже розпарсений масив `AstFunc`, кожен зводить у
`ir.Func`:

```ts
fns.map((f) => ({ name: f.name, body: toBlock(f.block) }));
```

Це те, у що впадає вихід `parseTree` (через `fromTree`/`fromAst` у [[Публічний-API-rombik|build.ts]]).

### `toBlock(n): Block`

Зводить вузол-блок у `ir.Block`. Важлива деталь — **інлайнінг вкладених блоків**:

```ts
if (c.kind === 'block') stmts.push(...toBlock(c).stmts);  // розгортаємо, а не вкладаємо
else { const node = toNode(c); if (node) stmts.push(node); }
```

Це потрібно, бо парсер загортає `with`/`try`/`match` у проміжні `block`-вузли —
розкладка не має бачити цю штучну вкладеність.

### `toNode(n): Node | null` (внутрішня)

Switch по `kind` → конкретний варіант union `ir.Node`:

```
process → Process   io → IO       terminal → Terminal   call → Call
if → If             for → For{spec, body, else}        while → While{cond, body, else}
dowhile → DoWhile   infloop → InfLoop  break → Break    continue → Continue
(інше → null, тихо пропускається)
```

Зверни увагу:
- для `for` поле AST-JSON зветься `cond`, а в `ir.For` — `spec` («змінна = початок,
  кінець, крок»). Конвертер це мапить.
- `for`/`while` несуть **`else`** (гілка `for/else`/`while/else`) — береться з
  `n.else`. → [[IR-проміжне-представлення]].
- `break`/`continue` — обидва стрибки без фігури.
- `match`/`switch` сюди не доходять окремим видом: парсер зводить їх у ланцюг `if`.

## Пов'язане

- [[Чому-AST-JSON-як-контракт]]
- [[IR-проміжне-представлення]]
- [[Парсер-Python]]
- [[Публічний-API-rombik]]
