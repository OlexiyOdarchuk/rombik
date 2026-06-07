// Пакет python парсить Python-код у ir. Використовуємо РІДНИЙ Python `ast`
// (через python3 у підпроцесі): найточніший парсер Python, без cgo. Сам ast
// віддає спрощений JSON, а Go зводить його в ir.
//
// Вимога рантайму: python3 (3.9+, заради ast.unparse). Це адаптер парсера —
// ядро (layout/svg) від нього не залежить.
package python

import (
	_ "embed"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/OlexiyOdarchuk/rombik/pkg/ir"
	"github.com/OlexiyOdarchuk/rombik/pkg/parser/astjson"
)

// pyScript — той самий парсер, що й вантажить браузер у Pyodide (parser.py).
// CLI запускає його через python3 (читає stdin, друкує JSON).
//
//go:embed parser.py
var pyScript string

// Script повертає текст парсера — щоб фронтенд міг віддати його Pyodide.
func Script() string { return pyScript }

// AST повертає сирий AST-JSON від python3 (формат astjson). Саме це в браузері
// видаватиме Pyodide, виконуючи pyScript, — далі astjson.FromJSON будує ir.
func AST(code string) ([]byte, error) {
	cmd := exec.Command("python3", "-c", pyScript)
	cmd.Stdin = strings.NewReader(code)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			msg := strings.TrimSpace(string(ee.Stderr))
			if i := strings.LastIndex(msg, "\n"); i >= 0 {
				msg = strings.TrimSpace(msg[i+1:]) // лише останній (чистий) рядок, без трейсбека
			}
			if msg == "" {
				msg = "не вдалося розібрати код"
			}
			return nil, errors.New(msg)
		}
		return nil, fmt.Errorf("не вдалося запустити python3: %w (чи встановлено?)", err)
	}
	return out, nil
}

// ParseAll: код → AST через python3 → ir. У браузері той самий шлях, лише AST
// дає Pyodide замість підпроцесу.
func ParseAll(code string) ([]ir.Func, error) {
	out, err := AST(code)
	if err != nil {
		return nil, err
	}
	return astjson.FromJSON(out)
}
