package astjson

import (
	"testing"

	"flowgen/pkg/ir"
)

func TestFromJSON(t *testing.T) {
	data := []byte(`[{"name":"f","block":{"kind":"block","stmts":[
		{"kind":"io","text":"Ввід x"},
		{"kind":"if","cond":"x > 0","then":{"kind":"block","stmts":[{"kind":"process","text":"a"}]},"else":{"kind":"block","stmts":[]}},
		{"kind":"terminal","text":"Повернути x"}
	]}}]`)
	funcs, err := FromJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(funcs) != 1 || funcs[0].Name != "f" {
		t.Fatalf("очікувалась 1 функція «f», маємо %+v", funcs)
	}
	stmts := funcs[0].Body.Stmts
	if len(stmts) != 3 {
		t.Fatalf("очікувалось 3 інструкції, маємо %d", len(stmts))
	}
	if _, ok := stmts[0].(*ir.IO); !ok {
		t.Error("перша має бути IO")
	}
	if _, ok := stmts[1].(*ir.If); !ok {
		t.Error("друга має бути If")
	}
	if _, ok := stmts[2].(*ir.Terminal); !ok {
		t.Error("третя має бути Terminal")
	}
}

func TestNestedBlockInlined(t *testing.T) {
	// вкладений block має інлайнитись (як from with/try)
	data := []byte(`[{"name":"m","block":{"kind":"block","stmts":[
		{"kind":"block","stmts":[{"kind":"process","text":"a"},{"kind":"process","text":"b"}]}
	]}}]`)
	funcs, _ := FromJSON(data)
	if n := len(funcs[0].Body.Stmts); n != 2 {
		t.Errorf("вкладений block мав інлайнитись у 2 інструкції, маємо %d", n)
	}
}
