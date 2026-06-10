// Звіряє TS svg.render(diagram) з еталонним SVG Go-рушія, байт-у-байт.
import { render } from '../src/lib/engine/svg.ts';
import { readFileSync, readdirSync } from 'fs';

let total = 0, pass = 0;
const fails = [];
for (const gf of readdirSync('./test/golden')) {
  const g = JSON.parse(readFileSync('./test/golden/' + gf, 'utf8'));
  for (const fn of g.functions) {
    total++;
    const got = render(fn.diagram);
    if (got === fn.svg) { pass++; continue; }
    // знайти першу розбіжність
    let i = 0; while (i < got.length && i < fn.svg.length && got[i] === fn.svg[i]) i++;
    fails.push({ file: g.file, name: fn.name,
      pos: i,
      exp: fn.svg.slice(Math.max(0, i - 30), i + 40),
      got: got.slice(Math.max(0, i - 30), i + 40),
      lenExp: fn.svg.length, lenGot: got.length });
  }
}
console.log(`SVG-парність: ${pass}/${total}`);
for (const x of fails.slice(0, 8)) {
  console.log(`\n  ✘ ${x.file}:${x.name}  @${x.pos}  (len exp=${x.lenExp} got=${x.lenGot})`);
  console.log(`    EXP …${x.exp}…`);
  console.log(`    GOT …${x.got}…`);
}
process.exit(fails.length ? 1 : 0);
