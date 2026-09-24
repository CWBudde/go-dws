package interp

import "testing"

func TestCompoundIndexAssignment_EvaluatesTargetOnce(t *testing.T) {
	for _, tt := range []struct {
		name, source, want string
	}{
		{
			name: "array index",
			source: `var a := [10, 20];
var calls := 0;
function NextIndex: Integer;
begin Inc(calls); Result := calls - 1; end;
a[NextIndex()] += 5;
PrintLn(calls);
PrintLn(a[0]);
PrintLn(a[1]);`,
			want: "1\n15\n20\n",
		},
		{
			name: "RHS replaces member receiver",
			source: `type THolder = class Values: array of Integer; end;
var first := THolder.Create;
var second := THolder.Create;
first.Values := [10];
second.Values := [100];
var target := first;
function RightSide: Integer;
begin target := second; Result := 5; end;
target.Values[0] += RightSide();
PrintLn(first.Values[0]);
PrintLn(second.Values[0]);`,
			want: "15\n100\n",
		},
		{
			name: "nested array indices",
			source: `var a : array of array of Integer;
a.SetLength(2);
a[0] := [10];
a[1] := [100];
var outerCalls, innerCalls: Integer;
function OuterIndex: Integer;
begin Inc(outerCalls); Result := outerCalls - 1; end;
function InnerIndex: Integer;
begin Inc(innerCalls); Result := 0; end;
a[OuterIndex()][InnerIndex()] += 5;
PrintLn(outerCalls);
PrintLn(innerCalls);
PrintLn(a[0][0]);
PrintLn(a[1][0]);`,
			want: "1\n1\n15\n100\n",
		},
		{
			name: "associative key",
			source: `var a: array [String] of Integer;
a['first'] := 10;
a['second'] := 100;
var calls := 0;
function NextKey: String;
begin
  Inc(calls);
  if calls = 1 then Result := 'first' else Result := 'second';
end;
a[NextKey()] += 5;
PrintLn(calls);
PrintLn(a['first']);
PrintLn(a['second']);`,
			want: "1\n15\n100\n",
		},
		{
			name: "record array field",
			source: `type TBox = record Values: array of Integer; end;
var box: TBox;
box.Values := [10, 100];
var calls := 0;
function NextIndex: Integer;
begin Inc(calls); Result := calls - 1; end;
box.Values[NextIndex()] += 5;
PrintLn(calls);
PrintLn(box.Values[0]);
PrintLn(box.Values[1]);`,
			want: "1\n15\n100\n",
		},
		{
			name: "string field",
			source: `type TBox = class Text: String; end;
var box := TBox.Create;
box.Text := 'ab';
var calls := 0;
function NextIndex: Integer;
begin Inc(calls); Result := calls; end;
box.Text[NextIndex()] += '';
PrintLn(calls);
PrintLn(box.Text);`,
			want: "1\nab\n",
		},
		{
			name: "named indexed property",
			source: `var reads, writes, receivers, indices, rightSides: Integer;
type THolder = class
  Value: Integer;
  function GetItem(i: Integer): Integer;
  begin Inc(reads); Result := Value + i; end;
  procedure SetItem(i, v: Integer);
  begin Inc(writes); Value := v - i; end;
  property Items[i: Integer]: Integer read GetItem write SetItem;
end;
var first := THolder.Create;
var second := THolder.Create;
first.Value := 10;
second.Value := 100;
function Receiver: THolder;
begin Inc(receivers); if receivers = 1 then Result := first else Result := second; end;
function NextIndex: Integer;
begin Inc(indices); Result := 1; end;
function RightSide: Integer;
begin Inc(rightSides); Result := 5; end;
Receiver().Items[NextIndex()] += RightSide();
PrintLn(first.Value);
PrintLn(second.Value);
PrintLn(receivers);
PrintLn(indices);
PrintLn(reads);
PrintLn(writes);
PrintLn(rightSides);`,
			want: "15\n100\n1\n1\n1\n1\n1\n",
		},
		{
			name: "default indexed property",
			source: `var reads, writes, indices: Integer;
type THolder = class
  Value: Integer;
  function GetItem(i: Integer): Integer;
  begin Inc(reads); Result := Value; end;
  procedure SetItem(i, v: Integer);
  begin Inc(writes); Value := v; end;
  property Items[i: Integer]: Integer read GetItem write SetItem; default;
end;
var h := THolder.Create;
h.Value := 10;
function NextIndex: Integer;
begin Inc(indices); Result := 0; end;
h[NextIndex()] += 5;
PrintLn(h.Value);
PrintLn(indices);
PrintLn(reads);
PrintLn(writes);`,
			want: "15\n1\n1\n1\n",
		},
		{
			name: "class indexed property",
			source: `var indices, reads, writes: Integer;
type THolder = class
  class var Value: Integer;
  class function GetItem(i: Integer): Integer;
  begin Inc(reads); Result := Value; end;
  class procedure SetItem(i, v: Integer);
  begin Inc(writes); Value := v; end;
  class property Items[i: Integer]: Integer read GetItem write SetItem;
end;
THolder.Value := 10;
function NextIndex: Integer;
begin Inc(indices); Result := 0; end;
THolder.Items[NextIndex()] += 5;
PrintLn(THolder.Value);
PrintLn(indices);
PrintLn(reads);
PrintLn(writes);`,
			want: "15\n1\n1\n1\n",
		},
		{
			name: "index failure skips getter RHS and setter",
			source: `var indices, reads, rightSides, writes: Integer;
type THolder = class
  function GetItem(i: Integer): Integer;
  begin Inc(reads); Result := 10; end;
  procedure SetItem(i, v: Integer);
  begin Inc(writes); end;
  property Items[i: Integer]: Integer read GetItem write SetItem;
end;
var h := THolder.Create;
function NextIndex: Integer;
begin Inc(indices); raise Exception.Create('bad index'); end;
function RightSide: Integer;
begin Inc(rightSides); Result := 5; end;
try h.Items[NextIndex()] += RightSide();
except on E: Exception do PrintLn(E.Message); end;
PrintLn(indices);
PrintLn(reads);
PrintLn(rightSides);
PrintLn(writes);`,
			want: "bad index\n1\n0\n0\n0\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(t, runQuickwinScript(t, tt.source), tt.want)
		})
	}
}
