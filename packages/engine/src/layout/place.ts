// Розкладка координат — порт place()-сімейства з pkg/layout/layout.go. Рекурсивно
// роздає (x,y) вузлам IR і будує ребра-ламані. ПОРЯДОК додавання фігур/ребер має
// точно збігатися з Go — від нього залежить bodyExtent (зрізи за індексами).
import type { Diagram, Shape, Edge, Point, Kind } from '../diagram.ts';
import type { Node, Block, If, For, While, DoWhile, InfLoop } from '../ir.ts';
import type { ResolvedOptions } from './options.ts';
import { ioText, procText } from './options.ts';
import {
  boxH, diaH, termW, termH, vGap, branchGap, hGap, mergeGap, margin, hexH, arcGap,
  textW, diaW, hexW, nstmts, blockSize,
} from './measure.ts';

const P = (x: number, y: number): Point => ({ x, y });
const edge = (a: Point, b: Point): Edge => ({ points: [a, b] });

function term(cx: number, top: number, text: string): Shape {
  return { kind: 'terminator', x: cx - termW / 2, y: top, w: termW, h: termH, text };
}

function contentBottom(d: Diagram): number {
  let b = 0;
  for (const s of d.shapes) b = Math.max(b, s.y + s.h);
  for (const e of d.edges) for (const p of e.points) b = Math.max(b, p.y);
  return b;
}

// endsBlock — чи блок завершується (return/raise/exit/break, або if з обома гілками).
function endsBlock(blk: Block): boolean {
  if (blk.stmts.length === 0) return false;
  const x = blk.stmts[blk.stmts.length - 1];
  if (x.kind === 'terminal' || x.kind === 'break') return true;
  if (x.kind === 'if') return endsBlock(x.then) && endsBlock(x.else);
  return false;
}

// soleJump — Break/Continue, якщо блок це рівно один такий стрибок.
function soleJump(blk: Block): Node | null {
  if (blk.stmts.length === 1) {
    const s = blk.stmts[0];
    if (s.kind === 'break' || s.kind === 'continue') return s;
  }
  return null;
}

// termGuardLast — якщо останній стейтмент тіла це guard (if з порожнім else).
function termGuardLast(blk: Block): If | null {
  if (blk.stmts.length === 0) return null;
  const g = blk.stmts[blk.stmts.length - 1];
  if (g.kind === 'if' && nstmts(g.else) === 0 && nstmts(g.then) > 0) return g;
  return null;
}

// mapCalls — замінює виклики підпрограм на звичайні процеси (опція callAsProcess).
export function mapCalls(b: Block): void {
  b.stmts.forEach((n, i) => {
    switch (n.kind) {
      case 'call': b.stmts[i] = { kind: 'process', text: n.text }; break;
      case 'if': mapCalls(n.then); mapCalls(n.else); break;
      case 'for': case 'while': case 'dowhile': case 'infloop': mapCalls(n.body); break;
      case 'block': mapCalls(n); break;
    }
  });
}

// Builder несе стан розкладки одної схеми.
class Builder {
  d: Diagram = { shapes: [], edges: [], w: 0, h: 0 };
  ends: Point[] = [];
  loopBreaks: Point[][] = [];
  loopConts: Point[][] = [];
  o: ResolvedOptions;
  constructor(o: ResolvedOptions) { this.o = o; }

  pushLoop(): void { this.loopBreaks.push([]); this.loopConts.push([]); }
  popLoop(): [Point[], Point[]] {
    return [this.loopBreaks.pop() ?? [], this.loopConts.pop() ?? []];
  }
  recordBreak(p: Point): void { const n = this.loopBreaks.length; if (n > 0) this.loopBreaks[n - 1].push(p); }
  recordContinue(p: Point): void { const n = this.loopConts.length; if (n > 0) this.loopConts[n - 1].push(p); }

  bodyExtent(startS: number, startE: number, fallback: number): [number, number] {
    let right = fallback, left = fallback;
    for (let i = startS; i < this.d.shapes.length; i++) {
      const s = this.d.shapes[i];
      right = Math.max(right, s.x + s.w); left = Math.min(left, s.x);
    }
    for (let i = startE; i < this.d.edges.length; i++) {
      for (const p of this.d.edges[i].points) { right = Math.max(right, p.x); left = Math.min(left, p.x); }
    }
    return [right, left];
  }

  shiftX(dx: number): void {
    // Точки-об'єкти аліасяться між ребрами і ends (на відміну від Go, де Point —
    // значення-копія). Зсуваємо КОЖЕН об'єкт рівно раз, інакше спільна точка
    // (напр. ends[i], що є й кінцем ребра routeEnds) зсунеться двічі → діагональ.
    const seen = new Set<Point>();
    const shift = (p: Point) => { if (!seen.has(p)) { seen.add(p); p.x += dx; } };
    for (const s of this.d.shapes) s.x += dx;
    for (const e of this.d.edges) for (const p of e.points) shift(p);
    for (const p of this.ends) shift(p);
  }

  run(prog: Block): Diagram {
    const o = this.o;
    const [bw] = blockSize(prog, o);
    const w = bw + 2 * margin;
    const cx = w / 2;
    const top = margin;
    let bodyTop = top;
    if (!o.noStart) {
      this.d.shapes.push(term(cx, top, o.startText));
      bodyTop = top + termH + vGap;
      this.d.edges.push(edge(P(cx, top + termH), P(cx, bodyTop)));
    }
    const [exit, ended] = this.placeBlock(prog, cx, bodyTop);
    if (!ended) this.ends.push(exit);
    if (!o.noEnd && this.ends.length > 0) {
      let kY = contentBottom(this.d) + vGap;
      if (this.ends.length > 1) kY += mergeGap;
      this.routeEnds(cx, kY);
      this.d.shapes.push(term(cx, kY, o.endText));
    }
    let [right, left] = this.bodyExtent(0, 0, cx);
    const dx = margin - left;
    if (dx !== 0) { this.shiftX(dx); right += dx; }
    this.d.w = right + margin;
    this.d.h = contentBottom(this.d) + margin;
    return this.d;
  }

  routeEnds(cx: number, kY: number): void {
    if (this.ends.length === 0) return;
    if (this.ends.length === 1) {
      const e = this.ends[0];
      if (e.x > cx - 1 && e.x < cx + 1) {
        this.d.edges.push(edge(e, P(cx, kY)));
      } else {
        const my = e.y + mergeGap / 2;
        this.d.edges.push({ points: [e, { x: e.x, y: my }, { x: cx, y: my }, { x: cx, y: kY }] });
      }
      return;
    }
    const [right, left] = this.bodyExtent(0, 0, cx);
    const busRight = right + mergeGap, busLeft = left - mergeGap;
    const jy = kY - vGap;
    let minDropR = jy, minDropL = jy;
    let hasR = false, hasL = false;
    for (const e of this.ends) {
      const drop = e.y + mergeGap / 2;
      if (e.x < cx - 1) {
        if (drop < minDropL) minDropL = drop;
        this.d.edges.push({ arrowless: true, points: [e, { x: e.x, y: drop }, { x: busLeft, y: drop }] });
        hasL = true;
      } else {
        if (drop < minDropR) minDropR = drop;
        this.d.edges.push({ arrowless: true, points: [e, { x: e.x, y: drop }, { x: busRight, y: drop }] });
        hasR = true;
      }
    }
    if (hasL) this.d.edges.push({ arrowless: true, points: [{ x: busLeft, y: minDropL }, { x: busLeft, y: jy }, { x: cx, y: jy }] });
    if (hasR) this.d.edges.push({ arrowless: true, points: [{ x: busRight, y: minDropR }, { x: busRight, y: jy }, { x: cx, y: jy }] });
    this.d.edges.push(edge(P(cx, jy), P(cx, kY)));
  }

  routeBreaks(cx: number, contY: number, pts: Point[]): void {
    if (pts.length === 0) return;
    // break зливається в точку виходу циклу (contY) — як дуга виходу: БЕЗ вістря.
    // Головку дасть continuation із contY у наступну фігуру (інакше зайва головка
    // на стику). У Go тут вістря лишали — це був візуальний дефект.
    pts.forEach((p, i) => {
      if (p.x > cx - 1 && p.x < cx + 1) {
        this.d.edges.push({ arrowless: true, points: [p, { x: cx, y: contY }] });
        return;
      }
      const safeY = contY - (pts.length - 1 - i) * 7;
      const pts4: Point[] = [p, { x: p.x, y: safeY }, { x: cx, y: safeY }];
      if (safeY < contY - 0.5) pts4.push({ x: cx, y: contY });
      this.d.edges.push({ arrowless: true, points: pts4 });
    });
  }

  routeContinues(backX: number, contY: number, pts: Point[]): void {
    if (pts.length === 0) return;
    pts.forEach((p, i) => {
      const safeY = contY - vGap * 0.6 + i * 6;
      this.d.edges.push({ arrowless: true, points: [p, { x: p.x, y: safeY }, { x: backX, y: safeY }] });
    });
  }

  place(n: Node, cx: number, top: number): [Point, boolean] {
    switch (n.kind) {
      case 'process': return [this.leaf('process', procText(n.text, this.o), cx, top), false];
      case 'io': return [this.leaf('io', ioText(n.text, this.o), cx, top), false];
      case 'call': return [this.leaf('subprogram', n.text, cx, top), false];
      case 'terminal': return this.placeTerminal(n.text, cx, top);
      case 'block': return this.placeBlock(n, cx, top);
      case 'if': return this.placeIf(n, cx, top);
      case 'for': return this.placeFor(n, cx, top);
      case 'while': return this.placeWhile(n, cx, top);
      case 'dowhile': return [this.placeDoWhile(n, cx, top), false];
      case 'infloop': return [this.placeInfLoop(n, cx, top), false];
      case 'connector': {
        const cw = 46;
        this.d.shapes.push({ kind: 'connector', x: cx - cw / 2, y: top, w: cw, h: cw, text: n.text });
        // goto-стрибок: потік завершується тут (продовження — біля кружечка-мітки з тією ж літерою)
        return [P(cx, top + cw), !!n.jump];
      }
      case 'break': this.recordBreak(P(cx, top)); return [P(cx, top), true];
      case 'continue': this.recordContinue(P(cx, top)); return [P(cx, top), true];
    }
  }

  leaf(kind: Kind, text: string, cx: number, top: number): Point {
    const w = textW(text);
    this.d.shapes.push({ kind, x: cx - w / 2, y: top, w, h: boxH, text });
    return P(cx, top + boxH);
  }

  placeTerminal(text: string, cx: number, top: number): [Point, boolean] {
    const kind: Kind = this.o.returnAsIO ? 'io' : 'process';
    const exit = this.leaf(kind, text, cx, top);
    if (this.o.singleEnd) { this.ends.push(exit); return [exit, true]; }
    const endTop = exit.y + vGap;
    this.d.edges.push(edge(exit, P(cx, endTop)));
    this.d.shapes.push(term(cx, endTop, this.o.endText));
    return [P(cx, endTop + termH), true];
  }

  placeBlock(blk: Block, cx: number, top: number): [Point, boolean] {
    let cur = top;
    let exit = P(cx, top);
    for (let i = 0; i < blk.stmts.length; i++) {
      const s = blk.stmts[i];
      if (s.kind === 'break') { this.recordBreak(exit); return [exit, true]; }
      if (s.kind === 'continue') { this.recordContinue(exit); return [exit, true]; }
      if (i > 0) this.d.edges.push(edge(exit, P(cx, cur)));
      const [e, ended] = this.place(s, cx, cur);
      exit = e;
      if (ended) return [exit, true];
      cur = exit.y + vGap;
    }
    return [exit, false];
  }

  placeJumpGuard(cond: string, jump: Node, cx: number, top: number): [Point, boolean] {
    const dw = diaW(cond);
    this.d.shapes.push({ kind: 'decision', x: cx - dw / 2, y: top, w: dw, h: diaH, text: cond });
    const midY = top + diaH / 2;
    const isBreak = jump.kind === 'break';
    const side = isBreak ? -1 : 1;
    this.d.edges.push({ arrowless: true, label: this.o.yes, points: [
      { x: cx + side * dw / 2, y: midY }, { x: cx + side * (dw / 2 + hGap), y: midY },
    ] });
    const p = P(cx + side * (dw / 2 + hGap), midY);
    if (isBreak) this.recordBreak(p); else this.recordContinue(p);
    return [P(cx, top + diaH), false];
  }

  placeIf(n: If, cx: number, top: number): [Point, boolean] {
    if (nstmts(n.else) === 0) {
      const j = soleJump(n.then);
      if (j) return this.placeJumpGuard(n.cond, j, cx, top);
    }
    const dw = diaW(n.cond);
    this.d.shapes.push({ kind: 'decision', x: cx - dw / 2, y: top, w: dw, h: diaH, text: n.cond });
    const midY = top + diaH / 2;
    const branchTop = top + diaH + branchGap;
    const thenEmpty = nstmts(n.then) === 0;
    const elseEmpty = nstmts(n.else) === 0;
    const thenEnded = endsBlock(n.then);
    const elseEnded = endsBlock(n.else);
    if (elseEmpty && !thenEmpty && !thenEnded) return this.guard(n.then, this.o.yes, this.o.no, 1, cx, dw, midY, branchTop);
    if (thenEmpty && !elseEmpty && !elseEnded) return this.guard(n.else, this.o.no, this.o.yes, -1, cx, dw, midY, branchTop);
    if (elseEmpty && !thenEmpty && thenEnded) return this.guardTerm(n.then, this.o.no, this.o.yes, 1, cx, dw, midY, branchTop);
    if (thenEmpty && !elseEmpty && elseEnded) return this.guardTerm(n.else, this.o.yes, this.o.no, -1, cx, dw, midY, branchTop);
    const [tw, th] = this.branchSizeP(n.then);
    const [ew, eh] = this.branchSizeP(n.else);
    const thenCx = cx - Math.max(dw / 2 + 24, tw / 2 + hGap / 2);
    const elseCx = cx + Math.max(dw / 2 + 24, ew / 2 + hGap / 2);
    const mergeY = branchTop + Math.max(th, eh) + mergeGap;
    const te = this.branch(n.then, this.o.yes, cx, cx - dw / 2, midY, thenCx, branchTop, mergeY);
    const ee = this.branch(n.else, this.o.no, cx, cx + dw / 2, midY, elseCx, branchTop, mergeY);
    return [P(cx, mergeY), te && ee];
  }

  branchSizeP(b: Block): [number, number] {
    if (nstmts(b) === 0) return [0, 0];
    return blockSize(b, this.o);
  }

  guard(body: Block, downLabel: string, sideLabel: string, side: number, cx: number, dw: number, midY: number, branchTop: number): [Point, boolean] {
    const [bw, bh] = blockSize(body, this.o);
    this.d.edges.push({ label: downLabel, points: [{ x: cx, y: midY + diaH / 2 }, { x: cx, y: branchTop }] });
    const [exit, ended] = this.placeBlock(body, cx, branchTop);
    const mergeY = branchTop + bh + mergeGap;
    const vx = cx + side * dw / 2;
    const sideX = cx + side * Math.max(dw / 2 + 24, bw / 2 + hGap);
    this.d.edges.push({ arrowless: true, label: sideLabel, points: [
      { x: vx, y: midY }, { x: sideX, y: midY }, { x: sideX, y: mergeY }, { x: cx, y: mergeY },
    ] });
    if (!ended) this.d.edges.push({ arrowless: true, points: [exit, { x: cx, y: mergeY }] });
    return [P(cx, mergeY), false];
  }

  guardTerm(body: Block, contLabel: string, actLabel: string, side: number, cx: number, dw: number, midY: number, branchTop: number): [Point, boolean] {
    const [bw, bh] = blockSize(body, this.o);
    const actCx = cx + side * Math.max(dw / 2 + 24, bw / 2 + hGap / 2);
    this.d.edges.push({ label: actLabel, points: [
      { x: cx + side * dw / 2, y: midY }, { x: actCx, y: midY }, { x: actCx, y: branchTop },
    ] });
    this.placeBlock(body, actCx, branchTop);
    const mergeY = branchTop + bh + mergeGap;
    this.d.edges.push({ arrowless: true, label: contLabel, points: [
      { x: cx, y: midY + diaH / 2 }, { x: cx, y: mergeY },
    ] });
    return [P(cx, mergeY), false];
  }

  branch(blk: Block, label: string, cx: number, vx: number, midY: number, bcx: number, branchTop: number, mergeY: number): boolean {
    if (blk.stmts.length === 0) {
      this.d.edges.push({ arrowless: true, label, points: [
        { x: vx, y: midY }, { x: bcx, y: midY }, { x: bcx, y: mergeY }, { x: cx, y: mergeY },
      ] });
      return false;
    }
    if (blk.stmts.length === 1 && blk.stmts[0].kind === 'break') {
      this.d.edges.push({ arrowless: true, label, points: [{ x: vx, y: midY }, { x: bcx, y: midY }] });
      this.recordBreak(P(bcx, midY));
      return true;
    }
    this.d.edges.push({ label, points: [{ x: vx, y: midY }, { x: bcx, y: midY }, { x: bcx, y: branchTop }] });
    const [exit, ended] = this.placeBlock(blk, bcx, branchTop);
    if (!ended) this.d.edges.push({ arrowless: true, points: [exit, { x: exit.x, y: mergeY }, { x: cx, y: mergeY }] });
    return ended;
  }

  placeFor(n: For, cx: number, top: number): [Point, boolean] {
    const hw = hexW(n.spec);
    this.d.shapes.push({ kind: 'loop', x: cx - hw / 2, y: top, w: hw, h: hexH, text: n.spec });
    const headCy = top + hexH / 2;
    const g = termGuardLast(n.body);
    if (g) {
      const cont = this.placeGuardLoopBody(n.body, g, cx, hw / 2, headCy, top + hexH, '', '');
      return this.placeLoopElse(n.else, cx, cont);
    }
    const bodyTop = top + hexH + vGap;
    this.d.edges.push(edge(P(cx, top + hexH), P(cx, bodyTop)));
    const startS = this.d.shapes.length, startE = this.d.edges.length;
    this.pushLoop();
    const [bodyExit] = this.placeBlock(n.body, cx, bodyTop);
    const [brks, conts] = this.popLoop();
    let cont = this.loopArcs(cx, hw / 2, headCy, startS, startE, bodyExit.y, '', conts);
    const [c2, ended] = this.placeLoopElse(n.else, cx, cont);
    this.routeBreaks(cx, c2.y, brks);
    return [c2, ended];
  }

  placeWhile(n: While, cx: number, top: number): [Point, boolean] {
    const dw = diaW(n.cond);
    this.d.shapes.push({ kind: 'decision', x: cx - dw / 2, y: top, w: dw, h: diaH, text: n.cond });
    const headCy = top + diaH / 2;
    const g = termGuardLast(n.body);
    if (g) {
      const cont = this.placeGuardLoopBody(n.body, g, cx, dw / 2, headCy, top + diaH, this.o.yes, this.o.no);
      return this.placeLoopElse(n.else, cx, cont);
    }
    const bodyTop = top + diaH + vGap;
    this.d.edges.push({ label: this.o.yes, points: [{ x: cx, y: top + diaH }, { x: cx, y: bodyTop }] });
    const startS = this.d.shapes.length, startE = this.d.edges.length;
    this.pushLoop();
    const [bodyExit] = this.placeBlock(n.body, cx, bodyTop);
    const [brks, conts] = this.popLoop();
    const cont = this.loopArcs(cx, dw / 2, headCy, startS, startE, bodyExit.y, this.o.no, conts);
    const [c2, ended] = this.placeLoopElse(n.else, cx, cont);
    this.routeBreaks(cx, c2.y, brks);
    return [c2, ended];
  }

  placeDoWhile(n: DoWhile, cx: number, top: number): Point {
    const startS = this.d.shapes.length, startE = this.d.edges.length;
    this.pushLoop();
    const [bodyExit] = this.placeBlock(n.body, cx, top);
    const [brks, conts] = this.popLoop();
    const diaTop = bodyExit.y + vGap;
    const dw = diaW(n.cond);
    this.d.shapes.push({ kind: 'decision', x: cx - dw / 2, y: diaTop, w: dw, h: diaH, text: n.cond });
    const diaCy = diaTop + diaH / 2;
    this.d.edges.push(edge(bodyExit, P(cx, diaTop)));
    const [right] = this.bodyExtent(startS, startE, cx);
    const backX = right + arcGap;
    const contY = diaTop + diaH + vGap;
    this.routeContinues(backX, contY, conts);
    const mergeY = top - vGap / 2;
    this.d.edges.push({ label: this.o.no, arrowless: true, points: [
      { x: cx + dw / 2, y: diaCy }, { x: backX, y: diaCy }, { x: backX, y: mergeY }, { x: cx, y: mergeY },
    ] });
    this.d.edges.push({ label: this.o.yes, arrowless: true, points: [{ x: cx, y: diaTop + diaH }, { x: cx, y: contY }] });
    this.routeBreaks(cx, contY, brks);
    return P(cx, contY);
  }

  placeInfLoop(n: InfLoop, cx: number, top: number): Point {
    const startS = this.d.shapes.length, startE = this.d.edges.length;
    this.pushLoop();
    const [bodyExit] = this.placeBlock(n.body, cx, top);
    const [brks, conts] = this.popLoop();
    const [right] = this.bodyExtent(startS, startE, cx);
    const backX = right + arcGap;
    const contY = bodyExit.y + vGap;
    this.routeContinues(backX, contY, conts);
    const mergeY = top - vGap / 2;
    const drop = bodyExit.y + mergeGap / 2;
    this.d.edges.push({ arrowless: true, points: [
      { x: cx, y: bodyExit.y }, { x: cx, y: drop }, { x: backX, y: drop }, { x: backX, y: mergeY }, { x: cx, y: mergeY },
    ] });
    this.routeBreaks(cx, contY, brks);
    return P(cx, contY);
  }

  loopArcs(cx: number, headHalf: number, headCy: number, startS: number, startE: number, bodyBottom: number, exitLabel: string, conts: Point[]): Point {
    let [right, left] = this.bodyExtent(startS, startE, cx);
    right = Math.max(right, cx + headHalf);
    left = Math.min(left, cx - headHalf);
    const backX = right + arcGap;
    const leftX = left - arcGap;
    const contY = bodyBottom + vGap;
    this.routeContinues(backX, contY, conts);
    const drop = bodyBottom + mergeGap / 2;
    this.d.edges.push({ points: [
      { x: cx, y: bodyBottom }, { x: cx, y: drop }, { x: backX, y: drop }, { x: backX, y: headCy }, { x: cx + headHalf, y: headCy },
    ] });
    this.d.edges.push({ label: exitLabel, arrowless: true, points: [
      { x: cx - headHalf, y: headCy }, { x: leftX, y: headCy }, { x: leftX, y: contY }, { x: cx, y: contY },
    ] });
    return P(cx, contY);
  }

  placeLoopGuard(g: If, cx: number, headHalf: number, headCy: number, headBottom: number, startS: number, startE: number, entryLabel: string, exitLabel: string): [Point, number] {
    const dw = diaW(g.cond);
    const diaTop = headBottom + vGap;
    this.d.edges.push({ label: entryLabel, points: [{ x: cx, y: headBottom }, { x: cx, y: diaTop }] });
    this.d.shapes.push({ kind: 'decision', x: cx - dw / 2, y: diaTop, w: dw, h: diaH, text: g.cond });
    const diaMidY = diaTop + diaH / 2;
    const actTop = diaTop + diaH + branchGap;
    this.d.edges.push({ label: this.o.yes, points: [{ x: cx, y: diaTop + diaH }, { x: cx, y: actTop }] });
    const [exit, ended] = this.placeBlock(g.then, cx, actTop);
    let [right, left] = this.bodyExtent(startS, startE, cx);
    right = Math.max(right, cx + headHalf);
    left = Math.min(left, cx - headHalf);
    const backX = right + arcGap;
    if (ended) {
      this.d.edges.push({ label: this.o.no, points: [
        { x: cx + dw / 2, y: diaMidY }, { x: backX, y: diaMidY }, { x: backX, y: headCy }, { x: cx + headHalf, y: headCy },
      ] });
    } else {
      this.d.edges.push({ label: this.o.no, arrowless: true, points: [{ x: cx + dw / 2, y: diaMidY }, { x: backX, y: diaMidY }] });
      const drop = exit.y + mergeGap / 2;
      this.d.edges.push({ points: [
        { x: cx, y: exit.y }, { x: cx, y: drop }, { x: backX, y: drop }, { x: backX, y: headCy }, { x: cx + headHalf, y: headCy },
      ] });
    }
    const contY = contentBottom(this.d) + vGap;
    const leftX = left - arcGap;
    this.d.edges.push({ label: exitLabel, arrowless: true, points: [
      { x: cx - headHalf, y: headCy }, { x: leftX, y: headCy }, { x: leftX, y: contY }, { x: cx, y: contY },
    ] });
    return [P(cx, contY), backX];
  }

  placeGuardLoopBody(body: Block, g: If, cx: number, headHalf: number, headCy: number, headBottom: number, entryLabel: string, exitLabel: string): Point {
    this.pushLoop();
    const startS = this.d.shapes.length, startE = this.d.edges.length;
    let fromY = headBottom, el = entryLabel;
    const pre = body.stmts.slice(0, body.stmts.length - 1);
    if (pre.length > 0) {
      const bodyTop = headBottom + vGap;
      this.d.edges.push({ label: entryLabel, points: [{ x: cx, y: headBottom }, { x: cx, y: bodyTop }] });
      const [exit] = this.placeBlock({ kind: 'block', stmts: pre }, cx, bodyTop);
      fromY = exit.y; el = '';
    }
    const [cont, backX] = this.placeLoopGuard(g, cx, headHalf, headCy, fromY, startS, startE, el, exitLabel);
    const [brks, conts] = this.popLoop();
    this.routeContinues(backX, cont.y, conts);
    this.routeBreaks(cx, cont.y, brks);
    return cont;
  }

  placeLoopElse(els: Block, cx: number, cont: Point): [Point, boolean] {
    if (nstmts(els) === 0) return [cont, false];
    const top = cont.y + vGap;
    this.d.edges.push(edge(cont, P(cx, top)));
    return this.placeBlock(els, cx, top);
  }
}

// layoutProgram розкладає тіло програми у повну схему (Початок → тіло → Кінець).
export function layoutProgram(prog: Block, o: ResolvedOptions): Diagram {
  if (o.callAsProcess) mapCalls(prog);
  return new Builder(o).run(prog);
}
