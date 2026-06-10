---
tags: [component, parser]
---

# Парсер Python

**Модуль:** `parser/treesitter.ts` (`packages/engine/src/parser/treesitter.ts`)

Перетворює дерево **tree-sitter** на [[Чому-AST-JSON-як-контракт|AST-JSON]]. Один
фронтенд для **двох мов одразу** — Python І C++. Раніше було два окремі парсери (CLI на
`python3 ast`+`parser.py` і браузерний `parser.js`) — **обидва видалено**; після
міграції Go→TS лишився **єдиний** парсер на tree-sitter.

> Назва файлу нотатки лишилася «Парсер-Python», але контент — про спільний
> tree-sitter-парсер Python+C++.

## Експорт

```ts
export function parseTree(tree: TSTree, lang: Lang): AstFunc[]
export interface TSNode { type; text; isNamed; childCount; namedChildCount;
                          child(i); childForFieldName(name); namedChildren }
export interface TSTree { rootNode: TSNode }
export type Lang = 'python' | 'cpp'
```

- `parseTree(tree, lang)` — головна точка входу: дерево tree-sitter + мова → масив
  схем `AstFunc[]`.
- `TSNode`/`TSTree` — **мінімальний структурний інтерфейс** вузла tree-sitter (сумісний
  із `web-tree-sitter` Node), щоб рушій не залежав від конкретної версії пакета й не
  мав DOM-залежностей.
- `lang` керує мово-специфічними гілками: `isPy`/`isCpp`.

> [!warning] Безпека
> tree-sitter лише **розбирає** синтаксис — код НЕ виконується. Тому безпечно
> годувати парсер довільним публічним вводом.

## Збір функцій

Перед розбором тіл — два проходи:

- `collectDefinedFunctions(root, defined)` — рекурсивно збирає імена всіх
  `function_definition` (для C++ розкручує `declarator` до `function_declarator`).
  `defined` потрібен, щоб відрізнити **виклик власної функції** (→ підпрограма) від
  звичайного виразу.
- `collect(node)` — обходить корінь: кожен `function_definition` → окремий `AstFunc`;
  усе поза функціями збирається в `mainStmts`.

Якщо є код поза функціями — він іде окремою схемою з іменем `main` (або `програма`,
якщо `main` уже зайнятий). Якщо схем нема взагалі — повертається порожня `main`.

### Параметри як ввід

```ts
if (paramsText) b.stmts.unshift({ kind: 'io', text: 'Ввід ' + oneline(paramsText) });
```

Параметри функції показуємо вхідним паралелограмом (конвенція курсової схеми). Імена
параметрів витягуються по-різному для Python (`parameters` → identifier/`name`) і C++
(`function_declarator` → `parameters` → `declarator`).

## Відображення конструкцій

Повна таблиця — у [[Підтримувані-конструкції-Python]]. Ключові рішення `stmt()`:

| Конструкція | AST-JSON kind | Нюанс |
|--------|---------------|-------|
| `input(...)` / C++ `cin >>` як значення присвоєння/декларації | `io` «Ввід …» | `hasInputNode` шукає `input`/`cin` будь-де у виразі |
| `print(...)` / `cout << …` | `io` «Вивід …» | рядки → «…», порожній `print()` → «Вивід порожнього рядка» |
| `for i in range(a,b,c)` (Py) | `for` | spec: «i = a, b-1, c»; `endof` робить `stop-1` (range напіввідкритий) |
| C++ `for(init; cond; upd)` | `for` | spec складається з init/cond/update |
| `while True: … if C: break` (останнім) | `dowhile` | післяумова; `cond` = умова break (`breakIfNode`) |
| `while True:` (break десь усередині) | `infloop` | без ромба-умови |
| `while C:` / `for …:` | `while`/`for` | + поле `else` (гілка `for/else`, `while/else`, `unwrapElse`) |
| C++ `do { … } while(C)` | `dowhile` | післяумова напряму |
| `break` / `continue` | `break`/`continue` | стрибок без фігури |
| `match/case` (Py) / `switch/case` (C++) | `if` (ланцюг) | патерни/мітки → каскад `if` (`buildCascade`) |
| `return v` / `raise`/`throw` / `exit()` | `terminal` | «Повернути v» / «Помилка: e» / «Вихід» |
| `x = моя_функція()` | `call` | лише якщо `defined.has(funcName)` |
| `with ... :` | `block` | «відкрити: …» + тіло |
| `try/except` (+ `catch`/`finally`) | `block` | тіло try + обробники як вкладені `if «Виняток?»` |

## Дрібниці форматування

- **`oneline` + `MAXLEN=64`** — будь-який текст у блоці зводиться в один рядок і
  обрізається «…», щоб не ламати геометрію.
- **`formatArg`** — рядкові аргументи беруться в лапки «…»; f-рядки лишаються читабельні.
- **`endof`** — верхня межа `range`: для літерала рахує `n-1`, для `X+1` дає `X`,
  інакше `expr - 1`.
- **C++ `cout`/`cin`** — регулярками розбираються ланцюжки `<<`/`>>` у читабельний
  текст вводу/виводу.

## Що ігнорується

`stmt()` повертає `null` (фігура не малюється) для докстрінгів/рядкових виразів,
`comment`, `import`/`import_from`/`preproc_include`/`preproc_def`, а також вкладених
`function_definition`/`class_definition` — вкладені `def` йдуть **окремими схемами**
(через рекурсивний `collect`), класи пропускаються як зайвий шум.

## Пов'язане

- [[Чому-AST-JSON-як-контракт]]
- [[astjson-конвертер]]
- [[Підтримувані-конструкції-Python]]
- [[Браузерний-рушій]]
