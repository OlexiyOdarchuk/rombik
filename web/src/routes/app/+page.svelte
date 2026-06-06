<script>
	import { generate, warmup } from '$lib/engine.js';
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
	let opts = $state({ callAsProcess: false });
	let funcs = $state([]); // [{name, svg, diagram}]
	let status = $state('Готовий. Натисни «Побудувати».');
	let busy = $state(false);
	let errored = $state(false);

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
			const res = await generate(code, opts, (s) => (status = s));
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
	<div class="mb-3 flex flex-wrap items-center gap-3">
		<button
			onclick={build}
			disabled={busy}
			class="rounded-lg bg-blue-600 px-5 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700 disabled:opacity-50"
		>
			{busy ? 'Будую…' : 'Побудувати схему'}
		</button>

		<label class="flex items-center gap-2 text-sm text-slate-600">
			<input
				type="checkbox"
				bind:checked={opts.callAsProcess}
				onchange={() => funcs.length && build()}
				class="rounded border-slate-300"
			/>
			Виклик звичайним блоком (не ДСТУ-підпрограмою)
		</label>

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
			<textarea
				bind:value={code}
				spellcheck="false"
				class="min-h-0 flex-1 resize-none rounded-b-xl bg-white p-4 font-mono text-sm text-slate-800 outline-none"
			></textarea>
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
