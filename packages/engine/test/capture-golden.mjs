// Регенерує golden-еталони з ПОТОЧНОГО TS-рушія (Go видалено — це регрес-снапшоти).
// Парність із Go доведено історично (git); відтепер golden ловлять регресії TS.
import { readFileSync, writeFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { createRequire } from 'node:module';
import { parseTree, fromAst, splitFromAst, renderSvg, renderTypst, renderExcalidraw } from '../src/index.ts';
const here = import.meta.dirname, web = join(here, '..', '..', '..', 'web');
const req = createRequire(join(web, 'package.json'));
const ts = await import(req.resolve('web-tree-sitter')); await ts.Parser.init();
const cache = {};
async function P(lang) { if (!cache[lang]) { const p = new ts.Parser(); p.setLanguage(await ts.Language.load(join(web, 'static', lang === 'cpp' ? 'tree-sitter-cpp.wasm' : 'tree-sitter-python.wasm'))); cache[lang] = p; } return cache[lang]; }
for (const f of readdirSync(join(here, 'corpus'))) {
  const lang = f.endsWith('.cpp') ? 'cpp' : 'python';
  const tree = (await P(lang)).parse(readFileSync(join(here, 'corpus', f), 'utf8'));
  const ast = parseTree(tree, lang);
  const figs = fromAst(ast, {});
  writeFileSync(join(here, 'golden', `${f}.json`), JSON.stringify({
    file: f, lang, options: {}, astJSON: ast,
    functions: figs.map((r) => ({
      name: r.name, diagram: r.diagram, svg: renderSvg(r.diagram),
      typst: renderTypst(r.diagram), excalidraw: renderExcalidraw(r.diagram),
      split: splitFromAst(ast, {}, r.name, 900).map((p) => ({ name: p.name, svg: renderSvg(p.diagram) })),
    })),
  }, null, 1));
  console.log(`  ${f}: ${figs.length} фігур`);
}
console.log('Готово.'); process.exit(0);
