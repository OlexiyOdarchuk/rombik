// Пакет rombik — високорівневий публічний API: код → блок-схеми за ДСТУ.
// Об'єднує парсер, розкладку й рендер. Для бібліотечного використання:
//
//	res, err := rombik.FromPython(code, rombik.Options{})
//	for _, f := range res {
//	    os.WriteFile(f.Name+".svg", []byte(f.SVG()), 0o644)
//	}
//
// У браузері (WASM, без python3) парсер дає AST окремо (Pyodide), а тут —
// rombik.FromAST(astJSON, opts).
package rombik

import (
	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
	"github.com/OlexiyOdarchuk/rombik/pkg/ir"
	"github.com/OlexiyOdarchuk/rombik/pkg/layout"
	"github.com/OlexiyOdarchuk/rombik/pkg/parser/astjson"
	"github.com/OlexiyOdarchuk/rombik/pkg/parser/python"

	"github.com/OlexiyOdarchuk/rombik/pkg/render/excalidraw"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/raster"
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

// PDF малює схему напряму у PDF (через fpdf — без компілятора, працює в WASM).
func (r Result) PDF() ([]byte, error) { return raster.PDF(r.Diagram) }

// PNG растеризує схему нативно (scale — пікселів на одиницю; типово 2).
func (r Result) PNG(scale float64) ([]byte, error) { return raster.PNG(r.Diagram, scale) }

// Excalidraw повертає схему у форматі .excalidraw (для excalidraw.com).
func (r Result) Excalidraw() string { return excalidraw.Render(r.Diagram) }

// FromPython: Python-код → схеми (потребує python3 у системі; не для WASM).
func FromPython(code string, opts Options) ([]Result, error) {
	funcs, err := python.ParseAll(code)
	if err != nil {
		return nil, err
	}
	return build(funcs, opts), nil
}

// FromAST: вже розібраний AST-JSON (формат astjson) → схеми. Працює будь-де,
// зокрема у WASM (AST дає Pyodide).
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
		d.Caption = f.Name      // підпис за замовч. — ім'я функції (редагований у фронті)
		d.FigNum = i + 1        // «Рисунок N» за порядком у файлі
		d.CapWord = opts.CapWord // слово підпису («Рисунок»/«Рис.»/своє)
		res[i] = Result{Name: f.Name, Diagram: d}
	}
	return res
}
