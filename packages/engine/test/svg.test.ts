// Golden-парність: TS-рендерер має давати БАЙТ-У-БАЙТ той самий SVG, що Go-рушій.
// Еталони в test/golden/*.json зняті з Go (capture-golden.mjs) — оракул міграції.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { render } from '../src/render/svg.ts';

const goldenDir = join(import.meta.dirname, 'golden');

interface GoldenFn { name: string; diagram: unknown; svg: string; }
interface Golden { file: string; functions: GoldenFn[]; }

for (const gf of readdirSync(goldenDir)) {
  const g: Golden = JSON.parse(readFileSync(join(goldenDir, gf), 'utf8'));
  for (const fn of g.functions) {
    test(`svg: ${g.file} → ${fn.name}`, () => {
      assert.equal(render(fn.diagram as never), fn.svg);
    });
  }
}
