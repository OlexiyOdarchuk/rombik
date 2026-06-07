---
tags: [component, parser]
---

# astjson — конвертер

**Пакет:** `pkg/parser/astjson` · **Файл:** `astjson.go`

Єдина точка перетворення **[[Чому-AST-JSON-як-контракт|AST-JSON]] → [[IR-проміжне-представлення|IR]]**.
Спільна для всіх мов і середовищ — і для `python3` (CLI), і для Pyodide (браузер).

## Типи на дроті

```go
type Node struct {
    Kind  string `json:"kind"`
    Text  string `json:"text"`   // process/io/call/terminal
    Cond  string `json:"cond"`   // if/for/while/dowhile (для for — Spec)
    Stmts []Node `json:"stmts"`  // block
    Then  *Node  `json:"then"`
    Else  *Node  `json:"else"`
    Body  *Node  `json:"body"`
}
type Func struct {
    Name  string `json:"name"`
    Block Node   `json:"block"`
}
```

Пласка структура з усіма можливими полями; `Kind` — рядок-дискримінатор. Чому не
десеріалізувати одразу в `ir` — пояснено в [[Чому-AST-JSON-як-контракт]].

## Три функції

### `FromJSON(data []byte) ([]ir.Func, error)`

Точка входу. Розбирає JSON-масив `Func`, кожен зводить у `ir.Func`:

```go
res = append(res, ir.Func{Name: fns[i].Name, Body: ToBlock(&fns[i].Block)})
```

Це **те, що кличе WASM** напряму (`cmd/wasm`), і те, у що впадає `python.ParseAll`.

### `ToBlock(n *Node) *ir.Block`

Зводить вузол-блок у `ir.Block`. Важлива деталь — **інлайнінг вкладених блоків**:

```go
if c.Kind == "block" {
    b.Stmts = append(b.Stmts, ToBlock(c).Stmts...)  // розгортаємо, а не вкладаємо
    continue
}
```

Це потрібно, бо `parser.py` загортає `with`/`try` у проміжні `block`-вузли —
розкладка не має бачити цю штучну вкладеність.

### `ToNode(n *Node) ir.Node`

Switch по `Kind` → конкретний тип `ir`:

```
process → ir.Process   io → ir.IO        terminal → ir.Terminal
call → ir.Call         if → ir.If        for → ir.For{Spec, Body, Else}
while → ir.While{Cond, Body, Else}        dowhile → ir.DoWhile   infloop → ir.InfLoop
break → ir.Break       continue → ir.Continue   (інше → nil, тихо пропускається)
```

Зверни увагу:
- для `for` поле JSON зветься `Cond`, а в `ir.For` — `Spec` («змінна = початок, кінець,
  крок»). Конвертер це мапить.
- `for`/`while` тепер несуть і **`Else`** (гілка `for/else`/`while/else`) — береться з
  поля `n.Else`. → [[IR-проміжне-представлення]].
- `continue` додано разом із `break` (обидва — стрибки без фігури).
- `match` сюди не доходить окремим видом: парсер зводить його в ланцюг `if`.

## Пов'язане

- [[Чому-AST-JSON-як-контракт]]
- [[IR-проміжне-представлення]]
- [[Парсер-Python]]
- [[WASM-міст]]
