// Рушій у браузері, без сервера й без Go-WASM:
//   web-tree-sitter парсить код → Tree
//   @rombik/engine (чистий TS) бере Tree → схеми → SVG / Typst / Excalidraw
// PNG/PDF поки через лінивий растер-WASM (заміниться браузерним canvas).
import { base } from '$app/paths';
import {
	parseTree, fromAst, splitFromAst,
	renderSvg, renderSvgAll as engRenderSvgAll,
	renderTypst as engRenderTypst, renderTypstAll as engRenderTypstAll, renderTypstFragment, renderTypstFragmentAll,
	renderExcalidraw as engRenderExcalidraw, renderExcalidrawAll as engRenderExcalidrawAll,
} from '@rombik/engine';

let initPromise = null;
let parser = null;
let langs = {}; // 'python' | 'cpp' -> Language

function loadScript(src) {
	return new Promise((resolve, reject) => {
		const s = document.createElement('script');
		s.src = src;
		s.onload = resolve;
		s.onerror = () => reject(new Error('не вдалося завантажити ' + src));
		document.head.appendChild(s);
	});
}

async function instantiateWasm(resp, importObject) {
	if (WebAssembly.instantiateStreaming) {
		try {
			return (await WebAssembly.instantiateStreaming(resp, importObject)).instance;
		} catch (e) {
			console.warn('WASM streaming failed, fallback to arrayBuffer', e);
		}
	}
	const bytes = await (await resp).arrayBuffer();
	return (await WebAssembly.instantiate(bytes, importObject)).instance;
}

// init: лише tree-sitter (рушій тепер чистий TS, wasm не потрібен).
async function init(onProgress) {
	onProgress?.('Завантаження Tree-sitter…');
	const modulePath = `${base}/tree-sitter.js`;
	const { Parser, Language } = await import(/* @vite-ignore */ modulePath);
	await Parser.init({ locateFile: () => `${base}/tree-sitter.wasm` });
	parser = new Parser();
	langs.python = await Language.load(`${base}/tree-sitter-python.wasm`);
	langs.cpp = await Language.load(`${base}/tree-sitter-cpp.wasm`);
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
		const lang = options.lang === 'cpp' ? 'cpp' : 'python';
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

// --- Нативний PNG/PDF (поки через лінивий растер-WASM; tdewolff/canvas ~16 МБ) ---
// TODO: замінити на браузерний SVG→canvas і викинути растер-WASM.
let rasterPromise = null;

function loadRaster(onProgress) {
	if (!rasterPromise)
		rasterPromise = (async () => {
			onProgress?.('Завантаження рушія PNG/PDF…');
			if (!globalThis.Go) await loadScript(`${base}/wasm_exec.js`);
			const go = new globalThis.Go();
			const instance = await instantiateWasm(fetch(`${base}/rombik-raster.wasm`), go.importObject);
			go.run(instance);
		})();
	return rasterPromise;
}

function b64ToBytes(b64) {
	const bin = atob(b64);
	const bytes = new Uint8Array(bin.length);
	for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
	return bytes;
}

/** Нативний PDF схеми (Uint8Array). cap = { caption, figNum, capWord, capFormat }. */
export async function renderPdf(diagram, cap, onProgress) {
	await loadRaster(onProgress);
	const res = JSON.parse(globalThis.rombikPdf(JSON.stringify(diagram), JSON.stringify(cap)));
	if (res.error) throw new Error(res.error);
	return b64ToBytes(res.pdf);
}

/** Нативний PNG схеми (Uint8Array). scale — пікселів на одиницю. */
export async function renderPng(diagram, cap, scale, onProgress) {
	await loadRaster(onProgress);
	const res = JSON.parse(globalThis.rombikPng(JSON.stringify(diagram), JSON.stringify(cap), scale));
	if (res.error) throw new Error(res.error);
	return b64ToBytes(res.png);
}

/** Один PNG з УСІХ схем (Uint8Array). */
export async function renderPngAll(diagrams, scale, onProgress) {
	await loadRaster(onProgress);
	const res = JSON.parse(globalThis.rombikPngAll(JSON.stringify(diagrams), scale));
	if (res.error) throw new Error(res.error);
	return b64ToBytes(res.png);
}

/** Один багатосторінковий PDF з УСІХ схем (Uint8Array). */
export async function renderPdfAll(diagrams, onProgress) {
	await loadRaster(onProgress);
	const res = JSON.parse(globalThis.rombikPdfAll(JSON.stringify(diagrams)));
	if (res.error) throw new Error(res.error);
	return b64ToBytes(res.pdf);
}
