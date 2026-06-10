// Парсер: дерево tree-sitter (Python/C++) → AST-JSON (IR-контракт). Порт
// web/src/lib/parser.js у TS. Один фронтенд для обох мов; ядро (layout) від нього
// не залежить (споживає AST-JSON). Дзеркалить логіку parser.py (CLI колись).
import type { AstNode, AstFunc } from '../astjson.ts';

// TSNode — мінімальний інтерфейс вузла tree-sitter (структурно сумісний із
// web-tree-sitter Node), щоб не залежати від версії пакета.
export interface TSNode {
  id: number;
  type: string;
  text: string;
  isNamed: boolean;
  childCount: number;
  namedChildCount: number;
  child(index: number): TSNode | null;
  childForFieldName(name: string): TSNode | null;
  readonly namedChildren: TSNode[];
}
export interface TSTree { rootNode: TSNode; }
export type Lang = 'python' | 'cpp';

const MAXLEN = 64;
function oneline(t: string | null | undefined): string {
  if (!t) return '?';
  t = t.replace(/\s+/g, ' ').trim();
  return t.length <= MAXLEN ? t : t.substring(0, MAXLEN - 1) + '…';
}

function collectDefinedFunctions(node: TSNode | null, defined: Set<string>): void {
  if (!node) return;
  if (node.type === 'function_definition') {
    let nameNode = node.childForFieldName('name');
    if (!nameNode && node.childForFieldName('declarator')) {
      let decl = node.childForFieldName('declarator');
      while (decl && decl.type !== 'identifier' && decl.type !== 'function_declarator') {
        decl = decl.childForFieldName('declarator');
      }
      if (decl && decl.type === 'function_declarator') {
        const n = decl.childForFieldName('declarator');
        if (n && n.type === 'identifier') nameNode = n;
      }
    }
    if (nameNode) defined.add(nameNode.text);
  }
  for (let i = 0; i < node.childCount; i++) collectDefinedFunctions(node.child(i), defined);
}

function hasInputNode(node: TSNode | null): boolean {
  if (!node) return false;
  if (node.type === 'call' || node.type === 'call_expression') {
    const func = node.childForFieldName('function');
    if (func && (func.text === 'input' || func.text === 'cin' || func.text === 'std::cin')) return true;
  }
  if (node.text === 'cin' || node.text === 'std::cin') return true;
  for (let i = 0; i < node.namedChildCount; i++) {
    if (hasInputNode(node.namedChildren[i])) return true;
  }
  return false;
}

function getCallName(node: TSNode | null): string | null {
  if (!node) return null;
  if (node.type === 'call' || node.type === 'call_expression') {
    const func = node.childForFieldName('function');
    if (func) return func.text;
  }
  return null;
}

// endof — верхня межа range як «stop - 1», зі згортанням «X + 1» → «X» і констант.
function endof(node: TSNode | null): string {
  if (!node) return '?';
  if (node.type === 'integer') {
    const n = parseInt(node.text, 10);
    if (!Number.isNaN(n)) return String(n - 1);
  }
  if (node.type === 'binary_operator') {
    const op = node.childForFieldName('operator');
    const right = node.childForFieldName('right');
    const left = node.childForFieldName('left');
    if (op && op.text === '+' && right && right.text === '1' && left) return left.text;
  }
  return node.text + ' - 1';
}

// unwrapElse — розгортає вузол else_clause у його тіло (для for/while-else; if
// розгортає інлайн). Без цього else_clause летить у block() як невідомий вузол і
// стає текстовим дампом замість розібраної гілки.
function unwrapElse(alt: TSNode | null): TSNode | null {
  if (alt && alt.type === 'else_clause') {
    return alt.childForFieldName('body') || alt.childForFieldName('consequence') ||
      (alt.namedChildCount > 0 ? alt.namedChildren[0] : null);
  }
  return alt;
}

function isTrueNode(n: TSNode | null): boolean {
  return !!n && (n.type === 'true' || (n.type === 'integer' && n.text === '1'));
}
function breakIfNode(st: TSNode | null): TSNode | null {
  if (!st || st.type !== 'if_statement' || st.childForFieldName('alternative')) return null;
  const body = st.childForFieldName('consequence');
  if (!body) return null;
  const kids = body.namedChildren.filter((c) => c.type !== 'comment');
  if (kids.length === 1 && kids[0].type === 'break_statement') return st.childForFieldName('condition');
  return null;
}

function formatArg(argNode: TSNode): string {
  if (argNode.type === 'string') {
    const text = argNode.text;
    if (text.startsWith('f"') || text.startsWith("f'")) return `«${text}»`;
    if (text.length >= 2 && (text[0] === '"' || text[0] === "'")) return `«${text.substring(1, text.length - 1)}»`;
  }
  return argNode.text;
}

// parseTree — головна функція: tree-sitter Tree + мова → масив схем (AstFunc).
export function parseTree(tree: TSTree, lang: Lang): AstFunc[] {
  const defined = new Set<string>();
  collectDefinedFunctions(tree.rootNode, defined);

  const isCpp = lang === 'cpp';
  const isPy = lang === 'python';

  function stmt(s: TSNode | null): AstNode | null {
    if (!s) return null;

    if (s.type === 'if_statement' || s.type === 'elif_clause') {
      const cond = s.childForFieldName('condition');
      const conseq = s.childForFieldName('consequence') || s.childForFieldName('body');
      let alt = s.childForFieldName('alternative');
      if (alt && alt.type === 'else_clause') {
        let altBody = alt.childForFieldName('body') || alt.childForFieldName('consequence');
        if (!altBody && alt.namedChildCount > 0) altBody = alt.namedChildren[0];
        alt = altBody;
      }
      let condText = cond ? cond.text : '?';
      if (isCpp && condText.startsWith('(') && condText.endsWith(')')) condText = condText.substring(1, condText.length - 1);
      return { kind: 'if', cond: oneline(condText), then: block(conseq), else: block(alt) };
    }

    if (s.type === 'while_statement') {
      const cond = s.childForFieldName('condition');
      const body = s.childForFieldName('body');
      if (isPy && isTrueNode(cond) && body) {
        const stmts = body.namedChildren.filter((c) => c.type !== 'comment');
        const last = stmts[stmts.length - 1];
        const brkCond = breakIfNode(last ?? null);
        if (brkCond) {
          return { kind: 'dowhile', cond: oneline(brkCond.text),
            body: { kind: 'block', stmts: stmts.slice(0, -1).map(stmt).filter(Boolean) as AstNode[] } };
        }
        return { kind: 'infloop', body: block(body) };
      }
      let condText = cond ? cond.text : '?';
      if (isCpp && condText.startsWith('(') && condText.endsWith(')')) condText = condText.substring(1, condText.length - 1);
      return { kind: 'while', cond: oneline(condText), body: block(body), else: block(unwrapElse(s.childForFieldName('alternative'))) };
    }

    if (s.type === 'for_statement') {
      if (isPy) {
        const left = s.childForFieldName('left');
        const right = s.childForFieldName('right');
        let condText = `${left ? left.text : '?'} ∈ ${right ? right.text : '?'}`;
        if (right && (right.type === 'call' || right.type === 'call_expression')) {
          const func = getCallName(right);
          if (func === 'range') {
            const args = right.childForFieldName('arguments');
            if (args && args.namedChildCount > 0) {
              const argNodes = args.namedChildren;
              const leftText = left ? left.text : '?';
              if (argNodes.length === 1) condText = `${leftText} = 0, ${endof(argNodes[0])}, 1`;
              else if (argNodes.length === 2) condText = `${leftText} = ${argNodes[0].text}, ${endof(argNodes[1])}, 1`;
              else if (argNodes.length >= 3) condText = `${leftText} = ${argNodes[0].text}, ${endof(argNodes[1])}, ${argNodes[2].text}`;
            }
          }
        }
        return { kind: 'for', cond: oneline(condText), body: block(s.childForFieldName('body')), else: block(unwrapElse(s.childForFieldName('alternative'))) };
      }
      if (isCpp) {
        const init = s.childForFieldName('initializer');
        const cond = s.childForFieldName('condition');
        const upd = s.childForFieldName('update');
        const body = s.childForFieldName('body');
        let spec = '';
        if (init) {
          if (init.type === 'declaration') {
            const decls: string[] = [];
            for (let i = 0; i < init.namedChildCount; i++) {
              if (init.namedChildren[i].type === 'init_declarator') decls.push(init.namedChildren[i].text);
            }
            spec += decls.length > 0 ? decls.join(', ') : init.text.replace(';', '');
          } else {
            spec += init.text.replace(';', '');
          }
        }
        if (cond) spec += (spec ? ', ' : '') + cond.text.replace(';', '');
        if (upd) spec += (spec ? ', ' : '') + upd.text;
        return { kind: 'for', cond: oneline(spec), body: block(body), else: block(null) };
      }
    }

    if (s.type === 'for_range_loop') { // C++ range-based for: for (int x : arr)
      const decl = s.childForFieldName('declarator');
      const right = s.childForFieldName('right');
      const body = s.childForFieldName('body');
      const cond = `${decl ? decl.text : '?'} ∈ ${right ? right.text : '?'}`;
      return { kind: 'for', cond: oneline(cond), body: block(body), else: block(null) };
    }

    if (s.type === 'do_statement') {
      const cond = s.childForFieldName('condition');
      const body = s.childForFieldName('body');
      let condText = cond ? cond.text : '?';
      if (isCpp && condText.startsWith('(') && condText.endsWith(')')) condText = condText.substring(1, condText.length - 1);
      return { kind: 'dowhile', cond: oneline(condText), body: block(body) };
    }

    if (s.type === 'switch_statement') {
      const cond = s.childForFieldName('condition');
      const body = s.childForFieldName('body');
      let subj = cond ? cond.text : '?';
      if (isCpp && subj.startsWith('(') && subj.endsWith(')')) subj = subj.substring(1, subj.length - 1);
      // Зібрати case-и; порожні мітки (fallthrough «case 1: case 2: case 3:») зливаються
      // з наступним непорожнім — їхні значення йдуть в одну умову через АБО.
      const cases: { vals: string[]; node: TSNode; valNode: TSNode | null }[] = [];
      let pendingVals: string[] = [];
      if (body && (body.type === 'compound_statement' || body.type === 'block')) {
        for (let i = 0; i < body.namedChildCount; i++) {
          const c = body.namedChildren[i];
          if (c.type !== 'case_statement') continue;
          const valNode = c.childForFieldName('value');
          const hasBody = c.namedChildren.some((ch) => !(valNode && ch.id === valNode.id));
          if (!hasBody) { // лише мітка, без тіла — накопичити для наступного
            if (valNode) pendingVals.push(valNode.text);
            continue;
          }
          const vals = [...pendingVals];
          if (valNode) vals.push(valNode.text);
          pendingVals = [];
          cases.push({ vals, node: c, valNode });
        }
      }
      const blockify = (node: AstNode): AstNode => (node.kind === 'block' ? node : { kind: 'block', stmts: [node] });
      const buildCascade = (i: number): AstNode => {
        if (i >= cases.length) return { kind: 'block', stmts: [] };
        const c = cases[i];
        const condText = c.vals.length ? c.vals.map((v) => `${subj} == ${v}`).join(' || ') : '';
        const caseStmts: AstNode[] = [];
        for (let j = 0; j < c.node.namedChildCount; j++) {
          const child = c.node.namedChildren[j];
          if (c.valNode && child.id === c.valNode.id) continue; // пропустити саме значення case
          const mapped = stmt(child);
          if (mapped) caseStmts.push(mapped);
        }
        const caseBlock: AstNode = { kind: 'block', stmts: caseStmts };
        if (condText === '') return caseBlock; // default
        return { kind: 'if', cond: oneline(condText.replace(';', '')), then: caseBlock, else: blockify(buildCascade(i + 1)) };
      };
      return buildCascade(0);
    }

    if (s.type === 'match_statement') {
      const subjNode = s.childForFieldName('subject');
      const subj = subjNode ? subjNode.text : '?';
      const body = s.childForFieldName('body');
      const cases: { pat: string | null; body: TSNode | null }[] = [];
      if (body && body.type === 'block') {
        for (let i = 0; i < body.namedChildCount; i++) {
          const c = body.namedChildren[i];
          if (c.type === 'case_clause') {
            const pat = c.namedChildren.find((n) => n.type === 'case_pattern');
            const conseq = c.childForFieldName('consequence') || c.childForFieldName('body') || c.namedChildren.find((n) => n.type === 'block') || null;
            cases.push({ pat: pat ? pat.text : null, body: conseq });
          }
        }
      }
      const blockify = (node: AstNode): AstNode => (node.kind === 'block' ? node : { kind: 'block', stmts: [node] });
      const buildCascade = (i: number): AstNode => {
        if (i >= cases.length) return { kind: 'block', stmts: [] };
        const c = cases[i];
        let condText = c.pat ? `${subj} == ${c.pat}` : '';
        if (c.pat === '_' || c.pat === 'case _') condText = '';
        if (condText === '') return block(c.body);
        return { kind: 'if', cond: oneline(condText), then: block(c.body), else: blockify(buildCascade(i + 1)) };
      };
      return buildCascade(0);
    }

    if (s.type === 'try_statement') {
      const bodyNode = s.childForFieldName('body');
      const outStmts = block(bodyNode).stmts ?? [];
      const handlers: TSNode[] = [];
      let finBlock: TSNode | null = null;
      for (let i = 0; i < s.namedChildCount; i++) {
        const c = s.namedChildren[i];
        if (c.type === 'except_clause' || c.type === 'catch_clause') handlers.push(c);
        if (c.type === 'finally_clause') finBlock = c;
      }
      const buildHandlers = (i: number): AstNode => {
        if (i >= handlers.length) return { kind: 'block', stmts: [] };
        const h = handlers[i];
        let exType = '';
        if (h.type === 'except_clause') {
          const val = h.namedChildren.find((n) => n.type !== 'block');
          if (val) exType = ' ' + val.text;
        } else if (h.type === 'catch_clause') {
          const p = h.childForFieldName('parameters');
          if (p) {
            let text = p.text;
            if (text.startsWith('(') && text.endsWith(')')) text = text.substring(1, text.length - 1);
            exType = ' ' + text;
          }
        }
        let hBody = h.childForFieldName('body');
        if (!hBody) hBody = h.namedChildren[h.namedChildCount - 1];
        return { kind: 'block', stmts: [{ kind: 'if', cond: oneline(`Виняток${exType}?`), then: block(hBody), else: buildHandlers(i + 1) }] };
      };
      if (handlers.length > 0) outStmts.push(...(buildHandlers(0).stmts ?? []));
      if (finBlock) {
        const finBody = finBlock.namedChildren[finBlock.namedChildCount - 1];
        outStmts.push(...(block(finBody).stmts ?? []));
      }
      return { kind: 'block', stmts: outStmts };
    }

    if (s.type === 'with_statement') {
      const bodyNode = s.childForFieldName('body');
      const items: string[] = [];
      for (let i = 0; i < s.namedChildCount; i++) {
        const c = s.namedChildren[i];
        if (c.type === 'with_item') items.push(c.text);
        else if (c !== bodyNode && c.type !== 'with_item' && c.type !== 'block' && c.type !== 'compound_statement') items.push(c.text);
      }
      const stmts: AstNode[] = [{ kind: 'process', text: oneline('відкрити: ' + items.join(', ')) }];
      if (bodyNode) stmts.push(...(block(bodyNode).stmts ?? []));
      return { kind: 'block', stmts };
    }

    if (s.type === 'break_statement') return { kind: 'break' };
    if (s.type === 'continue_statement') return { kind: 'continue' };

    if (s.type === 'return_statement') {
      let val = '';
      const vals = s.namedChildren.filter((c) => c.type !== 'comment');
      if (vals.length > 0) val = vals.map((c) => c.text).join(' ');
      return { kind: 'terminal', text: val ? oneline('Повернути ' + val.replace(';', '')) : 'Повернути' };
    }

    if (s.type === 'raise_statement' || s.type === 'throw_statement') {
      const cause = s.namedChildren.length > 0 ? s.namedChildren[0].text : '';
      return { kind: 'terminal', text: oneline('Помилка' + (cause ? ': ' + cause : '')) };
    }

    if (s.type === 'expression_statement') {
      const expr = s.namedChildren[0];
      if (!expr) return { kind: 'process', text: oneline(s.text.replace(';', '')) };
      if (expr.type === 'string' || expr.type === 'concatenated_string') return null;

      if (isCpp && (expr.type === 'binary_expression' || expr.type === 'shift_expression')) {
        const text = expr.text.trim();
        if (text.startsWith('cout') || text.startsWith('std::cout')) {
          let ioText = text.replace(/^(?:std::)?cout\s*<<\s*/, '').replace(/<<\s*(?:std::)?endl$/, '').replace(/<<\s*(?:std::)?endl\s*/, '').replace(/<<\s*/g, ' ');
          ioText = ioText.replace(/"([^"]*)"/g, '«$1»');
          return { kind: 'io', text: oneline('Вивід ' + ioText.replace(';', '')) };
        }
        if (text.startsWith('cin') || text.startsWith('std::cin')) {
          const ioText = text.replace(/^(?:std::)?cin\s*>>\s*/, '').replace(/>>\s*/g, ', ');
          return { kind: 'io', text: oneline('Ввід ' + ioText.replace(';', '')) };
        }
      }

      if (expr.type === 'call' || expr.type === 'call_expression') {
        const funcName = getCallName(expr);
        if (funcName === 'print' || funcName === 'cout' || funcName === 'std::cout') {
          const args = expr.childForFieldName('arguments');
          if (args && args.namedChildCount > 0) return { kind: 'io', text: oneline('Вивід ' + args.namedChildren.map(formatArg).join(' ')) };
          return { kind: 'io', text: 'Вивід порожнього рядка' };
        }
        if (funcName === 'input' || funcName === 'cin' || funcName === 'std::cin') {
          const args = expr.childForFieldName('arguments');
          if (args && args.namedChildCount > 0) return { kind: 'io', text: oneline('Ввід ' + args.namedChildren.map(formatArg).join(' ')) };
          return { kind: 'io', text: 'Ввід' };
        }
        if (funcName === 'exit' || funcName === 'quit') return { kind: 'terminal', text: 'Вихід' };
        if (funcName && defined.has(funcName)) return { kind: 'call', text: oneline(expr.text.replace(';', '')) };
      }

      if (expr.type === 'assignment' || expr.type === 'assignment_expression') {
        if (hasInputNode(expr.childForFieldName('right'))) {
          const left = expr.childForFieldName('left');
          if (left) return { kind: 'io', text: oneline('Ввід ' + left.text.replace(';', '')) };
        }
        const rhs = expr.childForFieldName('right');
        if (rhs && (rhs.type === 'call' || rhs.type === 'call_expression')) {
          const funcName = getCallName(rhs);
          if (funcName && defined.has(funcName)) return { kind: 'call', text: oneline(expr.text.replace(';', '')) };
        }
      }

      return { kind: 'process', text: oneline(expr.text.replace(';', '')) };
    }

    if (s.type === 'block' || s.type === 'compound_statement') return block(s);

    if (s.type === 'declaration') {
      if (hasInputNode(s)) {
        const decl = s.childForFieldName('declarator');
        const inner = decl?.childForFieldName('declarator');
        if (decl && inner) return { kind: 'io', text: oneline('Ввід ' + inner.text.replace(';', '')) };
        return { kind: 'io', text: oneline('Ввід ' + s.text.replace(';', '')) };
      }
      if (isCpp) {
        const decls: string[] = [];
        for (let i = 0; i < s.namedChildCount; i++) {
          if (s.namedChildren[i].type === 'init_declarator') decls.push(s.namedChildren[i].text);
        }
        if (decls.length > 0) return { kind: 'process', text: oneline(decls.join(', ').replace(/;/g, '')) };
        return null;
      }
      return { kind: 'process', text: oneline(s.text.replace(';', '')) };
    }

    if (s.isNamed) {
      if (s.type === 'function_definition' || s.type === 'class_definition' ||
          s.type === 'class_specifier' || s.type === 'struct_specifier' ||
          s.type === 'using_declaration' || s.type === 'namespace_definition' ||
          s.type === 'import_statement' || s.type === 'import_from_statement' ||
          s.type === 'preproc_include' || s.type === 'preproc_def' ||
          s.type === 'comment' || s.type === 'string') return null;
      return { kind: 'process', text: oneline(s.text.replace(';', '')) };
    }
    return null;
  }

  function block(node: TSNode | null): AstNode {
    if (!node) return { kind: 'block', stmts: [] };
    const stmts: AstNode[] = [];
    const children = node.type === 'block' || node.type === 'compound_statement' ? node.namedChildren : [node];
    for (const c of children) {
      const mapped = stmt(c);
      if (mapped) stmts.push(mapped);
    }
    return { kind: 'block', stmts };
  }

  const funcs: AstFunc[] = [];
  const mainStmts: AstNode[] = [];

  function collect(node: TSNode | null, prefix = ''): void {
    if (!node) return;
    // C++ клас/структура: кожен метод-тіло — окрема схема «Клас::метод». Поля й
    // специфікатори доступу ігноруємо (інакше протекли б у «main»).
    if (node.type === 'class_specifier' || node.type === 'struct_specifier') {
      const cn = node.childForFieldName('name');
      const list = node.childForFieldName('body');
      if (list) for (const k of list.namedChildren) if (k.type === 'function_definition') collect(k, (cn ? cn.text + '::' : '') + prefix);
      return;
    }
    if (node.type === 'function_definition') {
      let nameNode = node.childForFieldName('name');
      let paramsText = '';
      const bodyNode = node.childForFieldName('body');

      if (isCpp) {
        let decl = node.childForFieldName('declarator');
        while (decl && decl.type !== 'identifier' && decl.type !== 'function_declarator') decl = decl.childForFieldName('declarator');
        if (decl && decl.type === 'function_declarator') {
          const n = decl.childForFieldName('declarator');
          if (n && (n.type === 'identifier' || n.type === 'field_identifier')) nameNode = n; // field_identifier — метод у класі
          const p = decl.childForFieldName('parameters');
          if (p) {
            const paramsList: string[] = [];
            for (let i = 0; i < p.namedChildCount; i++) {
              const paramDecl = p.namedChildren[i];
              if (paramDecl.type === 'parameter_declaration' || paramDecl.type === 'optional_parameter_declaration') {
                const d = paramDecl.childForFieldName('declarator');
                paramsList.push(d ? d.text : paramDecl.text);
              } else {
                paramsList.push(paramDecl.text);
              }
            }
            paramsText = paramsList.join(', ');
          }
        }
      } else if (isPy) {
        const p = node.childForFieldName('parameters');
        if (p) {
          const names: string[] = [];
          for (const pc of p.namedChildren) {
            if (pc.type === 'comment') continue;
            if (pc.type === 'identifier') { names.push(pc.text); continue; }
            const nm = pc.childForFieldName('name');
            if (nm) { names.push(nm.text); continue; }
            const id = pc.namedChildren.find((c) => c.type === 'identifier');
            names.push(id ? id.text : pc.text);
          }
          paramsText = names.join(', ');
        }
      }

      const fnName = prefix + (nameNode ? nameNode.text : 'unknown');
      const b = block(bodyNode);
      if (paramsText && paramsText.trim() !== '') (b.stmts ??= []).unshift({ kind: 'io', text: 'Ввід ' + oneline(paramsText) });
      funcs.push({ name: fnName, block: b });

      if (bodyNode && bodyNode.type === 'block') {
        for (const k of bodyNode.namedChildren) if (k.type === 'function_definition') collect(k, prefix);
      }
    } else {
      const mapped = stmt(node);
      if (mapped && mapped.kind !== 'block') mainStmts.push(mapped);
      else if (mapped && mapped.kind === 'block') mainStmts.push(...(mapped.stmts ?? []));
    }
  }

  for (const c of tree.rootNode.namedChildren) collect(c);

  const out: AstFunc[] = [...funcs];
  if (mainStmts.length > 0) out.push({ name: defined.has('main') ? 'програма' : 'main', block: { kind: 'block', stmts: mainStmts } });
  if (out.length === 0) out.push({ name: 'main', block: { kind: 'block', stmts: [] } });
  return out;
}
