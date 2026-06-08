// Пакет excalidraw експортує diagram.Diagram у формат Excalidraw (.excalidraw,
// JSON) — щоб схему можна було відкрити й доредагувати на excalidraw.com.
// Кожна фігура → елемент (rectangle/diamond/ellipse/line) + текст у тій самій
// групі; ребро → arrow. Координати ті самі, що в SVG (вісь y вниз).
package excalidraw

import (
	"encoding/json"
	"fmt"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
)

const (
	stroke  = "#1e1e1e"
	updated = 1700000000000 // фіксований час (у WASM нема годинника) — Excalidraw байдуже
	gap     = 44.0          // проміжок між схемами у RenderAll
)

// Render повертає .excalidraw-документ (JSON) для однієї схеми.
func Render(d *diagram.Diagram) string {
	b := &builder{}
	b.diagram(d, 0, 0)
	return doc(b.els)
}

// RenderAll складає всі схеми в один .excalidraw (вертикально, як у SVG-all).
func RenderAll(ds []*diagram.Diagram) string {
	b := &builder{}
	y := 0.0
	for _, d := range ds {
		b.diagram(d, 0, y)
		y += diagHeight(d) + gap
	}
	return doc(b.els)
}

func doc(els []any) string {
	m := map[string]any{
		"type":     "excalidraw",
		"version":  2,
		"source":   "https://github.com/OlexiyOdarchuk/rombik",
		"elements": els,
		"appState": map[string]any{"gridSize": nil, "viewBackgroundColor": "#ffffff"},
		"files":    map[string]any{},
	}
	bb, _ := json.MarshalIndent(m, "", "  ")
	return string(bb)
}

func diagHeight(d *diagram.Diagram) float64 {
	if d.CaptionLine() != "" {
		return d.H + 30
	}
	return d.H
}

type builder struct {
	n   int
	els []any
}

func (b *builder) id() string { b.n++; return fmt.Sprintf("r%d", b.n) }

// base — спільні поля елемента Excalidraw.
func (b *builder) base(typ string, x, y, w, h float64, group string) map[string]any {
	seed := b.n * 100003 // детермінований (без рандому — щоб працювало в WASM)
	gids := []string{}
	if group != "" {
		gids = []string{group}
	}
	return map[string]any{
		"id": b.id(), "type": typ, "x": x, "y": y, "width": w, "height": h,
		"angle": 0, "strokeColor": stroke, "backgroundColor": "transparent",
		"fillStyle": "solid", "strokeWidth": 1.5, "strokeStyle": "solid",
		"roughness": 1, "opacity": 100, "groupIds": gids, "frameId": nil,
		"roundness": nil, "seed": seed, "version": 1, "versionNonce": seed,
		"isDeleted": false, "boundElements": nil, "updated": updated, "link": nil, "locked": false,
	}
}

func (b *builder) add(m map[string]any) { b.els = append(b.els, m) }

// text — окремий текстовий елемент по центру (cx,cy), у групі group.
func (b *builder) text(s string, cx, cy float64, group string, size float64) {
	if s == "" {
		return
	}
	w := float64(len([]rune(s)))*size*0.55 + 6
	h := size * 1.25
	t := b.base("text", cx-w/2, cy-h/2, w, h, group)
	t["text"] = s
	t["fontSize"] = size
	t["fontFamily"] = 1 // Excalifont (рукописний; сучасний має кирилицю)
	t["textAlign"] = "center"
	t["verticalAlign"] = "middle"
	t["containerId"] = nil
	t["originalText"] = s
	t["lineHeight"] = 1.25
	b.add(t)
}

func (b *builder) diagram(d *diagram.Diagram, ox, oy float64) {
	for _, e := range d.Edges {
		b.arrow(e, ox, oy)
	}
	for _, s := range d.Shapes {
		b.shape(s, ox, oy)
	}
	if cap := d.CaptionLine(); cap != "" {
		b.text(cap, ox+d.W/2, oy+d.H+15, "", 16)
	}
}

func (b *builder) shape(s diagram.Shape, ox, oy float64) {
	x, y, w, h := s.X+ox, s.Y+oy, s.W, s.H
	cx, cy := x+w/2, y+h/2
	g := fmt.Sprintf("g%d", b.n) // спільна група для фігури й її тексту
	switch s.Kind {
	case diagram.Process:
		b.add(b.base("rectangle", x, y, w, h, g))
	case diagram.Decision:
		b.add(b.base("diamond", x, y, w, h, g))
	case diagram.Terminator:
		m := b.base("rectangle", x, y, w, h, g)
		m["roundness"] = map[string]any{"type": 3} // округлений (≈ стадіон)
		b.add(m)
	case diagram.Connector:
		r := w
		if h < w {
			r = h
		}
		b.add(b.base("ellipse", cx-r/2, cy-r/2, r, r, g))
	case diagram.InOut:
		sk := h * 0.4
		b.polygon(x, y, w, h, [][2]float64{{sk, 0}, {w, 0}, {w - sk, h}, {0, h}}, g)
	case diagram.Hexagon:
		sk := h * 0.5
		b.polygon(x, y, w, h, [][2]float64{{sk, 0}, {w - sk, 0}, {w, h / 2}, {w - sk, h}, {sk, h}, {0, h / 2}}, g)
	case diagram.Predef:
		b.add(b.base("rectangle", x, y, w, h, g))
		b.lineSeg(x+9, y, x+9, y+h, g)
		b.lineSeg(x+w-9, y, x+w-9, y+h, g)
	}
	b.text(s.Text, cx, cy, g, 16)
}

// polygon — закрита ламана (line) як фігура; pts відносні до (x,y).
func (b *builder) polygon(x, y, w, h float64, pts [][2]float64, g string) {
	m := b.base("line", x, y, w, h, g)
	p := make([][2]float64, 0, len(pts)+1)
	p = append(p, pts...)
	p = append(p, pts[0]) // замикаємо
	m["points"] = p
	m["lastCommittedPoint"] = nil
	b.add(m)
}

func (b *builder) lineSeg(x1, y1, x2, y2 float64, g string) {
	m := b.base("line", x1, y1, x2-x1, y2-y1, g)
	m["points"] = [][2]float64{{0, 0}, {x2 - x1, y2 - y1}}
	m["lastCommittedPoint"] = nil
	b.add(m)
}

func (b *builder) arrow(e diagram.Edge, ox, oy float64) {
	if len(e.Points) < 2 {
		return
	}
	x0, y0 := e.Points[0].X+ox, e.Points[0].Y+oy
	var minX, minY, maxX, maxY float64 = 0, 0, 0, 0
	pts := make([][2]float64, len(e.Points))
	for i, p := range e.Points {
		dx, dy := (p.X+ox)-x0, (p.Y+oy)-y0
		pts[i] = [2]float64{dx, dy}
		minX, minY = min(minX, dx), min(minY, dy)
		maxX, maxY = max(maxX, dx), max(maxY, dy)
	}
	m := b.base("arrow", x0, y0, maxX-minX, maxY-minY, "")
	m["points"] = pts
	m["lastCommittedPoint"] = nil
	m["startBinding"] = nil
	m["endBinding"] = nil
	m["startArrowhead"] = nil
	if e.Arrowless {
		m["endArrowhead"] = nil
	} else {
		m["endArrowhead"] = "arrow"
	}
	b.add(m)
	if e.Label != "" && len(e.Points) >= 2 {
		lx, ly, _ := diagram.LabelAnchor(e.Points[0], e.Points[1])
		b.text(e.Label, lx, ly, "", 14)
	}
}
