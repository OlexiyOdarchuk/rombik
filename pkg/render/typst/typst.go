// Пакет typst малює diagram.Diagram у нативний Typst (через CeTZ). На відміну
// від SVG, результат — це Typst-код, що компілюється у векторну графіку без
// зовнішніх картинок. Координати точні (ті самі, що в SVG), тож вигляд збігається.
package typst

import (
	"fmt"
	"strings"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
)

// Render повертає САМОДОСТАТНІЙ Typst-документ: сторінка авто-розміру (щільно
// огортає схему — не A4 з обрізанням), за потреби — підпис «Рисунок N» через
// figure. Береш цей код і одразу компілюєш у тісний PDF.
func Render(d *diagram.Diagram) string {
	var b strings.Builder
	preamble(&b, d)
	figure(&b, d)
	return b.String()
}

// RenderAll повертає ОДИН Typst-документ з усіма схемами (кожна — окрема
// сторінка авто-розміру, figure з авто-нумерацією «Рисунок 1, 2, …»). Слово й
// роздільник підпису беремо з першої схеми (вони глобальні).
func RenderAll(ds []*diagram.Diagram) string {
	if len(ds) == 0 {
		return ""
	}
	var b strings.Builder
	preamble(&b, ds[0])
	for i, d := range ds {
		if i > 0 {
			b.WriteString("#pagebreak()\n")
		}
		figure(&b, d)
	}
	return b.String()
}

// preamble — шапка документа: import, авто-сторінка, шрифт, помічник #flowchart.
func preamble(b *strings.Builder, d *diagram.Diagram) {
	b.WriteString(`#import "@preview/cetz:0.3.4"` + "\n")
	b.WriteString("#set page(width: auto, height: auto, margin: 14pt)\n")
	b.WriteString("#set text(" + font + ")\n")
	fmt.Fprintf(b, "#set figure.caption(separator: [%s])\n", d.CapSeparator()) // напр. « — »
	supplement := "[" + d.CapSupplement() + "]"
	if !d.CapHasWord() { // шаблон без {word} — без слова-supplement
		supplement = "none"
	}
	b.WriteString(`#let flowchart(body, caption: none) = figure(` + "\n" +
		"  body, caption: caption, supplement: " + supplement + `, kind: "flowchart", numbering: "1",` + "\n)\n")
}

// figure — одна схема: cetz.canvas, за потреби загорнутий у #flowchart-підпис.
func figure(b *strings.Builder, d *diagram.Diagram) {
	canvas := renderCanvas(d)
	if d.Caption != "" {
		// #%q — рядкова форма, щоб спецсимволи Typst у підписі не ламали документ.
		fmt.Fprintf(b, "#flowchart(caption: [#%q])[\n%s]\n", d.Caption, canvas)
	} else {
		b.WriteString(canvas)
	}
}

// Fragment — лише блок cetz.canvas (без преамбули/сторінки/підпису), щоб
// вставити у свій .typ-документ (де import cetz уже є). Зручно для звіту.
func Fragment(d *diagram.Diagram) string { return renderCanvas(d) }

// FragmentAll — кілька canvas-фрагментів поспіль (через порожній рядок).
func FragmentAll(ds []*diagram.Diagram) string {
	var b strings.Builder
	for i, d := range ds {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(renderCanvas(d))
	}
	return b.String()
}

// renderCanvas — сам блок cetz.canvas зі схемою (без сторінки/підпису).
func renderCanvas(d *diagram.Diagram) string {
	var b strings.Builder
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
			lx, ly, align := diagram.LabelAnchor(e.Points[0], e.Points[1])
			anchor := "center"
			switch align {
			case "start":
				anchor = "west"
			case "end":
				anchor = "east"
			}
			fmt.Fprintf(&b, "  content((%.1f, %.1f), text(%ssize: 12pt)[#%q], anchor: %q)\n", lx, fy(ly), font, e.Label, anchor)
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
	case diagram.Connector:
		r := s.W / 2
		if s.H < s.W {
			r = s.H / 2
		}
		fmt.Fprintf(b, "  circle((%.1f, %.1f), radius: %.1f%s)\n", cx, fy(cy), r, fill)
	}
	// Текст по центру (рядкова форма #%q — щоб спецсимволи [ ] * не ламали Typst).
	fmt.Fprintf(b, "  content((%.1f, %.1f), text(%ssize: 14pt)[#%q])\n", cx, fy(cy), font, s.Text)
}

// font — Times New Roman (вимога курсових), із serif-фолбеками; перший наявний виграє.
const font = `font: ("Times New Roman", "Liberation Serif", "DejaVu Serif"), `
