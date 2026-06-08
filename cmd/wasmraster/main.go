//go:build js && wasm

// Окремий «важкий» WASM для нативного PNG/PDF (tdewolff/canvas). Вантажиться у
// браузері ЛЕНИВО — лише на першу вимогу експорту PNG/PDF, щоб не роздувати
// початкове завантаження (головний rombik.wasm лишається легким).
//
// Реєструє JS-глобальні rombikPng(diagramJSON, capJSON?, scale?) і
// rombikPdf(diagramJSON, capJSON?) → JSON-рядок {png|pdf: base64} або {error}.
package main

import (
	"encoding/base64"
	"encoding/json"
	"syscall/js"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/raster"
)

// parseDiagram розбирає diagram-JSON і застосовує перекриття підпису (capJSON).
func parseDiagram(args []js.Value) (*diagram.Diagram, error) {
	var d diagram.Diagram
	if err := json.Unmarshal([]byte(args[0].String()), &d); err != nil {
		return nil, err
	}
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		var c struct {
			Caption   *string `json:"caption"`
			FigNum    *int    `json:"figNum"`
			CapWord   *string `json:"capWord"`
			CapFormat *string `json:"capFormat"`
		}
		_ = json.Unmarshal([]byte(args[1].String()), &c)
		if c.Caption != nil {
			d.Caption = *c.Caption
		}
		if c.FigNum != nil {
			d.FigNum = *c.FigNum
		}
		if c.CapWord != nil {
			d.CapWord = *c.CapWord
		}
		if c.CapFormat != nil {
			d.CapFormat = *c.CapFormat
		}
	}
	return &d, nil
}

func png(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return out(map[string]any{"error": "немає diagram-JSON"})
	}
	d, err := parseDiagram(args)
	if err != nil {
		return out(map[string]any{"error": "розбір diagram: " + err.Error()})
	}
	scale := 2.0
	if len(args) > 2 && !args[2].IsUndefined() && !args[2].IsNull() {
		scale = args[2].Float()
	}
	b, err := raster.PNG(d, scale)
	if err != nil {
		return out(map[string]any{"error": "png: " + err.Error()})
	}
	return out(map[string]any{"png": base64.StdEncoding.EncodeToString(b)})
}

func pdf(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return out(map[string]any{"error": "немає diagram-JSON"})
	}
	d, err := parseDiagram(args)
	if err != nil {
		return out(map[string]any{"error": "розбір diagram: " + err.Error()})
	}
	b, err := raster.PDF(d)
	if err != nil {
		return out(map[string]any{"error": "pdf: " + err.Error()})
	}
	return out(map[string]any{"pdf": base64.StdEncoding.EncodeToString(b)})
}

// pdfAll(diagramsJSON) -> {pdf: base64} — один багатосторінковий PDF з усіх схем.
func pdfAll(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return out(map[string]any{"error": "немає diagram-JSON"})
	}
	var ds []*diagram.Diagram
	if err := json.Unmarshal([]byte(args[0].String()), &ds); err != nil {
		return out(map[string]any{"error": "розбір: " + err.Error()})
	}
	b, err := raster.PDFAll(ds)
	if err != nil {
		return out(map[string]any{"error": "pdf: " + err.Error()})
	}
	return out(map[string]any{"pdf": base64.StdEncoding.EncodeToString(b)})
}

// pngAll(diagramsJSON, scale?) -> {png: base64} — усі схеми в одному PNG.
func pngAll(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return out(map[string]any{"error": "немає diagram-JSON"})
	}
	var ds []*diagram.Diagram
	if err := json.Unmarshal([]byte(args[0].String()), &ds); err != nil {
		return out(map[string]any{"error": "розбір: " + err.Error()})
	}
	scale := 2.0
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		scale = args[1].Float()
	}
	b, err := raster.PNGAll(ds, scale)
	if err != nil {
		return out(map[string]any{"error": "png: " + err.Error()})
	}
	return out(map[string]any{"png": base64.StdEncoding.EncodeToString(b)})
}

func out(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"серіалізація"}`
	}
	return string(b)
}

func wrap(f func(js.Value, []js.Value) any) func(js.Value, []js.Value) any {
	return func(this js.Value, args []js.Value) (res any) {
		defer func() {
			if r := recover(); r != nil {
				res = `{"error":"паніка у wasmraster модулі"}`
			}
		}()
		return f(this, args)
	}
}

func main() {
	js.Global().Set("rombikPng", js.FuncOf(wrap(png)))
	js.Global().Set("rombikPdf", js.FuncOf(wrap(pdf)))
	js.Global().Set("rombikPdfAll", js.FuncOf(wrap(pdfAll)))
	js.Global().Set("rombikPngAll", js.FuncOf(wrap(pngAll)))
	select {}
}
