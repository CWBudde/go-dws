type
   TRec = record
      FData : array [0..3] of String;
      function GetIdx(i : Integer) : String;
      begin
         Result := FData[i];
      end;
      property Idx[i : Integer] : String read GetIdx; default;
   end;

var r : TRec;

PrintLn(r['oops']);
PrintLn(r[1]);
