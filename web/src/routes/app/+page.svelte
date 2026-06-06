<script>
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
	let opts = $state({ callAsProcess: false });
	let funcs = $state([]); // [{name, svg}]
	let active = $state(0);
	let status = $state('Готовий. Натисни «Побудувати».');
	let busy = $state(false);

	// Тимчасова заглушка прев'ю, поки підключаємо двигун (наступний крок).
	const PLACEHOLDER = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 320 120" font-family="Inter" font-size="13">
		<rect width="320" height="120" fill="#f8fafc"/>
		<text x="160" y="60" text-anchor="middle" fill="#94a3b8">Тут зʼявиться схема</text></svg>`;

	async function build() {
		busy = true;
		status = 'Будую…';
		try {
			// TODO: підключити engine.generate(code, opts) — Pyodide + flowgen.wasm
			await new Promise((r) => setTimeout(r, 150));
			funcs = [{ name: 'grade', svg: PLACEHOLDER }];
			active = 0;
			status = `Готово: ${funcs.length} схем(а). Двигун підключаємо в наступному кроці.`;
		} catch (e) {
			status = 'Помилка: ' + e;
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head>
	<title>Редактор — flowgen</title>
</svelte:head>

<div class="mx-auto flex h-[calc(100vh-4rem)] max-w-7xl flex-col px-4 py-4">
	<!-- toolbar -->
	<div class="mb-3 flex flex-wrap items-center gap-3">
		<button
			onclick={build}
			disabled={busy}
			class="rounded-lg bg-blue-600 px-5 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700 disabled:opacity-50"
		>
			{busy ? 'Будую…' : 'Побудувати схему'}
		</button>

		<label class="flex items-center gap-2 text-sm text-slate-600">
			<input type="checkbox" bind:checked={opts.callAsProcess} class="rounded border-slate-300" />
			Виклик звичайним блоком (не ДСТУ-підпрограмою)
		</label>

		<div class="ml-auto flex gap-2">
			<button class="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-400" disabled>
				Експорт SVG
			</button>
			<button class="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-400" disabled>
				Експорт PNG
			</button>
		</div>
	</div>

	<!-- split -->
	<div class="grid min-h-0 flex-1 gap-3 lg:grid-cols-2">
		<!-- code -->
		<div class="flex min-h-0 flex-col rounded-xl border border-slate-200 bg-white">
			<div class="border-b border-slate-200 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-slate-500">
				Код Python
			</div>
			<textarea
				bind:value={code}
				spellcheck="false"
				class="min-h-0 flex-1 resize-none rounded-b-xl bg-white p-4 font-mono text-sm text-slate-800 outline-none"
			></textarea>
		</div>

		<!-- preview -->
		<div class="flex min-h-0 flex-col rounded-xl border border-slate-200 bg-white">
			<div class="flex items-center gap-1 border-b border-slate-200 px-3 py-2">
				{#if funcs.length}
					{#each funcs as f, i (f.name)}
						<button
							onclick={() => (active = i)}
							class="rounded-md px-3 py-1 text-sm font-medium transition
								{active === i ? 'bg-blue-50 text-blue-700' : 'text-slate-500 hover:bg-slate-100'}"
						>
							{f.name}
						</button>
					{/each}
				{:else}
					<span class="px-1 text-xs font-semibold uppercase tracking-wide text-slate-500">Схема</span>
				{/if}
			</div>
			<div class="grid min-h-0 flex-1 place-items-center overflow-auto grid-bg p-4">
				{#if funcs.length}
					<!-- eslint-disable-next-line svelte/no-at-html-tags -->
					{@html funcs[active].svg}
				{:else}
					<p class="text-sm text-slate-400">Встав код і натисни «Побудувати схему»</p>
				{/if}
			</div>
		</div>
	</div>

	<p class="mt-3 text-xs text-slate-500">{status}</p>
</div>
