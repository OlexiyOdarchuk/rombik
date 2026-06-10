// SVG-рендерер — порт pkg/render/svg. Малює diagram точними ДСТУ-примітивами.
import type { Diagram, Shape, Point } from '../diagram.ts';
import { captionLine, labelAnchor } from '../diagram.ts';

const CAP_GAP = 30;
const FONT_ATTR = `"'Times New Roman', 'Liberation Serif', 'DejaVu Serif', serif"`;
const ARROW_DEFS =
  `<defs><marker id="arr" markerWidth="9" markerHeight="9" refX="7.5" refY="3" orient="auto">` +
  `<path d="M0,0 L8,3 L0,6 Z" fill="#222"/></marker></defs>`;

// f/f0 — дзеркало Go %.1f / %.0f (strconv FormatFloat 'f': до найближчого, рівні —
// «до парного»). Tie детектуємо ТОЧНО: scaled має дробову частину рівно 0.5 (для
// k.d5, точно представного як double, множення на 10^prec лишається точним).
function fmtFixed(x: number, prec: number): string {
  if (Object.is(x, -0)) x = 0;
  const neg = x < 0, a = neg ? -x : x;
  // Tie детектуємо на ІСТИННОМУ значенні double (довгий toFixed), бо a*10^prec
  // саме по собі округлюється й створює фальшиві tie (напр. 146.65*10 → 1466.5).
  const ext = a.toFixed(prec + 18);
  const dot = ext.indexOf('.');
  const roundDigit = ext[dot + prec + 1];
  const rest = ext.slice(dot + prec + 2);
  if (roundDigit === '5' && /^0+$/.test(rest)) {  // рівний tie → до парного
    const keptLast = +((ext.slice(0, dot) + ext.slice(dot + 1, dot + 1 + prec)).slice(-1));
    const m = Math.pow(10, prec);
    const n = Math.floor(a * m);                  // a*m == n+0.5 точно для істинного tie
    const r = keptLast % 2 === 1 ? n + 1 : n;
    const s = (r / m).toFixed(prec);
    return neg && r !== 0 ? '-' + s : s;
  }
  const s = a.toFixed(prec);                      // не tie → toFixed = до найближчого
  return neg && a !== 0 ? '-' + s : s;
}
const f = (n: number) => fmtFixed(n, 1);
const f0 = (n: number) => fmtFixed(n, 0);

function esc(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function pathOf(pts: Point[]): string {
  let s = '';
  for (let i = 0; i < pts.length; i++) {
    s += (i === 0 ? 'M' : 'L') + f(pts[i].x) + ' ' + f(pts[i].y) + ' ';
  }
  return s.trim();
}

function diagHeight(d: Diagram): number {
  return captionLine(d) !== '' ? d.h + CAP_GAP : d.h;
}

function renderShape(b: string[], s: Shape): void {
  const stroke = 'fill="#fdfdfd" stroke="#222" stroke-width="1.5"';
  switch (s.kind) {
    case 'terminator':
      b.push(`<rect x="${f(s.x)}" y="${f(s.y)}" width="${f(s.w)}" height="${f(s.h)}" rx="${f(s.h / 2)}" ${stroke}/>`);
      break;
    case 'process':
      b.push(`<rect x="${f(s.x)}" y="${f(s.y)}" width="${f(s.w)}" height="${f(s.h)}" ${stroke}/>`);
      break;
    case 'decision': {
      const cx = s.x + s.w / 2, cy = s.y + s.h / 2;
      b.push(`<polygon points="${f(cx)},${f(s.y)} ${f(s.x + s.w)},${f(cy)} ${f(cx)},${f(s.y + s.h)} ${f(s.x)},${f(cy)}" ${stroke}/>`);
      break;
    }
    case 'io': {
      const sk = s.h * 0.4;
      b.push(`<polygon points="${f(s.x + sk)},${f(s.y)} ${f(s.x + s.w)},${f(s.y)} ${f(s.x + s.w - sk)},${f(s.y + s.h)} ${f(s.x)},${f(s.y + s.h)}" ${stroke}/>`);
      break;
    }
    case 'loop': {
      const sk = s.h * 0.5, cy = s.y + s.h / 2;
      b.push(`<polygon points="${f(s.x + sk)},${f(s.y)} ${f(s.x + s.w - sk)},${f(s.y)} ${f(s.x + s.w)},${f(cy)} ${f(s.x + s.w - sk)},${f(s.y + s.h)} ${f(s.x + sk)},${f(s.y + s.h)} ${f(s.x)},${f(cy)}" ${stroke}/>`);
      break;
    }
    case 'subprogram': {
      const inn = 9;
      b.push(`<rect x="${f(s.x)}" y="${f(s.y)}" width="${f(s.w)}" height="${f(s.h)}" ${stroke}/>`);
      b.push(`<line x1="${f(s.x + inn)}" y1="${f(s.y)}" x2="${f(s.x + inn)}" y2="${f(s.y + s.h)}" stroke="#222" stroke-width="1.5"/>`);
      b.push(`<line x1="${f(s.x + s.w - inn)}" y1="${f(s.y)}" x2="${f(s.x + s.w - inn)}" y2="${f(s.y + s.h)}" stroke="#222" stroke-width="1.5"/>`);
      break;
    }
    case 'connector': {
      const cx = s.x + s.w / 2, cy = s.y + s.h / 2;
      b.push(`<circle cx="${f(cx)}" cy="${f(cy)}" r="${f(Math.min(s.w, s.h) / 2)}" ${stroke}/>`);
      break;
    }
  }
  b.push(`<text x="${f(s.x + s.w / 2)}" y="${f(s.y + s.h / 2)}" text-anchor="middle" dominant-baseline="middle" fill="#111">${esc(s.text)}</text>`);
}

function writeBody(b: string[], d: Diagram): void {
  const cap = captionLine(d);
  for (const e of d.edges) {
    const marker = e.arrowless ? '' : ' marker-end="url(#arr)"';
    b.push(`<path d="${pathOf(e.points)}" fill="none" stroke="#222" stroke-width="1.5"${marker}/>`);
    if (e.label && e.points.length >= 2) {
      const { x, y, align } = labelAnchor(e.points[0], e.points[1]);
      b.push(`<text x="${f(x)}" y="${f(y)}" text-anchor="${align}" font-size="12" fill="#444">${esc(e.label)}</text>`);
    }
  }
  for (const s of d.shapes) renderShape(b, s);
  if (cap !== '') {
    b.push(`<text x="${f(d.w / 2)}" y="${f(d.h + CAP_GAP * 0.6)}" text-anchor="middle" font-size="14" fill="#111">${esc(cap)}</text>`);
  }
}

export function render(d: Diagram): string {
  const totalH = diagHeight(d);
  const b: string[] = [];
  b.push(`<svg xmlns="http://www.w3.org/2000/svg" width="${f0(d.w)}" height="${f0(totalH)}" viewBox="0 0 ${f0(d.w)} ${f0(totalH)}" font-family=${FONT_ATTR} font-size="14">`);
  b.push(ARROW_DEFS);
  b.push('<rect width="100%" height="100%" fill="#ffffff"/>');
  writeBody(b, d);
  b.push('</svg>');
  return b.join('');
}
