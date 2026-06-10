// Поділ на частини: TS splitFromAst має дати ті самі частини (byte-svg), що Go
// rombikSplit (maxH=900). Покриває конектори, NoStart/NoEnd, підписи «частина N з M».
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { splitFromAst } from '../src/build.ts';
import { renderSvg } from '../src/index.ts';

const goldenDir = join(import.meta.dirname, 'golden');

interface GoldenFn { name: string; split: { name: string; svg: string }[]; }
interface Golden { file: string; astJSON: unknown; functions: GoldenFn[]; }

for (const gf of readdirSync(goldenDir)) {
  const g: Golden = JSON.parse(readFileSync(join(goldenDir, gf), 'utf8'));
  for (const fn of g.functions) {
    test(`split: ${g.file} → ${fn.name} (${fn.split.length} ч.)`, () => {
      const parts = splitFromAst(g.astJSON as never, {}, fn.name, 900);
      assert.equal(parts.length, fn.split.length, 'кількість частин');
      parts.forEach((p, i) => {
        assert.equal(p.name, fn.split[i].name, `назва частини #${i}`);
        assert.equal(renderSvg(p.diagram), fn.split[i].svg, `svg частини ${p.name}`);
      });
    });
  }
}
