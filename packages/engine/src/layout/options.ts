// Опції рендера (порт layout.Options) + застосування слів вводу/виводу і зняття
// тип-анотацій. Вхідні опції — усі необов'язкові; resolveOptions заповнює ДСТУ-замовч.

export interface Options {
  callAsProcess?: boolean;
  singleEnd?: boolean;
  yes?: string;
  no?: string;
  inWord?: string;
  outWord?: string;
  startText?: string;
  endText?: string;
  stripTypes?: boolean;
  returnAsIO?: boolean;
  capWord?: string;
  noStart?: boolean;
  noEnd?: boolean;
  // ДСТУ: Початок/Кінець лише для main; інші функції → вхід/вихід (з цими словами).
  mainOnlyTerminators?: boolean;
  entryText?: string;
  exitText?: string;
  // формат лічильникового for: 'comma' (i = 0, 9, 1) | 'range' (i = 0..9) | 'verbose' (i від 0 до 9)
  forFormat?: 'comma' | 'range' | 'verbose';
}

export interface ResolvedOptions {
  callAsProcess: boolean;
  singleEnd: boolean;
  yes: string;
  no: string;
  inWord: string;
  outWord: string;
  startText: string;
  endText: string;
  stripTypes: boolean;
  returnAsIO: boolean;
  capWord: string;
  noStart: boolean;
  noEnd: boolean;
  mainOnlyTerminators: boolean;
  entryText: string;
  exitText: string;
}

export function resolveOptions(o: Options = {}): ResolvedOptions {
  return {
    callAsProcess: o.callAsProcess ?? false,
    singleEnd: o.singleEnd ?? false,
    yes: o.yes || 'Так',
    no: o.no || 'Ні',
    inWord: o.inWord || 'Ввід',
    outWord: o.outWord || 'Вивід',
    startText: o.startText || 'Початок',
    endText: o.endText || 'Кінець',
    stripTypes: o.stripTypes ?? false,
    returnAsIO: o.returnAsIO ?? false,
    capWord: o.capWord ?? '',
    noStart: o.noStart ?? false,
    noEnd: o.noEnd ?? false,
    mainOnlyTerminators: o.mainOnlyTerminators ?? false,
    entryText: o.entryText || 'Вхід',
    exitText: o.exitText || 'Вихід',
  };
}

// typeAnnRe — «name: type =» -> «name =». \w як у Go RE2 — лише ASCII.
const typeAnnRe = /^([\w.]+)\s*:\s*[^=]+=/;

// ioText застосовує обрані слова вводу/виводу до тексту IO.
export function ioText(t: string, o: ResolvedOptions): string {
  if (t.startsWith('Ввід')) return o.inWord + t.slice('Ввід'.length);
  if (t.startsWith('Вивід')) return o.outWord + t.slice('Вивід'.length);
  return t;
}

// procText прибирає тип-анотацію, якщо ввімкнено опцію.
export function procText(t: string, o: ResolvedOptions): string {
  return o.stripTypes ? t.replace(typeAnnRe, '$1 =') : t;
}
