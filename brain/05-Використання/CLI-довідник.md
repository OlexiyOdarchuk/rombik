---
tags: [usage, cli]
---

# CLI — довідник

**Команда:** `cmd/rombik` · `go run ./cmd/rombik [прапорці]`

Перетворює Python-файл на блок-схему. Формат виводу — за розширенням `-o`. PNG і PDF —
**нативні** (без зовнішніх бінарників). → [[Растровий-рендер-PNG-PDF]].

## Прапорці

| Прапорець | Тип | Типово | Призначення |
|-----------|-----|--------|-------------|
| `-py FILE` | string | `""` | Python-файл (без нього — демо). |
| `-o FILE` | string | `out.svg` | Вихід. Розширення → формат: `.svg`/`.png`/`.pdf`/`.typ`/`.json`/`.excalidraw`. |
| `-fn NAME` | string | `""` | Малювати лише функцію з цим іменем. |
| `-calls-plain` | bool | `false` | Виклики підпрограм → звичайний прямокутник (`CallAsProcess`). |
| `-single-end` | bool | `false` | Один спільний «Кінець» (`SingleEnd`). |
| `-scale N` | float | `2` | Щільність PNG (пікселів на одиницю). |
| `-caption S` | string | `""` | Підпис схеми (інакше — ім'я функції; `«-»` — без підпису). |
| `-fignum N` | int | `0` | Номер «Рисунок N» (0 — за порядком функцій). |
| `-figword S` | string | `""` | Слово підпису: «Рисунок» (замовч.), «Рис.» тощо. |
| `-capformat S` | string | `""` | Шаблон, напр. `«{num}. {text}»` (замовч. `«{word} {num} — {text}»`). |

`-calls-plain`/`-single-end` — підмножина [[Опції-рендера]]; решта структурних опцій
(слова I/O, тексти термінаторів, StripTypes, ReturnAsIO) доступні через WASM/фронтенд.
Підпис — [[Diagram-модель-геометрії|поля Caption/FigNum/CapWord/CapFormat]].

## Формати виводу (`write`)

```
.json → json.MarshalIndent(diagram)   — сира геометрія
.png  → raster.PNG(d, scale)          — НАТИВНИЙ растр (tdewolff/canvas, без rsvg)
.pdf  → raster.PDF / raster.PDFAll    — НАТИВНИЙ PDF (без typst-бінарника)
.typ  → typst.Render / RenderAll      — вихідний код Typst (CeTZ)
.excalidraw → excalidraw.Render / All — формат Excalidraw (JSON)
інше  → svg.Render                    — SVG-текст
```

> [!note] rsvg-convert більше НЕ потрібен
> Раніше `.png` йшов через `rsvg-convert`. Тепер PNG і PDF малюються нативно в Go. У
> коментарях файлу ще лишилися згадки rsvg — це історія, код їх не кличе.

## Кілька функцій

- **Одна** функція (або відфільтрована `-fn`) → точно у `-o`.
- **Кілька** функцій:
  - `.pdf`, `.typ` і `.excalidraw` → **один спільний документ** (`PDFAll`/`RenderAll`, наскрізна нумерація);
  - інші формати → файли `<основа>_<функція>.<ext>`.

## Приклади

```bash
go run ./cmd/rombik                                        # демо → out.svg
go run ./cmd/rombik -py examples/grade.py -o grade.svg
go run ./cmd/rombik -py examples/grade.py -o grade.png -scale 3
go run ./cmd/rombik -py examples/course.py -o course.pdf   # усі функції → один PDF
go run ./cmd/rombik -py examples/course.py -o course.typ   # Typst для вставки в курсову
go run ./cmd/rombik -py examples/course.py -o course.excalidraw # Excalidraw-формат
go run ./cmd/rombik -py examples/course.py -fn matrix_gen -o m.svg
go run ./cmd/rombik -py f.py -figword "Рис." -capformat "{word} {num}. {text}" -o s.pdf
```

## Залежності рантайму

- **`python3` 3.9+** — обов'язково для CLI (рідний `ast`).
- PNG/PDF/SVG/Typst — **нічого зовнішнього** (усе в Go-бінарнику). Готові бінарники на
  6 платформ — у GitHub Releases ([[Збірка-і-запуск]]).

## Пов'язане

- [[Опції-рендера]] · [[Підтримувані-конструкції-Python]]
- [[Растровий-рендер-PNG-PDF]] · [[Typst-рендер]]
- [[Публічний-API-rombik]] · [[Конвеєр-обробки]]
