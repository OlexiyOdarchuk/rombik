// Знімає еталони з Go-рушія: для кожного файлу корпусу → {functions:[{name,diagram,svg}]}.
// Це оракул парності для TS-порту.
import { Parser, Language } from 'web-tree-sitter';
import { parseTreeToAstJson } from '../src/lib/parser.js';
import { readFileSync, writeFileSync, readdirSync } from 'fs';

await Parser.init();
eval(readFileSync('./static/wasm_exec.js', 'utf8'));
const go = new globalThis.Go();
const { instance } = await WebAssembly.instantiate(readFileSync('./static/rombik.wasm'), go.importObject);
go.run(instance);
await new Promise(r => setTimeout(r, 0));

const langWasm = { py: 'tree-sitter-python.wasm', cpp: 'tree-sitter-cpp.wasm' };
const cache = {};
async function langFor(ext) {
  const key = ext === 'cpp' ? 'cpp' : 'py';
  if (!cache[key]) { const p = new Parser(); p.setLanguage(await Language.load('./static/' + langWasm[key])); cache[key] = [p, key]; }
  return cache[key];
}

for (const f of readdirSync('./test/corpus')) {
  const ext = f.endsWith('.cpp') ? 'cpp' : 'py';
  const [parser, lang] = await langFor(ext);
  const tree = parser.parse(readFileSync('./test/corpus/' + f, 'utf8'));
  const astJSON = parseTreeToAstJson(tree, lang);
  const res = JSON.parse(globalThis.rombikGenerate(astJSON, JSON.stringify({})));
  const golden = { file: f, lang, options: {}, astJSON: JSON.parse(astJSON),
    functions: res.functions.map(fn => ({ name: fn.name, diagram: fn.diagram, svg: fn.svg })) };
  writeFileSync(`./test/golden/${f}.json`, JSON.stringify(golden, null, 1));
  console.log(`  ${f}: ${res.functions.length} фігур`);
}
console.log('Еталони знято.');
process.exit(0);
