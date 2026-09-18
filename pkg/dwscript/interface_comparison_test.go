package dwscript

import (
	"bytes"
	"strings"
	"testing"
)

const interfaceComparisonDeclarations = `
type ILeft = interface
  function LeftValue: Integer;
end;
type IRight = interface
  function RightValue: Integer;
end;
type ILeftAlias = ILeft;
type IRightAlias = IRight;
type TBoth = class(TObject, ILeft, IRight)
  function LeftValue: Integer; begin Result := 1; end;
  function RightValue: Integer; begin Result := 2; end;
end;
`

func TestInterfaceComparison_Identity(t *testing.T) {
	for _, tc := range []struct {
		name, source, want string
	}{
		{
			name: "unrelated interfaces on same object",
			source: `var Obj := TBoth.Create;
var L: ILeft := Obj;
var R: IRight := Obj;
PrintLn(L = R); PrintLn(R = L);
PrintLn(L <> R); PrintLn(R <> L);`,
			want: "True\nTrue\nFalse\nFalse\n",
		},
		{
			name: "unrelated interfaces on different objects",
			source: `var L: ILeft := TBoth.Create;
var R: IRight := TBoth.Create;
PrintLn(L = R); PrintLn(R = L);
PrintLn(L <> R); PrintLn(R <> L);`,
			want: "False\nFalse\nTrue\nTrue\n",
		},
		{
			name: "separate wrappers and copied reference",
			source: `var Obj := TBoth.Create;
var L: ILeft := Obj;
var Other: ILeft := Obj;
var Copy := L;
PrintLn(L = Other); PrintLn(L <> Other);
PrintLn(L = Copy); PrintLn(L <> Copy);`,
			want: "True\nFalse\nTrue\nFalse\n",
		},
		{
			name: "interface aliases",
			source: `var Obj := TBoth.Create;
var LeftRef: ILeft := Obj;
var RightRef: IRight := Obj;
var L: ILeftAlias := LeftRef;
var R: IRightAlias := RightRef;
PrintLn(L = R); PrintLn(L <> R);`,
			want: "True\nFalse\n",
		},
		{
			name: "nil interfaces and nil literal",
			source: `var L: ILeft;
var R: IRight;
PrintLn(L = R); PrintLn(L <> R);
PrintLn(L = nil); PrintLn(nil = L);
PrintLn(L <> nil); PrintLn(nil <> L);
L := TBoth.Create;
PrintLn(L = R); PrintLn(R = L);
PrintLn(L <> R); PrintLn(R <> L);
PrintLn(L = nil); PrintLn(nil = L);
PrintLn(L <> nil); PrintLn(nil <> L);`,
			want: "True\nFalse\nTrue\nTrue\nFalse\nFalse\nFalse\nFalse\nTrue\nTrue\nFalse\nFalse\nTrue\nTrue\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			engine, err := New(WithTypeCheck(true), WithOutput(&output))
			if err != nil {
				t.Fatal(err)
			}
			program, err := engine.Compile(interfaceComparisonDeclarations + tc.source)
			if err != nil {
				t.Fatalf("Compile: %v", err)
			}
			if _, err := engine.Run(program); err != nil {
				t.Fatalf("Run: %v", err)
			}
			if got := output.String(); got != tc.want {
				t.Errorf("output = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestInterfaceComparison_InvalidOperands(t *testing.T) {
	for _, tc := range []struct {
		name, statement, diagnostic string
	}{
		{"less than", "PrintLn(L < R);", "Invalid Operands"},
		{"greater than", "PrintLn(L > R);", "Invalid Operands"},
		{"less than or equal", "PrintLn(L <= R);", "Invalid Operands"},
		{"greater than or equal", "PrintLn(L >= R);", "Invalid Operands"},
		{"integer equality", "PrintLn(L = 1);", "cannot compare"},
		{"string inequality", "PrintLn('value' <> L);", "cannot compare"},
		{"implicit unrelated assignment", "L := R;", "Incompatible types"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine, err := New(WithTypeCheck(true))
			if err != nil {
				t.Fatal(err)
			}
			_, err = engine.Compile(interfaceComparisonDeclarations + "var L: ILeft; var R: IRight;\n" + tc.statement)
			if err == nil || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("Compile error = %v, want diagnostic containing %q", err, tc.diagnostic)
			}
		})
	}
}
