// Пакет rombik — високорівневий публічний API: AST-JSON → блок-схеми за ДСТУ.
// Об'єднує розкладку й рендер. Для бібліотечного використання:
//
//	res, err := rombik.FromAST(astJSON, rombik.Options{})
//	for _, f := range res {
//	    os.WriteFile(f.Name+".svg", []byte(f.SVG()), 0o644)
//	}
//
// Парсинг коду — поза цим пакетом: фронтенд (tree-sitter у вебі/Node) дає AST-JSON,
// а тут FromAST(astJSON) зводить його в схеми. Так ядро не залежить ні від мови,
// ні від рантайму парсера (python3 більше не потрібен).
//
// PNG/PDF тут НЕ надаються навмисно: вони тягнуть важке дерево tdewolff/canvas
// (растеризація, шрифти, латех, ~55 модулів). Щоб імпорт цього пакета заради SVG
// не витягував усе те, растрові формати винесено в окремий пакет render/raster:
//
//	import "github.com/OlexiyOdarchuk/rombik/pkg/render/raster"
//	png, err := raster.PNG(res[0].Diagram, 2)  // Result.Diagram — публічний
//	pdf, err := raster.PDF(res[0].Diagram)
package rombik

import (
	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
	"github.com/OlexiyOdarchuk/rombik/pkg/ir"
	"github.com/OlexiyOdarchuk/rombik/pkg/layout"
	"github.com/OlexiyOdarchuk/rombik/pkg/parser/astjson"

	"github.com/OlexiyOdarchuk/rombik/pkg/render/excalidraw"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/svg"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/typst"
)

// Options — перемикачі рендера (див. layout.Options).
type Options = layout.Options

// Result — одна побудована схема (на функцію).
type Result struct {
	Name    string
	Diagram *diagram.Diagram
}

// SVG повертає схему у форматі SVG.
func (r Result) SVG() string { return svg.Render(r.Diagram) }

// Typst повертає самодостатній Typst-документ (компілюється у тісний PDF).
func (r Result) Typst() string { return typst.Render(r.Diagram) }

// PNG/PDF — у пакеті render/raster (важкий canvas-стек винесено окремо, див. док
// пакета): raster.PNG(r.Diagram, scale), raster.PDF(r.Diagram).

// Excalidraw повертає схему у форматі .excalidraw (для excalidraw.com).
func (r Result) Excalidraw() string { return excalidraw.Render(r.Diagram) }

// FromAST: розібраний AST-JSON (формат astjson) → схеми. Це єдина точка входу
// з коду: парсер (tree-sitter у вебі/Node) дає AST-JSON, а далі — мова-агностик.
func FromAST(astJSON []byte, opts Options) ([]Result, error) {
	funcs, err := astjson.FromJSON(astJSON)
	if err != nil {
		return nil, err
	}
	return build(funcs, opts), nil
}

// FromIR: вже готовий список програм (ir) → схеми. Для тих, хто будує ir сам.
func FromIR(funcs []ir.Func, opts Options) []Result { return build(funcs, opts) }

func build(funcs []ir.Func, opts Options) []Result {
	res := make([]Result, len(funcs))
	for i, f := range funcs {
		d := layout.Build(f.Body, opts)
		d.Caption = f.Name       // підпис за замовч. — ім'я функції (редагований у фронті)
		d.FigNum = i + 1         // «Рисунок N» за порядком у файлі
		d.CapWord = opts.CapWord // слово підпису («Рисунок»/«Рис.»/своє)
		res[i] = Result{Name: f.Name, Diagram: d}
	}
	return res
}
