// Форматування чисел — дзеркало Go strconv FormatFloat 'f' (%.1f / %.0f):
// округлення до найближчого, рівні — «до парного». Tie детектуємо на ІСТИННОМУ
// значенні double (довгий toFixed), бо a*10^prec округлюється і дає фальшиві tie
// (напр. 146.65*10 → 1466.5). Спільне для svg/typst-рендерерів.
export function fmtFixed(x: number, prec: number): string {
  if (Object.is(x, -0)) x = 0;
  const neg = x < 0, a = neg ? -x : x;
  const ext = a.toFixed(prec + 18);
  const dot = ext.indexOf('.');
  const roundDigit = ext[dot + prec + 1];
  const rest = ext.slice(dot + prec + 2);
  if (roundDigit === '5' && /^0+$/.test(rest)) {  // рівний tie → до парного
    const keptLast = +((ext.slice(0, dot) + ext.slice(dot + 1, dot + 1 + prec)).slice(-1));
    const m = Math.pow(10, prec);
    const n = Math.floor(a * m);                  // a*m == n+0.5 точно для істинного tie
    const r = keptLast % 2 === 1 ? n + 1 : n;
    const s = (r / m).toFixed(prec);
    return neg && r !== 0 ? '-' + s : s;
  }
  const s = a.toFixed(prec);                      // не tie → toFixed = до найближчого
  return neg && a !== 0 ? '-' + s : s;
}

export const f1 = (n: number) => fmtFixed(n, 1);
export const f0 = (n: number) => fmtFixed(n, 0);
