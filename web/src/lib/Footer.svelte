<script>
	const hash = typeof __COMMIT_HASH__ !== 'undefined' ? __COMMIT_HASH__ : 'dev';
	const date = typeof __COMMIT_DATE__ !== 'undefined' ? __COMMIT_DATE__ : 'today';
</script>

<footer class="border-t border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
	<div class="mx-auto flex max-w-6xl flex-col gap-4 px-5 py-8 text-sm text-slate-500 dark:text-slate-400 sm:flex-row sm:items-center sm:justify-between">
		<div class="flex flex-col gap-1">
			<p>
				© {new Date().getFullYear()} rombik — блок-схеми за ДСТУ 19.701-90. Зробив
				<a href="https://ishawyha.dev" class="font-medium text-blue-600 hover:underline dark:text-blue-400">Олексій Одарчук</a>.
			</p>
			<p class="text-xs text-slate-400 dark:text-slate-500">
				Останнє оновлення: {date} (commit <a href="https://github.com/OlexiyOdarchuk/rombik/commit/{hash}" target="_blank" rel="noopener" class="font-mono hover:text-slate-600 hover:underline dark:hover:text-slate-300">{hash}</a>)
			</p>
		</div>
		<div class="flex gap-4">
			<a href="/guide" class="hover:text-slate-800 dark:hover:text-slate-200">Як це працює</a>
			<a href="/app" class="hover:text-slate-800 dark:hover:text-slate-200">Редактор</a>
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
