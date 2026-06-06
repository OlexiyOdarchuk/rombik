package route

import "testing"

func TestRouteAvoidsObstacle(t *testing.T) {
	a, b := Pt{X: 100, Y: 0}, Pt{X: 100, Y: 300}
	obs := []Rect{{X: 60, Y: 120, W: 80, H: 40}} // блокує прямий шлях по x=100
	p := Route(a, b, obs)

	if len(p) < 3 {
		t.Fatalf("очікувався обхід (≥3 точки), маємо пряму: %v", p)
	}
	if p[0] != a || p[len(p)-1] != b {
		t.Fatalf("кінці шляху не збігаються: %v", p)
	}
	in := obs[0].inflate(pad)
	for i := 0; i+1 < len(p); i++ {
		if blocked(p[i], p[i+1], []Rect{in}) {
			t.Errorf("сегмент %v-%v ріже перешкоду", p[i], p[i+1])
		}
	}
}

func TestRouteStraightWhenClear(t *testing.T) {
	p := Route(Pt{X: 0, Y: 0}, Pt{X: 0, Y: 100}, nil)
	if len(p) != 2 {
		t.Errorf("чистий шлях має бути прямим (2 точки), маємо %v", p)
	}
}
