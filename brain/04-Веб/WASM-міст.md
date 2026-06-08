---
tags: [web, wasm, component]
---

# WASM-міст

**Пакети:** `cmd/wasm` (легкий) + `cmd/wasmraster` (важкий) · тег `//go:build js && wasm`

Go-рушій у браузері поділено на **два WASM-модулі**, бо растровий рендер (PNG/PDF)
тягне важкі залежності, які не потрібні для звичайного перегляду. Жоден із них **не
імпортує** `parser/python` (там `os/exec`) — Python розбирає Pyodide, сюди приходить
готовий AST-JSON. → [[Розділення-відповідальностей]].

## Модуль 1 — `cmd/wasm` (легкий, `rombik.wasm` ~4 МБ)

Парсинг-результат → SVG/Typst. Реєструє три глобальні функції:

```go
js.Global().Set("rombikGenerate", js.FuncOf(generate))   // AST-JSON → схеми (svg+typst+diagram)
js.Global().Set("rombikRenderOne", js.FuncOf(renderOne)) // дешевий ре-рендер однієї схеми
js.Global().Set("rombikTypstAll", js.FuncOf(typstAll))   // експорт усіх схем одним .typ
select {}                                                 // тримаємо модуль живим
```

| Функція | Вхід | Вихід (JSON) |
|---------|------|--------------|
| `rombikGenerate(astJSON, optionsJSON?)` | AST-JSON + опції | `{functions:[{name, svg, typst, diagram, ...}]}` |
| `rombikRenderOne(diagramJSON, captionJSON?)` | один `Diagram` + правки підпису | `{svg, typst}` |
| `rombikTypstAll(diagramsJSON)` | масив `Diagram` | `{typst}` |

`rombikRenderOne` — ключ до **живого редагування підпису**: фронт міняє лише
`Caption/FigNum/CapWord` і просить перемалювати ОДНУ схему, **без повторного парсингу**.
Саме заради цього `Diagram` має `UnmarshalJSON` ([[Diagram-модель-геометрії]]).

## Модуль 2 — `cmd/wasmraster` (важкий, `rombik-raster.wasm` ~16 МБ)

Нативний PNG/PDF через [[Растровий-рендер-PNG-PDF|raster]] (tdewolff/canvas). Фронт
вантажить його **ліниво** — лише на першому експорті PNG/PDF. Реєструє:

```go
js.Global().Set("rombikPng",    js.FuncOf(png))     // Diagram → PNG (base64)
js.Global().Set("rombikPdf",    js.FuncOf(pdf))     // Diagram → PDF
js.Global().Set("rombikPdfAll", js.FuncOf(pdfAll))  // масив Diagram → багатосторінковий PDF
```

## Чому два модулі

- **Швидкий старт:** для перегляду схем досить легкого 4 МБ, а не 16 МБ.
- **Lazy:** важкий растровий код вантажиться, лише коли реально треба PNG/PDF.
- **Та сама межа:** обидва беруть `Diagram`/AST-JSON і кличуть спільне ядро
  (`layout.Build` + рендери) — як і CLI. → [[Конвеєр-обробки]].

## select{} — чому

`main` навмисно висить на `select{}`: інакше після виходу з `main` зареєстровані
функції зникли б. У JS `go.run(instance)` не `await`-ять саме тому. → [[Браузерний-рушій]].

## Пов'язане

- [[Браузерний-рушій]] · [[Фронтенд-SvelteKit]]
- [[astjson-конвертер]] · [[Layout-рушій-розкладки]]
- [[Растровий-рендер-PNG-PDF]] · [[Typst-рендер]]
- [[Розділення-відповідальностей]]
