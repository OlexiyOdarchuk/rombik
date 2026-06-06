// Пакет ir — логічне (мова-агностик) представлення алгоритму як структури
// керування. Сюди парсер зводить будь-який код (Python/C++/…), а layout
// перетворює це на геометрію. Жодного рендера тут — лише структура.
package ir

// Node — елемент алгоритму. Реалізації: Process, IO, If (далі — While тощо).
type Node interface{ node() }

// Process — дія/обчислення (прямокутник ДСТУ).
type Process struct{ Text string }

func (*Process) node() {}

// IO — ввід/вивід (паралелограм ДСТУ).
type IO struct{ Text string }

func (*IO) node() {}

// If — розгалуження (ромб ДСТУ) з гілками «Так»/«Ні».
type If struct {
	Cond string
	Then *Block
	Else *Block
}

func (*If) node() {}

// Block — послідовність вузлів (виконуються згори вниз).
type Block struct{ Stmts []Node }

func (*Block) node() {}

// NewBlock — зручний конструктор послідовності.
func NewBlock(stmts ...Node) *Block { return &Block{Stmts: stmts} }
