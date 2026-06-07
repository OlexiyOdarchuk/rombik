# Парсер Python -> спрощений AST у JSON (формат astjson).
# Працює у двох середовищах з одного джерела:
#   * CLI: код читається зі stdin, результат друкується у stdout;
#   * браузер (Pyodide): фронтенд заздалегідь ставить глобальну `src`,
#     а результат бере з глобальної `_out` (і/або зі stdout).
# Жодного виконання коду — лише ast.parse (безпечно для публічного вводу).
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
def is_exit(c):
    # exit()/quit()/sys.exit()
    if call_name(c) in ("exit", "quit"):
        return True
    return (isinstance(c, ast.Call) and isinstance(c.func, ast.Attribute)
            and c.func.attr == "exit")
def has_input(node):
    return any(call_name(n)=="input" for n in ast.walk(node))
def io(text):
    return {"kind":"io","text":text}
def endof(e):
    # кінець range — це stop-1; для літерала рахуємо, для «X + 1» -> «X»,
    # інакше «expr - 1»
    if isinstance(e, ast.Constant) and isinstance(e.value, int):
        return str(e.value-1)
    if isinstance(e, ast.BinOp) and isinstance(e.op, ast.Add) \
       and isinstance(e.right, ast.Constant) and e.right.value == 1:
        return expr(e.left)
    return expr(e)+" - 1"
def is_docstring(s):
    # рядок-літерал сам по собі (докстрінг/коментар) — не малюємо
    return isinstance(s, ast.Expr) and isinstance(s.value, ast.Constant) \
        and isinstance(s.value.value, str)
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
def match_cond(subj, pat):
    # умова-текст для патерна match; "" -> завжди істина (wildcard/capture)
    if isinstance(pat, ast.MatchValue):
        return subj+" == "+expr(pat.value)
    if isinstance(pat, ast.MatchSingleton):
        return subj+" is "+repr(pat.value)
    if isinstance(pat, ast.MatchOr):
        parts = [match_cond(subj, p) for p in pat.patterns]
        return "" if "" in parts else " or ".join(parts)
    if isinstance(pat, ast.MatchAs):
        return "" if pat.pattern is None else match_cond(subj, pat.pattern)
    try: return subj+" ~ "+ast.unparse(pat)        # складні патерни — текстом
    except Exception: return subj+" ~ ?"
def lower_match(m):
    # match/case -> ланцюг if/elif (кожен кейс — ромб; case _ -> else)
    subj = expr(m.subject)
    def build(i):
        if i >= len(m.cases):
            return {"kind":"block","stmts":[]}
        c = m.cases[i]
        cond = match_cond(subj, c.pattern)
        if c.guard is not None:
            cond = (cond+" and "+expr(c.guard)) if cond else expr(c.guard)
        if cond == "":                              # завжди-істинний — решта недосяжна
            return block(c.body)
        return {"kind":"block","stmts":[{"kind":"if","cond":oneline(cond),
                "then":block(c.body),"else":build(i+1)}]}
    return build(0)
def lower_try(t):
    # try/except/finally -> тіло + ланцюг if «Виняток <Тип>?» (обробники) + finally.
    # Чесно показуємо обробку винятків, а не мовчки її викидаємо.
    out = block(t.body)["stmts"]
    if t.handlers:
        def build(i):
            if i >= len(t.handlers):
                return {"kind":"block","stmts":[]}
            h = t.handlers[i]
            typ = (" "+expr(h.type)) if h.type else ""
            return {"kind":"block","stmts":[{"kind":"if","cond":oneline("Виняток"+typ+"?"),
                    "then":block(h.body),"else":build(i+1)}]}
        out += build(0)["stmts"]
    out += block(t.orelse)["stmts"]                 # try/else: шлях без винятку
    out += block(t.finalbody)["stmts"]              # finally: завжди наприкінці
    return {"kind":"block","stmts":out}
def stmt(s):
    if isinstance(s, ast.If):
        return {"kind":"if","cond":expr(s.test),"then":block(s.body),"else":block(s.orelse)}
    if isinstance(s, ast.Break):
        return {"kind":"break"}
    if isinstance(s, ast.Continue):
        return {"kind":"continue"}
    if isinstance(s, ast.While):
        # ідіома післяумови: while True: … if COND: break
        if is_true(s.test) and s.body and break_if(s.body[-1]):
            return {"kind":"dowhile","cond":expr(s.body[-1].test),"body":block(s.body[:-1])}
        # while True з break десь усередині — нескінченний цикл (без ромба)
        if is_true(s.test):
            return {"kind":"infloop","body":block(s.body)}
        return {"kind":"while","cond":expr(s.test),"body":block(s.body),"else":block(s.orelse)}
    if isinstance(s, ast.For):
        return {"kind":"for","cond":forspec(s),"body":block(s.body),"else":block(s.orelse)}
    if getattr(ast, "Match", None) and isinstance(s, ast.Match):  # 3.10+
        return lower_match(s)
    if isinstance(s, (ast.Assign, ast.AugAssign, ast.AnnAssign)):
        if isinstance(s, ast.Assign) and has_input(s.value):
            return io("Ввід "+", ".join(expr(t) for t in s.targets))
        if calls_defined(s.value):  # x = моя_функція(...) -> підпрограма
            return {"kind":"call","text":expr(s)}
        return {"kind":"process","text":expr(s)}
    if isinstance(s, ast.Return):
        return {"kind":"terminal","text":("Повернути "+expr(s.value)) if s.value else "Повернути"}
    if isinstance(s, ast.Raise):
        return {"kind":"terminal","text":("Помилка: "+expr(s.exc)) if s.exc else "Помилка"}
    if isinstance(s, ast.Expr) and isinstance(s.value, ast.Call):
        nm = call_name(s.value)
        if nm=="print":
            a = s.value.args
            return io("Вивід порожнього рядка" if not a else "Вивід "+" ".join(arg(x) for x in a))
        if nm=="input": return io("Ввід "+" ".join(arg(a) for a in s.value.args))
        if is_exit(s.value): return {"kind":"terminal","text":"Вихід"}
        if nm in defined: return {"kind":"call","text":expr(s.value)}
    if isinstance(s, ast.With):  # контекст-менеджер: заголовок + тіло
        items = ", ".join((expr(i.context_expr)+(" → "+expr(i.optional_vars) if i.optional_vars else "")) for i in s.items)
        return {"kind":"block","stmts":[{"kind":"process","text":oneline("відкрити: "+items)}]+[stmt(x) for x in s.body]}
    if isinstance(s, ast.Try):
        return lower_try(s)
    return {"kind":"process","text":expr(s)}
SKIP = (ast.Pass, ast.Import, ast.ImportFrom, ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef)
def block(stmts):
    # вкладені def/class не малюємо в тілі — вони йдуть окремими схемами
    return {"kind":"block","stmts":[stmt(s) for s in stmts
            if not isinstance(s, SKIP) and not is_docstring(s)]}
def funcblock(fn):
    # параметри функції — як вхідний паралелограм (конвенція курсових схем)
    b = block(fn.body)
    params = [a.arg for a in fn.args.args]
    if params:
        b["stmts"].insert(0, io("Ввід "+", ".join(params)))
    return b
try:
    src  # у браузері Pyodide ставить цю глобальну заздалегідь
except NameError:
    src = sys.stdin.read()
mod = ast.parse(src)
def collect(body, acc):
    # функції рекурсивно (зокрема вкладені) -> кожна окремою схемою.
    # У класи не заходимо (методи-дандери — зайвий шум).
    for s in body:
        if isinstance(s, (ast.FunctionDef, ast.AsyncFunctionDef)):
            acc.append(s); collect(s.body, acc)
funcs = []
collect(mod.body, funcs)
defined = {fn.name for fn in funcs}  # заповнюємо ДО розбору тіл
rest = [s for s in mod.body if not isinstance(s, (ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef, ast.Import, ast.ImportFrom))]
out = []
if funcs:
    for fn in funcs:
        out.append({"name":fn.name, "block":funcblock(fn)})
    if [s for s in rest if not is_docstring(s)]:
        out.append({"name":"main" if "main" not in defined else "програма", "block":block(rest)})
else:
    out.append({"name":"main", "block":block(mod.body)})
_out = json.dumps(out, ensure_ascii=False)
print(_out)
