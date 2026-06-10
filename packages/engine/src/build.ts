// Оркестрація: AST-JSON → схеми. Порт rombik.build/FromAST. Кожна функція файлу
// стає окремою схемою з підписом «ім'я» і номером «Рисунок N».
import type { Diagram } from './diagram.ts';
import type { Options } from './layout/options.ts';
import { resolveOptions } from './layout/options.ts';
import { fromJson, type AstFunc } from './astjson.ts';
import { layoutProgram } from './layout/place.ts';
import { parseTree, type TSTree, type Lang } from './parser/treesitter.ts';

export interface Result { name: string; diagram: Diagram; }

// fromTree: дерево tree-sitter (Python/C++) → схеми. Повний шлях код→схема (парсер
// у браузері/Node дає Tree, далі все в TS).
export function fromTree(tree: TSTree, lang: Lang, opts: Options = {}): Result[] {
  return fromAst(parseTree(tree, lang), opts);
}

// fromAst: розібраний AST-JSON (рядок або масив) + опції → список схем.
export function fromAst(astJSON: string | AstFunc[], opts: Options = {}): Result[] {
  const o = resolveOptions(opts);
  const funcs = fromJson(astJSON);
  return funcs.map((f, i) => {
    const d = layoutProgram(f.body, o);
    d.caption = f.name;     // підпис за замовч. — ім'я функції
    d.figNum = i + 1;       // «Рисунок N» за порядком
    d.capWord = o.capWord;  // слово підпису
    return { name: f.name, diagram: d };
  });
}
