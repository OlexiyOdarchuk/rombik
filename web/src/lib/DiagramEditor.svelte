<script>
	// Інтерактивний редактор схеми (Фаза 1): вибір, перетягування з рероутом ребер,
	// правка тексту, додавання/видалення фігур, конектор «А-в-кружечку».
	// Працює на копії моделі; Зберегти → віддає diagram {shapes,edges,w,h} назад.
	let { diagram, onsave, oncancel } = $props();

	const MARGIN = 24;
	// Палітра типів: [kind, підпис, типова ширина/висота].
	const PALETTE = [
		['process', 'Дія', 150, 46],
		['decision', 'Умова', 150, 76],
		['io', 'Ввід/вивід', 150, 46],
		['terminator', 'Початок/Кінець', 130, 42],
		['loop', 'Цикл', 150, 50],
		['subprogram', 'Підпрограма', 150, 46],
		['connector', 'Конектор', 44, 44]
	];

	let nodes = $state([]);
	let edges = $state([]);
	let W = $state(100);
	let H = $state(100);
	let selId = $state(null);
	let editId = $state(null);
	let svgEl;
	let nid = 0;

	// Завантажуємо модель із diagram (копія, щоб не псувати оригінал до Зберегти).
	$effect(() => {
		load(diagram);
	});

	function load(d) {
		nid = 0;
		nodes = (d.shapes ?? []).map((s) => ({
			id: 'n' + nid++,
			kind: s.kind,
			x: s.x,
			y: s.y,
			w: s.w,
			h: s.h,
			text: s.text ?? ''
		}));
		edges = (d.edges ?? []).map((e, i) => {
			const pts = (e.points ?? []).map((p) => ({ x: p.x, y: p.y }));
			return {
				id: 'e' + i,
				points: pts,
				label: e.label ?? '',
				arrowless: !!e.arrowless,
				fromId: nodeAt(pts[0]),
				toId: nodeAt(pts[pts.length - 1])
			};
		});
		W = d.w ?? 100;
		H = d.h ?? 100;
		selId = null;
		editId = null;
	}

	const nodeById = (id) => nodes.find((n) => n.id === id);

	// Чи точка в межах якоїсь фігури (для прив'язки кінців ребер).
	function nodeAt(pt, margin = 8) {
		if (!pt) return null;
		for (const n of nodes) {
			if (pt.x >= n.x - margin && pt.x <= n.x + n.w + margin && pt.y >= n.y - margin && pt.y <= n.y + n.h + margin)
				return n.id;
		}
		return null;
	}

	// Ортогональний рероут ребра між фігурами from→to (після переміщення).
	function reroute(e) {
		const a = nodeById(e.fromId);
		const b = nodeById(e.toId);
		if (!a || !b) return;
		const acx = a.x + a.w / 2,
			bcx = b.x + b.w / 2,
			acy = a.y + a.h / 2,
			bcy = b.y + b.h / 2;
		const dx = bcx - acx,
			dy = bcy - acy;
		if (Math.abs(dy) >= Math.abs(dx)) {
			const ay = dy >= 0 ? a.y + a.h : a.y;
			const by = dy >= 0 ? b.y : b.y + b.h;
			e.points =
				acx === bcx
					? [{ x: acx, y: ay }, { x: bcx, y: by }]
					: [{ x: acx, y: ay }, { x: acx, y: (ay + by) / 2 }, { x: bcx, y: (ay + by) / 2 }, { x: bcx, y: by }];
		} else {
			const ax = dx >= 0 ? a.x + a.w : a.x;
			const bx = dx >= 0 ? b.x : b.x + b.w;
			e.points =
				acy === bcy
					? [{ x: ax, y: acy }, { x: bx, y: bcy }]
					: [{ x: ax, y: acy }, { x: (ax + bx) / 2, y: acy }, { x: (ax + bx) / 2, y: bcy }, { x: bx, y: bcy }];
		}
	}

	// Клієнтські координати → координати схеми (через CTM — коректно при масштабі).
	function toSvg(clientX, clientY) {
		const p = svgEl.createSVGPoint();
		p.x = clientX;
		p.y = clientY;
		const m = svgEl.getScreenCTM().inverse();
		const r = p.matrixTransform(m);
		return { x: r.x, y: r.y };
	}

	let drag = null;
	function startDrag(ev, n) {
		if (editId) return;
		selId = n.id;
		const s = toSvg(ev.clientX, ev.clientY);
		drag = { id: n.id, ox: n.x, oy: n.y, sx: s.x, sy: s.y };
		ev.stopPropagation();
	}
	function onMove(ev) {
		if (!drag) return;
		const s = toSvg(ev.clientX, ev.clientY);
		const n = nodeById(drag.id);
		n.x = drag.ox + (s.x - drag.sx);
		n.y = drag.oy + (s.y - drag.sy);
		for (const e of edges) if (e.fromId === n.id || e.toId === n.id) reroute(e);
		nodes = [...nodes];
		edges = [...edges];
	}
	const endDrag = () => (drag = null);

	function addNode(kind, w, h) {
		const id = 'n' + nid++;
		nodes.push({ id, kind, x: W / 2 - w / 2, y: H / 2 - h / 2, w, h, text: kind === 'connector' ? 'А' : 'текст' });
		nodes = [...nodes];
		selId = id;
		editId = id;
	}

	function delSel() {
		if (!selId) return;
		nodes = nodes.filter((n) => n.id !== selId);
		edges = edges.filter((e) => e.fromId !== selId && e.toId !== selId);
		selId = null;
	}

	function onKey(ev) {
		if (editId) return;
		if ((ev.key === 'Delete' || ev.key === 'Backspace') && selId) {
			ev.preventDefault();
			delSel();
		}
	}

	// Текст фігури → автоширина (щоб уміщався), окрім конектора (фіксоване коло).
	function setText(n, v) {
		n.text = v;
		if (n.kind !== 'connector') n.w = Math.max(n.w, v.length * 8.5 + 32);
		nodes = [...nodes];
	}

	function save() {
		// Нормалізуємо межі: зсуваємо все так, щоб усе було видимим, рахуємо W/H.
		let minX = Infinity,
			minY = Infinity,
			maxX = -Infinity,
			maxY = -Infinity;
		const pts = [];
		for (const n of nodes) pts.push([n.x, n.y], [n.x + n.w, n.y + n.h]);
		for (const e of edges) for (const p of e.points) pts.push([p.x, p.y]);
		for (const [x, y] of pts) {
			minX = Math.min(minX, x);
			minY = Math.min(minY, y);
			maxX = Math.max(maxX, x);
			maxY = Math.max(maxY, y);
		}
		if (!isFinite(minX)) {
			minX = minY = 0;
			maxX = maxY = 100;
		}
		const dx = MARGIN - minX,
			dy = MARGIN - minY;
		const d = {
			shapes: nodes.map((n) => ({ kind: n.kind, x: n.x + dx, y: n.y + dy, w: n.w, h: n.h, text: n.text })),
			edges: edges.map((e) => ({
				points: e.points.map((p) => ({ x: p.x + dx, y: p.y + dy })),
				label: e.label || undefined,
				arrowless: e.arrowless || undefined
			})),
			w: maxX - minX + 2 * MARGIN,
			h: maxY - minY + 2 * MARGIN
		};
		onsave?.(d);
	}

	// --- геометрія фігур для SVG ---
	function poly(n) {
		const { x, y, w, h } = n;
		const cx = x + w / 2,
			cy = y + h / 2;
		switch (n.kind) {
			case 'decision':
				return `${cx},${y} ${x + w},${cy} ${cx},${y + h} ${x},${cy}`;
			case 'io': {
				const s = h * 0.4;
				return `${x + s},${y} ${x + w},${y} ${x + w - s},${y + h} ${x},${y + h}`;
			}
			case 'loop': {
				const s = h * 0.5;
				return `${x + s},${y} ${x + w - s},${y} ${x + w},${cy} ${x + w - s},${y + h} ${x + s},${y + h} ${x},${cy}`;
			}
		}
		return '';
	}
	const pathD = (pts) => pts.map((p, i) => `${i ? 'L' : 'M'}${p.x} ${p.y}`).join(' ');
</script>

<svelte:window onpointermove={onMove} onpointerup={endDrag} onkeydown={onKey} />

<div class="fixed inset-0 z-50 flex flex-col bg-slate-900/60 backdrop-blur-sm">
	<!-- тулбар -->
	<div class="flex items-center gap-2 border-b border-slate-200 bg-white px-4 py-2 dark:border-slate-700 dark:bg-slate-900">
		<span class="text-sm font-semibold text-slate-700 dark:text-slate-200">Редактор схеми</span>
		<div class="ml-2 flex flex-wrap gap-1">
			{#each PALETTE as [kind, label, w, h] (kind)}
				<button
					onclick={() => addNode(kind, w, h)}
					class="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800"
				>
					+ {label}
				</button>
			{/each}
		</div>
		<button
			onclick={delSel}
			disabled={!selId}
			class="rounded border border-red-200 px-2 py-1 text-xs font-medium text-red-600 transition hover:bg-red-50 disabled:opacity-40 dark:border-red-900 dark:text-red-400 dark:hover:bg-red-950"
		>
			Видалити
		</button>
		<div class="ml-auto flex gap-2">
			<button onclick={() => oncancel?.()} class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm text-slate-700 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-800">
				Скасувати
			</button>
			<button onclick={save} class="rounded-lg bg-blue-600 px-4 py-1.5 text-sm font-semibold text-white transition hover:bg-blue-700">
				Зберегти
			</button>
		</div>
	</div>

	<!-- полотно -->
	<div class="min-h-0 flex-1 overflow-auto bg-white p-6 dark:bg-slate-950">
		<svg
			bind:this={svgEl}
			viewBox="0 0 {W} {H}"
			width={W}
			height={H}
			class="mx-auto block max-w-full"
			style="height:auto"
			font-family="Arial, sans-serif"
			font-size="14"
			onpointerdown={() => (selId = null)}
			role="presentation"
		>
			<defs>
				<marker id="ed-arr" markerWidth="9" markerHeight="9" refX="7.5" refY="3" orient="auto">
					<path d="M0,0 L8,3 L0,6 Z" fill="#222" />
				</marker>
			</defs>
			<rect width={W} height={H} fill="#fff" />

			{#each edges as e (e.id)}
				<path d={pathD(e.points)} fill="none" stroke="#222" stroke-width="1.5" marker-end={e.arrowless ? '' : 'url(#ed-arr)'} />
				{#if e.label}
					<text x={e.points[0].x + 6} y={e.points[0].y - 6} font-size="12" fill="#444">{e.label}</text>
				{/if}
			{/each}

			{#each nodes as n (n.id)}
				<g
					class="cursor-move"
					onpointerdown={(ev) => startDrag(ev, n)}
					ondblclick={() => (editId = n.id)}
					role="presentation"
				>
					{#if n.kind === 'process' || n.kind === 'subprogram'}
						<rect x={n.x} y={n.y} width={n.w} height={n.h} fill="#fdfdfd" stroke={selId === n.id ? '#2563eb' : '#222'} stroke-width={selId === n.id ? 2.5 : 1.5} />
						{#if n.kind === 'subprogram'}
							<line x1={n.x + 9} y1={n.y} x2={n.x + 9} y2={n.y + n.h} stroke="#222" stroke-width="1.5" />
							<line x1={n.x + n.w - 9} y1={n.y} x2={n.x + n.w - 9} y2={n.y + n.h} stroke="#222" stroke-width="1.5" />
						{/if}
					{:else if n.kind === 'terminator'}
						<rect x={n.x} y={n.y} width={n.w} height={n.h} rx={n.h / 2} fill="#fdfdfd" stroke={selId === n.id ? '#2563eb' : '#222'} stroke-width={selId === n.id ? 2.5 : 1.5} />
					{:else if n.kind === 'connector'}
						<circle cx={n.x + n.w / 2} cy={n.y + n.h / 2} r={Math.min(n.w, n.h) / 2} fill="#fdfdfd" stroke={selId === n.id ? '#2563eb' : '#222'} stroke-width={selId === n.id ? 2.5 : 1.5} />
					{:else}
						<polygon points={poly(n)} fill="#fdfdfd" stroke={selId === n.id ? '#2563eb' : '#222'} stroke-width={selId === n.id ? 2.5 : 1.5} />
					{/if}

					{#if editId === n.id}
						<foreignObject x={n.x} y={n.y + n.h / 2 - 12} width={n.w} height="24">
							<input
								value={n.text}
								oninput={(ev) => setText(n, ev.currentTarget.value)}
								onblur={() => (editId = null)}
								onkeydown={(ev) => ev.key === 'Enter' && (editId = null)}
								onpointerdown={(ev) => ev.stopPropagation()}
								class="h-full w-full border-0 bg-transparent text-center text-sm outline-none"
								autofocus
							/>
						</foreignObject>
					{:else}
						<text x={n.x + n.w / 2} y={n.y + n.h / 2} text-anchor="middle" dominant-baseline="middle" fill="#111" pointer-events="none">{n.text}</text>
					{/if}
				</g>
			{/each}
		</svg>
	</div>
	<p class="border-t border-slate-200 bg-white px-4 py-1.5 text-xs text-slate-400 dark:border-slate-700 dark:bg-slate-900">
		Перетягни фігуру мишкою · подвійний клік — змінити текст · Delete — видалити · додавай фігури з палітри
	</p>
</div>
