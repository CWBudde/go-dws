package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompile_PropertyOperatorDiagnostics(t *testing.T) {
	for _, fixture := range []string{
		"FailureScripts/property_error3", "FailureScripts/property_error4",
		"FailureScripts/class_operator3", "OperatorOverloadFail/operator_overload1",
	} {
		t.Run(fixture, func(t *testing.T) {
			base := filepath.Join("../../testdata/fixtures", fixture)
			source, err := os.ReadFile(base + ".pas")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(base + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			assertDiagnostics(t, string(source), base+".pas", strings.Split(strings.TrimSpace(string(expected)), "\n"))
		})
	}
}

func TestCompile_PropertyAccessorStop(t *testing.T) {
	for _, accessor := range []string{"read", "write"} {
		source := "type T = class\n  property P: Integer " + accessor + "\n  {comment}\n  ;\n  property Q: Unknown;\nend;\nMissing;"
		assertDiagnostics(t, source, "<test>", []string{`Syntax Error: Name expected [line: 4, column: 3]`})
	}
}

func TestCompile_PropertyValidationOrderAndAnchors(t *testing.T) {
	tests := []struct {
		name, member, property string
		messages               []string
	}{
		{"setter result before arity", "function GeT: Integer;", "P: String write get", []string{`Procedure expected`}},
		{"getter result before arity", "function GeT(a: Integer): Integer;", "P: String read get", []string{`Field/method "GeT" has an incompatible type`}},
		{"getter arity", "function GeT(a: Integer): Integer;", "P: Integer read get", []string{`Method "GeT" has incompatible parameters`}},
		{"setter arity", "procedure SeT;", "P: Integer write set", []string{`Method "SeT" has incompatible parameters`}},
		{"indexed getter", "function GeT(a: Integer): Integer;", "P[i: String]: Integer read get", []string{`Parameter 0 - Type "String" expected (instead of "Integer")`, `Method "GeT" has incompatible parameters`}},
		{"numeric getter index", "function GeT(a: Float): Integer;", "P[i: Integer]: Integer read get", []string{`Parameter 0 - Type "Integer" expected (instead of "Float")`, `Method "GeT" has incompatible parameters`}},
		{"constant casing", "const MiXeD = 1;", "P: String read mixed", []string{`Field/method "MiXeD" has an incompatible type`}},
		{"setter value type", "procedure SeT(v: String);", "P: Integer write set", []string{`Method "SeT" has incompatible parameters`}},
		{"read field", "FiElD: Integer;", "P: String read field", []string{`Field/method "FiElD" has an incompatible type`}},
		{"write field", "FiElD: Integer;", "P: String write field", []string{`Symbol "FiElD" has an incompatible type`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := strings.LastIndex(tt.property, " ")
			source := "type T = class\n  " + tt.member + "\n  property " + tt.property[:split] + "\n  {padding}\n  " + tt.property[split+1:] + ";\n  property Stop: Integer read ;\nend;"
			want := make([]string, 0, len(tt.messages)+1)
			for _, message := range tt.messages {
				want = append(want, "Syntax Error: "+message+" [line: 5, column: 3]")
			}
			want = append(want, `Syntax Error: Name expected [line: 6, column: 31]`)
			assertDiagnostics(t, source, "<test>", want)
		})
	}
}

func TestCompile_GlobalOperatorValidation(t *testing.T) {
	tests := []struct {
		name, source string
		want         []string
	}{
		{"result before binding arity", "function F: String; begin Result := ''; end;\noperator + (Integer) : Integer uses F;", []string{
			`Syntax Error: Expected 2 parameters (instead of 1) [line: 2, column: 20]`,
			`Syntax Error: Result type should be "Integer" [line: 2, column: 37]`,
		}},
		{"binding arity uses operator arity", "function F: Integer; begin Result := 1; end;\noperator + (Integer) : Integer uses F;", []string{
			`Syntax Error: Expected 2 parameters (instead of 1) [line: 2, column: 20]`,
			`Syntax Error: Expected 2 parameters (instead of 0) [line: 2, column: 37]`,
		}},
		{"binding type", "function F(a: String; b: Integer): Integer; begin Result := b; end;\noperator + (Integer, Integer) : Integer uses F;", []string{
			`Syntax Error: Parameter 0 - Type "Integer" expected (instead of "String") [line: 2, column: 46]`,
		}},
		{"var binding parameter", "function F(var a: Integer; b: Integer): Integer; begin Result := b; end;\noperator + (Integer, Integer) : Integer uses F;", []string{
			`Syntax Error: Parameter 0 - Var-parameter forbidden [line: 2, column: 46]`,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertDiagnostics(t, tt.source, "<test>", tt.want) })
	}
}

func TestCompile_ValidPropertyAndOperatorDeclarations(t *testing.T) {
	for _, source := range []string{
		"type T = class F: Integer; property P: Integer read F write F; end; var o := T.Create; o.P := 2; PrintLn(o.P);",
		"type T = class function Get(i: Integer): Integer; begin Result := i; end; procedure Put(i, v: Integer); begin end; property P[i: Integer]: Integer read Get write Put; end; var o := T.Create; o.P[1] := 2; PrintLn(o.P[1]);",
		"function F(s: String; i: Integer): String; begin Result := s; end; operator + (String, Integer): String uses F; PrintLn('a' + 1);",
		"type T = record X: Integer; end; function F(v: T): T; begin Result.X := -v.X; end; operator - (T): T uses F; var v: T; var w := -v;",
	} {
		t.Run(source, func(t *testing.T) { assertDiagnostics(t, source, "<test>", nil) })
	}
}

func TestCompile_GlobalOperatorOperandAnchor(t *testing.T) {
	source := "function F(a,b: Integer): Integer; begin Result := a+b; end;\noperator + (Integer\n) {comment}\n: Integer uses F;"
	assertDiagnostics(t, source, "<test>", []string{`Syntax Error: Expected 2 parameters (instead of 1) [line: 3, column: 1]`})
}

func TestCompile_OperatorBindingRejectsNumericConversions(t *testing.T) {
	for _, tt := range []struct {
		name, source, message string
		column                int
	}{
		{"result", "function F(a,b: Integer): Integer; begin Result := a; end;\noperator + (Integer, Integer): Float uses F;", `Result type should be "Float"`, 43},
		{"operand", "function F(a,b: Integer): Integer; begin Result := a+b; end;\noperator + (Float, Integer): Integer uses F;", `Parameter 0 - Type "Float" expected (instead of "Integer")`, 43},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertDiagnostics(t, tt.source, "<test>", []string{fmt.Sprintf("Syntax Error: %s [line: 2, column: %d]", tt.message, tt.column)})
		})
	}
}

func TestCompile_VariantPropertyIndexSignature(t *testing.T) {
	source := "type T = class function Get(i: Integer): Integer; begin Result := i; end; property P[i: Variant]: Integer read Get; end;"
	assertDiagnostics(t, source, "<test>", nil)
}

func TestCompile_OperatorBindingRejectsArrayElementConversion(t *testing.T) {
	source := "function F(a,b: Integer): array of Integer; begin Result := [a]; end;\noperator + (Integer, Integer): array of Variant uses F;"
	assertDiagnostics(t, source, "<test>", []string{`Syntax Error: Result type should be "array of Variant" [line: 2, column: 54]`})
}
