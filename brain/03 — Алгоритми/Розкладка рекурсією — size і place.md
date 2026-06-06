---
tags: [algorithm, layout, important]
---

# Розкладка рекурсією — size і place

Як вкладене дерево [[IR — проміжне представлення|IR]] стає координатами. Це
фундамент усього [[Layout — рушій розкладки|layout]].

> [!important] Головна ідея
> Структурований код — це **дерево**, а не довільний граф. Тож не потрібен складний
> graph-layout (як у Graphviz). Достатньо двох рекурсивних обходів:
> **`size` рахує габарити знизу вгору, `place` роздає координати згори вниз.**

## Фаза 1: size (знизу вгору)

`size(n) (w, h)` повертає габарити прямокутника, у який впишеться все піддерево вузла.
Рекурсія: габарити складеного вузла рахуються з габаритів дітей.

```go
func (b *build) size(n ir.Node) (w, h float64) {
    switch x := n.(type) {
    case *ir.Process: return textW(...), boxH        // листя: ширина під текст
    case *ir.If:                                      // складене: з дітей
        tw, th := b.branchSize(x.Then)
        ew, eh := b.branchSize(x.Else)
        h := diaH + branchGap + max(th, eh) + mergeGap
        ...
    case *ir.For:
        bw, bh := b.blockSize(x.Body)
        return max(hexW(x.Spec), bw) + 2*arcGap, hexH + vGap + bh + vGap
    }
}
```

`blockSize(blk)` для послідовності — **сума висот** дітей + `vGap` між ними, ширина =
максимум ширин:

```go
for i, s := range blk.Stmts {
    sw, sh := b.size(s)
    w = max(w, sw)
    h += sh
    if i > 0 { h += vGap }
}
```

Навіщо рахувати габарити наперед: щоб **відцентрувати** гілки if і винести дуги
циклів, треба знати ширину піддерева ще ДО того, як ставити його координати.

### Тонкість: порожні гілки

`branchSize` повертає `(0,0)` для порожньої гілки — щоб вона **не резервувала
ширину** і не зміщувала протилежну гілку вбік. Це вмикає guard-розкладку
([[Розкладка if — guard і симетрія]]).

## Фаза 2: place (згори вниз)

`place(n, cx, top) (exit, ended)` малює вузол так, щоб його **верх-центр** був у
`(cx, top)`. Повертає:

- **`exit`** — точку **низ-центр**, звідки піде наступне ребро;
- **`ended`** — чи гілка завершилась (`return/raise/exit/break`); якщо так — решта
  інструкцій недосяжна, далі не малюємо.

```go
func (b *build) place(n ir.Node, cx, top float64) (diagram.Point, bool) {
    switch x := n.(type) {
    case *ir.Process: return b.leaf(diagram.Process, ..., cx, top), false
    case *ir.If:      return b.placeIf(x, cx, top)
    case *ir.For:     return b.placeFor(x, cx, top), false
    ...
    }
}
```

### placeBlock — послідовність

```go
func (b *build) placeBlock(blk *ir.Block, cx, top float64) (Point, bool) {
    cur := top
    for i, s := range blk.Stmts {
        if i > 0 { /* ребро від попереднього exit до cur */ }
        exit, ended = b.place(s, cx, cur)
        if ended { return exit, true }   // далі недосяжно
        cur = exit.Y + vGap
    }
    return exit, false
}
```

Кожен наступний блок ставиться на `exit.Y + vGap` нижче, з'єднується ребром. Усе
вирівняно по осі `cx`.

## Чому верх-центр як точка прив'язки

Уся схема **вертикальна й симетрична** відносно осі `cx`. Прив'язка за центром зверху
робить додавання блоків тривіальним: «став наступний на тій самій осі, нижче». Гілки й
дуги розходяться симетрично вліво/вправо від `cx`.

## Дві фази на прикладі циклу for

1. **size:** `placeFor` спершу хоче знати ширину тіла (`bw`), щоб винести дуги на
   `arcGap` назовні → `size(For)` = `max(hexW, bw) + 2*arcGap`.
2. **place:** ставить шестикутник у `cx`, тіло під ним, тоді `loopArcs` малює дугу
   повернення праворуч (на `bw/2 + arcGap`) і вихід ліворуч.

→ [[Розкладка циклів]].

## Пов'язане

- [[Layout — рушій розкладки]]
- [[Розкладка if — guard і симетрія]]
- [[Розкладка циклів]]
- [[Зведення виходів у Кінець]]
