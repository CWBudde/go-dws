package interp

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestHighWithTypeMetaValues tests High() with type meta-values.
// Task 9.134: High() should work with type names (Integer, Float, Boolean).
func TestHighWithTypeMetaValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "High(Integer)",
			input:    "PrintLn(High(Integer));",
			expected: "9223372036854775807\n", // math.MaxInt64
		},
		{
			name:     "High(Boolean)",
			input:    "PrintLn(High(Boolean));",
			expected: "true\n",
		},
		{
			name:     "High(Integer) in expression",
			input:    "var x: Integer := High(Integer); PrintLn(x);",
			expected: "9223372036854775807\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
			}

			var buf bytes.Buffer
			interp := New(&buf)
			val := interp.Eval(program)

			// Check for errors
			if errVal, ok := val.(*ErrorValue); ok {
				t.Fatalf("evaluation error: %s", errVal.Message)
			}

			output := buf.String()
			if output != tt.expected {
				t.Errorf("output mismatch:\ngot:  %q\nwant: %q", output, tt.expected)
			}
		})
	}
}

// TestLowWithTypeMetaValues tests Low() with type meta-values.
// Task 9.134: Low() should work with type names (Integer, Float, Boolean).
func TestLowWithTypeMetaValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Low(Integer)",
			input:    "PrintLn(Low(Integer));",
			expected: "-9223372036854775808\n", // math.MinInt64
		},
		{
			name:     "Low(Boolean)",
			input:    "PrintLn(Low(Boolean));",
			expected: "false\n",
		},
		{
			name:     "Low(Integer) in expression",
			input:    "var x: Integer := Low(Integer); PrintLn(x);",
			expected: "-9223372036854775808\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
			}

			var buf bytes.Buffer
			interp := New(&buf)
			val := interp.Eval(program)

			// Check for errors
			if errVal, ok := val.(*ErrorValue); ok {
				t.Fatalf("evaluation error: %s", errVal.Message)
			}

			output := buf.String()
			if output != tt.expected {
				t.Errorf("output mismatch:\ngot:  %q\nwant: %q", output, tt.expected)
			}
		})
	}
}

// TestHighLowWithFloatType tests High/Low with Float type meta-value.
func TestHighLowWithFloatType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkFn  func(t *testing.T, output string)
	}{
		{
			name:  "High(Float)",
			input: "PrintLn(High(Float));",
			checkFn: func(t *testing.T, output string) {
				// Check that the output contains a very large number
				// math.MaxFloat64 is approximately 1.7976931348623157e+308
				if !strings.Contains(output, "e+") {
					t.Errorf("expected scientific notation in output, got: %q", output)
				}
			},
		},
		{
			name:  "Low(Float)",
			input: "PrintLn(Low(Float));",
			checkFn: func(t *testing.T, output string) {
				// Check that the output contains a very large negative number
				if !strings.Contains(output, "-") {
					t.Errorf("expected negative number in output, got: %q", output)
				}
				if !strings.Contains(output, "e+") {
					t.Errorf("expected scientific notation in output, got: %q", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
			}

			var buf bytes.Buffer
			interp := New(&buf)
			val := interp.Eval(program)

			// Check for errors
			if errVal, ok := val.(*ErrorValue); ok {
				t.Fatalf("evaluation error: %s", errVal.Message)
			}

			output := buf.String()
			tt.checkFn(t, output)
		})
	}
}

// TestHighLowDirectValues tests High/Low return the correct runtime values.
func TestHighLowDirectValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkFn  func(t *testing.T, val Value)
	}{
		{
			name:  "High(Integer) returns IntegerValue",
			input: "High(Integer)",
			checkFn: func(t *testing.T, val Value) {
				intVal, ok := val.(*IntegerValue)
				if !ok {
					t.Fatalf("expected IntegerValue, got %T", val)
				}
				if intVal.Value != math.MaxInt64 {
					t.Errorf("expected %d, got %d", math.MaxInt64, intVal.Value)
				}
			},
		},
		{
			name:  "Low(Integer) returns IntegerValue",
			input: "Low(Integer)",
			checkFn: func(t *testing.T, val Value) {
				intVal, ok := val.(*IntegerValue)
				if !ok {
					t.Fatalf("expected IntegerValue, got %T", val)
				}
				if intVal.Value != math.MinInt64 {
					t.Errorf("expected %d, got %d", math.MinInt64, intVal.Value)
				}
			},
		},
		{
			name:  "High(Boolean) returns true",
			input: "High(Boolean)",
			checkFn: func(t *testing.T, val Value) {
				boolVal, ok := val.(*BooleanValue)
				if !ok {
					t.Fatalf("expected BooleanValue, got %T", val)
				}
				if !boolVal.Value {
					t.Errorf("expected true, got false")
				}
			},
		},
		{
			name:  "Low(Boolean) returns false",
			input: "Low(Boolean)",
			checkFn: func(t *testing.T, val Value) {
				boolVal, ok := val.(*BooleanValue)
				if !ok {
					t.Fatalf("expected BooleanValue, got %T", val)
				}
				if boolVal.Value {
					t.Errorf("expected false, got true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
			}

			var buf bytes.Buffer
			interp := New(&buf)
			val := interp.Eval(program)

			// Check for errors
			if errVal, ok := val.(*ErrorValue); ok {
				t.Fatalf("evaluation error: %s", errVal.Message)
			}

			tt.checkFn(t, val)
		})
	}
}

// TestTypeMetaValueInEnvironment tests that type meta-values are registered in the environment.
// Task 9.133: Type names should be available as identifiers.
func TestTypeMetaValueInEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Integer type meta-value exists",
			input:    "PrintLn(Integer);",
			expected: "Integer\n",
		},
		{
			name:     "Float type meta-value exists",
			input:    "PrintLn(Float);",
			expected: "Float\n",
		},
		{
			name:     "Boolean type meta-value exists",
			input:    "PrintLn(Boolean);",
			expected: "Boolean\n",
		},
		{
			name:     "String type meta-value exists",
			input:    "PrintLn(String);",
			expected: "String\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %s", strings.Join(p.Errors(), "\n"))
			}

			var buf bytes.Buffer
			interp := New(&buf)
			val := interp.Eval(program)

			// Check for errors
			if errVal, ok := val.(*ErrorValue); ok {
				t.Fatalf("evaluation error: %s", errVal.Message)
			}

			output := buf.String()
			if output != tt.expected {
				t.Errorf("output mismatch:\ngot:  %q\nwant: %q", output, tt.expected)
			}
		})
	}
}
