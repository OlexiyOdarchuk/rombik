// tsc із rewriteRelativeImportExtensions переписує .ts→.js у .js-емісії, але НЕ у
// .d.ts (відома межа TS). Доганяємо вручну: у всіх dist/**/*.d.ts відносні
// специфікатори './x.ts' → './x.js', щоб типи резолвились у споживача.
import { readdirSync, readFileSync, writeFileSync, statSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const dist = join(dirname(fileURLToPath(import.meta.url)), '..', 'dist');
const RE = /(from\s+['"])(\.[^'"]+)\.ts(['"])/g;

let fixed = 0;
function walk(dir) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) walk(p);
    else if (name.endsWith('.d.ts')) {
      const src = readFileSync(p, 'utf8');
      const out = src.replace(RE, '$1$2.js$3');
      if (out !== src) { writeFileSync(p, out); fixed++; }
    }
  }
}
walk(dist);
console.log(`fix-dts: переписано .ts→.js у ${fixed} .d.ts-файлах`);
