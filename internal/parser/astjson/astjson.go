// Пакет astjson — спільний контракт «спрощений AST у JSON → ir». Будь-який
// фронтенд парсера (Python через ast, C++ через tree-sitter, браузер через
// Pyodide/web-tree-sitter) видає цей самий формат, а тут єдина конвертація в ir.
// Завдяки цьому конвертацію переиспользуют усі мови й працює вона і в WASM
// (без жодних підпроцесів).
package astjson

import (
	"encoding/json"
	"fmt"

	"flowgen/internal/ir"
)

// Node — вузол спрощеного дерева. Поля використовуються залежно від Kind:
//
//	process/io/call — Text
//	if              — Cond, Then, Else
//	for/while/dowhile — Cond, Body
//	block           — Stmts (вкладені block інлайняться)
type Node struct {
	Kind  string `json:"kind"`
	Text  string `json:"text"`
	Cond  string `json:"cond"`
	Stmts []Node `json:"stmts"`
	Then  *Node  `json:"then"`
	Else  *Node  `json:"else"`
	Body  *Node  `json:"body"`
}

// Func — іменована програма (функція або тіло модуля) → окрема схема.
type Func struct {
	Name  string `json:"name"`
	Block Node   `json:"block"`
}

// FromJSON розбирає JSON-масив Func і зводить його в список ir.Func. Це й є
// точка входу для WASM: на вхід — рядок від парсера, на вихід — готовий ir.
func FromJSON(data []byte) ([]ir.Func, error) {
	var fns []Func
	if err := json.Unmarshal(data, &fns); err != nil {
		return nil, fmt.Errorf("розбір AST-JSON: %w", err)
	}
	res := make([]ir.Func, 0, len(fns))
	for i := range fns {
		res = append(res, ir.Func{Name: fns[i].Name, Body: ToBlock(&fns[i].Block)})
	}
	return res, nil
}

// ToBlock зводить вузол-блок у ir.Block (вкладені block-вузли інлайняться).
func ToBlock(n *Node) *ir.Block {
	b := &ir.Block{}
	if n == nil {
		return b
	}
	for i := range n.Stmts {
		c := &n.Stmts[i]
		if c.Kind == "block" {
			b.Stmts = append(b.Stmts, ToBlock(c).Stmts...)
			continue
		}
		if node := ToNode(c); node != nil {
			b.Stmts = append(b.Stmts, node)
		}
	}
	return b
}

// ToNode зводить один вузол у ir.Node (nil — якщо тип невідомий).
func ToNode(n *Node) ir.Node {
	switch n.Kind {
	case "process":
		return &ir.Process{Text: n.Text}
	case "io":
		return &ir.IO{Text: n.Text}
	case "terminal":
		return &ir.Terminal{Text: n.Text}
	case "call":
		return &ir.Call{Text: n.Text}
	case "if":
		return &ir.If{Cond: n.Cond, Then: ToBlock(n.Then), Else: ToBlock(n.Else)}
	case "for":
		return &ir.For{Spec: n.Cond, Body: ToBlock(n.Body)}
	case "while":
		return &ir.While{Cond: n.Cond, Body: ToBlock(n.Body)}
	case "dowhile":
		return &ir.DoWhile{Cond: n.Cond, Body: ToBlock(n.Body)}
	}
	return nil
}
