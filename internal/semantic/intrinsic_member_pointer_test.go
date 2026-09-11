package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// analyzeIntrinsicPointerSource parses and analyzes src, returning the
// accumulated semantic errors.
func analyzeIntrinsicPointerSource(t *testing.T, src string) []string {
	t.Helper()
	p := parser.New(lexer.New(src))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	a := NewAnalyzer()
	_ = a.Analyze(program)
	return a.Errors()
}

// TestIntrinsicMemberPointer_Accepted covers the positions in which TObject's
// intrinsic parameterless members (ClassName, ClassType) form a pointer instead
// of being read eagerly.
func TestIntrinsicMemberPointer_Accepted(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "address-of class type through metaclass",
			input: `var proc := @TObject.ClassType;`,
		},
		{
			name:  "address-of class name through metaclass",
			input: `var proc := @TObject.ClassName;`,
		},
		{
			name: "pointer reads its result in receiver position",
			input: `
				var proc := @TObject.ClassType;
				PrintLn(proc.ClassName);
			`,
		},
		{
			name: "class type coerces into array of parameterless function",
			input: `
				var a : array of function : TClass;
				a.Add(TObject.ClassType);
			`,
		},
		{
			name: "class name coerces into array of parameterless function",
			input: `
				var a : array of function : String;
				a.Add(TObject.ClassName);
			`,
		},
		{
			name: "derived class type is accepted in a TClass-returning slot",
			input: `
				type TClassA = class(TObject);
				var a : array of function : TClass;
				a.Add(TClassA.ClassType);
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errs := analyzeIntrinsicPointerSource(t, tt.input); len(errs) != 0 {
				t.Errorf("expected no semantic errors, got: %v", errs)
			}
		})
	}
}

// TestIntrinsicMemberPointer_Rejected keeps the intrinsic coercion narrow: it
// applies only to parameterless targets whose result type actually accepts the
// intrinsic's own result.
func TestIntrinsicMemberPointer_Rejected(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "result type mismatch is still reported",
			input: `
				var a : array of function : Integer;
				a.Add(TObject.ClassName);
			`,
			expectedErr: "Incompatible parameter types",
		},
		{
			name: "a target with parameters does not capture the intrinsic",
			input: `
				var a : array of function (x : Integer) : String;
				a.Add(TObject.ClassName);
			`,
			expectedErr: "Incompatible parameter types",
		},
		{
			name: "unbound instance method pointers remain unsupported",
			input: `
				type TThing = class
					function Value : Integer;
				end;
				function TThing.Value : Integer;
				begin
					Result := 1;
				end;
				var p := @TThing.Value;
			`,
			expectedErr: "unbound method pointers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := analyzeIntrinsicPointerSource(t, tt.input)
			found := false
			for _, err := range errs {
				if strings.Contains(err, tt.expectedErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected an error containing %q, got: %v", tt.expectedErr, errs)
			}
		})
	}
}
