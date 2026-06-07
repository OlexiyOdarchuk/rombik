// Пакет svg малює diagram.Diagram у SVG. Це адаптер рендера: залежить лише від
// diagram. Фігури — точні ДСТУ-примітиви, повний контроль до пікселя.
package svg

import (
	"fmt"
	"strings"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
)

// capGap — висота, що резервується під підпис «Рисунок N — …» знизу.
const capGap = 30

// diagHeight — повна висота схеми разом із місцем під підпис.
func diagHeight(d *diagram.Diagram) float64 {
	if d.CaptionLine() != "" {
		return d.H + capGap
	}
	return d.H
}

const arrowDefs = `<defs><marker id="arr" markerWidth="9" markerHeight="9" refX="7.5" refY="3" orient="auto">` +
	`<path d="M0,0 L8,3 L0,6 Z" fill="#222"/></marker></defs>`

// Render повертає SVG-рядок для діаграми.
func Render(d *diagram.Diagram) string {
	var b strings.Builder
	totalH := diagHeight(d)
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" font-family=%s font-size="14">`,
		d.W, totalH, d.W, totalH, fontAttr)
	b.WriteString(arrowDefs)
	b.WriteString(`<rect width="100%" height="100%" fill="#ffffff"/>`)
	writeBody(&b, d)
	b.WriteString(`</svg>`)
	return b.String()
}

// fontAttr — шрифт схеми: Times New Roman (вимога курсових), із serif-фолбеками.
const fontAttr = `"'Times New Roman', 'Liberation Serif', 'DejaVu Serif', serif"`

// RenderAll складає всі схеми в ОДИН SVG (вертикально, по центру, із проміжком).
func RenderAll(ds []*diagram.Diagram) string {
	if len(ds) == 0 {
		return ""
	}
	if len(ds) == 1 {
		return Render(ds[0])
	}
	const gap = 44.0
	var maxW, totalH float64
	for i, d := range ds {
		if d.W > maxW {
			maxW = d.W
		}
		totalH += diagHeight(d)
		if i > 0 {
			totalH += gap
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" font-family=%s font-size="14">`,
		maxW, totalH, maxW, totalH, fontAttr)
	b.WriteString(arrowDefs)
	b.WriteString(`<rect width="100%" height="100%" fill="#ffffff"/>`)
	y := 0.0
	for _, d := range ds {
		dx := (maxW - d.W) / 2 // центруємо по горизонталі
		fmt.Fprintf(&b, `<g transform="translate(%.1f,%.1f)">`, dx, y)
		writeBody(&b, d)
		b.WriteString(`</g>`)
		y += diagHeight(d) + gap
	}
	b.WriteString(`</svg>`)
	return b.String()
}

// writeBody малює саму схему (ребра, фігури, підпис) — без обгортки <svg>/фону.
func writeBody(b *strings.Builder, d *diagram.Diagram) {
	cap := d.CaptionLine()
	// Спершу ребра (під фігурами), потім фігури.
	for _, e := range d.Edges {
		marker := ` marker-end="url(#arr)"`
		if e.Arrowless {
			marker = ""
		}
		fmt.Fprintf(b, `<path d="%s" fill="none" stroke="#222" stroke-width="1.5"%s/>`,
			pathOf(e.Points), marker)
		if e.Label != "" && len(e.Points) >= 2 {
			lx, ly, anchor := diagram.LabelAnchor(e.Points[0], e.Points[1])
			fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="%s" font-size="12" fill="#444">%s</text>`,
				lx, ly, anchor, esc(e.Label))
		}
	}
	for _, s := range d.Shapes {
		renderShape(b, s)
	}
	// Підпис схеми («Рисунок N — …») — по центру під схемою.
	if cap != "" {
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="14" fill="#111">%s</text>`,
			d.W/2, d.H+capGap*0.6, esc(cap))
	}
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
	case diagram.Hexagon:
		sk := s.H * 0.5 // скоси з боків
		cy := s.Y + s.H/2
		fmt.Fprintf(b, `<polygon points="%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f" %s/>`,
			s.X+sk, s.Y, s.X+s.W-sk, s.Y, s.X+s.W, cy, s.X+s.W-sk, s.Y+s.H, s.X+sk, s.Y+s.H, s.X, cy, stroke)
	case diagram.Predef:
		fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" %s/>`, s.X, s.Y, s.W, s.H, stroke)
		const in = 9 // відступ бокових рисок
		fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#222" stroke-width="1.5"/>`, s.X+in, s.Y, s.X+in, s.Y+s.H)
		fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#222" stroke-width="1.5"/>`, s.X+s.W-in, s.Y, s.X+s.W-in, s.Y+s.H)
	case diagram.Connector:
		cx, cy := s.X+s.W/2, s.Y+s.H/2
		fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="%.1f" %s/>`, cx, cy, min(s.W, s.H)/2, stroke)
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
