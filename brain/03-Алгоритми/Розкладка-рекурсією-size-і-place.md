---
tags: [algorithm, layout, important]
---

# Розкладка рекурсією — size і place

Як вкладене дерево [[IR-проміжне-представлення|IR]] стає координатами. Це
фундамент усього [[Layout-рушій-розкладки|layout]] (`packages/engine/src/layout/`).

> [!important] Головна ідея
> Структурований код — це **дерево**, а не довільний граф. Тож не потрібен складний
> graph-layout (як у Graphviz). Достатньо двох рекурсивних обходів:
> **`size` рахує габарити знизу вгору, `place` роздає координати згори вниз.**

## Фаза 1: size (знизу вгору)

`size(n, o): [w, h]` (модуль `layout/measure.ts`) повертає габарити прямокутника, у
який впишеться все піддерево вузла. Рекурсія: габарити складеного вузла рахуються з
габаритів дітей.

```ts
export function size(n: Node, o: Options): [number, number] {
  switch (n.kind) {
    case 'process': return [textW(procText(n.text, o)), boxH]; // листя: ширина під текст
    case 'if': {                                               // складене: з дітей
      const [tw, th] = branchSize(n.then, o);
      const [ew, eh] = branchSize(n.else, o);
      const h = diaH + branchGap + Math.max(th, eh) + mergeGap;
      // ...
    }
    case 'for': {
      const [bw, bh] = blockSize(n.body, o);
      return withElse(n.else, Math.max(hexW(n.spec), bw) + 2 * arcGap,
                      hexH + vGap + bh + vGap, o);
    }
  }
}
```

`blockSize(blk, o)` для послідовності — **сума висот** дітей + `vGap` між ними, ширина =
максимум ширин:

```ts
b.stmts.forEach((s, i) => {
  const [sw, sh] = size(s, o);
  w = Math.max(w, sw);
  h += sh;
  if (i > 0) h += vGap;
});
```

Навіщо рахувати габарити наперед: щоб **відцентрувати** гілки if і винести дуги
циклів, треба знати ширину піддерева ще ДО того, як ставити його координати.

### Тонкість: порожні гілки

`branchSize` повертає `[0, 0]` для порожньої гілки — щоб вона **не резервувала
ширину** і не зміщувала протилежну гілку вбік. Це вмикає guard-розкладку
([[Розкладка-if-guard-і-симетрія]]).

## Фаза 2: place (згори вниз)

`place(n, cx, top): [exit, ended]` (модуль `layout/place.ts`, клас `Builder`) малює
вузол так, щоб його **верх-центр** був у `(cx, top)`. Повертає:

- **`exit`** — точку **низ-центр**, звідки піде наступне ребро;
- **`ended`** — чи гілка завершилась (`return/raise/exit/break`); якщо так — решта
  інструкцій недосяжна, далі не малюємо.

```ts
place(n: Node, cx: number, top: number): [Point, boolean] {
  switch (n.kind) {
    case 'process': return [this.leaf('process', procText(n.text, this.o), cx, top), false];
    case 'if':      return this.placeIf(n, cx, top);
    case 'for':     return this.placeFor(n, cx, top);
    // ...
  }
}
```

### placeBlock — послідовність

```ts
placeBlock(blk: Block, cx: number, top: number): [Point, boolean] {
  let cur = top, exit = P(cx, top);
  for (let i = 0; i < blk.stmts.length; i++) {
    const s = blk.stmts[i];
    if (i > 0) this.d.edges.push(edge(exit, P(cx, cur))); // ребро від попереднього exit
    const [e, ended] = this.place(s, cx, cur);
    exit = e;
    if (ended) return [exit, true];   // далі недосяжно
    cur = exit.y + vGap;
  }
  return [exit, false];
}
```

Кожен наступний блок ставиться на `exit.y + vGap` нижче, з'єднується ребром. Усе
вирівняно по осі `cx`.

## Чому верх-центр як точка прив'язки

Уся схема **вертикальна й симетрична** відносно осі `cx`. Прив'язка за центром зверху
робить додавання блоків тривіальним: «став наступний на тій самій осі, нижче». Гілки й
дуги розходяться симетрично вліво/вправо від `cx`.

## Дві фази на прикладі циклу for

1. **size:** `placeFor` спершу хоче знати ширину тіла (`bw`), щоб винести дуги на
   `arcGap` назовні → `size(For)` = `max(hexW, bw) + 2*arcGap`.
2. **place:** ставить шестикутник у `cx`, тіло під ним, тоді `loopArcs` малює дугу
   повернення праворуч і вихід ліворуч.

→ [[Розкладка-циклів]].

> [!note] TS-специфіка точок
> На відміну від Go (де `diagram.Point` — значення-копія), у TS точки — це
> об'єкти-посилання, які аліасяться між ребрами та `ends`. Тому `shiftX` (фінальне
> вирівнювання схеми по `margin`) зсуває кожен об'єкт рівно раз через `Set<Point>` —
> інакше спільна точка зсунулась би двічі й дала б діагональ.

## Пов'язане

- [[Layout-рушій-розкладки]]
- [[Розкладка-if-guard-і-симетрія]]
- [[Розкладка-циклів]]
- [[Зведення-виходів-у-Кінець]]
</content>
</invoke>
