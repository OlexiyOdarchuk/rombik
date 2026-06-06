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
)

// Point — точка у координатах діаграми (y росте вниз).
type Point struct{ X, Y float64 }

// Shape — фігура. X,Y — лівий верхній кут; текст центрується.
type Shape struct {
	Kind Kind
	X, Y float64
	W, H float64
	Text string
}

// Edge — стрілка як ортогональна ламана (масив точок) з опційною позначкою.
// Arrowless — без вістря (для з'єднань у точці злиття гілок).
type Edge struct {
	Points    []Point
	Label     string
	Arrowless bool
}

// Diagram — повна розкладена схема.
type Diagram struct {
	Shapes []Shape
	Edges  []Edge
	W, H   float64
}
