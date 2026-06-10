#!/usr/bin/env -S node --experimental-strip-types
// rombik CLI — код → блок-схема, на тому самому TS-рушії (rombik-engine).
// Запуск:  node --experimental-strip-types tools/rombik.mjs <файл> [опції]
//    або:  ./tools/rombik.mjs <файл> [опції]   (після chmod +x)
//    або:  npm run cli -- <файл> [опції]
import { parseArgs } from 'node:util';
import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { join, dirname, extname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { Parser, Language } from 'web-tree-sitter';
import { fromTree, parseTree, splitFromAst, renderSvg, renderTypst, renderExcalidraw } from 'rombik-engine';

// wasm-граматики: біля скрипта (зібраний dist/) або у web/static (dev із кореня).
const selfDir = dirname(fileURLToPath(import.meta.url));
const grammars = existsSync(join(selfDir, 'tree-sitter-python.wasm'))
  ? selfDir
  : join(selfDir, '..', 'web', 'static');

const { values, positionals } = parseArgs({
  allowPositionals: true,
  options: {
    out:          { type: 'string',  short: 'o' },                 // вихід (default: stdout)
    format:       { type: 'string',  short: 't', default: 'svg' }, // svg | typ | json | excalidraw
    lang:         { type: 'string',  short: 'l' },                 // py | cpp | c (default: за розширенням)
    fn:           { type: 'string' },                              // лише функція з цим іменем
    'single-end': { type: 'boolean', default: false },             // один спільний Кінець
    split:        { type: 'string' },                              // розбити на частини ≤ N (висоти)
    help:         { type: 'boolean', short: 'h', default: false },
  },
});

function help() {
  console.log(`rombik — код (Python/C++) → блок-схема ДСТУ 19.701-90

Використання:
  rombik <файл> [опції]        ( '-' = читати stdin )

Опції:
  -o, --out FILE     вихід (без нього — stdout); кілька функцій → FILE_<імʼя>.<ext>
  -t, --format FMT   svg | typ | json | excalidraw            (типово svg)
  -l, --lang LANG    py | cpp | c                              (типово за розширенням)
      --fn NAME      малювати лише функцію NAME
      --single-end   один спільний «Кінець» (інакше — на кожен return/raise)
      --split N      розбити схему на частини, не вищі за N
  -h, --help

Приклади:
  rombik examples/grade.py -o grade.svg
  rombik prog.cpp -t typ > prog.typ
  cat prog.py | rombik - -t json

PNG / PDF (растр — зовнішнім конвертером SVG):
  rombik grade.py | rsvg-convert -f pdf -o grade.pdf
  rombik grade.py | rsvg-convert -f png -o grade.png      # також: cairosvg, inkscape`);
}

if (values.help || positionals.length === 0) { help(); process.exit(0); }

const file = positionals[0];
const src = file === '-' ? readFileSync(0, 'utf8') : readFileSync(file, 'utf8');
const ext = file === '-' ? '' : extname(file).toLowerCase();
// C — підмножина C++: парситься тією ж граматикою (lang='cpp').
const cppExt = ['.cpp', '.cc', '.cxx', '.hpp', '.h', '.cs', '.c'];
const lang = values.lang === 'py' ? 'python'
  : values.lang === 'cpp' || values.lang === 'c' ? 'cpp'
  : cppExt.includes(ext) ? 'cpp' : 'python';

// У зібраному dist/ ядро web-tree-sitter лежить поруч (locateFile); у dev його
// знаходить node_modules сам.
await Parser.init(existsSync(join(grammars, 'web-tree-sitter.wasm')) ? { locateFile: (f) => join(grammars, f) } : undefined);
const parser = new Parser();
parser.setLanguage(await Language.load(join(grammars, lang === 'cpp' ? 'tree-sitter-cpp.wasm' : 'tree-sitter-python.wasm')));
const tree = parser.parse(src);
const opts = { singleEnd: values['single-end'] };

let results;
if (values.split) {
  const ast = parseTree(tree, lang);
  const name = values.fn ?? ast[0]?.name;
  results = splitFromAst(ast, opts, name, Number(values.split));
} else {
  results = fromTree(tree, lang, opts);
  if (values.fn) results = results.filter((r) => r.name === values.fn);
}
if (results.length === 0) { console.error('Нічого не знайдено (перевір --fn / синтаксис).'); process.exit(1); }

const render = (d) => {
  switch (values.format) {
    case 'svg': return renderSvg(d);
    case 'typ': case 'typst': return renderTypst(d);
    case 'excalidraw': return renderExcalidraw(d);
    case 'json': return JSON.stringify(d, null, 1);
    default: throw new Error('невідомий формат: ' + values.format);
  }
};

if (!values.out || values.out === '-') {
  process.stdout.write(results.map((r) => render(r.diagram)).join('\n'));
} else if (results.length === 1) {
  writeFileSync(values.out, render(results[0].diagram));
  console.error(`Готово: ${values.out}`);
} else {
  const base = values.out.replace(/\.[^.]+$/, '');
  const ex = extname(values.out) || '.' + values.format;
  for (const r of results) writeFileSync(`${base}_${r.name}${ex}`, render(r.diagram));
  console.error(`Готово: ${results.length} схем(и) → ${base}_*${ex}`);
}
