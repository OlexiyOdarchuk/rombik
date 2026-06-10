<script>
    import { base } from "$app/paths";
    import { openBug } from "$lib/bug.svelte.js";
    const features = [
        {
            title: "Точно за ДСТУ 19.701-90",
            body: "Овал, прямокутник, ромб, паралелограм, шестикутник циклу, символ підпрограми — усе як вимагає стандарт і викладачі.",
            icon: "M3 7h18M3 12h18M3 17h18",
        },
        {
            title: "Python, C, C++ та Pascal",
            body: "Встав код — отримай схему. if/else, switch, цикли, функції, класи, ввід-вивід. Вкладеність будь-якої глибини.",
            icon: "M8 9l-4 3 4 3M16 9l4 3-4 3",
        },
        {
            title: "Редагуй як хочеш",
            body: "Зміни текст і форму блоку, перетягни його, проведи стрілку. Виділяй блоки рамкою-ласо — ділиш схему на частини або об'єднуєш назад.",
            icon: "M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z",
        },
        {
            title: "Усе в браузері",
            body: "Нічого не встановлюєш, код нікуди не йде — усе локально. Експорт у SVG, PNG, PDF, Typst і Excalidraw.",
            icon: "M3 12a9 9 0 1 0 18 0 9 9 0 0 0-18 0M3 12h18M12 3a15 15 0 0 1 0 18",
        },
    ];

    const steps = [
        {
            n: "1",
            t: "Встав код",
            d: "Функцію чи цілу програму — Python, C, C++ або Pascal.",
        },
        {
            n: "2",
            t: "Отримай схему",
            d: "Рушій будує блок-схему за ДСТУ — миттєво.",
        },
        {
            n: "3",
            t: "Підправ і експортуй",
            d: "Зміни що треба й завантаж SVG/PNG у свій документ.",
        },
    ];
</script>

<svelte:head>
    <title>rombik — автоматичні блок-схеми за ДСТУ з коду Python, C, C++, Pascal</title>
    <meta
        name="description"
        content="Генератор блок-схем онлайн. Встав код на Python, C, C++ чи Pascal — отримай готову блок-схему за ДСТУ 19.701-90. Експорт у SVG, PNG, PDF, Typst та Excalidraw."
    />
    <meta
        name="keywords"
        content="блок-схема, алгоритм, ДСТУ 19.701-90, python блок-схема, c++ блок-схема, pascal блок-схема, c блок-схема, генератор блок-схем, візуалізація алгоритмів, курсова, диплом"
    />
    <meta property="og:title" content="rombik — блок-схеми за секунду" />
    <meta
        property="og:description"
        content="Безкоштовний генератор правильних блок-схем з коду Python, C, C++ і Pascal прямо у вашому браузері."
    />
    <meta property="og:type" content="website" />
</svelte:head>

<!-- HERO -->
<section
    class="relative overflow-hidden border-b border-slate-200 grid-bg dark:border-slate-800 bg-gradient-to-b from-slate-50 to-white dark:from-slate-950 dark:to-slate-900"
>
    <div
        class="mx-auto grid max-w-6xl items-center gap-10 px-5 py-20 lg:grid-cols-2 lg:py-28"
    >
        <div class="rise">
            <span
                class="inline-flex items-center gap-2 rounded-full border border-blue-200/60 bg-blue-50/80 px-4 py-1.5 text-xs font-bold uppercase tracking-wider text-blue-700 backdrop-blur-sm dark:border-blue-800/50 dark:bg-blue-900/30 dark:text-blue-400 shadow-sm"
            >
                ДСТУ 19.701-90 · Локально у браузері
            </span>
            <h1
                class="mt-6 text-4xl font-extrabold leading-tight tracking-tight text-slate-900 dark:text-white sm:text-5xl lg:text-6xl"
            >
                Блок-схеми з коду <br /><span
                    class="bg-gradient-to-r from-blue-600 to-indigo-500 bg-clip-text text-transparent"
                    >за лічені секунди</span
                >
            </h1>
            <p
                class="mt-6 max-w-lg text-lg text-slate-600 leading-relaxed dark:text-slate-300"
            >
                Встав код на Python, C, C++ чи Pascal — отримай готову
                блок-схему за державним стандартом. Підправ у візуальному
                редакторі й додай у будь-який документ. Усе безкоштовно та надійно.
            </p>
            <div class="mt-10 flex flex-wrap gap-4">
                <a
                    href="{base}/app"
                    class="group relative inline-flex items-center justify-center overflow-hidden rounded-xl bg-blue-600 px-8 py-3.5 font-semibold text-white transition-all hover:bg-blue-700 hover:shadow-lg hover:shadow-blue-500/30 hover:-translate-y-0.5"
                >
                    <span class="mr-2">Побудувати схему</span>
                    <svg
                        class="h-4 w-4 transition-transform group-hover:translate-x-1"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        ><path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M14 5l7 7m0 0l-7 7m7-7H3"
                        /></svg
                    >
                </a>
                <a
                    href="{base}/guide"
                    class="inline-flex items-center justify-center rounded-xl border border-slate-200 bg-white/50 px-8 py-3.5 font-semibold text-slate-700 backdrop-blur-sm transition-all hover:border-slate-300 hover:bg-slate-50 hover:shadow-sm dark:border-slate-700 dark:bg-slate-800/50 dark:text-slate-200 dark:hover:border-slate-600 dark:hover:bg-slate-800 hover:-translate-y-0.5"
                >
                    Як це працює
                </a>
            </div>
        </div>

        <!-- міні-схема: справжній вивід рушія -->
        <div class="rise relative group">
            <!-- Розмитий фоновий градієнт для свічення -->
            <div
                class="absolute -inset-1 rounded-3xl bg-gradient-to-r from-blue-600 to-indigo-500 opacity-20 blur transition duration-1000 group-hover:opacity-30 dark:opacity-40"
            ></div>
            <div
                class="relative rounded-2xl border border-white/50 bg-white/80 p-8 shadow-2xl backdrop-blur-xl dark:border-slate-700/50 dark:bg-slate-900/80"
            >
                <svg
                    viewBox="0 0 472 564"
                    class="mx-auto h-auto w-full max-w-[400px] text-slate-900 dark:text-slate-200"
                    font-family="Outfit, sans-serif"
                    font-size="14"
                >
                    <defs
                        ><marker
                            id="arr"
                            markerWidth="9"
                            markerHeight="9"
                            refX="7.5"
                            refY="3"
                            orient="auto"
                            ><path
                                d="M0,0 L8,3 L0,6 Z"
                                fill="currentColor"
                            /></marker
                        ></defs
                    >
                    <path
                        d="M236.0 86.0 L236.0 130.0"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                        marker-end="url(#arr)"
                    />
                    <path
                        d="M236.0 176.0 L236.0 220.0"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                        marker-end="url(#arr)"
                    />
                    <path
                        d="M161.0 258.0 L121.8 258.0 L121.8 344.0"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                        marker-end="url(#arr)"
                    />
                    <text
                        x="155.0"
                        y="251.0"
                        text-anchor="end"
                        font-size="12"
                        fill="currentColor"
                        class="opacity-70">Так</text
                    >
                    <path
                        d="M311.0 258.0 L341.8 258.0 L341.8 344.0"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                        marker-end="url(#arr)"
                    />
                    <text
                        x="317.0"
                        y="251.0"
                        text-anchor="start"
                        font-size="12"
                        fill="currentColor"
                        class="opacity-70">Ні</text
                    >
                    <path
                        d="M121.8 390.0 L121.8 434.0 L236.0 434.0"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                    />
                    <path
                        d="M341.8 390.0 L341.8 434.0 L236.0 434.0"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                    />
                    <path
                        d="M236.0 434.0 L236.0 478.0"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                        marker-end="url(#arr)"
                    />
                    <rect
                        x="171.0"
                        y="44.0"
                        width="130.0"
                        height="42.0"
                        rx="21.0"
                        class="fill-white dark:fill-slate-800"
                        stroke="currentColor"
                        stroke-width="1.5"
                    />
                    <text
                        x="236.0"
                        y="65.0"
                        text-anchor="middle"
                        dominant-baseline="middle"
                        fill="currentColor">Початок</text
                    >
                    <polygon
                        points="189.4,130.0 301.0,130.0 282.6,176.0 171.0,176.0"
                        class="fill-white dark:fill-slate-800"
                        stroke="currentColor"
                        stroke-width="1.5"
                    />
                    <text
                        x="236.0"
                        y="153.0"
                        text-anchor="middle"
                        dominant-baseline="middle"
                        fill="currentColor">Ввід n</text
                    >
                    <polygon
                        points="236.0,220.0 311.0,258.0 236.0,296.0 161.0,258.0"
                        class="fill-white dark:fill-slate-800"
                        stroke="currentColor"
                        stroke-width="1.5"
                    />
                    <text
                        x="236.0"
                        y="258.0"
                        text-anchor="middle"
                        dominant-baseline="middle"
                        fill="currentColor">n > 0</text
                    >
                    <polygon
                        points="62.4,344.0 199.5,344.0 181.1,390.0 44.0,390.0"
                        class="fill-white dark:fill-slate-800"
                        stroke="currentColor"
                        stroke-width="1.5"
                    />
                    <text
                        x="121.8"
                        y="367.0"
                        text-anchor="middle"
                        dominant-baseline="middle"
                        fill="currentColor">Вивід «Додатне»</text
                    >
                    <polygon
                        points="273.9,344.0 428.0,344.0 409.6,390.0 255.5,390.0"
                        class="fill-white dark:fill-slate-800"
                        stroke="currentColor"
                        stroke-width="1.5"
                    />
                    <text
                        x="341.8"
                        y="367.0"
                        text-anchor="middle"
                        dominant-baseline="middle"
                        fill="currentColor">Вивід «Недодатне»</text
                    >
                    <rect
                        x="171.0"
                        y="478.0"
                        width="130.0"
                        height="42.0"
                        rx="21.0"
                        class="fill-white dark:fill-slate-800"
                        stroke="currentColor"
                        stroke-width="1.5"
                    />
                    <text
                        x="236.0"
                        y="499.0"
                        text-anchor="middle"
                        dominant-baseline="middle"
                        fill="currentColor">Кінець</text
                    >
                </svg>
            </div>
        </div>
    </div>
</section>

<!-- ПРОБЛЕМА -->
<section class="mx-auto max-w-4xl px-5 py-20 text-center">
    <h2 class="text-3xl font-bold text-slate-900 dark:text-slate-100">
        Малювати схеми вручну — це години роботи
    </h2>
    <p
        class="mx-auto mt-4 max-w-2xl text-lg text-slate-600 dark:text-slate-300"
    >
        Звіт, лабораторна, практична, курсова, диплом — будь-де, де є алгоритм,
        його треба зобразити схемою за ДСТУ. Лінійки, фігури, стрілки, переробки
        після кожної правки коду. rombik робить це за секунду й дозволяє
        підправити результат під будь-які вимоги.
    </p>
</section>

<!-- МОЖЛИВОСТІ -->
<section
    class="border-y border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900"
>
    <div class="mx-auto max-w-6xl px-5 py-20">
        <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {#each features as f (f.title)}
                <div
                    class="group rounded-2xl border border-slate-200 bg-white/50 p-6 backdrop-blur-sm transition-all hover:border-blue-300 hover:bg-white hover:shadow-xl hover:shadow-blue-500/5 dark:border-slate-800 dark:bg-slate-900/50 dark:hover:border-blue-700/50 dark:hover:bg-slate-800 hover:-translate-y-1"
                >
                    <div
                        class="grid h-12 w-12 place-items-center rounded-xl bg-blue-50 text-blue-600 transition-colors group-hover:bg-blue-600 group-hover:text-white dark:bg-blue-900/40 dark:text-blue-400 dark:group-hover:bg-blue-600 dark:group-hover:text-white"
                    >
                        <svg
                            width="24"
                            height="24"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        >
                            <path d={f.icon} />
                        </svg>
                    </div>
                    <h3
                        class="mt-5 text-lg font-bold text-slate-900 dark:text-white"
                    >
                        {f.title}
                    </h3>
                    <p
                        class="mt-2 text-sm leading-relaxed text-slate-600 dark:text-slate-400"
                    >
                        {f.body}
                    </p>
                </div>
            {/each}
        </div>
    </div>
</section>

<!-- ЯК ПРАЦЮЄ -->
<section class="mx-auto max-w-6xl px-5 py-20">
    <h2
        class="text-center text-3xl font-bold text-slate-900 dark:text-slate-100"
    >
        Як це працює?
    </h2>
    <div class="mt-12 grid gap-8 md:grid-cols-3">
        {#each steps as s (s.n)}
            <div
                class="relative rounded-2xl border border-slate-200 dark:border-slate-800 bg-white p-6 shadow-sm dark:bg-slate-800"
            >
                <div
                    class="absolute -top-5 left-6 grid h-10 w-10 place-items-center rounded-xl bg-blue-600 font-bold text-white shadow-lg shadow-blue-500/30"
                >
                    {s.n}
                </div>
                <h3
                    class="mt-4 text-lg font-semibold text-slate-900 dark:text-slate-100"
                >
                    {s.t}
                </h3>
                <p class="mt-2 text-slate-600 dark:text-slate-300">{s.d}</p>
            </div>
        {/each}
    </div>
</section>

<!-- FAQ -->
<section
    class="border-t border-slate-200 bg-slate-50 py-20 dark:border-slate-800 dark:bg-slate-900/50"
>
    <div class="mx-auto max-w-4xl px-5">
        <h2
            class="text-center text-3xl font-bold text-slate-900 dark:text-slate-100"
        >
            Часті запитання
        </h2>
        <div class="mt-10 space-y-4">
            <details
                class="group rounded-2xl border border-slate-200 bg-white p-6 shadow-sm open:border-blue-300 dark:border-slate-700 dark:bg-slate-800 dark:open:border-blue-500"
            >
                <summary
                    class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100"
                >
                    Які мови підтримуються?
                    <span
                        class="text-blue-500 transition-transform group-open:rotate-45"
                        >+</span
                    >
                </summary>
                <p class="mt-4 text-slate-600 dark:text-slate-300">
                    Повністю підтримуються <strong>Python</strong>,
                    <strong>C</strong>, <strong>C++</strong> та
                    <strong>Pascal</strong>. Усі базові конструкції (умови, цикли,
                    винятки, функції/процедури) автоматично розпізнаються і
                    будуються за ДСТУ. Деталі — на сторінці «Як це працює».
                </p>
            </details>
            <details
                class="group rounded-2xl border border-slate-200 bg-white p-6 shadow-sm open:border-blue-300 dark:border-slate-700 dark:bg-slate-800 dark:open:border-blue-500"
            >
                <summary
                    class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100"
                >
                    Як скопіювати схему у Word/Google Docs?
                    <span
                        class="text-blue-500 transition-transform group-open:rotate-45"
                        >+</span
                    >
                </summary>
                <p class="mt-4 text-slate-600 dark:text-slate-300">
                    Найпростіше — натиснути «Копіювати ▾» і вибрати PNG, а потім
                    просто вставити у документ (Ctrl+V). Також можна завантажити
                    SVG для векторної якості.
                </p>
            </details>
            <details
                class="group rounded-2xl border border-slate-200 bg-white p-6 shadow-sm open:border-blue-300 dark:border-slate-700 dark:bg-slate-800 dark:open:border-blue-500"
            >
                <summary
                    class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100"
                >
                    Чи працює це офлайн?
                    <span
                        class="text-blue-500 transition-transform group-open:rotate-45"
                        >+</span
                    >
                </summary>
                <p class="mt-4 text-slate-600 dark:text-slate-300">
                    Так! Після першого завантаження сторінки все працює
                    безпосередньо у вашому браузері: код розбирає Tree-sitter, а
                    схему будує рушій на TypeScript — без сервера. Ваш код та
                    схеми нікуди не відправляються.
                </p>
            </details>
            <details
                class="group rounded-2xl border border-slate-200 bg-white p-6 shadow-sm open:border-blue-300 dark:border-slate-700 dark:bg-slate-800 dark:open:border-blue-500"
            >
                <summary
                    class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100"
                >
                    Що робити, якщо схема вийшла занадто довгою чи широкою?
                    <span
                        class="text-blue-500 transition-transform group-open:rotate-45"
                        >+</span
                    >
                </summary>
                <p class="mt-4 text-slate-600 dark:text-slate-300">
                    Великі алгоритми можна розбити на частини. Кнопка
                    «✂ Розбити» під схемою ріже її автоматично за висотою. А у
                    візуальному редакторі ділиш вручну: виділяєш блоки рамкою-ласо
                    (Shift+тяг) → «⊞ Окрема схема», а «⊟ Об'єднати» зливає назад.
                    На стиках самі з'являються конектори-кружечки (А, Б, В…).
                </p>
            </details>
            <details
                class="group rounded-2xl border border-slate-200 bg-white p-6 shadow-sm open:border-blue-300 dark:border-slate-700 dark:bg-slate-800 dark:open:border-blue-500"
            >
                <summary
                    class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100"
                >
                    Як змінити текст усередині блоку або його форму?
                    <span
                        class="text-blue-500 transition-transform group-open:rotate-45"
                        >+</span
                    >
                </summary>
                <p class="mt-4 text-slate-600 dark:text-slate-300">
                    Натисніть «✎ Редагувати» під згенерованою схемою.
                    Відкриється візуальний редактор, де ви зможете двічі
                    клікнути на текст, щоб змінити його, додати нові блоки,
                    провести стрілки чи змінити їх маршрут.
                </p>
            </details>
            <details
                class="group rounded-2xl border border-slate-200 bg-white p-6 shadow-sm open:border-blue-300 dark:border-slate-700 dark:bg-slate-800 dark:open:border-blue-500"
            >
                <summary
                    class="flex cursor-pointer items-center justify-between font-semibold text-slate-900 dark:text-slate-100"
                >
                    Я хочу використовувати українські слова (Так/Ні,
                    Ввід/Вивід). Як це зробити?
                    <span
                        class="text-blue-500 transition-transform group-open:rotate-45"
                        >+</span
                    >
                </summary>
                <p class="mt-4 text-slate-600 dark:text-slate-300">
                    За замовчуванням усе вже українською — «Так/Ні», «Ввід/Вивід».
                    А якщо викладач хоче інакше, відкрий «Налаштування» (значок
                    шестерні) → вкладка «Текст та мітки» і постав будь-які свої
                    слова: підписи гілок (Так/Ні · Yes/No · +/−), слова вводу-виводу,
                    шаблон підпису «Рисунок N».
                </p>
            </details>
        </div>
    </div>
</section>

<!-- OPEN SOURCE -->
<section
    class="border-t border-slate-200 bg-white py-20 dark:border-slate-800 dark:bg-slate-900"
>
    <div class="mx-auto max-w-4xl px-5 text-center">
        <div
            class="inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300"
        >
            <svg
                width="32"
                height="32"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><path
                    d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"
                ></path></svg
            >
        </div>
        <h2 class="mt-6 text-3xl font-bold text-slate-900 dark:text-slate-100">
            100% Відкритий код
        </h2>
        <p class="mt-4 text-lg text-slate-600 dark:text-slate-400">
            rombik розробляється відкрито на GitHub. Рушій і фронтенд — повністю
            на TypeScript (Svelte 5), усе працює у браузері без сервера. Ви
            можете долучитися до розробки, повідомити про баг або запропонувати
            нову функцію.
        </p>
        <div class="mt-8 flex flex-wrap items-center justify-center gap-4">
            <a
                href="https://github.com/OlexiyOdarchuk/rombik"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-2 rounded-xl border border-slate-300 bg-white px-6 py-3 font-semibold text-slate-700 shadow-sm transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700"
            >
                Зірочка на GitHub
            </a>
            <button
                onclick={openBug}
                class="inline-flex items-center gap-2 rounded-xl border border-amber-300 bg-amber-50 px-6 py-3 font-semibold text-amber-700 shadow-sm transition hover:bg-amber-100 dark:border-amber-800 dark:bg-amber-950/30 dark:text-amber-400 dark:hover:bg-amber-950/60"
            >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
                Повідомити про помилку
            </button>
            <a
                href="https://ishawyha.dev/projects"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-2 rounded-xl border border-transparent px-6 py-3 font-semibold text-blue-600 transition hover:underline dark:text-blue-400"
            >
                Більше проєктів від ishawyha →
            </a>
        </div>
    </div>
</section>

<!-- CTA -->
<section
    class="bg-slate-900 dark:border-t dark:border-slate-800 dark:bg-slate-950"
>
    <div class="mx-auto max-w-4xl px-5 py-16 text-center">
        <h2 class="text-3xl font-bold text-white">Спробуй просто зараз</h2>
        <p class="mt-3 text-slate-300">
            Безкоштовно, у браузері, без реєстрації.
        </p>
        <a
            href="{base}/app"
            class="mt-7 inline-block rounded-xl bg-blue-600 px-7 py-3 font-semibold text-white shadow-lg transition hover:bg-blue-500"
        >
            Відкрити редактор →
        </a>
    </div>
</section>
