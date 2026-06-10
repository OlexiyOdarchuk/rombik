// Бандлить CLI (рушій + tree-sitter-обгортку) в один самодостатній файл і кладе
// поруч 3 wasm-граматики. Джерело CLI — спільне tools/rombik.mjs; рушій береться
// з джерел (умова експорту rombik-source). Результат → packages/cli/dist/, що й
// публікується як пакет `rombik` (bin → npx rombik).
import { build } from 'esbuild';
import { copyFileSync, mkdirSync, readFileSync, writeFileSync, chmodSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const root = join(here, '..', '..');
const dist = join(here, 'dist');
mkdirSync(dist, { recursive: true });

await build({
  entryPoints: [join(root, 'tools', 'rombik.mjs')],
  bundle: true,
  platform: 'node',
  format: 'esm',
  target: 'node22',
  conditions: ['rombik-source'], // rombik-engine → TS-джерела (бандлимо), не dist
  outfile: join(dist, 'rombik.mjs'),
});

// шебанг для прямого запуску (npx/bin); прибрати TS-флаги зі споживчого шебангу
const out = join(dist, 'rombik.mjs');
const js = readFileSync(out, 'utf8').replace(/^#![^\n]*\n/, '');
writeFileSync(out, '#!/usr/bin/env node\n' + js);
chmodSync(out, 0o755);

// wasm поруч (rombik.mjs шукає їх біля себе через selfDir)
copyFileSync(join(root, 'node_modules/web-tree-sitter/web-tree-sitter.wasm'), join(dist, 'web-tree-sitter.wasm'));
copyFileSync(join(root, 'web/static/tree-sitter-python.wasm'), join(dist, 'tree-sitter-python.wasm'));
copyFileSync(join(root, 'web/static/tree-sitter-cpp.wasm'), join(dist, 'tree-sitter-cpp.wasm'));
copyFileSync(join(root, 'web/static/tree-sitter-pascal.wasm'), join(dist, 'tree-sitter-pascal.wasm'));

console.log('Зібрано: packages/cli/dist/rombik.mjs (+ 3 wasm)');
