// @rombik/engine — публічний API рушія: код/AST → блок-схема за ДСТУ.
export type { Diagram, Shape, Edge, Point, Kind } from './diagram.ts';
export { captionLine, labelAnchor } from './diagram.ts';
export type { Node, Func, Block, If, For, While } from './ir.ts';
export type { AstNode, AstFunc } from './astjson.ts';
export { fromJson } from './astjson.ts';
export type { Options } from './layout/options.ts';
export { layoutProgram } from './layout/place.ts';
export { fromAst, type Result } from './build.ts';
export { render as renderSvg } from './render/svg.ts';
export { render as renderTypst, renderAll as renderTypstAll, fragment as renderTypstFragment } from './render/typst.ts';
export { render as renderExcalidraw, renderAll as renderExcalidrawAll } from './render/excalidraw.ts';
