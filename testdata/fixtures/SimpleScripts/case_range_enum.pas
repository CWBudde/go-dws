type TColor = (Red, Green, Blue, Yellow);
type TDisj = (dOne = 1, dTen = 10, dTwo = 2);

procedure PrintColorRange(const value : TColor);
begin
   case value of
      Red..Blue : PrintLn('cool');
      Yellow : PrintLn('warm');
   else
      PrintLn('elsewhere');
   end;
end;

var c : TColor;
for c := Low(TColor) to High(TColor) do
   PrintColorRange(c);

// Ranges compare ordinal values, exactly like the <= and >= operators, so
// dTen (ordinal 10) falls outside dOne..dTwo (ordinals 1..2) even though it
// is declared between them.
procedure PrintDisjRange(const value : TDisj);
begin
   Print(IntToStr(Ord(value)) + ' ');
   case value of
      dOne..dTwo : Print('in');
   else
      Print('out');
   end;
   // A case range must always agree with the comparison operators.
   if (value >= dOne) and (value <= dTwo) then
      PrintLn(' cmp-in')
   else
      PrintLn(' cmp-out');
end;

PrintDisjRange(dOne);
PrintDisjRange(dTen);
PrintDisjRange(dTwo);

procedure PrintWideDisjRange(const value : TDisj);
begin
   Print(IntToStr(Ord(value)) + ' ');
   case value of
      dOne..dTen : Print('in');
   else
      Print('out');
   end;
   if (value >= dOne) and (value <= dTen) then
      PrintLn(' cmp-in')
   else
      PrintLn(' cmp-out');
end;

PrintWideDisjRange(dOne);
PrintWideDisjRange(dTen);
PrintWideDisjRange(dTwo);

// Aliases share an ordinal, so an alias of the lower bound is inside the range
// just as "=" reports it equal.
type TAlias = (aOne = 1, aAlias = 1, aTwo = 2);

PrintLn('aOne = aAlias: ' + BoolToStr(aOne = aAlias));
case aOne of
   aAlias..aTwo : PrintLn('alias in range');
else
   PrintLn('alias out of range');
end;

// Reversed bounds never match.
case Green of
   Blue..Red : PrintLn('oopsie!');
else
   PrintLn('reversed does not match');
end;

case Green of
   Red..Red : PrintLn('oopsie!');
   Blue..Yellow : PrintLn('oopsie!');
else
   PrintLn('no match');
end;
