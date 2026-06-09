// Рушій у браузері, без сервера:
//   web-tree-sitter парсить код і формує AST-JSON
//   rombik.wasm (Go) бере AST-JSON + опції -> {functions:[{name, svg, diagram}]}
import { base } from '$app/paths';
import { parseTreeToAstJson } from './parser.js';

let initPromise = null; // ініціалізація один раз
let parser = null;
let langs = {}; // 'python' -> Language, 'cpp' -> Language

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

// onProgress(stage) — для індикатора.
async function init(onProgress) {
	onProgress?.('Завантаження Tree-sitter…');
	const modulePath = `${base}/tree-sitter.js`;
	const { Parser, Language } = await import(/* @vite-ignore */ modulePath);
	await Parser.init({
		locateFile: () => `${base}/tree-sitter.wasm`
	});
	parser = new Parser();
	langs.python = await Language.load(`${base}/tree-sitter-python.wasm`);
	langs.cpp = await Language.load(`${base}/tree-sitter-cpp.wasm`);

	onProgress?.('Завантаження рушія…');
	await loadScript(`${base}/wasm_exec.js`);
	const go = new globalThis.Go();
	const instance = await instantiateWasm(fetch(`${base}/rombik.wasm`), go.importObject);
	go.run(instance); // НЕ await: main блокується на select{}, лишаючись живим
}

/** Готує середовище (ідемпотентно). Можна викликати заздалегідь для прогріву. */
export function warmup(onProgress) {
	if (!initPromise) initPromise = init(onProgress);
	return initPromise;
}

/**
 * generate(code, options) -> { functions:[{name, svg, diagram}] } або { error }.
 */
export async function generate(code, options = {}, onProgress) {
	await warmup(onProgress);
	onProgress?.('Будую схему…');

	try {
		const langStr = options.lang === 'cpp' ? 'cpp' : 'python';
		parser.setLanguage(langs[langStr]);
		const tree = parser.parse(code);
		
		const astJSON = parseTreeToAstJson(tree, langStr);
		lastAst = astJSON; // тримаємо для розбивки (потрібен AST, не лише diagram)
		const res = JSON.parse(globalThis.rombikGenerate(astJSON, JSON.stringify(options)));
		
		if (tree.rootNode.hasError) {
			res.warning = "У коді є синтаксичні помилки. Деякі блоки можуть бути згенеровані неправильно.";
		}
		
		return res;
	} catch (e) {
		return { error: 'рушій: ' + (e?.message ?? e) };
	}
}

let lastAst = null;

/** Ріже схему функції на зв'язані частини (кнопка «Розбити на частини»).
 *  Повертає { parts:[{name,caption,figNum,svg,typst,diagram}] } або { error }. */
export function splitSchema(name, maxH, options = {}) {
	if (!lastAst) return { error: 'спершу побудуй схему' };
	try {
		return JSON.parse(globalThis.rombikSplit(lastAst, JSON.stringify(options), name, maxH));
	} catch (e) {
		return { error: 'розбивка: ' + (e?.message ?? e) };
	}
}

/**
 * Дешевий ре-рендер однієї схеми після зміни підпису (без розбору коду).
 * cap = { caption, figNum, capWord }. Повертає { svg, typst } або { error }.
 */
export function renderCaption(diagram, cap) {
	try {
		return JSON.parse(globalThis.rombikRenderOne(JSON.stringify(diagram), JSON.stringify(cap)));
	} catch (e) {
		return { error: 'ре-рендер: ' + (e?.message ?? e) };
	}
}

// --- Нативний PNG/PDF (важкий raster-wasm, вантажиться лениво) ---
// Окремий Go-WASM (tdewolff/canvas) ~16 МБ — тягнемо ЛИШЕ на першу вимогу
// експорту PNG/PDF, щоб не роздувати початкове завантаження.
let rasterPromise = null;

function loadRaster(onProgress) {
	if (!rasterPromise)
		rasterPromise = (async () => {
			onProgress?.('Завантаження рушія PNG/PDF…');
			if (!globalThis.Go) await loadScript(`${base}/wasm_exec.js`);
			const go = new globalThis.Go();
			const instance = await instantiateWasm(fetch(`${base}/rombik-raster.wasm`), go.importObject);
			go.run(instance); // НЕ await: main блокується на select{}
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

/** Typst однієї схеми. fragment=true → лише cetz.canvas (без преамбули). */
export function renderTypst(diagram, fragment = false) {
	const res = JSON.parse(globalThis.rombikTypstOne(JSON.stringify(diagram), fragment));
	if (res.error) throw new Error(res.error);
	return res.typst;
}

/** Один Typst з УСІХ схем. fragment=true → лише canvas-блоки (без преамбули). */
export function renderTypstAll(diagrams, fragment = false) {
	const res = JSON.parse(globalThis.rombikTypstAll(JSON.stringify(diagrams), fragment));
	if (res.error) throw new Error(res.error);
	return res.typst;
}

/** Один SVG з УСІХ схем (вертикально). */
export function renderSvgAll(diagrams) {
	const res = JSON.parse(globalThis.rombikSvgAll(JSON.stringify(diagrams)));
	if (res.error) throw new Error(res.error);
	return res.svg;
}

/** Схема у форматі .excalidraw (для excalidraw.com). */
export function renderExcalidraw(diagram) {
	const res = JSON.parse(globalThis.rombikExcalidraw(JSON.stringify(diagram)));
	if (res.error) throw new Error(res.error);
	return res.excalidraw;
}

/** Усі схеми в одному .excalidraw. */
export function renderExcalidrawAll(diagrams) {
	const res = JSON.parse(globalThis.rombikExcalidrawAll(JSON.stringify(diagrams)));
	if (res.error) throw new Error(res.error);
	return res.excalidraw;
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

// Витягуємо людську суть із трейсбеку Python (останній рядок — «SyntaxError: …»).
function cleanPyError(msg) {
	const lines = msg.trim().split('\n').filter(Boolean);
	const last = lines[lines.length - 1] || msg;
	return last.replace(/^\s*/, '');
}
