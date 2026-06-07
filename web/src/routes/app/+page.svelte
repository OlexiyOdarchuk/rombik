<script>
	import { generate, warmup, renderCaption, renderPdf, renderPng } from '$lib/engine.js';
	import CodeEditor from '$lib/CodeEditor.svelte';
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

	// Налаштування (галочки/списки) -> опції двигуна.
	let s = $state({
		singleEnd: false, // false = Кінець на кожен вихід
		callAsProcess: false,
		stripTypes: false,
		returnAsIO: false,
		branch: 'так', // так | yes | pm
		io: 'short', // short | verbose | imperative
		showCaption: true, // підпис «Рисунок N — …» під схемою
		capWord: 'Рисунок', // слово підпису (Рисунок / Рис. / своє)
		pngScale: 3 // якість PNG (пікселів на одиницю)
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
		capWord: s.capWord
	});

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
		download(`${f.name}.typ`, f.typst, 'text/plain');
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

	function exportSvg(f) {
		download(`${f.name}.svg`, f.svg, 'image/svg+xml');
	}

	async function exportPng(f) {
		busy = true;
		errored = false;
		try {
			const png = await renderPng(f.diagram, capOpts(f), s.pngScale, (st) => (status = st));
			download(`${f.name}.png`, png, 'image/png');
			status = 'PNG готовий.';
		} catch (e) {
			errored = true;
			status = 'PNG не вдався: ' + (e?.message ?? e);
		} finally {
			busy = false;
		}
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

		<!-- Налаштування (випадайна панель) -->
		<div class="relative">
			<button
				onclick={() => (showSettings = !showSettings)}
				class="flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-400 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:hover:border-slate-600"
			>
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
					<circle cx="12" cy="12" r="3" />
					<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
				</svg>
				Налаштування
			</button>

			{#if showSettings}
				<button class="fixed inset-0 z-10 cursor-default" onclick={() => (showSettings = false)} aria-label="Закрити"></button>
				<div class="absolute left-0 top-full z-20 mt-2 w-72 space-y-4 rounded-xl border border-slate-200 bg-white p-4 shadow-xl dark:border-slate-700 dark:bg-slate-900">
					<div class="space-y-2.5">
						<p class="text-xs font-semibold uppercase tracking-wide text-slate-400">Структура</p>
						{#each [['singleEnd', 'Один Кінець на всю схему'], ['callAsProcess', 'Виклик — звичайним блоком'], ['stripTypes', 'Без тип-анотацій'], ['returnAsIO', 'return — паралелограмом']] as [key, label] (key)}
							<label class="flex cursor-pointer items-center gap-2.5 text-sm text-slate-700 dark:text-slate-300">
								<input type="checkbox" bind:checked={s[key]} onchange={reapply} class="rounded border-slate-300" />
								{label}
							</label>
						{/each}
					</div>
					<div class="space-y-2.5 border-t border-slate-100 pt-3 dark:border-slate-700">
						<p class="text-xs font-semibold uppercase tracking-wide text-slate-400">Підписи</p>
						<label class="flex items-center justify-between text-sm text-slate-700 dark:text-slate-300">
							Гілки
							<select bind:value={s.branch} onchange={reapply} class="rounded-md border-slate-300 py-1 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200">
								<option value="так">Так / Ні</option>
								<option value="yes">Yes / No</option>
								<option value="pm">+ / −</option>
							</select>
						</label>
						<label class="flex items-center justify-between text-sm text-slate-700 dark:text-slate-300">
							Ввід/вивід
							<select bind:value={s.io} onchange={reapply} class="rounded-md border-slate-300 py-1 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200">
								<option value="short">Ввід / Вивід</option>
								<option value="verbose">Введення / Виведення</option>
								<option value="imperative">Ввести / Вивести</option>
							</select>
						</label>
					</div>
					<div class="space-y-2.5 border-t border-slate-100 pt-3 dark:border-slate-700">
						<p class="text-xs font-semibold uppercase tracking-wide text-slate-400">Підпис схеми</p>
						<label class="flex cursor-pointer items-center gap-2.5 text-sm text-slate-700 dark:text-slate-300">
							<input type="checkbox" bind:checked={s.showCaption} onchange={() => funcs.length && reCaptionAll()} class="rounded border-slate-300" />
							Показувати підпис «Рисунок N»
						</label>
						<label class="flex items-center justify-between gap-2 text-sm text-slate-700 dark:text-slate-300">
							Слово підпису
							<input
								type="text"
								bind:value={s.capWord}
								oninput={() => funcs.length && reCaptionAll()}
								placeholder="Рисунок"
								class="w-28 rounded-md border-slate-300 py-1 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
							/>
						</label>
						<p class="text-xs text-slate-400">Напр. «Рисунок», «Рис.», «Figure». Номер і текст — під кожною схемою.</p>
						<label class="flex items-center justify-between text-sm text-slate-700 dark:text-slate-300">
							Якість PNG
							<select bind:value={s.pngScale} class="rounded-md border-slate-300 py-1 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200">
								<option value={2}>2× (екран)</option>
								<option value={3}>3× (друк)</option>
								<option value={4}>4× (макс.)</option>
							</select>
						</label>
					</div>
				</div>
			{/if}
		</div>

		{#if funcs.length}
			<span class="ml-auto text-sm text-slate-500 dark:text-slate-400">{funcs.length} схем</span>
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
					{#each funcs as f (f.name)}
						<div class="rounded-lg border border-slate-200 bg-white shadow-sm dark:border-slate-700 dark:bg-slate-800">
							<div class="flex flex-wrap items-center gap-2 border-b border-slate-100 px-3 py-2 dark:border-slate-700">
								<span class="font-mono text-xs text-slate-400 dark:text-slate-500">{f.name}</span>
								<div class="flex shrink-0 gap-1">
									<button onclick={() => exportSvg(f)} class="rounded border border-slate-200 px-2 py-1 text-xs font-medium text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700">SVG</button>
									<button onclick={() => exportPng(f)} disabled={busy} class="rounded border border-slate-200 px-2 py-1 text-xs font-medium text-slate-600 transition hover:bg-slate-100 disabled:opacity-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700">PNG</button>
									<button onclick={() => exportTypst(f)} class="rounded border border-slate-200 px-2 py-1 text-xs font-medium text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700">Typst</button>
									<button onclick={() => exportPdf(f)} disabled={busy} class="rounded border border-blue-200 bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700 transition hover:bg-blue-100 disabled:opacity-50 dark:border-blue-900 dark:bg-blue-950 dark:text-blue-300 dark:hover:bg-blue-900">PDF</button>
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
							<div class="schema m-3 grid place-items-center overflow-auto rounded-md bg-white p-3 dark:ring-1 dark:ring-slate-700">{@html f.svg}</div>
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

	<p class="mt-3 text-xs {errored ? 'text-red-600 dark:text-red-400' : 'text-slate-500 dark:text-slate-400'}">{status}</p>
</div>
