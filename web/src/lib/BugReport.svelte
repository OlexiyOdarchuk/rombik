<script>
	// Модалка звороту: збирає нік у Telegram + опис помилки (+ необов'язково код) і
	// шле POST на той самий Cloudflare-воркер, що форвардить у Telegram. Без GitHub —
	// студентам зручніше. Воркер розрізняє звернення за полем `type`.
	import { bugOpen, closeBug } from '$lib/bug.svelte.js';

	const ENDPOINT = 'https://s-uah-miner-telemetry.ishawyha.workers.dev';
	const TG = 'https://t.me/NeShawyha';

	let tg = $state('');
	let message = $state('');
	let code = $state('');
	let company = $state(''); // honeypot — боти заповнюють, люди ні
	let status = $state('idle'); // idle | sending | success | error

	const valid = $derived(tg.trim().length > 0 && message.trim().length > 0);

	async function submit(e) {
		e.preventDefault();
		if (!valid || status === 'sending') return;
		status = 'sending';
		try {
			const res = await fetch(ENDPOINT, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					type: 'rombik-bug',
					contact: tg.trim(),
					message: message.trim(),
					code: code.trim(),
					page: typeof location !== 'undefined' ? location.pathname : '',
					ua: typeof navigator !== 'undefined' ? navigator.userAgent : '',
					company
				})
			});
			if (!res.ok) throw new Error(String(res.status));
			status = 'success';
			tg = message = code = '';
		} catch {
			status = 'error';
		}
	}

	function close() {
		status = 'idle';
		closeBug();
	}
	function onKey(e) {
		if (e.key === 'Escape') close();
	}
</script>

<svelte:window onkeydown={bugOpen() ? onKey : undefined} />

{#if bugOpen()}
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
	<div class="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-sm" onclick={close}>
		<div
			class="w-full max-w-lg rounded-2xl border border-slate-200 bg-white p-6 shadow-xl dark:border-slate-700 dark:bg-slate-900"
			onclick={(e) => e.stopPropagation()}
			role="dialog"
			aria-modal="true"
			tabindex="-1"
		>
			<div class="flex items-start justify-between gap-4">
				<div>
					<h2 class="text-lg font-bold text-slate-900 dark:text-slate-100">Повідомити про помилку</h2>
					<p class="mt-1 text-sm text-slate-500 dark:text-slate-400">Схема намалювалась не так? Напиши — і я полагоджу.</p>
				</div>
				<button onclick={close} aria-label="Закрити" class="rounded-lg p-1.5 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12" /></svg>
				</button>
			</div>

			{#if status === 'success'}
				<div class="mt-6 rounded-xl bg-green-50 p-5 text-center dark:bg-green-950/40">
					<p class="font-semibold text-green-700 dark:text-green-400">Дякую! 🙌</p>
					<p class="mt-1 text-sm text-slate-600 dark:text-slate-300">Я напишу тобі в Telegram, щойно гляну.</p>
					<button onclick={close} class="mt-4 rounded-lg bg-blue-600 px-5 py-2 text-sm font-semibold text-white transition hover:bg-blue-700">Закрити</button>
				</div>
			{:else}
				<form onsubmit={submit} class="mt-5 space-y-4">
					<label class="block">
						<span class="text-sm font-medium text-slate-700 dark:text-slate-200">Твій нік у Telegram <span class="text-red-500">*</span></span>
						<input
							bind:value={tg}
							placeholder="@username"
							class="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 outline-none transition focus:border-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100"
						/>
						<span class="mt-1 block text-xs text-slate-400">Щоб я міг відповісти й уточнити.</span>
					</label>

					<label class="block">
						<span class="text-sm font-medium text-slate-700 dark:text-slate-200">Що сталося <span class="text-red-500">*</span></span>
						<textarea
							bind:value={message}
							rows="3"
							placeholder="Напр.: switch на 3 case намалював лише перший…"
							class="mt-1 w-full resize-y rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 outline-none transition focus:border-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100"
						></textarea>
					</label>

					<label class="block">
						<span class="text-sm font-medium text-slate-700 dark:text-slate-200">Код, що зламався <span class="font-normal text-slate-400">(необов'язково, дуже допомагає)</span></span>
						<textarea
							bind:value={code}
							rows="4"
							placeholder="Встав сюди шматок коду…"
							class="mt-1 w-full resize-y rounded-lg border border-slate-300 px-3 py-2 font-mono text-xs text-slate-900 outline-none transition focus:border-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100"
						></textarea>
					</label>

					<!-- honeypot: приховано від людей, видно ботам -->
					<input bind:value={company} tabindex="-1" autocomplete="off" class="absolute -left-[9999px]" aria-hidden="true" />

					{#if status === 'error'}
						<p class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950/40 dark:text-red-400">
							Не вдалося надіслати. Напиши напряму:
							<a href={TG} target="_blank" rel="noopener" class="font-semibold underline">@NeShawyha</a>
						</p>
					{/if}

					<div class="flex items-center justify-between gap-3 pt-1">
						<a href={TG} target="_blank" rel="noopener" class="text-sm text-slate-500 hover:text-slate-700 hover:underline dark:text-slate-400">або написати в Telegram →</a>
						<button
							type="submit"
							disabled={!valid || status === 'sending'}
							class="rounded-lg bg-blue-600 px-5 py-2 text-sm font-semibold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
						>
							{status === 'sending' ? 'Надсилаю…' : 'Надіслати'}
						</button>
					</div>
				</form>
			{/if}
		</div>
	</div>
{/if}
