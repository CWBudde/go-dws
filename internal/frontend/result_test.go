package frontend

import (
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
		`Syntax Error: cannot index non-array type Integer [line: 9, column: 7]`,
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
		`Syntax Error: Read access of property should be a static method [line: 10, column: 38]`,
		`Syntax Error: Write access of property should be a static method [line: 16, column: 51]`,
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
		`Syntax Error: property 'ReadOnlyValue' is read-only [line: 19, column: 9]`,
		`Syntax Error: property 'WriteOnlyValue' is write-only [line: 20, column: 14]`,
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

	foundParsing := false
	foundSemantic := false
	for _, diag := range result.Diagnostics {
		if diag.Phase == PhaseParsing {
			foundParsing = true
		}
		if diag.Phase == PhaseSemantic {
			foundSemantic = true
		}
	}
	if !foundParsing || !foundSemantic {
		t.Fatalf("expected both parsing and semantic diagnostics, got: %+v", result.Diagnostics)
	}
}

func TestCompile_SkipsSemanticDiagnosticsAfterBlockingParserError(t *testing.T) {
	result := compileParsedResult(&Result{
		Program: &ast.Program{},
		Diagnostics: parserDiagnostics([]*parser.ParserError{
			parser.NewParserError(lexer.Position{Line: 1, Column: 1}, 1, "test", "E_UNKNOWN_PARSER_STATE"),
		}),
	}, "if then", "blocking_parser_only.pas", semantic.HintsLevelPedantic)

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
