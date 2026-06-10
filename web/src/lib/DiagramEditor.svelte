<script>
	// Візуальний редактор схеми (Фаза 2): безмежне полотно (пан+зум), редаговані
	// стрілки (вибір, тягання вузлів і кінців, переприв'язка, малювання нових),
	// авто-ортогональний рероут при русі блоків, розділення конектором.
	// diagrams — масив {name, diagram} (усі функції коду на одному полотні);
	// diagram — застарілий одиночний вхід (огортається в масив).
	let { diagrams = null, diagram = null, onsave, oncancel } = $props();

	const MARGIN = 24;
	// Кольори рамок груп (кожна функція/схема — своя група).
	const GROUP_COLORS = ['#2563eb', '#16a34a', '#9333ea', '#ea580c', '#0891b2', '#db2777', '#65a30d', '#e11d48'];
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
	let groups = $state([]); // [{id, name, color}] — групи = окремі схеми
	let selNodes = $state(new Set()); // мультивибір вузлів (id)
	let lasso = $state(null); // {x0,y0,x1,y1} під час рамки-ласо
	let sel = $state(null); // {type:'node'|'edge', id}
	let editId = $state(null); // фігура, чий текст редагуємо
	let editLabel = $state(null); // ребро, чий підпис (Так/Ні) редагуємо
	let guides = $state([]); // напрямні-магніти під час перетягування
	let view = $state({ x: 40, y: 40, scale: 1 });
	let gEl; // внутрішня <g> з трансформацією (для коректного перерахунку координат)
	let nid = 0,
		eid = 0,
		gseq = 0; // лічильник унікальних id груп

	// Чи вузол у виділенні (мультивибір АБО одиночний primary).
	const isNodeSel = (id) => selNodes.has(id) || (sel?.type === 'node' && sel.id === id);

	// --- історія (undo/redo) ---
	let past = [];
	let futureH = [];
	const snap = () => JSON.stringify({ nodes, edges, groups });
	function remember() {
		past.push(snap());
		if (past.length > 60) past.shift();
		futureH = [];
	}
	function restore(s) {
		const o = JSON.parse(s);
		nodes = o.nodes;
		edges = o.edges;
		groups = o.groups ?? groups;
		sel = null;
		selNodes = new Set();
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

	$effect(() => load(diagrams ?? diagram));

	// load — усі схеми на одне полотно: кожна зміщується праворуч у власну колонку,
	// її блоки дістають group = id групи (= окрема вихідна схема).
	function load(input) {
		const list = Array.isArray(input) ? input : input ? [{ name: '', diagram: input }] : [];
		let k = 0, j = 0, gx = 0;
		const ns = [], es = [], grps = [];
		for (let gi = 0; gi < list.length; gi++) {
			const it = list[gi];
			const d = it.diagram ?? it;
			const gid = 'g' + gi;
			grps.push({ id: gid, name: it.name ?? '', color: GROUP_COLORS[gi % GROUP_COLORS.length] });
			const offX = gx;
			const gn = (d.shapes ?? []).map((s) => ({ id: 'n' + k++, group: gid, kind: s.kind, x: s.x + offX, y: s.y, w: s.w, h: s.h, text: s.text ?? '' }));
			const at = (pt) => {
				if (!pt) return null;
				for (const n of gn) if (pt.x >= n.x - 8 && pt.x <= n.x + n.w + 8 && pt.y >= n.y - 8 && pt.y <= n.y + n.h + 8) return n.id;
				return null;
			};
			const ge = (d.edges ?? []).map((e) => {
				const pts = (e.points ?? []).map((p) => ({ x: p.x + offX, y: p.y }));
				const lp = labelAnchor(pts[0], pts[1]);
				return { id: 'e' + j++, points: pts, label: e.label ?? '', lx: lp.x, ly: lp.y, arrowless: !!e.arrowless, fromId: at(pts[0]), toId: at(pts[pts.length - 1]), manual: false };
			});
			ns.push(...gn); es.push(...ge);
			gx += (d.w ?? 400) + 80;
		}
		nid = k; eid = j; gseq = list.length;
		nodes = ns; edges = es; groups = grps;
		sel = null; selNodes = new Set(); editId = null;
	}

	const nodeById = (id) => nodes.find((n) => n.id === id);

	// Рамки груп: bbox блоків кожної групи + підпис («Рисунок N» або імʼя функції).
	let groupFrames = $derived.by(() => {
		if (groups.length <= 1) return []; // одна схема — рамка зайва
		const m = new Map();
		for (const n of nodes) {
			let f = m.get(n.group);
			if (!f) { f = { group: n.group, minX: Infinity, minY: Infinity, maxX: -Infinity, maxY: -Infinity }; m.set(n.group, f); }
			f.minX = Math.min(f.minX, n.x); f.minY = Math.min(f.minY, n.y);
			f.maxX = Math.max(f.maxX, n.x + n.w); f.maxY = Math.max(f.maxY, n.y + n.h);
		}
		const PAD = 16;
		return [...m.values()].map((f) => {
			const gi = groups.findIndex((g) => g.id === f.group);
			const g = groups[gi];
			return {
				color: g?.color ?? '#94a3b8',
				name: g?.name || `Рисунок ${gi + 1}`,
				x: f.minX - PAD, y: f.minY - PAD - 22,
				w: f.maxX - f.minX + 2 * PAD, h: f.maxY - f.minY + 2 * PAD + 22
			};
		});
	});

	// Позиція підпису ребра (Так/Ні) — як labelAnchor у @rombik/engine (далі від ромба).
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
		if (act) return;
		// Shift+тяг фону — рамка-ласо (мультивибір); звичайний тяг — пан + зняти виділення.
		if (ev.shiftKey) {
			const w = toWorld(ev.clientX, ev.clientY);
			act = { kind: 'lasso', pid: ev.pointerId };
			lasso = { x0: w.x, y0: w.y, x1: w.x, y1: w.y };
			return;
		}
		sel = null;
		selNodes = new Set();
		editId = null;
		act = { kind: 'pan', pid: ev.pointerId, sx: ev.clientX, sy: ev.clientY, vx: view.x, vy: view.y };
	}

	function nodeDown(ev, n) {
		ev.stopPropagation();
		if (act) return;
		if (editId && editId !== n.id) editId = null;
		// Shift-клік — лише перемкнути вузол у мультивиборі (без тягання).
		if (ev.shiftKey) {
			const s = new Set(selNodes);
			s.has(n.id) ? s.delete(n.id) : s.add(n.id);
			selNodes = s;
			sel = s.size ? { type: 'node', id: n.id } : null;
			return;
		}
		// Клік по вузлу поза мультивибором — почати з нього (інакше тягнемо всю групу).
		if (!selNodes.has(n.id)) selNodes = new Set([n.id]);
		sel = { type: 'node', id: n.id };
		const w = toWorld(ev.clientX, ev.clientY);
		act = { kind: 'node', pid: ev.pointerId, id: n.id, ox: n.x, oy: n.y, wx: w.x, wy: w.y };
	}

	function edgeDown(ev, e) {
		ev.stopPropagation();
		sel = { type: 'edge', id: e.id };
	}

	function labelDown(ev, e) {
		ev.stopPropagation();
		if (act) return;
		sel = { type: 'edge', id: e.id };
		const w = toWorld(ev.clientX, ev.clientY);
		act = { kind: 'label', pid: ev.pointerId, id: e.id, ox: e.lx, oy: e.ly, wx: w.x, wy: w.y };
	}

	function vertexDown(ev, e, i) {
		ev.stopPropagation();
		if (act) return;
		sel = { type: 'edge', id: e.id };
		act = { kind: 'vertex', pid: ev.pointerId, id: e.id, i, end: i === 0 || i === e.points.length - 1 };
		e.manual = true;
	}

	function portDown(ev, n, side) {
		ev.stopPropagation();
		if (act) return;
		const w = toWorld(ev.clientX, ev.clientY);
		act = { kind: 'draw', pid: ev.pointerId, from: n.id, side, x: w.x, y: w.y };
	}

	function onMove(ev) {
		if (!act || (act.pid !== undefined && ev.pointerId !== act.pid)) return;
		// Знімок в історію — лише при ПЕРШОМУ русі (не на простий клік-вибір).
		if (!act.saved && (act.kind === 'node' || act.kind === 'vertex' || act.kind === 'label')) {
			remember();
			act.saved = true;
		}
		const w = toWorld(ev.clientX, ev.clientY);
		if (act.kind === 'pan') {
			view.x = act.vx + (ev.clientX - act.sx);
			view.y = act.vy + (ev.clientY - act.sy);
		} else if (act.kind === 'lasso') {
			lasso = { ...lasso, x1: w.x, y1: w.y };
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
			// мультивибір (>1) — рухаємо всю групу разом; інакше лише цей вузол
			const moving = selNodes.has(n.id) && selNodes.size > 1 ? selNodes : new Set([n.id]);
			for (const id of moving) {
				const nn = nodeById(id);
				if (!nn) continue;
				nn.x += dx;
				nn.y += dy;
				rerouteFor(id, dx, dy);
			}
			nodes = [...nodes];
			edges = [...edges];
		} else if (act.kind === 'label') {
			const e = edges.find((x) => x.id === act.id);
			e.lx = act.ox + (w.x - act.wx);
			e.ly = act.oy + (w.y - act.wy);
			edges = [...edges];
		} else if (act.kind === 'vertex') {
			const e = edges.find((x) => x.id === act.id);
			let nx = w.x, ny = w.y;
			const T = 10 / view.scale;
			guides = [];
			// 1) 90 degree snap with prev/next point
			if (act.i > 0) {
				const p = e.points[act.i - 1];
				if (Math.abs(nx - p.x) < T) { nx = p.x; guides.push({ t: 'v', v: nx }); }
				if (Math.abs(ny - p.y) < T) { ny = p.y; guides.push({ t: 'h', v: ny }); }
			}
			if (act.i < e.points.length - 1) {
				const p = e.points[act.i + 1];
				if (Math.abs(nx - p.x) < T) { nx = p.x; guides.push({ t: 'v', v: nx }); }
				if (Math.abs(ny - p.y) < T) { ny = p.y; guides.push({ t: 'h', v: ny }); }
			}
			// 2) Snap to node centers
			for (const o of nodes) {
				const cX = o.x + o.w / 2, cY = o.y + o.h / 2;
				if (Math.abs(nx - cX) < T) { nx = cX; guides.push({ t: 'v', v: nx }); }
				if (Math.abs(ny - cY) < T) { ny = cY; guides.push({ t: 'h', v: ny }); }
			}
			e.points[act.i] = { x: nx, y: ny };
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
		if (!act || (act.pid !== undefined && ev.pointerId !== act.pid)) return;
		const w = toWorld(ev.clientX, ev.clientY);
		if (act.kind === 'lasso') {
			// вибрати вузли, чий центр у прямокутнику
			const lx0 = Math.min(lasso.x0, lasso.x1), lx1 = Math.max(lasso.x0, lasso.x1);
			const ly0 = Math.min(lasso.y0, lasso.y1), ly1 = Math.max(lasso.y0, lasso.y1);
			const s = new Set();
			for (const n of nodes) {
				const cx = n.x + n.w / 2, cy = n.y + n.h / 2;
				if (cx >= lx0 && cx <= lx1 && cy >= ly0 && cy <= ly1) s.add(n.id);
			}
			selNodes = s;
			sel = s.size ? { type: 'node', id: [...s][0] } : null;
			lasso = null;
			act = null;
			return;
		}
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
		// мультивибір — видалити всі виділені вузли та дотичні ребра
		if (selNodes.size > 1) {
			remember();
			const ids = selNodes;
			nodes = nodes.filter((n) => !ids.has(n.id));
			edges = edges.filter((e) => !ids.has(e.fromId) && !ids.has(e.toId));
			selNodes = new Set();
			sel = null;
			return;
		}
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
		// нащадки вперед: додаємо тільки ті блоки, в які можна потрапити ЛИШЕ через цей вузол (strict descendants)
		const set = new Set([nodeId]);
		let changed = true;
		while (changed) {
			changed = false;
			for (const e of edges) {
				if (e.fromId && e.toId && set.has(e.fromId) && !set.has(e.toId)) {
					let hasOutsideIncoming = false;
					for (const inE of edges) {
						if (inE.toId === e.toId && !set.has(inE.fromId)) {
							hasOutsideIncoming = true;
							break;
						}
					}
					if (!hasOutsideIncoming) {
						set.add(e.toId);
						changed = true;
					}
				}
			}
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

	// Виділені вузли → окрема група (= окрема вихідна схема). Порожні групи прибираємо.
	function groupSelection() {
		if (selNodes.size < 1) return;
		remember();
		const k = gseq++;
		const gid = 'g' + k;
		groups.push({ id: gid, name: '', color: GROUP_COLORS[k % GROUP_COLORS.length] });
		for (const n of nodes) if (selNodes.has(n.id)) n.group = gid;
		const used = new Set(nodes.map((n) => n.group));
		groups = groups.filter((g) => used.has(g.id));
		nodes = [...nodes];
		selNodes = new Set();
		sel = null;
	}

	// nodeIdAt/groupOfPoint — до якої групи (схеми) належить точка ребра. Без авто-
	// детекції компонентів: групи задає користувач, тут лише розкладаємо ребра по них.
	function nodeIdAt(pt) {
		for (const n of nodes) if (pt.x >= n.x - 8 && pt.x <= n.x + n.w + 8 && pt.y >= n.y - 8 && pt.y <= n.y + n.h + 8) return n.id;
		return null;
	}
	function groupOfPoint(pt) {
		const id = nodeIdAt(pt);
		if (id) return nodeById(id).group;
		let best = null, bd = Infinity; // junction (стик) — найближча за центром фігура
		for (const n of nodes) { const d = (pt.x - (n.x + n.w / 2)) ** 2 + (pt.y - (n.y + n.h / 2)) ** 2; if (d < bd) { bd = d; best = n.group; } }
		return best;
	}

	// save — одна вихідна схема НА ГРУПУ. Жодної авто-фрагментації: розкладка строго
	// за group вузлів. Ребро йде в групу свого кінця.
	function save() {
		const buckets = new Map();
		const bucket = (g) => { if (!buckets.has(g)) buckets.set(g, { nodes: [], edges: [] }); return buckets.get(g); };
		for (const n of nodes) bucket(n.group).nodes.push(n);
		for (const e of edges) {
			const g = groupOfPoint(e.points[0]) ?? groupOfPoint(e.points[e.points.length - 1]);
			if (g != null) bucket(g).edges.push(e);
		}
		const out = [...buckets.values()].filter((b) => b.nodes.length).map((b) => {
			let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
			const acc = (x, y) => { minX = Math.min(minX, x); minY = Math.min(minY, y); maxX = Math.max(maxX, x); maxY = Math.max(maxY, y); };
			for (const n of b.nodes) { acc(n.x, n.y); acc(n.x + n.w, n.y + n.h); }
			for (const e of b.edges) for (const p of e.points) acc(p.x, p.y);
			if (!isFinite(minX)) { minX = minY = 0; maxX = maxY = 100; }
			const dx = MARGIN - minX, dy = MARGIN - minY;
			return {
				_minX: minX,
				shapes: b.nodes.map((n) => ({ kind: n.kind, x: n.x + dx, y: n.y + dy, w: n.w, h: n.h, text: n.text })),
				edges: b.edges.map((e) => ({ points: e.points.map((p) => ({ x: p.x + dx, y: p.y + dy })), label: e.label || undefined, arrowless: e.arrowless || undefined })),
				w: maxX - minX + 2 * MARGIN,
				h: maxY - minY + 2 * MARGIN
			};
		}).sort((a, b) => a._minX - b._minX).map(({ _minX, ...rest }) => rest);
		onsave?.(out.length === 1 ? out[0] : out);
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
	<div class="flex flex-col gap-2 border-b border-slate-200 bg-white px-3 py-2 shadow-sm dark:border-slate-700 dark:bg-slate-900 sm:flex-row sm:items-center">
		<!-- Верхній ряд на мобілці: назва та головні кнопки -->
		<div class="flex items-center justify-between sm:w-auto">
			<span class="text-sm font-bold text-slate-700 dark:text-slate-200">Редактор схеми</span>
			<div class="flex items-center gap-1.5 sm:hidden">
				<button onclick={() => oncancel?.()} class="rounded-lg border border-slate-300 px-3 py-1 text-xs text-slate-700 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-200 dark:hover:bg-slate-800">✕ Вийти</button>
				<button onclick={save} class="rounded-lg bg-blue-600 px-3 py-1 text-xs font-semibold text-white transition hover:bg-blue-700">Зберегти</button>
			</div>
		</div>

		<!-- Прокручуваний ряд кнопок для фігур -->
		<div class="flex flex-1 items-center gap-1 overflow-x-auto pb-1 sm:pb-0 scrollbar-hide">
			{#each PALETTE as [kind, label, w, h] (kind)}
				<button onclick={() => addNode(kind, w, h)} class="shrink-0 rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800">+ {label}</button>
			{/each}
			<div class="mx-1 h-5 w-px shrink-0 bg-slate-200 dark:bg-slate-700"></div>
			{#if selNodes.size > 0}
				<button onclick={groupSelection} class="shrink-0 rounded border border-emerald-300 px-2 py-1 text-xs font-medium text-emerald-700 transition hover:bg-emerald-50 dark:border-emerald-800 dark:text-emerald-400 dark:hover:bg-emerald-950">⊞ Окрема схема ({selNodes.size})</button>
			{/if}
			{#if sel?.type === 'node' && selNodes.size <= 1 && nodeById(sel.id)?.kind !== 'connector'}
				<button onclick={() => splitAt(sel.id)} class="shrink-0 rounded border border-amber-300 px-2 py-1 text-xs font-medium text-amber-700 transition hover:bg-amber-50 dark:border-amber-800 dark:text-amber-400 dark:hover:bg-amber-950">✂ Розділити</button>
			{/if}
			<button onclick={delSel} disabled={!sel && selNodes.size === 0} class="shrink-0 rounded border border-red-200 px-2 py-1 text-xs font-medium text-red-600 transition hover:bg-red-50 disabled:opacity-40 dark:border-red-900 dark:text-red-400 dark:hover:bg-red-950">Видалити</button>
			<button onclick={undo} title="Скасувати" class="grid h-7 w-7 shrink-0 place-items-center rounded border border-slate-300 text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800">↶</button>
			<button onclick={redo} title="Повторити" class="grid h-7 w-7 shrink-0 place-items-center rounded border border-slate-300 text-slate-600 transition hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800">↷</button>
		</div>

		<!-- Зум і зберегти (тільки для десктопа) -->
		<div class="hidden shrink-0 items-center gap-1.5 sm:flex">
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
	<!-- полотно ЗАВЖДИ світле (схема — чорне-по-білому; темна лише обгортка) -->
	<svg class="min-h-0 flex-1 touch-none bg-white text-slate-900" style="font-family: 'Times New Roman', 'Liberation Serif', 'DejaVu Serif', serif" onpointerdown={bgDown} onwheel={onWheel} role="presentation">
		<defs>
			<marker id="ed-arr" markerWidth="9" markerHeight="9" refX="7.5" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6 Z" fill="currentColor" /></marker>
			<pattern id="ed-grid" width="28" height="28" patternUnits="userSpaceOnUse" patternTransform="translate({view.x},{view.y}) scale({view.scale})">
				<path d="M28 0H0V28" fill="none" class="stroke-slate-200" stroke-width="1" />
			</pattern>
		</defs>
		<rect width="100%" height="100%" fill="url(#ed-grid)" />
		<g bind:this={gEl} transform="translate({view.x},{view.y}) scale({view.scale})">
			<!-- рамки груп (кожна = окрема вихідна схема) -->
			{#each groupFrames as gf (gf.name + ',' + gf.x)}
				<rect x={gf.x} y={gf.y} width={gf.w} height={gf.h} rx="12" fill={gf.color} fill-opacity="0.04" stroke={gf.color} stroke-width={1.5 / view.scale} stroke-dasharray="8 5" opacity="0.75" pointer-events="none" />
				<text x={gf.x + 12} y={gf.y + 17} font-size="13" font-weight="bold" fill={gf.color} pointer-events="none">{gf.name}</text>
			{/each}
			<!-- рамка-ласо -->
			{#if lasso}
				<rect x={Math.min(lasso.x0, lasso.x1)} y={Math.min(lasso.y0, lasso.y1)} width={Math.abs(lasso.x1 - lasso.x0)} height={Math.abs(lasso.y1 - lasso.y0)} fill="#2563eb" fill-opacity="0.08" stroke="#2563eb" stroke-width={1 / view.scale} stroke-dasharray="5 4" pointer-events="none" />
			{/if}
			<!-- ребра -->
			{#each edges as e (e.id)}
				<!-- товста невидима «хіт-зона» для зручного кліку -->
				<path d={pathD(e.points)} fill="none" stroke="transparent" stroke-width="12" class="cursor-pointer" onpointerdown={(ev) => edgeDown(ev, e)} role="presentation" />
				<path d={pathD(e.points)} fill="none" stroke={isSel('edge', e.id) ? '#2563eb' : 'currentColor'} stroke-width={isSel('edge', e.id) ? 2.5 : 1.5} marker-end={e.arrowless ? '' : 'url(#ed-arr)'} pointer-events="none" />
				{#if editLabel === e.id}
					<foreignObject x={e.lx - 6} y={e.ly - 16} width="90" height="22">
				<!-- svelte-ignore a11y_autofocus -->
				<input value={e.label} oninput={(ev) => { e.label = ev.currentTarget.value; edges = [...edges]; }} onblur={() => (editLabel = null)} onkeydown={(ev) => ev.key === 'Enter' && (editLabel = null)} onpointerdown={(ev) => ev.stopPropagation()} class="w-full rounded border border-blue-300 bg-white px-1 text-xs text-slate-900 outline-none" autofocus />
					</foreignObject>
				{:else if e.label}
					<text x={e.lx} y={e.ly} text-anchor="middle" font-size="12" fill={isSel('edge', e.id) ? '#2563eb' : 'currentColor'} class="cursor-move" onpointerdown={(ev) => labelDown(ev, e)} ondblclick={() => (editLabel = e.id)} role="presentation">{e.label}</text>
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
						<rect x={n.x} y={n.y} width={n.w} height={n.h} class="fill-white" stroke={isNodeSel(n.id) ? '#2563eb' : 'currentColor'} stroke-width={isNodeSel(n.id) ? 2.5 : 1.5} />
						{#if n.kind === 'subprogram'}
							<line x1={n.x + 9} y1={n.y} x2={n.x + 9} y2={n.y + n.h} stroke="currentColor" stroke-width="1.5" />
							<line x1={n.x + n.w - 9} y1={n.y} x2={n.x + n.w - 9} y2={n.y + n.h} stroke="currentColor" stroke-width="1.5" />
						{/if}
					{:else if n.kind === 'terminator'}
						<rect x={n.x} y={n.y} width={n.w} height={n.h} rx={n.h / 2} class="fill-white" stroke={isNodeSel(n.id) ? '#2563eb' : 'currentColor'} stroke-width={isNodeSel(n.id) ? 2.5 : 1.5} />
					{:else if n.kind === 'connector'}
						<circle cx={n.x + n.w / 2} cy={n.y + n.h / 2} r={Math.min(n.w, n.h) / 2} class="fill-white" stroke={isNodeSel(n.id) ? '#2563eb' : 'currentColor'} stroke-width={isNodeSel(n.id) ? 2.5 : 1.5} />
					{:else}
						<polygon points={poly(n)} class="fill-white" stroke={isNodeSel(n.id) ? '#2563eb' : 'currentColor'} stroke-width={isNodeSel(n.id) ? 2.5 : 1.5} />
					{/if}

					{#if editId === n.id}
						<foreignObject x={n.x} y={n.y + n.h / 2 - 12} width={n.w} height="24">
							<!-- svelte-ignore a11y_autofocus -->
							<input value={n.text} oninput={(ev) => setText(n, ev.currentTarget.value)} onblur={() => (editId = null)} onkeydown={(ev) => ev.key === 'Enter' && (editId = null)} onpointerdown={(ev) => ev.stopPropagation()} class="h-full w-full border-0 bg-transparent text-center text-sm text-slate-900 outline-none" autofocus />
						</foreignObject>
					{:else}
						<text x={n.x + n.w / 2} y={n.y + n.h / 2} text-anchor="middle" dominant-baseline="middle" fill="currentColor" pointer-events="none">{n.text}</text>
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

	<!-- Плаваючі кнопки зуму для мобільного -->
	<div class="fixed bottom-12 right-4 flex flex-col gap-2 sm:hidden">
		<button onclick={() => zoom(1.2)} class="grid h-10 w-10 place-items-center rounded-full bg-white shadow-lg border border-slate-200 text-lg font-bold text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200">+</button>
		<button onclick={() => zoom(1 / 1.2)} class="grid h-10 w-10 place-items-center rounded-full bg-white shadow-lg border border-slate-200 text-lg font-bold text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200">−</button>
		<button onclick={fit} class="grid h-10 w-10 place-items-center rounded-full bg-white shadow-lg border border-slate-200 text-[10px] font-bold text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200">Фіт</button>
	</div>

	<p class="border-t border-slate-200 bg-white px-4 py-1.5 text-xs text-slate-400 dark:border-slate-700 dark:bg-slate-900">
		Тягни фон — рух полотна · Shift+тяг — рамка-ласо · Shift+клік — додати у вибір · ⊞ — виділене в окрему схему · колесо — масштаб · подвійний клік — текст · ✂ — розділити схему
	</p>
</div>
