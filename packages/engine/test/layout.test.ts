// End-to-end golden-парність: AST-JSON → layout → svg має дати БАЙТ-У-БАЙТ той самий
// вивід, що Go-рушій. Це перевіряє серце (layout) проти еталонів зі старого Go.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { fromAst } from '../src/build.ts';
import { renderSvg } from '../src/index.ts';

const goldenDir = join(import.meta.dirname, 'golden');

interface Golden {
  file: string;
  options: Record<string, unknown>;
  astJSON: unknown;
  functions: { name: string; svg: string }[];
}

for (const gf of readdirSync(goldenDir)) {
  const g: Golden = JSON.parse(readFileSync(join(goldenDir, gf), 'utf8'));
  test(`e2e: ${g.file}`, () => {
    const results = fromAst(g.astJSON as never, g.options);
    assert.equal(results.length, g.functions.length, 'кількість функцій');
    results.forEach((r, i) => {
      const exp = g.functions[i];
      assert.equal(r.name, exp.name, `назва #${i}`);
      assert.equal(renderSvg(r.diagram), exp.svg, `svg → ${r.name}`);
    });
  });
}
