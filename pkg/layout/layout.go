// Пакет layout перетворює логічний ir у геометричний diagram. Уся «магія» тут,
// і вона НЕ про графи: структурований код — це вкладене дерево, тож розкладка
// рекурсивна. size() рахує габарити піддерева, build.place() роздає координати.
//
// Усі шляхи, що завершують функцію (return/raise/exit і природний вихід),
// збираються у build.ends і зводяться в термінатор «Кінець».
package layout

import (
	"regexp"
	"strings"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
	"github.com/OlexiyOdarchuk/rombik/pkg/ir"
	"github.com/OlexiyOdarchuk/rombik/pkg/route"
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
	// SingleEnd: усі виходи зводити в ОДИН Кінець (інакше — окремий «Кінець»
	// на кожен return/raise/exit; обидва варіанти за ДСТУ).
	SingleEnd bool `json:"singleEnd"`
	// Yes/No: підписи гілок розв'язку (типово «Так»/«Ні»).
	Yes string `json:"yes"`
	No  string `json:"no"`
	// InWord/OutWord: слова вводу/виводу (типово «Ввід»/«Вивід»; інші варіанти —
	// «Введення»/«Виведення», «Ввести»/«Вивести»).
	InWord  string `json:"inWord"`
	OutWord string `json:"outWord"`
	// StartText/EndText: текст термінаторів (типово «Початок»/«Кінець»).
	StartText string `json:"startText"`
	EndText   string `json:"endText"`
	// StripTypes: прибирати тип-анотації (a: float = 3.1 -> a = 3.1).
	StripTypes bool `json:"stripTypes"`
	// ReturnAsIO: return малювати паралелограмом (інакше — прямокутником).
	ReturnAsIO bool `json:"returnAsIO"`
}

// build несе стан розкладки: полотно, точки до єдиного Кінця, стек збирачів
// break і опції.
type build struct {
	d          *diagram.Diagram
	ends       []diagram.Point
	loopBreaks [][]diagram.Point
	singleEnd  bool
	yes, no    string
	inWord     string
	outWord    string
	endText    string
	stripTypes bool
	retAsIO    bool
}

// typeAnnRe — «name: type =» -> «name =» (для зняття тип-анотацій).
var typeAnnRe = regexp.MustCompile(`^([\w.]+)\s*:\s*[^=]+=`)

// ioText застосовує обрані слова вводу/виводу до тексту ir.IO.
func (b *build) ioText(t string) string {
	if r, ok := strings.CutPrefix(t, "Ввід"); ok {
		return b.inWord + r
	}
	if r, ok := strings.CutPrefix(t, "Вивід"); ok {
		return b.outWord + r
	}
	return t
}

// procText прибирає тип-анотацію, якщо ввімкнено опцію.
func (b *build) procText(t string) string {
	if b.stripTypes {
		return typeAnnRe.ReplaceAllString(t, "$1 =")
	}
	return t
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
	b := &build{
		d: &diagram.Diagram{}, singleEnd: opts.SingleEnd,
		yes: opts.Yes, no: opts.No, inWord: opts.InWord, outWord: opts.OutWord,
		endText: opts.EndText, stripTypes: opts.StripTypes, retAsIO: opts.ReturnAsIO,
	}
	def := func(p *string, d string) {
		if *p == "" {
			*p = d
		}
	}
	def(&b.yes, "Так")
	def(&b.no, "Ні")
	def(&b.inWord, "Ввід")
	def(&b.outWord, "Вивід")
	def(&b.endText, "Кінець")
	startText := opts.StartText
	if startText == "" {
		startText = "Початок"
	}
	bw, _ := b.blockSize(prog)
	w := bw + 2*margin
	cx := w / 2

	top := float64(margin)
	b.d.Shapes = append(b.d.Shapes, term(cx, top, startText))

	bodyTop := top + termH + vGap
	b.d.Edges = append(b.d.Edges, edge(P(cx, top+termH), P(cx, bodyTop)))
	exit, ended := b.placeBlock(prog, cx, bodyTop)
	if !ended { // природний вихід веде у Кінець (у per-exit це єдиний спільний)
		b.ends = append(b.ends, exit)
	}

	// Спільний Кінець — лише якщо є що в нього зводити (у per-exit режимі
	// return/raise мають свої локальні Кінці й сюди не потрапляють).
	if len(b.ends) > 0 {
		kY := contentBottom(b.d) + vGap
		if len(b.ends) > 1 {
			kY += mergeGap // запас під шину, щоб не різала фігури
		}
		b.routeEnds(cx, kY)
		b.d.Shapes = append(b.d.Shapes, term(cx, kY, b.endText))
	}

	b.d.W = w
	b.d.H = contentBottom(b.d) + margin
	return b.d
}

// routeEnds зводить усі точки-виходи у Кінець маршрутизатором (обхід фігур).
func (b *build) routeEnds(cx, kY float64) {
	if len(b.ends) == 0 {
		return
	}
	obs := b.rects()
	if len(b.ends) == 1 { // одне ребро прямо в Кінець (зі стрілкою)
		b.d.Edges = append(b.d.Edges, routed(route.Route(toPt(b.ends[0]), route.Pt{X: cx, Y: kY}, obs), false))
		return
	}
	// Кілька: кожен — у точку збору над Кінцем (без вістря), тоді одна стрілка.
	jy := kY - vGap
	for _, e := range b.ends {
		b.d.Edges = append(b.d.Edges, routed(route.Route(toPt(e), route.Pt{X: cx, Y: jy}, obs), true))
	}
	b.d.Edges = append(b.d.Edges, edge(P(cx, jy), P(cx, kY)))
}

// rects — фігури як перешкоди для маршрутизатора.
func (b *build) rects() []route.Rect {
	rs := make([]route.Rect, len(b.d.Shapes))
	for i, s := range b.d.Shapes {
		rs[i] = route.Rect{X: s.X, Y: s.Y, W: s.W, H: s.H}
	}
	return rs
}

func toPt(p diagram.Point) route.Pt { return route.Pt{X: p.X, Y: p.Y} }

func routed(path []route.Pt, arrowless bool) diagram.Edge {
	pts := make([]diagram.Point, len(path))
	for i, p := range path {
		pts[i] = diagram.Point{X: p.X, Y: p.Y}
	}
	return diagram.Edge{Points: pts, Arrowless: arrowless}
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

func (b *build) size(n ir.Node) (w, h float64) {
	switch x := n.(type) {
	case *ir.Process:
		return textW(b.procText(x.Text)), boxH
	case *ir.IO:
		return textW(b.ioText(x.Text)), boxH
	case *ir.Call:
		return textW(x.Text), boxH
	case *ir.Terminal:
		if b.singleEnd {
			return textW(x.Text), boxH
		}
		return max(textW(x.Text), termW), boxH + vGap + termH // блок + локальний Кінець
	case *ir.Block:
		return b.blockSize(x)
	case *ir.If:
		tw, th := b.branchSize(x.Then)
		ew, eh := b.branchSize(x.Else)
		h := diaH + branchGap + max(th, eh) + mergeGap
		if (len(x.Then.Stmts) == 0) != (len(x.Else.Stmts) == 0) { // guard: одна гілка порожня
			return max(diaW(x.Cond), max(tw, ew)+2*hGap), h
		}
		return max(diaW(x.Cond), tw+hGap+ew), h
	case *ir.For:
		bw, bh := b.blockSize(x.Body)
		return max(hexW(x.Spec), bw) + 2*arcGap, hexH + vGap + bh + vGap
	case *ir.While:
		bw, bh := b.blockSize(x.Body)
		return max(diaW(x.Cond), bw) + 2*arcGap, diaH + vGap + bh + vGap
	case *ir.DoWhile:
		bw, bh := b.blockSize(x.Body)
		return max(diaW(x.Cond), bw) + 2*arcGap, bh + vGap + diaH + vGap
	case *ir.InfLoop:
		bw, bh := b.blockSize(x.Body)
		return bw + 2*arcGap, bh + vGap
	case *ir.Break:
		return 0, 0 // без фігури
	}
	return minBoxW, boxH
}

// branchSize — габарити гілки if; порожня гілка не резервує ширини (щоб не
// зміщувати протилежну гілку вбік).
func (b *build) branchSize(blk *ir.Block) (w, h float64) {
	if blk == nil || len(blk.Stmts) == 0 {
		return 0, 0
	}
	return b.blockSize(blk)
}

func (b *build) blockSize(blk *ir.Block) (w, h float64) {
	if blk == nil || len(blk.Stmts) == 0 {
		return minBoxW, 0
	}
	for i, s := range blk.Stmts {
		sw, sh := b.size(s)
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
		return b.leaf(diagram.Process, b.procText(x.Text), cx, top), false
	case *ir.IO:
		return b.leaf(diagram.InOut, b.ioText(x.Text), cx, top), false
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

// placeTerminal — return/raise/exit: звичайний прямокутник. У режимі singleEnd
// веде у спільний Кінець (через b.ends), інакше — свій локальний «Кінець».
func (b *build) placeTerminal(text string, cx, top float64) (diagram.Point, bool) {
	kind := diagram.Process
	if b.retAsIO {
		kind = diagram.InOut
	}
	exit := b.leaf(kind, text, cx, top)
	if b.singleEnd {
		b.ends = append(b.ends, exit)
		return exit, true
	}
	endTop := exit.Y + vGap
	b.d.Edges = append(b.d.Edges, edge(exit, P(cx, endTop)))
	b.d.Shapes = append(b.d.Shapes, term(cx, endTop, b.endText))
	return P(cx, endTop+termH), true
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
		return b.guard(n.Then, b.yes, b.no, +1, cx, dw, midY, branchTop)
	}
	if thenEmpty && !elseEmpty {
		return b.guard(n.Else, b.no, b.yes, -1, cx, dw, midY, branchTop)
	}

	// Обидві гілки — симетрично.
	tw, th := b.branchSize(n.Then)
	ew, eh := b.branchSize(n.Else)
	total := tw + hGap + ew
	thenCx := cx - total/2 + tw/2
	elseCx := cx + total/2 - ew/2
	mergeY := branchTop + max(th, eh) + mergeGap

	thenEnded := b.branch(n.Then, b.yes, cx, cx-dw/2, midY, thenCx, branchTop, mergeY)
	elseEnded := b.branch(n.Else, b.no, cx, cx+dw/2, midY, elseCx, branchTop, mergeY)
	return P(cx, mergeY), thenEnded && elseEnded
}

// guard малює «if cond: BODY» без else: дія прямо вниз (downLabel), порожня
// гілка — вбік (sideLabel) і вниз до злиття. side=+1 праворуч, -1 ліворуч.
func (b *build) guard(body *ir.Block, downLabel, sideLabel string, side, cx, dw, midY, branchTop float64) (diagram.Point, bool) {
	bw, bh := b.blockSize(body)
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

// endsBlock — чи блок завершується (останній стейтмент return/raise/exit/break,
// або if, де обидві гілки завершуються).
func endsBlock(blk *ir.Block) bool {
	if blk == nil || len(blk.Stmts) == 0 {
		return false
	}
	switch x := blk.Stmts[len(blk.Stmts)-1].(type) {
	case *ir.Terminal, *ir.Break:
		return true
	case *ir.If:
		return endsBlock(x.Then) && endsBlock(x.Else)
	}
	return false
}

// termGuardLast — якщо останній стейтмент тіла це guard (if із завершальною дією
// і порожнім else), повертає його. Тоді «Ні» цього guard природно є дугою
// повернення циклу (вгору в заголовок), без зайвого «вниз-вгору».
func termGuardLast(blk *ir.Block) *ir.If {
	if blk == nil || len(blk.Stmts) == 0 {
		return nil
	}
	g, ok := blk.Stmts[len(blk.Stmts)-1].(*ir.If)
	if ok && len(g.Else.Stmts) == 0 && endsBlock(g.Then) {
		return g
	}
	return nil
}

// placeLoopGuard малює тіло-guard циклу: ромб-умова, Так→завершальна дія (вниз),
// Ні→дуга повернення ВГОРУ в заголовок (праворуч). Повертає вихід циклу.
// headHalf/headCy — піввисота-по-X і центр заголовка; headBottom — його низ;
// entryLabel — підпис ребра заголовок→ромб (для while це «Так»).
func (b *build) placeLoopGuard(g *ir.If, cx, headHalf, headCy, headBottom float64, entryLabel, exitLabel string) diagram.Point {
	dw := diaW(g.Cond)
	diaTop := headBottom + vGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: entryLabel, Points: []diagram.Point{{X: cx, Y: headBottom}, {X: cx, Y: diaTop}}})
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: diaTop, W: dw, H: diaH, Text: g.Cond})
	diaMidY := diaTop + diaH/2

	// Так → вниз → завершальна дія (вона сама йде у свій Кінець).
	actTop := diaTop + diaH + branchGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: b.yes, Points: []diagram.Point{{X: cx, Y: diaTop + diaH}, {X: cx, Y: actTop}}})
	b.placeBlock(g.Then, cx, actTop)

	aw, _ := b.blockSize(g.Then)
	half := max(dw, aw) / 2
	// Ні → правий кут ромба → ВГОРУ в правий кут заголовка (дуга повернення).
	backX := cx + half + arcGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: b.no, Points: []diagram.Point{
		{X: cx + dw/2, Y: diaMidY}, {X: backX, Y: diaMidY}, {X: backX, Y: headCy}, {X: cx + headHalf, Y: headCy},
	}})
	// Вихід циклу: лівий кут заголовка → вниз нижче всього → центр.
	contY := contentBottom(b.d) + vGap
	leftX := cx - half - arcGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Label: exitLabel, Points: []diagram.Point{
		{X: cx - headHalf, Y: headCy}, {X: leftX, Y: headCy}, {X: leftX, Y: contY}, {X: cx, Y: contY},
	}})
	return P(cx, contY)
}

// placeGuardLoopBody розкладає тіло циклу, що ЗАКІНЧУЄТЬСЯ guard-ом: спершу
// попередні інструкції (звичайно), тоді останній guard, чиє «Ні» — дуга
// повернення вгору в заголовок. Обробляє break усередині попередніх інструкцій.
func (b *build) placeGuardLoopBody(body *ir.Block, g *ir.If, cx, headHalf, headCy, headBottom float64, entryLabel, exitLabel string) diagram.Point {
	b.pushLoop()
	fromY, el := headBottom, entryLabel
	if pre := body.Stmts[:len(body.Stmts)-1]; len(pre) > 0 {
		bodyTop := headBottom + vGap
		b.d.Edges = append(b.d.Edges, diagram.Edge{Label: entryLabel, Points: []diagram.Point{{X: cx, Y: headBottom}, {X: cx, Y: bodyTop}}})
		exit, _ := b.placeBlock(&ir.Block{Stmts: pre}, cx, bodyTop)
		fromY, el = exit.Y, ""
	}
	cont := b.placeLoopGuard(g, cx, headHalf, headCy, fromY, el, exitLabel)
	brks := b.popLoop()
	b.routeBreaks(cx, cont.Y, brks)
	return cont
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

	// Тіло закінчується guard-ом → його «Ні» одразу дугою вгору в заголовок.
	if g := termGuardLast(n.Body); g != nil {
		return b.placeGuardLoopBody(n.Body, g, cx, hw/2, headCy, top+hexH, "", "")
	}

	bw, _ := b.blockSize(n.Body)
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

	if g := termGuardLast(n.Body); g != nil {
		return b.placeGuardLoopBody(n.Body, g, cx, dw/2, headCy, top+diaH, b.yes, b.no)
	}

	bw, _ := b.blockSize(n.Body)
	bodyTop := top + diaH + vGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: b.yes, Points: []diagram.Point{{X: cx, Y: top + diaH}, {X: cx, Y: bodyTop}}})
	b.pushLoop()
	bodyExit, _ := b.placeBlock(n.Body, cx, bodyTop)
	brks := b.popLoop()
	cont := b.loopArcs(cx, dw/2, headCy, bw/2, bodyExit.Y, b.no)
	b.routeBreaks(cx, cont.Y, brks)
	return cont
}

// placeDoWhile — цикл з післяумовою: тіло згори, ромб-умова знизу, Так→вихід,
// Ні→дуга повернення справа до лінії входу (повтор тіла).
func (b *build) placeDoWhile(n *ir.DoWhile, cx, top float64) diagram.Point {
	bw, _ := b.blockSize(n.Body)
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
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: b.no, Arrowless: true, Points: []diagram.Point{
		{X: cx + dw/2, Y: diaCy}, {X: backX, Y: diaCy}, {X: backX, Y: mergeY}, {X: cx, Y: mergeY},
	}})
	contY := diaTop + diaH + vGap
	// Вихід — без вістря: голову дасть наступне ребро (вхід у фігуру/Кінець).
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: b.yes, Arrowless: true, Points: []diagram.Point{
		{X: cx, Y: diaTop + diaH}, {X: cx, Y: contY},
	}})
	b.routeBreaks(cx, contY, brks)
	return P(cx, contY)
}

// placeInfLoop — нескінченний цикл while True: тіло + безумовна дуга повернення
// справа; вихід — лише через break (routeBreaks).
func (b *build) placeInfLoop(n *ir.InfLoop, cx, top float64) diagram.Point {
	bw, _ := b.blockSize(n.Body)
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
