// Excalidraw СЕМАНТИЧНА парність: .excalidraw — це JSON для excalidraw.com, де
// важлива структура, не байти (Go сортує ключі + екранує <>& ). Тож порівнюємо
// розпарсені об'єкти (deepEqual ігнорує порядок ключів і екранування).
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { render } from '../src/render/excalidraw.ts';

const goldenDir = join(import.meta.dirname, 'golden');

interface GoldenFn { name: string; diagram: unknown; excalidraw: string; }
interface Golden { file: string; functions: GoldenFn[]; }

for (const gf of readdirSync(goldenDir)) {
  const g: Golden = JSON.parse(readFileSync(join(goldenDir, gf), 'utf8'));
  for (const fn of g.functions) {
    test(`excalidraw: ${g.file} → ${fn.name}`, () => {
      assert.deepEqual(JSON.parse(render(fn.diagram as never)), JSON.parse(fn.excalidraw));
    });
  }
}
