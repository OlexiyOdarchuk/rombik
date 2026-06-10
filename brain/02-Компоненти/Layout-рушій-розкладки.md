---
tags: [component, layout, core, important]
---

# Layout — рушій розкладки

**Каталог:** `layout/` (`packages/engine/src/layout/`) · ⭐ **серце проєкту**
**Файли:** `measure.ts` (size-сімейство), `options.ts` (опції + текст), `place.ts` (клас `Builder`)

Перетворює логічний [[IR-проміжне-представлення|ir]] у геометричний
[[Diagram-модель-геометрії|diagram]]. Уся «магія» тут — і вона **не про графи**:
структурований код — це вкладене дерево, тож розкладка **рекурсивна**. Алгоритм
портовано з Go **байт-у-байт** (стережуть golden-тести).

> [!important] Дві фази
> 1. **`size`** (`measure.ts`) — знизу вгору рахує габарити піддерева (скільки місця треба).
> 2. **`place`** (`place.ts`) — згори вниз роздає координати, додаючи фігури й ребра.
>
> Деталі обох фаз — окрема нотатка: [[Розкладка-рекурсією-size-і-place]].

## measure.ts — константи (точки)

```ts
boxH=46  minBoxW=130  charW=8.5  padX=28        // блоки й текст
diaH=76  minDiaW=150                            // ромб
termW=130  termH=42                             // термінатор
hexH=50                                         // шестикутник
vGap=44  hGap=56  branchGap=48  mergeGap=44     // відступи
margin=44  arcGap=38                            // поля; винос дуг циклу
```

Ширина під текст: `textW(s) = max(runeLen(s)*charW + padX, minBoxW)`. Ромб і
шестикутник мають свої формули (`diaW`, `hexW`) із запасом на скоси. Ширина тексту —
**оцінка** через `charW` (як у Go), без вимірювання браузером, щоб розкладка була
однакова всюди (веб, Node) і трималась golden-парності — рушій **без DOM-залежностей**.

Експорти `measure.ts`: `size(n, o)`, `blockSize(b, o)`, `branchSize(b, o)`, `textW`,
`diaW`, `hexW`, `nstmts` + усі константи.

## options.ts — перемикачі рендера

```ts
export interface Options {
  callAsProcess?: boolean;  // виклик → звичайний прямокутник
  singleEnd?: boolean;      // один спільний «Кінець»
  yes?, no?: string;        // підписи гілок («Так»/«Ні»)
  inWord?, outWord?: string;// слова вводу/виводу
  startText?, endText?: string; // текст термінаторів
  stripTypes?: boolean;     // прибрати тип-анотації
  returnAsIO?: boolean;     // return — паралелограмом
  capWord?: string;         // слово-supplement підпису
  noStart?, noEnd?: boolean;// без термінатора Початок/Кінець (для частин split)
}
```

`resolveOptions(o)` заповнює ДСТУ-замовчування → `ResolvedOptions` (усі поля
обов'язкові). Усі опції необов'язкові — фронт передає лише змінені галочки. Кожна
опція детально — [[Опції-рендера]].

Текстові перетворення тут же:
- **`ioText(t, o)`** — застосовує обрані `inWord`/`outWord` («Ввід» → «Введення» тощо).
- **`procText(t, o)`** — за `stripTypes` прибирає тип-анотацію регуляркою
  `typeAnnRe` (`name: type =` → `name =`).

## place.ts — клас Builder (стан розкладки)

```ts
class Builder {
  d: Diagram = { shapes: [], edges: [], w: 0, h: 0 };  // полотно, що наповнюється
  ends: Point[] = [];          // точки-виходи → у «Кінець»
  loopBreaks: Point[][] = [];  // СТЕК збирачів break (на кожен вкладений цикл)
  loopConts: Point[][] = [];   // СТЕК збирачів continue
  o: ResolvedOptions;
}
```

- **`ends`** — куди звести всі виходи функції. → [[Зведення-виходів-у-Кінець]].
- **`loopBreaks`/`loopConts`** — стеки: `pushLoop`/`popLoop`/`recordBreak`/`recordContinue`.
  Кожен `break`/`continue` пишеться у вершину стека (поточний цикл).

> [!warning] shiftX і аліасинг точок
> У TS `Point` — об'єкт (посилання), а в Go — значення-копія. Спільна точка може бути
> і в `ends`, і кінцем ребра `routeEnds`. Тому `shiftX(dx)` зсуває **кожен об'єкт рівно
> раз** (через `Set<Point>`), інакше спільна точка зсунеться двічі → діагональ.

## layoutProgram — головна функція

```ts
export function layoutProgram(prog: Block, o: ResolvedOptions): Diagram
```

(внутрішньо створює `Builder` і кличе `run`):

1. Якщо `callAsProcess` — `mapCalls` рекурсивно замінює `Call` → `Process`.
2. `blockSize(prog)` → ширина полотна `w = bw + 2*margin`, центр `cx = w/2`.
3. Якщо не `noStart` — ставить термінатор «Початок», ребро вниз.
4. `placeBlock(prog, cx, bodyTop)` → розкладає тіло, повертає точку виходу й `ended`.
5. Якщо тіло не завершилось `return`'ом — додає природний вихід у `ends`.
6. Якщо не `noEnd` і `ends` непорожній — `routeEnds` + термінатор «Кінець».
7. `bodyExtent` + `shiftX` вирівнюють по лівому полю; `d.w`/`d.h` = розмір полотна.

## Карта методів place

| Метод | Вузол | Нотатка-деталь |
|-------|-------|----------------|
| `placeBlock` | `Block` | послідовність + ребра між блоками |
| `leaf` | Process/IO/Call | одна фігура |
| `placeTerminal` | `Terminal` | return/raise → у «Кінець» |
| `placeIf` / `guard` / `guardTerm` / `branch` | `If` | [[Розкладка-if-guard-і-симетрія]] |
| `placeFor` / `placeWhile` / `placeDoWhile` / `placeInfLoop` | цикли | [[Розкладка-циклів]] |
| `loopArcs` / `placeLoopGuard` / `placeLoopElse` | — | дуги повернення/виходу циклу, for/while-else |
| `routeEnds` / `routeBreaks` | — | [[Зведення-виходів-у-Кінець]] |

## Маршрутизація ребер

`routeEnds` зводить виходи у «Кінець» **простими ортогональними ламаними** (прямо
вниз, якщо вихід на осі `cx`; інакше через проміжну точку на рівні `mergeGap/2`).
Решта ребер також — прості ортогональні ламані, побудовані прямо в `place*`-методах.
(Окремого A*-маршрутизатора більше немає — він не знадобився після байт-у-байт порту.)

## Пов'язане

- [[Розкладка-рекурсією-size-і-place]] ⭐ як працюють size/place
- [[Розкладка-if-guard-і-симетрія]]
- [[Розкладка-циклів]]
- [[Зведення-виходів-у-Кінець]]
- [[Diagram-модель-геометрії]]
