<script>
	const shapes = [
		{ name: 'Термінатор', use: 'Початок / Кінець', svg: 'terminator' },
		{ name: 'Процес', use: 'Дія, обчислення, присвоєння', svg: 'process' },
		{ name: 'Розв’язок', use: 'Умова (if / while)', svg: 'decision' },
		{ name: 'Ввід-вивід', use: 'input / print, cin / cout', svg: 'io' },
		{ name: 'Межа циклу', use: 'Цикл for з лічильником', svg: 'loop' },
		{ name: 'Підпрограма', use: 'Виклик власної функції', svg: 'subprogram' }
	];

	const supported = [
		'Послідовність, присвоєння, вирази',
		'Розгалуження if / elif / else (будь-яка вкладеність)',
		'Цикл for (шестикутник «i = 0, n-1, 1»)',
		'Цикл while (ромб-передумова)',
		'Цикл while True … break (ромб-післяумова)',
		'Ввід / вивід: input/print → «Ввід …» / «Вивід …»',
		'Функції → окрема схема на кожну; параметри → ввід; return → паралелограм',
		'Виклик власної функції → символ підпрограми'
	];
</script>

<svelte:head>
	<title>Як це працює — rombik</title>
</svelte:head>

<section class="mx-auto max-w-4xl px-5 py-16">
	<h1 class="text-4xl font-bold tracking-tight text-slate-900">Як це працює</h1>
	<p class="mt-4 text-lg text-slate-600">
		rombik розбирає твій код справжнім парсером Python, будує логічне дерево алгоритму й
		розкладає його у блок-схему за ДСТУ 19.701-90. Усе локально в браузері — код нікуди не
		надсилається.
	</p>

	<h2 class="mt-14 text-2xl font-bold text-slate-900">Фігури за стандартом</h2>
	<div class="mt-6 grid gap-4 sm:grid-cols-2">
		{#each shapes as s (s.name)}
			<div class="flex items-center gap-4 rounded-xl border border-slate-200 bg-white p-4">
				<svg viewBox="0 0 120 56" width="96" height="46" class="shrink-0" stroke="#334155" stroke-width="1.5" fill="#fff">
					{#if s.svg === 'terminator'}
						<rect x="6" y="14" width="108" height="28" rx="14" />
					{:else if s.svg === 'process'}
						<rect x="6" y="14" width="108" height="28" />
					{:else if s.svg === 'decision'}
						<polygon points="60,8 114,28 60,48 6,28" />
					{:else if s.svg === 'io'}
						<polygon points="22,14 114,14 98,42 6,42" />
					{:else if s.svg === 'loop'}
						<polygon points="20,14 100,14 114,28 100,42 20,42 6,28" />
					{:else if s.svg === 'subprogram'}
						<rect x="6" y="14" width="108" height="28" />
						<line x1="16" y1="14" x2="16" y2="42" /><line x1="104" y1="14" x2="104" y2="42" />
					{/if}
				</svg>
				<div>
					<p class="font-semibold text-slate-900">{s.name}</p>
					<p class="text-sm text-slate-500">{s.use}</p>
				</div>
			</div>
		{/each}
	</div>

	<h2 class="mt-14 text-2xl font-bold text-slate-900">Що вже підтримується</h2>
	<ul class="mt-6 grid gap-2 sm:grid-cols-2">
		{#each supported as item (item)}
			<li class="flex items-start gap-2 text-slate-700">
				<svg class="mt-1 shrink-0 text-blue-600" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
					<path d="M5 13l4 4L19 7" />
				</svg>
				{item}
			</li>
		{/each}
	</ul>

	<div class="mt-14 rounded-2xl border border-blue-200 bg-blue-50 p-6">
		<h3 class="text-lg font-semibold text-slate-900">Скоро: C++ та редактор</h3>
		<p class="mt-2 text-slate-600">
			Додаємо мову C++, ручне перетягування блоків і розбиття великої схеми на частини через
			конектори. Опції рендера (наприклад, виклик звичайним блоком замість символу
			підпрограми) — галочками просто в редакторі.
		</p>
	</div>

	<div class="mt-12 text-center">
		<a href="/app" class="inline-block rounded-xl bg-blue-600 px-7 py-3 font-semibold text-white transition hover:bg-blue-700">
			Спробувати редактор →
		</a>
	</div>
</section>
