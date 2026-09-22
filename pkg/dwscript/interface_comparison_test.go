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
			name: "class and interface identity",
			source: `var Obj := TBoth.Create;
var Other := TBoth.Create;
var L: ILeft := Obj;
PrintLn(Obj = L); PrintLn(L = Obj);
PrintLn(Obj <> L); PrintLn(L <> Obj);
PrintLn(Other = L); PrintLn(L = Other);
PrintLn(Other <> L); PrintLn(L <> Other);`,
			want: "True\nTrue\nFalse\nFalse\nFalse\nFalse\nTrue\nTrue\n",
		},
		{
			name: "class and interface nil identity",
			source: `var Obj: TBoth;
var L: ILeft;
PrintLn(Obj = L); PrintLn(L = Obj);
PrintLn(Obj <> L); PrintLn(L <> Obj);
Obj := TBoth.Create;
PrintLn(Obj = L); PrintLn(L = Obj);
PrintLn(Obj <> L); PrintLn(L <> Obj);
L := Obj;
Obj := nil;
PrintLn(Obj = L); PrintLn(L = Obj);
PrintLn(Obj <> L); PrintLn(L <> Obj);`,
			want: "True\nTrue\nFalse\nFalse\nFalse\nFalse\nTrue\nTrue\nFalse\nFalse\nTrue\nTrue\n",
		},
		{
			name: "direct object assignment to interface aliases",
			source: `type IAliasChain = ILeftAlias;
type TChild = class(TBoth) end;
type TChildAlias = TChild;
var Obj: TChildAlias := TChild.Create;
var Initialized: IAliasChain := Obj;
var Assigned: ILeftAlias;
Assigned := Obj;
var Underlying: ILeft := Initialized;
var Copied: ILeftAlias := Underlying;
PrintLn(Initialized.LeftValue);
PrintLn(Assigned.LeftValue);
PrintLn(Copied.LeftValue);
PrintLn(Initialized); PrintLn(Assigned);
PrintLn(Obj = Initialized); PrintLn(Initialized = Obj);
PrintLn(Initialized = Assigned); PrintLn(Initialized <> Copied);
Assigned := nil;
PrintLn(Assigned = nil);
Assigned := Obj;
PrintLn(Assigned.LeftValue);`,
			want: "1\n1\n1\nTInterfaceSymbol\nTInterfaceSymbol\nTrue\nTrue\nTrue\nFalse\nTrue\n1\n",
		},
		{
			name: "nil alias initializer preserves interface assignment",
			source: `var L: ILeftAlias := nil;
PrintLn(L = nil);
L := TBoth.Create;
PrintLn(L.LeftValue);
PrintLn(L);
L := nil;
PrintLn(L = nil);`,
			want: "True\n1\nTInterfaceSymbol\nTrue\n",
		},
		{
			name: "nil object initializer preserves aliased interface type",
			source: `var Obj: TBoth;
var L: ILeftAlias := Obj;
PrintLn(L = Obj); PrintLn(Obj = L);
L := TBoth.Create;
PrintLn(L.LeftValue); PrintLn(L);`,
			want: "True\nTrue\n1\nTInterfaceSymbol\n",
		},
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
		{"incompatible object comparison", "PrintLn(TObject.Create = L);", "cannot compare"},
		{"mixed reference ordering", "PrintLn(TBoth.Create < L);", "Invalid Operands"},
		{"incompatible object alias initializer", "var A: ILeftAlias := TObject.Create;", "Cannot assign"},
		{"incompatible object alias assignment", "var A: ILeftAlias; A := TObject.Create;", "Cannot assign"},
		{"unrelated interface alias assignment", "var A: ILeftAlias; A := R;", "Incompatible types"},
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
