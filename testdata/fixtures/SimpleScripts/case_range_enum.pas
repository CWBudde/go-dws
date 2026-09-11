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

// Ranges follow declaration order, so dTen sits between dOne and dTwo
// even though its ordinal value is 10.
procedure PrintDisjRange(const value : TDisj);
begin
   Print(IntToStr(Ord(value)) + ' ');
   case value of
      dOne..dTwo : PrintLn('in');
   else
      PrintLn('out');
   end;
end;

PrintDisjRange(dOne);
PrintDisjRange(dTen);
PrintDisjRange(dTwo);

procedure PrintNarrowDisjRange(const value : TDisj);
begin
   Print(IntToStr(Ord(value)) + ' ');
   case value of
      dOne..dTen : PrintLn('in');
   else
      PrintLn('out');
   end;
end;

PrintNarrowDisjRange(dOne);
PrintNarrowDisjRange(dTen);
PrintNarrowDisjRange(dTwo);

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
