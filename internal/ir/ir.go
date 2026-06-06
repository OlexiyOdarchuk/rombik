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

// Return — повернення з функції: паралелограм «Повернути …» + термінатор Кінець.
// Завершує гілку (потік далі не йде).
type Return struct{ Text string }

func (*Return) node() {}

// Call — виклик підпрограми (предетермінований процес ДСТУ — прямокутник із
// боковими рисками). Лише для функцій, визначених у цьому ж файлі.
type Call struct{ Text string }

func (*Call) node() {}

// If — розгалуження (ромб ДСТУ) з гілками «Так»/«Ні».
type If struct {
	Cond string
	Then *Block
	Else *Block
}

func (*If) node() {}

// For — цикл з лічильником (шестикутник ДСТУ «початок циклу»). Spec — підпис
// у форматі «змінна = початок, кінець, крок» (напр. «i = 0, n-1, 1»).
type For struct {
	Spec string
	Body *Block
}

func (*For) node() {}

// While — цикл з передумовою: ромб згори, тіло, дуга повернення.
type While struct {
	Cond string
	Body *Block
}

func (*While) node() {}

// DoWhile — цикл з післяумовою (ідіома «while True: … if cond: break»):
// тіло згори, ромб-умова виходу знизу, дуга повернення. Cond — умова break.
type DoWhile struct {
	Cond string
	Body *Block
}

func (*DoWhile) node() {}

// Block — послідовність вузлів (виконуються згори вниз).
type Block struct{ Stmts []Node }

func (*Block) node() {}

// NewBlock — зручний конструктор послідовності.
func NewBlock(stmts ...Node) *Block { return &Block{Stmts: stmts} }

// Func — іменована програма (функція або тіло модуля). Один файл із кількома
// def дає кілька Func → кожна йде в окрему схему.
type Func struct {
	Name string
	Body *Block
}
