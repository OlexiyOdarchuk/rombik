// Збирає standalone CLI у dist/rombik.mjs
import { build } from 'esbuild';
import { copyFileSync, mkdirSync, readFileSync, writeFileSync, chmodSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const dist = join(root, 'dist');
mkdirSync(dist, { recursive: true });
await build({
  entryPoints: [join(root, 'tools', 'rombik.mjs')],
  bundle: true, platform: 'node', format: 'esm', target: 'node22',
  conditions: ['rombik-source'], // rombik-engine → TS-джерела (не зібраний dist)
  outfile: join(dist, 'rombik.mjs'),
});
// замінити шебанг (--experimental-strip-types) на звичайний node
const out = join(dist, 'rombik.mjs');
let js = readFileSync(out, 'utf8').replace(/^#![^\n]*\n/, '');
writeFileSync(out, '#!/usr/bin/env node\n' + js);
chmodSync(out, 0o755);
// wasm поруч: ядро web-tree-sitter + граматики
copyFileSync(join(root, 'node_modules/web-tree-sitter/web-tree-sitter.wasm'), join(dist, 'web-tree-sitter.wasm'));
copyFileSync(join(root, 'web/static/tree-sitter-python.wasm'), join(dist, 'tree-sitter-python.wasm'));
copyFileSync(join(root, 'web/static/tree-sitter-cpp.wasm'), join(dist, 'tree-sitter-cpp.wasm'));
console.log('Зібрано: dist/rombik.mjs (+ 3 wasm)');
