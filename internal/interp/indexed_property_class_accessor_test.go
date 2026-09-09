package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// TestIndexedPropertyWithClassMethodAccessor covers an indexed property whose
// accessor is a class method. DWScript allows such a property to be reached through
// the class name — the accessor needs no instance — as well as through an instance.
// Both used to fail: the accessor lookup consulted only the instance method table,
// and VisitIndexExpression had no metaclass receiver branch at all.
func TestIndexedPropertyWithClassMethodAccessor(t *testing.T) {
	const decl = `
type TC = class
	class var Store : array [0..4] of String;
	class function Get(i : Integer) : String;
	begin
		Result := Store[i];
	end;
	class procedure Put(i : Integer; const v : String);
	begin
		Store[i] := v;
	end;
	property Prop[i : Integer] : String read Get write Put;
end;
`

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "read through the class name",
			input:    decl + "TC.Prop[1] := 'a';\nPrintLn(TC.Prop[1]);\n",
			expected: "a\n",
		},
		{
			name:     "read through an instance",
			input:    decl + "var o := TC.Create;\no.Prop[2] := 'b';\nPrintLn(o.Prop[2]);\n",
			expected: "b\n",
		},
		{
			name:     "write through the class name, read through an instance",
			input:    decl + "TC.Prop[3] := 'c';\nvar o := TC.Create;\nPrintLn(o.Prop[3]);\n",
			expected: "c\n",
		},
		{
			name:     "write through an instance, read through the class name",
			input:    decl + "var o := TC.Create;\no.Prop[4] := 'd';\nPrintLn(TC.Prop[4]);\n",
			expected: "d\n",
		},
		{
			name: "expression accessor through the class name",
			input: `
type TC = class
	class var Store : array [0..4] of Integer;
	property Prop[i : Integer] : Integer read (Store[i] * 2);
end;
TC.Store[1] := 21;
PrintLn(TC.Prop[1]);
`,
			expected: "42\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			interp := New(&out)
			if result := interpret(interp, tt.input); isError(result) {
				t.Fatalf("interpreter error: %s", result.String())
			}
			if out.String() != tt.expected {
				t.Errorf("got %q, want %q", out.String(), tt.expected)
			}
		})
	}
}

// TestIndexedPropertyMetaclassDiagnostics pins the compile-time rules for reaching an
// indexed property through a class name. Semantic analysis used to accept anything
// the indexed path saw, disagreeing with the evaluator about what it could execute.
func TestIndexedPropertyMetaclassDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name: "instance-method getter is not reachable through the class",
			input: `
type TC = class
	FStore : array [0..4] of String;
	function Get(i : Integer) : String;
	begin
		Result := FStore[i];
	end;
	property Prop[i : Integer] : String read Get;
end;
PrintLn(TC.Prop[1]);
`,
			want: []string{
				"Read access of property should be a static method",
				"Class method or constructor expected",
			},
		},
		{
			name: "instance-method setter is not reachable through the class",
			input: `
type TC = class
	FStore : array [0..4] of String;
	class function Get(i : Integer) : String;
	begin
		Result := 'x';
	end;
	procedure Put(i : Integer; const v : String);
	begin
		FStore[i] := v;
	end;
	property Prop[i : Integer] : String read Get write Put;
end;
TC.Prop[1] := 'a';
`,
			want: []string{
				"Write access of property should be a static method",
				"Class method or constructor expected",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := frontend.Compile(tt.input, "test.dws", semantic.HintsLevelPedantic)
			diagnostics := strings.Join(result.DiagnosticStrings(), "\n")
			for _, want := range tt.want {
				if !strings.Contains(diagnostics, want) {
					t.Errorf("missing diagnostic %q in:\n%s", want, diagnostics)
				}
			}
		})
	}
}
