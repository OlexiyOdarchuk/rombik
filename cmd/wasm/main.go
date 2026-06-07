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
	"github.com/OlexiyOdarchuk/rombik/pkg/render/excalidraw"
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

func argBool(args []js.Value, i int) bool {
	return len(args) > i && !args[i].IsUndefined() && !args[i].IsNull() && args[i].Bool()
}

// typstOne(diagramJSON, fragment?) -> {typst}. fragment=true → лише cetz.canvas
// (без преамбули/сторінки) для вставки у свій .typ.
func typstOne(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return result(map[string]any{"error": "немає diagram-JSON"})
	}
	var d diagram.Diagram
	if err := json.Unmarshal([]byte(args[0].String()), &d); err != nil {
		return result(map[string]any{"error": "розбір: " + err.Error()})
	}
	if argBool(args, 1) {
		return result(map[string]any{"typst": typst.Fragment(&d)})
	}
	return result(map[string]any{"typst": typst.Render(&d)})
}

// typstAll(diagramsJSON, fragment?) -> {typst} — усі схеми одним Typst (масив
// diagram із уже виставленими полями підпису). Для «експортувати все».
func typstAll(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return result(map[string]any{"error": "немає diagram-JSON"})
	}
	var ds []*diagram.Diagram
	if err := json.Unmarshal([]byte(args[0].String()), &ds); err != nil {
		return result(map[string]any{"error": "розбір: " + err.Error()})
	}
	if argBool(args, 1) {
		return result(map[string]any{"typst": typst.FragmentAll(ds)})
	}
	return result(map[string]any{"typst": typst.RenderAll(ds)})
}

// excal(diagramJSON) -> {excalidraw} — одна схема у форматі .excalidraw.
func excal(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return result(map[string]any{"error": "немає diagram-JSON"})
	}
	var d diagram.Diagram
	if err := json.Unmarshal([]byte(args[0].String()), &d); err != nil {
		return result(map[string]any{"error": "розбір: " + err.Error()})
	}
	return result(map[string]any{"excalidraw": excalidraw.Render(&d)})
}

// excalAll(diagramsJSON) -> {excalidraw} — усі схеми в одному .excalidraw.
func excalAll(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return result(map[string]any{"error": "немає diagram-JSON"})
	}
	var ds []*diagram.Diagram
	if err := json.Unmarshal([]byte(args[0].String()), &ds); err != nil {
		return result(map[string]any{"error": "розбір: " + err.Error()})
	}
	return result(map[string]any{"excalidraw": excalidraw.RenderAll(ds)})
}

// svgAll(diagramsJSON) -> {svg} — усі схеми в одному SVG (вертикально).
func svgAll(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return result(map[string]any{"error": "немає diagram-JSON"})
	}
	var ds []*diagram.Diagram
	if err := json.Unmarshal([]byte(args[0].String()), &ds); err != nil {
		return result(map[string]any{"error": "розбір: " + err.Error()})
	}
	return result(map[string]any{"svg": svg.RenderAll(ds)})
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
	js.Global().Set("rombikTypstOne", js.FuncOf(typstOne))
	js.Global().Set("rombikTypstAll", js.FuncOf(typstAll))
	js.Global().Set("rombikSvgAll", js.FuncOf(svgAll))
	js.Global().Set("rombikExcalidraw", js.FuncOf(excal))
	js.Global().Set("rombikExcalidrawAll", js.FuncOf(excalAll))
	select {} // тримаємо модуль живим
}
