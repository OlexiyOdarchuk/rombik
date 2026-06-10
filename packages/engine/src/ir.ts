// IR — логічне (мова-агностик) дерево керування. Порт pkg/ir. Layout споживає це.
// Дискримінований union за полем kind. else/body/then завжди присутні як Block
// (порожній, якщо гілки нема) — на відміну від Go *Block (nil), щоб уникнути null-перевірок.

export interface Process { kind: 'process'; text: string; }
export interface IO { kind: 'io'; text: string; }
export interface Call { kind: 'call'; text: string; }
export interface Terminal { kind: 'terminal'; text: string; }
export interface If { kind: 'if'; cond: string; then: Block; else: Block; }
export interface For { kind: 'for'; spec: string; body: Block; else: Block; }
export interface While { kind: 'while'; cond: string; body: Block; else: Block; }
export interface DoWhile { kind: 'dowhile'; cond: string; body: Block; }
export interface InfLoop { kind: 'infloop'; body: Block; }
export interface Break { kind: 'break'; }
export interface Continue { kind: 'continue'; }
export interface Connector { kind: 'connector'; text: string; jump?: boolean; }
export interface Block { kind: 'block'; stmts: Node[]; }

export type Node =
  | Process | IO | Call | Terminal
  | If | For | While | DoWhile | InfLoop
  | Break | Continue | Connector | Block;

// Func — іменована програма (функція або тіло модуля) → окрема схема.
export interface Func { name: string; body: Block; main?: boolean; }
