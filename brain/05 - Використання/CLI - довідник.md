---
tags: [usage, cli]
---

# CLI — довідник

**Команда:** `cmd/rombik` · `go run ./cmd/rombik [прапорці]`

Перетворює Python-файл на блок-схему. Формат виводу — за розширенням `-o`.

## Прапорці

| Прапорець       | Тип    | Типово    | Призначення |
|-----------------|--------|-----------|-------------|
| `-py FILE`      | string | `""`      | Python-файл. Без нього — захардкоджене демо (`demo()`). |
| `-o FILE`       | string | `out.svg` | Вихід. Розширення → формат: `.svg` / `.png` / `.json`. |
| `-fn NAME`      | string | `""`      | Малювати лише функцію з цим іменем. |
| `-calls-plain`  | bool   | `false`   | Виклики підпрограм → звичайний прямокутник (опція `CallAsProcess`). |
| `-single-end`   | bool   | `false`   | Один спільний «Кінець» (опція `SingleEnd`). |
| `-scale N`      | float  | `2`       | Масштаб PNG (роздільність). |

Прапорці `-calls-plain` і `-single-end` — це підмножина [[Опції рендера]]; решта опцій
(слова вводу/виводу, тексти термінаторів, StripTypes, ReturnAsIO) доступні через WASM
з фронтенду, але **не** виведені окремими прапорцями CLI.

## Формати виводу (`write`)

```
.json → json.MarshalIndent(diagram)        — сира геометрія
.png  → svg.Render → rsvg-convert -z scale — растр (потрібен librsvg)
інше  → svg.Render                          — SVG-текст
```

PNG вимагає `rsvg-convert` у `$PATH`; без нього — зрозуміла помилка з підказкою
поставити librsvg або вивести `.svg`.

## Кілька функцій

- **Одна функція** у файлі (або відфільтрована `-fn`) → точно у `-o`.
- **Кілька функцій** → файли `<основа>_<функція>.<ext>`. Напр. `-o out.svg` дасть
  `out_grade.svg`, `out_main.svg`.

Кожна функція = окрема схема (так їх ділить [[Парсер Python|parser.py]]).

## Приклади

```bash
go run ./cmd/rombik                                   # демо → out.svg
go run ./cmd/rombik -py examples/grade.py -o grade.svg
go run ./cmd/rombik -py examples/grade.py -o grade.png -scale 3
go run ./cmd/rombik -py examples/course.py -fn matrix_gen -o matrix.svg
go run ./cmd/rombik -py examples/course.py -o схема.json   # геометрія в JSON
go run ./cmd/rombik -py examples/course.py -single-end -calls-plain -o s.svg
```

## Що друкує

Після запису — рядок-підсумок:

```
Готово: grade.svg (472×564, фігур: 7, ребер: 9)
```

## Залежності рантайму

- **`python3` 3.9+** — обов'язково (рідний `ast`, заради `ast.unparse`).
- **`rsvg-convert`** (librsvg) — лише для `.png`.

## Пов'язане

- [[Опції рендера]]
- [[Підтримувані конструкції Python]]
- [[Конвеєр обробки]]
