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
	<title>flowgen — блок-схеми за ДСТУ з коду Python</title>
</svelte:head>

<!-- HERO -->
<section class="relative overflow-hidden border-b border-slate-200 grid-bg">
	<div class="mx-auto grid max-w-6xl items-center gap-10 px-5 py-20 lg:grid-cols-2 lg:py-28">
		<div class="rise">
			<span class="inline-flex items-center gap-2 rounded-full border border-blue-200 bg-blue-50 px-3 py-1 text-xs font-semibold text-blue-700">
				ДСТУ 19.701-90 · без реєстрації
			</span>
			<h1 class="mt-5 text-4xl font-bold leading-tight tracking-tight text-slate-900 sm:text-5xl">
				Блок-схеми з коду <span class="text-blue-600">за лічені секунди</span>
			</h1>
			<p class="mt-5 max-w-md text-lg text-slate-600">
				Встав код на Python — отримай готову блок-схему за державним стандартом. Підправ
				у візуальному редакторі й додай у будь-який документ.
			</p>
			<div class="mt-8 flex flex-wrap gap-3">
				<a
					href="/app"
					class="rounded-xl bg-blue-600 px-6 py-3 font-semibold text-white shadow-sm transition hover:bg-blue-700"
				>
					Побудувати схему →
				</a>
				<a
					href="/guide"
					class="rounded-xl border border-slate-300 bg-white px-6 py-3 font-semibold text-slate-700 transition hover:border-slate-400"
				>
					Як це працює
				</a>
			</div>
		</div>

		<!-- міні-схема: справжній вивід двигуна -->
		<div class="rise rounded-2xl border border-slate-200 bg-white p-6 shadow-xl shadow-slate-200/50">
			<svg viewBox="0 0 472 564" class="mx-auto h-auto w-full max-w-[240px]" font-family="Inter, sans-serif" font-size="14">
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
	<h2 class="text-3xl font-bold text-slate-900">Малювати схеми вручну — це години роботи</h2>
	<p class="mx-auto mt-4 max-w-2xl text-lg text-slate-600">
		Звіт, лабораторна, практична, курсова, диплом — будь-де, де є алгоритм, його треба
		зобразити схемою за ДСТУ. Лінійки, фігури, стрілки, переробки після кожної правки коду.
		flowgen робить це за секунду й дозволяє підправити результат під будь-які вимоги.
	</p>
</section>

<!-- МОЖЛИВОСТІ -->
<section class="border-y border-slate-200 bg-white">
	<div class="mx-auto max-w-6xl px-5 py-20">
		<div class="grid gap-6 sm:grid-cols-2">
			{#each features as f (f.title)}
				<div class="rounded-2xl border border-slate-200 p-6 transition hover:border-blue-300 hover:shadow-md">
					<div class="grid h-11 w-11 place-items-center rounded-xl bg-blue-50 text-blue-600">
						<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
							<path d={f.icon} />
						</svg>
					</div>
					<h3 class="mt-4 text-lg font-semibold text-slate-900">{f.title}</h3>
					<p class="mt-2 text-slate-600">{f.body}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<!-- ЯК ПРАЦЮЄ -->
<section class="mx-auto max-w-6xl px-5 py-20">
	<h2 class="text-center text-3xl font-bold text-slate-900">Три кроки</h2>
	<div class="mt-12 grid gap-8 md:grid-cols-3">
		{#each steps as s (s.n)}
			<div class="relative rounded-2xl border border-slate-200 bg-white p-6">
				<div class="grid h-10 w-10 place-items-center rounded-full bg-blue-600 font-bold text-white">
					{s.n}
				</div>
				<h3 class="mt-4 text-lg font-semibold text-slate-900">{s.t}</h3>
				<p class="mt-2 text-slate-600">{s.d}</p>
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
