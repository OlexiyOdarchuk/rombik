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

// --- Typst → PDF у браузері (typst.ts) ---
// Компілятор важкий (~10 МБ wasm), тож вантажимо ЛИШЕ на першу вимогу PDF.
const TYPST_VER = '0.7.0';
const TYPST_CDN = (pkg, file) => `https://cdn.jsdelivr.net/npm/@myriaddreamin/${pkg}@${TYPST_VER}/pkg/${file}`;
let typstPromise = null;

async function getTypst() {
	if (!typstPromise)
		typstPromise = (async () => {
			const { $typst } = await import('@myriaddreamin/typst.ts/dist/esm/contrib/snippet.mjs');
			$typst.setCompilerInitOptions({
				getModule: () => TYPST_CDN('typst-ts-web-compiler', 'typst_ts_web_compiler_bg.wasm')
			});
			$typst.setRendererInitOptions({
				getModule: () => TYPST_CDN('typst-ts-renderer', 'typst_ts_renderer_bg.wasm')
			});
			return $typst;
		})();
	return typstPromise;
}

/** Компілює Typst-код у PDF (Uint8Array). Потребує мережі (cetz-пакет + wasm). */
export async function typstToPdf(code, onProgress) {
	onProgress?.('Завантаження Typst-компілятора…');
	const $typst = await getTypst();
	onProgress?.('Компіляція PDF…');
	return await $typst.pdf({ mainContent: code });
}

// Витягуємо людську суть із трейсбеку Python (останній рядок — «SyntaxError: …»).
function cleanPyError(msg) {
	const lines = msg.trim().split('\n').filter(Boolean);
	const last = lines[lines.length - 1] || msg;
	return last.replace(/^\s*/, '');
}
