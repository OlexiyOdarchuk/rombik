package rombik

import (
	"fmt"

	"github.com/OlexiyOdarchuk/rombik/pkg/ir"
	"github.com/OlexiyOdarchuk/rombik/pkg/layout"
	"github.com/OlexiyOdarchuk/rombik/pkg/parser/astjson"
)

// SplitFromAST: AST-JSON + ім'я функції → частини схеми (для WASM-кнопки розбивки).
func SplitFromAST(astJSON []byte, opts Options, name string, maxH float64) ([]Result, error) {
	funcs, err := astjson.FromJSON(astJSON)
	if err != nil {
		return nil, err
	}
	for _, f := range funcs {
		if f.Name == name {
			return SplitByHeight(f, opts, maxH), nil
		}
	}
	return nil, fmt.Errorf("функцію %q не знайдено", name)
}

// abcRunes — літери для конекторів розриву (А, Б, В…).
var abcRunes = []rune("АБВГДЕЖЗИКЛМНОПРСТУФ")

func letter(i int) string {
	if i >= 0 && i < len(abcRunes) {
		return string(abcRunes[i])
	}
	// Фолбек для дуже довгих схем (>20 частин), щоб уникнути паніки та дублікатів
	return fmt.Sprintf("А%d", i-len(abcRunes)+1)
}

// SplitByHeight ріже схему функції на кілька зв'язаних частин так, щоб кожна
// була не вища за maxH (одиниць). Ріже на межах інструкцій ВЕРХНЬОГО рівня, на
// стику ставить пару конекторів «А» (вихід попередньої → вхід наступної). Якщо
// схема й так уміщається — повертає одну Result (без змін).
func SplitByHeight(f ir.Func, opts Options, maxH float64) []Result {
	stmts := f.Body.Stmts
	if len(stmts) <= 1 {
		return build([]ir.Func{f}, opts)
	}
	// Жадібно групуємо інструкції, поки висота частини не перевищить поріг.
	var groups [][]ir.Node
	cur := []ir.Node{}
	for _, s := range stmts {
		cand := append(append([]ir.Node{}, cur...), s)
		h := layout.Build(&ir.Block{Stmts: cand}, opts).H
		if h > maxH && len(cur) > 0 {
			groups = append(groups, cur)
			cur = []ir.Node{s}
		} else {
			cur = cand
		}
	}
	if len(cur) > 0 {
		groups = append(groups, cur)
	}
	if len(groups) <= 1 { // ділити нема сенсу
		return build([]ir.Func{f}, opts)
	}

	res := make([]Result, len(groups))
	for i, g := range groups {
		body := make([]ir.Node, 0, len(g)+2)
		if i > 0 { // вхідний конектор (з'єднує з попередньою частиною)
			body = append(body, &ir.Connector{Text: letter(i - 1)})
		}
		body = append(body, g...)
		if i < len(groups)-1 { // вихідний конектор (на наступну частину)
			body = append(body, &ir.Connector{Text: letter(i)})
		}
		o := opts
		o.NoStart = i > 0
		o.NoEnd = i < len(groups)-1
		d := layout.Build(&ir.Block{Stmts: body}, o)
		d.Caption = fmt.Sprintf("%s (частина %d з %d)", f.Name, i+1, len(groups))
		d.FigNum = i + 1
		d.CapWord = opts.CapWord
		res[i] = Result{Name: fmt.Sprintf("%s_ч%d", f.Name, i+1), Diagram: d}
	}
	return res
}
