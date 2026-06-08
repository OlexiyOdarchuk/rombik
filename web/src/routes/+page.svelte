<script>
	const features = [
		{
			title: 'Точно за ДСТУ 19.701-90',
			body: 'Овал, прямокутник, ромб, паралелограм, шестикутник циклу, символ підпрограми — усе як вимагає стандарт і викладачі.',
			icon: 'M3 7h18M3 12h18M3 17h18'
		},
		{
			title: 'Прямо з коду Python',
			body: 'Встав код — отримай схему. if/elif/else, for, while, while True/break, функції, ввід-вивід. Вкладеність будь-якої глибини.',
			icon: 'M8 9l-4 3 4 3M16 9l4 3-4 3'
		},
		{
			title: 'Редагуй як хочеш',
			body: 'Зміни текст у блоці, форму, перетягни блок, проведи стрілку. Розбий велику схему на частини конекторами.',
			icon: 'M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z'
		},
		{
			title: 'Усе в браузері',
			body: 'Нічого не встановлюєш, код нікуди не йде — розбір відбувається локально. Експорт у SVG та PNG.',
			icon: 'M3 12a9 9 0 1 0 18 0 9 9 0 0 0-18 0M3 12h18M12 3a15 15 0 0 1 0 18'
		}
	];

	const steps = [
		{ n: '1', t: 'Встав код', d: 'Свою функцію або цілу програму на Python.' },
		{ n: '2', t: 'Отримай схему', d: 'Двигун будує блок-схему за ДСТУ — миттєво.' },
		{ n: '3', t: 'Підправ і експортуй', d: 'Зміни що треба й завантаж SVG/PNG у свій документ.' }
	];
</script>

<svelte:head>
	<title>rombik — блок-схеми за ДСТУ з коду Python</title>
</svelte:head>

<!-- HERO -->
<section class="relative overflow-hidden border-b border-slate-200 grid-bg dark:border-slate-800 bg-gradient-to-b from-slate-50 to-white dark:from-slate-950 dark:to-slate-900">
	<div class="mx-auto grid max-w-6xl items-center gap-10 px-5 py-20 lg:grid-cols-2 lg:py-28">
		<div class="rise">
			<span class="inline-flex items-center gap-2 rounded-full border border-blue-200/60 bg-blue-50/80 px-4 py-1.5 text-xs font-bold uppercase tracking-wider text-blue-700 backdrop-blur-sm dark:border-blue-800/50 dark:bg-blue-900/30 dark:text-blue-400 shadow-sm">
				ДСТУ 19.701-90 · Локально у браузері
			</span>
			<h1 class="mt-6 text-4xl font-extrabold leading-tight tracking-tight text-slate-900 dark:text-white sm:text-5xl lg:text-6xl">
				Блок-схеми з коду <br/><span class="bg-gradient-to-r from-blue-600 to-indigo-500 bg-clip-text text-transparent">за лічені секунди</span>
			</h1>
			<p class="mt-6 max-w-lg text-lg text-slate-600 leading-relaxed dark:text-slate-300">
				Встав код на Python — отримай готову блок-схему за державним стандартом. Підправ
				у візуальному редакторі й додай у будь-який документ. Усе безкоштовно.
			</p>
			<div class="mt-10 flex flex-wrap gap-4">
				<a
					href="/app"
					class="group relative inline-flex items-center justify-center overflow-hidden rounded-xl bg-blue-600 px-8 py-3.5 font-semibold text-white transition-all hover:bg-blue-700 hover:shadow-lg hover:shadow-blue-500/30 hover:-translate-y-0.5"
				>
					<span class="mr-2">Побудувати схему</span>
					<svg class="h-4 w-4 transition-transform group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/></svg>
				</a>
				<a
					href="/guide"
					class="inline-flex items-center justify-center rounded-xl border border-slate-200 bg-white/50 px-8 py-3.5 font-semibold text-slate-700 backdrop-blur-sm transition-all hover:border-slate-300 hover:bg-slate-50 hover:shadow-sm dark:border-slate-700 dark:bg-slate-800/50 dark:text-slate-200 dark:hover:border-slate-600 dark:hover:bg-slate-800 hover:-translate-y-0.5"
				>
					Як це працює
				</a>
			</div>
		</div>

		<!-- міні-схема: справжній вивід двигуна -->
		<div class="rise relative group">
			<!-- Розмитий фоновий градієнт для свічення -->
			<div class="absolute -inset-1 rounded-3xl bg-gradient-to-r from-blue-600 to-indigo-500 opacity-20 blur transition duration-1000 group-hover:opacity-30 dark:opacity-40"></div>
			<div class="relative rounded-2xl border border-white/50 bg-white/80 p-8 shadow-2xl backdrop-blur-xl dark:border-slate-700/50 dark:bg-slate-900/80">
				<svg viewBox="0 0 472 564" class="mx-auto h-auto w-full max-w-[400px]" font-family="Outfit, sans-serif" font-size="14">
				<defs><marker id="arr" markerWidth="9" markerHeight="9" refX="7.5" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6 Z" fill="#222"/></marker></defs>
				<path d="M236.0 86.0 L236.0 130.0" fill="none" stroke="#222" stroke-width="1.5" marker-end="url(#arr)"/>
				<path d="M236.0 176.0 L236.0 220.0" fill="none" stroke="#222" stroke-width="1.5" marker-end="url(#arr)"/>
				<path d="M161.0 258.0 L121.8 258.0 L121.8 344.0" fill="none" stroke="#222" stroke-width="1.5" marker-end="url(#arr)"/>
				<text x="155.0" y="251.0" text-anchor="end" font-size="12" fill="#444">Так</text>
				<path d="M311.0 258.0 L341.8 258.0 L341.8 344.0" fill="none" stroke="#222" stroke-width="1.5" marker-end="url(#arr)"/>
				<text x="317.0" y="251.0" text-anchor="start" font-size="12" fill="#444">Ні</text>
				<path d="M121.8 390.0 L121.8 434.0 L236.0 434.0" fill="none" stroke="#222" stroke-width="1.5"/>
				<path d="M341.8 390.0 L341.8 434.0 L236.0 434.0" fill="none" stroke="#222" stroke-width="1.5"/>
				<path d="M236.0 434.0 L236.0 478.0" fill="none" stroke="#222" stroke-width="1.5" marker-end="url(#arr)"/>
				<rect x="171.0" y="44.0" width="130.0" height="42.0" rx="21.0" fill="#fdfdfd" stroke="#222" stroke-width="1.5"/>
				<text x="236.0" y="65.0" text-anchor="middle" dominant-baseline="middle" fill="#111">Початок</text>
				<polygon points="189.4,130.0 301.0,130.0 282.6,176.0 171.0,176.0" fill="#fdfdfd" stroke="#222" stroke-width="1.5"/>
				<text x="236.0" y="153.0" text-anchor="middle" dominant-baseline="middle" fill="#111">Ввід n</text>
				<polygon points="236.0,220.0 311.0,258.0 236.0,296.0 161.0,258.0" fill="#fdfdfd" stroke="#222" stroke-width="1.5"/>
				<text x="236.0" y="258.0" text-anchor="middle" dominant-baseline="middle" fill="#111">n &gt; 0</text>
				<polygon points="62.4,344.0 199.5,344.0 181.1,390.0 44.0,390.0" fill="#fdfdfd" stroke="#222" stroke-width="1.5"/>
				<text x="121.8" y="367.0" text-anchor="middle" dominant-baseline="middle" fill="#111">Вивід «Додатне»</text>
				<polygon points="273.9,344.0 428.0,344.0 409.6,390.0 255.5,390.0" fill="#fdfdfd" stroke="#222" stroke-width="1.5"/>
				<text x="341.8" y="367.0" text-anchor="middle" dominant-baseline="middle" fill="#111">Вивід «Недодатне»</text>
				<rect x="171.0" y="478.0" width="130.0" height="42.0" rx="21.0" fill="#fdfdfd" stroke="#222" stroke-width="1.5"/>
				<text x="236.0" y="499.0" text-anchor="middle" dominant-baseline="middle" fill="#111">Кінець</text>
			</svg>
		</div>
	</div>
</section>

<!-- ПРОБЛЕМА -->
<section class="mx-auto max-w-4xl px-5 py-20 text-center">
	<h2 class="text-3xl font-bold text-slate-900 dark:text-slate-100">Малювати схеми вручну — це години роботи</h2>
	<p class="mx-auto mt-4 max-w-2xl text-lg text-slate-600 dark:text-slate-300">
		Звіт, лабораторна, практична, курсова, диплом — будь-де, де є алгоритм, його треба
		зобразити схемою за ДСТУ. Лінійки, фігури, стрілки, переробки після кожної правки коду.
		rombik робить це за секунду й дозволяє підправити результат під будь-які вимоги.
	</p>
</section>

<!-- МОЖЛИВОСТІ -->
<section class="border-y border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
	<div class="mx-auto max-w-6xl px-5 py-20">
		<div class="grid gap-6 sm:grid-cols-2">
			{#each features as f (f.title)}
				<div class="group rounded-2xl border border-slate-200 bg-white/50 p-8 backdrop-blur-sm transition-all hover:border-blue-300 hover:bg-white hover:shadow-xl hover:shadow-blue-500/5 dark:border-slate-800 dark:bg-slate-900/50 dark:hover:border-blue-700/50 dark:hover:bg-slate-800 hover:-translate-y-1">
					<div class="grid h-14 w-14 place-items-center rounded-2xl bg-blue-50 text-blue-600 transition-colors group-hover:bg-blue-600 group-hover:text-white dark:bg-blue-900/40 dark:text-blue-400 dark:group-hover:bg-blue-600 dark:group-hover:text-white">
						<svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<path d={f.icon} />
						</svg>
					</div>
					<h3 class="mt-6 text-xl font-bold text-slate-900 dark:text-white">{f.title}</h3>
					<p class="mt-3 leading-relaxed text-slate-600 dark:text-slate-400">{f.body}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<!-- ЯК ПРАЦЮЄ -->
<section class="mx-auto max-w-6xl px-5 py-20">
	<h2 class="text-center text-3xl font-bold text-slate-900 dark:text-slate-100">Три кроки</h2>
	<div class="mt-12 grid gap-8 md:grid-cols-3">
		{#each steps as s (s.n)}
			<div class="relative rounded-2xl border border-slate-200 dark:border-slate-800 bg-white p-6 dark:bg-slate-800">
				<div class="grid h-10 w-10 place-items-center rounded-full bg-blue-600 font-bold text-white">
					{s.n}
				</div>
				<h3 class="mt-4 text-lg font-semibold text-slate-900 dark:text-slate-100">{s.t}</h3>
				<p class="mt-2 text-slate-600 dark:text-slate-300">{s.d}</p>
			</div>
		{/each}
	</div>
</section>

<!-- CTA -->
<section class="bg-slate-900">
	<div class="mx-auto max-w-4xl px-5 py-16 text-center">
		<h2 class="text-3xl font-bold text-white">Спробуй просто зараз</h2>
		<p class="mt-3 text-slate-300">Безкоштовно, у браузері, без реєстрації.</p>
		<a
			href="/app"
			class="mt-7 inline-block rounded-xl bg-blue-600 px-7 py-3 font-semibold text-white shadow-lg transition hover:bg-blue-500"
		>
			Відкрити редактор →
		</a>
	</div>
</section>
