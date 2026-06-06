// Команда flowgen — перший зріз: захардкоджений алгоритм → ДСТУ-блок-схема в SVG.
// Парсер коду (tree-sitter) додамо наступним кроком; зараз доводимо ядро
// (IR → layout → SVG).
//
//	go run ./cmd/flowgen   → пише out.svg
package main

import (
	"fmt"
	"os"

	"flowgen/internal/ir"
	"flowgen/internal/layout"
	"flowgen/internal/render/svg"
)

func main() {
	// Приклад: ввести n; s := n*n; якщо s>100 — «велике», інакше «мале».
	prog := ir.NewBlock(
		&ir.IO{Text: "Ввести n"},
		&ir.Process{Text: "s := n * n"},
		&ir.If{
			Cond: "s > 100",
			Then: ir.NewBlock(&ir.Process{Text: "Вивести «велике»"}),
			Else: ir.NewBlock(&ir.Process{Text: "Вивести «мале»"}),
		},
	)

	d := layout.Build(prog)
	out := svg.Render(d)

	if err := os.WriteFile("out.svg", []byte(out), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "запис:", err)
		os.Exit(1)
	}
	fmt.Printf("Готово: out.svg (%.0f×%.0f, фігур: %d, ребер: %d)\n", d.W, d.H, len(d.Shapes), len(d.Edges))
}
