package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestTypeConversionWithExperimentalPasses tests type conversions and coercions
// with experimental passes enabled.
// Task 6.1.2.6: Type Conversion/Coercion
func TestTypeConversionWithExperimentalPasses(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name: "implicit Integer to Float in assignment",
			input: `
				var i: Integer := 42;
				var f: Float;
				f := i;
			`,
			expectError: false,
		},
		{
			name: "implicit Integer to Float in variable declaration",
			input: `
				var i: Integer := 42;
				var f: Float := i;
			`,
			expectError: false,
		},
		{
			name: "integer literal to Float variable",
			input: `
				var f: Float := 42;
			`,
			expectError: false,
		},
		{
			name: "integer literal in Float expression",
			input: `
				var f: Float;
				f := 10 + 20;
			`,
			expectError: false,
		},
		{
			name: "Integer to Float in function parameter",
			input: `
				procedure TakeFloat(f: Float); begin end;
				var i: Integer := 42;
				TakeFloat(i);
			`,
			expectError: false,
		},
		{
			name: "Integer literal to Float parameter",
			input: `
				procedure TakeFloat(f: Float); begin end;
				TakeFloat(42);
			`,
			expectError: false,
		},
		{
			name: "Integer to Float in function return",
			input: `
				function GetFloat: Float;
				begin
					var i: Integer := 42;
					Result := i;
				end;
			`,
			expectError: false,
		},
		{
			name: "no implicit Float to Integer",
			input: `
				var f: Float := 3.14;
				var i: Integer;
				i := f;
			`,
			expectError:   true,
			errorContains: "cannot assign",
		},
		{
			name: "Integer and Float in arithmetic",
			input: `
				var i: Integer := 10;
				var f: Float := 3.14;
				var result: Float;
				result := i + f;
			`,
			expectError: false,
		},
		{
			name: "Integer to Float in comparison",
			input: `
				var i: Integer := 10;
				var f: Float := 10.5;
				var b: Boolean;
				b := i < f;
			`,
			expectError: false,
		},
		{
			name: "Integer to Float in record field",
			input: `
				type TRec = record
					f: Float;
				end;
				var r: TRec;
				var i: Integer := 42;
				r.f := i;
			`,
			expectError: false,
		},
		{
			name: "Integer to Float in array element",
			input: `
				var arr: array of Float;
				var i: Integer := 42;
				arr.Add(i);
			`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Use experimental passes
			analyzer := NewAnalyzerWithExperimentalPasses()
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				errors := analyzer.Errors()
				if len(errors) == 0 {
					t.Errorf("Expected error messages but got none")
					return
				}
				if tt.errorContains != "" {
					found := false
					for _, errMsg := range errors {
						if strings.Contains(errMsg, tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error containing '%s', got errors: %v",
							tt.errorContains, errors)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v\nErrors: %v", err, analyzer.Errors())
				}
			}
		})
	}
}

// TestIntegerLiteralToFloatContext tests integer literal type inference
// Task 6.1.2.6: Integer literal to Float context inference
func TestIntegerLiteralToFloatContext(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "integer literal in Float declaration",
			input: `
				var f: Float := 100;
			`,
			expectError: false,
		},
		{
			name: "integer literal in Float assignment",
			input: `
				var f: Float;
				f := 42;
			`,
			expectError: false,
		},
		{
			name: "integer literal as Float function parameter",
			input: `
				function Square(x: Float): Float;
				begin
					Result := x * x;
				end;
				var r: Float := Square(5);
			`,
			expectError: false,
		},
		{
			name: "integer literal in Float arithmetic",
			input: `
				var f: Float := 10;
				f := f + 5;
			`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Use experimental passes
			analyzer := NewAnalyzerWithExperimentalPasses()
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v\nErrors: %v", err, analyzer.Errors())
				}
			}
		})
	}
}

// TestNumericCoercionInBinaryOperations tests type coercion in expressions
// Task 6.1.2.6: Support numeric type promotion in operations
func TestNumericCoercionInBinaryOperations(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "Integer + Float returns Float",
			input: `
				var i: Integer := 10;
				var f: Float := 3.14;
				var result: Float := i + f;
			`,
			expectError: false,
		},
		{
			name: "Float + Integer returns Float",
			input: `
				var f: Float := 3.14;
				var i: Integer := 10;
				var result: Float := f + i;
			`,
			expectError: false,
		},
		{
			name: "Integer * Float returns Float",
			input: `
				var i: Integer := 10;
				var f: Float := 2.5;
				var result: Float := i * f;
			`,
			expectError: false,
		},
		{
			name: "Integer - Float returns Float",
			input: `
				var i: Integer := 100;
				var f: Float := 3.14;
				var result: Float := i - f;
			`,
			expectError: false,
		},
		{
			name: "Integer / Float returns Float",
			input: `
				var i: Integer := 100;
				var f: Float := 2.0;
				var result: Float := i / f;
			`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Use experimental passes
			analyzer := NewAnalyzerWithExperimentalPasses()
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v\nErrors: %v", err, analyzer.Errors())
				}
			}
		})
	}
}
