<script>
	import { base } from '$app/paths';
	import { openBug } from '$lib/bug.svelte.js';
	const shapes = [
		{ name: 'Термінатор', use: 'Початок / Кінець', svg: 'terminator' },
		{ name: 'Процес', use: 'Дія, обчислення, присвоєння', svg: 'process' },
		{ name: 'Розв’язок', use: 'Умова (if / while)', svg: 'decision' },
		{ name: 'Ввід-вивід', use: 'input / print, cin / cout', svg: 'io' },
		{ name: 'Межа циклу', use: 'Цикл for з лічильником', svg: 'loop' },
		{ name: 'Підпрограма', use: 'Виклик власної функції', svg: 'subprogram' }
	];

	const pythonFeatures = [
		'Послідовність, присвоєння, вирази',
		'if / elif / else (будь-яка вкладеність)',
		'match / case',
		'for … in range (шестикутник «i = 0, n−1, 1»)',
		'while (передумова), for/else, while/else',
		'while True … break (післяумова), нескінченний цикл',
		'break / continue / return / raise / exit',
		'try / except / finally, with',
		'input / print → «Ввід …» / «Вивід …»',
		'Виклик власної функції → підпрограма'
	];
	const cppFeatures = [
		'if / else, тернарник, складені умови',
		'switch — усі case + default; fallthrough (case 1: case 2:) → одна умова через «||»',
		'for (класичний) і range-based «for (int x : arr)»',
		'while, do / while, вкладені цикли',
		'break / continue / return, рекурсія',
		'cin / cout → «Ввід …» / «Вивід …»',
		'try / catch',
		'Методи класів і структур → окрема схема «Клас::метод»'
	];
</script>

<svelte:head>
	<title>Як це працює — rombik</title>
</svelte:head>

<section class="mx-auto max-w-4xl px-5 py-16">
	<h1 class="text-4xl font-bold tracking-tight text-slate-900 dark:text-slate-100">Як це працює</h1>
	<p class="mt-4 text-lg text-slate-600 dark:text-slate-300">
		rombik розбирає твій код справжнім парсером (<strong>Python</strong> і <strong>C++</strong>),
		будує логічне дерево алгоритму й розкладає його у блок-схему за ДСТУ 19.701-90. Усе
		локально в браузері — код нікуди не надсилається.
	</p>

	<h2 class="mt-14 text-2xl font-bold text-slate-900 dark:text-slate-100">Фігури за стандартом</h2>
	<div class="mt-6 grid gap-4 sm:grid-cols-2">
		{#each shapes as s (s.name)}
			<div class="flex items-center gap-4 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-800 dark:bg-slate-800">
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
					<p class="font-semibold text-slate-900 dark:text-slate-100">{s.name}</p>
					<p class="text-sm text-slate-500 dark:text-slate-400">{s.use}</p>
				</div>
			</div>
		{/each}
	</div>

	<h2 class="mt-14 text-2xl font-bold text-slate-900 dark:text-slate-100">Що вже підтримується</h2>
	<p class="mt-3 text-slate-600 dark:text-slate-300">
		Кожна функція стає окремою схемою; її параметри малюються вхідним паралелограмом,
		а <code>return</code> — паралелограмом виходу.
	</p>
	<div class="mt-6 grid gap-6 md:grid-cols-2">
		{#each [['Python', pythonFeatures], ['C++', cppFeatures]] as [lang, items] (lang)}
			<div class="rounded-2xl border border-slate-200 bg-white p-5 dark:border-slate-800 dark:bg-slate-800">
				<h3 class="mb-3 flex items-center gap-2 text-lg font-bold text-slate-900 dark:text-slate-100">
					<span class="rounded-md bg-blue-600 px-2 py-0.5 text-sm text-white">{lang}</span>
					<span class="text-sm font-normal text-slate-500 dark:text-slate-400">повна підтримка</span>
				</h3>
				<ul class="space-y-2">
					{#each items as item (item)}
						<li class="flex items-start gap-2 text-sm text-slate-700 dark:text-slate-300">
							<svg class="mt-0.5 shrink-0 text-blue-600" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
								<path d="M5 13l4 4L19 7" />
							</svg>
							{item}
						</li>
					{/each}
				</ul>
			</div>
		{/each}
	</div>

	<h2 class="mt-14 text-2xl font-bold text-slate-900 dark:text-slate-100">Інтерактивний візуальний редактор</h2>
	<div class="mt-6 rounded-2xl border border-blue-200 bg-blue-50 p-6 dark:border-blue-900 dark:bg-blue-950/40">
		<p class="text-slate-600 dark:text-slate-300">
			Окрім автоматичної генерації, у rombik є повноцінний <strong>візуальний редактор</strong>
			(кнопка «✎ Редагувати» на картці схеми). Ви можете:
		</p>
		<ul class="mt-3 list-inside list-disc space-y-1 text-slate-600 dark:text-slate-300">
			<li>Перетягувати блоки (примагнічуються до сітки), рухати цілу виділену групу разом</li>
			<li>Змінювати маршрут стрілок або малювати власні з'єднання; правити текст подвійним кліком</li>
			<li><strong>Ділити на схеми вручну:</strong> виділяй блоки рамкою-ласо (Shift+тяг фону) чи Shift-кліком, тоді <strong>«⊞ Окрема схема»</strong> виносить їх в окремий рисунок (рамкою)</li>
			<li><strong>«⊟ Об'єднати»</strong> — зливає кілька схем назад в одну</li>
			<li>Стрілки між схемами автоматично стають <strong>парами конекторів</strong> (кружечки А, Б, В…)</li>
			<li><strong>«Редагувати всі»</strong> — усі функції коду на одному полотні, щоб ділити блоки між ними</li>
			<li>Undo / redo, безмежне полотно з пан/зумом</li>
		</ul>
	</div>

	<h2 class="mt-14 text-2xl font-bold text-slate-900 dark:text-slate-100">Експорт та формати</h2>
	<p class="mt-4 text-slate-600 dark:text-slate-300">
		Після побудови схеми, ви можете зберегти її у різних форматах:
	</p>
	<ul class="mt-4 grid gap-2 sm:grid-cols-2">
		<li class="flex items-start gap-2 text-slate-700 dark:text-slate-300"><strong class="w-24 text-blue-600">SVG / PNG</strong> Ідеально для вставки в Word, Google Docs чи презентації</li>
		<li class="flex items-start gap-2 text-slate-700 dark:text-slate-300"><strong class="w-24 text-blue-600">PDF</strong> Векторний формат для друку або курсових робіт</li>
		<li class="flex items-start gap-2 text-slate-700 dark:text-slate-300"><strong class="w-24 text-blue-600">Typst</strong> Нативний код для сучасних систем верстки наукових робіт</li>
		<li class="flex items-start gap-2 text-slate-700 dark:text-slate-300"><strong class="w-24 text-blue-600">Excalidraw</strong> Відкривайте схему в Excalidraw для подальшого рукописного редагування</li>
	</ul>



	<div class="mt-12 text-center">
		<a href="{base}/app" class="inline-block rounded-xl bg-blue-600 px-7 py-3 font-semibold text-white transition hover:bg-blue-700">
			Спробувати редактор →
		</a>
	</div>

	<div class="mt-12 rounded-2xl border border-amber-200 bg-amber-50 p-6 text-center dark:border-amber-900/50 dark:bg-amber-950/30">
		<p class="font-semibold text-slate-900 dark:text-slate-100">Схема побудувалась не так? Є ідея?</p>
		<p class="mt-1 text-sm text-slate-600 dark:text-slate-300">
			Якщо якась конструкція намалювалась неправильно — напиши, додай шматок коду, і я полагоджу.
		</p>
		<button
			onclick={openBug}
			class="mt-4 inline-flex items-center gap-2 rounded-xl border border-amber-300 bg-white px-5 py-2.5 font-semibold text-amber-700 transition hover:bg-amber-100 dark:border-amber-800 dark:bg-slate-900 dark:text-amber-400 dark:hover:bg-slate-800"
		>
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
			Повідомити про помилку
		</button>
	</div>
</section>
