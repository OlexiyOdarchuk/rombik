// Пакет layout перетворює логічний ir у геометричний diagram. Уся «магія» тут,
// і вона НЕ про графи: структурований код — це вкладене дерево, тож розкладка
// рекурсивна. size() рахує габарити піддерева, build.place() роздає координати.
//
// Усі шляхи, що завершують функцію (return/raise/exit і природний вихід),
// збираються у build.ends і зводяться в ЄДИНИЙ термінатор «Кінець».
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

// hexW — ширина шестикутника під підпис (із запасом на скоси з боків).
func hexW(spec string) float64 { return max(minBoxW, textW(spec)+hexH) }

// Options — перемикачі рендера (галочки в інтерфейсі). Серіалізується в/з JSON.
type Options struct {
	// CallAsProcess: виклик підпрограми малювати звичайним прямокутником
	// (не ДСТУ-символом «підпрограма») — на вимогу деяких викладачів.
	CallAsProcess bool `json:"callAsProcess"`
}

// build несе стан розкладки: полотно, точки до єдиного Кінця і стек збирачів
// break для поточних циклів.
type build struct {
	d          *diagram.Diagram
	ends       []diagram.Point
	loopBreaks [][]diagram.Point
}

func (b *build) pushLoop() { b.loopBreaks = append(b.loopBreaks, nil) }

func (b *build) popLoop() []diagram.Point {
	n := len(b.loopBreaks) - 1
	pts := b.loopBreaks[n]
	b.loopBreaks = b.loopBreaks[:n]
	return pts
}

func (b *build) recordBreak(p diagram.Point) {
	if n := len(b.loopBreaks); n > 0 {
		b.loopBreaks[n-1] = append(b.loopBreaks[n-1], p)
	}
}

// routeBreaks зводить точки break до виходу циклу (contY). Без вістря — голову
// дасть наступне ребро (вхід у фігуру/Кінець).
func (b *build) routeBreaks(cx, contY float64, pts []diagram.Point) {
	for _, p := range pts {
		if p.X > cx-1 && p.X < cx+1 {
			b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{p, P(cx, contY)}})
		} else {
			b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{p, P(p.X, contY), P(cx, contY)}})
		}
	}
}

// Build розкладає програму (тіло) у повну діаграму: Початок → тіло → Кінець.
func Build(prog *ir.Block, opts Options) *diagram.Diagram {
	if opts.CallAsProcess {
		mapCalls(prog)
	}
	bw, _ := blockSize(prog)
	w := bw + 2*margin
	cx := w / 2

	b := &build{d: &diagram.Diagram{}}
	top := float64(margin)
	b.d.Shapes = append(b.d.Shapes, term(cx, top, "Початок"))

	bodyTop := top + termH + vGap
	b.d.Edges = append(b.d.Edges, edge(P(cx, top+termH), P(cx, bodyTop)))
	exit, ended := b.placeBlock(prog, cx, bodyTop)
	if !ended { // природний вихід теж веде у Кінець
		b.ends = append(b.ends, exit)
	}

	// Єдиний Кінець — нижче всього вмісту; усі виходи зводимо до нього.
	// Для кількох виходів лишаємо запас під горизонтальну шину, щоб вона не
	// проходила по краю фігур.
	kY := contentBottom(b.d) + vGap
	if len(b.ends) > 1 {
		kY += mergeGap
	}
	b.routeEnds(cx, kY)
	b.d.Shapes = append(b.d.Shapes, term(cx, kY, "Кінець"))

	b.d.W = w
	b.d.H = kY + termH + margin
	return b.d
}

// routeEnds зводить усі точки-виходи у єдиний Кінець однією шиною знизу.
func (b *build) routeEnds(cx, kY float64) {
	switch len(b.ends) {
	case 0:
		return
	case 1:
		p := b.ends[0]
		if p.X == cx {
			b.d.Edges = append(b.d.Edges, edge(p, P(cx, kY)))
		} else { // одна суцільна ламана зі стрілкою в Кінець
			b.d.Edges = append(b.d.Edges, diagram.Edge{Points: []diagram.Point{
				p, P(p.X, kY-mergeGap), P(cx, kY-mergeGap), P(cx, kY)}})
		}
		return
	}
	// Кілька виходів: ведемо прямо вниз до шини, якщо стовпець чистий; на бічну
	// рейку — лише ті, під ким реально є фігура (щоб не різати її).
	cy := kY - mergeGap
	leftRail := margin * 0.5
	rightRail := 2*cx - margin*0.5
	for _, p := range b.ends {
		if !b.blockedDown(p.X, p.Y, cy) { // стовпець чистий — прямо вниз, тоді до центру
			if p.X > cx-1 && p.X < cx+1 {
				b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{p, P(cx, cy)}})
			} else {
				b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{p, P(p.X, cy), P(cx, cy)}})
			}
			continue
		}
		rail := leftRail // блокує — обходимо рейкою (вниз від центру, тоді вбік)
		if p.X > cx {
			rail = rightRail
		}
		drop := p.Y + mergeGap/2
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
			p, P(p.X, drop), P(rail, drop), P(rail, cy), P(cx, cy)}})
	}
	b.d.Edges = append(b.d.Edges, edge(P(cx, cy), P(cx, kY))) // шина → Кінець
}

// blockedDown — чи є фігура в стовпці x між fromY і toY (тобто прямий шлях
// униз перетнув би її).
func (b *build) blockedDown(x, fromY, toY float64) bool {
	for _, s := range b.d.Shapes {
		if s.Y > fromY+1 && s.Y < toY && s.X-1 < x && x < s.X+s.W+1 {
			return true
		}
	}
	return false
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
	case *ir.Terminal:
		return textW(x.Text), boxH
	case *ir.Block:
		return blockSize(x)
	case *ir.If:
		tw, th := branchSize(x.Then)
		ew, eh := branchSize(x.Else)
		h := diaH + branchGap + max(th, eh) + mergeGap
		if (len(x.Then.Stmts) == 0) != (len(x.Else.Stmts) == 0) { // guard: одна гілка порожня
			return max(diaW(x.Cond), max(tw, ew)+2*hGap), h
		}
		return max(diaW(x.Cond), tw+hGap+ew), h
	case *ir.For:
		bw, bh := blockSize(x.Body)
		return max(hexW(x.Spec), bw) + 2*arcGap, hexH + vGap + bh + vGap
	case *ir.While:
		bw, bh := blockSize(x.Body)
		return max(diaW(x.Cond), bw) + 2*arcGap, diaH + vGap + bh + vGap
	case *ir.DoWhile:
		bw, bh := blockSize(x.Body)
		return max(diaW(x.Cond), bw) + 2*arcGap, bh + vGap + diaH + vGap
	case *ir.InfLoop:
		bw, bh := blockSize(x.Body)
		return bw + 2*arcGap, bh + vGap
	case *ir.Break:
		return 0, 0 // без фігури
	}
	return minBoxW, boxH
}

// branchSize — габарити гілки if; порожня гілка не резервує ширини (щоб не
// зміщувати протилежну гілку вбік).
func branchSize(blk *ir.Block) (w, h float64) {
	if blk == nil || len(blk.Stmts) == 0 {
		return 0, 0
	}
	return blockSize(blk)
}

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
// place малює вузол з верхом-центром у (cx, top); повертає точку виходу (низ-центр)
// і прапорець «гілка завершена» (return/raise/exit — потік далі не йде).

func (b *build) place(n ir.Node, cx, top float64) (diagram.Point, bool) {
	switch x := n.(type) {
	case *ir.Process:
		return b.leaf(diagram.Process, x.Text, cx, top), false
	case *ir.IO:
		return b.leaf(diagram.InOut, x.Text, cx, top), false
	case *ir.Call:
		return b.leaf(diagram.Predef, x.Text, cx, top), false
	case *ir.Terminal:
		return b.placeTerminal(x.Text, cx, top)
	case *ir.Block:
		return b.placeBlock(x, cx, top)
	case *ir.If:
		return b.placeIf(x, cx, top)
	case *ir.For:
		return b.placeFor(x, cx, top), false
	case *ir.While:
		return b.placeWhile(x, cx, top), false
	case *ir.DoWhile:
		return b.placeDoWhile(x, cx, top), false
	case *ir.InfLoop:
		return b.placeInfLoop(x, cx, top), false
	case *ir.Break:
		// стрибок на вихід циклу — без фігури; з'єднання зробить routeBreaks
		b.recordBreak(P(cx, top))
		return P(cx, top), true
	}
	return P(cx, top), false
}

func (b *build) leaf(kind diagram.Kind, text string, cx, top float64) diagram.Point {
	w := textW(text)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: kind, X: cx - w/2, Y: top, W: w, H: boxH, Text: text})
	return P(cx, top+boxH)
}

// placeTerminal — return/raise/exit: звичайний прямокутник, який веде у Кінець.
func (b *build) placeTerminal(text string, cx, top float64) (diagram.Point, bool) {
	exit := b.leaf(diagram.Process, text, cx, top)
	b.ends = append(b.ends, exit)
	return exit, true
}

func (b *build) placeBlock(blk *ir.Block, cx, top float64) (diagram.Point, bool) {
	cur := top
	exit := P(cx, top)
	for i, s := range blk.Stmts {
		// break не має фігури — з'єднання від попереднього виходу зробить
		// routeBreaks; стрілку-голову в нікуди не малюємо.
		if _, isBreak := s.(*ir.Break); isBreak {
			b.recordBreak(exit)
			return exit, true
		}
		if i > 0 {
			b.d.Edges = append(b.d.Edges, edge(exit, P(cx, cur)))
		}
		var ended bool
		exit, ended = b.place(s, cx, cur)
		if ended {
			return exit, true // решта інструкцій недосяжні
		}
		cur = exit.Y + vGap
	}
	return exit, false
}

func (b *build) placeIf(n *ir.If, cx, top float64) (diagram.Point, bool) {
	dw := diaW(n.Cond)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: top, W: dw, H: diaH, Text: n.Cond})
	midY := top + diaH/2
	branchTop := top + diaH + branchGap

	thenEmpty := len(n.Then.Stmts) == 0
	elseEmpty := len(n.Else.Stmts) == 0

	// Guard (одна гілка порожня): дія прямо вниз, порожня гілка обходить збоку.
	if elseEmpty && !thenEmpty {
		return b.guard(n.Then, "Так", "Ні", +1, cx, dw, midY, branchTop)
	}
	if thenEmpty && !elseEmpty {
		return b.guard(n.Else, "Ні", "Так", -1, cx, dw, midY, branchTop)
	}

	// Обидві гілки — симетрично.
	tw, th := branchSize(n.Then)
	ew, eh := branchSize(n.Else)
	total := tw + hGap + ew
	thenCx := cx - total/2 + tw/2
	elseCx := cx + total/2 - ew/2
	mergeY := branchTop + max(th, eh) + mergeGap

	thenEnded := b.branch(n.Then, "Так", cx, cx-dw/2, midY, thenCx, branchTop, mergeY)
	elseEnded := b.branch(n.Else, "Ні", cx, cx+dw/2, midY, elseCx, branchTop, mergeY)
	return P(cx, mergeY), thenEnded && elseEnded
}

// guard малює «if cond: BODY» без else: дія прямо вниз (downLabel), порожня
// гілка — вбік (sideLabel) і вниз до злиття. side=+1 праворуч, -1 ліворуч.
func (b *build) guard(body *ir.Block, downLabel, sideLabel string, side, cx, dw, midY, branchTop float64) (diagram.Point, bool) {
	bw, bh := blockSize(body)
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: downLabel, Points: []diagram.Point{
		{X: cx, Y: midY + diaH/2}, {X: cx, Y: branchTop},
	}})
	exit, ended := b.placeBlock(body, cx, branchTop)
	mergeY := branchTop + bh + mergeGap

	vx := cx + side*dw/2
	sideX := cx + side*(bw/2+hGap)
	b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Label: sideLabel, Points: []diagram.Point{
		{X: vx, Y: midY}, {X: sideX, Y: midY}, {X: sideX, Y: mergeY}, {X: cx, Y: mergeY},
	}})
	if !ended {
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{exit, {X: cx, Y: mergeY}}})
	}
	return P(cx, mergeY), false // порожня гілка завжди дає продовження
}

// branch малює одну гілку if від кута ромба (vx,midY). Повертає, чи гілка
// завершилась (return/raise/exit). Порожня гілка — ОДНА суцільна лінія до
// злиття без стрілки-голови в нікуди.
func (b *build) branch(blk *ir.Block, label string, cx, vx, midY, bcx, branchTop, mergeY float64) bool {
	if len(blk.Stmts) == 0 {
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Label: label, Points: []diagram.Point{
			{X: vx, Y: midY}, {X: bcx, Y: midY}, {X: bcx, Y: mergeY}, {X: cx, Y: mergeY},
		}})
		return false
	}
	if len(blk.Stmts) == 1 {
		if _, ok := blk.Stmts[0].(*ir.Break); ok { // гілка «if …: break» — на вихід циклу
			b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Label: label, Points: []diagram.Point{
				{X: vx, Y: midY}, {X: bcx, Y: midY},
			}})
			b.recordBreak(P(bcx, midY))
			return true
		}
	}
	// Непорожня: стрілка з підписом у верх першого блоку, далі — злиття.
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: label, Points: []diagram.Point{
		{X: vx, Y: midY}, {X: bcx, Y: midY}, {X: bcx, Y: branchTop},
	}})
	exit, ended := b.placeBlock(blk, bcx, branchTop)
	if !ended {
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
			exit, {X: exit.X, Y: mergeY}, {X: cx, Y: mergeY},
		}})
	}
	return ended
}

// placeFor — цикл for: шестикутник згори, тіло під ним, дуга повернення справа,
// вихід зліва вниз (як у курсових схемах на fletcher).
func (b *build) placeFor(n *ir.For, cx, top float64) diagram.Point {
	hw := hexW(n.Spec)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Hexagon, X: cx - hw/2, Y: top, W: hw, H: hexH, Text: n.Spec})
	headCy := top + hexH/2

	bw, _ := blockSize(n.Body)
	bodyTop := top + hexH + vGap
	b.d.Edges = append(b.d.Edges, edge(P(cx, top+hexH), P(cx, bodyTop)))
	b.pushLoop()
	bodyExit, _ := b.placeBlock(n.Body, cx, bodyTop)
	brks := b.popLoop()
	cont := b.loopArcs(cx, hw/2, headCy, bw/2, bodyExit.Y, "")
	b.routeBreaks(cx, cont.Y, brks)
	return cont
}

// placeWhile — цикл while: ромб-передумова, Так→тіло, дуга повернення справа,
// Ні→вихід зліва вниз.
func (b *build) placeWhile(n *ir.While, cx, top float64) diagram.Point {
	dw := diaW(n.Cond)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: top, W: dw, H: diaH, Text: n.Cond})
	headCy := top + diaH/2

	bw, _ := blockSize(n.Body)
	bodyTop := top + diaH + vGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: "Так", Points: []diagram.Point{{X: cx, Y: top + diaH}, {X: cx, Y: bodyTop}}})
	b.pushLoop()
	bodyExit, _ := b.placeBlock(n.Body, cx, bodyTop)
	brks := b.popLoop()
	cont := b.loopArcs(cx, dw/2, headCy, bw/2, bodyExit.Y, "Ні")
	b.routeBreaks(cx, cont.Y, brks)
	return cont
}

// placeDoWhile — цикл з післяумовою: тіло згори, ромб-умова знизу, Так→вихід,
// Ні→дуга повернення справа до лінії входу (повтор тіла).
func (b *build) placeDoWhile(n *ir.DoWhile, cx, top float64) diagram.Point {
	bw, _ := blockSize(n.Body)
	b.pushLoop()
	bodyExit, _ := b.placeBlock(n.Body, cx, top)
	brks := b.popLoop()

	diaTop := bodyExit.Y + vGap
	dw := diaW(n.Cond)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: diaTop, W: dw, H: diaH, Text: n.Cond})
	diaCy := diaTop + diaH/2
	b.d.Edges = append(b.d.Edges, edge(bodyExit, P(cx, diaTop)))

	backX := cx + bw/2 + arcGap
	mergeY := top - vGap/2
	// Без вістря: вливається в лінію входу (не у фігуру) — щоб не було «двох голів».
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: "Ні", Arrowless: true, Points: []diagram.Point{
		{X: cx + dw/2, Y: diaCy}, {X: backX, Y: diaCy}, {X: backX, Y: mergeY}, {X: cx, Y: mergeY},
	}})
	contY := diaTop + diaH + vGap
	// Вихід — без вістря: голову дасть наступне ребро (вхід у фігуру/Кінець).
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: "Так", Arrowless: true, Points: []diagram.Point{
		{X: cx, Y: diaTop + diaH}, {X: cx, Y: contY},
	}})
	b.routeBreaks(cx, contY, brks)
	return P(cx, contY)
}

// placeInfLoop — нескінченний цикл while True: тіло + безумовна дуга повернення
// справа; вихід — лише через break (routeBreaks).
func (b *build) placeInfLoop(n *ir.InfLoop, cx, top float64) diagram.Point {
	bw, _ := blockSize(n.Body)
	b.pushLoop()
	bodyExit, _ := b.placeBlock(n.Body, cx, top)
	brks := b.popLoop()

	// Безумовна дуга повернення справа: низ тіла → праворуч → вгору → вхід.
	// Із центру низу (трохи вниз) і без вістря (вливається в лінію входу).
	backX := cx + bw/2 + arcGap
	mergeY := top - vGap/2
	drop := bodyExit.Y + mergeGap/2
	b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
		{X: cx, Y: bodyExit.Y}, {X: cx, Y: drop}, {X: backX, Y: drop}, {X: backX, Y: mergeY}, {X: cx, Y: mergeY},
	}})

	contY := bodyExit.Y + vGap
	b.routeBreaks(cx, contY, brks)
	return P(cx, contY)
}

// loopArcs малює дугу повернення (низ тіла → праворуч → вгору → правий кут
// заголовка) і дугу виходу (лівий кут заголовка → ліворуч → вниз → центр).
func (b *build) loopArcs(cx, headHalf, headCy, bodyHalf, bodyBottom float64, exitLabel string) diagram.Point {
	backX := cx + bodyHalf + arcGap
	leftX := cx - bodyHalf - arcGap
	contY := bodyBottom + vGap
	// Дуга повернення — у правий кут заголовка; виходить із центру низу тіла
	// (спершу трохи вниз, тоді вбік — не «з кута»).
	drop := bodyBottom + mergeGap/2
	b.d.Edges = append(b.d.Edges, diagram.Edge{Points: []diagram.Point{
		{X: cx, Y: bodyBottom}, {X: cx, Y: drop}, {X: backX, Y: drop}, {X: backX, Y: headCy}, {X: cx + headHalf, Y: headCy},
	}})
	// Вихід із циклу — без вістря: голову дасть наступне ребро (вхід у фігуру).
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: exitLabel, Arrowless: true, Points: []diagram.Point{
		{X: cx - headHalf, Y: headCy}, {X: leftX, Y: headCy}, {X: leftX, Y: contY}, {X: cx, Y: contY},
	}})
	return P(cx, contY)
}

// --- дрібні помічники ---

func P(x, y float64) diagram.Point { return diagram.Point{X: x, Y: y} }

func term(cx, top float64, text string) diagram.Shape {
	return diagram.Shape{Kind: diagram.Terminator, X: cx - termW/2, Y: top, W: termW, H: termH, Text: text}
}

func edge(a, b diagram.Point) diagram.Edge {
	return diagram.Edge{Points: []diagram.Point{a, b}}
}

// contentBottom — найнижча точка серед фігур і ребер (для розміщення Кінця).
func contentBottom(d *diagram.Diagram) float64 {
	var b float64
	for _, s := range d.Shapes {
		b = max(b, s.Y+s.H)
	}
	for _, e := range d.Edges {
		for _, p := range e.Points {
			b = max(b, p.Y)
		}
	}
	return b
}
