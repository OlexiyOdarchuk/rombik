// Вимірювальний шар layout — порт size()-сімейства з pkg/layout/layout.go.
// Рахує габарити піддерева ДО роздачі координат (place). Детермінований: ширина
// тексту — оцінка charW (як у Go), без вимірювання браузером, щоб розкладка була
// однакова всюди (веб, Node) і трималась golden-парності.
import type { Node, Block } from '../ir.ts';
import type { ResolvedOptions as Options } from './options.ts';
import { ioText, procText } from './options.ts';

// Розміри й відступи (у точках) — точно як у Go.
export const boxH = 46;
export const minBoxW = 130;
export const charW = 8.5;
export const padX = 28;
export const diaH = 76;
export const minDiaW = 150;
export const termW = 130;
export const termH = 42;
export const vGap = 44;
export const branchGap = 48;
export const hGap = 56;
export const mergeGap = 44;
export const margin = 44;
export const hexH = 50;
export const arcGap = 38;

const runeLen = (s: string): number => [...s].length;

export function textW(s: string): number {
  return Math.max(runeLen(s) * charW + padX, minBoxW);
}
export function diaW(cond: string): number {
  return Math.max(minDiaW, runeLen(cond) * charW + 60);
}
export function hexW(spec: string): number {
  return Math.max(minBoxW, textW(spec) + hexH);
}

// nstmts — кількість інструкцій блоку (Block завжди не-null у TS-IR).
export function nstmts(b: Block | undefined): number {
  return b ? b.stmts.length : 0;
}

// size — габарити вузла (w, h). Порт build.size.
export function size(n: Node, o: Options): [number, number] {
  switch (n.kind) {
    case 'process': return [textW(procText(n.text, o)), boxH];
    case 'io': return [textW(ioText(n.text, o)), boxH];
    case 'call': return [textW(n.text), boxH];
    case 'terminal':
      if (o.singleEnd) return [textW(n.text), boxH];
      return [Math.max(textW(n.text), termW), boxH + vGap + termH];
    case 'block': return blockSize(n, o);
    case 'if': {
      const [tw, th] = branchSize(n.then, o);
      const [ew, eh] = branchSize(n.else, o);
      const h = diaH + branchGap + Math.max(th, eh) + mergeGap;
      if ((nstmts(n.then) === 0) !== (nstmts(n.else) === 0)) { // guard: одна гілка порожня
        return [Math.max(diaW(n.cond), Math.max(tw, ew) + 2 * hGap), h];
      }
      return [Math.max(diaW(n.cond), tw + hGap + ew), h];
    }
    case 'for': {
      const [bw, bh] = blockSize(n.body, o);
      return withElse(n.else, Math.max(hexW(n.spec), bw) + 2 * arcGap, hexH + vGap + bh + vGap, o);
    }
    case 'while': {
      const [bw, bh] = blockSize(n.body, o);
      return withElse(n.else, Math.max(diaW(n.cond), bw) + 2 * arcGap, diaH + vGap + bh + vGap, o);
    }
    case 'dowhile': {
      const [bw, bh] = blockSize(n.body, o);
      return [Math.max(diaW(n.cond), bw) + 2 * arcGap, bh + vGap + diaH + vGap];
    }
    case 'infloop': {
      const [bw, bh] = blockSize(n.body, o);
      return [bw + 2 * arcGap, bh + vGap];
    }
    case 'break':
    case 'continue': return [0, 0];
    case 'connector': return [46, 46];
  }
}

// branchSize — габарити гілки if; порожня гілка не резервує ширини.
export function branchSize(b: Block, o: Options): [number, number] {
  if (nstmts(b) === 0) return [0, 0];
  return blockSize(b, o);
}

// withElse — доповнює габарити циклу розміром гілки else (for/while), якщо вона є.
function withElse(els: Block, w: number, h: number, o: Options): [number, number] {
  if (nstmts(els) === 0) return [w, h];
  const [ew, eh] = blockSize(els, o);
  return [Math.max(w, ew), h + vGap + eh];
}

export function blockSize(b: Block | undefined, o: Options): [number, number] {
  if (!b || b.stmts.length === 0) return [minBoxW, 0];
  let w = 0, h = 0;
  b.stmts.forEach((s, i) => {
    const [sw, sh] = size(s, o);
    w = Math.max(w, sw);
    h += sh;
    if (i > 0) h += vGap;
  });
  return [w, h];
}
