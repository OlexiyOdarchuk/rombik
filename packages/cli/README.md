# rombik (CLI)

Код (**Python / C / C++ / Pascal**) → блок-схема алгоритму за **ДСТУ 19.701‑90**, прямо в терміналі.
Самодостатній: тягне рушій і tree-sitter-граматики в один файл — потрібен лише **Node 22+**.

```bash
npx rombik examples/grade.py -o grade.svg     # без встановлення
# або глобально:
npm i -g rombik && rombik prog.cpp -t typ > prog.typ
```

## Використання

```
rombik <файл> [опції]        ( '-' = читати stdin )

  -o, --out FILE     вихід (без нього — stdout); кілька функцій → FILE_<імʼя>.<ext>
  -t, --format FMT   svg | typ | json | excalidraw            (типово svg)
  -l, --lang LANG    py | cpp | c | pas                        (типово за розширенням)
      --fn NAME      малювати лише функцію NAME
      --single-end   один спільний «Кінець» (інакше — на кожен return/raise)
      --split N      розбити схему на частини, не вищі за N
  -h, --help
```

## Приклади

```bash
rombik grade.py -o grade.svg
rombik prog.cpp -t typ > prog.typ
cat prog.py | rombik - -t json

# PNG / PDF — растеризуй SVG зовнішнім конвертером:
rombik grade.py | rsvg-convert -f pdf -o grade.pdf   # також cairosvg / inkscape
```

PNG/PDF самого CLI не дає (потрібен растеризатор) — на сайті ж є, а тут зручно
конвеєром у `rsvg-convert`.

Програмний доступ до рушія — пакет [`rombik-engine`](https://www.npmjs.com/package/rombik-engine).
Частина [rombik](https://github.com/OlexiyOdarchuk/rombik). MIT.
