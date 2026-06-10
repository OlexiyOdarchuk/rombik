---
tags: [component, ir, core]
---

# IR — проміжне представлення

**Модуль:** `ir.ts` (`packages/engine/src/ir.ts`)

Логічне (мова-агностик) представлення алгоритму як **структури керування**. Сюди
парсер зводить будь-який код; звідси [[Layout-рушій-розкладки|layout]] робить
геометрію. **Жодного рендера й жодних координат** — лише структура.

## Дискримінований union Node

```ts
export type Node =
  | Process | IO | Call | Terminal
  | If | For | While | DoWhile | InfLoop
  | Break | Continue | Connector | Block;
```

Кожен варіант має літеральне поле `kind` — це **discriminated union** (sum type
«по-TS»). Розкладка робить `switch (n.kind)`, і TypeScript звужує тип у кожній гілці.

## Вузли-листя (текст → одна фігура)

| Тип | Поля | Фігура ДСТУ | Сенс |
|-----|------|-------------|------|
| `Process` | `text` | прямокутник | дія/обчислення |
| `IO` | `text` | паралелограм | ввід/вивід |
| `Terminal` | `text` | прямокутник → веде в «Кінець» | `return`/`raise`/`exit` |
| `Call` | `text` | прямокутник із рисками | виклик функції з цього ж файлу |
| `Break` | — | — (без фігури) | вихід із циклу |
| `Continue` | — | — (без фігури) | стрибок на заголовок циклу (наступна ітерація) |

> `Terminal` малюється звичайним прямокутником (або паралелограмом за опцією
> `returnAsIO`), але **завершує гілку** і веде у термінатор «Кінець». →
> [[Зведення-виходів-у-Кінець]].

## Керівні вузли

```ts
interface If      { kind: 'if'; cond: string; then: Block; else: Block; }   // ромб + дві гілки
interface For     { kind: 'for'; spec: string; body: Block; else: Block; }  // шестикутник + тіло (+ for/else)
interface While   { kind: 'while'; cond: string; body: Block; else: Block; }// передумова: ромб згори (+ while/else)
interface DoWhile { kind: 'dowhile'; cond: string; body: Block; }           // післяумова: ромб знизу
interface InfLoop { kind: 'infloop'; body: Block; }                         // while True без умови
interface Connector { kind: 'connector'; text: string; }                    // з'єднувач «А» (розбиття схеми)
```

- **`For.spec`** — готовий підпис «i = 0, n-1, 1» (його будує парсер, не layout).
- **`For.else`/`While.else`** — гілка `for/else`/`while/else`: виконується після
  **нормального** завершення циклу (без `break`); порожній `Block` — якщо її нема.
- **`DoWhile.cond`** — умова `break` з ідіоми `while True: … if cond: break` (або
  C++ `do…while`).
- **`InfLoop`** — `while True` із break-ами десь усередині; виходить лише через `break`.
- **`Connector`** — з'єднувач для розбиття довгої схеми на частини. → [[Публічний-API-rombik|splitByHeight]].

> [!note] Гілки завжди присутні
> `then`/`else`/`body` — завжди `Block` (порожній, якщо гілки нема), на відміну від
> Go `*Block` (nil). Це прибирає null-перевірки в layout.

> `Break` і `Continue` — стрибки без фігури: `Break` веде на вихід циклу, `Continue` —
> на заголовок (дуга повторної ітерації). → [[Розкладка-циклів]].

Як кожен із них розкладається — [[Розкладка-циклів]], [[Розкладка-if-guard-і-симетрія]].

## Контейнери

```ts
interface Block { kind: 'block'; stmts: Node[]; }     // послідовність (згори вниз)
interface Func  { name: string; body: Block; }        // іменована програма → окрема схема
```

`Func` — одна функція (або тіло модуля). Файл із кількома `def` дає кілька `Func`,
кожна йде в окрему діаграму.

## Чому IR окремо від AST і від Diagram

- **Від AST** (`AstNode`) — бо AST «на дроті» плаский і нетипізований; IR —
  типобезпечний union із літеральними `kind`. → [[astjson-конвертер]].
- **Від Diagram** — бо IR не має координат; це чиста логіка «що за чим іде». Геометрію
  додає [[Layout-рушій-розкладки|layout]]. Одне IR теоретично можна розкласти
  по-різному.

## Пов'язане

- [[astjson-конвертер]]
- [[Layout-рушій-розкладки]]
- [[Розкладка-рекурсією-size-і-place]]
