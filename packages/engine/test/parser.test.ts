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
    p.setLanguage(await ts.Language.load(join(web, 'static', lang === 'cpp' ? 'tree-sitter-cpp.wasm' : 'tree-sitter-python.wasm')));
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
