// Пакет diagram — геометричний результат розкладки: фігури з координатами і
// ребра-стрілки. Це «контракт» між layout (виробляє) і render (малює). Пуста
// структура даних, без логіки й залежностей.
package diagram

// Kind — тип фігури ДСТУ.
type Kind int

const (
	Terminator Kind = iota // початок/кінець — овал (стадіон)
	Process                // дія — прямокутник
	Decision               // умова — ромб
	InOut                  // ввід/вивід — паралелограм
	Hexagon                // початок циклу (for) — шестикутник
	Predef                 // виклик підпрограми — прямокутник з боковими рисками
)

// String — машинна назва типу (для JSON/фронтенду).
func (k Kind) String() string {
	switch k {
	case Terminator:
		return "terminator"
	case Process:
		return "process"
	case Decision:
		return "decision"
	case InOut:
		return "io"
	case Hexagon:
		return "loop"
	case Predef:
		return "subprogram"
	}
	return "unknown"
}

// MarshalJSON віддає тип фігури як читабельний рядок, а не число.
func (k Kind) MarshalJSON() ([]byte, error) { return []byte(`"` + k.String() + `"`), nil }

// Point — точка у координатах діаграми (y росте вниз).
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Shape — фігура. X,Y — лівий верхній кут; текст центрується.
type Shape struct {
	Kind Kind    `json:"kind"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	W    float64 `json:"w"`
	H    float64 `json:"h"`
	Text string  `json:"text"`
}

// Edge — стрілка як ортогональна ламана (масив точок) з опційною позначкою.
// Arrowless — без вістря (для з'єднань у точці злиття гілок).
type Edge struct {
	Points    []Point `json:"points"`
	Label     string  `json:"label,omitempty"`
	Arrowless bool    `json:"arrowless,omitempty"`
}

// Diagram — повна розкладена схема.
type Diagram struct {
	Shapes []Shape `json:"shapes"`
	Edges  []Edge  `json:"edges"`
	W      float64 `json:"w"`
	H      float64 `json:"h"`
}
