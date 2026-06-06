<script>
	import { generate, warmup } from '$lib/engine.js';
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
		io: 'short' // short | verbose | imperative
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
			outWord
		};
	}
	// Перебудувати при зміні налаштування, якщо схеми вже є.
	const reapply = () => funcs.length && build();

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

	function exportSvg(f) {
		const blob = new Blob([f.svg], { type: 'image/svg+xml' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `${f.name}.svg`;
		a.click();
		URL.revokeObjectURL(url);
	}

	function exportPng(f) {
		const scale = 2;
		const img = new Image();
		const url = URL.createObjectURL(new Blob([f.svg], { type: 'image/svg+xml' }));
		img.onload = () => {
			const c = document.createElement('canvas');
			c.width = f.diagram.w * scale;
			c.height = f.diagram.h * scale;
			const ctx = c.getContext('2d');
			ctx.scale(scale, scale);
			ctx.drawImage(img, 0, 0);
			URL.revokeObjectURL(url);
			c.toBlob((b) => {
				const u = URL.createObjectURL(b);
				const a = document.createElement('a');
				a.href = u;
				a.download = `${f.name}.png`;
				a.click();
				URL.revokeObjectURL(u);
			});
		};
		img.src = url;
	}
</script>

<svelte:head>
	<title>Редактор — flowgen</title>
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
				class="flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-400"
			>
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
					<circle cx="12" cy="12" r="3" />
					<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
				</svg>
				Налаштування
			</button>

			{#if showSettings}
				<button class="fixed inset-0 z-10 cursor-default" onclick={() => (showSettings = false)} aria-label="Закрити"></button>
				<div class="absolute left-0 top-full z-20 mt-2 w-72 space-y-4 rounded-xl border border-slate-200 bg-white p-4 shadow-xl">
					<div class="space-y-2.5">
						<p class="text-xs font-semibold uppercase tracking-wide text-slate-400">Структура</p>
						{#each [['singleEnd', 'Один Кінець на всю схему'], ['callAsProcess', 'Виклик — звичайним блоком'], ['stripTypes', 'Без тип-анотацій'], ['returnAsIO', 'return — паралелограмом']] as [key, label] (key)}
							<label class="flex cursor-pointer items-center gap-2.5 text-sm text-slate-700">
								<input type="checkbox" bind:checked={s[key]} onchange={reapply} class="rounded border-slate-300" />
								{label}
							</label>
						{/each}
					</div>
					<div class="space-y-2.5 border-t border-slate-100 pt-3">
						<p class="text-xs font-semibold uppercase tracking-wide text-slate-400">Підписи</p>
						<label class="flex items-center justify-between text-sm text-slate-700">
							Гілки
							<select bind:value={s.branch} onchange={reapply} class="rounded-md border-slate-300 py-1 text-sm">
								<option value="так">Так / Ні</option>
								<option value="yes">Yes / No</option>
								<option value="pm">+ / −</option>
							</select>
						</label>
						<label class="flex items-center justify-between text-sm text-slate-700">
							Ввід/вивід
							<select bind:value={s.io} onchange={reapply} class="rounded-md border-slate-300 py-1 text-sm">
								<option value="short">Ввід / Вивід</option>
								<option value="verbose">Введення / Виведення</option>
								<option value="imperative">Ввести / Вивести</option>
							</select>
						</label>
					</div>
				</div>
			{/if}
		</div>

		{#if funcs.length}
			<span class="ml-auto text-sm text-slate-500">{funcs.length} схем</span>
		{/if}
	</div>

	<!-- split -->
	<div class="grid min-h-0 flex-1 gap-3 lg:grid-cols-2">
		<!-- code -->
		<div class="flex min-h-0 flex-col rounded-xl border border-slate-200 bg-white">
			<div class="border-b border-slate-200 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-slate-500">
				Код Python
			</div>
			<div class="min-h-0 flex-1 overflow-hidden rounded-b-xl">
				<CodeEditor bind:value={code} />
			</div>
		</div>

		<!-- preview: усі схеми списком -->
		<div class="flex min-h-0 flex-col rounded-xl border border-slate-200 bg-white">
			<div class="border-b border-slate-200 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-slate-500">
				{funcs.length ? `Схеми (${funcs.length})` : 'Схема'}
			</div>
			<div class="min-h-0 flex-1 space-y-4 overflow-auto grid-bg p-4">
				{#if funcs.length}
					{#each funcs as f (f.name)}
						<div class="rounded-lg border border-slate-200 bg-white shadow-sm">
							<div class="flex items-center justify-between gap-2 border-b border-slate-100 px-3 py-2">
								<span class="truncate font-mono text-sm font-semibold text-slate-700">{f.name}</span>
								<div class="flex shrink-0 gap-1">
									<button
										onclick={() => exportSvg(f)}
										class="rounded border border-slate-200 px-2 py-1 text-xs font-medium text-slate-600 transition hover:bg-slate-100"
									>
										SVG
									</button>
									<button
										onclick={() => exportPng(f)}
										class="rounded border border-slate-200 px-2 py-1 text-xs font-medium text-slate-600 transition hover:bg-slate-100"
									>
										PNG
									</button>
								</div>
							</div>
							<!-- eslint-disable-next-line svelte/no-at-html-tags -->
							<div class="schema grid place-items-center overflow-auto p-3">{@html f.svg}</div>
						</div>
					{/each}
				{:else}
					<p class="grid h-full place-items-center text-sm text-slate-400">
						Встав код і натисни «Побудувати схему»
					</p>
				{/if}
			</div>
		</div>
	</div>

	<p class="mt-3 text-xs {errored ? 'text-red-600' : 'text-slate-500'}">{status}</p>
</div>
