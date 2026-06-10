# rombik.

**Код → блок-схема за ДСТУ 19.701-90 (SVG / PNG / PDF / Typst / Excalidraw).**

rombik перетворює код на **Python та C++** на акуратну блок-схему алгоритму
(ГОСТ/ДСТУ 19.701-90 — той самий стандарт, що вимагають у курсових і лабораторних).
Працює **повністю у браузері**, без сервера: код нікуди не надсилається. Рушій —
чистий **TypeScript** (`rombik-engine`), тож тим самим кодом можна користуватись
і в Node-скриптах.

```python
def grade(score):
    name = input("Ваше ім'я: ")
    print("Привіт,", name)
    total = score + 5
    if total >= 90:
        print("Відмінно")
    else:
        if total >= 60:
            print("Задовільно")
        else:
            print("Незадовільно")
    print("Готово")
```

> Початок → паралелограм «Ввід score» → «Ввід name» → процес «total = score + 5»
> → ромб «total >= 90» (Так/Ні) → вкладені гілки → «Вивід «Готово»» → Кінець.

---

## Можливості

- **Точні ДСТУ-примітиви** — термінатор (овал), процес (прямокутник), розв'язок (ромб),
  ввід/вивід (паралелограм), початок циклу (шестикутник), підпрограма (прямокутник
  із боковими рисками).
- **Розпізнавання керівних конструкцій** — `if/elif/else`, `match/case`,
  `try/except/finally`, `with`, `for ... in range`, `while`, `for/else`, `while/else`,
  ідіоми `while True: … if cond: break` (післяумова) та нескінченний цикл, `break`,
  `continue`, `return/raise/exit`, виклики локальних функцій → символ підпрограми.
- **Повна підтримка C++** — `if/else`, `switch` (усі `case` + `default`, fallthrough
  `case 1: case 2:` → одна умова через `||`), `for` (класичний і range‑based `for (x : arr)`),
  `while`, `do/while`, вкладені цикли, `break`/`continue`, кілька `return`, рекурсія,
  `cin`/`cout` (ввід/вивід), `try`/`catch`, **методи класів** — кожен окремою схемою `Клас::метод`.
- **Кожна функція — окрема схема.** Параметри функції малюються вхідним паралелограмом.
- **Підпис «Рисунок N»** — конфігуроване слово/шаблон/нумерація (за ДСТУ); шрифт Times New Roman 14.
- **Налаштування під вимоги викладача** — слова вводу/виводу, підписи гілок, один
  спільний «Кінець» чи окремий на кожен вихід, зняття тип-анотацій тощо.
- **П'ять форматів виводу** — **SVG**, **PNG** і **PDF** (браузерний canvas),
  **Typst** (CeTZ — для вставки в курсову; є фрагмент-режим), **Excalidraw** (`.excalidraw`
  — доредагувати на excalidraw.com). «Завантажити всі» — у будь-якому форматі.
- **Візуальний редактор** (у браузері) — безмежне полотно з пан/зумом, перетягування
  блоків, редаговані стрілки, правка тексту/підписів, конектори, undo/redo. **Ручне
  ділення:** виділяєш блоки рамкою-ласо (Shift+тяг) чи Shift-кліком → **«⊞ Окрема схема»**
  виносить їх у свою схему (рамкою), **«⊟ Об'єднати»** — зливає схеми назад. Крос-групові
  стрілки самі стають парами конекторів. «Редагувати всі» — усі функції на одному полотні.
- **Без бекенду** — SvelteKit + Tree-sitter (WASM-граматики) + TS-рушій. Синтаксичні
  помилки показуються по-людськи, а не призводять до падіння.

---

## Запуск (браузерна версія)

Потрібен лише **Node 22+**. Монорепо на npm-workspaces.

```bash
npm install            # з кореня — лінкує rombik-engine у веб
npm run dev            # http://localhost:5173
npm run build:web      # статика у web/build/
```

**Nix:** `nix develop` дає готове середовище (Node 22 + `rsvg-convert` для PNG/PDF) —
без ручної возні з версіями.

---

## Бібліотека / скриптинг (`rombik-engine`)

Чистий TS, без DOM/фреймворку — працює у браузері, Node, будь-де. Парсинг дає
tree-sitter (Tree), далі все в рушії:

```ts
import { fromTree, renderSvg, splitFromAst } from 'rombik-engine';
// tree — результат web-tree-sitter (Python/C++)
const figs = fromTree(tree, 'python', { singleEnd: true });
for (const f of figs) writeFileSync(`${f.name}.svg`, renderSvg(f.diagram));
```

Інші точки входу: `fromAst(astJSON, opts)` (готовий AST-JSON), `parseTree(tree, lang)`
(tree → AST-JSON), `renderTypst` / `renderExcalidraw`, `splitFromAst(ast, opts, name, maxH)`.
PNG/PDF — браузерний SVG→canvas (див. `web/src/lib/engine.js`).

### CLI — пакет `rombik`

Найпростіше — через npm (потрібен лише **Node 22+**):

```bash
npx rombik examples/grade.py -o grade.svg     # без встановлення
npm i -g rombik && rombik prog.cpp -t typ     # глобально
```

Або зібрати standalone-копію вручну (бандл рушія + tree-sitter в один файл, **без
node_modules і TS-флагів** — просто `node`):

```bash
npm run build:cli                                   # → dist/rombik.mjs (+ wasm поруч)

node dist/rombik.mjs examples/grade.py -o grade.svg
./dist/rombik.mjs prog.cpp -t typ > prog.typ         # shebang; з будь-якої теки
cat prog.py | ./dist/rombik.mjs - -t json            # stdin → JSON-геометрія

# PNG / PDF — растеризуй SVG зовнішнім конвертером (rsvg-convert / cairosvg / inkscape):
./dist/rombik.mjs grade.py | rsvg-convert -f pdf -o grade.pdf
./dist/rombik.mjs grade.py | rsvg-convert -f png -o grade.png
```

> Папку `dist/` (скрипт + 3 wasm, ~4 МБ) можна скопіювати куди завгодно — працює з
> чистим Node 22+, нічого не встановлюючи. Для dev без бандла: `npm run cli -- file.py`.

Опції: `-o/--out`, `-t/--format svg|typ|json|excalidraw`, `-l/--lang py|cpp`, `--fn NAME`,
`--single-end`, `--split N`. `-h` — повна довідка. **PNG/PDF у CLI** немає (потрібен
браузерний canvas) — растеризуй SVG зовнішнім тулом, як вище; на сайті ж PNG/PDF є.

---

## Як це працює (конвеєр)

```
код (Python / C++)
   │  tree-sitter (WASM-граматика)            ← єдиний парсер: браузер і Node
   ▼
Tree → parser/treesitter  →  AST-JSON          (мова-агностик контракт)
   ▼
astjson → IR              (логічне дерево алгоритму, без геометрії)
   │  layout  (рекурсивна розкладка, шинна маршрутизація поворотів/виходів)
   ▼
Diagram   (фігури з координатами + ребра-ламані + підпис)
   │
   ├─ render/svg        → SVG        ├─ render/excalidraw → .excalidraw
   ├─ render/typst      → Typst      └─ (веб) SVG → canvas → PNG / PDF (jsPDF)
```

Кожен етап — окремий модуль із однією відповідальністю. Ядро (`ir` → `layout` →
`diagram` → рендери) не залежить ні від мови, ні від DOM, ні від формату.

---

## Структура репозиторію

```
packages/engine/        rombik-engine — TS-рушій (без DOM/фреймворку)
  src/
    parser/treesitter   tree-sitter → AST-JSON (Python + C++)
    astjson · ir        AST-JSON → IR (логічне дерево)
    layout/             розкладка IR → геометрія (серце проєкту)
    diagram             модель геометрії + підпис
    render/{svg,typst,excalidraw}
    build               fromAst / fromTree / splitFromAst (оркестрація)
  test/                 golden-парність (node:test): corpus + заморожені еталони
web/                    SvelteKit-фронтенд (CodeMirror, редактор схеми, експорт, PNG/PDF на canvas)
examples/               приклади коду
brain/                  📚 документація проєкту (Obsidian-vault)
```

---

## Розробка

```bash
npm test                              # тести рушія (golden-парність, node:test)
npm run typecheck                     # tsc --noEmit (strict)
npm run dev                           # дев-сервер вебу
```

> **Інваріант якості:** кожен модуль рушія стереже `packages/engine/test/golden` —
> заморожені байт-у-байт еталони (історично зняті з Go-рушія, з якого проєкт мігрував).

---

## Документація

У теці [`brain/`](brain/) — повний «другий мозок» проєкту: задум, архітектура, розбір
модулів, алгоритми розкладки. Відкривайте як Obsidian-vault або з [`brain/Home.md`](brain/Home.md).

---

## Залежності

- **Рушій (`rombik-engine`):** нуль рантайм-залежностей (чистий TS). Тести —
  `web-tree-sitter` (граматики) + вбудований `node:test`.
- **Веб:** SvelteKit (Svelte 5) + Tailwind v4, CodeMirror, `web-tree-sitter`, `jspdf`.
