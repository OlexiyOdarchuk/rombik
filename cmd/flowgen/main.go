// Команда flowgen: код → ДСТУ-блок-схема в SVG.
//
//	go run ./cmd/flowgen                 → демо (захардкоджений алгоритм)
//	go run ./cmd/flowgen -py file.py     → схема з Python-файлу
//	go run ./cmd/flowgen -py file.py -o схема.svg
package main

import (
	"flag"
	"fmt"
	"os"

	"flowgen/internal/ir"
	"flowgen/internal/layout"
	"flowgen/internal/parser/python"
	"flowgen/internal/render/svg"
)

func main() {
	pyFile := flag.String("py", "", "Python-файл для парсингу (інакше — демо)")
	outFile := flag.String("o", "out.svg", "вихідний SVG")
	flag.Parse()

	var prog *ir.Block
	if *pyFile != "" {
		code, err := os.ReadFile(*pyFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "читання:", err)
			os.Exit(1)
		}
		prog, err = python.Parse(string(code))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		prog = demo()
	}

	d := layout.Build(prog)
	if err := os.WriteFile(*outFile, []byte(svg.Render(d)), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "запис:", err)
		os.Exit(1)
	}
	fmt.Printf("Готово: %s (%.0f×%.0f, фігур: %d, ребер: %d)\n", *outFile, d.W, d.H, len(d.Shapes), len(d.Edges))
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
