// Модель геометрії — порт pkg/diagram. Контракт між layout (виробляє) і render
// (малює). Чиста структура даних. Kind серіалізується рядком (як у Go MarshalJSON).

export type Kind =
  | 'terminator' // початок/кінець — овал
  | 'process'    // дія — прямокутник
  | 'decision'   // умова — ромб
  | 'io'         // ввід/вивід — паралелограм
  | 'loop'       // початок циклу (for) — шестикутник
  | 'subprogram' // виклик підпрограми — прямокутник з боковими рисками
  | 'connector'; // з'єднувач «А» — коло

export interface Point { x: number; y: number; }

export interface Shape {
  kind: Kind;
  x: number; y: number; w: number; h: number;
  text: string;
}

export interface Edge {
  points: Point[];
  label?: string;
  arrowless?: boolean;
}

export interface Diagram {
  shapes: Shape[];
  edges: Edge[];
  w: number;
  h: number;
  caption?: string;
  figNum?: number;
  capWord?: string;
  capFormat?: string;
}

export const CAPTION_WORD = 'Рисунок';
export const CAP_FORMAT_DEFAULT = '{word} {num} — {text}';

// labelAnchor — позиція й вирівнювання підпису ребра (Так/Ні) за першим сегментом.
export function labelAnchor(p0: Point, p1: Point): { x: number; y: number; align: string } {
  if (p1.y === p0.y && p1.x !== p0.x) { // горизонтальний сегмент
    let off = (p1.x - p0.x) / 2;
    if (off > 34) off = 34;
    else if (off < -34) off = -34;
    return { x: p0.x + off, y: p0.y - 9, align: 'middle' };
  }
  if (p1.x === p0.x) { // вертикальний сегмент
    return { x: p0.x + 10, y: (p0.y + p1.y) / 2, align: 'start' };
  }
  return { x: p0.x + 8, y: p0.y - 8, align: 'start' };
}

export function capSupplement(d: Diagram): string { return d.capWord || CAPTION_WORD; }
function capFmt(d: Diagram): string { return d.capFormat || CAP_FORMAT_DEFAULT; }

// capSeparator — роздільник між номером і текстом (для нативного підпису Typst-figure).
export function capSeparator(d: Diagram): string {
  const fmt = capFmt(d);
  const i = fmt.indexOf('{num}');
  const j = fmt.indexOf('{text}');
  if (i < 0 || j < 0 || j < i) return ' — ';
  return fmt.slice(i + '{num}'.length, j);
}

// capHasWord — чи містить шаблон слово-supplement (для Typst-figure).
export function capHasWord(d: Diagram): boolean { return capFmt(d).includes('{word}'); }

// captionLine — повний рядок підпису («Рисунок N — текст») за шаблоном.
export function captionLine(d: Diagram): string {
  if (!d.caption) return '';
  if (!d.figNum || d.figNum <= 0) return d.caption;
  return capFmt(d)
    .replaceAll('{word}', capSupplement(d))
    .replaceAll('{num}', String(d.figNum))
    .replaceAll('{text}', d.caption)
    .trim();
}
