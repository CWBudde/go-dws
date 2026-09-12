package frontend

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestCompile_CollectsParserDiagnostics(t *testing.T) {
	result := Compile("var x := );", "parser_error.pas", semantic.HintsLevelPedantic)

	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("expected parser diagnostics")
	}
	if !result.HasFatalDiagnostics() {
		t.Fatal("expected fatal parser diagnostics")
	}
	if !result.SemanticAttempted {
		t.Fatal("expected semantic analysis to run on recovered parse output")
	}
	if result.HasSemanticBlockingDiagnostics() {
		t.Fatal("expected recovered parser diagnostics to remain semantic-recoverable")
	}
	foundParsing := false
	for _, diag := range result.Diagnostics {
		if diag.Phase == PhaseParsing {
			foundParsing = true
			break
		}
	}
	if !foundParsing {
		t.Fatal("expected at least one parsing diagnostic")
	}
}

func TestParserDiagnosticBlocksSemantic(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		block bool
	}{
		{name: "recoverable invalid expression", code: parser.ErrInvalidExpression, block: false},
		{name: "recoverable missing end", code: parser.ErrMissingEnd, block: false},
		{name: "recoverable expected identifier", code: parser.ErrExpectedIdent, block: false},
		{name: "unknown parser code blocks", code: "E_UNKNOWN_PARSER_STATE", block: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.NewParserError(lexer.Position{Line: 1, Column: 1}, 1, "test", tt.code)
			got := parserDiagnosticBlocksSemantic(err)
			if got != tt.block {
				t.Fatalf("parserDiagnosticBlocksSemantic(%q) = %v, want %v", tt.code, got, tt.block)
			}
		})
	}
}

func TestCompile_CollectsSemanticDiagnostics(t *testing.T) {
	source := `
var i: Integer;
begin
	i := 'oops';
end;
`

	result := Compile(source, "semantic_error.pas", semantic.HintsLevelPedantic)

	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if !result.SemanticAttempted {
		t.Fatal("expected semantic analysis to run")
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("expected semantic diagnostics")
	}

	foundSemantic := false
	for _, diag := range result.Diagnostics {
		if diag.Phase == PhaseSemantic {
			foundSemantic = true
			break
		}
	}
	if !foundSemantic {
		t.Fatal("expected at least one semantic diagnostic")
	}
}

func TestExtractPosition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantMsg  string
		wantLine int
		wantCol  int
	}{
		{
			name:     "dwscript format",
			input:    `Syntax Error: Unknown name "Bug" [line: 2, column: 15]`,
			wantLine: 2,
			wantCol:  15,
			wantMsg:  `Syntax Error: Unknown name "Bug"`,
		},
		{
			name:     "at suffix",
			input:    "cannot access private field 'Field' of class 'TTest' at 16:2",
			wantLine: 16,
			wantCol:  2,
			wantMsg:  "cannot access private field 'Field' of class 'TTest'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, col, msg := extractPosition(tt.input)
			if line != tt.wantLine || col != tt.wantCol || msg != tt.wantMsg {
				t.Fatalf("extractPosition(%q) = (%d, %d, %q), want (%d, %d, %q)",
					tt.input, line, col, msg, tt.wantLine, tt.wantCol, tt.wantMsg)
			}
		})
	}
}

func TestCompile_RendersStructuredSemanticSyntaxErrorsInDWScriptFormat(t *testing.T) {
	source := `
type TEnum = (eZero, eOne, eTwo);

var ab1 : array [False..True] of Integer;
var ab2 : array [True..False] of Integer;

var v : Variant;

ab1[1]:=1;
ab1['1']:=1;
ab1[True]:=1;
ab1[Integer(True)]:=1;
ab1[eZero]:=1;
ab1[Integer(eZero)]:=1;
ab1[v]:=1;
`

	result := Compile(source, "array_index_bool.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Lower bound is greater than upper bound [line: 5, column: 23]`,
		`Syntax Error: Array index expected "Boolean" but got "Integer" [line: 9, column: 5]`,
		`Syntax Error: Array index expected "Boolean" but got "String" [line: 10, column: 5]`,
		`Syntax Error: Array index expected "Boolean" but got "Integer" [line: 12, column: 5]`,
		`Syntax Error: Array index expected "Boolean" but got "TEnum" [line: 13, column: 5]`,
		`Syntax Error: Array index expected "Boolean" but got "Integer" [line: 14, column: 5]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_AnchorsStructuredArrayBoundErrorsAtDWScriptPositions(t *testing.T) {
	source := `type a = array ['aa'] of String;`

	result := Compile(source, "array_error5.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Bound isn't of an ordinal type [line: 1, column: 16]`,
		`Syntax Error: ".." expected [line: 1, column: 21]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredArrayIndexAndDimensionDiagnostics(t *testing.T) {
	source := `
var s: String;
var x: Integer;
var arr: array of Integer;

arr := new Integer[3.14];

x := s['hello'];
x := x[0];
`

	result := Compile(source, "array_bucket.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: array dimension 1 must be integer, got Float [line: 6, column: 20]`,
		`Syntax Error: Array index expected "Integer" but got "String" [line: 8, column: 8]`,
		`Syntax Error: Array expected [line: 9, column: 7]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredInOperatorArrayMembershipDiagnostics(t *testing.T) {
	source := `
var ints: array of Integer;
var ok: Boolean := 1 in ints;
var bad: Boolean := 's' in ints;
`

	result := Compile(source, "in_array_typecheck.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Incompatible types: "String" and "Integer" [line: 4, column: 25]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStaticArrayLiteralSizeMismatchesCompileSide(t *testing.T) {
	source := `const a : array [0..2] of Integer = (
	1,
	2,
	3,
	4
	);

const b : array [0..2] of Integer = (
	1,
	2
	);`

	result := Compile(source, "array_const_item_count.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Incompatible types: "array [0..2] of Integer" and "array [0..3] of Integer" [line: 1, column: 37]`,
		`Syntax Error: Incompatible types: "array [0..2] of Integer" and "array [0..1] of Integer" [line: 8, column: 37]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_AnchorsStaticArrayLiteralAssignmentMismatchAtLiteral(t *testing.T) {
	source := `var xmlWhiteSpace = [' '];
xmlWhiteSpace := [' ', #9];`

	result := Compile(source, "var_static_array.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Incompatible types: Cannot assign "array [0..1] of String" to "array [0..0] of String" [line: 2, column: 18]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_AnchorsIndexedStaticArrayLiteralMismatchAtAssignment(t *testing.T) {
	source := `var crafts : array of array [0..19] of Integer;
begin
	crafts[0] := [$711E1F88, $711E039F, $0DBF, $0DD6, $099F, $0F7A, $097A];
end;`

	result := Compile(source, "array_item_mismatch1.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Incompatible types: Cannot assign "array [0..6] of Integer" to "array [0..19] of Integer" [line: 3, column: 12]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_NormalizesOperandDiagnostics(t *testing.T) {
	source := `
var v : Variant;
v := '123' + 123;
v := 123 + '123';
`

	result := Compile(source, "operand_normalization.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Invalid Operands [line: 3, column: 12]`,
		`Syntax Error: Invalid Operands [line: 4, column: 10]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_SuppressesFollowOnInferenceErrorAfterFailedInitializer(t *testing.T) {
	source := `
procedure Bug;
begin
end;

var i := Bug*Bug;
`

	result := Compile(source, "failed_initializer_follow_on.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Incompatible operands [line: 6, column: 13]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_NormalizesArrayHelperArityAndAnchorDiagnostics(t *testing.T) {
	source := `var a : array of Integer;

Print(a.Low(1));

Print(a.setlength(1, 2));

a.Delete;`

	result := Compile(source, "array_helper_arity_bucket.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: No arguments expected [line: 3, column: 14]`,
		`Hint: "setlength" does not match case of declaration ("SetLength") [line: 5, column: 9]`,
		`Syntax Error: Too many arguments [line: 5, column: 23]`,
		`Syntax Error: Expression expected [line: 5, column: 24]`,
		`Syntax Error: More arguments expected [line: 7, column: 9]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_NormalizesArrayHelperCallbackDiagnostics(t *testing.T) {
	source := `var a : array of Integer;

a.ForEach(StrToInt);

a.ForEach(IntToStr);

type
	TRec = record
		Field : Boolean;
	end;

var ar : array of TRec;
ar.Sort(@CompareStr);
ar.Sort(CompareStr);`

	result := Compile(source, "array_helper_callback_bucket.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: There is no overloaded version of "StrToInt" that can be called with these arguments [line: 3, column: 11]`,
		`Syntax Error: Incompatible parameter types - "procedure (Integer)" expected (instead of "Integer") [line: 3, column: 11]`,
		`Syntax Error: Incompatible parameter types - "procedure (Integer)" expected (instead of "function IntToStr(Integer): String") [line: 5, column: 11]`,
		`Syntax Error: More arguments expected [line: 13, column: 10]`,
		`Syntax Error: Incompatible types: "function (TRec, TRec): Integer" and "function CompareStr(String, String): Integer" [line: 13, column: 9]`,
		`Syntax Error: Incompatible parameter types - "function (TRec, TRec): Integer" expected (instead of "nil") [line: 13, column: 9]`,
		`Syntax Error: More arguments expected [line: 14, column: 9]`,
		`Syntax Error: Incompatible parameter types - "function (TRec, TRec): Integer" expected (instead of "Integer") [line: 14, column: 9]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RecoversUnitPrefixFamilies(t *testing.T) {
	source := `unit Bug.Test;

uses Bug.'123';`

	result := Compile(source, "unit_prefix8.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Name expected [line: 3, column: 10]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RecoversIncompleteVarDeclarations(t *testing.T) {
	source := `var a : bug`

	result := Compile(source, "var_incomplete4.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Unknown name "bug" [line: 1, column: 9]`,
		`Syntax Error: ";" expected [line: 1, column: 9]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredIndexedPropertyIndexDiagnostics(t *testing.T) {
	source := `
type
	TBox = class
		function GetItem(i: Integer): String; begin Result := ''; end;
		property Items[i: Integer]: String read GetItem; default;
	end;
var
	box: TBox;
	s: String;
begin
	s := box['bad'];
	s := box.Items['bad'];
end;
`

	result := Compile(source, "indexed_property_bucket.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Array index expected "Integer" but got "String" [line: 11, column: 11]`,
		`Syntax Error: Array index expected "Integer" but got "String" [line: 12, column: 17]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredPropertyDeclarationAccessorDiagnostics(t *testing.T) {
	source := `
type
	TReadField = class
		FCount: Integer;
		class property Count: Integer read FCount;
	end;

	TReadMethod = class
		function GetCount: Integer; begin Result := 0; end;
		class property Count: Integer read GetCount;
	end;

	TWriteMethod = class
		class var FCount: Integer;
		procedure SetCount(value: Integer); begin end;
		class property Count: Integer read FCount write SetCount;
	end;
`

	result := Compile(source, "property_decl_accessors.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Class member expected [line: 5, column: 38]`,
		`Syntax Error: Class method or constructor expected [line: 10, column: 38]`,
		`Syntax Error: Class method or constructor expected [line: 16, column: 51]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredPropertyDeclarationSignatureDiagnostics(t *testing.T) {
	source := `
type
	TBad = class
		function GetCount: String; begin Result := ''; end;
		procedure SetCount(value: String); begin end;
		property Count: Integer read GetCount write SetCount;
	end;
`

	result := Compile(source, "property_decl_signature.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: property 'Count' getter method 'GetCount' returns String, expected Integer [line: 6, column: 3]`,
		`Syntax Error: property 'Count' setter method 'SetCount' value parameter has type String, expected Integer [line: 6, column: 3]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredPropertyUseSiteDiagnostics(t *testing.T) {
	source := `
type
	TAccess = class
		FValue: Integer;
		property ReadOnlyValue: Integer read FValue;
		procedure SetWriteOnlyValue(value: Integer); begin end;
		property WriteOnlyValue: Integer write SetWriteOnlyValue;
		property Value: Integer read FValue write FValue;
	end;

	TStaticWrite = class
		class procedure SetValue(value: Integer); begin end;
		property Value: Integer write SetValue;
	end;
var
	access: TAccess;
	x: Integer;
begin
	access.ReadOnlyValue := 1;
	x := access.WriteOnlyValue;
	access.Value := 'bad';
	x := TAccess.ReadOnlyValue;
	TAccess.Value := 1;
	TStaticWrite.Value := 'bad';
end;
`

	result := Compile(source, "property_use_site_bucket.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Cannot set a value for a read-only property [line: 19, column: 9]`,
		`Syntax Error: Cannot read a write only property [line: 20, column: 14]`,
		`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 21, column: 9]`,
		`Syntax Error: Object reference needed to read/write an object field [line: 22, column: 15]`,
		`Syntax Error: Object reference needed to read/write an object field [line: 23, column: 10]`,
		`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 24, column: 15]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersForwardedAccessorMethodsAfterPropertyUseErrors(t *testing.T) {
	source := `
type
	TMyClass = class
		class procedure SetVal(v: Integer);
		property Val: Integer write SetVal;

		procedure SetProp(v: Integer);
		property Prop: Integer write SetProp;
	end;

TMyClass.Val := 'bad';
TMyClass.Prop := 123;
`

	result := Compile(source, "property_error8.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Argument 0 expects type "Integer" instead of "String" [line: 11, column: 10]`,
		`Syntax Error: Write access of property should be a static method [line: 12, column: 10]`,
		`Syntax Error: Class method or constructor expected [line: 12, column: 10]`,
		`Syntax Error: Method "SetVal" of class "TMyClass" not implemented [line: 4, column: 19]`,
		`Syntax Error: Method "SetProp" of class "TMyClass" not implemented [line: 7, column: 13]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersNoDefaultClassPropertyDiagnostic(t *testing.T) {
	source := `type
	TTest = class
		Field: Integer;
	end;
var i := TTest.Create[1];`

	result := Compile(source, "property_no_default.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := `Syntax Error: Class "TTest" has no default property [line: 5, column: 22]`
	if len(got) != 1 || got[0] != want {
		t.Fatalf("diagnostics = %v, want [%s]", got, want)
	}
}

func TestCompile_PrefersStructuredSemanticDiagnosticsOverLegacyMirrors(t *testing.T) {
	source := `
type
	TTest = class
		FValue: Integer;
		property Value: Integer read FValue write FValue;
	end;
begin
	TTest.Value := 1;
end;
`

	result := Compile(source, "structured_preferred.pas", semantic.HintsLevelPedantic)
	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if result.Analyzer == nil {
		t.Fatal("expected semantic analyzer to run")
	}
	if len(result.Analyzer.StructuredErrors()) == 0 {
		t.Fatal("expected structured semantic errors")
	}
	if len(result.Analyzer.Errors()) == 0 {
		t.Fatal("expected mirrored legacy semantic errors")
	}

	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Object reference needed to read/write an object field [line: 8, column: 8]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_DedupesStructuredAndLegacySemanticVisibilityErrors(t *testing.T) {
	source := `
type
   TTest = class
      private
         Field : Integer;
   end;
type
   TSubTest = class (TTest)
      protected
         Field2 : Integer;
         procedure Stuff;
   end;

procedure TSubTest.Stuff;
begin
   Field2:=1;
   Field:=2;
end;

var o := TSubTest.Create;
o.Field:=1;`

	result := Compile(source, "visibility4.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Member symbol "Field" is not visible from this scope [line: 17, column: 4]`,
		`Syntax Error: Member symbol "Field" is not visible from this scope [line: 21, column: 3]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredMethodAndClassVarVisibilityDiagnostics(t *testing.T) {
	source := `
type
   TTest = class
      private
         class var Hidden : Integer;
         function Helper : Integer;
   end;

function TTest.Helper : Integer;
begin
   Result := 1;
end;

var o := TTest.Create;
var x : Integer;
x := o.Helper();
x := TTest.Hidden;
`

	result := Compile(source, "visibility_methods.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Member symbol "Helper" is not visible from this scope [line: 16, column: 8]`,
		`Syntax Error: Member symbol "Hidden" is not visible from this scope [line: 17, column: 12]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredMemberAndHelperMissDiagnostics(t *testing.T) {
	source := `
type
   IThing = interface
   end;

   TStringHelper = helper for String
      function ToUpper: String;
   end;

   TPoint = class
      X : Integer;
   end;

var
   i : IThing;
   s : String;
   p : TPoint;
   a : Integer := i.Missing;
   b : String := s.Reverse;
   c : Integer := p.Y;
`

	result := Compile(source, "member_helper_bucket.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: There is no accessible member with name "Missing" for type IThing [line: 18, column: 21]`,
		`Syntax Error: There is no accessible member with name "Reverse" for type String [line: 19, column: 20]`,
		`Syntax Error: There is no accessible member with name "Y" for type TPoint [line: 20, column: 21]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersRecordKeywordMemberMissAfterRecoverableParserError(t *testing.T) {
	source := `
type
TPoint = record
	X: Integer;
end;

var
broken := ;
p: TPoint;
a: Integer := p.end;
`

	result := Compile(source, "record_keyword_member_recovery.pas", semantic.HintsLevelPedantic)
	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if !result.SemanticAttempted {
		t.Fatal("expected semantic analysis to run after recoverable parser error")
	}

	got := result.DiagnosticStrings()
	wantRecordMiss := `Syntax Error: There is no accessible member with name "end" for type TPoint [line: 10, column: 17]`
	foundParsing := false
	foundRecordMiss := false
	for _, diag := range result.Diagnostics {
		if diag.Phase == PhaseParsing {
			foundParsing = true
		}
		if diag.Render() == wantRecordMiss {
			foundRecordMiss = true
		}
	}

	if !foundParsing {
		t.Fatalf("expected parsing diagnostic, got: %v", got)
	}
	if !foundRecordMiss {
		t.Fatalf("expected record member miss diagnostic %q, got: %v", wantRecordMiss, got)
	}
}

func TestCompile_RendersStructuredUnknownNameDiagnostics(t *testing.T) {
	source := `
Foo();
`

	result := Compile(source, "unknown_name.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Unknown name "Foo" [line: 2, column: 5]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersStructuredNoOverloadDiagnostics(t *testing.T) {
	source := `
function Pick(x: Integer): Integer;
begin
	Result := x;
end;

function Pick(x: String): String;
begin
	Result := x;
end;

Pick(True);
`

	result := Compile(source, "no_overload.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: There is no overloaded version of "Pick" that can be called with these arguments [line: 12, column: 5]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_RendersCompileSideAbstractInstantiationDiagnostics(t *testing.T) {
	source := `
type
	TAbstract = class abstract
		function GetValue: Integer; abstract;
	end;

	TIncomplete = class(TAbstract)
	end;

var
	a: TAbstract;
	b: TIncomplete;
begin
	a := new TAbstract();
	a := TAbstract.Create();
	a := TAbstract.Create;
	b := TIncomplete.Create();
end;
`

	result := Compile(source, "abstract_compile_side.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Error: Trying to create an instance of an abstract class [line: 14, column: 7]`,
		`Error: Trying to create an instance of an abstract class [line: 15, column: 7]`,
		`Error: Trying to create an instance of an abstract class [line: 16, column: 17]`,
		`Error: Trying to create an instance of an abstract class [line: 17, column: 7]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSortDiagnostics_DeferredMethodNotImplementedComesLast(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:  `Method "SetVal" of class "TMyClass" not implemented`,
			Phase:    PhaseSemantic,
			Line:     3,
			Column:   23,
			Severity: SeverityError,
		},
		{
			Message:  `Syntax Error: Argument 0 expects type "Integer" instead of "String"`,
			Phase:    PhaseSemantic,
			Line:     10,
			Column:   10,
			Severity: SeverityError,
		},
	}

	sortDiagnostics(diags)

	if diags[0].Line != 10 || diags[1].Line != 3 {
		t.Fatalf("unexpected order after sort: %+v", diags)
	}
}

func TestSortDiagnostics_DeferredClassIncompleteComesLast(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:  `Class "TMyClass" isn't defined completely`,
			Phase:    PhaseSemantic,
			Line:     3,
			Column:   1,
			Severity: SeverityError,
		},
		{
			Message:  `Syntax Error: Unknown name "Foo"`,
			Phase:    PhaseSemantic,
			Line:     10,
			Column:   10,
			Severity: SeverityError,
		},
	}

	sortDiagnostics(diags)

	if diags[0].Line != 10 || diags[1].Line != 3 {
		t.Fatalf("unexpected order after sort: %+v", diags)
	}
}

func TestSortDiagnostics_ParserComesBeforeSemanticAtSameLocation(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:  `Syntax Error: Unknown name "Foo"`,
			Phase:    PhaseSemantic,
			Line:     4,
			Column:   7,
			Severity: SeverityError,
		},
		{
			Message:  `Expression expected`,
			Phase:    PhaseParsing,
			Line:     4,
			Column:   7,
			Severity: SeverityError,
		},
	}

	sortDiagnostics(diags)

	if diags[0].Phase != PhaseParsing || diags[1].Phase != PhaseSemantic {
		t.Fatalf("unexpected order after sort: %+v", diags)
	}
}

// A hint and an error on the same line keep the order they were emitted in, in
// both directions: upstream writes the two streams as it compiles, so neither
// severity nor column decides. `use_proc_result1` wants the hint at column 10
// before the error at column 8; `array_in1` wants the error at column 6 before
// the hint at column 17. The analyzer produces each in the right order, and the
// sort must not disturb either.
func TestSortDiagnostics_SameLineHintKeepsEmittedOrder(t *testing.T) {
	hint := Diagnostic{
		Message:  `Hint: "print" does not match case of declaration ("Print")`,
		Phase:    PhaseSemantic,
		Line:     2,
		Column:   10,
		Severity: SeverityHint,
	}
	err := Diagnostic{
		Message:  `Syntax Error: Incompatible operands`,
		Phase:    PhaseSemantic,
		Line:     2,
		Column:   8,
		Severity: SeverityError,
	}

	hintFirst := []Diagnostic{hint, err}
	sortDiagnostics(hintFirst)
	if hintFirst[0].Severity != SeverityHint || hintFirst[1].Severity != SeverityError {
		t.Fatalf("hint emitted first did not stay first: %+v", hintFirst)
	}

	errFirst := []Diagnostic{err, hint}
	sortDiagnostics(errFirst)
	if errFirst[0].Severity != SeverityError || errFirst[1].Severity != SeverityHint {
		t.Fatalf("error emitted first did not stay first: %+v", errFirst)
	}
}

func TestSortDiagnostics_ArgumentCountComesBeforeFollowOnTypeError(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:  `Syntax Error: Incompatible parameter types - "function (TRec, TRec): Integer" expected (instead of "nil")`,
			Phase:    PhaseSemantic,
			Line:     12,
			Column:   9,
			Severity: SeverityError,
		},
		{
			Message:  `Syntax Error: More arguments expected`,
			Phase:    PhaseSemantic,
			Line:     12,
			Column:   10,
			Severity: SeverityError,
		},
	}

	sortDiagnostics(diags)

	if diags[0].Message != `Syntax Error: More arguments expected` {
		t.Fatalf("unexpected order after sort: %+v", diags)
	}
}

func TestSortDiagnostics_StaticClassInstantiationBeforeNoInstances(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:  `Syntax Error: Class "TObj" is static, no instances allowed`,
			Phase:    PhaseSemantic,
			Line:     7,
			Column:   7,
			Severity: SeverityError,
		},
		{
			Message:  `Syntax Error: Class "TObj" is static, instantiation not allowed`,
			Phase:    PhaseSemantic,
			Line:     7,
			Column:   15,
			Severity: SeverityError,
		},
	}

	sortDiagnostics(diags)

	if diags[0].Message != `Syntax Error: Class "TObj" is static, instantiation not allowed` {
		t.Fatalf("unexpected order after sort: %+v", diags)
	}
}

func TestFilterDiagnostics_ReplacesGenericMemberMissWithVisibility(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:  `There is no accessible member with name "Field" for type TTest`,
			Rendered: `Syntax Error: There is no accessible member with name "Field" for type TTest [line: 5, column: 10]`,
			Phase:    PhaseSemantic,
			Line:     5,
			Column:   10,
			Severity: SeverityError,
			Fatal:    true,
		},
		{
			Message:  `Member symbol "Field" is not visible from this scope`,
			Rendered: `Syntax Error: Member symbol "Field" is not visible from this scope [line: 5, column: 10]`,
			Phase:    PhaseSemantic,
			Line:     5,
			Column:   10,
			Severity: SeverityError,
			Fatal:    true,
		},
	}

	sortDiagnostics(diags)
	got := filterDiagnostics(diags)

	if len(got) != 1 {
		t.Fatalf("expected 1 diagnostic after filtering, got %d: %+v", len(got), got)
	}
	if got[0].Message != `Member symbol "Field" is not visible from this scope` {
		t.Fatalf("unexpected filtered diagnostic: %+v", got[0])
	}
}

func TestFilterDiagnostics_ReplacesGenericMetaclassMemberErrorWithObjectReference(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:  `Syntax Error: Class method or constructor expected`,
			Rendered: `Syntax Error: Class method or constructor expected [line: 8, column: 8]`,
			Phase:    PhaseSemantic,
			Line:     8,
			Column:   8,
			Severity: SeverityError,
			Fatal:    true,
		},
		{
			Message:  `Syntax Error: Object reference needed to read/write an object field`,
			Rendered: `Syntax Error: Object reference needed to read/write an object field [line: 8, column: 8]`,
			Phase:    PhaseSemantic,
			Line:     8,
			Column:   8,
			Severity: SeverityError,
			Fatal:    true,
		},
	}

	sortDiagnostics(diags)
	got := filterDiagnostics(diags)

	if len(got) != 1 {
		t.Fatalf("expected 1 diagnostic after filtering, got %d: %+v", len(got), got)
	}
	if got[0].Message != `Syntax Error: Object reference needed to read/write an object field` {
		t.Fatalf("unexpected filtered diagnostic: %+v", got[0])
	}
}

func TestCompile_MixedStructuredAndLegacySemanticDiagnosticsRemainStable(t *testing.T) {
	source := `
type
	TTest = class
		FValue: Integer;
		property Value: Integer read FValue write FValue;
	end;
var
	o: TTest;
	i: Integer;
begin
	TTest.Value := 1;
	i := 'oops';
end;
`

	result := Compile(source, "mixed_semantic_sources.pas", semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Object reference needed to read/write an object field [line: 11, column: 8]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 12, column: 4]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_CollectsSemanticDiagnosticsAfterRecoverableParserError(t *testing.T) {
	source := `
var broken := ;
var i: Integer;
begin
	i := 'oops';
end;
`

	result := Compile(source, "recoverable_parser_plus_semantic.pas", semantic.HintsLevelPedantic)
	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if !result.SemanticAttempted {
		t.Fatal("expected semantic analysis to run after recoverable parser error")
	}

	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Expression expected [line: 2, column: 15]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 5, column: 4]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_CollectsSemanticDiagnosticsAfterMissingEnd(t *testing.T) {
	source := `
var i: Integer;
begin
	i := 'oops';
`

	result := Compile(source, "missing_end_plus_semantic.pas", semantic.HintsLevelPedantic)
	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if !result.SemanticAttempted {
		t.Fatal("expected semantic analysis to run after missing end parser error")
	}

	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: End of block expected [line: 3, column: 1]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 4, column: 4]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_OrdersSameLineRecoverableParserBeforeSemanticDiagnostics(t *testing.T) {
	source := `var broken := ; var i: Integer; i := 'oops';`

	result := Compile(source, "same_line_recoverable_parser_plus_semantic.pas", semantic.HintsLevelPedantic)
	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if !result.SemanticAttempted {
		t.Fatal("expected semantic analysis to run after recoverable parser error")
	}

	got := result.DiagnosticStrings()
	want := []string{
		`Syntax Error: Expression expected [line: 1, column: 15]`,
		`Syntax Error: Incompatible types: Cannot assign "String" to "Integer" [line: 1, column: 35]`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_SkipsSemanticDiagnosticsAfterBlockingParserError(t *testing.T) {
	result := compileParsedResult(&Result{
		Program: &ast.Program{},
		Diagnostics: parserDiagnostics([]*parser.ParserError{
			parser.NewParserError(lexer.Position{Line: 1, Column: 1}, 1, "test", "E_UNKNOWN_PARSER_STATE"),
		}),
	}, "if then", Options{Filename: "blocking_parser_only.pas", HintsLevel: semantic.HintsLevelPedantic})

	if result == nil {
		t.Fatal("expected non-nil compile result")
	}
	if result.SemanticAttempted {
		t.Fatal("expected semantic analysis to be skipped after blocking parser error")
	}
	for _, diag := range result.Diagnostics {
		if diag.Phase == PhaseSemantic {
			t.Fatalf("did not expect semantic diagnostics, got: %+v", result.Diagnostics)
		}
	}
}

// diagnostics is the shared assertion for the tests below: DWScript's output is
// an ordered list, so both the text and the order are part of the expectation.
func assertDiagnostics(t *testing.T, source, filename string, want []string) {
	t.Helper()
	result := Compile(source, filename, semantic.HintsLevelPedantic)
	got := result.DiagnosticStrings()
	if len(got) != len(want) {
		t.Fatalf("expected %d diagnostics, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diagnostic %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompile_WarnsOnAssignmentToForLoopVariable(t *testing.T) {
	// The anchor differs by the shape of the write: an ordinary assignment is
	// reported at its target, while a loop header that reuses an enclosing
	// loop's variable is reported at the `:=` — the token upstream's scanner
	// holds once it has read the name.
	assertDiagnostics(t, `var i : Integer;
for i:=1 to 10 do i:=1;`, "for_var_usage.pas", []string{
		`Warning: Assignment to FOR-Loop variable [line: 2, column: 19]`,
	})

	assertDiagnostics(t, `var i : Integer;
for i:=1 to 10 do 
   for i:=1 to 10 do ;`, "for_var_usage2.pas", []string{
		`Warning: Assignment to FOR-Loop variable [line: 3, column: 9]`,
		`Hint: Empty FOR loop [line: 3, column: 22]`,
	})
}

func TestCompile_RendersForInLoopVariableMismatch(t *testing.T) {
	// The loop header converts nothing, so the Integer-to-Float promotion an
	// assignment would perform is an error here. Both diagnostics are anchored
	// at the `in`, and the empty-body hint follows the error on the same line.
	assertDiagnostics(t, `var i : Integer;
var f : Float;

var ia : array of Integer;
var fa : array of Float;

for i in ia do ;
for f in ia do ;

for i in fa do ;
for f in fa do ;`, "for_in1.pas", []string{
		`Hint: Empty FOR loop [line: 7, column: 16]`,
		`Syntax Error: Incompatible types: "Float" and "Integer" [line: 8, column: 7]`,
		`Hint: Empty FOR loop [line: 8, column: 16]`,
		`Syntax Error: Incompatible types: "Integer" and "Float" [line: 10, column: 7]`,
		`Hint: Empty FOR loop [line: 10, column: 16]`,
		`Hint: Empty FOR loop [line: 11, column: 16]`,
	})
}

func TestCompile_SplitsNonEnumerableForInCollections(t *testing.T) {
	// Naming a type that is not an enumeration still builds a loop, so the
	// empty-body hint follows; an expression that is not a container does not,
	// and nothing further is reported about the loop.
	assertDiagnostics(t, `var i : Integer;

for i in Integer do ;`, "for_in2.pas", []string{
		`Syntax Error: Enumeration expected [line: 3, column: 7]`,
		`Hint: Empty FOR loop [line: 3, column: 21]`,
	})

	assertDiagnostics(t, `var i : Integer;

for i in PrintLn('bug') do ;`, "for_error3.pas", []string{
		`Syntax Error: Array expected [line: 3, column: 7]`,
	})
}

func TestCompile_IteratingAStringDrawsNoEmptyForLoopHint(t *testing.T) {
	// Upstream compiles a string for-in into a character walk rather than a
	// container loop, and that path emits no empty-body hint.
	assertDiagnostics(t, `var s := '';
for var c in s do;`, "for_var_in_string.pas", nil)
}

func TestCompile_KeepsHintsInEmissionOrderAgainstErrors(t *testing.T) {
	// A hint is never reordered against an error, on a shared line as much as
	// across lines: the name-resolution hint at column 10 precedes the enclosing
	// expression's error at column 8 because it was produced first.
	assertDiagnostics(t, `var x := 0;
x := x + print('');`, "use_proc_result1.pas", []string{
		`Hint: "print" does not match case of declaration ("Print") [line: 2, column: 10]`,
		`Syntax Error: Incompatible operands [line: 2, column: 8]`,
	})
}

// TestCompile_CallConventionHintAnchorsLikeEveryOtherDiagnostic pins that the
// hint a method's calling convention produces is the one a free routine produces:
// same sentence, and a position rendered as "[line: L, column: C]" rather than
// formatted into the message text as "at L:C". A hint that writes its own
// position can never match a fixture (SimpleScripts/call_conventions).
func TestCompile_CallConventionHintAnchorsLikeEveryOtherDiagnostic(t *testing.T) {
	const src = `type
   TMyClass = class
      procedure M; stdcall; begin end;
   end;
procedure F; register;
begin
end;`

	result := Compile(src, "call_conventions.pas", semantic.HintsLevelPedantic)
	if result == nil {
		t.Fatal("expected non-nil compile result")
	}

	var hints []Diagnostic
	for _, diag := range result.Diagnostics {
		if strings.Contains(diag.Message, "onvention") {
			hints = append(hints, diag)
		}
	}
	if len(hints) != 2 {
		t.Fatalf("got %d calling-convention hints %+v, want one for the method and one for the routine", len(hints), hints)
	}
	for _, hint := range hints {
		if !strings.Contains(hint.Message, "Call convention") ||
			!strings.Contains(hint.Message, "is not supported and ignored") {
			t.Fatalf("hint %q does not use the shared sentence", hint.Message)
		}
		// A hint that formatted its own "at L:C" would leave the structured
		// position empty, which is what made the method form unmatchable.
		if hint.Line == 0 || hint.Column == 0 {
			t.Fatalf("hint %q carries no structured position (%d:%d)", hint.Message, hint.Line, hint.Column)
		}
		if strings.Contains(hint.Message, " at ") {
			t.Fatalf("hint %q formats its own position", hint.Message)
		}
	}
}
