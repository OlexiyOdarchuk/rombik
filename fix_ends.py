with open("pkg/layout/layout.go", "r") as f:
    content = f.read()

old_func = """func (b *build) routeEnds(cx, kY float64) {
	if len(b.ends) == 0 {
		return
	}
	if len(b.ends) == 1 { // один вихід — прямо в Кінець (з невеликим вигином, якщо не по осі)
		e := b.ends[0]
		if e.X > cx-1 && e.X < cx+1 {
			b.d.Edges = append(b.d.Edges, edge(e, P(cx, kY)))
		} else {
			my := e.Y + mergeGap/2
			b.d.Edges = append(b.d.Edges, diagram.Edge{Points: []diagram.Point{e, {X: e.X, Y: my}, {X: cx, Y: my}, {X: cx, Y: kY}}})
		}
		return
	}
	right, _ := b.bodyExtent(0, 0, cx) // найправіша точка всього графа
	busX := right + mergeGap           // коридор шини — гарантовано вільний
	jy := kY - vGap                    // точка збору над Кінцем
	minDrop := jy
	for _, e := range b.ends {
		drop := e.Y + mergeGap/2
		if drop < minDrop {
			minDrop = drop
		}
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
			e, {X: e.X, Y: drop}, {X: busX, Y: drop}, // вниз з-під блоку, тоді вправо до шини
		}})
	}
	b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
		{X: busX, Y: minDrop}, {X: busX, Y: jy}, {X: cx, Y: jy}, // шина вниз → у центр
	}})
	b.d.Edges = append(b.d.Edges, edge(P(cx, jy), P(cx, kY))) // → у Кінець (зі стрілкою)
}"""

new_func = """func (b *build) routeEnds(cx, kY float64) {
	if len(b.ends) == 0 {
		return
	}
	if len(b.ends) == 1 {
		e := b.ends[0]
		if e.X > cx-1 && e.X < cx+1 {
			b.d.Edges = append(b.d.Edges, edge(e, P(cx, kY)))
		} else {
			my := e.Y + mergeGap/2
			b.d.Edges = append(b.d.Edges, diagram.Edge{Points: []diagram.Point{e, {X: e.X, Y: my}, {X: cx, Y: my}, {X: cx, Y: kY}}})
		}
		return
	}
	right, left := b.bodyExtent(0, 0, cx)
	busRight := right + mergeGap
	busLeft := left - mergeGap
	jy := kY - vGap
	minDropR := jy
	minDropL := jy
	hasR, hasL := false, false

	for _, e := range b.ends {
		drop := e.Y + mergeGap/2
		if e.X < cx-1 {
			if drop < minDropL {
				minDropL = drop
			}
			b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
				e, {X: e.X, Y: drop}, {X: busLeft, Y: drop},
			}})
			hasL = true
		} else {
			if drop < minDropR {
				minDropR = drop
			}
			b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
				e, {X: e.X, Y: drop}, {X: busRight, Y: drop},
			}})
			hasR = true
		}
	}

	if hasL {
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
			{X: busLeft, Y: minDropL}, {X: busLeft, Y: jy}, {X: cx, Y: jy},
		}})
	}
	if hasR {
		b.d.Edges = append(b.d.Edges, diagram.Edge{Arrowless: true, Points: []diagram.Point{
			{X: busRight, Y: minDropR}, {X: busRight, Y: jy}, {X: cx, Y: jy},
		}})
	}
	b.d.Edges = append(b.d.Edges, edge(P(cx, jy), P(cx, kY)))
}"""

content = content.replace(old_func, new_func)

with open("pkg/layout/layout.go", "w") as f:
    f.write(content)
