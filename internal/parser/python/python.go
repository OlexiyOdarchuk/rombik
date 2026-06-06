// Пакет python парсить Python-код у ir. Використовуємо РІДНИЙ Python `ast`
// (через python3 у підпроцесі): найточніший парсер Python, без cgo. Сам ast
// віддає спрощений JSON, а Go зводить його в ir.
//
// Вимога рантайму: python3 (3.9+, заради ast.unparse). Це адаптер парсера —
// ядро (layout/svg) від нього не залежить.
package python

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"flowgen/internal/ir"
)

// pyScript читає код зі stdin і друкує спрощене дерево у JSON.
const pyScript = `
import sys, ast, json
def expr(e):
    try: return ast.unparse(e).strip()
    except Exception: return "?"
def arg(a):
    if isinstance(a, ast.Constant) and isinstance(a.value, str):
        return "«"+a.value+"»"
    return expr(a)
def call_name(c):
    return c.func.id if isinstance(c, ast.Call) and isinstance(c.func, ast.Name) else None
def has_input(node):
    return any(call_name(n)=="input" for n in ast.walk(node))
def io(text):
    return {"kind":"io","text":text}
def endof(e):
    # кінець range — це stop-1; для літерала рахуємо, інакше «expr - 1»
    if isinstance(e, ast.Constant) and isinstance(e.value, int):
        return str(e.value-1)
    return expr(e)+" - 1"
def forspec(s):
    # «змінна = початок, кінець, крок» для for ... in range(...)
    tgt = expr(s.target)
    it = s.iter
    if call_name(it)=="range":
        a = it.args
        if len(a)==1: return tgt+" = 0, "+endof(a[0])+", 1"
        if len(a)==2: return tgt+" = "+expr(a[0])+", "+endof(a[1])+", 1"
        if len(a)>=3: return tgt+" = "+expr(a[0])+", "+endof(a[1])+", "+expr(a[2])
    return tgt+" ∈ "+expr(it)
def stmt(s):
    if isinstance(s, ast.If):
        return {"kind":"if","cond":expr(s.test),"then":block(s.body),"else":block(s.orelse)}
    if isinstance(s, ast.While):
        return {"kind":"while","cond":expr(s.test),"body":block(s.body)}
    if isinstance(s, ast.For):
        return {"kind":"for","cond":forspec(s),"body":block(s.body)}
    if isinstance(s, (ast.Assign, ast.AugAssign, ast.AnnAssign)):
        if isinstance(s, ast.Assign) and has_input(s.value):
            return io("Ввід "+", ".join(expr(t) for t in s.targets))
        return {"kind":"process","text":expr(s)}
    if isinstance(s, ast.Return):
        return {"kind":"process","text":("повернути "+expr(s.value)) if s.value else "повернути"}
    if isinstance(s, ast.Expr) and isinstance(s.value, ast.Call):
        nm = call_name(s.value)
        if nm=="print": return io("Вивід "+" ".join(arg(a) for a in s.value.args))
        if nm=="input": return io("Ввід "+" ".join(arg(a) for a in s.value.args))
    if isinstance(s, ast.Expr):
        return {"kind":"process","text":expr(s)}
    return {"kind":"process","text":expr(s)}
def block(stmts):
    return {"kind":"block","stmts":[stmt(s) for s in stmts if not isinstance(s,(ast.Pass,ast.Import,ast.ImportFrom))]}
src = sys.stdin.read()
mod = ast.parse(src)
body = mod.body
if len(body)==1 and isinstance(body[0], ast.FunctionDef):
    body = body[0].body
print(json.dumps(block(body), ensure_ascii=False))
`

// pyNode — форма вузла у JSON від скрипта.
type pyNode struct {
	Kind  string   `json:"kind"`
	Text  string   `json:"text"`
	Cond  string   `json:"cond"`
	Stmts []pyNode `json:"stmts"`
	Then  *pyNode  `json:"then"`
	Else  *pyNode  `json:"else"`
	Body  *pyNode  `json:"body"`
}

// Parse зводить Python-код у ir.Block.
func Parse(code string) (*ir.Block, error) {
	cmd := exec.Command("python3", "-c", pyScript)
	cmd.Stdin = strings.NewReader(code)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("python parse: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("python parse: %w (чи встановлено python3?)", err)
	}
	var root pyNode
	if err := json.Unmarshal(out, &root); err != nil {
		return nil, fmt.Errorf("розбір JSON від python: %w", err)
	}
	return toBlock(&root), nil
}

func toBlock(n *pyNode) *ir.Block {
	b := &ir.Block{}
	if n == nil {
		return b
	}
	for i := range n.Stmts {
		if node := toNode(&n.Stmts[i]); node != nil {
			b.Stmts = append(b.Stmts, node)
		}
	}
	return b
}

func toNode(n *pyNode) ir.Node {
	switch n.Kind {
	case "process":
		return &ir.Process{Text: n.Text}
	case "io":
		return &ir.IO{Text: n.Text}
	case "if":
		return &ir.If{Cond: n.Cond, Then: toBlock(n.Then), Else: toBlock(n.Else)}
	case "for":
		return &ir.For{Spec: n.Cond, Body: toBlock(n.Body)}
	case "while":
		return &ir.While{Cond: n.Cond, Body: toBlock(n.Body)}
	}
	return nil
}
