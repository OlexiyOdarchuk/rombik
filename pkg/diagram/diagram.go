// Пакет diagram — геометричний результат розкладки: фігури з координатами і
// ребра-стрілки. Це «контракт» між layout (виробляє) і render (малює). Пуста
// структура даних, без логіки й залежностей.
package diagram

import (
	"errors"
	"strconv"
	"strings"
)

// Kind — тип фігури ДСТУ.
type Kind int

const (
	Terminator Kind = iota // початок/кінець — овал (стадіон)
	Process                // дія — прямокутник
	Decision               // умова — ромб
	InOut                  // ввід/вивід — паралелограм
	Hexagon                // початок циклу (for) — шестикутник
	Predef                 // виклик підпрограми — прямокутник з боковими рисками
	Connector              // з'єднувач «А-в-кружечку» — коло (розрив схеми на частини)
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
	case Connector:
		return "connector"
	}
	return "unknown"
}

// MarshalJSON віддає тип фігури як читабельний рядок, а не число.
func (k Kind) MarshalJSON() ([]byte, error) { return []byte(`"` + k.String() + `"`), nil }

// kindByName — зворотний бік String() (для UnmarshalJSON: фронт шле diagram назад).
var kindByName = map[string]Kind{
	"terminator": Terminator, "process": Process, "decision": Decision,
	"io": InOut, "loop": Hexagon, "subprogram": Predef, "connector": Connector,
}

// UnmarshalJSON приймає тип фігури як рядок («process») або число — щоб JSON,
// віддане фронтенду, розбиралося назад без втрат (PDF/PNG/ре-рендер підпису).
func (k *Kind) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if v, ok := kindByName[s]; ok {
		*k = v
		return nil
	}
	if n, err := strconv.Atoi(s); err == nil { // запасний шлях — сире число
		*k = Kind(n)
		return nil
	}
	return errors.New("невідомий тип фігури: " + s)
}

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
//
// Caption — підпис схеми (за замовч. ім'я функції; редагований). FigNum — номер
// «Рисунок N» для растрових форматів (SVG/PNG), де нема авто-нумерації; у Typst
// нумерує сам figure. CapWord — слово-supplement («Рисунок»/«Рис.»/своє); порожнє
// = ДСТУ-замовч. Порожній Caption — без підпису.
type Diagram struct {
	Shapes    []Shape `json:"shapes"`
	Edges     []Edge  `json:"edges"`
	W         float64 `json:"w"`
	H         float64 `json:"h"`
	Caption   string  `json:"caption,omitempty"`
	FigNum    int     `json:"figNum,omitempty"`
	CapWord   string  `json:"capWord,omitempty"`
	CapFormat string  `json:"capFormat,omitempty"` // шаблон: {word} {num} — {text}
}

// CaptionWord — слово-supplement підпису за замовчуванням (ДСТУ: «Рисунок»).
const CaptionWord = "Рисунок"

// CapFormatDefault — шаблон підпису за замовчуванням (ДСТУ: «Рисунок N — текст»).
const CapFormatDefault = "{word} {num} — {text}"

// CapSupplement — слово підпису (своє або ДСТУ-замовч.).
func (d *Diagram) CapSupplement() string {
	if d.CapWord != "" {
		return d.CapWord
	}
	return CaptionWord
}

// CapFmt — шаблон підпису (свій або замовч.).
func (d *Diagram) CapFmt() string {
	if d.CapFormat != "" {
		return d.CapFormat
	}
	return CapFormatDefault
}

// CaptionLine — повний рядок підпису для растрових форматів за шаблоном.
// Порожньо, якщо підпису нема; без номера ({num}<=0) — лише текст.
func (d *Diagram) CaptionLine() string {
	if d.Caption == "" {
		return ""
	}
	if d.FigNum <= 0 {
		return d.Caption
	}
	r := strings.NewReplacer(
		"{word}", d.CapSupplement(),
		"{num}", strconv.Itoa(d.FigNum),
		"{text}", d.Caption,
	)
	return strings.TrimSpace(r.Replace(d.CapFmt()))
}

// CapSeparator — роздільник між номером і текстом (частина шаблону між {num} і
// {text}) — для нативного підпису Typst-figure. Замовч. « — ».
func (d *Diagram) CapSeparator() string {
	f := d.CapFmt()
	i := strings.Index(f, "{num}")
	j := strings.Index(f, "{text}")
	if i < 0 || j < 0 || j < i {
		return " — "
	}
	return f[i+len("{num}") : j]
}

// CapHasWord — чи містить шаблон слово-supplement (для Typst-figure).
func (d *Diagram) CapHasWord() bool { return strings.Contains(d.CapFmt(), "{word}") }
