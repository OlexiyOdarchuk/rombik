//go:build js && wasm

// Точка входу для браузера (WebAssembly). НЕ імпортує пакет python (там os/exec,
// якого у WASM нема): Python розбирає Pyodide, а ми отримуємо вже готовий
// AST-JSON. Реєструє в JS глобальну функцію flowgenGenerate.
//
// Збірка:
//
//	GOOS=js GOARCH=wasm go build -o web/flowgen.wasm ./cmd/wasm
//
// Виклик з JS:
//
//	const res = JSON.parse(flowgenGenerate(astJSON, optionsJSON))
//	// res = { functions: [{name, svg, diagram}], error? }
package main

import (
	"encoding/json"
	"syscall/js"

	"flowgen/internal/diagram"
	"flowgen/internal/layout"
	"flowgen/internal/parser/astjson"
	"flowgen/internal/render/svg"
)

type outFunc struct {
	Name    string           `json:"name"`
	SVG     string           `json:"svg"`
	Diagram *diagram.Diagram `json:"diagram"`
}

// generate(astJSON, optionsJSON?) -> JSON-рядок з усіма схемами або {error}.
func generate(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return result(map[string]any{"error": "немає AST-JSON"})
	}
	var opts layout.Options
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		_ = json.Unmarshal([]byte(args[1].String()), &opts)
	}
	funcs, err := astjson.FromJSON([]byte(args[0].String()))
	if err != nil {
		return result(map[string]any{"error": err.Error()})
	}
	res := make([]outFunc, 0, len(funcs))
	for _, f := range funcs {
		d := layout.Build(f.Body, opts)
		res = append(res, outFunc{Name: f.Name, SVG: svg.Render(d), Diagram: d})
	}
	return result(map[string]any{"functions": res})
}

func result(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"серіалізація результату"}`
	}
	return string(b)
}

func main() {
	js.Global().Set("flowgenGenerate", js.FuncOf(generate))
	select {} // тримаємо модуль живим
}
