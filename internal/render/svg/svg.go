// Пакет svg малює diagram.Diagram у SVG. Це адаптер рендера: залежить лише від
// diagram. Фігури — точні ДСТУ-примітиви, повний контроль до пікселя.
// (Другий адаптер — Typst/fletcher — додамо пізніше тим самим інтерфейсом.)
package svg

import (
	"fmt"
	"strings"

	"flowgen/internal/diagram"
)

// Render повертає SVG-рядок для діаграми.
func Render(d *diagram.Diagram) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" font-family="Arial, sans-serif" font-size="14">`,
		d.W, d.H, d.W, d.H)
	// Маркер-стрілка.
	b.WriteString(`<defs><marker id="arr" markerWidth="9" markerHeight="9" refX="7.5" refY="3" orient="auto">` +
		`<path d="M0,0 L8,3 L0,6 Z" fill="#222"/></marker></defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#ffffff"/>`)

	// Спершу ребра (під фігурами), потім фігури.
	for _, e := range d.Edges {
		marker := ` marker-end="url(#arr)"`
		if e.Arrowless {
			marker = ""
		}
		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="#222" stroke-width="1.5"%s/>`,
			pathOf(e.Points), marker)
		if e.Label != "" && len(e.Points) >= 2 {
			lx := (e.Points[0].X+e.Points[1].X)/2 + 4
			ly := (e.Points[0].Y+e.Points[1].Y)/2 - 5
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="#444">%s</text>`, lx, ly, esc(e.Label))
		}
	}
	for _, s := range d.Shapes {
		renderShape(&b, s)
	}
	b.WriteString(`</svg>`)
	return b.String()
}

func renderShape(b *strings.Builder, s diagram.Shape) {
	const stroke = `fill="#fdfdfd" stroke="#222" stroke-width="1.5"`
	switch s.Kind {
	case diagram.Terminator:
		fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" %s/>`,
			s.X, s.Y, s.W, s.H, s.H/2, stroke)
	case diagram.Process:
		fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" %s/>`, s.X, s.Y, s.W, s.H, stroke)
	case diagram.Decision:
		cx, cy := s.X+s.W/2, s.Y+s.H/2
		fmt.Fprintf(b, `<polygon points="%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f" %s/>`,
			cx, s.Y, s.X+s.W, cy, cx, s.Y+s.H, s.X, cy, stroke)
	case diagram.InOut:
		sk := s.H * 0.4 // нахил паралелограма
		fmt.Fprintf(b, `<polygon points="%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f" %s/>`,
			s.X+sk, s.Y, s.X+s.W, s.Y, s.X+s.W-sk, s.Y+s.H, s.X, s.Y+s.H, stroke)
	}
	// Текст по центру фігури.
	fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" dominant-baseline="middle" fill="#111">%s</text>`,
		s.X+s.W/2, s.Y+s.H/2, esc(s.Text))
}

// pathOf будує SVG-шлях із ламаної.
func pathOf(pts []diagram.Point) string {
	var b strings.Builder
	for i, p := range pts {
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&b, "%s%.1f %.1f ", cmd, p.X, p.Y)
	}
	return strings.TrimSpace(b.String())
}

func esc(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
