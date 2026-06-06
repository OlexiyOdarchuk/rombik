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
defined = set()  # функції, визначені у файлі -> малюємо як підпрограму
MAXLEN = 64      # обрізаємо задовгий текст у блоці
def oneline(t):
    # будь-який текст у блок — в один рядок (захист від складних конструкцій)
    t = " ".join(t.split())
    return t if len(t)<=MAXLEN else t[:MAXLEN-1]+"…"
def expr(e):
    try: return oneline(ast.unparse(e))
    except Exception: return "?"
def fstr(j):
    # f-рядок -> читабельний текст із {виразами}
    out = ""
    for p in j.values:
        if isinstance(p, ast.Constant): out += str(p.value)
        elif isinstance(p, ast.FormattedValue): out += "{"+expr(p.value)+"}"
    return out
def arg(a):
    if isinstance(a, ast.Constant) and isinstance(a.value, str):
        return "«"+a.value+"»"
    if isinstance(a, ast.JoinedStr):
        return "«"+fstr(a)+"»"
    return expr(a)
def call_name(c):
    return c.func.id if isinstance(c, ast.Call) and isinstance(c.func, ast.Name) else None
def calls_defined(e):
    return isinstance(e, ast.Call) and call_name(e) in defined
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
def is_true(t):
    return isinstance(t, ast.Constant) and t.value in (True, 1)
def break_if(st):
    # «if COND: break» без else і єдиним break у тілі
    return (isinstance(st, ast.If) and not st.orelse
            and len(st.body)==1 and isinstance(st.body[0], ast.Break))
def stmt(s):
    if isinstance(s, ast.If):
        return {"kind":"if","cond":expr(s.test),"then":block(s.body),"else":block(s.orelse)}
    if isinstance(s, ast.While):
        # ідіома післяумови: while True: … if COND: break
        if is_true(s.test) and s.body and break_if(s.body[-1]):
            return {"kind":"dowhile","cond":expr(s.body[-1].test),"body":block(s.body[:-1])}
        return {"kind":"while","cond":expr(s.test),"body":block(s.body)}
    if isinstance(s, ast.For):
        return {"kind":"for","cond":forspec(s),"body":block(s.body)}
    if isinstance(s, (ast.Assign, ast.AugAssign, ast.AnnAssign)):
        if isinstance(s, ast.Assign) and has_input(s.value):
            return io("Ввід "+", ".join(expr(t) for t in s.targets))
        if calls_defined(s.value):  # x = моя_функція(...) -> підпрограма
            return {"kind":"call","text":expr(s)}
        return {"kind":"process","text":expr(s)}
    if isinstance(s, ast.Return):
        return io("Повернути "+expr(s.value)) if s.value else io("Повернути")
    if isinstance(s, ast.Expr) and isinstance(s.value, ast.Call):
        nm = call_name(s.value)
        if nm=="print": return io("Вивід "+" ".join(arg(a) for a in s.value.args))
        if nm=="input": return io("Ввід "+" ".join(arg(a) for a in s.value.args))
        if nm in defined: return {"kind":"call","text":expr(s.value)}
    if isinstance(s, ast.With):  # контекст-менеджер: заголовок + тіло
        items = ", ".join((expr(i.context_expr)+(" → "+expr(i.optional_vars) if i.optional_vars else "")) for i in s.items)
        return {"kind":"block","stmts":[{"kind":"process","text":oneline("відкрити: "+items)}]+[stmt(x) for x in s.body]}
    return {"kind":"process","text":expr(s)}
def block(stmts):
    return {"kind":"block","stmts":[stmt(s) for s in stmts if not isinstance(s,(ast.Pass,ast.Import,ast.ImportFrom))]}
def funcblock(fn):
    # параметри функції — як вхідний паралелограм (конвенція курсових схем)
    b = block(fn.body)
    params = [a.arg for a in fn.args.args]
    if params:
        b["stmts"].insert(0, io("Ввід "+", ".join(params)))
    return b
src = sys.stdin.read()
mod = ast.parse(src)
funcs = [s for s in mod.body if isinstance(s, ast.FunctionDef)]
defined = {fn.name for fn in funcs}  # заповнюємо ДО розбору тіл
rest  = [s for s in mod.body if not isinstance(s, (ast.FunctionDef, ast.Import, ast.ImportFrom))]
out = []
if funcs:
    for fn in funcs:
        out.append({"name":fn.name, "block":funcblock(fn)})
    if rest:
        out.append({"name":"main", "block":block(rest)})
else:
    out.append({"name":"main", "block":block(mod.body)})
print(json.dumps(out, ensure_ascii=False))
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

// pyFunc — іменована програма у JSON (функція або тіло модуля).
type pyFunc struct {
	Name  string `json:"name"`
	Block pyNode `json:"block"`
}

// ParseAll зводить Python-код у список іменованих програм (по одній на функцію;
// або одну «main», якщо функцій нема).
func ParseAll(code string) ([]ir.Func, error) {
	cmd := exec.Command("python3", "-c", pyScript)
	cmd.Stdin = strings.NewReader(code)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("python parse: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("python parse: %w (чи встановлено python3?)", err)
	}
	var fns []pyFunc
	if err := json.Unmarshal(out, &fns); err != nil {
		return nil, fmt.Errorf("розбір JSON від python: %w", err)
	}
	res := make([]ir.Func, 0, len(fns))
	for i := range fns {
		res = append(res, ir.Func{Name: fns[i].Name, Body: toBlock(&fns[i].Block)})
	}
	return res, nil
}

func toBlock(n *pyNode) *ir.Block {
	b := &ir.Block{}
	if n == nil {
		return b
	}
	for i := range n.Stmts {
		c := &n.Stmts[i]
		if c.Kind == "block" { // вкладений блок (напр. from with) — інлайнимо
			b.Stmts = append(b.Stmts, toBlock(c).Stmts...)
			continue
		}
		if node := toNode(c); node != nil {
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
	case "call":
		return &ir.Call{Text: n.Text}
	case "if":
		return &ir.If{Cond: n.Cond, Then: toBlock(n.Then), Else: toBlock(n.Else)}
	case "for":
		return &ir.For{Spec: n.Cond, Body: toBlock(n.Body)}
	case "while":
		return &ir.While{Cond: n.Cond, Body: toBlock(n.Body)}
	case "dowhile":
		return &ir.DoWhile{Cond: n.Cond, Body: toBlock(n.Body)}
	}
	return nil
}
