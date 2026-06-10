program Grade;
var score, total: integer;
begin
  write('Введіть бал: ');
  readln(score);
  total := score + 5;
  if total >= 90 then
    writeln('Відмінно')
  else if total >= 60 then
    writeln('Задовільно')
  else
    writeln('Незадовільно');
  writeln('Готово');
end.
