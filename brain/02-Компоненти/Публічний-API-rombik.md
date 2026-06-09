---
tags: [component, api, facade]
---

# Публічний API (pkg/rombik)

**Пакет:** `pkg/rombik` · **Файл:** `rombik.go`

Високорівневий **фасад**: об'єднує парсер, розкладку й усі рендери в кілька зручних
викликів. Це точка входу для **бібліотечного** використання (`go get` цього модуля в
інший проєкт). CLI і WASM теж спираються на ці ж пакети.

## Типи

```go
type Options = layout.Options       // ті самі перемикачі, що й розкладка ([[Опції-рендера]])

type Result struct {
    Name    string
    Diagram *diagram.Diagram
}
```

## Конструктори (код → схеми)

```go
func FromPython(code string, opts Options) ([]Result, error) // потребує python3 (не WASM)
func FromAST(astJSON []byte, opts Options) ([]Result, error) // AST дає Tree-sitter — працює в WASM
func SplitFromAST(astJSON []byte, opts Options, name string, maxH float64) ([]Result, error) // розбиття довгої схеми на частини через конектори
func FromIR(funcs []ir.Func, opts Options) []Result          // для тих, хто будує ir сам
```

Усі троє йдуть через внутрішній `build`, який на кожну функцію кличе
[[Layout-рушій-розкладки|layout.Build]] і **засіває підпис**: `Caption` = ім'я функції,
`FigNum` = порядковий номер (1..N), `CapWord` = опція. → [[Diagram-модель-геометрії]].

## Рендери на результаті

```go
func (r Result) SVG() string                       // → svg.Render
func (r Result) Typst() string                     // → typst.Render
func (r Result) PDF() ([]byte, error)              // → raster.PDF (нативно)
func (r Result) PNG(scale float64) ([]byte, error) // → raster.PNG (нативно)
func (r Result) Excalidraw() string                // → excalidraw.Render (нативно)
```

Один `Result` → будь-який із п'яти форматів. Жодного зовнішнього бінарника.

## Приклад

```go
res, _ := rombik.FromPython(code, rombik.Options{SingleEnd: true})
for _, f := range res {
    os.WriteFile(f.Name+".svg", []byte(f.SVG()), 0o644)
    png, _ := f.PNG(3)
    os.WriteFile(f.Name+".png", png, 0o644)
}
```

## Чому фасад

- **Зручність:** бібліотечному споживачу не треба знати про `astjson`, `layout`,
  окремі рендери — лише `From*` + методи `Result`.
- **Єдине засівання підпису:** логіка «ім'я функції → Caption, номер → FigNum» в одному
  місці, а не дублюється в CLI/WASM.
- **Межі лишаються чисті:** фасад нічого не додає до ядра — лише склеює. CLI і WASM
  можуть і не користуватися ним, а звертатися до пакетів напряму (так і роблять заради
  тонкого контролю). → [[Розділення-відповідальностей]].

## Пов'язане

- [[Layout-рушій-розкладки]] · [[Diagram-модель-геометрії]]
- [[SVG-рендерер]] · [[Typst-рендер]] · [[Растровий-рендер-PNG-PDF]]
- [[CLI-довідник]] · [[WASM-міст]]
