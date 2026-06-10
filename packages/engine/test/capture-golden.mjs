// МІГРАЦІЙНИЙ оракул: astJSON дає TS-парсер (валідований), а layout/svg/typst/excal/
// split — Go-рушій (rombik.wasm) як незалежний еталон. Так golden-тести стережуть
// парність TS-layout проти Go. Прибрати, коли Go видалено (golden заморозяться).
import { readFileSync, writeFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { createRequire } from 'node:module';
import { parseTree } from '../src/parser/treesitter.ts';
const here = import.meta.dirname, web = join(here,'..','..','..','web');
const req = createRequire(join(web,'package.json'));
const ts = await import(req.resolve('web-tree-sitter')); await ts.Parser.init();
globalThis.eval(readFileSync(join(web,'static','wasm_exec.js'),'utf8'));
const go = new globalThis.Go();
const { instance } = await WebAssembly.instantiate(readFileSync(join(web,'static','rombik.wasm')), go.importObject);
go.run(instance); await new Promise(r=>setTimeout(r,0));
const cache={};
async function P(lang){ if(!cache[lang]){const p=new ts.Parser(); p.setLanguage(await ts.Language.load(join(web,'static',lang==='cpp'?'tree-sitter-cpp.wasm':'tree-sitter-python.wasm'))); cache[lang]=p;} return cache[lang]; }
for (const f of readdirSync(join(here,'corpus'))){
  const lang = f.endsWith('.cpp')?'cpp':'py'; const tlang = lang==='cpp'?'cpp':'python';
  const tree = (await P(tlang)).parse(readFileSync(join(here,'corpus',f),'utf8'));
  const astArr = parseTree(tree, tlang);
  const astJSON = JSON.stringify(astArr);
  const res = JSON.parse(globalThis.rombikGenerate(astJSON, '{}'));
  writeFileSync(join(here,'golden',`${f}.json`), JSON.stringify({
    file:f, lang:tlang, options:{}, astJSON:astArr,
    functions: res.functions.map(fn=>{ const dj=JSON.stringify(fn.diagram);
      return { name:fn.name, diagram:fn.diagram, svg:fn.svg,
        typst: JSON.parse(globalThis.rombikTypstOne(dj,false)).typst,
        excalidraw: JSON.parse(globalThis.rombikExcalidraw(dj)).excalidraw,
        split: (JSON.parse(globalThis.rombikSplit(astJSON,'{}',fn.name,900)).parts||[]).map(p=>({name:p.name,svg:p.svg})) };
    })
  }, null, 1));
  console.log(`  ${f}: ${res.functions.length} фігур`);
}
console.log('Готово.'); process.exit(0);
