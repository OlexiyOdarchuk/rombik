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
	// CapWord: слово підпису схеми («Рисунок»/«Рис.»/своє; порожнє — «Рисунок»).
	CapWord string `json:"capWord"`
	// NoStart/NoEnd: не малювати «Початок»/«Кінець» — для частин схеми між
	// конекторами (розбивка великої схеми на кілька зв'язаних частин).
	NoStart bool `json:"noStart"`
	NoEnd   bool `json:"noEnd"`
}

// build несе стан розкладки: полотно, точки до єдиного Кінця, стек збирачів
// break і опції.
type build struct {
	d          *diagram.Diagram
	ends       []diagram.Point
	loopBreaks [][]diagram.Point
	loopConts  [][]diagram.Point // точки continue поточних циклів (вкладені)
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

func (b *build) pushLoop() {
	b.loopBreaks = append(b.loopBreaks, nil)
	b.loopConts = append(b.loopConts, nil)
}

// popLoop знімає поточний цикл зі стеку й повертає його точки break та continue.
func (b *build) popLoop() (brks, conts []diagram.Point) {
	n := len(b.loopBreaks) - 1
	brks, conts = b.loopBreaks[n], b.loopConts[n]
	b.loopBreaks, b.loopConts = b.loopBreaks[:n], b.loopConts[:n]
	return
}

func (b *build) recordBreak(p diagram.Point) {
	if n := len(b.loopBreaks); n > 0 {
		b.loopBreaks[n-1] = append(b.loopBreaks[n-1], p)
	}
}

func (b *build) recordContinue(p diagram.Point) {
	if n := len(b.loopConts); n > 0 {
		b.loopConts[n-1] = append(b.loopConts[n-1], p)
	}
}

// routeBreaks зводить точки break до виходу циклу (contY): кожен break спускається
// ВЛАСНОЮ колонкою прямо до рівня виходу циклу, тоді горизонтально в центр (cx).
// Спуск до самого contY (а не до рівня над дугою повернення) — принципово: інакше
// вертикаль break лягає на колонку дуги повернення (обидві на осі cx трохи вище
// contY) і візуально «вливається» в петлю циклу, хоча має йти на вихід. Спускаючись
// нижче дуги, break лише ПЕРЕТИНАЄ її горизонталь чистим прямим кутом. O(N), без
// перетинів із фігурами тіла (всі вище). Кілька break-ів фанняться по рівню.
func (b *build) routeBreaks(cx, contY float64, pts []diagram.Point) {
	if len(pts) == 0 {
		return
	}
	for i, p := range pts {
		if p.X > cx-1 && p.X < cx+1 { // вже на осі — прямо вниз
			b.d.Edges = append(b.d.Edges, diagram.Edge{Points: []diagram.Point{
				p, {X: cx, Y: contY},
			}})
			continue
		}
		safeY := contY - float64(len(pts)-1-i)*7 // i=останній → рівно contY
		pts4 := []diagram.Point{p, {X: p.X, Y: safeY}, {X: cx, Y: safeY}}
		if safeY < contY-0.5 {
			pts4 = append(pts4, diagram.Point{X: cx, Y: contY})
		}
		b.d.Edges = append(b.d.Edges, diagram.Edge{Points: pts4})
	}
}

// routeContinues заводить точки continue у колонку дуги повернення (backX).
// Щоб не перетинати фігури сусідніх гілок (напр. діаманти while), лінія 
// спускається до безпечного рівня safeY (відразу над contY), йде горизонтально
// до backX, де зливається з дугою повернення.
func (b *build) routeContinues(backX, contY float64, pts []diagram.Point) {
	if len(pts) == 0 {
		return
	}
	for i, p := range pts {
		safeY := contY - vGap*0.6 + float64(i)*6
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
			p, 
			{X: p.X, Y: safeY}, 
			{X: backX, Y: safeY},
		}})
	}
}

// Build розкладає програму (тіло) у повну діаграму: Початок → тіло → Кінець.
func Build(prog *ir.Block, opts Options) *diagram.Diagram {
	ensureBlocks(prog) // нормалізуємо nil-блоки (захист від ручних IR)
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
	bodyTop := top
	if !opts.NoStart { // частини після розриву починаються конектором, не «Початком»
		b.d.Shapes = append(b.d.Shapes, term(cx, top, startText))
		bodyTop = top + termH + vGap
		b.d.Edges = append(b.d.Edges, edge(P(cx, top+termH), P(cx, bodyTop)))
	}
	exit, ended := b.placeBlock(prog, cx, bodyTop)
	if !ended { // природний вихід веде у Кінець (у per-exit це єдиний спільний)
		b.ends = append(b.ends, exit)
	}

	// Спільний Кінець — лише якщо є що в нього зводити. NoEnd: частина
	// закінчується конектором (розрив), Кінця нема.
	if !opts.NoEnd && len(b.ends) > 0 {
		kY := contentBottom(b.d) + vGap
		if len(b.ends) > 1 {
			kY += mergeGap // запас під шину, щоб не різала фігури
		}
		b.routeEnds(cx, kY)
		b.d.Shapes = append(b.d.Shapes, term(cx, kY, b.endText))
	}

	// Нормалізація: дуги циклів можуть виходити за [margin, w-margin] (бічні
	// обходи ширші за фігури). Зсуваємо все так, щоб найлівіша точка була на
	// margin, і беремо ширину за фактичним вмістом.
	right, left := b.bodyExtent(0, 0, cx)
	if dx := margin - left; dx != 0 {
		b.shiftX(dx)
		right += dx
	}
	b.d.W = right + margin
	b.d.H = contentBottom(b.d) + margin
	return b.d
}

// routeEnds зводить усі точки-виходи (return/raise/exit у single-end) в один
// Кінець ШИНОЮ: магістраль — вертикальна колонка праворуч від усього, прокладена
// в гарантовано вільному просторі. Жодного пошуку шляху (A*) — детермінована
// геометрія O(n): кожен вихід вниз з-під блоку → вправо до шини, усі по шині вниз
// → у центр над Кінцем → у Кінець. Саме так роблять у методичках (магістраль збоку).
func (b *build) routeEnds(cx, kY float64) {
	if len(b.ends) == 0 {
		return
	}
	if len(b.ends) == 1 {
		e := b.ends[0]
		if e.X > cx-1 && e.X < cx+1 {
			b.d.Edges = append(b.d.Edges, edge(e, P(cx, kY)))
		} else {
			my := e.Y + mergeGap/2
			b.d.Edges = append(b.d.Edges, diagram.Edge{Points: []diagram.Point{e, {X: e.X, Y: my}, {X: cx, Y: my}, {X: cx, Y: kY}}})
		}
		return
	}
	right, left := b.bodyExtent(0, 0, cx)
	busRight := right + mergeGap
	busLeft := left - mergeGap
	jy := kY - vGap
	minDropR := jy
	minDropL := jy
	hasR, hasL := false, false

	for _, e := range b.ends {
		drop := e.Y + mergeGap/2
		if e.X < cx-1 {
			if drop < minDropL {
				minDropL = drop
			}
			b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
				e, {X: e.X, Y: drop}, {X: busLeft, Y: drop},
			}})
			hasL = true
		} else {
			if drop < minDropR {
				minDropR = drop
			}
			b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
				e, {X: e.X, Y: drop}, {X: busRight, Y: drop},
			}})
			hasR = true
		}
	}

	if hasL {
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
			{X: busLeft, Y: minDropL}, {X: busLeft, Y: jy}, {X: cx, Y: jy},
		}})
	}
	if hasR {
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
			{X: busRight, Y: minDropR}, {X: busRight, Y: jy}, {X: cx, Y: jy},
		}})
	}
	b.d.Edges = append(b.d.Edges, edge(P(cx, jy), P(cx, kY)))
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
		case *ir.InfLoop:
			mapCalls(x.Body)
		case *ir.Block:
			mapCalls(x)
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
		if (nstmts(x.Then) == 0) != (nstmts(x.Else) == 0) { // guard: одна гілка порожня
			return max(diaW(x.Cond), max(tw, ew)+2*hGap), h
		}
		return max(diaW(x.Cond), tw+hGap+ew), h
	case *ir.For:
		bw, bh := b.blockSize(x.Body)
		w, h := max(hexW(x.Spec), bw)+2*arcGap, hexH+vGap+bh+vGap
		return b.withElse(x.Else, w, h)
	case *ir.While:
		bw, bh := b.blockSize(x.Body)
		w, h := max(diaW(x.Cond), bw)+2*arcGap, diaH+vGap+bh+vGap
		return b.withElse(x.Else, w, h)
	case *ir.DoWhile:
		bw, bh := b.blockSize(x.Body)
		return max(diaW(x.Cond), bw) + 2*arcGap, bh + vGap + diaH + vGap
	case *ir.InfLoop:
		bw, bh := b.blockSize(x.Body)
		return bw + 2*arcGap, bh + vGap
	case *ir.Break, *ir.Continue:
		return 0, 0 // без фігури
	case *ir.Connector:
		return 46, 46
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

// nstmts — кількість інструкцій блоку, безпечно для nil (IR можуть будувати
// вручну, лишивши Then/Else == nil).
func nstmts(blk *ir.Block) int {
	if blk == nil {
		return 0
	}
	return len(blk.Stmts)
}

// withElse доповнює габарити циклу розміром гілки else (for/while), якщо вона є.
func (b *build) withElse(els *ir.Block, w, h float64) (float64, float64) {
	if nstmts(els) == 0 {
		return w, h
	}
	ew, eh := b.blockSize(els)
	return max(w, ew), h + vGap + eh
}

// ensureBlocks гарантує, що жоден під-блок IR не nil (Then/Else/Body) — інакше
// глибше буде SIGSEGV. Безпечне доповнення до nil-перевірок у branchSize/blockSize:
// нормалізуємо дерево один раз на вході, далі всі звертання безпечні.
func ensureBlocks(blk *ir.Block) {
	fix := func(p **ir.Block) {
		if *p == nil {
			*p = &ir.Block{}
		}
		ensureBlocks(*p)
	}
	if blk == nil {
		return
	}
	for _, s := range blk.Stmts {
		switch x := s.(type) {
		case *ir.If:
			fix(&x.Then)
			fix(&x.Else)
		case *ir.For:
			fix(&x.Body)
			fix(&x.Else)
		case *ir.While:
			fix(&x.Body)
			fix(&x.Else)
		case *ir.DoWhile:
			fix(&x.Body)
		case *ir.InfLoop:
			fix(&x.Body)
		}
	}
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
		return b.placeFor(x, cx, top)
	case *ir.While:
		return b.placeWhile(x, cx, top)
	case *ir.DoWhile:
		return b.placeDoWhile(x, cx, top), false
	case *ir.InfLoop:
		return b.placeInfLoop(x, cx, top), false
	case *ir.Connector:
		const cw = 46.0
		b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Connector, X: cx - cw/2, Y: top, W: cw, H: cw, Text: x.Text})
		return P(cx, top+cw), false
	case *ir.Break:
		// стрибок на вихід циклу — без фігури; з'єднання зробить routeBreaks
		b.recordBreak(P(cx, top))
		return P(cx, top), true
	case *ir.Continue:
		// стрибок на наступну ітерацію — без фігури; з'єднання зробить routeContinues
		b.recordContinue(P(cx, top))
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
		// break/continue не мають фігури — з'єднання від попереднього виходу
		// зробить routeBreaks/routeContinues; стрілку-голову в нікуди не малюємо.
		if _, isBreak := s.(*ir.Break); isBreak {
			b.recordBreak(exit)
			return exit, true
		}
		if _, isCont := s.(*ir.Continue); isCont {
			b.recordContinue(exit)
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

// soleJump повертає Break/Continue, якщо блок — це рівно один такий стрибок.
func soleJump(blk *ir.Block) ir.Node {
	if nstmts(blk) == 1 {
		switch blk.Stmts[0].(type) {
		case *ir.Break, *ir.Continue:
			return blk.Stmts[0]
		}
	}
	return nil
}

// placeJumpGuard малює «if cond: break/continue» (else порожній): ромб-умова,
// стрибок (Так) — убік (break ліворуч до виходу циклу, continue праворуч до дуги
// повернення), а основний потік (Ні) — прямо вниз до наступної інструкції. Так
// стрибок не тягнеться по центру крізь подальші блоки тіла.
func (b *build) placeJumpGuard(cond string, jump ir.Node, cx, top float64) (diagram.Point, bool) {
	dw := diaW(cond)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: top, W: dw, H: diaH, Text: cond})
	midY := top + diaH/2
	_, isBreak := jump.(*ir.Break)
	side := 1.0 // continue → праворуч (до дуги повернення)
	if isBreak {
		side = -1 // break → ліворуч (до виходу циклу)
	}
	b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Label: b.yes, Points: []diagram.Point{
		{X: cx + side*dw/2, Y: midY}, {X: cx + side*(dw/2+hGap), Y: midY},
	}})
	p := P(cx+side*(dw/2+hGap), midY)
	if isBreak {
		b.recordBreak(p)
	} else {
		b.recordContinue(p)
	}
	return P(cx, top+diaH), false // Ні — прямо вниз (основний потік)
}

func (b *build) placeIf(n *ir.If, cx, top float64) (diagram.Point, bool) {
	// «if cond: break/continue» — стрибок убік, основний потік прямо вниз.
	if nstmts(n.Else) == 0 {
		if j := soleJump(n.Then); j != nil {
			return b.placeJumpGuard(n.Cond, j, cx, top)
		}
	}

	dw := diaW(n.Cond)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: top, W: dw, H: diaH, Text: n.Cond})
	midY := top + diaH/2
	branchTop := top + diaH + branchGap

	thenEmpty := nstmts(n.Then) == 0
	elseEmpty := nstmts(n.Else) == 0
	thenEnded := endsBlock(n.Then)
	elseEnded := endsBlock(n.Else)

	// Guard (лише одна гілка непорожня): дія прямо вниз, порожня обходить збоку.
	if elseEmpty && !thenEmpty && !thenEnded {
		return b.guard(n.Then, b.yes, b.no, +1, cx, dw, midY, branchTop)
	}
	if thenEmpty && !elseEmpty && !elseEnded {
		return b.guard(n.Else, b.no, b.yes, -1, cx, dw, midY, branchTop)
	}
	// Непорожня гілка ЗАКІНЧУЄТЬСЯ (break/return), друга порожня: завершальну дію
	// НЕ можна лишати на центральній осі (продовження зіткнеться з нею). Але й
	// бічна дужка злиття (двостороння гілка) тут зайва — дія назад не зливається.
	// Тож: дія йде ВБІК і вниз (break/return маршрутизують окремо), а ПРОДОВЖЕННЯ
	// (порожня гілка) — прямо вниз по центру. На один рівень «сходів» менше.
	if elseEmpty && !thenEmpty && thenEnded {
		return b.guardTerm(n.Then, b.no, b.yes, +1, cx, dw, midY, branchTop)
	}
	if thenEmpty && !elseEmpty && elseEnded {
		return b.guardTerm(n.Else, b.yes, b.no, -1, cx, dw, midY, branchTop)
	}

	// Обидві гілки (або обидві порожні) — розносимо по боках.
	tw, th := b.branchSize(n.Then)
	ew, eh := b.branchSize(n.Else)
	// Гарантуємо, що стрілки виходять з ромба назовні, а гілки не накладаються.
	thenCx := cx - max(dw/2+24, tw/2+hGap/2)
	elseCx := cx + max(dw/2+24, ew/2+hGap/2)
	mergeY := branchTop + max(th, eh) + mergeGap

	thenEnded = b.branch(n.Then, b.yes, cx, cx-dw/2, midY, thenCx, branchTop, mergeY)
	elseEnded = b.branch(n.Else, b.no, cx, cx+dw/2, midY, elseCx, branchTop, mergeY)
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
	sideX := cx + side*max(dw/2+24, bw/2+hGap)
	b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Label: sideLabel, Points: []diagram.Point{
		{X: vx, Y: midY}, {X: sideX, Y: midY}, {X: sideX, Y: mergeY}, {X: cx, Y: mergeY},
	}})
	if !ended {
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{exit, {X: cx, Y: mergeY}}})
	}
	return P(cx, mergeY), false // порожня гілка завжди дає продовження
}

// guardTerm малює «if cond: {дія; break/return}» з ПОРОЖНІМ else: завершальна дія
// (actLabel) — ВБІК і вниз; вона сама завершується (break/return маршрутизують
// окремо), тож НЕ зливається назад. Продовження (порожня гілка, contLabel) — прямо
// ВНИЗ по центру до точки нижче дії. На відміну від двосторонньої гілки, тут нема
// бічної дужки злиття для завершальної дії — на один рівень «сходів» менше.
// side=+1 — дію праворуч, -1 — ліворуч. Ромб уже розміщено в placeIf.
func (b *build) guardTerm(body *ir.Block, contLabel, actLabel string, side, cx, dw, midY, branchTop float64) (diagram.Point, bool) {
	bw, bh := b.blockSize(body)
	actCx := cx + side*max(dw/2+24, bw/2+hGap/2)
	// Дія: кут ромба → вбік → вниз у тіло (зі стрілкою у перший блок).
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: actLabel, Points: []diagram.Point{
		{X: cx + side*dw/2, Y: midY}, {X: actCx, Y: midY}, {X: actCx, Y: branchTop},
	}})
	b.placeBlock(body, actCx, branchTop) // тіло завершується само (break/return)
	// Продовження: ромб → прямо вниз по центру (без вістря — голову дасть наступне
	// ребро від точки злиття).
	mergeY := branchTop + bh + mergeGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Label: contLabel, Points: []diagram.Point{
		{X: cx, Y: midY + diaH/2}, {X: cx, Y: mergeY},
	}})
	return P(cx, mergeY), false
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

// termGuardLast — якщо останній стейтмент тіла це guard (if із порожнім else),
// повертає його. Тоді «Ні» цього guard природно зливається з дугою повернення
// циклу (вгору в заголовок), без зайвого довгого треку «вниз-вбік-вниз». Працює
// і коли дія завершується (break/return — «Ні» САМ є дугою), і коли ні (тоді
// дугою є вихід тіла, а «Ні» лише вливається в її колонку).
func termGuardLast(blk *ir.Block) *ir.If {
	if blk == nil || len(blk.Stmts) == 0 {
		return nil
	}
	g, ok := blk.Stmts[len(blk.Stmts)-1].(*ir.If)
	if ok && nstmts(g.Else) == 0 && nstmts(g.Then) > 0 {
		return g
	}
	return nil
}

// placeLoopGuard малює тіло-guard циклу: ромб-умова, Так→завершальна дія (вниз),
// Ні→дуга повернення ВГОРУ в заголовок (праворуч). Повертає вихід циклу.
// headHalf/headCy — піввисота-по-X і центр заголовка; headBottom — його низ;
// entryLabel — підпис ребра заголовок→ромб (для while це «Так»).
func (b *build) placeLoopGuard(g *ir.If, cx, headHalf, headCy, headBottom float64, startS, startE int, entryLabel, exitLabel string) (diagram.Point, float64) {
	dw := diaW(g.Cond)
	diaTop := headBottom + vGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: entryLabel, Points: []diagram.Point{{X: cx, Y: headBottom}, {X: cx, Y: diaTop}}})
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: diaTop, W: dw, H: diaH, Text: g.Cond})
	diaMidY := diaTop + diaH/2

	// Так → вниз → дія тіла.
	actTop := diaTop + diaH + branchGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: b.yes, Points: []diagram.Point{{X: cx, Y: diaTop + diaH}, {X: cx, Y: actTop}}})
	exit, ended := b.placeBlock(g.Then, cx, actTop)

	// Колонки дуг — за реальним краєм УСЬОГО тіла, щоб дуга повернення не різала
	// бічні обходи внутрішніх guard-ів. Але й не ближче за кут ЗАГОЛОВКА (він буває
	// ширший за тіло): інакше колона дуги стає майже впритул до вершини заголовка,
	// завершальний горизонтальний сегмент майже нульовий — і вістря «ламається».
	right, left := b.bodyExtent(startS, startE, cx)
	right = max(right, cx+headHalf)
	left = min(left, cx-headHalf)
	backX := right + arcGap
	if ended {
		// Дія завершується (break/return): назад вертатись нема чому, тож «Ні» САМ
		// є дугою повернення — правий кут ромба → ВГОРУ в правий кут заголовка.
		b.d.Edges = append(b.d.Edges, diagram.Edge{Label: b.no, Points: []diagram.Point{
			{X: cx + dw/2, Y: diaMidY}, {X: backX, Y: diaMidY}, {X: backX, Y: headCy}, {X: cx + headHalf, Y: headCy},
		}})
	} else {
		// Дія продовжується: дугою повернення є ВИХІД тіла (низ → backX → вгору в
		// заголовок), а «Ні» лише коротко вливається в ту саму колонку backX на рівні
		// ромба — без довгого треку вниз через усе тіло.
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Label: b.no, Points: []diagram.Point{
			{X: cx + dw/2, Y: diaMidY}, {X: backX, Y: diaMidY},
		}})
		drop := exit.Y + mergeGap/2
		b.d.Edges = append(b.d.Edges, diagram.Edge{Points: []diagram.Point{
			{X: cx, Y: exit.Y}, {X: cx, Y: drop}, {X: backX, Y: drop}, {X: backX, Y: headCy}, {X: cx + headHalf, Y: headCy},
		}})
	}
	// Вихід циклу: лівий кут заголовка → вниз нижче всього → центр.
	contY := contentBottom(b.d) + vGap
	leftX := left - arcGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Label: exitLabel, Points: []diagram.Point{
		{X: cx - headHalf, Y: headCy}, {X: leftX, Y: headCy}, {X: leftX, Y: contY}, {X: cx, Y: contY},
	}})
	return P(cx, contY), backX
}

// placeGuardLoopBody розкладає тіло циклу, що ЗАКІНЧУЄТЬСЯ guard-ом: спершу
// попередні інструкції (звичайно), тоді останній guard, чиє «Ні» — дуга
// повернення вгору в заголовок. Обробляє break усередині попередніх інструкцій.
func (b *build) placeGuardLoopBody(body *ir.Block, g *ir.If, cx, headHalf, headCy, headBottom float64, entryLabel, exitLabel string) diagram.Point {
	b.pushLoop()
	startS, startE := len(b.d.Shapes), len(b.d.Edges)
	fromY, el := headBottom, entryLabel
	if pre := body.Stmts[:len(body.Stmts)-1]; len(pre) > 0 {
		bodyTop := headBottom + vGap
		b.d.Edges = append(b.d.Edges, diagram.Edge{Label: entryLabel, Points: []diagram.Point{{X: cx, Y: headBottom}, {X: cx, Y: bodyTop}}})
		exit, _ := b.placeBlock(&ir.Block{Stmts: pre}, cx, bodyTop)
		fromY, el = exit.Y, ""
	}
	cont, backX := b.placeLoopGuard(g, cx, headHalf, headCy, fromY, startS, startE, el, exitLabel)
	brks, conts := b.popLoop()
	b.routeContinues(backX, cont.Y, conts)
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
func (b *build) placeFor(n *ir.For, cx, top float64) (diagram.Point, bool) {
	hw := hexW(n.Spec)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Hexagon, X: cx - hw/2, Y: top, W: hw, H: hexH, Text: n.Spec})
	headCy := top + hexH/2

	// Тіло закінчується guard-ом → його «Ні» одразу дугою вгору в заголовок.
	if g := termGuardLast(n.Body); g != nil {
		cont := b.placeGuardLoopBody(n.Body, g, cx, hw/2, headCy, top+hexH, "", "")
		return b.placeLoopElse(n.Else, cx, cont)
	}

	bodyTop := top + hexH + vGap
	b.d.Edges = append(b.d.Edges, edge(P(cx, top+hexH), P(cx, bodyTop)))
	startS, startE := len(b.d.Shapes), len(b.d.Edges)
	b.pushLoop()
	bodyExit, _ := b.placeBlock(n.Body, cx, bodyTop)
	brks, conts := b.popLoop()
	cont := b.loopArcs(cx, hw/2, headCy, startS, startE, bodyExit.Y, "", conts)
	cont, ended := b.placeLoopElse(n.Else, cx, cont) // for/else: після нормального виходу
	b.routeBreaks(cx, cont.Y, brks)                  // break оминає else
	return cont, ended
}

// placeWhile — цикл while: ромб-передумова, Так→тіло, дуга повернення справа,
// Ні→вихід зліва вниз.
func (b *build) placeWhile(n *ir.While, cx, top float64) (diagram.Point, bool) {
	dw := diaW(n.Cond)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: top, W: dw, H: diaH, Text: n.Cond})
	headCy := top + diaH/2

	if g := termGuardLast(n.Body); g != nil {
		cont := b.placeGuardLoopBody(n.Body, g, cx, dw/2, headCy, top+diaH, b.yes, b.no)
		return b.placeLoopElse(n.Else, cx, cont)
	}

	bodyTop := top + diaH + vGap
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: b.yes, Points: []diagram.Point{{X: cx, Y: top + diaH}, {X: cx, Y: bodyTop}}})
	startS, startE := len(b.d.Shapes), len(b.d.Edges)
	b.pushLoop()
	bodyExit, _ := b.placeBlock(n.Body, cx, bodyTop)
	brks, conts := b.popLoop()
	cont := b.loopArcs(cx, dw/2, headCy, startS, startE, bodyExit.Y, b.no, conts)
	cont, ended := b.placeLoopElse(n.Else, cx, cont)
	b.routeBreaks(cx, cont.Y, brks)
	return cont, ended
}

// placeDoWhile — цикл з післяумовою: тіло згори, ромб-умова знизу, Так→вихід,
// Ні→дуга повернення справа до лінії входу (повтор тіла).
func (b *build) placeDoWhile(n *ir.DoWhile, cx, top float64) diagram.Point {
	startS, startE := len(b.d.Shapes), len(b.d.Edges)
	b.pushLoop()
	bodyExit, _ := b.placeBlock(n.Body, cx, top)
	brks, conts := b.popLoop()

	diaTop := bodyExit.Y + vGap
	dw := diaW(n.Cond)
	b.d.Shapes = append(b.d.Shapes, diagram.Shape{Kind: diagram.Decision, X: cx - dw/2, Y: diaTop, W: dw, H: diaH, Text: n.Cond})
	diaCy := diaTop + diaH/2
	b.d.Edges = append(b.d.Edges, edge(bodyExit, P(cx, diaTop)))

	right, _ := b.bodyExtent(startS, startE, cx)
	backX := right + arcGap
	contY := diaTop + diaH + vGap
	b.routeContinues(backX, contY, conts) // continue → наступна ітерація (вгору) через дугу
	mergeY := top - vGap/2
	// Без вістря: вливається в лінію входу (не у фігуру) — щоб не було «двох голів».
	b.d.Edges = append(b.d.Edges, diagram.Edge{Label: b.no, Arrowless: true, Points: []diagram.Point{
		{X: cx + dw/2, Y: diaCy}, {X: backX, Y: diaCy}, {X: backX, Y: mergeY}, {X: cx, Y: mergeY},
	}})
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
	startS, startE := len(b.d.Shapes), len(b.d.Edges)
	b.pushLoop()
	bodyExit, _ := b.placeBlock(n.Body, cx, top)
	brks, conts := b.popLoop()

	// Безумовна дуга повернення справа: низ тіла → праворуч → вгору → вхід.
	// Із центру низу (трохи вниз) і без вістря (вливається в лінію входу).
	right, _ := b.bodyExtent(startS, startE, cx)
	backX := right + arcGap
	contY := bodyExit.Y + vGap
	b.routeContinues(backX, contY, conts)
	mergeY := top - vGap/2
	drop := bodyExit.Y + mergeGap/2
	b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
		{X: cx, Y: bodyExit.Y}, {X: cx, Y: drop}, {X: backX, Y: drop}, {X: backX, Y: mergeY}, {X: cx, Y: mergeY},
	}})

	b.routeBreaks(cx, contY, brks)
	return P(cx, contY)
}

// loopArcs малює дугу повернення (низ тіла → праворуч → вгору → правий кут
// заголовка) і дугу виходу (лівий кут заголовка → ліворуч → вниз → центр).
func (b *build) loopArcs(cx, headHalf, headCy float64, startS, startE int, bodyBottom float64, exitLabel string, conts []diagram.Point) diagram.Point {
	right, left := b.bodyExtent(startS, startE, cx)
	right = max(right, cx+headHalf) // дуга огинає і ширший за тіло заголовок (див. placeLoopGuard)
	left = min(left, cx-headHalf)
	backX := right + arcGap
	leftX := left - arcGap
	contY := bodyBottom + vGap
	b.routeContinues(backX, contY, conts) // continue → колонка дуги повернення → вгору
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

// placeLoopElse розкладає гілку for/else (while/else) — виконується після
// НОРМАЛЬНОГО завершення циклу. cont — точка виходу циклу; повертає нову точку
// продовження (нижче else). Якщо else порожній — повертає cont без змін.
// Break, що оминає else, маршрутизують ПІСЛЯ цього виклику (на новий cont).
func (b *build) placeLoopElse(els *ir.Block, cx float64, cont diagram.Point) (diagram.Point, bool) {
	if nstmts(els) == 0 {
		return cont, false
	}
	top := cont.Y + vGap
	b.d.Edges = append(b.d.Edges, edge(cont, P(cx, top)))
	return b.placeBlock(els, cx, top)
}

// --- дрібні помічники ---

func P(x, y float64) diagram.Point { return diagram.Point{X: x, Y: y} }

func term(cx, top float64, text string) diagram.Shape {
	return diagram.Shape{Kind: diagram.Terminator, X: cx - termW/2, Y: top, W: termW, H: termH, Text: text}
}

func edge(a, b diagram.Point) diagram.Edge {
	return diagram.Edge{Points: []diagram.Point{a, b}}
}

// shiftX зсуває всі фігури й точки ребер по X (для нормалізації полотна).
func (b *build) shiftX(dx float64) {
	for i := range b.d.Shapes {
		b.d.Shapes[i].X += dx
	}
	for i := range b.d.Edges {
		for j := range b.d.Edges[i].Points {
			b.d.Edges[i].Points[j].X += dx
		}
	}
	for i := range b.ends {
		b.ends[i].X += dx
	}
}

// bodyExtent — найправіша/найлівіша X серед фігур і ребер, доданих ПІСЛЯ маркера
// (startS, startE). Дає реальну ширину тіла з урахуванням бічних обходів (Ні-гілки
// внутрішніх guard-ів тощо), щоб дуги циклу огинали ВСЕ тіло й не різали внутрішні лінії.
// bodyExtent — правий/лівий край вмісту, доданого після маркера (startS, startE).
// fallback — значення, якщо вмісту нема (порожнє тіло циклу, напр. «for: pass»):
// БЕЗ нього повертався б ±Inf → координата -Inf → json.Marshal тихо падає й на
// виході порожній файл. Тому стартуємо з fallback (зазвичай cx), а не з Inf.
func (b *build) bodyExtent(startS, startE int, fallback float64) (right, left float64) {
	right, left = fallback, fallback
	for _, s := range b.d.Shapes[startS:] {
		right, left = max(right, s.X+s.W), min(left, s.X)
	}
	for _, e := range b.d.Edges[startE:] {
		for _, p := range e.Points {
			right, left = max(right, p.X), min(left, p.X)
		}
	}
	return
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
