// astjson — контракт «спрощений AST у JSON → IR». Порт pkg/parser/astjson.
// Будь-який фронтенд парсера (tree-sitter тощо) видає цей формат; тут єдина
// конвертація в IR. Вкладені block-вузли інлайняться.
import type { Block, Node, Func } from './ir.ts';

// AstNode — вузол спрощеного дерева. Поля залежать від kind (див. ToNode).
export interface AstNode {
  kind: string;
  text?: string;
  cond?: string;
  stmts?: AstNode[];
  then?: AstNode | null;
  else?: AstNode | null;
  body?: AstNode | null;
}
export interface AstFunc { name: string; block: AstNode; }

// fromJson: масив Func (рядок або вже розпарсений) → список IR-Func.
export function fromJson(data: string | AstFunc[]): Func[] {
  const fns: AstFunc[] = typeof data === 'string' ? JSON.parse(data) : data;
  return fns.map((f) => ({ name: f.name, body: toBlock(f.block) }));
}

// toBlock зводить вузол-блок у IR-Block (вкладені block інлайняться).
export function toBlock(n: AstNode | null | undefined): Block {
  const stmts: Node[] = [];
  for (const c of n?.stmts ?? []) {
    if (c.kind === 'block') stmts.push(...toBlock(c).stmts);
    else {
      const node = toNode(c);
      if (node) stmts.push(node);
    }
  }
  return { kind: 'block', stmts };
}

// toNode зводить один вузол у IR-Node (null — якщо тип невідомий).
function toNode(n: AstNode): Node | null {
  switch (n.kind) {
    case 'process': return { kind: 'process', text: n.text ?? '' };
    case 'io': return { kind: 'io', text: n.text ?? '' };
    case 'terminal': return { kind: 'terminal', text: n.text ?? '' };
    case 'call': return { kind: 'call', text: n.text ?? '' };
    case 'if': return { kind: 'if', cond: n.cond ?? '', then: toBlock(n.then), else: toBlock(n.else) };
    case 'for': return { kind: 'for', spec: n.cond ?? '', body: toBlock(n.body), else: toBlock(n.else) };
    case 'while': return { kind: 'while', cond: n.cond ?? '', body: toBlock(n.body), else: toBlock(n.else) };
    case 'dowhile': return { kind: 'dowhile', cond: n.cond ?? '', body: toBlock(n.body) };
    case 'infloop': return { kind: 'infloop', body: toBlock(n.body) };
    case 'break': return { kind: 'break' };
    case 'continue': return { kind: 'continue' };
  }
  return null;
}
