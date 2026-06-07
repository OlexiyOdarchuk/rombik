package python

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/OlexiyOdarchuk/rombik/pkg/ir"
)

func TestParseAll(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 не встановлено")
	}
	src := "def f(x):\n" +
		"    print(x)\n" +
		"    if x > 0:\n" +
		"        return x\n" +
		"    return -x\n"
	funcs, err := ParseAll(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(funcs) != 1 || funcs[0].Name != "f" {
		t.Fatalf("очікувалась функція «f», маємо %+v", funcs)
	}
	// параметр x -> вхідний блок «Ввід x» першим
	first, ok := funcs[0].Body.Stmts[0].(*ir.IO)
	if !ok || first.Text != "Ввід x" {
		t.Errorf("перший блок мав бути «Ввід x», маємо %+v", funcs[0].Body.Stmts[0])
	}
}

func TestParseSyntaxError(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 не встановлено")
	}
	if _, err := ParseAll("def f(:\n  pass"); err == nil {
		t.Error("очікувалась помилка синтаксису")
	}
}

// walk рекурсивно обходить IR, викликаючи fn для кожного вузла.
func walk(n ir.Node, fn func(ir.Node)) {
	if n == nil {
		return
	}
	fn(n)
	switch x := n.(type) {
	case *ir.Block:
		for _, s := range x.Stmts {
			walk(s, fn)
		}
	case *ir.If:
		walk(x.Then, fn)
		walk(x.Else, fn)
	case *ir.For:
		walk(x.Body, fn)
		walk(x.Else, fn)
	case *ir.While:
		walk(x.Body, fn)
		walk(x.Else, fn)
	case *ir.DoWhile:
		walk(x.Body, fn)
	case *ir.InfLoop:
		walk(x.Body, fn)
	}
}

func findIfCond(body *ir.Block, substr string) bool {
	found := false
	walk(body, func(n ir.Node) {
		if i, ok := n.(*ir.If); ok && strings.Contains(i.Cond, substr) {
			found = true
		}
	})
	return found
}

// TestParseNewConstructs — match/try/continue/for-else більше НЕ губляться тихо.
func TestParseNewConstructs(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 не встановлено")
	}

	t.Run("continue", func(t *testing.T) {
		src := "def f(xs):\n    for x in xs:\n        if x < 0:\n            continue\n        print(x)\n"
		funcs, err := ParseAll(src)
		if err != nil {
			t.Fatal(err)
		}
		got := false
		walk(funcs[0].Body, func(n ir.Node) {
			if _, ok := n.(*ir.Continue); ok {
				got = true
			}
		})
		if !got {
			t.Error("continue не потрапив у IR")
		}
	})

	t.Run("for-else", func(t *testing.T) {
		src := "def f(xs):\n    for x in xs:\n        pass\n    else:\n        return True\n"
		funcs, err := ParseAll(src)
		if err != nil {
			t.Fatal(err)
		}
		got := false
		walk(funcs[0].Body, func(n ir.Node) {
			if fr, ok := n.(*ir.For); ok && fr.Else != nil && len(fr.Else.Stmts) > 0 {
				got = true
			}
		})
		if !got {
			t.Error("гілку for/else втрачено")
		}
	})

	t.Run("try-except", func(t *testing.T) {
		src := "def f(x):\n    try:\n        y = 1 / x\n    except ZeroDivisionError:\n        y = 0\n    return y\n"
		funcs, err := ParseAll(src)
		if err != nil {
			t.Fatal(err)
		}
		if !findIfCond(funcs[0].Body, "Виняток") {
			t.Error("обробку except втрачено (нема ромба «Виняток…»)")
		}
	})

	t.Run("match", func(t *testing.T) {
		src := "def f(c):\n    match c:\n        case 1:\n            return \"a\"\n        case _:\n            return \"b\"\n"
		funcs, err := ParseAll(src)
		if err != nil {
			t.Fatal(err)
		}
		if !findIfCond(funcs[0].Body, "== 1") {
			t.Error("match не знижено в if/elif (нема умови «c == 1»)")
		}
	})
}
