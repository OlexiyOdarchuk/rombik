// Пакет layout перетворює логічний ir у геометричний diagram. Уся «магія» тут,
// і вона НЕ про графи: структурований код — це вкладене дерево, тож розкладка
// рекурсивна. size() рахує габарити піддерева, place() роздає координати.
package layout

import (
	"flowgen/internal/diagram"
	"flowgen/internal/ir"
)

// Розміри й відступи (у точках).
const (
	boxH      = 46  // висота прямокутника/паралелограма
	minBoxW   = 130 // мінімальна ширина блоку
	charW     = 8.5 // приблизна ширина символу (шрифт ~14px)
	padX      = 28  // горизонтальні поля тексту в блоці
	diaH      = 76  // висота ромба
	minDiaW   = 150 // мінімальна ширина ромба
	termW     = 130 // ширина термінатора (початок/кінець)
	termH     = 42
	vGap      = 44 // вертикальний відступ між послідовними блоками
	branchGap = 48 // від ромба до верху гілок
	hGap      = 56 // горизонтальний відступ між гілками if
	mergeGap  = 44 // від низу гілок до точки злиття
	margin    = 44 // зовнішні поля діаграми
	hexH      = 50 // висота шестикутника «початок циклу»
	arcGap    = 38 // горизонтальний винос дуг циклу за межі тіла
)

// textW — приблизна ширина блоку під текст.
func textW(s string) float64 {
	w := float64(len([]rune(s)))*charW + padX
	return max(w, minBoxW)
}

// diaW — ширина ромба під умову (ромбу треба запас, бо текст у середній смузі).
func diaW(cond string) float64 {
	return max(minDiaW, float64(len([]rune(cond)))*charW+60)
}

// Options — перемикачі рендера (галочки в інтерфейсі). Серіалізується в/з JSON,
// тож фронтенд передає їх як об'єкт.
type Options struct {
	// CallAsProcess: виклик підпрограми малювати звичайним прямокутником
	// (не ДСТУ-символом «підпрограма») — на вимогу деяких викладачів.
	CallAsProcess bool `json:"callAsProcess"`
}

// Build розкладає програму (тіло) у повну діаграму: Початок → тіло → Кінець.
func Build(prog *ir.Block, opts Options) *diagram.Diagram {
	if opts.CallAsProcess {
		mapCalls(prog)
	}
	bw, _ := blockSize(prog)
	w := bw + 2*margin
	cx := w / 2

	d := &diagram.Diagram{}
	top := float64(margin)

	// Початок.
	d.Shapes = append(d.Shapes, term(cx, top, "Початок"))
	exit := diagram.Point{X: cx, Y: top + termH}

	// Тіло.
	bodyTop := exit.Y + vGap
	d.Edges = append(d.Edges, edge(exit, diagram.Point{X: cx, Y: bodyTop}))
	exit = placeBlock(d, prog, cx, bodyTop)

	// Кінець.
	endTop := exit.Y + vGap
	d.Edges = append(d.Edges, edge(exit, diagram.Point{X: cx, Y: endTop}))
	d.Shapes = append(d.Shapes, term(cx, endTop, "Кінець"))

	d.W = w
	d.H = endTop + termH + margin
	return d
}

// mapCalls рекурсивно замінює виклики підпрограм на звичайні процеси (опція).
func mapCalls(b *ir.Block) {
	if b == nil {
		return
	}
	for i, n := range b.Stmts {
		switch x := n.(type) {
		case *ir.Call:
			b.Stmts[i] = &ir.Process{Text: x.Text}
		case *ir.If:
			mapCalls(x.Then)
			mapCalls(x.Else)
		case *ir.For:
			mapCalls(x.Body)
		case *ir.While:
			mapCalls(x.Body)
		case *ir.DoWhile:
			mapCalls(x.Body)
		}
	}
}

// --- розмір (габарити піддерева) ---

func size(n ir.Node) (w, h float64) {
	switch x := n.(type) {
	case *ir.Process:
		return textW(x.Text), boxH
	case *ir.IO:
		return textW(x.Text), boxH
	case *ir.Call:
		return textW(x.Text), boxH
	case *ir.Block:
		return blockSize(x)
	case *ir.If:
		tw, th := blockSize(x.Then)
		ew, eh := blockSize(x.Else)
		return max(diaW(x.Cond), tw+hGap+ew), diaH + branchGap + max(th, eh) + mergeGap
	case *ir.For:
		bw, bh := blockSize(x.Body)
		return max(hexW(x.Spec), bw) + 2*arcGap, hexH + vGap + bh + vGap
	case *ir.While:
		bw, bh := blockSize(x.Body)
		return max(diaW(x.Cond), bw) + 2*arcGap, diaH + vGap + bh + vGap
	case *ir.DoWhile:
		bw, bh := blockSize(x.Body)
		return max(diaW(x.Cond), bw) + 2*arcGap, bh + vGap + diaH + vGap
	}
	return minBoxW, boxH
}

// hexW — ширина шестикутника під підпис (із запасом на скоси з боків).
func hexW(spec string) float64 { return max(minBoxW, textW(spec)+hexH) }

func blockSize(b *ir.Block) (w, h float64) {
	if b == nil || len(b.Stmts) == 0 {
		return minBoxW, 0
	}
	for i, s := range b.Stmts {
		sw, sh := size(s)
		w = max(w, sw)
		h += sh
		if i > 0 {
			h += vGap
		}
	}
	return w, h
}

// --- розкладка (координати) ---
// place малює вузол з верхом-центром у (cx, top); повертає точку виходу (низ-центр).

func place(d *diagram.Diagram, n ir.Node, cx, top float64) diagram.Point {
	switch x := n.(type) {
	case *ir.Process:
		return leaf(d, diagram.Process, x.Text, cx, top)
	case *ir.IO:
		return leaf(d, diagram.InOut, x.Text, cx, top)
	case *ir.Call:
		return leaf(d, diagram.Predef, x.Text, cx, top)
	case *ir.Block:
		return placeBlock(d, x, cx, top)
	case *ir.If:
		return placeIf(d, x, cx, top)
	case *ir.For:
		return placeFor(d, x, cx, top)
	case *ir.While:
		return placeWhile(d, x, cx, top)
	case *ir.DoWhile:
		return placeDoWhile(d, x, cx, top)
	}
	return diagram.Point{X: cx, Y: top}
}

func leaf(d *diagram.Diagram, kind diagram.Kind, text string, cx, top float64) diagram.Point {
	w := textW(text)
	d.Shapes = append(d.Shapes, diagram.Shape{Kind: kind, X: cx - w/2, Y: top, W: w, H: boxH, Text: text})
	return diagram.Point{X: cx, Y: top + boxH}
}

func placeBlock(d *diagram.Diagram, b *ir.Block, cx, top float64) diagram.Point {
	cur := top
	exit := diagram.Point{X: cx, Y: top}
	for i, s := range b.Stmts {
		if i > 0 {
			d.Edges = append(d.Edges, edge(exit, diagram.Point{X: cx, Y: cur}))
		}
		exit = place(d, s, cx, cur)
		cur = exit.Y + vGap
	}
	return exit
}

func placeIf(d *diagram.Diagram, n *ir.If, cx, top float64) diagram.Point {
	dw := diaW(n.Cond)
	d.Shapes = append(d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: top, W: dw, H: diaH, Text: n.Cond})
	midY := top + diaH/2
	diaBottom := top + diaH

	tw, th := blockSize(n.Then)
	ew, eh := blockSize(n.Else)
	total := tw + hGap + ew
	thenCx := cx - total/2 + tw/2
	elseCx := cx + total/2 - ew/2
	branchTop := diaBottom + branchGap

	// Стрілки від ромба до верху гілок (ортогонально), з підписами Так/Ні.
	d.Edges = append(d.Edges, diagram.Edge{
		Points: []diagram.Point{{X: cx - dw/2, Y: midY}, {X: thenCx, Y: midY}, {X: thenCx, Y: branchTop}},
		Label:  "Так",
	})
	d.Edges = append(d.Edges, diagram.Edge{
		Points: []diagram.Point{{X: cx + dw/2, Y: midY}, {X: elseCx, Y: midY}, {X: elseCx, Y: branchTop}},
		Label:  "Ні",
	})

	thenExit := placeBlock(d, n.Then, thenCx, branchTop)
	elseExit := placeBlock(d, n.Else, elseCx, branchTop)

	// Точка злиття гілок — по центру, нижче найвищої гілки.
	mergeY := branchTop + max(th, eh) + mergeGap
	d.Edges = append(d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{thenExit, {X: thenExit.X, Y: mergeY}, {X: cx, Y: mergeY}}})
	d.Edges = append(d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{elseExit, {X: elseExit.X, Y: mergeY}, {X: cx, Y: mergeY}}})
	return diagram.Point{X: cx, Y: mergeY}
}

// placeFor — цикл for: шестикутник згори, тіло під ним, дуга повернення справа,
// вихід зліва вниз (як у курсових схемах на fletcher).
func placeFor(d *diagram.Diagram, n *ir.For, cx, top float64) diagram.Point {
	hw := hexW(n.Spec)
	d.Shapes = append(d.Shapes, diagram.Shape{Kind: diagram.Hexagon, X: cx - hw/2, Y: top, W: hw, H: hexH, Text: n.Spec})
	headCy := top + hexH/2

	bw, _ := blockSize(n.Body)
	bodyTop := top + hexH + vGap
	d.Edges = append(d.Edges, edge(diagram.Point{X: cx, Y: top + hexH}, diagram.Point{X: cx, Y: bodyTop}))
	bodyExit := placeBlock(d, n.Body, cx, bodyTop)
	return loopArcs(d, cx, hw/2, headCy, bw/2, bodyExit.Y, "")
}

// placeWhile — цикл while: ромб-передумова, Так→тіло, дуга повернення справа,
// Ні→вихід зліва вниз.
func placeWhile(d *diagram.Diagram, n *ir.While, cx, top float64) diagram.Point {
	dw := diaW(n.Cond)
	d.Shapes = append(d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: top, W: dw, H: diaH, Text: n.Cond})
	headCy := top + diaH/2

	bw, _ := blockSize(n.Body)
	bodyTop := top + diaH + vGap
	d.Edges = append(d.Edges, diagram.Edge{Label: "Так", Points: []diagram.Point{{X: cx, Y: top + diaH}, {X: cx, Y: bodyTop}}})
	bodyExit := placeBlock(d, n.Body, cx, bodyTop)
	return loopArcs(d, cx, dw/2, headCy, bw/2, bodyExit.Y, "Ні")
}

// placeDoWhile — цикл з післяумовою: тіло згори, ромб-умова знизу, Так→вихід,
// Ні→дуга повернення справа до лінії входу (повтор тіла).
func placeDoWhile(d *diagram.Diagram, n *ir.DoWhile, cx, top float64) diagram.Point {
	bw, _ := blockSize(n.Body)
	bodyExit := placeBlock(d, n.Body, cx, top)

	diaTop := bodyExit.Y + vGap
	dw := diaW(n.Cond)
	d.Shapes = append(d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: diaTop, W: dw, H: diaH, Text: n.Cond})
	diaCy := diaTop + diaH/2
	d.Edges = append(d.Edges, edge(bodyExit, diagram.Point{X: cx, Y: diaTop}))

	// Ні — дуга повернення справа, вгору до лінії входу над тілом.
	backX := cx + bw/2 + arcGap
	mergeY := top - vGap/2
	d.Edges = append(d.Edges, diagram.Edge{Label: "Ні", Points: []diagram.Point{
		{X: cx + dw/2, Y: diaCy}, {X: backX, Y: diaCy}, {X: backX, Y: mergeY}, {X: cx, Y: mergeY},
	}})
	// Так — вихід униз.
	contY := diaTop + diaH + vGap
	d.Edges = append(d.Edges, diagram.Edge{Label: "Так", Points: []diagram.Point{
		{X: cx, Y: diaTop + diaH}, {X: cx, Y: contY},
	}})
	return diagram.Point{X: cx, Y: contY}
}

// loopArcs малює дугу повернення (низ тіла → праворуч → вгору → правий кут
// заголовка) і дугу виходу (лівий кут заголовка → ліворуч → вниз → центр).
// headHalf — піввисота заголовка по X (до бічного кута), повертає точку виходу.
func loopArcs(d *diagram.Diagram, cx, headHalf, headCy, bodyHalf, bodyBottom float64, exitLabel string) diagram.Point {
	backX := cx + bodyHalf + arcGap
	leftX := cx - bodyHalf - arcGap
	contY := bodyBottom + vGap
	// Дуга повернення — у правий кут заголовка.
	d.Edges = append(d.Edges, diagram.Edge{Points: []diagram.Point{
		{X: cx, Y: bodyBottom}, {X: backX, Y: bodyBottom}, {X: backX, Y: headCy}, {X: cx + headHalf, Y: headCy},
	}})
	// Дуга виходу — з лівого кута заголовка вниз до продовження.
	d.Edges = append(d.Edges, diagram.Edge{Label: exitLabel, Points: []diagram.Point{
		{X: cx - headHalf, Y: headCy}, {X: leftX, Y: headCy}, {X: leftX, Y: contY}, {X: cx, Y: contY},
	}})
	return diagram.Point{X: cx, Y: contY}
}

// --- дрібні помічники ---

func term(cx, top float64, text string) diagram.Shape {
	return diagram.Shape{Kind: diagram.Terminator, X: cx - termW/2, Y: top, W: termW, H: termH, Text: text}
}

func edge(a, b diagram.Point) diagram.Edge {
	return diagram.Edge{Points: []diagram.Point{a, b}}
}
