// Команда flowgen: код → ДСТУ-блок-схема в SVG.
//
//	go run ./cmd/flowgen                 → демо (захардкоджений алгоритм)
//	go run ./cmd/flowgen -py file.py     → схема(и) з Python-файлу
//	go run ./cmd/flowgen -py file.py -o схема.svg
//	go run ./cmd/flowgen -py file.py -fn matrix_gen   → лише одна функція
package main

import (
	"flag"
	"fmt"
	"os"
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
	flag.Parse()

	// Демо, коли без -py.
	if *pyFile == "" {
		write(layout.Build(demo()), *outFile)
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

	// Одна схема → точно в -o; кілька → <основа>_<функція>.svg.
	if len(funcs) == 1 {
		write(layout.Build(funcs[0].Body), *outFile)
		return
	}
	base, ext := splitExt(*outFile)
	for _, f := range funcs {
		write(layout.Build(f.Body), fmt.Sprintf("%s_%s%s", base, f.Name, ext))
	}
}

// write рендерить діаграму в SVG-файл і друкує підсумок.
func write(d *diagram.Diagram, out string) {
	if err := os.WriteFile(out, []byte(svg.Render(d)), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "запис:", err)
		os.Exit(1)
	}
	fmt.Printf("Готово: %s (%.0f×%.0f, фігур: %d, ребер: %d)\n", out, d.W, d.H, len(d.Shapes), len(d.Edges))
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
