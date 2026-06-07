package diagram

import (
	"encoding/json"
	"testing"
)

// TestKindRoundTrip — Kind має розбиратися назад із рядка (фронт шле diagram у
// wasm для PDF/PNG/ре-рендеру підпису). Без UnmarshalJSON тут була помилка.
func TestKindRoundTrip(t *testing.T) {
	for _, k := range []Kind{Terminator, Process, Decision, InOut, Hexagon, Predef} {
		b, err := json.Marshal(k)
		if err != nil {
			t.Fatal(err)
		}
		var got Kind
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("%v: %v", k, err)
		}
		if got != k {
			t.Errorf("round-trip %v → %s → %v", k, b, got)
		}
	}
}

// TestDiagramRoundTrip — повна діаграма marshal→unmarshal без втрат.
func TestDiagramRoundTrip(t *testing.T) {
	d := &Diagram{
		W: 100, H: 200, Caption: "f", FigNum: 3,
		Shapes: []Shape{{Kind: Decision, Text: "x>0"}, {Kind: Hexagon, Text: "i"}},
		Edges:  []Edge{{Points: []Point{{X: 1}, {Y: 2}}, Label: "Так"}},
	}
	b, _ := json.Marshal(d)
	var got Diagram
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Shapes[0].Kind != Decision || got.Shapes[1].Kind != Hexagon {
		t.Errorf("типи фігур втрачено: %+v", got.Shapes)
	}
}
