// Typst byte-парність: TS-рендерер має давати той самий Typst-код, що Go.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { render } from '../src/render/typst.ts';

const goldenDir = join(import.meta.dirname, 'golden');

interface GoldenFn { name: string; diagram: unknown; typst: string; }
interface Golden { file: string; functions: GoldenFn[]; }

for (const gf of readdirSync(goldenDir)) {
  const g: Golden = JSON.parse(readFileSync(join(goldenDir, gf), 'utf8'));
  for (const fn of g.functions) {
    test(`typst: ${g.file} → ${fn.name}`, () => {
      assert.equal(render(fn.diagram as never), fn.typst);
    });
  }
}
