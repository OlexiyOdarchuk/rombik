package python

import (
	"os/exec"
	"testing"

	"github.com/OlexiyOdarchuk/flowgen/pkg/ir"
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
