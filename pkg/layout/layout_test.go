package layout

import (
	"testing"

	"github.com/OlexiyOdarchuk/rombik/pkg/diagram"
	"github.com/OlexiyOdarchuk/rombik/pkg/ir"
)

func countText(d *diagram.Diagram, text string) int {
	n := 0
	for _, s := range d.Shapes {
		if s.Text == text {
			n++
		}
	}
	return n
}

func countKind(d *diagram.Diagram, k diagram.Kind) int {
	n := 0
	for _, s := range d.Shapes {
		if s.Kind == k {
			n++
		}
	}
	return n
}

func hasEdgeLabel(d *diagram.Diagram, label string) bool {
	for _, e := range d.Edges {
		if e.Label == label {
			return true
		}
	}
	return false
}

func TestBuildBasic(t *testing.T) {
	prog := ir.NewBlock(&ir.IO{Text: "Ввід n"}, &ir.Process{Text: "s = n * n"})
	d := Build(prog, Options{})
	// Початок + Ввід + Процес + Кінець
	if got := len(d.Shapes); got != 4 {
		t.Errorf("фігур = %d, очікувалось 4", got)
	}
	if countText(d, "Початок") != 1 || countText(d, "Кінець") != 1 {
		t.Error("має бути рівно один Початок і один Кінець")
	}
	if d.W <= 0 || d.H <= 0 {
		t.Error("полотно має ненульовий розмір")
	}
}

func TestPerExitVsSingleEnd(t *testing.T) {
	prog := ir.NewBlock(&ir.If{
		Cond: "x > 0",
		Then: ir.NewBlock(&ir.Terminal{Text: "Повернути 1"}),
		Else: ir.NewBlock(&ir.Terminal{Text: "Повернути 2"}),
	})
	perExit := countText(Build(prog, Options{}), "Кінець")
	single := countText(Build(prog, Options{SingleEnd: true}), "Кінець")
	if perExit != 2 {
		t.Errorf("per-exit: очікувалось 2 Кінці, маємо %d", perExit)
	}
	if single != 1 {
		t.Errorf("single-end: очікувався 1 Кінець, маємо %d", single)
	}
}

func TestBranchLabels(t *testing.T) {
	prog := ir.NewBlock(&ir.If{
		Cond: "x",
		Then: ir.NewBlock(&ir.Process{Text: "a"}),
		Else: ir.NewBlock(&ir.Process{Text: "b"}),
	})
	d := Build(prog, Options{Yes: "Yes", No: "No"})
	if !hasEdgeLabel(d, "Yes") || !hasEdgeLabel(d, "No") {
		t.Error("мітки гілок Yes/No не застосувались")
	}
}

func TestShapesForConstructs(t *testing.T) {
	prog := ir.NewBlock(
		&ir.For{Spec: "i = 0, 4, 1", Body: ir.NewBlock(&ir.Process{Text: "s += i"})},
		&ir.While{Cond: "k < 5", Body: ir.NewBlock(&ir.Process{Text: "k += 1"})},
	)
	d := Build(prog, Options{})
	if countKind(d, diagram.Hexagon) != 1 {
		t.Error("for має давати один шестикутник")
	}
	if countKind(d, diagram.Decision) != 1 {
		t.Error("while має давати один ромб")
	}
}

func TestCallAsProcess(t *testing.T) {
	prog := ir.NewBlock(&ir.Call{Text: "f(x)"})
	if countKind(Build(prog, Options{}), diagram.Predef) != 1 {
		t.Error("виклик має бути символом підпрограми")
	}
	if countKind(Build(prog, Options{CallAsProcess: true}), diagram.Predef) != 0 {
		t.Error("з CallAsProcess виклик має стати звичайним блоком")
	}
}

func TestStripTypes(t *testing.T) {
	prog := ir.NewBlock(&ir.Process{Text: "a: float = 3.1"})
	d := Build(prog, Options{StripTypes: true})
	if countText(d, "a = 3.1") != 1 {
		t.Errorf("тип-анотацію не знято: %+v", d.Shapes)
	}
}

// TestNilBlocksNoPanic — ручний IR із nil-блоками (Else/Body) не має падати.
func TestNilBlocksNoPanic(t *testing.T) {
	cases := []*ir.Block{
		ir.NewBlock(&ir.If{Cond: "x > 0", Then: ir.NewBlock(&ir.Process{Text: "a"})}), // Else nil
		ir.NewBlock(&ir.If{Cond: "x > 0"}),                                            // обидві гілки nil
		ir.NewBlock(&ir.For{Spec: "i"}),                                               // Body і Else nil
		ir.NewBlock(&ir.While{Cond: "c"}),                                             // Body і Else nil
	}
	for i, prog := range cases {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("кейс %d: паніка на nil-блоці: %v", i, r)
				}
			}()
			Build(prog, Options{})
		}()
	}
}

// TestContinueNoShape — continue не має давати фігуру (це стрибок, а не блок).
func TestContinueNoShape(t *testing.T) {
	body := ir.NewBlock(
		&ir.If{Cond: "x < 0", Then: ir.NewBlock(&ir.Continue{}), Else: ir.NewBlock()},
		&ir.Process{Text: "p"},
	)
	d := Build(ir.NewBlock(&ir.For{Spec: "x", Body: body}), Options{})
	if countText(d, "continue") != 0 {
		t.Error("continue не має бути намальований фігурою")
	}
	if countKind(d, diagram.Decision) != 1 {
		t.Error("умова continue має бути ромбом")
	}
}

// TestLoopElse — гілка for/else має малюватись після виходу з циклу.
func TestLoopElse(t *testing.T) {
	prog := ir.NewBlock(&ir.For{
		Spec: "x", Body: ir.NewBlock(&ir.Process{Text: "a"}),
		Else: ir.NewBlock(&ir.Process{Text: "кінець-циклу"}),
	})
	d := Build(prog, Options{})
	if countText(d, "кінець-циклу") != 1 {
		t.Error("гілку else циклу не намальовано")
	}
}
