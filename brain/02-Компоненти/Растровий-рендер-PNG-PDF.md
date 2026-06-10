---
tags: [component, render, raster, png, pdf]
---

# Растровий рендер (PNG/PDF)

**Модуль:** `web/src/lib/engine.js` (браузерний шар, не рушій)

PNG і PDF **більше не в рушії** і **не через `tdewolff/canvas`**. Після міграції
Go→TS растеризація живе у **браузері**: рушій `rombik-engine` (чистий TS, без DOM)
видає [[SVG-рендерер|SVG]], а перетворення SVG→растр робить сам браузер.

> [!important] Що змінилося
> Стара Go-растеризація (`pkg/render/raster` + `tdewolff/canvas` + вшитий `font.ttf`,
> а ще раніше зовнішні `rsvg-convert`/`typst`) — **історія**. Тепер рушій не залежить
> від графіки взагалі; PNG/PDF дає тонкий браузерний код поверх SVG. → [[Ідеї-та-межі]].

## Конвеєр

```
Diagram → renderSvg (TS-рушій) → <img src=blob:svg> → <canvas> → PNG
                                                     → jsPDF (растр посторінково) → PDF
```

- **`svgToCanvas(svg, scale)`** — SVG-рядок у `Blob` → `new Image()` → малюється у
  `<canvas>` з білим фоном і масштабом `scale` (`ctx.setTransform`). Браузер сам
  растеризує вектор.
- **`canvasToBytes(canvas)`** — `canvas.toBlob('image/png')` → `Uint8Array`.
- **`svgsToPdf(svgs, scale=3)`** — кожен SVG → окрема сторінка PDF: канва високої
  щільності вставляється через `jsPDF.addImage` (PNG-data-URL). Орієнтація сторінки
  (`l`/`p`) — за співвідношенням `w`/`h`. Розмір сторінки в пунктах = розмір схеми.
  `jspdf` підвантажується **ліниво** (`await import('jspdf')`) — лише коли треба PDF.

## API (експорти engine.js)

```js
export async function renderPng(diagram, cap, scale = 2, onProgress)  // PNG однієї схеми
export async function renderPngAll(diagrams, scale = 2, onProgress)   // один PNG з усіх схем
export async function renderPdf(diagram, cap, onProgress)             // PDF однієї схеми
export async function renderPdfAll(diagrams, onProgress)              // багатосторінковий PDF
```

`cap = { caption, figNum, capWord, capFormat }` доливається в `diagram` перед SVG-рендером,
щоб у растр потрапив підпис «Рисунок N — …» (`captionLine`, бо браузерний шлях не має
авто-нумерації, на відміну від [[Typst-рендер|Typst]]).

## Чому так

- **Рушій чистий.** `rombik-engine` не тягне ні графічних бібліотек, ні DOM — він
  лише будує `Diagram` і текстові формати (SVG/Typst/excalidraw). Усе важке —
  у браузері, де воно й так є.
- **Без зовнішніх бінарників і без WASM-растеризатора.** Раніше браузер ліниво
  підвантажував окремий ~16 МБ `rombik-raster.wasm`; тепер PNG/PDF дає сам браузер
  (canvas) + легкий `jspdf`.

## Пов'язане

- [[Diagram-модель-геометрії]]
- [[SVG-рендерер]] · [[Typst-рендер]]
- [[Браузерний-рушій]] · [[CLI-довідник]]
- [[Ідеї-та-межі]]
