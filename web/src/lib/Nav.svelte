<script>
	import { page } from '$app/state';
	import ThemeToggle from '$lib/ThemeToggle.svelte';

	const links = [
		{ href: '/', label: 'Головна' },
		{ href: '/guide', label: 'Як це працює' },
		{ href: '/app', label: 'Редактор' }
	];
	let menuOpen = $state(false);
</script>

<header class="sticky top-0 z-40 border-b border-slate-200/70 bg-paper/80 backdrop-blur dark:border-slate-800/70">
	<nav class="mx-auto flex h-16 max-w-6xl items-center justify-between px-5">
		<a href="/" class="flex items-center gap-2 font-semibold text-slate-900 dark:text-slate-100">
			<span class="grid h-8 w-8 place-items-center rounded-lg bg-blue-600 text-white shadow-sm">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round">
					<path d="M12 5 L19 12 L12 19 L5 12 Z" />
					<path d="M12 2 V5" /><path d="M5 12 H2" /><path d="M19 12 H22" />
				</svg>
			</span>
			<span>rombik<span class="text-blue-600">.</span></span>
		</a>

		<div class="hidden items-center gap-1 md:flex">
			{#each links as l (l.href)}
				<a
					href={l.href}
					class="rounded-lg px-3 py-2 text-sm font-medium transition hover:bg-slate-100 dark:hover:bg-slate-800
						{page.url.pathname === l.href
						? 'text-blue-700 dark:text-blue-400'
						: 'text-slate-600 dark:text-slate-300'}"
				>
					{l.label}
				</a>
			{/each}
			<ThemeToggle />
			<a
				href="/app"
				class="ml-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700"
			>
				Спробувати
			</a>
		</div>

		<div class="flex items-center gap-2 md:hidden">
			<ThemeToggle />
			<button onclick={() => (menuOpen = !menuOpen)} class="p-2 text-slate-600 dark:text-slate-300" aria-label="Відкрити меню">
				<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d={menuOpen ? "M18 6L6 18M6 6l12 12" : "M3 12h18M3 6h18M3 18h18"} />
				</svg>
			</button>
		</div>
	</nav>

	{#if menuOpen}
		<div class="border-t border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-900 md:hidden">
			<div class="flex flex-col gap-2">
				{#each links as l (l.href)}
					<a
						href={l.href}
						onclick={() => (menuOpen = false)}
						class="rounded-lg px-4 py-3 text-base font-medium transition hover:bg-slate-100 dark:hover:bg-slate-800
							{page.url.pathname === l.href
							? 'text-blue-700 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20'
							: 'text-slate-600 dark:text-slate-300'}"
					>
						{l.label}
					</a>
				{/each}
				<a
					href="/app"
					onclick={() => (menuOpen = false)}
					class="mt-2 rounded-lg bg-blue-600 px-4 py-3 text-center text-base font-semibold text-white shadow-sm transition hover:bg-blue-700"
				>
					Спробувати
				</a>
			</div>
		</div>
	{/if}
</header>
