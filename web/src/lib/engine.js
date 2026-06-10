// Рушій у браузері, без сервера й без Go-WASM:
//   web-tree-sitter парсить код → Tree
//   rombik-engine (чистий TS) бере Tree → схеми → SVG / Typst / Excalidraw
// PNG/PDF поки через лінивий растер-WASM (заміниться браузерним canvas).
import { base } from '$app/paths';
import {
	parseTree, fromAst, splitFromAst,
	renderSvg, renderSvgAll as engRenderSvgAll,
	renderTypst as engRenderTypst, renderTypstAll as engRenderTypstAll, renderTypstFragment, renderTypstFragmentAll,
	renderExcalidraw as engRenderExcalidraw, renderExcalidrawAll as engRenderExcalidrawAll,
} from 'rombik-engine';

let initPromise = null;
let parser = null;
let langs = {}; // 'python' | 'cpp' -> Language

// init: лише tree-sitter (рушій тепер чистий TS, Go-wasm не потрібен).
async function init(onProgress) {
	onProgress?.('Завантаження Tree-sitter…');
	const modulePath = `${base}/tree-sitter.js`;
	const { Parser, Language } = await import(/* @vite-ignore */ modulePath);
	await Parser.init({ locateFile: () => `${base}/tree-sitter.wasm` });
	parser = new Parser();
	langs.python = await Language.load(`${base}/tree-sitter-python.wasm`);
	langs.cpp = await Language.load(`${base}/tree-sitter-cpp.wasm`);
	langs.pascal = await Language.load(`${base}/tree-sitter-pascal.wasm`);
}

/** Готує середовище (ідемпотентно). Можна викликати заздалегідь для прогріву. */
export function warmup(onProgress) {
	if (!initPromise) initPromise = init(onProgress);
	return initPromise;
}

let lastAst = null; // AST останньої генерації (для розбивки)

/** generate(code, options) -> { functions:[{name, svg, diagram}], warning? } | { error }. */
export async function generate(code, options = {}, onProgress) {
	await warmup(onProgress);
	onProgress?.('Будую схему…');
	try {
		// C — підмножина C++: та сама граматика й гілка парсера (lang='cpp').
		const lang = options.lang === 'pascal' ? 'pascal'
			: options.lang === 'cpp' || options.lang === 'c' ? 'cpp' : 'python';
		parser.setLanguage(langs[lang]);
		const tree = parser.parse(code);
		lastAst = parseTree(tree, lang);
		const functions = fromAst(lastAst, options).map((r) => ({
			name: r.name, diagram: r.diagram, svg: renderSvg(r.diagram),
		}));
		const res = { functions };
		if (tree.rootNode.hasError) {
			res.warning = 'У коді є синтаксичні помилки. Деякі блоки можуть бути згенеровані неправильно.';
		}
		return res;
	} catch (e) {
		return { error: 'рушій: ' + (e?.message ?? e) };
	}
}

/** Ріже схему функції на зв'язані частини (кнопка «Розбити на частини»). */
export function splitSchema(name, maxH, options = {}) {
	if (!lastAst) return { error: 'спершу побудуй схему' };
	try {
		const parts = splitFromAst(lastAst, options, name, maxH).map((r) => ({
			name: r.name, caption: r.diagram.caption, figNum: r.diagram.figNum,
			svg: renderSvg(r.diagram), typst: engRenderTypst(r.diagram), diagram: r.diagram,
		}));
		return { parts };
	} catch (e) {
		return { error: 'розбивка: ' + (e?.message ?? e) };
	}
}

/** Дешевий ре-рендер однієї схеми після зміни підпису. cap = { caption, figNum, capWord, capFormat }. */
export function renderCaption(diagram, cap) {
	try {
		const d = { ...diagram, ...cap };
		return { svg: renderSvg(d), typst: engRenderTypst(d) };
	} catch (e) {
		return { error: 'ре-рендер: ' + (e?.message ?? e) };
	}
}

/** Typst однієї схеми. fragment=true → лише cetz.canvas (без преамбули). */
export function renderTypst(diagram, fragment = false) {
	return fragment ? renderTypstFragment(diagram) : engRenderTypst(diagram);
}

/** Один Typst з УСІХ схем. fragment=true → лише canvas-блоки. */
export function renderTypstAll(diagrams, fragment = false) {
	return fragment ? renderTypstFragmentAll(diagrams) : engRenderTypstAll(diagrams);
}

/** Один SVG з УСІХ схем (вертикально). */
export function renderSvgAll(diagrams) {
	return engRenderSvgAll(diagrams);
}

/** Схема у форматі .excalidraw (для excalidraw.com). */
export function renderExcalidraw(diagram) {
	return engRenderExcalidraw(diagram);
}

/** Усі схеми в одному .excalidraw. */
export function renderExcalidrawAll(diagrams) {
	return engRenderExcalidrawAll(diagrams);
}

// --- PNG/PDF у браузері (SVG→canvas), без Go-WASM ---
// SVG рендерить рушій (TS), а растеризацію робить сам браузер: <img> зі SVG →
// <canvas> → PNG; PDF — той самий растр посторінково через jsPDF (лінивий import).

function svgDims(svg) {
	const m = svg.match(/width="([\d.]+)" height="([\d.]+)"/);
	return m ? { w: parseFloat(m[1]), h: parseFloat(m[2]) } : { w: 800, h: 600 };
}

// svgToCanvas: SVG-рядок → <canvas> у масштабі scale (білий фон, як у Go-растрі).
async function svgToCanvas(svg, scale) {
	const { w, h } = svgDims(svg);
	const url = URL.createObjectURL(new Blob([svg], { type: 'image/svg+xml;charset=utf-8' }));
	try {
		const img = new Image();
		await new Promise((res, rej) => {
			img.onload = () => res();
			img.onerror = () => rej(new Error('не вдалося растеризувати SVG'));
			img.src = url;
		});
		const canvas = document.createElement('canvas');
		canvas.width = Math.max(1, Math.round(w * scale));
		canvas.height = Math.max(1, Math.round(h * scale));
		const ctx = canvas.getContext('2d');
		ctx.fillStyle = '#ffffff';
		ctx.fillRect(0, 0, canvas.width, canvas.height);
		ctx.setTransform(scale, 0, 0, scale, 0, 0);
		ctx.drawImage(img, 0, 0, w, h);
		return { canvas, w, h };
	} finally {
		URL.revokeObjectURL(url);
	}
}

async function canvasToBytes(canvas) {
	const blob = await new Promise((res) => canvas.toBlob(res, 'image/png'));
	return new Uint8Array(await blob.arrayBuffer());
}

// svgsToPdf: кожен SVG — окрема сторінка PDF (растр високої щільності у jsPDF).
async function svgsToPdf(svgs, scale = 3) {
	const { jsPDF } = await import('jspdf');
	let pdf = null;
	for (const svg of svgs) {
		const { canvas, w, h } = await svgToCanvas(svg, scale);
		const orient = w > h ? 'l' : 'p';
		if (!pdf) pdf = new jsPDF({ unit: 'pt', format: [w, h], orientation: orient });
		else pdf.addPage([w, h], orient);
		pdf.addImage(canvas.toDataURL('image/png'), 'PNG', 0, 0, w, h);
	}
	if (!pdf) pdf = new jsPDF();
	return new Uint8Array(pdf.output('arraybuffer'));
}

/** PNG схеми (Uint8Array). cap = { caption, figNum, capWord, capFormat }; scale — щільність. */
export async function renderPng(diagram, cap, scale = 2, onProgress) {
	onProgress?.('Рендер PNG…');
	const { canvas } = await svgToCanvas(renderSvg({ ...diagram, ...cap }), scale);
	return canvasToBytes(canvas);
}

/** Один PNG з УСІХ схем (Uint8Array). */
export async function renderPngAll(diagrams, scale = 2, onProgress) {
	onProgress?.('Рендер PNG…');
	const { canvas } = await svgToCanvas(engRenderSvgAll(diagrams), scale);
	return canvasToBytes(canvas);
}

/** PDF схеми (Uint8Array). */
export async function renderPdf(diagram, cap, onProgress) {
	onProgress?.('Рендер PDF…');
	return svgsToPdf([renderSvg({ ...diagram, ...cap })]);
}

/** Багатосторінковий PDF з УСІХ схем (Uint8Array). */
export async function renderPdfAll(diagrams, onProgress) {
	onProgress?.('Рендер PDF…');
	return svgsToPdf(diagrams.map((d) => renderSvg(d)));
}
