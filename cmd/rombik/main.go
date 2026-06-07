// Команда rombik: код → ДСТУ-блок-схема в SVG.
//
//	go run ./cmd/rombik                 → демо (захардкоджений алгоритм)
//	go run ./cmd/rombik -py file.py     → схема(и) з Python-файлу
//	go run ./cmd/rombik -py file.py -o схема.svg
//	go run ./cmd/rombik -py file.py -o схема.png        → одразу PNG (rsvg-convert)
//	go run ./cmd/rombik -py file.py -o схема.png -scale 3
//	go run ./cmd/rombik -py file.py -fn matrix_gen      → лише одна функція
//
// Формат вихідного файлу — за розширенням -o: .svg, .png (через rsvg-convert),
// .json (дані для фронтенду).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
	"github.com/OlexiyOdarchuk/rombik/pkg/ir"
	"github.com/OlexiyOdarchuk/rombik/pkg/rombik"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/excalidraw"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/raster"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/svg"
	"github.com/OlexiyOdarchuk/rombik/pkg/render/typst"
)

func main() {
	pyFile := flag.String("py", "", "Python-файл для парсингу (інакше — демо)")
	outFile := flag.String("o", "out.svg", "вихідний SVG (для кількох функцій — основа імені)")
	fnName := flag.String("fn", "", "малювати лише функцію з цим іменем")
	callPlain := flag.Bool("calls-plain", false, "виклики підпрограм — звичайним прямокутником (не ДСТУ-символом)")
	singleEnd := flag.Bool("single-end", false, "один спільний Кінець (інакше — на кожен return/raise)")
	scale := flag.Float64("scale", 2, "масштаб для PNG (роздільність)")
	caption := flag.String("caption", "", "підпис схеми (інакше — ім'я функції; «-» — без підпису)")
	figNum := flag.Int("fignum", 0, "номер «Рисунок N» (0 — за порядком функцій)")
	figWord := flag.String("figword", "", "слово підпису: Рисунок (замовч.), Рис. тощо")
	figFmt := flag.String("capformat", "", "шаблон підпису, напр. «{num}. {text}» (замовч. «{word} {num} — {text}»)")
	flag.Parse()
	opts := rombik.Options{CallAsProcess: *callPlain, SingleEnd: *singleEnd, CapWord: *figWord}

	var funcs []rombik.Result
	if *pyFile == "" {
		funcs = rombik.FromIR([]ir.Func{{Name: "main", Body: demo()}}, opts)
	} else {
		code, err := os.ReadFile(*pyFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "читання:", err)
			os.Exit(1)
		}
		funcs, err = rombik.FromPython(string(code), opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if *fnName != "" {
		funcs = filterByName(funcs, *fnName)
		if len(funcs) == 0 {
			fmt.Fprintf(os.Stderr, "функцію %q не знайдено\n", *fnName)
			os.Exit(1)
		}
	}

	// Перекриття підпису з CLI: «-» вимикає, інакше задає текст; -fignum задає
	// стартовий номер (далі по порядку).
	for i, f := range funcs {
		switch {
		case *caption == "-":
			f.Diagram.Caption = ""
		case *caption != "":
			f.Diagram.Caption = *caption
		}
		if *figNum > 0 {
			f.Diagram.FigNum = *figNum + i
		}
		f.Diagram.CapFormat = *figFmt
	}

	// Одна схема → точно в -o; кілька → <основа>_<функція>.<ext>.
	if len(funcs) == 1 {
		write(funcs[0].Diagram, *outFile, *scale)
		return
	}
	base, ext := splitExt(*outFile)
	// .typ/.pdf/.excalidraw для кількох функцій → ОДИН документ з усіма схемами.
	if low := strings.ToLower(ext); low == ".typ" || low == ".pdf" || low == ".excalidraw" {
		writeCombined(funcs, *outFile, low)
		return
	}
	for _, f := range funcs {
		write(f.Diagram, fmt.Sprintf("%s_%s%s", base, f.Name, ext), *scale)
	}
}

// writeCombined зводить усі схеми в один .typ або .pdf документ.
func writeCombined(funcs []rombik.Result, out, ext string) {
	ds := make([]*diagram.Diagram, len(funcs))
	for i, f := range funcs {
		ds[i] = f.Diagram
	}
	if ext == ".typ" {
		writeFile(out, []byte(typst.RenderAll(ds)))
	} else if ext == ".excalidraw" {
		writeFile(out, []byte(excalidraw.RenderAll(ds)))
	} else {
		b, err := raster.PDFAll(ds)
		if err != nil {
			fmt.Fprintln(os.Stderr, "pdf:", err)
			os.Exit(1)
		}
		writeFile(out, b)
	}
	fmt.Printf("Готово: %s (%d схем одним документом)\n", out, len(funcs))
}

// write серіалізує діаграму у файл за розширенням: .json → дані, .png/.pdf →
// нативний растр/вектор (raster, без зовнішніх бінарників), .typ → Typst, інакше → SVG.
func write(d *diagram.Diagram, out string, scale float64) {
	switch strings.ToLower(filepath.Ext(out)) {
	case ".json":
		b, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "json:", err)
			os.Exit(1)
		}
		writeFile(out, b)
	case ".png":
		b, err := raster.PNG(d, scale)
		if err != nil {
			fmt.Fprintln(os.Stderr, "png:", err)
			os.Exit(1)
		}
		writeFile(out, b)
	case ".pdf":
		b, err := raster.PDF(d)
		if err != nil {
			fmt.Fprintln(os.Stderr, "pdf:", err)
			os.Exit(1)
		}
		writeFile(out, b)
	case ".typ":
		writeFile(out, []byte(typst.Render(d)))
	case ".excalidraw":
		writeFile(out, []byte(excalidraw.Render(d)))
	default:
		writeFile(out, []byte(svg.Render(d)))
	}
	fmt.Printf("Готово: %s (%.0f×%.0f, фігур: %d, ребер: %d)\n", out, d.W, d.H, len(d.Shapes), len(d.Edges))
}

func writeFile(out string, data []byte) {
	if err := os.WriteFile(out, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "запис:", err)
		os.Exit(1)
	}
}

func filterByName(fns []rombik.Result, name string) []rombik.Result {
	var res []rombik.Result
	for _, f := range fns {
		if f.Name == name {
			res = append(res, f)
		}
	}
	return res
}

// splitExt ділить "out.svg" → ("out", ".svg").
func splitExt(p string) (string, string) {
	ext := filepath.Ext(p)
	return strings.TrimSuffix(p, ext), ext
}

// demo — захардкоджений приклад (коли без -py).
func demo() *ir.Block {
	return ir.NewBlock(
		&ir.IO{Text: "Ввести n"},
		&ir.Process{Text: "s := n * n"},
		&ir.If{
			Cond: "s > 100",
			Then: ir.NewBlock(&ir.Process{Text: "Вивести «велике»"}),
			Else: ir.NewBlock(&ir.Process{Text: "Вивести «мале»"}),
		},
	)
}
