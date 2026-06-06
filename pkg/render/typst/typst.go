// Пакет typst малює diagram.Diagram у нативний Typst (через CeTZ). На відміну
// від SVG, результат — це Typst-код, що компілюється у векторну графіку без
// зовнішніх картинок. Координати точні (ті самі, що в SVG), тож вигляд збігається.
package typst

import (
	"fmt"
	"strings"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
)

// Render повертає самодостатній Typst-фрагмент (import + cetz.canvas).
func Render(d *diagram.Diagram) string {
	var b strings.Builder
	b.WriteString(`#import "@preview/cetz:0.3.4"` + "\n")
	b.WriteString("#cetz.canvas(length: 1pt, {\n")
	b.WriteString("  import cetz.draw: *\n")
	b.WriteString("  set-style(stroke: 1.5pt, fill: none)\n")

	fy := func(y float64) float64 { return d.H - y } // CeTZ: вісь y вгору

	for _, e := range d.Edges {
		pts := make([]string, len(e.Points))
		for i, p := range e.Points {
			pts[i] = fmt.Sprintf("(%.1f, %.1f)", p.X, fy(p.Y))
		}
		mark := ""
		if !e.Arrowless {
			mark = `, mark: (end: ">")`
		}
		fmt.Fprintf(&b, "  line(%s%s)\n", strings.Join(pts, ", "), mark)
		if e.Label != "" && len(e.Points) >= 2 {
			p := e.Points[0]
			fmt.Fprintf(&b, "  content((%.1f, %.1f), text(%ssize: 12pt)[#%q])\n", p.X, fy(p.Y)+9, font, e.Label)
		}
	}
	for _, s := range d.Shapes {
		shape(&b, s, fy)
	}

	b.WriteString("})\n")
	return b.String()
}

func shape(b *strings.Builder, s diagram.Shape, fy func(float64) float64) {
	x1, y1, x2, y2 := s.X, s.Y, s.X+s.W, s.Y+s.H
	cx, cy := s.X+s.W/2, s.Y+s.H/2
	const fill = ", fill: white"
	poly := func(pts ...[2]float64) {
		parts := make([]string, len(pts))
		for i, p := range pts {
			parts[i] = fmt.Sprintf("(%.1f, %.1f)", p[0], p[1])
		}
		fmt.Fprintf(b, "  line(%s, close: true%s)\n", strings.Join(parts, ", "), fill)
	}

	switch s.Kind {
	case diagram.Terminator:
		fmt.Fprintf(b, "  rect((%.1f, %.1f), (%.1f, %.1f), radius: %.1f%s)\n", x1, fy(y2), x2, fy(y1), s.H/2, fill)
	case diagram.Process:
		fmt.Fprintf(b, "  rect((%.1f, %.1f), (%.1f, %.1f)%s)\n", x1, fy(y2), x2, fy(y1), fill)
	case diagram.Decision:
		poly([2]float64{cx, fy(y1)}, [2]float64{x2, fy(cy)}, [2]float64{cx, fy(y2)}, [2]float64{x1, fy(cy)})
	case diagram.InOut:
		sk := s.H * 0.4
		poly([2]float64{x1 + sk, fy(y1)}, [2]float64{x2, fy(y1)}, [2]float64{x2 - sk, fy(y2)}, [2]float64{x1, fy(y2)})
	case diagram.Hexagon:
		sk := s.H * 0.5
		poly([2]float64{x1 + sk, fy(y1)}, [2]float64{x2 - sk, fy(y1)}, [2]float64{x2, fy(cy)},
			[2]float64{x2 - sk, fy(y2)}, [2]float64{x1 + sk, fy(y2)}, [2]float64{x1, fy(cy)})
	case diagram.Predef:
		fmt.Fprintf(b, "  rect((%.1f, %.1f), (%.1f, %.1f)%s)\n", x1, fy(y2), x2, fy(y1), fill)
		fmt.Fprintf(b, "  line((%.1f, %.1f), (%.1f, %.1f))\n", x1+9, fy(y1), x1+9, fy(y2))
		fmt.Fprintf(b, "  line((%.1f, %.1f), (%.1f, %.1f))\n", x2-9, fy(y1), x2-9, fy(y2))
	}
	// Текст по центру (рядкова форма #%q — щоб спецсимволи [ ] * не ламали Typst).
	fmt.Fprintf(b, "  content((%.1f, %.1f), text(%ssize: 14pt)[#%q])\n", cx, fy(cy), font, s.Text)
}

// font — sans-шрифт (як у SVG), з фолбеками; перші наявні в системі виграють.
const font = `font: ("Arial", "Liberation Sans", "DejaVu Sans", "Helvetica"), `
