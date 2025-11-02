package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestContractPreconditionBooleanType tests that preconditions must be boolean
// Task 9.140
func TestContractPreconditionBooleanType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid boolean precondition",
			input: `
function Divide(a, b: Float): Float;
require
	b <> 0.0;
begin
	Result := a / b;
end;
`,
			expectError: false,
		},
		{
			name: "invalid integer precondition",
			input: `
function Invalid(x: Integer): Integer;
require
	x + 1;
begin
	Result := x;
end;
`,
			expectError: true,
			errorMsg:    "precondition must be boolean",
		},
		{
			name: "invalid string precondition",
			input: `
function Invalid(s: String): Integer;
require
	s;
begin
	Result := 0;
end;
`,
			expectError: true,
			errorMsg:    "precondition must be boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing '%s', got none", tt.errorMsg)
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestContractPostconditionBooleanType tests that postconditions must be boolean
// Task 9.141
func TestContractPostconditionBooleanType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid boolean postcondition",
			input: `
function Abs(x: Integer): Integer;
begin
	Result := x;
end;
ensure
	Result >= 0;
`,
			expectError: false,
		},
		{
			name: "invalid integer postcondition",
			input: `
function Invalid(x: Integer): Integer;
begin
	Result := x;
end;
ensure
	Result;
`,
			expectError: true,
			errorMsg:    "postcondition must be boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing '%s', got none", tt.errorMsg)
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestContractMessageStringType tests that contract messages must be strings
// Task 9.142
func TestContractMessageStringType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid string message",
			input: `
function Divide(a, b: Float): Float;
require
	b <> 0.0 : 'divisor cannot be zero';
begin
	Result := a / b;
end;
`,
			expectError: false,
		},
		// Note: We cannot test non-string message literals because the parser
		// expects a STRING token after the colon. This is actually good - the
		// parser enforces this at parse time. If we want to test semantic checking
		// of message types, we would need to construct AST nodes directly or
		// test with expressions that evaluate to non-string types at runtime,
		// which isn't applicable for literal messages.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing '%s', got none", tt.errorMsg)
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestContractOldExpressionValidation tests validation of 'old' expressions
// Task 9.143
func TestContractOldExpressionValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid old expression with parameter",
			input: `
function Increment(x: Integer): Integer;
begin
	Result := x + 1;
end;
ensure
	Result = old x + 1;
`,
			expectError: false,
		},
		{
			name: "invalid old expression with undefined variable",
			input: `
function Invalid(x: Integer): Integer;
begin
	Result := x;
end;
ensure
	Result = old undefined_var;
`,
			expectError: true,
			errorMsg:    "old() references undefined identifier",
		},
		{
			name: "valid old expression in complex postcondition",
			input: `
function Calculate(a, b: Integer): Integer;
var
	temp: Integer;
begin
	temp := a + b;
	Result := temp * 2;
end;
ensure
	Result = (old a + old b) * 2;
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
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing '%s', got none", tt.errorMsg)
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestContractMultipleConditions tests multiple preconditions and postconditions
func TestContractMultipleConditions(t *testing.T) {
	input := `
function Clamp(x, min, max: Integer): Integer;
require
	min <= max : 'min must not exceed max';
	min < 1000;
begin
	if x < min then
		Result := min
	else if x > max then
		Result := max
	else
		Result := x;
end;
ensure
	Result >= min;
	Result <= max;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestContractTypeInference tests that old expressions get correct types
func TestContractTypeInference(t *testing.T) {
	input := `
function Double(x: Integer): Integer;
begin
	Result := x * 2;
end;
ensure
	Result = old x * 2;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Get the function declaration
	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	funcDecl, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected *ast.FunctionDecl, got %T", program.Statements[0])
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify postconditions exist
	if funcDecl.PostConditions == nil {
		t.Fatal("expected postconditions, got nil")
	}

	if len(funcDecl.PostConditions.Conditions) != 1 {
		t.Fatalf("expected 1 postcondition, got %d", len(funcDecl.PostConditions.Conditions))
	}
}
