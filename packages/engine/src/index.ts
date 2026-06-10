// @rombik/engine — публічний API рушія: код/AST → блок-схема за ДСТУ.
// Чистий TS, без залежностей від DOM чи фреймворку. Споживачі: веб, Node-скрипти.
export type { Diagram, Shape, Edge, Point, Kind } from './diagram.ts';
export { captionLine, labelAnchor } from './diagram.ts';
export type { Node, Func, Block, If, For, While } from './ir.ts';
export type { AstNode, AstFunc } from './astjson.ts';
export { fromJson } from './astjson.ts';
export { render as renderSvg } from './render/svg.ts';
