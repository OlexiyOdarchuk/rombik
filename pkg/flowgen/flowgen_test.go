package flowgen

import (
	"strings"
	"testing"
)

func TestFromASTtoSVG(t *testing.T) {
	ast := []byte(`[{"name":"f","block":{"kind":"block","stmts":[
		{"kind":"io","text":"Вивід «привіт»"}
	]}}]`)
	res, err := FromAST(ast, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].Name != "f" {
		t.Fatalf("очікувалась 1 схема «f», маємо %+v", res)
	}
	svg := res[0].SVG()
	if !strings.HasPrefix(svg, "<svg") || !strings.Contains(svg, "Вивід «привіт»") {
		t.Error("SVG некоректний або без тексту блоку")
	}
}

func TestFromASTbadJSON(t *testing.T) {
	if _, err := FromAST([]byte("не json"), Options{}); err == nil {
		t.Error("очікувалась помилка на некоректному JSON")
	}
}
