<script>
	// Візуальний редактор схеми (Фаза 2): безмежне полотно (пан+зум), редаговані
	// стрілки (вибір, тягання вузлів і кінців, переприв'язка, малювання нових),
	// авто-ортогональний рероут при русі блоків, розділення конектором.
	let { diagram, onsave, oncancel } = $props();

	const MARGIN = 24;
	const PALETTE = [
		['process', 'Дія', 150, 46],
		['decision', 'Умова', 150, 76],
		['io', 'Ввід/вивід', 150, 46],
		['terminator', 'Початок/Кінець', 130, 42],
		['loop', 'Цикл', 150, 50],
		['subprogram', 'Підпрограма', 150, 46],
		['connector', 'Конектор', 46, 46]
	];

	let nodes = $state([]);
	let edges = $state([]);
	let sel = $state(null); // {type:'node'|'edge', id}
	let editId = $state(null); // фігура, чий текст редагуємо
	let editLabel = $state(null); // ребро, чий підпис (Так/Ні) редагуємо
	let guides = $state([]); // напрямні-магніти під час перетягування
	let view = $state({ x: 40, y: 40, scale: 1 });
	let gEl; // внутрішня <g> з трансформацією (для коректного перерахунку координат)
	let nid = 0,
		eid = 0;

	// --- історія (undo/redo) ---
	let past = [];
	let futureH = [];
	const snap = () => JSON.stringify({ nodes, edges });
	function remember() {
		past.push(snap());
		if (past.length > 60) past.shift();
		futureH = [];
	}
	function restore(s) {
		const o = JSON.parse(s);
		nodes = o.nodes;
		edges = o.edges;
		sel = null;
		editId = null;
		editLabel = null;
	}
	function undo() {
		if (!past.length) return;
		futureH.push(snap());
		restore(past.pop());
	}
	function redo() {
		if (!futureH.length) return;
		past.push(snap());
		restore(futureH.pop());
	}

	$effect(() => load(diagram));

	function load(d) {
		let k = 0;
		const ns = (d.shapes ?? []).map((s) => ({ id: 'n' + k++, kind: s.kind, x: s.x, y: s.y, w: s.w, h: s.h, text: s.text ?? '' }));
		const at = (pt) => {
			if (!pt) return null;
			for (const n of ns) if (pt.x >= n.x - 8 && pt.x <= n.x + n.w + 8 && pt.y >= n.y - 8 && pt.y <= n.y + n.h + 8) return n.id;
			return null;
		};
		let j = 0;
		const es = (d.edges ?? []).map((e) => {
			const pts = (e.points ?? []).map((p) => ({ x: p.x, y: p.y }));
			const lp = labelAnchor(pts[0], pts[1]);
			return {
				id: 'e' + j++,
				points: pts,
				label: e.label ?? '',
				lx: lp.x,
				ly: lp.y, // позиція підпису (рухається окремо)
				arrowless: !!e.arrowless,
				fromId: at(pts[0]),
				toId: at(pts[pts.length - 1]),
				manual: false
			};
		});
		nid = k;
		eid = j;
		nodes = ns;
		edges = es;
		sel = null;
		editId = null;
	}

	const nodeById = (id) => nodes.find((n) => n.id === id);

	// Позиція підпису ребра (Так/Ні) — як у Go diagram.LabelAnchor (далі від ромба).
	function labelAnchor(p0, p1) {
		if (!p0) return { x: 0, y: 0 };
		if (!p1) return { x: p0.x + 8, y: p0.y - 8 };
		if (p1.y === p0.y && p1.x !== p0.x) {
			let off = (p1.x - p0.x) / 2;
			off = Math.max(-34, Math.min(34, off));
			return { x: p0.x + off, y: p0.y - 9 };
		}
		if (p1.x === p0.x) return { x: p0.x + 10, y: (p0.y + p1.y) / 2 };
		return { x: p0.x + 8, y: p0.y - 8 };
	}

	// --- координати: екран → світ (через CTM внутрішньої <g>, враховує пан+зум) ---
	function toWorld(clientX, clientY) {
		const p = gEl.ownerSVGElement.createSVGPoint();
		p.x = clientX;
		p.y = clientY;
		const r = p.matrixTransform(gEl.getScreenCTM().inverse());
		return { x: r.x, y: r.y };
	}

	// --- ортогональний маршрут між фігурами (порт-до-порту) ---
	function portRoute(a, b) {
		const acx = a.x + a.w / 2,
			acy = a.y + a.h / 2,
			bcx = b.x + b.w / 2,
			bcy = b.y + b.h / 2;
		const dx = bcx - acx,
			dy = bcy - acy;
		if (Math.abs(dy) >= Math.abs(dx)) {
			const ay = dy >= 0 ? a.y + a.h : a.y;
			const by = dy >= 0 ? b.y : b.y + b.h;
			if (Math.abs(acx - bcx) < 1) return [{ x: acx, y: ay }, { x: bcx, y: by }];
			const my = (ay + by) / 2;
			return [{ x: acx, y: ay }, { x: acx, y: my }, { x: bcx, y: my }, { x: bcx, y: by }];
		}
		const ax = dx >= 0 ? a.x + a.w : a.x;
		const bx = dx >= 0 ? b.x : b.x + b.w;
		if (Math.abs(acy - bcy) < 1) return [{ x: ax, y: acy }, { x: bx, y: bcy }];
		const mx = (ax + bx) / 2;
		return [{ x: ax, y: acy }, { x: mx, y: acy }, { x: mx, y: bcy }, { x: bx, y: bcy }];
	}

	// Рероут усіх ребер, дотичних до фігури. Авто-ребра — повний ортогональний
	// маршрут; ручні (користувач тягав вузол) — лише зсув приєднаного кінця.
	function rerouteFor(nodeId, dx, dy) {
		for (const e of edges) {
			const touchFrom = e.fromId === nodeId,
				touchTo = e.toId === nodeId;
			if (!touchFrom && !touchTo) continue;
			const a = nodeById(e.fromId),
				b = nodeById(e.toId);
			if (!e.manual && a && b) {
				e.points = portRoute(a, b);
			} else {
				if (touchFrom && e.points[0]) {
					e.points[0].x += dx;
					e.points[0].y += dy;
				}
				if (touchTo && e.points.length) {
					e.points[e.points.length - 1].x += dx;
					e.points[e.points.length - 1].y += dy;
				}
			}
		}
	}

	// Магнічення: підрівнюємо позицію фігури до країв/центрів інших фігур.
	function magnet(nx, ny, node) {
		const T = 7 / view.scale;
		const myX = [nx, nx + node.w / 2, nx + node.w];
		const myY = [ny, ny + node.h / 2, ny + node.h];
		let rx = nx,
			ry = ny,
			gx = null,
			gy = null;
		for (const o of nodes) {
			if (o.id === node.id) continue;
			const oX = [o.x, o.x + o.w / 2, o.x + o.w];
			const oY = [o.y, o.y + o.h / 2, o.y + o.h];
			for (let i = 0; i < 3; i++) for (const v of oX) if (gx === null && Math.abs(myX[i] - v) < T) ((rx = nx + (v - myX[i])), (gx = v));
			for (let i = 0; i < 3; i++) for (const v of oY) if (gy === null && Math.abs(myY[i] - v) < T) ((ry = ny + (v - myY[i])), (gy = v));
		}
		const g = [];
		if (gx !== null) g.push({ t: 'v', v: gx });
		if (gy !== null) g.push({ t: 'h', v: gy });
		return { x: rx, y: ry, g };
	}

	// --- взаємодія (одна машина станів через pointer) ---
	let act = $state(null); // {kind, ...}

	function bgDown(ev) {
		// фон: пан полотна + зняти виділення
		sel = null;
		editId = null;
		act = { kind: 'pan', sx: ev.clientX, sy: ev.clientY, vx: view.x, vy: view.y };
	}

	function nodeDown(ev, n) {
		ev.stopPropagation();
		if (editId && editId !== n.id) editId = null;
		sel = { type: 'node', id: n.id };
		const w = toWorld(ev.clientX, ev.clientY);
		act = { kind: 'node', id: n.id, ox: n.x, oy: n.y, wx: w.x, wy: w.y };
	}

	function edgeDown(ev, e) {
		ev.stopPropagation();
		sel = { type: 'edge', id: e.id };
	}

	function labelDown(ev, e) {
		ev.stopPropagation();
		sel = { type: 'edge', id: e.id };
		const w = toWorld(ev.clientX, ev.clientY);
		act = { kind: 'label', id: e.id, ox: e.lx, oy: e.ly, wx: w.x, wy: w.y };
	}

	function vertexDown(ev, e, i) {
		ev.stopPropagation();
		sel = { type: 'edge', id: e.id };
		act = { kind: 'vertex', id: e.id, i, end: i === 0 || i === e.points.length - 1 };
		e.manual = true;
	}

	function portDown(ev, n, side) {
		ev.stopPropagation();
		const w = toWorld(ev.clientX, ev.clientY);
		act = { kind: 'draw', from: n.id, side, x: w.x, y: w.y };
	}

	function onMove(ev) {
		if (!act) return;
		// Знімок в історію — лише при ПЕРШОМУ русі (не на простий клік-вибір).
		if (!act.saved && (act.kind === 'node' || act.kind === 'vertex' || act.kind === 'label')) {
			remember();
			act.saved = true;
		}
		const w = toWorld(ev.clientX, ev.clientY);
		if (act.kind === 'pan') {
			view.x = act.vx + (ev.clientX - act.sx);
			view.y = act.vy + (ev.clientY - act.sy);
		} else if (act.kind === 'node') {
			const n = nodeById(act.id);
			let nx = act.ox + (w.x - act.wx),
				ny = act.oy + (w.y - act.wy);
			const m = magnet(nx, ny, n);
			nx = m.x;
			ny = m.y;
			guides = m.g;
			const dx = nx - n.x,
				dy = ny - n.y;
			n.x = nx;
			n.y = ny;
			rerouteFor(n.id, dx, dy);
			nodes = [...nodes];
			edges = [...edges];
		} else if (act.kind === 'label') {
			const e = edges.find((x) => x.id === act.id);
			e.lx = act.ox + (w.x - act.wx);
			e.ly = act.oy + (w.y - act.wy);
			edges = [...edges];
		} else if (act.kind === 'vertex') {
			const e = edges.find((x) => x.id === act.id);
			e.points[act.i] = { x: w.x, y: w.y };
			edges = [...edges];
		} else if (act.kind === 'draw') {
			act.x = w.x;
			act.y = w.y;
			edges = [...edges]; // оновити пунктир-прев'ю
		}
	}

	function nodeAtPoint(pt) {
		for (const n of nodes) if (pt.x >= n.x && pt.x <= n.x + n.w && pt.y >= n.y && pt.y <= n.y + n.h) return n;
		return null;
	}

	function onUp(ev) {
		if (!act) return;
		const w = toWorld(ev.clientX, ev.clientY);
		if (act.kind === 'vertex' && act.end) {
			// відпустили кінець ребра над фігурою → переприв'язка
			const over = nodeAtPoint(w);
			const e = edges.find((x) => x.id === act.id);
			if (over) {
				if (act.i === 0) e.fromId = over.id;
				else e.toId = over.id;
			}
			edges = [...edges];
		} else if (act.kind === 'draw') {
			const over = nodeAtPoint(w);
			if (over && over.id !== act.from) {
				remember();
				const a = nodeById(act.from);
				const p = portRoute(a, over);
				edges.push({ id: 'e' + eid++, points: p, label: '', lx: p[0].x + 8, ly: p[0].y - 8, arrowless: false, fromId: act.from, toId: over.id, manual: false });
				edges = [...edges];
			}
		}
		guides = [];
		act = null;
	}

	function onWheel(ev) {
		ev.preventDefault();
		const w = toWorld(ev.clientX, ev.clientY);
		const f = ev.deltaY < 0 ? 1.1 : 1 / 1.1;
		const ns = Math.min(3, Math.max(0.2, view.scale * f));
		// масштабуємо навколо курсора
		view.x = ev.clientX - (gEl.ownerSVGElement.getBoundingClientRect().left + w.x * ns + 0);
		view.y = ev.clientY - (gEl.ownerSVGElement.getBoundingClientRect().top + w.y * ns + 0);
		view.scale = ns;
	}

	function zoom(f) {
		view.scale = Math.min(3, Math.max(0.2, view.scale * f));
	}
	function fit() {
		view = { x: 40, y: 40, scale: 1 };
	}

	function addNode(kind, w, h) {
		remember();
		const c = toWorld(gEl.ownerSVGElement.getBoundingClientRect().left + 300, gEl.ownerSVGElement.getBoundingClientRect().top + 200);
		const id = 'n' + nid++;
		nodes.push({ id, kind, x: c.x, y: c.y, w, h, text: kind === 'connector' ? nextLetter() : 'текст' });
		nodes = [...nodes];
		sel = { type: 'node', id };
		if (kind !== 'connector') editId = id;
	}

	function delSel() {
		if (!sel) return;
		remember();
		if (sel.type === 'node') {
			nodes = nodes.filter((n) => n.id !== sel.id);
			edges = edges.filter((e) => e.fromId !== sel.id && e.toId !== sel.id);
		} else {
			edges = edges.filter((e) => e.id !== sel.id);
		}
		sel = null;
	}

	function onKey(ev) {
		if (editId || editLabel) return;
		const mod = ev.ctrlKey || ev.metaKey;
		if (mod && (ev.key === 'z' || ev.key === 'Z')) {
			ev.preventDefault();
			ev.shiftKey ? redo() : undo();
			return;
		}
		if (mod && (ev.key === 'y' || ev.key === 'Y')) {
			ev.preventDefault();
			redo();
			return;
		}
		if ((ev.key === 'Delete' || ev.key === 'Backspace') && sel) {
			ev.preventDefault();
			delSel();
		}
	}

	function setText(n, v) {
		n.text = v;
		if (n.kind !== 'connector') n.w = Math.max(n.w, v.length * 8.5 + 32);
		nodes = [...nodes];
	}

	// Наступна вільна літера для конектора (А, Б, В…).
	const ABC = 'АБВГДЕЖЗИКЛМНОП';
	function nextLetter() {
		const used = new Set(nodes.filter((n) => n.kind === 'connector').map((n) => n.text));
		for (const c of ABC) if (!used.has(c)) return c;
		return 'А';
	}

	// Розділення: клік-на-блок → схема рветься надвоє парою конекторів «X».
	// Блок із усіма нащадками (разом із їхніми ребрами!) їде в окрему колонку
	// праворуч; на стику в частині-1 стає конектор-вихід «X», перед блоком —
	// конектор-вхід «X».
	function splitAt(nodeId) {
		const node = nodeById(nodeId);
		if (!node) return;
		remember();
		const letter = nextLetter();
		// нащадки вперед (BFS по вихідних ребрах), разом із блоком
		const set = new Set([nodeId]);
		const q = [nodeId];
		while (q.length) {
			const cur = q.shift();
			for (const e of edges) if (e.fromId === cur && e.toId && !set.has(e.toId)) { set.add(e.toId); q.push(e.toId); }
		}
		// зсув частини-2 у чисту колонку праворуч
		let maxX = -Infinity,
			sx = Infinity,
			sy = Infinity;
		for (const n of nodes) maxX = Math.max(maxX, n.x + n.w);
		for (const n of nodes) if (set.has(n.id)) ((sx = Math.min(sx, n.x)), (sy = Math.min(sy, n.y)));
		const offX = maxX + 90 - sx;
		const offY = 100 - sy; // верх частини-2 ~ під конектором-входом
		const ox0 = node.x,
			oy0 = node.y; // ОРИГІНАЛЬНЕ місце блоку (для конектора-виходу)
		// рухаємо вузли І їхні внутрішні ребра
		for (const n of nodes) if (set.has(n.id)) ((n.x += offX), (n.y += offY));
		for (const e of edges) if (set.has(e.fromId) && set.has(e.toId)) for (const p of e.points) ((p.x += offX), (p.y += offY));
		// конектори: вихід на місці блоку (частина-1), вхід над переміщеним блоком
		const exit = { id: 'n' + nid++, kind: 'connector', x: ox0 + node.w / 2 - 23, y: oy0, w: 46, h: 46, text: letter };
		const entry = { id: 'n' + nid++, kind: 'connector', x: node.x + node.w / 2 - 23, y: node.y - 92, w: 46, h: 46, text: letter };
		nodes.push(exit, entry);
		// межові ребра: вхідні в блок із частини-1 → у конектор-вихід; інші межові — рероут
		for (const e of edges) {
			const fin = set.has(e.fromId),
				tin = set.has(e.toId);
			if (fin === tin) continue;
			if (e.toId === nodeId && !fin) {
				e.toId = exit.id;
			}
			const a = nodeById(e.fromId),
				b = nodeById(e.toId);
			if (a && b) ((e.points = portRoute(a, b)), (e.manual = false));
		}
		const p = portRoute(entry, node);
		edges.push({ id: 'e' + eid++, points: p, label: '', lx: p[0].x + 8, ly: p[0].y - 8, arrowless: false, fromId: entry.id, toId: nodeId, manual: false });
		nodes = [...nodes];
		edges = [...edges];
		sel = null;
	}

	function save() {
		let minX = Infinity,
			minY = Infinity,
			maxX = -Infinity,
			maxY = -Infinity;
		const acc = (x, y) => {
			minX = Math.min(minX, x);
			minY = Math.min(minY, y);
			maxX = Math.max(maxX, x);
			maxY = Math.max(maxY, y);
		};
		for (const n of nodes) {
			acc(n.x, n.y);
			acc(n.x + n.w, n.y + n.h);
		}
		for (const e of edges) for (const p of e.points) acc(p.x, p.y);
		if (!isFinite(minX)) {
			minX = minY = 0;
			maxX = maxY = 100;
		}
		const dx = MARGIN - minX,
			dy = MARGIN - minY;
		onsave?.({
			shapes: nodes.map((n) => ({ kind: n.kind, x: n.x + dx, y: n.y + dy, w: n.w, h: n.h, text: n.text })),
			edges: edges.map((e) => ({ points: e.points.map((p) => ({ x: p.x + dx, y: p.y + dy })), label: e.label || undefined, arrowless: e.arrowless || undefined })),
			w: maxX - minX + 2 * MARGIN,
			h: maxY - minY + 2 * MARGIN
		});
	}

	// --- геометрія фігур ---
	function poly(n) {
		const { x, y, w, h } = n;
		const cx = x + w / 2,
			cy = y + h / 2;
		if (n.kind === 'decision') return `${cx},${y} ${x + w},${cy} ${cx},${y + h} ${x},${cy}`;
		if (n.kind === 'io') {
			const s = h * 0.4;
			return `${x + s},${y} ${x + w},${y} ${x + w - s},${y + h} ${x},${y + h}`;
		}
		if (n.kind === 'loop') {
			const s = h * 0.5;
			return `${x + s},${y} ${x + w - s},${y} ${x + w},${cy} ${x + w - s},${y + h} ${x + s},${y + h} ${x},${cy}`;
		}
		return '';
	}
	const pathD = (pts) => pts.map((p, i) => `${i ? 'L' : 'M'}${p.x} ${p.y}`).join(' ');
	const isSel = (t, id) => sel && sel.type === t && sel.id === id;
	const ports = (n) => [
		{ side: 'top', x: n.x + n.w / 2, y: n.y },
		{ side: 'bottom', x: n.x + n.w / 2, y: n.y + n.h },
		{ side: 'left', x: n.x, y: n.y + n.h / 2 },
		{ side: 'right', x: n.x + n.w, y: n.y + n.h / 2 }
	];
</script>

<svelte:window onpointermove={onMove} onpointerup={onUp} onkeydown={onKey} />

<div class="fixed inset-0 z-50 flex flex-col bg-slate-900/70 backdrop-blur-sm">
	<!-- тулбар -->
	<div class="flex flex-wrap items-center gap-2 border-b border-slate-200 bg-white px-4 py-2 dark:border-slate-700 dark:bg-slate-900">
		<span class="text-sm font-semibold text-slate-700 dark:text-slate-200">Редактор</span>
		<div class="flex flex-wrap gap-1">
			{#each PALETTE as [kind, label, w, h] (kind)}
				<button onclick={() => addNode(kind, w, h)} class="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800">+ {label}</button>
			{/each}
		</div>
		<div class="mx-1 h-5 w-px bg-slate-200 dark:bg-slate-700"></div>
		{#if sel?.type === 'node' && nodeById(sel.id)?.kind !== 'connector'}
			<button onclick={() => splitAt(sel.id)} class="rounded border border-amber-300 px-2 py-1 text-xs font-medium text-amber-700 transition hover:bg-amber-50 dark:border-amber-800 dark:text-amber-400 dark:hover:bg-amber-950">✂ Розділити тут</button>
		{/if}
		<button onclick={delSel} disabled={!sel} class="rounded border border-red-200 px-2 py-1 text-xs font-medium text-red-600 transition hover:bg-red-50 disabled:opacity-40 dark:border-red-900 dark:text-red-400 dark:hover:bg-red-950">Видалити</button>
		<button onclick={undo} title="Скасувати (Ctrl+Z)" class="grid h-7 w-7 place-items-center rounded border border-slate-300 text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800">↶</button>
		<button onclick={redo} title="Повторити (Ctrl+Shift+Z)" class="grid h-7 w-7 place-items-center rounded border border-slate-300 text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800">↷</button>
		<div class="ml-auto flex items-center gap-1.5">
			<button onclick={() => zoom(1 / 1.2)} class="grid h-7 w-7 place-items-center rounded border border-slate-300 text-slate-600 dark:border-slate-600 dark:text-slate-300">−</button>
			<span class="w-10 text-center text-xs text-slate-400">{Math.round(view.scale * 100)}%</span>
			<button onclick={() => zoom(1.2)} class="grid h-7 w-7 place-items-center rounded border border-slate-300 text-slate-600 dark:border-slate-600 dark:text-slate-300">+</button>
			<button onclick={fit} class="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 dark:border-slate-600 dark:text-slate-300">Скинути</button>
			<div class="mx-1 h-5 w-px bg-slate-200 dark:bg-slate-700"></div>
			<button onclick={() => oncancel?.()} class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm text-slate-700 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-800">Скасувати</button>
			<button onclick={save} class="rounded-lg bg-blue-600 px-4 py-1.5 text-sm font-semibold text-white transition hover:bg-blue-700">Зберегти</button>
		</div>
	</div>

	<!-- безмежне полотно -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<svg class="min-h-0 flex-1 touch-none bg-white dark:bg-slate-950" onpointerdown={bgDown} onwheel={onWheel} role="presentation">
		<defs>
			<marker id="ed-arr" markerWidth="9" markerHeight="9" refX="7.5" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6 Z" fill="#222" /></marker>
			<pattern id="ed-grid" width="28" height="28" patternUnits="userSpaceOnUse" patternTransform="translate({view.x},{view.y}) scale({view.scale})">
				<path d="M28 0H0V28" fill="none" stroke="#e2e8f0" stroke-width="1" />
			</pattern>
		</defs>
		<rect width="100%" height="100%" fill="url(#ed-grid)" />
		<g bind:this={gEl} transform="translate({view.x},{view.y}) scale({view.scale})">
			<!-- ребра -->
			{#each edges as e (e.id)}
				<!-- товста невидима «хіт-зона» для зручного кліку -->
				<path d={pathD(e.points)} fill="none" stroke="transparent" stroke-width="12" class="cursor-pointer" onpointerdown={(ev) => edgeDown(ev, e)} role="presentation" />
				<path d={pathD(e.points)} fill="none" stroke={isSel('edge', e.id) ? '#2563eb' : '#222'} stroke-width={isSel('edge', e.id) ? 2.5 : 1.5} marker-end={e.arrowless ? '' : 'url(#ed-arr)'} pointer-events="none" />
				{#if editLabel === e.id}
					<foreignObject x={e.lx - 6} y={e.ly - 16} width="90" height="22">
						<input value={e.label} oninput={(ev) => { e.label = ev.currentTarget.value; edges = [...edges]; }} onblur={() => (editLabel = null)} onkeydown={(ev) => ev.key === 'Enter' && (editLabel = null)} onpointerdown={(ev) => ev.stopPropagation()} class="w-full rounded border border-blue-300 bg-white px-1 text-xs outline-none dark:bg-slate-800 dark:text-slate-200" autofocus />
					</foreignObject>
				{:else if e.label}
					<text x={e.lx} y={e.ly} text-anchor="middle" font-size="12" fill={isSel('edge', e.id) ? '#2563eb' : '#444'} class="cursor-move" onpointerdown={(ev) => labelDown(ev, e)} ondblclick={() => (editLabel = e.id)} role="presentation">{e.label}</text>
				{/if}
				{#if isSel('edge', e.id)}
					{#each e.points as p, i (i)}
						<circle cx={p.x} cy={p.y} r="5" fill="#fff" stroke="#2563eb" stroke-width="2" class="cursor-grab" onpointerdown={(ev) => vertexDown(ev, e, i)} role="presentation" />
					{/each}
				{/if}
			{/each}

			<!-- напрямні-магніти (під час перетягування) -->
			{#each guides as gd (gd.t + gd.v)}
				{#if gd.t === 'v'}
					<line x1={gd.v} y1={-5000} x2={gd.v} y2={5000} stroke="#f43f5e" stroke-width={1 / view.scale} stroke-dasharray="4 3" pointer-events="none" />
				{:else}
					<line x1={-5000} y1={gd.v} x2={5000} y2={gd.v} stroke="#f43f5e" stroke-width={1 / view.scale} stroke-dasharray="4 3" pointer-events="none" />
				{/if}
			{/each}

			<!-- прев'ю нового ребра -->
			{#if act?.kind === 'draw' && nodeById(act.from)}
				{@const a = nodeById(act.from)}
				<line x1={a.x + a.w / 2} y1={a.y + a.h / 2} x2={act.x} y2={act.y} stroke="#2563eb" stroke-width="1.5" stroke-dasharray="5 4" />
			{/if}

			<!-- фігури -->
			{#each nodes as n (n.id)}
				<g class="cursor-move" onpointerdown={(ev) => nodeDown(ev, n)} ondblclick={() => { remember(); editId = n.id; }} role="presentation">
					{#if n.kind === 'process' || n.kind === 'subprogram'}
						<rect x={n.x} y={n.y} width={n.w} height={n.h} fill="#fdfdfd" stroke={isSel('node', n.id) ? '#2563eb' : '#222'} stroke-width={isSel('node', n.id) ? 2.5 : 1.5} />
						{#if n.kind === 'subprogram'}
							<line x1={n.x + 9} y1={n.y} x2={n.x + 9} y2={n.y + n.h} stroke="#222" stroke-width="1.5" />
							<line x1={n.x + n.w - 9} y1={n.y} x2={n.x + n.w - 9} y2={n.y + n.h} stroke="#222" stroke-width="1.5" />
						{/if}
					{:else if n.kind === 'terminator'}
						<rect x={n.x} y={n.y} width={n.w} height={n.h} rx={n.h / 2} fill="#fdfdfd" stroke={isSel('node', n.id) ? '#2563eb' : '#222'} stroke-width={isSel('node', n.id) ? 2.5 : 1.5} />
					{:else if n.kind === 'connector'}
						<circle cx={n.x + n.w / 2} cy={n.y + n.h / 2} r={Math.min(n.w, n.h) / 2} fill="#fdfdfd" stroke={isSel('node', n.id) ? '#2563eb' : '#222'} stroke-width={isSel('node', n.id) ? 2.5 : 1.5} />
					{:else}
						<polygon points={poly(n)} fill="#fdfdfd" stroke={isSel('node', n.id) ? '#2563eb' : '#222'} stroke-width={isSel('node', n.id) ? 2.5 : 1.5} />
					{/if}

					{#if editId === n.id}
						<foreignObject x={n.x} y={n.y + n.h / 2 - 12} width={n.w} height="24">
							<input value={n.text} oninput={(ev) => setText(n, ev.currentTarget.value)} onblur={() => (editId = null)} onkeydown={(ev) => ev.key === 'Enter' && (editId = null)} onpointerdown={(ev) => ev.stopPropagation()} class="h-full w-full border-0 bg-transparent text-center text-sm outline-none" autofocus />
						</foreignObject>
					{:else}
						<text x={n.x + n.w / 2} y={n.y + n.h / 2} text-anchor="middle" dominant-baseline="middle" fill="#111" pointer-events="none">{n.text}</text>
					{/if}

					<!-- порти для малювання стрілок (на виділеному блоці) -->
					{#if isSel('node', n.id) && n.kind !== 'connector'}
						{#each ports(n) as pt (pt.side)}
							<circle cx={pt.x} cy={pt.y} r="4.5" fill="#2563eb" class="cursor-crosshair" onpointerdown={(ev) => portDown(ev, n, pt.side)} role="presentation" />
						{/each}
					{/if}
				</g>
			{/each}
		</g>
	</svg>

	<p class="border-t border-slate-200 bg-white px-4 py-1.5 text-xs text-slate-400 dark:border-slate-700 dark:bg-slate-900">
		Тягни фон — рух полотна · колесо — масштаб · клік по стрілці — редагувати її вузли · сині порти — тягни нову стрілку · подвійний клік — текст · ✂ — розділити схему
	</p>
</div>
