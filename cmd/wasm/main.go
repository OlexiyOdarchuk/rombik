//go:build js && wasm

// Точка входу для браузера (WebAssembly). НЕ імпортує пакет python (там os/exec,
// якого у WASM нема): Python розбирає Pyodide, а ми отримуємо вже готовий
// AST-JSON. Реєструє в JS глобальну функцію rombikGenerate.
//
// Збірка:
//
//	GOOS=js GOARCH=wasm go build -o web/rombik.wasm ./cmd/wasm
//
// Виклик з JS:
//
//	const res = JSON.parse(rombikGenerate(astJSON, optionsJSON))
//	// res = { functions: [{name, svg, diagram}], error? }
package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/svg"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/typst"
	"github.com/OlexiyOdarchuk/rombik/pkg/rombik"
)

type outFunc struct {
	Name    string           `json:"name"`
	Caption string           `json:"caption"`
	FigNum  int              `json:"figNum"`
	SVG     string           `json:"svg"`
	Typst   string           `json:"typst"`
	Diagram *diagram.Diagram `json:"diagram"`
}

// generate(astJSON, optionsJSON?) -> JSON-рядок з усіма схемами або {error}.
func generate(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return result(map[string]any{"error": "немає AST-JSON"})
	}
	var opts rombik.Options
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		_ = json.Unmarshal([]byte(args[1].String()), &opts)
	}
	funcs, err := rombik.FromAST([]byte(args[0].String()), opts)
	if err != nil {
		return result(map[string]any{"error": err.Error()})
	}
	res := make([]outFunc, 0, len(funcs))
	for _, f := range funcs {
		res = append(res, outFunc{
			Name: f.Name, Caption: f.Diagram.Caption, FigNum: f.Diagram.FigNum,
			SVG: f.SVG(), Typst: f.Typst(), Diagram: f.Diagram,
		})
	}
	return result(map[string]any{"functions": res})
}

// renderOne(diagramJSON, captionJSON?) -> {svg, typst}. Дешевий ре-рендер після
// редагування підпису у фронті — без повторного розбору коду. captionJSON =
// {"caption":..,"figNum":..,"capWord":..} перекриває відповідні поля схеми.
func renderOne(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return result(map[string]any{"error": "немає diagram-JSON"})
	}
	var d diagram.Diagram
	if err := json.Unmarshal([]byte(args[0].String()), &d); err != nil {
		return result(map[string]any{"error": "розбір diagram: " + err.Error()})
	}
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		var cap struct {
			Caption *string `json:"caption"`
			FigNum  *int    `json:"figNum"`
			CapWord *string `json:"capWord"`
		}
		_ = json.Unmarshal([]byte(args[1].String()), &cap)
		if cap.Caption != nil {
			d.Caption = *cap.Caption
		}
		if cap.FigNum != nil {
			d.FigNum = *cap.FigNum
		}
		if cap.CapWord != nil {
			d.CapWord = *cap.CapWord
		}
	}
	return result(map[string]any{"svg": svg.Render(&d), "typst": typst.Render(&d)})
}

func result(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"серіалізація результату"}`
	}
	return string(b)
}

func main() {
	js.Global().Set("rombikGenerate", js.FuncOf(generate))
	js.Global().Set("rombikRenderOne", js.FuncOf(renderOne))
	select {} // тримаємо модуль живим
}
