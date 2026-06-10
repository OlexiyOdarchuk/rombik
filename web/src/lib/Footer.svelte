<script>
	import { base } from '$app/paths';
	const hash = typeof __COMMIT_HASH__ !== 'undefined' ? __COMMIT_HASH__ : 'dev';
	const date = typeof __COMMIT_DATE__ !== 'undefined' ? __COMMIT_DATE__ : 'today';
</script>

<footer class="border-t border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
	<div class="mx-auto flex max-w-6xl flex-col gap-4 px-5 py-8 text-sm text-slate-500 dark:text-slate-400 sm:flex-row sm:items-center sm:justify-between">
		<div class="flex flex-col gap-1">
			<p>
				© {new Date().getFullYear()} rombik — блок-схеми за ДСТУ 19.701-90. Зробив
				<a href="https://ishawyha.dev" class="font-medium text-blue-600 hover:underline dark:text-blue-400" target="_blank" rel="noopener">Олексій Одарчук (ishawyha)</a>.
			</p>
			<p class="text-xs text-slate-400 dark:text-slate-500">
				Останнє оновлення: {date} (commit <a href="https://github.com/OlexiyOdarchuk/rombik/commit/{hash}" target="_blank" rel="noopener" class="font-mono hover:text-slate-600 hover:underline dark:hover:text-slate-300">{hash}</a>)
			</p>
		</div>
		<div class="flex flex-wrap gap-4">
			<a href="{base}/guide" class="hover:text-slate-800 dark:hover:text-slate-200">Як це працює</a>
			<a href="{base}/app" class="hover:text-slate-800 dark:hover:text-slate-200">Редактор</a>
			<a
				href="https://github.com/OlexiyOdarchuk/rombik/issues/new"
				class="flex items-center gap-1 font-medium text-amber-600 hover:underline dark:text-amber-400"
				target="_blank"
				rel="noopener"
			>
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
				Повідомити про помилку
			</a>
			<a
				href="https://github.com/OlexiyOdarchuk/rombik"
				class="hover:text-slate-800 dark:hover:text-slate-200"
				target="_blank"
				rel="noopener"
			>
				GitHub
			</a>
		</div>
	</div>
</footer>
