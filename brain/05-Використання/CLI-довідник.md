---
tags: [usage, cli]
---

# CLI — довідник

**Команда:** `tools/rombik.mjs` — Node-скрипт на тому самому TS-рушії
(`@rombik/engine`). Жодного Go-бінарника й `python3`: парсинг робить web-tree-sitter
(wasm), розкладку та рендер — рушій. → [[Браузерний-рушій]].

Запуск (з кореня монорепо):

```bash
npm run cli -- examples/grade.py            # → npm run cli викликає скрипт
node --experimental-strip-types tools/rombik.mjs examples/grade.py
./tools/rombik.mjs examples/grade.py        # після chmod +x
```

Standalone-варіант (без `node_modules`/TS-флагів) — [[Збірка-і-запуск|build:cli]]:

```bash
npm run build:cli            # → dist/rombik.mjs (+ 3 wasm поруч)
node dist/rombik.mjs examples/grade.py -o grade.svg
```

Вхід — **файл** або `-` (читати stdin). Вихід — у `-o FILE` або в stdout (без `-o`).

## Опції

Розбираються через `node:util` `parseArgs`. Підтримуються довгі й короткі форми.

| Опція | Тип | Типово | Призначення |
|-------|-----|--------|-------------|
| `-o, --out FILE` | string | stdout | Файл виводу. Без нього — пишемо в stdout. Кілька функцій → `FILE_<імʼя>.<ext>`. `-o -` теж stdout. |
| `-t, --format FMT` | string | `svg` | `svg` \| `typ` (=`typst`) \| `json` \| `excalidraw`. |
| `-l, --lang LANG` | string | за розширенням | `py` \| `cpp`. Без нього мова визначається розширенням (`.cpp/.cc/.cxx/.hpp/.h/.cs` → C++, інакше Python; для stdin — Python). |
| `--fn NAME` | string | усі | Малювати лише функцію `NAME`. |
| `--single-end` | bool | `false` | Один спільний «Кінець» (`SingleEnd`) замість локального на кожен вихід. → [[Зведення-виходів-у-Кінець]]. |
| `--split N` | string/число | — | Розбити схему на частини, не вищі за `N` (одиниць висоти), з конекторами. Без `--fn` бере першу функцію файлу. |
| `-h, --help` | bool | `false` | Довідка. |

> [!note] Підмножина опцій рушія
> CLI відкриває лише `--single-end` зі структурних [[Опції-рендера|Options]]. Решта
> (слова I/O, тексти термінаторів, `stripTypes`, `returnAsIO`, `callAsProcess`,
> підпис/`capWord`) задаються програмно через рушій або у [[Фронтенд-SvelteKit|вебі]].

## Формати виводу

`-t` обирає рендерер рушія напряму (`tools/rombik.mjs`):

```
svg            → renderSvg(diagram)          — SVG-текст
typ | typst    → renderTypst(diagram)        — вихідний код Typst (CeTZ)
excalidraw     → renderExcalidraw(diagram)   — .excalidraw (JSON)
json           → JSON.stringify(diagram)     — сира геометрія
```

> [!warning] PNG/PDF у CLI поки нема
> Растеризація живе у браузері (SVG → `<canvas>` → PNG; PDF через jsPDF) і залежить
> від DOM/canvas, тож у Node-CLI її немає. Доступні `svg`/`typ`/`json`/`excalidraw`.
> PNG/PDF — лише у [[Фронтенд-SvelteKit|вебі]].

## Кілька функцій

- **Одна** функція (або відфільтрована `--fn`) у `-o` → точно в указаний файл.
- **Кілька** функцій із `-o FILE.ext` → окремі файли `FILE_<функція>.ext`.
- Без `-o` усі схеми йдуть у stdout підряд.

## Приклади

```bash
npm run cli -- examples/grade.py -o grade.svg
npm run cli -- examples/course.py --fn matrix_gen -o m.svg
npm run cli -- prog.cpp -t typ > prog.typ
cat prog.py | npm run cli -- - -t json
node dist/rombik.mjs examples/course.py -t excalidraw -o course.excalidraw
node dist/rombik.mjs examples/course.py --split 1200 --fn matrix_gen -o m.svg
```

## Пов'язане

- [[Опції-рендера]] · [[Підтримувані-конструкції-Python]]
- [[Збірка-і-запуск]] · [[Браузерний-рушій]]
- [[Публічний-API-rombik]] · [[Конвеєр-обробки]]
