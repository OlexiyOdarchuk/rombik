<script>
	import { page } from '$app/state';

	const links = [
		{ href: '/', label: 'Головна' },
		{ href: '/guide', label: 'Як це працює' },
		{ href: '/app', label: 'Редактор' }
	];
</script>

<header class="sticky top-0 z-40 border-b border-slate-200/70 bg-paper/80 backdrop-blur">
	<nav class="mx-auto flex h-16 max-w-6xl items-center justify-between px-5">
		<a href="/" class="flex items-center gap-2 font-semibold text-slate-900">
			<span class="grid h-8 w-8 place-items-center rounded-lg bg-blue-600 text-white shadow-sm">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round">
					<path d="M12 5 L19 12 L12 19 L5 12 Z" />
					<path d="M12 2 V5" /><path d="M12 19 V22" />
				</svg>
			</span>
			<span>rombik<span class="text-blue-600">.</span></span>
		</a>

		<div class="flex items-center gap-1">
			{#each links as l (l.href)}
				<a
					href={l.href}
					class="rounded-lg px-3 py-2 text-sm font-medium transition hover:bg-slate-100
						{page.url.pathname === l.href ? 'text-blue-700' : 'text-slate-600'}"
				>
					{l.label}
				</a>
			{/each}
			<a
				href="/app"
				class="ml-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700"
			>
				Спробувати
			</a>
		</div>
	</nav>
</header>
