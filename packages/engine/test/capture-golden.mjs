// МІГРАЦІЙНИЙ ІНСТРУМЕНТ (тимчасовий). Знімає golden-еталони з Go-рушія як оракул
// парності для TS-порту: для кожного файлу test/corpus/ → {astJSON, diagram, svg}.
// Тягне «старий світ» із ../../../web: web-tree-sitter, parser.js, rombik.wasm.
// Запуск: node packages/engine/test/capture-golden.mjs (потрібні web/node_modules і
// зібраний web/static/rombik.wasm — `./web/build-wasm.sh`).
// Прибрати, коли layout портовано і оракулом стануть самі коміченні снапшоти.
import { readFileSync, writeFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { createRequire } from 'node:module';

const here = import.meta.dirname;
const web = join(here, '..', '..', '..', 'web');
const require = createRequire(join(web, 'package.json'));
const { Parser, Language } = await import(require.resolve('web-tree-sitter'));
const { parseTreeToAstJson } = await import(join(web, 'src', 'lib', 'parser.js'));

await Parser.init();
globalThis.eval(readFileSync(join(web, 'static', 'wasm_exec.js'), 'utf8'));
const go = new globalThis.Go();
const { instance } = await WebAssembly.instantiate(readFileSync(join(web, 'static', 'rombik.wasm')), go.importObject);
go.run(instance);
await new Promise((r) => setTimeout(r, 0));

const langWasm = { py: 'tree-sitter-python.wasm', cpp: 'tree-sitter-cpp.wasm' };
const cache = {};
async function parserFor(ext) {
  const key = ext === 'cpp' ? 'cpp' : 'py';
  if (!cache[key]) {
    const p = new Parser();
    p.setLanguage(await Language.load(join(web, 'static', langWasm[key])));
    cache[key] = [p, key];
  }
  return cache[key];
}

const corpus = join(here, 'corpus');
const golden = join(here, 'golden');
for (const f of readdirSync(corpus)) {
  const ext = f.endsWith('.cpp') ? 'cpp' : 'py';
  const [parser, lang] = await parserFor(ext);
  const tree = parser.parse(readFileSync(join(corpus, f), 'utf8'));
  const astJSON = parseTreeToAstJson(tree, lang);
  const res = JSON.parse(globalThis.rombikGenerate(astJSON, JSON.stringify({})));
  writeFileSync(join(golden, `${f}.json`), JSON.stringify({
    file: f, lang, options: {}, astJSON: JSON.parse(astJSON),
    functions: res.functions.map((fn) => ({ name: fn.name, diagram: fn.diagram, svg: fn.svg })),
  }, null, 1));
  console.log(`  ${f}: ${res.functions.length} фігур`);
}
console.log('Еталони знято.');
process.exit(0);
