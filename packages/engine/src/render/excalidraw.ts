// Excalidraw-експорт — порт pkg/render/excalidraw. Кожна фігура → елемент
// (rectangle/diamond/ellipse/line) + текст у спільній групі; ребро → arrow.
// Координати ті самі, що в SVG (вісь y вниз). Детермінований (без рандому/годинника):
// seed від лічильника елементів — щоб працювало будь-де і відтворювалось.
import type { Diagram, Shape, Edge } from '../diagram.ts';
import { labelAnchor, captionLine } from '../diagram.ts';

const STROKE = '#1e1e1e';
const UPDATED = 1700000000000;
const GAP = 44;

type El = Record<string, unknown>;
type Pt = [number, number];

const runeLen = (s: string): number => [...s].length;

class Builder {
  n = 0;
  els: El[] = [];

  id(): string { this.n++; return `r${this.n}`; }
  add(m: El): void { this.els.push(m); }

  base(typ: string, x: number, y: number, w: number, h: number, group: string): El {
    const seed = this.n * 100003;
    const gids = group !== '' ? [group] : [];
    const id = this.id();
    return {
      id, type: typ, x, y, width: w, height: h,
      angle: 0, strokeColor: STROKE, backgroundColor: 'transparent',
      fillStyle: 'solid', strokeWidth: 1.5, strokeStyle: 'solid',
      roughness: 1, opacity: 100, groupIds: gids, frameId: null,
      roundness: null, seed, version: 1, versionNonce: seed,
      isDeleted: false, boundElements: null, updated: UPDATED, link: null, locked: false,
    };
  }

  text(s: string, cx: number, cy: number, group: string, size: number): void {
    if (s === '') return;
    const w = runeLen(s) * size * 0.55 + 6;
    const h = size * 1.25;
    const t = this.base('text', cx - w / 2, cy - h / 2, w, h, group);
    t.text = s;
    t.fontSize = size;
    t.fontFamily = 1;
    t.textAlign = 'center';
    t.verticalAlign = 'middle';
    t.containerId = null;
    t.originalText = s;
    t.lineHeight = 1.25;
    this.add(t);
  }

  diagram(d: Diagram, ox: number, oy: number): void {
    for (const e of d.edges) this.arrow(e, ox, oy);
    for (const s of d.shapes) this.shape(s, ox, oy);
    const cap = captionLine(d);
    if (cap !== '') this.text(cap, ox + d.w / 2, oy + d.h + 15, '', 16);
  }

  shape(s: Shape, ox: number, oy: number): void {
    const x = s.x + ox, y = s.y + oy, w = s.w, h = s.h;
    const cx = x + w / 2, cy = y + h / 2;
    const g = `g${this.n}`;
    switch (s.kind) {
      case 'process': this.add(this.base('rectangle', x, y, w, h, g)); break;
      case 'decision': this.add(this.base('diamond', x, y, w, h, g)); break;
      case 'terminator': {
        const m = this.base('rectangle', x, y, w, h, g);
        m.roundness = { type: 3 };
        this.add(m);
        break;
      }
      case 'connector': {
        const r = h < w ? h : w;
        this.add(this.base('ellipse', cx - r / 2, cy - r / 2, r, r, g));
        break;
      }
      case 'io': {
        const sk = h * 0.4;
        this.polygon(x, y, w, h, [[sk, 0], [w, 0], [w - sk, h], [0, h]], g);
        break;
      }
      case 'loop': {
        const sk = h * 0.5;
        this.polygon(x, y, w, h, [[sk, 0], [w - sk, 0], [w, h / 2], [w - sk, h], [sk, h], [0, h / 2]], g);
        break;
      }
      case 'subprogram':
        this.add(this.base('rectangle', x, y, w, h, g));
        this.lineSeg(x + 9, y, x + 9, y + h, g);
        this.lineSeg(x + w - 9, y, x + w - 9, y + h, g);
        break;
    }
    this.text(s.text, cx, cy, g, 16);
  }

  polygon(x: number, y: number, w: number, h: number, pts: Pt[], g: string): void {
    const m = this.base('line', x, y, w, h, g);
    m.points = [...pts, pts[0]];
    m.lastCommittedPoint = null;
    this.add(m);
  }

  lineSeg(x1: number, y1: number, x2: number, y2: number, g: string): void {
    const m = this.base('line', x1, y1, x2 - x1, y2 - y1, g);
    m.points = [[0, 0], [x2 - x1, y2 - y1]];
    m.lastCommittedPoint = null;
    this.add(m);
  }

  arrow(e: Edge, ox: number, oy: number): void {
    if (e.points.length < 2) return;
    const x0 = e.points[0].x + ox, y0 = e.points[0].y + oy;
    let minX = 0, minY = 0, maxX = 0, maxY = 0;
    const rel: Pt[] = e.points.map((p) => {
      const dx = (p.x + ox) - x0, dy = (p.y + oy) - y0;
      minX = Math.min(minX, dx); minY = Math.min(minY, dy);
      maxX = Math.max(maxX, dx); maxY = Math.max(maxY, dy);
      return [dx, dy];
    });
    const pts: Pt[] = rel.map((r) => [r[0] - minX, r[1] - minY]);
    const m = this.base('arrow', x0 + minX, y0 + minY, maxX - minX, maxY - minY, '');
    m.points = pts;
    m.lastCommittedPoint = null;
    m.startBinding = null;
    m.endBinding = null;
    m.startArrowhead = null;
    m.endArrowhead = e.arrowless ? null : 'arrow';
    this.add(m);
    if (e.label && e.points.length >= 2) {
      const { x: lx, y: ly } = labelAnchor(e.points[0], e.points[1]);
      this.text(e.label, lx + ox, ly + oy, '', 14);
    }
  }
}

function diagHeight(d: Diagram): number {
  return captionLine(d) !== '' ? d.h + 30 : d.h;
}

function doc(els: El[]): string {
  return JSON.stringify({
    type: 'excalidraw',
    version: 2,
    source: 'https://github.com/OlexiyOdarchuk/rombik',
    elements: els,
    appState: { gridSize: null, viewBackgroundColor: '#ffffff' },
    files: {},
  }, null, 2);
}

export function render(d: Diagram): string {
  const b = new Builder();
  b.diagram(d, 0, 0);
  return doc(b.els);
}

export function renderAll(ds: Diagram[]): string {
  const b = new Builder();
  let y = 0;
  for (const d of ds) { b.diagram(d, 0, y); y += diagHeight(d) + GAP; }
  return doc(b.els);
}
