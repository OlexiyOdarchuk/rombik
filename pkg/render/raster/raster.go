// Пакет raster малює diagram.Diagram у растрові/векторні формати (PNG, PDF)
// НАТИВНО в Go через tdewolff/canvas — без зовнішніх бінарників (rsvg/typst).
// Шрифт (кирилиця) вшито в бінарник. Геометрія — та сама, що в SVG-рендері.
package raster

import (
	"bytes"
	_ "embed"
	"image/color"
	"math"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	pdfrender "github.com/tdewolff/canvas/renderers/pdf"
)

//go:embed font.ttf
var fontTTF []byte

// Кольори — ті самі, що в SVG-рендері.
var (
	ink      = color.RGBA{0x22, 0x22, 0x22, 0xff} // лінії/обведення
	fillCol  = color.RGBA{0xfd, 0xfd, 0xfd, 0xff} // заливка фігур
	textCol  = color.RGBA{0x11, 0x11, 0x11, 0xff} // текст у фігурах
	labelCol = color.RGBA{0x44, 0x44, 0x44, 0xff} // підписи ребер
	whiteCol = color.RGBA{0xff, 0xff, 0xff, 0xff}
	noneCol  = color.RGBA{0, 0, 0, 0}
)

// 1 діаграмний пункт → мм (трактуємо одиниці діаграми як типографські пункти).
const ptPerUnit = 25.4 / 72.0

// capGap — місце під підпис «Рисунок N — …» знизу (як у SVG).
const capGap = 30.0

type rnd struct {
	ctx       *canvas.Context
	h         float64 // висота діаграми (для перевороту осі Y)
	main, lbl *canvas.FontFace
}

// x,y переводять координати діаграми (Y вниз) у canvas-мм (Y вгору).
func (r *rnd) x(v float64) float64 { return v * ptPerUnit }
func (r *rnd) y(v float64) float64 { return (r.h - v) * ptPerUnit }

// PNG растеризує діаграму (scale — пікселів на одиницю; типово 3).
func PNG(d *diagram.Diagram, scale float64) ([]byte, error) {
	c, err := build(d)
	if err != nil {
		return nil, err
	}
	if scale <= 0 {
		scale = 3
	}
	var buf bytes.Buffer
	if err := c.Write(&buf, renderers.PNG(canvas.DPI(72*scale))); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PDF малює діаграму у векторний PDF (сторінка у пунктах = розмір діаграми).
func PDF(d *diagram.Diagram) ([]byte, error) {
	c, err := build(d)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := c.Write(&buf, renderers.PDF()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PDFAll малює всі схеми в ОДИН багатосторінковий PDF (кожна — окрема сторінка
// свого розміру). Для експорту цілого звіту одним файлом.
func PDFAll(ds []*diagram.Diagram) ([]byte, error) {
	if len(ds) == 0 {
		return PDF(&diagram.Diagram{W: 100, H: 100})
	}
	var buf bytes.Buffer
	var doc *pdfrender.PDF
	for i, d := range ds {
		c, err := build(d)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			doc = pdfrender.New(&buf, c.W, c.H, nil) // перша сторінка
		} else {
			doc.NewPage(c.W, c.H) // наступна
		}
		c.RenderTo(doc)
	}
	if err := doc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func build(d *diagram.Diagram) (*canvas.Canvas, error) {
	fam := canvas.NewFontFamily("rombik")
	if err := fam.LoadFont(fontTTF, 0, canvas.FontRegular); err != nil {
		return nil, err
	}
	cap := d.CaptionLine()
	totalH := d.H
	if cap != "" {
		totalH += capGap // запас знизу під підпис
	}
	c := canvas.New(d.W*ptPerUnit, totalH*ptPerUnit)
	ctx := canvas.NewContext(c)
	ctx.SetStrokeCapper(canvas.RoundCap)
	ctx.SetStrokeJoiner(canvas.RoundJoin)
	r := &rnd{
		ctx:  ctx,
		h:    totalH,
		main: fam.Face(14, textCol, canvas.FontRegular, canvas.FontNormal),
		lbl:  fam.Face(12, labelCol, canvas.FontRegular, canvas.FontNormal),
	}
	// Білий фон.
	ctx.SetFillColor(whiteCol)
	ctx.SetStrokeColor(noneCol)
	ctx.DrawPath(0, 0, canvas.Rectangle(c.W, c.H))
	// Спершу ребра (під фігурами), потім фігури.
	for _, e := range d.Edges {
		r.edge(e)
	}
	for _, s := range d.Shapes {
		r.shape(s)
	}
	// Підпис схеми («Рисунок N — …») — по центру під схемою.
	if cap != "" {
		t := canvas.NewTextBox(r.main, cap, d.W*ptPerUnit, capGap*ptPerUnit, canvas.Center, canvas.Center, nil)
		r.ctx.DrawText(r.x(0), r.y(d.H), t)
	}
	return c, nil
}

func (r *rnd) shape(s diagram.Shape) {
	r.ctx.SetFillColor(fillCol)
	r.ctx.SetStrokeColor(ink)
	r.ctx.SetStrokeWidth(1.5 * ptPerUnit)
	cx, cy := s.X+s.W/2, s.Y+s.H/2
	switch s.Kind {
	case diagram.Terminator:
		r.ctx.DrawPath(r.x(s.X), r.y(s.Y+s.H), canvas.RoundedRectangle(s.W*ptPerUnit, s.H*ptPerUnit, (s.H/2)*ptPerUnit))
	case diagram.Process:
		r.ctx.DrawPath(r.x(s.X), r.y(s.Y+s.H), canvas.Rectangle(s.W*ptPerUnit, s.H*ptPerUnit))
	case diagram.Decision:
		r.poly([][2]float64{{cx, s.Y}, {s.X + s.W, cy}, {cx, s.Y + s.H}, {s.X, cy}})
	case diagram.InOut:
		sk := s.H * 0.4
		r.poly([][2]float64{{s.X + sk, s.Y}, {s.X + s.W, s.Y}, {s.X + s.W - sk, s.Y + s.H}, {s.X, s.Y + s.H}})
	case diagram.Hexagon:
		sk := s.H * 0.5
		r.poly([][2]float64{{s.X + sk, s.Y}, {s.X + s.W - sk, s.Y}, {s.X + s.W, cy}, {s.X + s.W - sk, s.Y + s.H}, {s.X + sk, s.Y + s.H}, {s.X, cy}})
	case diagram.Predef:
		r.ctx.DrawPath(r.x(s.X), r.y(s.Y+s.H), canvas.Rectangle(s.W*ptPerUnit, s.H*ptPerUnit))
		const in = 9.0 // бокові риски
		r.line(s.X+in, s.Y, s.X+in, s.Y+s.H)
		r.line(s.X+s.W-in, s.Y, s.X+s.W-in, s.Y+s.H)
	}
	r.centerText(s.Text, s.X, s.Y, s.W, s.H)
}

// poly малює заповнений+обведений багатокутник за точками діаграми.
func (r *rnd) poly(pts [][2]float64) {
	r.ctx.MoveTo(r.x(pts[0][0]), r.y(pts[0][1]))
	for _, p := range pts[1:] {
		r.ctx.LineTo(r.x(p[0]), r.y(p[1]))
	}
	r.ctx.Close()
	r.ctx.FillStroke()
}

// line — самотній штрих (для бокових рисок підпрограми).
func (r *rnd) line(x1, y1, x2, y2 float64) {
	r.ctx.SetFillColor(noneCol)
	r.ctx.SetStrokeColor(ink)
	r.ctx.MoveTo(r.x(x1), r.y(y1))
	r.ctx.LineTo(r.x(x2), r.y(y2))
	r.ctx.Stroke()
}

func (r *rnd) centerText(s string, X, Y, W, H float64) {
	if s == "" {
		return
	}
	t := canvas.NewTextBox(r.main, s, W*ptPerUnit, H*ptPerUnit, canvas.Center, canvas.Center, nil)
	r.ctx.DrawText(r.x(X), r.y(Y), t)
}

func (r *rnd) edge(e diagram.Edge) {
	if len(e.Points) < 2 {
		return
	}
	r.ctx.SetFillColor(noneCol)
	r.ctx.SetStrokeColor(ink)
	r.ctx.SetStrokeWidth(1.5 * ptPerUnit)
	r.ctx.MoveTo(r.x(e.Points[0].X), r.y(e.Points[0].Y))
	for _, p := range e.Points[1:] {
		r.ctx.LineTo(r.x(p.X), r.y(p.Y))
	}
	r.ctx.Stroke()
	if !e.Arrowless {
		n := len(e.Points)
		r.arrow(e.Points[n-2], e.Points[n-1])
	}
	if e.Label != "" {
		r.label(e.Label, e.Points[0], e.Points[1])
	}
}

// arrow малює заповнене вістря у точці b (напрям a→b).
func (r *rnd) arrow(a, b diagram.Point) {
	dx, dy := b.X-a.X, b.Y-a.Y
	l := math.Hypot(dx, dy)
	if l == 0 {
		return
	}
	ux, uy := dx/l, dy/l
	const al, aw = 11.0, 4.0
	bx, by := b.X-ux*al, b.Y-uy*al
	px, py := -uy, ux
	r.ctx.SetFillColor(ink)
	r.ctx.SetStrokeColor(noneCol)
	r.ctx.MoveTo(r.x(b.X), r.y(b.Y))
	r.ctx.LineTo(r.x(bx+px*aw), r.y(by+py*aw))
	r.ctx.LineTo(r.x(bx-px*aw), r.y(by-py*aw))
	r.ctx.Close()
	r.ctx.Fill()
}

// label ставить підпис (Так/Ні) біля початку ребра — як у SVG-рендері.
func (r *rnd) label(s string, p0, p1 diagram.Point) {
	const boxW = 60.0
	halign := canvas.Left
	lx, ly := p0.X+6, p0.Y-7
	switch {
	case p1.X == p0.X: // вертикальний сегмент — збоку, посередині
		ly = (p0.Y + p1.Y) / 2
	case p1.X < p0.X: // вліво — текст назовні зліва
		halign = canvas.Right
		lx, ly = p0.X-6-boxW, p0.Y-7
	}
	t := canvas.NewTextBox(r.lbl, s, boxW*ptPerUnit, 14*ptPerUnit, halign, canvas.Center, nil)
	r.ctx.DrawText(r.x(lx), r.y(ly)+7*ptPerUnit, t)
}
