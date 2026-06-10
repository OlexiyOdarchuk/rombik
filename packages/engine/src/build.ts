// Оркестрація: AST-JSON → схеми. Порт rombik.build/FromAST + split.go. Кожна
// функція файлу стає окремою схемою з підписом «ім'я» і номером «Рисунок N».
import type { Diagram } from './diagram.ts';
import type { Node, Block, Func } from './ir.ts';
import type { Options, ResolvedOptions } from './layout/options.ts';
import { resolveOptions } from './layout/options.ts';
import { fromJson, type AstFunc } from './astjson.ts';
import { layoutProgram } from './layout/place.ts';
import { parseTree, type TSTree, type Lang } from './parser/treesitter.ts';

export interface Result { name: string; diagram: Diagram; }

// buildResults — спільне: список IR-Func + опції → схеми (підпис=ім'я, Рисунок N).
function buildResults(funcs: Func[], o: ResolvedOptions): Result[] {
  return funcs.map((f, i) => {
    // ДСТУ: Початок/Кінець лише для main; підпрограми → Вхід/Вихід.
    const fo = o.mainOnlyTerminators && !f.main ? { ...o, startText: o.entryText, endText: o.exitText } : o;
    const d = layoutProgram(f.body, fo);
    d.caption = f.name;
    d.figNum = i + 1;
    d.capWord = o.capWord;
    return { name: f.name, diagram: d };
  });
}

// fromTree: дерево tree-sitter (Python/C++) → схеми. Повний шлях код→схема в TS.
export function fromTree(tree: TSTree, lang: Lang, opts: Options = {}): Result[] {
  return buildResults(fromJson(parseTree(tree, lang, { forFormat: opts.forFormat, returnWord: opts.returnWord, forEachWord: opts.forEachWord })), resolveOptions(opts));
}

// fromAst: розібраний AST-JSON (рядок або масив) + опції → список схем.
export function fromAst(astJSON: string | AstFunc[], opts: Options = {}): Result[] {
  return buildResults(fromJson(astJSON), resolveOptions(opts));
}

// --- поділ схеми на зв'язані частини (порт split.go) ---

const ABC = [...'АБВГДЕЖЗИКЛМНОПРСТУФ'];
function letter(i: number): string {
  return i >= 0 && i < ABC.length ? ABC[i] : `А${i - ABC.length + 1}`;
}

// splitByHeight ріже функцію на частини ≤ maxH (одиниць), на стиках — пара
// конекторів (вихід попередньої → вхід наступної). Якщо вміщається — одна схема.
export function splitByHeight(f: Func, opts: Options, maxH: number): Result[] {
  const o = resolveOptions(opts);
  const stmts = f.body.stmts;
  if (stmts.length <= 1) return buildResults([f], o);
  // Жадібно групуємо, поки висота частини не перевищить поріг.
  const groups: Node[][] = [];
  let cur: Node[] = [];
  for (const s of stmts) {
    const cand = [...cur, s];
    const h = layoutProgram({ kind: 'block', stmts: cand }, o).h;
    if (h > maxH && cur.length > 0) { groups.push(cur); cur = [s]; }
    else cur = cand;
  }
  if (cur.length > 0) groups.push(cur);
  if (groups.length <= 1) return buildResults([f], o);

  return groups.map((g, i) => {
    const body: Node[] = [];
    if (i > 0) body.push({ kind: 'connector', text: letter(i - 1) });
    body.push(...g);
    if (i < groups.length - 1) body.push({ kind: 'connector', text: letter(i) });
    const oo: ResolvedOptions = { ...o, noStart: i > 0, noEnd: i < groups.length - 1 };
    const d = layoutProgram({ kind: 'block', stmts: body }, oo);
    d.caption = `${f.name} (частина ${i + 1} з ${groups.length})`;
    d.figNum = i + 1;
    d.capWord = o.capWord;
    return { name: `${f.name}_ч${i + 1}`, diagram: d };
  });
}

// splitFromAst: AST-JSON + ім'я функції → частини (для кнопки розбивки).
export function splitFromAst(astJSON: string | AstFunc[], opts: Options, name: string, maxH: number): Result[] {
  const f = fromJson(astJSON).find((x) => x.name === name);
  if (!f) throw new Error(`функцію "${name}" не знайдено`);
  return splitByHeight(f, opts, maxH);
}
