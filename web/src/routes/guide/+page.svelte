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
	<h1 class="text-4xl font-bold tracking-tight text-slate-900 dark:text-slate-100">Як це працює</h1>
	<p class="mt-4 text-lg text-slate-600 dark:text-slate-300">
		rombik розбирає твій код справжнім парсером Python, будує логічне дерево алгоритму й
		розкладає його у блок-схему за ДСТУ 19.701-90. Усе локально в браузері — код нікуди не
		надсилається.
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
	<ul class="mt-6 grid gap-2 sm:grid-cols-2">
		{#each supported as item (item)}
			<li class="flex items-start gap-2 text-slate-700 dark:text-slate-300">
				<svg class="mt-1 shrink-0 text-blue-600" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
					<path d="M5 13l4 4L19 7" />
				</svg>
				{item}
			</li>
		{/each}
	</ul>

	<h2 class="mt-14 text-2xl font-bold text-slate-900 dark:text-slate-100">Інтерактивний візуальний редактор</h2>
	<div class="mt-6 rounded-2xl border border-blue-200 bg-blue-50 p-6 dark:border-blue-900 dark:bg-blue-950/40">
		<p class="text-slate-600 dark:text-slate-300">
			Окрім автоматичної генерації, у rombik є повноцінний <strong>візуальний редактор</strong>. 
			Ви можете:
		</p>
		<ul class="mt-3 list-inside list-disc space-y-1 text-slate-600 dark:text-slate-300">
			<li>Вручну перетягувати блоки (воВони примагнічуються до сітки)</li>
			<li>Змінювати маршрут стрілок або малювати власні з'єднання</li>
			<li>Редагувати текст усередині фігур подвійним кліком</li>
			<li><strong>Розбивати велику схему</strong> на сторінки за допомогою спеціальних конекторів (кружечки з літерами А, Б, В...)</li>
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

	<h2 class="mt-14 text-2xl font-bold text-slate-900 dark:text-slate-100">Найчастіші запитання (FAQ)</h2>
	<div class="mt-6 space-y-4">
		<details class="group rounded-xl border border-slate-200 bg-white p-5 open:border-blue-300 dark:border-slate-800 dark:bg-slate-900 dark:open:border-blue-500">
			<summary class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100">
				Чи потрібне підключення до Інтернету?
				<span class="text-blue-500 transition-transform group-open:rotate-45">+</span>
			</summary>
			<p class="mt-3 text-slate-600 dark:text-slate-300">
				Тільки для завантаження самої сторінки (близько 2 МБ). Далі рушій працює 100% офлайн у вашому браузері за допомогою технології WebAssembly. Ваші вихідні коди нікуди не відправляються.
			</p>
		</details>
		<details class="group rounded-xl border border-slate-200 bg-white p-5 open:border-blue-300 dark:border-slate-800 dark:bg-slate-900 dark:open:border-blue-500">
			<summary class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100">
				Що робити, якщо схема занадто довга для формату А4?
				<span class="text-blue-500 transition-transform group-open:rotate-45">+</span>
			</summary>
			<p class="mt-3 text-slate-600 dark:text-slate-300">
				Є два варіанти. Перший — скористайтеся автоматичним розбиттям: натисніть кнопку «✂ Розбити» під згенерованою схемою. Другий — відкрийте візуальний редактор (кнопка «✎ Редагувати»), виділіть блок, з якого хочете перенести частину схеми, і оберіть «✂ Розділити». Система сама вставить конектори з літерами.
			</p>
		</details>
		<details class="group rounded-xl border border-slate-200 bg-white p-5 open:border-blue-300 dark:border-slate-800 dark:bg-slate-900 dark:open:border-blue-500">
			<summary class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100">
				Як скопіювати схему в Word або Google Docs без втрати якості?
				<span class="text-blue-500 transition-transform group-open:rotate-45">+</span>
			</summary>
			<p class="mt-3 text-slate-600 dark:text-slate-300">
				Натисніть кнопку "Копіювати ▾" і оберіть <strong>PNG</strong>. Також ви можете експортувати схему у форматі <strong>SVG</strong> (Вставка → Малюнок) для максимально чіткого векторного відображення без пікселів.
			</p>
		</details>
		<details class="group rounded-xl border border-slate-200 bg-white p-5 open:border-blue-300 dark:border-slate-800 dark:bg-slate-900 dark:open:border-blue-500">
			<summary class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100">
				Я хочу використовувати українські слова (Так/Ні, Ввід/Вивід). Як це зробити?
				<span class="text-blue-500 transition-transform group-open:rotate-45">+</span>
			</summary>
			<p class="mt-3 text-slate-600 dark:text-slate-300">
				Всі ці речі можна налаштувати! Відкрийте меню "Налаштування" в панелі інструментів. Перейдіть на вкладку "Текст та мітки", де ви можете змінити слова "Yes", "No", "Input", "Print", "True", "False" на будь-які свої, включно з українськими відповідниками.
			</p>
		</details>
	</div>

	<div class="mt-12 text-center">
		<a href="/app" class="inline-block rounded-xl bg-blue-600 px-7 py-3 font-semibold text-white transition hover:bg-blue-700">
			Спробувати редактор →
		</a>
	</div>
</section>
