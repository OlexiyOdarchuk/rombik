// Typst-рендерер (через CeTZ) — порт pkg/render/typst. Вивід — Typst-код, що
// компілюється у вектор. Координати ті самі, що в SVG (вісь y CeTZ — вгору, тож fy).
// Має збігатися БАЙТ-У-БАЙТ із Go (golden).
import type { Diagram, Shape } from '../diagram.ts';
import { labelAnchor, captionLine, capSupplement, capSeparator, capHasWord } from '../diagram.ts';
import { f1 as f } from '../format.ts';

const font = 'font: ("Times New Roman", "Liberation Serif", "DejaVu Serif"), ';

// goQuote — дзеркало Go strconv.Quote / fmt %q: "..." з екрануванням. Друковані
// Unicode-символи (кирилиця, «», оператори) лишаються як є.
function goQuote(s: string): string {
  let out = '"';
  for (const ch of s) {
    const c = ch.codePointAt(0) as number;
    switch (ch) {
      case '"': out += '\\"'; continue;
      case '\\': out += '\\\\'; continue;
      case '\n': out += '\\n'; continue;
      case '\r': out += '\\r'; continue;
      case '\t': out += '\\t'; continue;
      case '\x07': out += '\\a'; continue;
      case '\b': out += '\\b'; continue;
      case '\f': out += '\\f'; continue;
      case '\v': out += '\\v'; continue;
    }
    if (c < 0x20 || c === 0x7f) out += '\\x' + c.toString(16).padStart(2, '0');
    else out += ch;
  }
  return out + '"';
}

export function render(d: Diagram): string {
  return preamble(d) + figure(d);
}

export function renderAll(ds: Diagram[]): string {
  if (ds.length === 0) return '';
  let s = preamble(ds[0]);
  ds.forEach((d, i) => { if (i > 0) s += '#pagebreak()\n'; s += figure(d); });
  return s;
}

export function fragment(d: Diagram): string { return renderCanvas(d); }
export function fragmentAll(ds: Diagram[]): string {
  return ds.map(renderCanvas).join('\n');
}

function preamble(d: Diagram): string {
  let supplement = '[' + capSupplement(d) + ']';
  if (!capHasWord(d)) supplement = 'none';
  return '#import "@preview/cetz:0.3.4"\n' +
    '#set page(width: auto, height: auto, margin: 14pt)\n' +
    '#set text(' + font + ')\n' +
    `#set figure.caption(separator: [${capSeparator(d)}])\n` +
    '#let flowchart(body, caption: none) = figure(\n' +
    '  body, caption: caption, supplement: ' + supplement + ', kind: "flowchart", numbering: "1",\n)\n';
}

function figure(d: Diagram): string {
  const canvas = renderCanvas(d);
  if (d.caption) return `#flowchart(caption: [#${goQuote(d.caption)}])[\n${canvas}]\n`;
  return canvas;
}

function renderCanvas(d: Diagram): string {
  const fy = (y: number) => d.h - y;
  let b = '#cetz.canvas(length: 1pt, {\n  import cetz.draw: *\n  set-style(stroke: 1.5pt, fill: none)\n';
  for (const e of d.edges) {
    const pts = e.points.map((p) => `(${f(p.x)}, ${f(fy(p.y))})`).join(', ');
    const mark = e.arrowless ? '' : ', mark: (end: ">")';
    b += `  line(${pts}${mark})\n`;
    if (e.label && e.points.length >= 2) {
      const { x: lx, y: ly, align } = labelAnchor(e.points[0], e.points[1]);
      const anchor = align === 'start' ? 'west' : align === 'end' ? 'east' : 'center';
      b += `  content((${f(lx)}, ${f(fy(ly))}), text(${font}size: 12pt)[#${goQuote(e.label)}], anchor: ${goQuote(anchor)})\n`;
    }
  }
  for (const s of d.shapes) b += shape(s, fy);
  return b + '})\n';
}

function shape(s: Shape, fy: (y: number) => number): string {
  const x1 = s.x, y1 = s.y, x2 = s.x + s.w, y2 = s.y + s.h;
  const cx = s.x + s.w / 2, cy = s.y + s.h / 2;
  const fill = ', fill: white';
  const poly = (...pts: [number, number][]) =>
    `  line(${pts.map((p) => `(${f(p[0])}, ${f(p[1])})`).join(', ')}, close: true${fill})\n`;
  let b = '';
  switch (s.kind) {
    case 'terminator':
      b += `  rect((${f(x1)}, ${f(fy(y2))}), (${f(x2)}, ${f(fy(y1))}), radius: ${f(s.h / 2)}${fill})\n`;
      break;
    case 'process':
      b += `  rect((${f(x1)}, ${f(fy(y2))}), (${f(x2)}, ${f(fy(y1))})${fill})\n`;
      break;
    case 'decision':
      b += poly([cx, fy(y1)], [x2, fy(cy)], [cx, fy(y2)], [x1, fy(cy)]);
      break;
    case 'io': {
      const sk = s.h * 0.4;
      b += poly([x1 + sk, fy(y1)], [x2, fy(y1)], [x2 - sk, fy(y2)], [x1, fy(y2)]);
      break;
    }
    case 'loop': {
      const sk = s.h * 0.5;
      b += poly([x1 + sk, fy(y1)], [x2 - sk, fy(y1)], [x2, fy(cy)], [x2 - sk, fy(y2)], [x1 + sk, fy(y2)], [x1, fy(cy)]);
      break;
    }
    case 'subprogram':
      b += `  rect((${f(x1)}, ${f(fy(y2))}), (${f(x2)}, ${f(fy(y1))})${fill})\n`;
      b += `  line((${f(x1 + 9)}, ${f(fy(y1))}), (${f(x1 + 9)}, ${f(fy(y2))}))\n`;
      b += `  line((${f(x2 - 9)}, ${f(fy(y1))}), (${f(x2 - 9)}, ${f(fy(y2))}))\n`;
      break;
    case 'connector': {
      const r = s.h < s.w ? s.h / 2 : s.w / 2;
      b += `  circle((${f(cx)}, ${f(fy(cy))}), radius: ${f(r)}${fill})\n`;
      break;
    }
  }
  b += `  content((${f(cx)}, ${f(fy(cy))}), text(${font}size: 14pt)[#${goQuote(s.text)}])\n`;
  return b;
}
