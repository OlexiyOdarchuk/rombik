// Двигун у браузері, без сервера:
//   Pyodide (CPython у WASM) виконує parser.py -> AST-JSON
//   rombik.wasm (Go) бере AST-JSON + опції -> {functions:[{name, svg, diagram}]}
import { base } from '$app/paths';

const PYODIDE_VER = '0.27.2';
const PYODIDE_CDN = `https://cdn.jsdelivr.net/pyodide/v${PYODIDE_VER}/full/`;

let initPromise = null; // ініціалізація один раз
let pyodide = null;
let parserSrc = '';

function loadScript(src) {
	return new Promise((resolve, reject) => {
		const s = document.createElement('script');
		s.src = src;
		s.onload = resolve;
		s.onerror = () => reject(new Error('не вдалося завантажити ' + src));
		document.head.appendChild(s);
	});
}

// onProgress(stage) — для індикатора («Завантаження Python…» тощо).
async function init(onProgress) {
	onProgress?.('Завантаження середовища Python…');
	await loadScript(PYODIDE_CDN + 'pyodide.js');
	pyodide = await globalThis.loadPyodide({ indexURL: PYODIDE_CDN });

	onProgress?.('Завантаження двигуна…');
	parserSrc = await (await fetch(`${base}/parser.py`)).text();
	await loadScript(`${base}/wasm_exec.js`);
	const go = new globalThis.Go();
	const bytes = await (await fetch(`${base}/rombik.wasm`)).arrayBuffer();
	const { instance } = await WebAssembly.instantiate(bytes, go.importObject);
	go.run(instance); // НЕ await: main блокується на select{}, лишаючись живим
}

/** Готує середовище (ідемпотентно). Можна викликати заздалегідь для прогріву. */
export function warmup(onProgress) {
	if (!initPromise) initPromise = init(onProgress);
	return initPromise;
}

/**
 * generate(code, options) -> { functions:[{name, svg, diagram}] } або { error }.
 * Розбір Python робить Pyodide (ast.parse, код НЕ виконується).
 */
export async function generate(code, options = {}, onProgress) {
	await warmup(onProgress);
	onProgress?.('Будую схему…');

	pyodide.globals.set('src', code);
	try {
		pyodide.runPython(parserSrc);
	} catch (e) {
		return { error: cleanPyError(String(e?.message ?? e)) };
	}
	const astJSON = pyodide.globals.get('_out');
	try {
		return JSON.parse(globalThis.rombikGenerate(astJSON, JSON.stringify(options)));
	} catch (e) {
		return { error: 'двигун: ' + (e?.message ?? e) };
	}
}

// Витягуємо людську суть із трейсбеку Python (останній рядок — «SyntaxError: …»).
function cleanPyError(msg) {
	const lines = msg.trim().split('\n').filter(Boolean);
	const last = lines[lines.length - 1] || msg;
	return last.replace(/^\s*/, '');
}
