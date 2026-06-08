<script>
	import {
		generate,
		warmup,
		splitSchema,
		renderCaption,
		renderPdf,
		renderPng,
		renderTypst,
		renderTypstAll,
		renderPdfAll,
		renderSvgAll,
		renderPngAll,
		renderExcalidraw,
		renderExcalidrawAll
	} from '$lib/engine.js';
	import CodeEditor from '$lib/CodeEditor.svelte';
	import DiagramEditor from '$lib/DiagramEditor.svelte';
	import { onMount } from 'svelte';

	const SAMPLE = `def grade(score):
    name = input("Ваше ім'я: ")
    print("Привіт,", name)
    total = score + 5
    if total >= 90:
        print("Відмінно")
    else:
        if total >= 60:
            print("Задовільно")
        else:
            print("Незадовільно")
    print("Готово")`;

	let code = $state(SAMPLE);
	let funcs = $state([]); // [{name, svg, diagram}]
	let status = $state('Готовий. Натисни «Побудувати».');
	let busy = $state(false);
	let errored = $state(false);
	let showSettings = $state(false);
	let sTab = $state('struct'); // struct | text | export
	let editingFn = $state(null); // схема, відкрита у візуальному редакторі

	function onEditorSave(d) {
		if (!editingFn) return;
		// Замінюємо геометрію, зберігаючи поля підпису.
		editingFn.diagram = { ...editingFn.diagram, shapes: d.shapes, edges: d.edges, w: d.w, h: d.h };
		reCaption(editingFn); // ре-рендер svg/typst з відредагованої схеми
		funcs = [...funcs];
		editingFn = null;
		status = 'Схему відредаговано.';
	}

	// Налаштування (галочки/списки) -> опції рушія.
	let s = $state({
		singleEnd: false, // false = Кінець на кожен вихід
		callAsProcess: false,
		stripTypes: false,
		returnAsIO: false,
		branch: 'так', // так | yes | pm
		io: 'short', // short | verbose | imperative
		showCaption: true, // підпис «Рисунок N — …» під схемою
		capWord: 'Рисунок', // слово підпису (Рисунок / Рис. / своє)
		capFormat: '{word} {num} — {text}', // шаблон підпису
		figStart: 1, // з якого номера нумерувати схеми
		pngScale: 3, // якість PNG (пікселів на одиницю)
		typstFragment: false, // Typst лише фрагмент (без преамбули) — для вставки у свій .typ
		font: "'Times New Roman', 'Liberation Serif', 'DejaVu Serif', serif" // шрифт для SVG
	});
	const BRANCH = { так: ['Так', 'Ні'], yes: ['Yes', 'No'], pm: ['+', '−'] };
	const IO = { short: ['Ввід', 'Вивід'], verbose: ['Введення', 'Виведення'], imperative: ['Ввести', 'Вивести'] };

	function engineOpts() {
		const [yes, no] = BRANCH[s.branch];
		const [inWord, outWord] = IO[s.io];
		return {
			singleEnd: s.singleEnd,
			callAsProcess: s.callAsProcess,
			stripTypes: s.stripTypes,
			returnAsIO: s.returnAsIO,
			yes,
			no,
			inWord,
			outWord,
			capWord: s.capWord
		};
	}
	// Перебудувати при зміні налаштування, якщо схеми вже є.
	const reapply = () => funcs.length && build();

	// Опції підпису для однієї схеми (з урахуванням глобального перемикача).
	const capOpts = (f) => ({
		caption: s.showCaption ? (f.caption ?? '') : '',
		figNum: Number(f.figNum) || 0,
		capWord: s.capWord,
		capFormat: s.capFormat
	});

	// Масив діаграм із УЖЕ виставленими полями підпису — для «експортувати все».
	const exportDiagrams = () => funcs.map((f) => ({ ...f.diagram, ...capOpts(f) }));

	// Перенумерувати схеми з глобального старту (s.figStart, далі по порядку).
	function renumber() {
		funcs.forEach((f, i) => (f.figNum = s.figStart + i));
		reCaptionAll();
	}

	// Дешевий ре-рендер підпису однієї схеми / усіх (без розбору коду).
	function reCaption(f) {
		const out = renderCaption(f.diagram, capOpts(f));
		if (out.error) {
			errored = true;
			status = out.error;
			return;
		}
		f.svg = out.svg;
		f.typst = out.typst;
		funcs = [...funcs]; // явний тригер — оновити й ПЕРЕГЛЯД (@html), не лише експорт
	}
	const reCaptionAll = () => {
		funcs.forEach((f) => {
			const out = renderCaption(f.diagram, capOpts(f));
			if (!out.error) {
				f.svg = out.svg;
				f.typst = out.typst;
			}
		});
		funcs = [...funcs];
	};

	// Прогріваємо середовище заздалегідь (вантажиться Pyodide ~6 МБ).
	onMount(() => {
		warmup((s) => (status = s)).then(
			() => (status = 'Готовий. Натисни «Побудувати».'),
			(e) => ((errored = true), (status = 'Не вдалося завантажити середовище: ' + e))
		);
	});

	async function build() {
		busy = true;
		errored = false;
		try {
			const res = await generate(code, engineOpts(), (st) => (status = st));
			if (res.error) {
				errored = true;
				status = res.error;
				return;
			}
			funcs = res.functions ?? [];
			if (funcs.length && s.figStart !== 1) renumber(); // глобальний старт нумерації
			status = funcs.length ? `Готово: ${funcs.length} схем.` : 'Порожньо: нема що малювати.';
		} catch (e) {
			errored = true;
			status = 'Помилка: ' + (e?.message ?? e);
		} finally {
			busy = false;
		}
	}

	function download(name, text, type) {
		const url = URL.createObjectURL(new Blob([text], { type }));
		const a = document.createElement('a');
		a.href = url;
		a.download = name;
		a.click();
		URL.revokeObjectURL(url);
	}

	function exportTypst(f) {
		const typ = s.typstFragment ? renderTypst({ ...f.diagram, ...capOpts(f) }, true) : f.typst;
		download(`${f.name}.typ`, typ, 'text/plain');
	}

	// --- Експорт УСІХ схем одним документом/зображенням ---
	function exportAllTypst() {
		try {
			download('схеми.typ', renderTypstAll(exportDiagrams(), s.typstFragment), 'text/plain');
			status = 'Typst (усі схеми) готовий.';
		} catch (e) {
			errored = true;
			status = 'Typst: ' + (e?.message ?? e);
		}
	}

	function exportAllSvg() {
		try {
			download('схеми.svg', renderSvgAll(exportDiagrams()), 'image/svg+xml');
			status = 'SVG (усі схеми) готовий.';
		} catch (e) {
			errored = true;
			status = 'SVG: ' + (e?.message ?? e);
		}
	}

	async function exportAllPng() {
		busy = true;
		errored = false;
		try {
			const png = await renderPngAll(exportDiagrams(), s.pngScale, (st) => (status = st));
			download('схеми.png', png, 'image/png');
			status = 'PNG (усі схеми) готовий.';
		} catch (e) {
			errored = true;
			status = 'PNG: ' + (e?.message ?? e);
		} finally {
			busy = false;
		}
	}

	async function exportAllPdf() {
		busy = true;
		errored = false;
		try {
			const pdf = await renderPdfAll(exportDiagrams(), (st) => (status = st));
			download('схеми.pdf', pdf, 'application/pdf');
			status = 'PDF (усі схеми) готовий.';
		} catch (e) {
			errored = true;
			status = 'PDF: ' + (e?.message ?? e);
		} finally {
			busy = false;
		}
	}

	async function exportPdf(f) {
		busy = true;
		errored = false;
		try {
			const pdf = await renderPdf(f.diagram, capOpts(f), (st) => (status = st));
			download(`${f.name}.pdf`, pdf, 'application/pdf');
			status = 'PDF готовий.';
		} catch (e) {
			errored = true;
			status = 'PDF не вдався: ' + (e?.message ?? e);
		} finally {
			busy = false;
		}
	}

	// Розбити схему на зв'язані частини конекторами (для завеликих/складних).
	function splitFn(f) {
		const res = splitSchema(f.name, 900, engineOpts());
		if (res.error) {
			errored = true;
			status = res.error;
			return;
		}
		if (!res.parts || res.parts.length <= 1) {
			status = 'Схема компактна — ділити нема сенсу.';
			return;
		}
		const i = funcs.indexOf(f);
		funcs = [...funcs.slice(0, i), ...res.parts, ...funcs.slice(i + 1)];
		status = `Розбито на ${res.parts.length} частин (конектори ○).`;
	}

	function exportSvg(f) { download(`${f.name}.svg`, f.svg.replace(/font-family="[^"]+"/, `font-family="${s.font}"`), 'image/svg+xml'); }
	function exportExcal(f) {
		try { download(`${f.name}.excalidraw`, renderExcalidraw({ ...f.diagram, ...capOpts(f) }), 'application/json'); status = 'Excalidraw готовий.'; }
		catch (e) { errored = true; status = 'Excalidraw: ' + (e?.message ?? e); }
	}
	function exportAllExcal() {
		try { download('схеми.excalidraw', renderExcalidrawAll(exportDiagrams()), 'application/json'); status = 'Excalidraw (усі) готовий.'; }
		catch (e) { errored = true; status = 'Excalidraw: ' + (e?.message ?? e); }
	}
	async function exportPng(f) {
		busy = true; errored = false;
		try { const png = await renderPng(f.diagram, capOpts(f), s.pngScale, (st) => (status = st)); download(`${f.name}.png`, png, 'image/png'); status = 'PNG готовий.'; }
		catch (e) { errored = true; status = 'PNG не вдався: ' + (e?.message ?? e); }
		finally { busy = false; }
	}

	async function copyToClipboard(textOrBytes, type = 'text/plain') {
		try {
			if (type === 'image/png') {
				const blob = new Blob([textOrBytes], { type });
				await navigator.clipboard.write([new ClipboardItem({ 'image/png': blob })]);
			} else {
				await navigator.clipboard.writeText(textOrBytes);
			}
			errored = false;
			status = 'Скопійовано в буфер обміну!';
			setTimeout(() => { if (status === 'Скопійовано в буфер обміну!') status = ''; }, 3000);
		} catch (e) {
			errored = true;
			status = 'Помилка копіювання: ' + (e?.message ?? e);
		}
	}
	function copySvg(f) { copyToClipboard(f.svg.replace(/font-family="[^"]+"/, `font-family="${s.font}"`), 'text/plain'); }
	function copyTypst(f) { copyToClipboard(s.typstFragment ? renderTypst({ ...f.diagram, ...capOpts(f) }, true) : f.typst, 'text/plain'); }
	function copyExcal(f) { copyToClipboard(renderExcalidraw({ ...f.diagram, ...capOpts(f) }), 'text/plain'); }
	async function copyPng(f) {
		busy = true; errored = false;
		try { const png = await renderPng(f.diagram, capOpts(f), s.pngScale, (st) => (status = st)); copyToClipboard(png, 'image/png'); }
		catch (e) { errored = true; status = 'PNG копіювання не вдалося: ' + (e?.message ?? e); }
		finally { busy = false; }
	}

	function copyAllSvg() { copyToClipboard(renderSvgAll(exportDiagrams()), 'text/plain'); }
	function copyAllTypst() { copyToClipboard(renderTypstAll(exportDiagrams(), s.typstFragment), 'text/plain'); }
	function copyAllExcal() { copyToClipboard(renderExcalidrawAll(exportDiagrams()), 'text/plain'); }
	async function copyAllPng() {
		busy = true; errored = false;
		try { const png = await renderPngAll(exportDiagrams(), s.pngScale, (st) => (status = st)); copyToClipboard(png, 'image/png'); }
		catch (e) { errored = true; status = 'PNG копіювання не вдалося: ' + (e?.message ?? e); }
		finally { busy = false; }
	}
</script>

<svelte:head>
	<title>Редактор — rombik</title>
</svelte:head>

<div class="mx-auto flex h-[calc(100vh-4rem)] max-w-7xl flex-col px-4 py-4">
	<!-- toolbar -->
	<div class="mb-3 flex items-center gap-3">
		<button
			onclick={build}
			disabled={busy}
			class="rounded-lg bg-blue-600 px-5 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700 disabled:opacity-50"
		>
			{busy ? 'Будую…' : 'Побудувати схему'}
		</button>

		<!-- Налаштування -->
		<button
			onclick={() => (showSettings = true)}
			class="flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-400 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:border-slate-600"
		>
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
				<circle cx="12" cy="12" r="3" />
				<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
			</svg>
			Налаштування
		</button>

		{#if showSettings}
			<div class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4 backdrop-blur-sm transition-opacity">
				<!-- svelte-ignore a11y_consider_explicit_label -->
				<button class="absolute inset-0 cursor-default" onclick={() => (showSettings = false)}></button>
				<!-- Modal Content -->
				<div class="rise relative flex max-h-full w-full max-w-2xl flex-col overflow-hidden rounded-2xl bg-white shadow-2xl dark:border dark:border-slate-800 dark:bg-slate-900">
					<!-- Header -->
					<div class="flex items-center justify-between border-b border-slate-100 px-6 py-4 dark:border-slate-800">
						<h2 class="text-xl font-bold text-slate-800 dark:text-slate-100">Налаштування генерації</h2>
						<button onclick={() => (showSettings = false)} class="rounded-lg p-2 text-slate-400 hover:bg-slate-100 hover:text-slate-600 dark:hover:bg-slate-800 dark:hover:text-slate-300">
							<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>
						</button>
					</div>
					<!-- Body -->
					<div class="flex min-h-[350px] flex-col sm:flex-row">
						<!-- Tabs Sidebar -->
						<div class="flex shrink-0 flex-row overflow-x-auto border-b border-slate-100 bg-slate-50/50 p-4 dark:border-slate-800 dark:bg-slate-900/50 sm:w-48 sm:flex-col sm:border-b-0 sm:border-r sm:space-x-0 sm:space-y-1 space-x-2">
							<button onclick={() => (sTab = 'struct')} class="rounded-lg px-4 py-2.5 text-left text-sm font-semibold transition-colors {sTab === 'struct' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300' : 'text-slate-600 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800'}">Структура</button>
							<button onclick={() => (sTab = 'text')} class="rounded-lg px-4 py-2.5 text-left text-sm font-semibold transition-colors {sTab === 'text' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300' : 'text-slate-600 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800'}">Текст і підписи</button>
							<button onclick={() => (sTab = 'export')} class="rounded-lg px-4 py-2.5 text-left text-sm font-semibold transition-colors {sTab === 'export' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300' : 'text-slate-600 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800'}">Експорт</button>
						</div>
						<!-- Tab Content -->
						<div class="flex-1 overflow-y-auto p-6 space-y-6">
							{#if sTab === 'struct'}
								<div class="space-y-3">
									<h3 class="text-sm font-semibold uppercase tracking-wide text-slate-400">Алгоритмічні блоки</h3>
									{#each [['singleEnd', 'Один Кінець на всю схему'], ['callAsProcess', 'Виклик — звичайним блоком'], ['stripTypes', 'Без тип-анотацій'], ['returnAsIO', 'return — паралелограмом']] as [key, label] (key)}
										<label class="flex cursor-pointer items-center gap-3 text-sm text-slate-700 dark:text-slate-300">
											<input type="checkbox" bind:checked={s[key]} onchange={reapply} class="rounded border-slate-300 h-4 w-4" />
											{label}
										</label>
									{/each}
								</div>
							{:else if sTab === 'text'}
								<div class="space-y-4">
									<h3 class="text-sm font-semibold uppercase tracking-wide text-slate-400">Формулювання всередині</h3>
									<label class="flex items-center justify-between text-sm text-slate-700 dark:text-slate-300">
										Текст гілок (if/while)
										<select bind:value={s.branch} onchange={reapply} class="w-40 rounded-md border-slate-300 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200">
											<option value="так">Так / Ні</option>
											<option value="yes">Yes / No</option>
											<option value="pm">+ / −</option>
										</select>
									</label>
									<label class="flex items-center justify-between text-sm text-slate-700 dark:text-slate-300">
										Текст вводу/виводу
										<select bind:value={s.io} onchange={reapply} class="w-40 rounded-md border-slate-300 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200">
											<option value="short">Ввід / Вивід</option>
											<option value="verbose">Введення / Виведення</option>
											<option value="imperative">Ввести / Вивести</option>
										</select>
									</label>

									<h3 class="mt-6 text-sm font-semibold uppercase tracking-wide text-slate-400">Підписи діаграм</h3>
									<label class="flex cursor-pointer items-center gap-3 text-sm text-slate-700 dark:text-slate-300">
										<input type="checkbox" bind:checked={s.showCaption} onchange={() => funcs.length && reCaptionAll()} class="rounded border-slate-300 h-4 w-4" />
										Генерувати підпис («Рисунок N...»)
									</label>
									<div class="grid grid-cols-2 gap-4">
										<div>
											<label class="block text-xs text-slate-500 mb-1">Слово підпису</label>
											<input type="text" bind:value={s.capWord} oninput={() => funcs.length && reCaptionAll()} placeholder="Рисунок" class="w-full rounded-md border-slate-300 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200" />
										</div>
										<div>
											<label class="block text-xs text-slate-500 mb-1">Нумерувати з</label>
											<input type="number" min="1" bind:value={s.figStart} oninput={() => funcs.length && renumber()} class="w-full rounded-md border-slate-300 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200" />
										</div>
									</div>
									<div>
										<label class="block text-xs text-slate-500 mb-1">Шаблон</label>
										<input type="text" bind:value={s.capFormat} oninput={() => funcs.length && reCaptionAll()} placeholder="{'{word} {num} — {text}'}" class="w-full rounded-md border-slate-300 py-1.5 font-mono text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200" />
										<p class="text-[10px] text-slate-400 mt-1">Доступно: <code>{'{word}'}</code> <code>{'{num}'}</code> <code>{'{text}'}</code></p>
									</div>
								</div>
							{:else if sTab === 'export'}
								<div class="space-y-4">
									<h3 class="text-sm font-semibold uppercase tracking-wide text-slate-400">Параметри файлів</h3>
									<label class="flex items-center justify-between text-sm text-slate-700 dark:text-slate-300">
										Шрифт схеми (SVG)
										<select bind:value={s.font} onchange={() => (funcs = [...funcs])} class="w-48 rounded-md border-slate-300 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200">
											<option value="'Times New Roman', 'Liberation Serif', 'DejaVu Serif', serif">Times New Roman</option>
											<option value="Arial, Helvetica, sans-serif">Arial</option>
											<option value="'Courier New', Courier, monospace">Courier New</option>
											<option value="'Outfit', sans-serif">Outfit (Інтерфейс)</option>
											<option value="'JetBrains Mono', monospace">JetBrains Mono (Код)</option>
										</select>
									</label>
									<label class="flex items-center justify-between text-sm text-slate-700 dark:text-slate-300">
										Якість PNG
										<select bind:value={s.pngScale} class="w-48 rounded-md border-slate-300 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200">
											<option value={2}>2× (Екранна якість)</option>
											<option value={3}>3× (Для друку)</option>
											<option value={4}>4× (Максимальна)</option>
										</select>
									</label>
									<label class="flex cursor-pointer items-center gap-3 text-sm text-slate-700 dark:text-slate-300 mt-4">
										<input type="checkbox" bind:checked={s.typstFragment} class="rounded border-slate-300 h-4 w-4" />
										<div>
											<p>Typst: фрагмент без преамбули</p>
											<p class="text-xs text-slate-500 dark:text-slate-400">Зручно для вставки в існуючі .typ документи</p>
										</div>
									</label>
								</div>
							{/if}
						</div>
					</div>
				</div>
			</div>
		{/if}
		{#if funcs.length}
			<div class="ml-auto flex items-center gap-3 border-l border-slate-200 pl-4 dark:border-slate-700">
				<span class="text-xs font-semibold text-slate-500 dark:text-slate-400">УСІ СХЕМИ ({funcs.length}):</span>
				<div class="group relative">
					<button class="flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 shadow-sm transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700">
						<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/></svg>
						Завантажити
					</button>
					<div class="absolute right-0 top-full z-10 hidden w-32 pt-2 group-hover:block">
						<div class="flex flex-col overflow-hidden rounded-md border border-slate-200 bg-white shadow-xl dark:border-slate-700 dark:bg-slate-800">
							{#each [['SVG', exportAllSvg], ['PNG', exportAllPng], ['Typst', exportAllTypst], ['PDF', exportAllPdf], ['Excalidraw', exportAllExcal]] as [label, fn] (label)}
								<button onclick={fn} disabled={busy} class="px-4 py-2 text-left text-xs font-medium transition hover:bg-slate-100 disabled:opacity-50 dark:hover:bg-slate-700">{label}</button>
							{/each}
						</div>
					</div>
				</div>
				<div class="group relative">
					<button class="flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 shadow-sm transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700">
						<svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
						Копіювати
					</button>
					<div class="absolute right-0 top-full z-10 hidden w-32 pt-2 group-hover:block">
						<div class="flex flex-col overflow-hidden rounded-md border border-slate-200 bg-white shadow-xl dark:border-slate-700 dark:bg-slate-800">
							{#each [['SVG', copyAllSvg], ['PNG', copyAllPng], ['Typst', copyAllTypst], ['Excalidraw', copyAllExcal]] as [label, fn] (label)}
								<button onclick={fn} disabled={busy} class="px-4 py-2 text-left text-xs font-medium transition hover:bg-slate-100 disabled:opacity-50 dark:hover:bg-slate-700">{label}</button>
							{/each}
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- split -->
	<div class="grid min-h-0 flex-1 gap-3 lg:grid-cols-2">
		<!-- code -->
		<div class="flex min-h-0 flex-col rounded-xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
			<div class="border-b border-slate-200 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-slate-500 dark:border-slate-800 dark:text-slate-400">
				Код Python
			</div>
			<div class="min-h-0 flex-1 overflow-hidden rounded-b-xl">
				<CodeEditor bind:value={code} />
			</div>
		</div>

		<!-- preview: усі схеми списком -->
		<div class="flex min-h-0 flex-col rounded-xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
			<div class="border-b border-slate-200 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-slate-500 dark:border-slate-800 dark:text-slate-400">
				{funcs.length ? `Схеми (${funcs.length})` : 'Схема'}
			</div>
			<div class="min-h-0 flex-1 space-y-4 overflow-auto grid-bg p-4">
				{#if funcs.length}
					{#each funcs as f, i (f.name + i)}
						<div class="rounded-lg border border-slate-200 bg-white shadow-sm dark:border-slate-700 dark:bg-slate-800">
							<div class="flex flex-wrap items-center gap-2 border-b border-slate-100 px-3 py-2 dark:border-slate-700">
								<span class="font-mono text-xs text-slate-400 dark:text-slate-500">{f.name}</span>
								<div class="flex shrink-0 gap-2">
									<button onclick={() => (editingFn = f)} class="rounded border border-blue-300 bg-blue-50/50 px-2.5 py-1 text-xs font-semibold text-blue-700 transition hover:bg-blue-100 dark:border-blue-800 dark:bg-blue-900/20 dark:text-blue-300 dark:hover:bg-blue-900/50">✎ Редагувати</button>
									<button onclick={() => splitFn(f)} class="rounded border border-amber-300 bg-amber-50/50 px-2.5 py-1 text-xs font-semibold text-amber-700 transition hover:bg-amber-100 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-400 dark:hover:bg-amber-900/50">✂ Розбити</button>
									<div class="h-4 w-px bg-slate-200 dark:bg-slate-600 self-center"></div>
									
									<div class="group relative">
										<button class="flex items-center gap-1 rounded border border-slate-200 px-2.5 py-1 text-xs font-semibold text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700">
											Завантажити ▾
										</button>
										<div class="absolute right-0 top-full z-10 hidden w-28 pt-1 group-hover:block">
											<div class="flex flex-col overflow-hidden rounded border border-slate-200 bg-white shadow-lg dark:border-slate-700 dark:bg-slate-800">
												{#each [['SVG', () => exportSvg(f)], ['PNG', () => exportPng(f)], ['Typst', () => exportTypst(f)], ['PDF', () => exportPdf(f)], ['Excalidraw', () => exportExcal(f)]] as [label, fn] (label)}
													<button onclick={fn} disabled={busy} class="px-3 py-1.5 text-left text-xs font-medium transition hover:bg-slate-100 disabled:opacity-50 dark:hover:bg-slate-700">{label}</button>
												{/each}
											</div>
										</div>
									</div>

									<div class="group relative">
										<button class="flex items-center gap-1 rounded border border-slate-200 px-2.5 py-1 text-xs font-semibold text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700">
											Копіювати ▾
										</button>
										<div class="absolute right-0 top-full z-10 hidden w-28 pt-1 group-hover:block">
											<div class="flex flex-col overflow-hidden rounded border border-slate-200 bg-white shadow-lg dark:border-slate-700 dark:bg-slate-800">
												{#each [['SVG', () => copySvg(f)], ['PNG', () => copyPng(f)], ['Typst', () => copyTypst(f)], ['Excalidraw', () => copyExcal(f)]] as [label, fn] (label)}
													<button onclick={fn} disabled={busy} class="px-3 py-1.5 text-left text-xs font-medium transition hover:bg-slate-100 disabled:opacity-50 dark:hover:bg-slate-700">{label}</button>
												{/each}
											</div>
										</div>
									</div>
								</div>
								{#if s.showCaption}
									<div class="flex w-full items-center gap-1.5">
										<span class="shrink-0 text-xs text-slate-500 dark:text-slate-400">{s.capWord || 'Рисунок'}</span>
										<input
											type="number"
											min="0"
											bind:value={f.figNum}
											oninput={() => reCaption(f)}
											class="w-14 rounded-md border-slate-300 px-1.5 py-1 text-xs dark:border-slate-600 dark:bg-slate-900 dark:text-slate-200"
										/>
										<span class="shrink-0 text-xs text-slate-400">—</span>
										<input
											type="text"
											bind:value={f.caption}
											oninput={() => reCaption(f)}
											placeholder="підпис схеми"
											class="min-w-0 flex-1 rounded-md border-slate-300 px-2 py-1 text-xs dark:border-slate-600 dark:bg-slate-900 dark:text-slate-200"
										/>
									</div>
								{/if}
							</div>
							<!-- eslint-disable-next-line svelte/no-at-html-tags -->
							<!-- схема — чорне-по-білому, тримаємо світлий «аркуш» навіть у темній темі -->
							<div class="schema m-3 grid place-items-center overflow-auto rounded-md bg-white p-3 dark:ring-1 dark:ring-slate-700">{@html f.svg.replace(/font-family="[^"]+"/, `font-family="${s.font}"`)}</div>
						</div>
					{/each}
				{:else}
					<p class="grid h-full place-items-center text-sm text-slate-400 dark:text-slate-500">
						Встав код і натисни «Побудувати схему»
					</p>
				{/if}
			</div>
		</div>
	</div>
	{#if errored}
		<div class="mt-4 rounded-xl border-l-4 border-red-500 bg-red-50 p-4 shadow-sm dark:border-red-600 dark:bg-red-950/30">
			<div class="flex gap-3">
				<svg class="mt-0.5 h-5 w-5 shrink-0 text-red-500 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
				</svg>
				<div class="overflow-hidden">
					<h3 class="text-sm font-bold text-red-800 dark:text-red-300">Помилка</h3>
					<pre class="mt-1 max-h-40 overflow-y-auto break-words whitespace-pre-wrap font-mono text-xs text-red-700 dark:text-red-400">{status}</pre>
				</div>
			</div>
		</div>
	{:else}
		<p class="mt-4 flex items-center gap-2 px-1 text-sm text-slate-500 dark:text-slate-400">{status}</p>
	{/if}
</div>

{#if editingFn}
	<DiagramEditor diagram={editingFn.diagram} onsave={onEditorSave} oncancel={() => (editingFn = null)} />
{/if}
