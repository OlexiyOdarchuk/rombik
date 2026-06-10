---
tags: [component, render, typst]
---

# Typst-рендер

**Модуль:** `render/typst.ts` (`packages/engine/src/render/typst.ts`)

Малює [[Diagram-модель-геометрії|Diagram]] у **вихідний код Typst** (через пакет
CeTZ). Це не картинка, а текст `.typ`, який Typst компілює у тісний, чіткий PDF.
Залежності — лише `diagram.ts`/`format.ts`: модуль генерує рядок, без DOM.

## API

```ts
export function render(d: Diagram): string         // один самодостатній .typ-документ
export function renderAll(ds: Diagram[]): string   // багато схем — один документ (#pagebreak)
export function fragment(d: Diagram): string        // лише cetz.canvas (без преамбули)
export function fragmentAll(ds: Diagram[]): string  // canvas-блоки без преамбули
```

(в [[Публічний-API-rombik|index.ts]]: `renderTypst`/`renderTypstAll`/`renderTypstFragment`/…)

## Чому Typst, а не картинка

- **Векторно й чітко.** CeTZ малює ті самі ДСТУ-фігури примітивами Typst → у PDF вони
  ідеальної якості, з нормальним текстом (виділяється, шукається), а не растром.
- **Нативна нумерація.** `figure()` Typst сам нумерує «Рисунок N» і підставляє
  supplement/separator — тож rombik передає лише слово (`capSupplement`) і роздільник
  (`capSeparator`), а не готовий рядок. Пор. із SVG/PNG/PDF, де номер вшито в текст
  ([[Diagram-модель-геометрії]]).
- **Вбудовується в курсову.** Студент бере `.typ` і вставляє у свій Typst-документ —
  схема стає частиною роботи з наскрізною нумерацією рисунків. Для цього й `fragment`
  (лише `cetz.canvas`, без преамбули).

## Як рендерить

- **Преамбула:** `#import "@preview/cetz:0.3.4"`, авто-розмір сторінки
  (`width: auto, height: auto, margin: 14pt`), налаштування supplement/separator для
  `figure` через `#let flowchart(...)`.
- **Координати** ті самі, що в [[SVG-рендерер|SVG]], але вісь y CeTZ дивиться **вгору**,
  тож рендер перевертає y (`fy`). Має збігатися **байт-у-байт** із Go (golden-тести).
- Кожна фігура `Diagram` → CeTZ-примітив (прямокутник/ромб/паралелограм/шестикутник/
  стадіон/підпрограма); ребра → CeTZ-лінії зі стрілками; підписи «Так/Ні».
- Текст — **14pt**, шрифт Times New Roman (Liberation/DejaVu Serif як fallback);
  підписи ребер — 12pt. Координати форматуються `f1` (Go-сумісно).
- **`goQuote`** — дзеркало Go `strconv.Quote`/`%q`: екранує рядок для Typst, лишаючи
  друковані Unicode-символи (кирилиця, «», оператори) як є.

## Де використовується

- **CLI:** `-o файл.typ` → `render`/`renderAll`. → [[CLI-довідник]].
- **Браузер:** експорт усіх схем одним документом (`renderTypstAll`) або фрагменти.
- **PDF без Typst-бінарника:** для тих, хто не має Typst, є **прямий** PDF у браузері
  (SVG→canvas+jsPDF) — без зовнішнього інструмента. → [[Растровий-рендер-PNG-PDF]].

## Пов'язане

- [[Diagram-модель-геометрії]]
- [[SVG-рендерер]] · [[Растровий-рендер-PNG-PDF]]
- [[CLI-довідник]]
