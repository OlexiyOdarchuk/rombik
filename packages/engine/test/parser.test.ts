// Парсер: TS parseTree на корпусі має дати той самий AST-JSON, що еталон (знятий
// через parser.js). Потребує web-tree-sitter + граматики з ../../../web (міграційно).
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { createRequire } from 'node:module';
import { parseTree, type Lang, type TSTree } from '../src/parser/treesitter.ts';

const here = import.meta.dirname;
const web = join(here, '..', '..', '..', 'web');
const require = createRequire(join(web, 'package.json'));
const ts: any = await import(require.resolve('web-tree-sitter'));
await ts.Parser.init();

const cache: Record<string, any> = {};
async function parserFor(lang: Lang) {
  if (!cache[lang]) {
    const p = new ts.Parser();
    const wasm = lang === 'cpp' ? 'tree-sitter-cpp.wasm' : lang === 'pascal' ? 'tree-sitter-pascal.wasm' : 'tree-sitter-python.wasm';
    p.setLanguage(await ts.Language.load(join(web, 'static', wasm)));
    cache[lang] = p;
  }
  return cache[lang];
}

const goldenDir = join(here, 'golden');
for (const gf of readdirSync(goldenDir)) {
  const g = JSON.parse(readFileSync(join(goldenDir, gf), 'utf8'));
  test(`parser: ${g.file}`, async () => {
    const p = await parserFor(g.lang as Lang);
    const tree: TSTree = p.parse(readFileSync(join(here, 'corpus', g.file), 'utf8'));
    assert.deepEqual(parseTree(tree, g.lang as Lang), g.astJSON);
  });
}

// --- switch (C++): регресія на «лише перший case» + fallthrough + зайвий value ---
async function astJson(src: string, lang: Lang = 'cpp') {
  const p = await parserFor(lang);
  return JSON.stringify(parseTree(p.parse(src), lang));
}

test('switch: усі case + default присутні (не лише перший)', async () => {
  const j = await astJson('void m(int o){ switch(o){ case 1: f(); break; case 2: g(); break; default: h(); } }');
  assert.match(j, /o == 1/);
  assert.match(j, /o == 2/);  // раніше губилося
  assert.match(j, /"h\(\)"/); // default-тіло теж
});

test('switch: fallthrough зливає мітки в одну умову через ||', async () => {
  const j = await astJson('void m(int d){ switch(d){ case 1: case 2: case 3: a(); break; default: b(); } }');
  assert.match(j, /d == 1 \|\| d == 2 \|\| d == 3/);
});

test('switch: значення case не протікає окремою дією', async () => {
  const j = await astJson('void m(int o){ switch(o){ case 1: f(); break; } }');
  assert.ok(!/"text":"1"/.test(j), 'значення "1" протекло як окремий вузол');
});

test('switch: умова веде в окрему гілку (вкладений if у else)', async () => {
  const ast = JSON.parse(await astJson('void m(int o){ switch(o){ case 1: f(); break; case 2: g(); break; } }'));
  // else першого case має бути блоком, що містить вкладений if (case 2)
  const sw = ast[0].block.stmts.find((s: any) => s.kind === 'if');
  assert.equal(sw.else.kind, 'block');
  assert.equal(sw.else.stmts[0].kind, 'if');
});

// --- інші C++ прогалини ---
test('range-for (for (int x : arr)) — це цикл, не дія', async () => {
  const ast = JSON.parse(await astJson('int s(int a[]){ int t=0; for (int x : a) { t += x; } return t; }'));
  const loop = ast[0].block.stmts.find((s: any) => s.kind === 'for');
  assert.ok(loop, 'range-for не став циклом');
  assert.match(loop.cond, /x ∈ a/);
});

test('методи класу — окремі схеми «Клас::метод»', async () => {
  const ast = JSON.parse(await astJson('class C { int v; public: void inc(){ v=v+1; } int get(){ return v; } };'));
  const names = ast.map((f: any) => f.name);
  assert.deepEqual(names, ['C::inc', 'C::get']);
  // поле v не протекло в окрему «main»-схему
  assert.ok(!names.includes('main') && !names.includes('програма'));
});

test('using namespace / класи не створюють зайвої схеми «програма»', async () => {
  const ast = JSON.parse(await astJson('using namespace std;\nint main(){ return 0; }'));
  assert.deepEqual(ast.map((f: any) => f.name), ['main']);
});

test('Python: методи класу → окремі схеми «Клас.метод», self прибрано', async () => {
  const ast = JSON.parse(await astJson('class C:\n    def __init__(self, x):\n        self.x = x\n    def get(self):\n        return self.x\n', 'python'));
  assert.deepEqual(ast.map((f: any) => f.name), ['C.__init__', 'C.get']);
  // self не потрапляє у «Ввід …»
  assert.deepEqual(ast[0].block.stmts[0], { kind: 'io', text: 'Ввід x' });
});

test('C++: шаблонна функція розпізнається (не сирий блок)', async () => {
  const ast = JSON.parse(await astJson('template <typename T>\nT maxOf(T a, T b){ if (a > b) return a; return b; }'));
  assert.deepEqual(ast.map((f: any) => f.name), ['maxOf']);
  assert.ok(!/template/.test(JSON.stringify(ast)), 'template протік у схему');
});

test('C++: enum не створює зайвої схеми «програма»', async () => {
  const ast = JSON.parse(await astJson('enum Color { RED, GREEN };\nvoid f(){ cout << 1; }'));
  assert.deepEqual(ast.map((f: any) => f.name), ['f']);
});

test('Python: list-comprehension розгортається в цикл', async () => {
  const ast = JSON.parse(await astJson('def f(n):\n    r = [i*i for i in range(n)]\n    return r\n', 'python'));
  const stmts = ast[0].block.stmts.flatMap((s: any) => s.kind === 'block' ? s.stmts : [s]);
  assert.ok(stmts.some((s: any) => s.kind === 'process' && s.text === 'r = []'), 'нема ініціалізації r = []');
  const loop = stmts.find((s: any) => s.kind === 'for');
  assert.ok(loop, 'comprehension не став циклом');
  assert.match(JSON.stringify(loop), /r\.append\(i\*i\)/);
});

test('Python: comprehension з if → ромб у тілі циклу', async () => {
  const ast = JSON.parse(await astJson('def f(xs):\n    e = [x for x in xs if x > 0]\n    return e\n', 'python'));
  const j = JSON.stringify(ast);
  assert.match(j, /"kind":"for"/);
  assert.match(j, /"cond":"x > 0"/); // if-клауза стала умовою
  assert.match(j, /e\.append\(x\)/);
});

test('Python: dict-comprehension → d = {} ; for … : d[k] = v', async () => {
  const ast = JSON.parse(await astJson('def f(ks, vs):\n    d = {k: v for k, v in zip(ks, vs)}\n    return d\n', 'python'));
  const j = JSON.stringify(ast);
  assert.match(j, /"d = \{\}"/);
  assert.match(j, /"d\[k\] = v"/);
});

test('Python: декоровані методи (@staticmethod) витягуються', async () => {
  const ast = JSON.parse(await astJson('class M:\n    @staticmethod\n    def sq(x):\n        return x*x\n    @classmethod\n    def make(cls):\n        return cls()\n', 'python'));
  assert.deepEqual(ast.map((f: any) => f.name), ['M.sq', 'M.make']);
});

test('Python: вкладений клас → Зовн.Внутр.метод', async () => {
  const ast = JSON.parse(await astJson('class O:\n    class I:\n        def ping(self):\n            return 1\n    def m(self):\n        return 2\n', 'python'));
  assert.deepEqual(ast.map((f: any) => f.name), ['O.I.ping', 'O.m']);
});

test('C++: функції в namespace витягуються', async () => {
  const ast = JSON.parse(await astJson('namespace m { int sq(int x){ return x*x; } int cube(int x){ return x*x*x; } }'));
  assert.deepEqual(ast.map((f: any) => f.name), ['sq', 'cube']);
});

test('C++: printf/scanf → ввід/вивід', async () => {
  const ast = JSON.parse(await astJson('void f(){ int x; scanf("%d", &x); printf("got %d", x); }'));
  const j = JSON.stringify(ast);
  assert.match(j, /"Ввід x"/);       // & прибрано, формат відкинуто
  assert.match(j, /Вивід .*got/);
});

test('C++: operator та деструктор мають правильні імена', async () => {
  const op = JSON.parse(await astJson('struct V { int x; V operator+(V o){ V r; return r; } };'));
  assert.ok(op.some((f: any) => f.name === 'V::operator+'), 'operator+ → unknown');
  const dt = JSON.parse(await astJson('class F { public: F(){} ~F(){ cout << 1; } };'));
  assert.deepEqual(dt.map((f: any) => f.name), ['F::F', 'F::~F']);
});

test('C++: out-of-line метод (Клас::метод) має правильне імʼя', async () => {
  const ast = JSON.parse(await astJson('int Counter::get(){ return value; }'));
  assert.deepEqual(ast.map((f: any) => f.name), ['Counter::get']);
});

test('C++: goto/мітка → конектори з тією ж літерою', async () => {
  const ast = JSON.parse(await astJson('void f(int n){\ni:\n  if (n > 0) { n--; goto i; }\n}'));
  const conns = JSON.stringify(ast).match(/"kind":"connector"[^}]*"text":"i"/g) ?? [];
  assert.ok(conns.length >= 2, 'goto+мітка мають дати 2 конектори «i»');
  assert.match(JSON.stringify(ast), /"kind":"connector"[^}]*"jump":true/); // goto — термінальний
});

// --- Pascal ---
test('Pascal: процедури/функції + головна програма → окремі схеми', async () => {
  const ast = JSON.parse(await astJson('program P;\nfunction sq(x: integer): integer;\nbegin\n  sq := x*x;\nend;\nbegin\n  writeln(sq(3));\nend.', 'pascal'));
  assert.deepEqual(ast.map((f: any) => f.name), ['sq', 'P']);
});

test('Pascal: writeln/readln → вивід/ввід, := → дія', async () => {
  const ast = JSON.parse(await astJson('program P;\nvar n: integer;\nbegin\n  readln(n);\n  n := n + 1;\n  writeln(n);\nend.', 'pascal'));
  const j = JSON.stringify(ast);
  assert.match(j, /"Ввід n"/);
  assert.match(j, /"n = n \+ 1"/); // := перетворено на =
  assert.match(j, /Вивід .*n/);
});

test('Pascal: for/while/repeat/case дають керівні вузли', async () => {
  const forA = JSON.parse(await astJson('program P;\nvar i: integer;\nbegin\n  for i := 1 to 10 do writeln(i);\nend.', 'pascal'));
  assert.match(JSON.stringify(forA), /"kind":"for".*"i = 1, 10, 1"/);
  const rep = JSON.parse(await astJson('program P;\nvar n: integer;\nbegin\n  repeat n := n+1; until n >= 5;\nend.', 'pascal'));
  assert.match(JSON.stringify(rep), /"kind":"dowhile".*не \(n >= 5\)/);
  const cas = JSON.parse(await astJson('program P;\nvar n: integer;\nbegin\n  case n of\n    1: writeln(1);\n    2: writeln(2);\n  end;\nend.', 'pascal'));
  assert.match(JSON.stringify(cas), /"kind":"if".*n = 1/);
});
