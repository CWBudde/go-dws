package dwscript

import (
	"fmt"
	"strings"
	"testing"
)

// These public Compile controls catch accepting accessor-only index defaults,
// reporting defaults before type/mode errors, and checking the setter Value.
func TestEngineCompile_ClassPropertyIndexDefaults(t *testing.T) {
	const readDefault = `Syntax Error: Parameter 0 (Declared) - default value at implementation does not match declaration or forward [line: 5, column: 2]`
	const readSummary = `Syntax Error: Method "Get" has incompatible parameters [line: 5, column: 2]`
	const writeDefault = `Syntax Error: Parameter 0 (Declared) - default value at implementation does not match declaration or forward [line: 6, column: 2]`
	const writeSummary = `Syntax Error: Method "Put" has incompatible parameters [line: 6, column: 2]`
	cases := []struct {
		name, reader, writer, indices, result, tail string
		want                                        []string
	}{
		{"noDefaults", "Access: Integer", "Store: Integer; Value: Integer", "Declared: Integer", "Integer", "", nil},
		{"readerDefault", "Access: Integer = 1", "Store: Integer; Value: Integer", "Declared: Integer", "Integer", "", []string{readDefault, readSummary}},
		{"writerDefault", "Access: Integer", "Store: Integer = 0; Value: Integer = 2", "Declared: Integer", "Integer", "", []string{writeDefault, writeSummary}},
		{"readBeforeWrite", "Access: Integer = 1", "Store: Integer = 0; Value: Integer = 2", "Declared: Integer", "Integer", "", []string{readDefault, readSummary, writeDefault, writeSummary}},
		{"zeroDefault", "Access: Integer = 0", "Store: Integer; Value: Integer", "Declared: Integer", "Integer", "", []string{readDefault, readSummary}},
		{"typeBeforeModeAndDefault", "Access: String = 'x'", "const Store: Integer; Value: Integer", "const Declared: Integer", "Integer", "", []string{`Syntax Error: Parameter 0 - Type "Integer" expected (instead of "String") [line: 5, column: 2]`, readSummary}},
		{"modeBeforeDefault", "Access: Integer = 1", "const Store: Integer; Value: Integer", "const Declared: Integer", "Integer", "", []string{`Syntax Error: Parameter 0 (Declared) - Const-parameter expected [line: 5, column: 2]`, readSummary}},
		{"writerTypeBeforeDefault", "Access: Integer", "Store: String = 'x'; Value: Integer = 2", "Declared: Integer", "Integer", "", []string{`Syntax Error: Parameter 0 - Type "Integer" expected (instead of "String") [line: 6, column: 2]`, writeSummary}},
		{"writerModeBeforeDefault", "const Access: Integer", "Store: Integer = 1; Value: Integer = 2", "const Declared: Integer", "Integer", "", []string{`Syntax Error: Parameter 0 (Declared) - Const-parameter expected [line: 6, column: 2]`, writeSummary}},
		{"setterValueDefaultExempt", "Access: Integer", "Store: Integer; Value: Integer = 2", "Declared: Integer", "Integer", "", nil},
		{"setterValueModeExempt", "Access: Integer", "Store: Integer; var Value: Integer", "Declared: Integer", "Integer", "", nil},
		{"setterValueTypeBeforeDefault", "Access: Integer", "Store: Integer = 1; Value: String = 'x'", "Declared: Integer", "Integer", "", []string{writeSummary}},
		{"countBeforeDefaults", "Access: Integer = 1", "Store: Integer = 1; Value: Integer = 2", "Declared, Other: Integer", "Integer", "", []string{readSummary, writeSummary}},
		{"getterResultBeforeDefault", "Access: Integer = 1", "Store: Integer; Value: Integer", "Declared: Integer", "Float", "", []string{`Syntax Error: Field/method "Get" has an incompatible type [line: 5, column: 2]`}},
		{"laterReachedName", "Access: Integer = 1", "Store: Integer; Value: Integer", "Declared: Integer", "Integer", "Missing;", []string{readDefault, readSummary, `Syntax Error: Unknown name "Missing" [line: 8, column: 1]`}},
		{"aliasesCaseAndNames", "Access: INTALIAS = 1", "Store: intalias; Value: INTEGER", "Declared: integer", "Integer", "", []string{readDefault, readSummary}},
		{"aliasesWithoutDefaults", "Access: INTALIAS", "Store: intalias; Value: INTEGER", "Declared: integer", "Integer", "", nil},
		{"multipleIndices", "A: Integer = 1; B: Integer = 2", "C: Integer = 3; D: Integer = 4; Value: Integer = 5", "Declared, Other: Integer", "Integer", "", []string{
			readDefault,
			`Syntax Error: Parameter 1 (Other) - default value at implementation does not match declaration or forward [line: 5, column: 2]`,
			readSummary,
			writeDefault,
			`Syntax Error: Parameter 1 (Other) - default value at implementation does not match declaration or forward [line: 6, column: 2]`,
			writeSummary,
		}},
		{"mixedIndexPriority", "A: String = 'x'; B: Integer = 2; C: Integer = 3", "const D: Integer; const E: Integer; F: Integer; Value: Integer", "const Declared, Other: Integer; Last: Integer", "Integer", "", []string{
			`Syntax Error: Parameter 0 - Type "Integer" expected (instead of "String") [line: 5, column: 2]`,
			`Syntax Error: Parameter 1 (Other) - Const-parameter expected [line: 5, column: 2]`,
			`Syntax Error: Parameter 2 (Last) - default value at implementation does not match declaration or forward [line: 5, column: 2]`,
			readSummary,
		}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			source := "type IntAlias = Integer; type T = class\n function Get(" + tt.reader + "): " + tt.result + "; begin Result := 0; end;\n procedure Put(" + tt.writer + "); begin end;\n property P[" + tt.indices + "]: Integer read\n Get write\n Put;\nend;\n" + tt.tail
			assertClassPropertyIndexCompile(t, source, tt.want)
		})
	}
}

// Selection must use the inherited/forwarded declaration signature, including
// its defaults, without revalidating an unrelated descendant override.
func TestEngineCompile_ClassPropertyIndexDefaults_SelectedSignatures(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         []string
	}{
		{"inheritedGetterDefault", `type T = class
 function Get(Access: Integer = 1): Integer; begin Result := 0; end;
end;
type U = class(T)
 property P[Declared: Integer]: Integer read
 Get;
end;`, []string{
			`Syntax Error: Parameter 0 (Declared) - default value at implementation does not match declaration or forward [line: 6, column: 2]`,
			`Syntax Error: Method "Get" has incompatible parameters [line: 6, column: 2]`,
		}},
		{"inheritedWriterDefault", `type T = class
 procedure Put(Store: Integer = 1; Value: Integer = 2); begin end;
end;
type U = class(T)
 property P[Declared: Integer]: Integer write
 Put;
end;`, []string{
			`Syntax Error: Parameter 0 (Declared) - default value at implementation does not match declaration or forward [line: 6, column: 2]`,
			`Syntax Error: Method "Put" has incompatible parameters [line: 6, column: 2]`,
		}},
		{"forwardedLexicalGetter", `type T = class
 function Get(Access: Integer): Integer; virtual; begin Result := 0; end;
 property P[Declared: Integer]: Integer read Get;
end;
type U = class(T)
 function Get(Access: Integer = 1): Integer; override; begin Result := 0; end;
 property Q[Other: Integer]: Integer read P;
end;`, nil},
		{"forwardedLexicalWriter", `type T = class
 procedure Put(Store: Integer; Value: Integer); virtual; begin end;
 property P[Declared: Integer]: Integer write Put;
end;
type U = class(T)
 procedure Put(Store: Integer = 1; Value: Integer = 2); override; begin end;
 property Q[Other: Integer]: Integer write P;
end;`, nil},
		{"forwardDeclarationDefault", `type T = class
 function Get(Access: Integer = 1): Integer;
 property P[Declared: Integer]: Integer read
 Get;
end;
function T.Get(Access: Integer): Integer; begin Result := 0; end;`, []string{
			`Syntax Error: Parameter 0 (Declared) - default value at implementation does not match declaration or forward [line: 4, column: 2]`,
			`Syntax Error: Method "Get" has incompatible parameters [line: 4, column: 2]`,
		}},
		{"indexDirectiveDefaultExempt", `type T = class
 function Get(Access: Integer = 1): Integer; begin Result := 0; end;
 procedure Put(Store: Integer = 1; Value: Integer = 2); begin end;
 property P: Integer index 7 read Get write Put;
end;`, nil},
		{"indexDirectiveTypePriority", `type T = class
 function Get(Access: String = 'x'): Integer; begin Result := 0; end;
 property P: Integer index 7 read
 Get;
end;`, []string{
			`Syntax Error: Method "Get" has incompatible parameters [line: 4, column: 2]`,
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertClassPropertyIndexCompile(t, tt.source, tt.want)
		})
	}
}

// The same checker serves interfaces; priority still holds where declared
// accessor defaults are currently absent from interface signature metadata.
func TestEngineCompile_PropertyIndexDefaults_InterfacePriority(t *testing.T) {
	// Interface summaries retain their legacy Syntax Error prefix in the
	// public Message; the formatter below adds the outer severity caption.
	for _, tt := range []struct {
		name, actual string
		want         []string
	}{
		{"typeBeforeMode", "String = 'x'", []string{
			`Syntax Error: Parameter 0 - Type "Integer" expected (instead of "String") [line: 4, column: 2]`,
			`Syntax Error: Syntax Error: Method "Get" has incompatible parameters [line: 4, column: 2]`,
		}},
		{"modeBeforeDefault", "Integer = 1", []string{
			`Syntax Error: Parameter 0 (Declared) - Const-parameter expected [line: 4, column: 2]`,
			`Syntax Error: Syntax Error: Method "Get" has incompatible parameters [line: 4, column: 2]`,
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "type ITest = interface\n function Get(Access: " + tt.actual + "): Integer;\n property P[const Declared: Integer]: Integer read\n Get;\nend;"
			assertClassPropertyIndexCompile(t, source, tt.want)
		})
	}
}

func assertClassPropertyIndexCompile(t *testing.T, source string, want []string) {
	t.Helper()
	engine, err := New()
	if err != nil {
		t.Fatal(err)
	}
	program, err := engine.Compile(source)
	if len(want) == 0 {
		if err != nil || program == nil {
			t.Fatalf("valid source: program nil = %t, error = %v", program == nil, err)
		}
		t.Log("Compile: program nil = false, diagnostics = []")
		return
	}
	if err == nil || program != nil {
		t.Fatalf("invalid source: program nil = %t, error = %v; want diagnostics:\n%s", program == nil, err, strings.Join(want, "\n"))
	}
	ce, ok := err.(*CompileError)
	if !ok {
		t.Fatalf("error = %T: %v", err, err)
	}
	got := make([]string, len(ce.Errors))
	for i, d := range ce.Errors {
		if d.Severity != SeverityError {
			t.Fatalf("diagnostic %d severity = %q", i, d.Severity)
		}
		got[i] = fmt.Sprintf("Syntax Error: %s [line: %d, column: %d]", d.Message, d.Line, d.Column)
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	t.Logf("Compile: program nil = true, diagnostics:\n%s", strings.Join(got, "\n"))
}
