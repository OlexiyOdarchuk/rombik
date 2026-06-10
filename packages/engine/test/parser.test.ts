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
