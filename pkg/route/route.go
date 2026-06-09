// Пакет route — ортогональний маршрутизатор ребер з обходом перешкод.
// Прокладає шлях між двома точками по сітці Ганана (лінії крізь краї всіх
// перешкод + кінці) методом A* зі штрафом за повороти, тож лінії виходять
// прямими й не ріжуть фігури.
package route

import (
	"container/heap"
	"math"
	"sort"
)

type Pt struct{ X, Y float64 }

// Rect — перешкода (фігура).
type Rect struct{ X, Y, W, H float64 }

const (
	pad  = 12.0 // зазор навколо фігур
	turn = 24.0 // штраф за поворот (рівні лінії)
)

// Dir — напрямок виходу/входу ребра (зовнішня нормаль у точці приєднання).
type Dir int

const (
	None Dir = iota
	Up
	Down
	Left
	Right
)

// Connect прокладає шлях from→to так, щоб ребро ВИХОДило з джерела в напрямку
// fromDir і ВХОДило в ціль із напрямку toDir (короткі стаби назовні від фігур),
// а середину огинає перешкоди. Так стрілки заходять у фігури рівно з потрібного
// боку, а не «крізь себе».
func Connect(from, to Pt, fromDir, toDir Dir, obstacles []Rect) []Pt {
	const off = pad + 6
	a := step(from, fromDir, off)
	z := step(to, toDir, off)
	mid := Route(a, z, obstacles)
	path := append([]Pt{from}, mid...)
	path = append(path, to)
	return simplify(dedup(path))
}

func step(p Pt, d Dir, off float64) Pt {
	switch d {
	case Up:
		return Pt{p.X, p.Y - off}
	case Down:
		return Pt{p.X, p.Y + off}
	case Left:
		return Pt{p.X - off, p.Y}
	case Right:
		return Pt{p.X + off, p.Y}
	}
	return p
}

func dedup(p []Pt) []Pt {
	out := p[:0:0]
	for _, q := range p {
		if len(out) == 0 || absf(q.X-out[len(out)-1].X) > 0.5 || absf(q.Y-out[len(out)-1].Y) > 0.5 {
			out = append(out, q)
		}
	}
	return out
}

func absf(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Route прокладає ортогональний шлях a→b, огинаючи obstacles. Фігури, що
// містять a або b (джерело/ціль), у перешкоди не йдуть.
func Route(a, b Pt, obstacles []Rect) []Pt {
	// Обмежуємо перешкоди локальною зоною (bbox a..b + щедрий буфер). Інакше сітка
	// Ганана будується по ВСІХ фігурах програми — і A* вішає браузер на великих
	// схемах. У зону потрапляють лише 3-5 локальних перешкод, незалежно від розміру.
	const buf = 340.0 // ~2 типові ширини блоку — місце обійти фігуру по дузі
	loX, hiX := min(a.X, b.X)-buf, max(a.X, b.X)+buf
	loY, hiY := min(a.Y, b.Y)-buf, max(a.Y, b.Y)+buf
	var obs []Rect
	for _, r := range obstacles {
		if r.contains(a) || r.contains(b) {
			continue
		}
		if r.X+r.W < loX || r.X > hiX || r.Y+r.H < loY || r.Y > hiY {
			continue // повністю поза зоною — не перешкода
		}
		obs = append(obs, r.inflate(pad))
	}

	xs := axis(a.X, b.X, obs, true)
	ys := axis(a.Y, b.Y, obs, false)
	ix, iy := index(xs, a.X), index(ys, a.Y)
	gx, gy := index(xs, b.X), index(ys, b.Y)
	nx, ny := len(xs), len(ys)

	// Стан = (вузол сітки, вісь останнього руху: 0 старт, 1 гориз, 2 верт).
	id := func(i, j, ax int) int { return (j*nx+i)*3 + ax }
	start := id(ix, iy, 0)

	dist := map[int]float64{start: 0}
	prev := map[int]int{}
	open := &pq{{state: start, f: 0}}
	seen := map[int]bool{}

	free := func(i1, j1, i2, j2 int) bool {
		return !blocked(Pt{xs[i1], ys[j1]}, Pt{xs[i2], ys[j2]}, obs)
	}

	var endState = -1
	for open.Len() > 0 {
		cur := heap.Pop(open).(node)
		if seen[cur.state] {
			continue
		}
		seen[cur.state] = true
		i, j, ax := (cur.state/3)%nx, (cur.state/3)/nx, cur.state%3
		if i == gx && j == gy {
			endState = cur.state
			break
		}
		d := dist[cur.state]
		// сусіди по 4 напрямках
		for _, s := range [4][3]int{{1, 0, 1}, {-1, 0, 1}, {0, 1, 2}, {0, -1, 2}} {
			ni, nj, nax := i+s[0], j+s[1], s[2]
			if ni < 0 || ni >= nx || nj < 0 || nj >= ny || !free(i, j, ni, nj) {
				continue
			}
			cost := math.Abs(xs[ni]-xs[i]) + math.Abs(ys[nj]-ys[j])
			if ax != 0 && ax != nax {
				cost += turn
			}
			ns := id(ni, nj, nax)
			if nd := d + cost; !seen[ns] {
				if old, ok := dist[ns]; !ok || nd < old {
					dist[ns] = nd
					prev[ns] = cur.state
					h := math.Abs(xs[gx]-xs[ni]) + math.Abs(ys[gy]-ys[nj])
					heap.Push(open, node{state: ns, f: nd + h})
				}
			}
		}
	}

	if endState < 0 { // не знайшли — пряма ламана (фолбек)
		return []Pt{a, {b.X, a.Y}, b}
	}
	// відновлюємо й спрощуємо
	var path []Pt
	for s := endState; ; s = prev[s] {
		i, j := (s/3)%nx, (s/3)/nx
		path = append(path, Pt{xs[i], ys[j]})
		if s == start {
			break
		}
	}
	reverse(path)
	return simplify(path)
}

func (r Rect) contains(p Pt) bool {
	return p.X >= r.X-1 && p.X <= r.X+r.W+1 && p.Y >= r.Y-1 && p.Y <= r.Y+r.H+1
}
func (r Rect) inflate(d float64) Rect { return Rect{r.X - d, r.Y - d, r.W + 2*d, r.H + 2*d} }

// blocked — чи axis-aligned відрізок a-b перетинає внутрішність якоїсь перешкоди.
func blocked(a, b Pt, obs []Rect) bool {
	mm := func(u, v float64) (float64, float64) {
		if u < v {
			return u, v
		}
		return v, u
	}
	minX, maxX := mm(a.X, b.X)
	minY, maxY := mm(a.Y, b.Y)
	for _, r := range obs {
		if maxX <= r.X || minX >= r.X+r.W || maxY <= r.Y || minY >= r.Y+r.H {
			continue
		}
		return true
	}
	return false
}

// axis збирає координатні лінії (краї перешкод + кінці) одного напрямку.
func axis(a, b float64, obs []Rect, horiz bool) []float64 {
	vals := []float64{a, b}
	for _, r := range obs {
		if horiz {
			vals = append(vals, r.X, r.X+r.W)
		} else {
			vals = append(vals, r.Y, r.Y+r.H)
		}
	}
	sort.Float64s(vals)
	out := vals[:0:0]
	for _, v := range vals {
		if len(out) == 0 || math.Abs(v-out[len(out)-1]) > 0.5 {
			out = append(out, v)
		}
	}
	return out
}

func index(xs []float64, v float64) int {
	for i, x := range xs {
		if math.Abs(x-v) <= 0.5 {
			return i
		}
	}
	return 0
}

func reverse(p []Pt) {
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
}

// simplify прибирає колінеарні проміжні точки.
func simplify(p []Pt) []Pt {
	if len(p) < 3 {
		return p
	}
	out := p[:1]
	for i := 1; i < len(p)-1; i++ {
		a, b, c := p[i-1], p[i], p[i+1]
		collinear := (math.Abs(a.X-b.X) < 0.5 && math.Abs(b.X-c.X) < 0.5) ||
			(math.Abs(a.Y-b.Y) < 0.5 && math.Abs(b.Y-c.Y) < 0.5)
		if !collinear {
			out = append(out, b)
		}
	}
	return append(out, p[len(p)-1])
}

// --- мінімальна купа для A* ---

type node struct {
	state int
	f     float64
}
type pq []node

func (q pq) Len() int           { return len(q) }
func (q pq) Less(i, j int) bool { return q[i].f < q[j].f }
func (q pq) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *pq) Push(x any)        { *q = append(*q, x.(node)) }
func (q *pq) Pop() any          { old := *q; n := len(old); it := old[n-1]; *q = old[:n-1]; return it }
