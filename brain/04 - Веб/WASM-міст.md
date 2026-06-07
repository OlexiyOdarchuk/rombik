---
tags: [web, wasm, component]
---

# WASM-міст

**Пакет:** `cmd/wasm` · **Файл:** `main.go` · тег збірки `//go:build js && wasm`

Точка входу Go-двигуна у браузері. Реєструє в JS глобальну функцію `rombikGenerate` і
лишається «живим». **Не імпортує** `parser/python` — Python розбирає Pyodide, сюди
приходить готовий AST-JSON. → [[Розділення відповідальностей]].

## Що реєструє

```go
func main() {
    js.Global().Set("rombikGenerate", js.FuncOf(generate))
    select {}   // тримаємо модуль живим (інакше Go-main завершиться і функція зникне)
}
```

`select{}` блокує `main` назавжди — без цього після виходу з `main` зареєстрована
функція стала б недоступна. У JS її не `await`-ять (`go.run` без await) саме тому —
[[Браузерний двигун]].

## Контракт generate

```
rombikGenerate(astJSON: string, optionsJSON?: string) -> string (JSON)
```

```go
func generate(_ js.Value, args []js.Value) any {
    // 1) опції (необов'язкові)
    if len(args) > 1 && !args[1].IsUndefined() {
        json.Unmarshal([]byte(args[1].String()), &opts)   // layout.Options
    }
    // 2) AST-JSON → ir
    funcs, err := astjson.FromJSON([]byte(args[0].String()))
    if err != nil { return result({"error": err.Error()}) }
    // 3) КОЖНУ функцію розкласти + відрендерити
    for _, f := range funcs {
        d := layout.Build(f.Body, opts)
        res = append(res, outFunc{Name: f.Name, SVG: svg.Render(d), Diagram: d})
    }
    return result({"functions": res})
}
```

Вихід:

```json
{ "functions": [ { "name": "...", "svg": "<svg>…</svg>", "diagram": {…} } ] }
// або
{ "error": "повідомлення" }
```

`svg` готовий до вставки в DOM; `diagram` — сира геометрія
([[Diagram - модель геометрії]]) для експорту JSON / альтернативного рендеру.

## Те саме ядро, що в CLI

`generate` робить рівно те, що й CLI після парсингу: `astjson.FromJSON` →
`layout.Build` → `svg.Render`. Порівняй із `cmd/rombik/main.go`. Різниця лише в
адаптерах входу/виходу — ядро байт-у-байт спільне. → [[Конвеєр обробки]].

## Опції з JS

`layout.Options` має JSON-теги (`json:"callAsProcess"` тощо), тож галочки інтерфейсу
серіалізуються в `optionsJSON` і десеріалізуються тут. Список — [[Опції рендера]].

## Пов'язане

- [[Браузерний двигун]]
- [[astjson - конвертер]]
- [[Layout - рушій розкладки]]
- [[Розділення відповідальностей]]
