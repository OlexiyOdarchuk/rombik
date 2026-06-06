// Команда flowgen: код → ДСТУ-блок-схема в SVG.
//
//	go run ./cmd/flowgen                 → демо (захардкоджений алгоритм)
//	go run ./cmd/flowgen -py file.py     → схема(и) з Python-файлу
//	go run ./cmd/flowgen -py file.py -o схема.svg
//	go run ./cmd/flowgen -py file.py -o схема.png        → одразу PNG (rsvg-convert)
//	go run ./cmd/flowgen -py file.py -o схема.png -scale 3
//	go run ./cmd/flowgen -py file.py -fn matrix_gen      → лише одна функція
//
// Формат вихідного файлу — за розширенням -o: .svg, .png (через rsvg-convert),
// .json (дані для фронтенду).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"flowgen/internal/diagram"
	"flowgen/internal/ir"
	"flowgen/internal/layout"
	"flowgen/internal/parser/python"
	"flowgen/internal/render/svg"
)

func main() {
	pyFile := flag.String("py", "", "Python-файл для парсингу (інакше — демо)")
	outFile := flag.String("o", "out.svg", "вихідний SVG (для кількох функцій — основа імені)")
	fnName := flag.String("fn", "", "малювати лише функцію з цим іменем")
	callPlain := flag.Bool("calls-plain", false, "виклики підпрограм — звичайним прямокутником (не ДСТУ-символом)")
	singleEnd := flag.Bool("single-end", false, "один спільний Кінець (інакше — на кожен return/raise)")
	scale := flag.Float64("scale", 2, "масштаб для PNG (роздільність)")
	flag.Parse()
	opts := layout.Options{CallAsProcess: *callPlain, SingleEnd: *singleEnd}

	// Демо, коли без -py.
	if *pyFile == "" {
		write(layout.Build(demo(), opts), *outFile, *scale)
		return
	}

	code, err := os.ReadFile(*pyFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "читання:", err)
		os.Exit(1)
	}
	funcs, err := python.ParseAll(string(code))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *fnName != "" {
		funcs = filterByName(funcs, *fnName)
		if len(funcs) == 0 {
			fmt.Fprintf(os.Stderr, "функцію %q не знайдено\n", *fnName)
			os.Exit(1)
		}
	}

	// Одна схема → точно в -o; кілька → <основа>_<функція>.<ext>.
	if len(funcs) == 1 {
		write(layout.Build(funcs[0].Body, opts), *outFile, *scale)
		return
	}
	base, ext := splitExt(*outFile)
	for _, f := range funcs {
		write(layout.Build(f.Body, opts), fmt.Sprintf("%s_%s%s", base, f.Name, ext), *scale)
	}
}

// write серіалізує діаграму у файл за розширенням: .json → дані, .png → растр
// (rsvg-convert), інакше → SVG.
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
		if err := svgToPNG(svg.Render(d), out, scale); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
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

// svgToPNG растеризує SVG у PNG через rsvg-convert (пакет librsvg).
func svgToPNG(svgText, out string, scale float64) error {
	if _, err := exec.LookPath("rsvg-convert"); err != nil {
		return fmt.Errorf("для PNG потрібен rsvg-convert (встанови librsvg) або вкажи -o файл.svg")
	}
	cmd := exec.Command("rsvg-convert", "-z", fmt.Sprintf("%g", scale), "-o", out)
	cmd.Stdin = strings.NewReader(svgText)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func filterByName(fns []ir.Func, name string) []ir.Func {
	var res []ir.Func
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
