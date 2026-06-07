---
tags: [component, layout, core, important]
---

# Layout — рушій розкладки

**Пакет:** `internal/layout` · **Файл:** `layout.go` · ⭐ **серце проєкту**

Перетворює логічний [[IR - проміжне представлення|ir]] у геометричний
[[Diagram - модель геометрії|diagram]]. Уся «магія» тут — і вона **не про графи**:
структурований код — це вкладене дерево, тож розкладка **рекурсивна**.

> [!important] Дві фази
> 1. **`size`** — знизу вгору рахує габарити піддерева (скільки місця треба).
> 2. **`place`** — згори вниз роздає координати, додаючи фігури й ребра.
>
> Деталі обох фаз — окрема нотатка: [[Розкладка рекурсією - size і place]].

## Константи (точки)

```
boxH=46  minBoxW=130  charW=8.5  padX=28        // блоки й текст
diaH=76  minDiaW=150                            // ромб
termW=130  termH=42                             // термінатор
hexH=50                                         // шестикутник
vGap=44  hGap=56  branchGap=48  mergeGap=44     // відступи
margin=44  arcGap=38                            // поля; винос дуг циклу
```

Ширина під текст: `textW(s) = len(s)*charW + padX`, не менше `minBoxW`. Ромб і
шестикутник мають свої формули (`diaW`, `hexW`) із запасом на скоси.

## Options — перемикачі рендера

```go
type Options struct {
    CallAsProcess bool   // виклик → звичайний прямокутник
    SingleEnd     bool   // один спільний «Кінець»
    Yes, No       string // підписи гілок («Так»/«Ні»)
    InWord, OutWord string // слова вводу/виводу
    StartText, EndText string // текст термінаторів
    StripTypes    bool   // прибрати тип-анотації
    ReturnAsIO    bool   // return — паралелограмом
}
```

Серіалізується в/з JSON (`json:"callAsProcess"` тощо) — так фронтенд передає галочки
у WASM. Кожна опція детально — [[Опції рендера]].

## build — стан розкладки

```go
type build struct {
    d          *diagram.Diagram   // полотно, що наповнюється
    ends       []diagram.Point    // точки-виходи → у «Кінець»
    loopBreaks [][]diagram.Point  // СТЕК збирачів break (на кожен вкладений цикл)
    singleEnd  bool
    yes, no, inWord, outWord, endText string
    stripTypes, retAsIO bool
}
```

- **`ends`** — куди звести всі виходи функції. → [[Зведення виходів у Кінець]].
- **`loopBreaks`** — стек: `pushLoop`/`popLoop`/`recordBreak`. Кожен `break` пишеться
  у вершину стека (поточний цикл), `routeBreaks` зводить їх до виходу циклу.

## Build — головна функція

```go
func Build(prog *ir.Block, opts Options) *diagram.Diagram
```

1. Якщо `CallAsProcess` — `mapCalls` рекурсивно замінює `ir.Call` → `ir.Process`.
2. Дефолти опцій («Так»/«Ні»/«Ввід»/«Вивід»/«Початок»/«Кінець»).
3. `blockSize(prog)` → ширина полотна `w = bw + 2*margin`, центр `cx = w/2`.
4. Ставить термінатор «Початок», ребро вниз.
5. `placeBlock(prog, cx, bodyTop)` → розкладає тіло, повертає точку виходу й `ended`.
6. Якщо тіло не завершилось `return`'ом — додає природний вихід у `ends`.
7. Якщо `ends` непорожній — `routeEnds` + термінатор «Кінець».
8. `W`, `H` = розмір полотна за `contentBottom`.

## Допоміжні текстові перетворення

- **`ioText`** — застосовує обрані `inWord`/`outWord` («Ввід» → «Введення» тощо).
- **`procText`** — за `StripTypes` прибирає тип-анотацію регуляркою
  `typeAnnRe` (`name: type =` → `name =`).

## Карта методів place

| Метод | Вузол | Нотатка-деталь |
|-------|-------|----------------|
| `placeBlock` | `Block` | послідовність + ребра між блоками |
| `leaf` | Process/IO/Call | одна фігура |
| `placeTerminal` | `Terminal` | return/raise → у «Кінець» |
| `placeIf` / `guard` / `branch` | `If` | [[Розкладка if - guard і симетрія]] |
| `placeFor` / `placeWhile` / `placeDoWhile` / `placeInfLoop` | цикли | [[Розкладка циклів]] |
| `loopArcs` | — | дуги повернення/виходу циклу |
| `routeEnds` / `routeBreaks` | — | [[Зведення виходів у Кінець]] |

## Маршрутизація ребер

`routeEnds` зводить виходи у «Кінець» через [[Route - маршрутизатор ребер|route.Route]]
(A* з обходом фігур). Решта ребер — прості ортогональні ламані, побудовані прямо в
`place*`-методах (не через A*). → [[A-зірка маршрутизація]].

## Пов'язане

- [[Розкладка рекурсією - size і place]] ⭐ як працюють size/place
- [[Розкладка if - guard і симетрія]]
- [[Розкладка циклів]]
- [[Зведення виходів у Кінець]]
- [[Diagram - модель геометрії]]
