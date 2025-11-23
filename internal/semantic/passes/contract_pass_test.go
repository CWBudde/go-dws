package passes_test

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/semantic"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestContractPassPreconditionBoolean tests that preconditions must be boolean
func TestContractPassPreconditionBoolean(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errorMsg    string
		expectError bool
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
	42;
begin
	Result := x;
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

			// Run all passes in order
			ctx := semantic.NewPassContext()

			// Pass 1: Declaration
			declPass := semantic.NewDeclarationPass()
			if err := declPass.Run(program, ctx); err != nil {
				t.Fatalf("declaration pass error: %v", err)
			}

			// Pass 2: Type Resolution
			typePass := semantic.NewTypeResolutionPass()
			if err := typePass.Run(program, ctx); err != nil {
				t.Fatalf("type resolution pass error: %v", err)
			}

			// Skip Pass 3: Validation (not needed for contract validation)

			// Pass 4: Contract Validation
			contractPass := semantic.NewContractPass()
			if err := contractPass.Run(program, ctx); err != nil {
				t.Fatalf("contract pass error: %v", err)
			}

			// Check for expected errors
			hasError := len(ctx.Errors) > 0
			if tt.expectError && !hasError {
				t.Fatalf("expected error containing '%s', got none", tt.errorMsg)
			}

			if tt.expectError {
				found := false
				for _, err := range ctx.Errors {
					if strings.Contains(err, tt.errorMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, ctx.Errors)
				}
			} else if hasError {
				t.Errorf("unexpected errors: %v", ctx.Errors)
			}
		})
	}
}

// TestContractPassPostconditionBoolean tests that postconditions must be boolean
func TestContractPassPostconditionBoolean(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errorMsg    string
		expectError bool
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
	42;
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

			// Run all passes in order
			ctx := semantic.NewPassContext()

			// Pass 1: Declaration
			declPass := semantic.NewDeclarationPass()
			if err := declPass.Run(program, ctx); err != nil {
				t.Fatalf("declaration pass error: %v", err)
			}

			// Pass 2: Type Resolution
			typePass := semantic.NewTypeResolutionPass()
			if err := typePass.Run(program, ctx); err != nil {
				t.Fatalf("type resolution pass error: %v", err)
			}

			// Skip Pass 3: Validation (not needed for contract validation)

			// Pass 4: Contract Validation
			contractPass := semantic.NewContractPass()
			if err := contractPass.Run(program, ctx); err != nil {
				t.Fatalf("contract pass error: %v", err)
			}

			// Check for expected errors
			hasError := len(ctx.Errors) > 0
			if tt.expectError && !hasError {
				t.Fatalf("expected error containing '%s', got none", tt.errorMsg)
			}

			if tt.expectError {
				found := false
				for _, err := range ctx.Errors {
					if strings.Contains(err, tt.errorMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, ctx.Errors)
				}
			} else if hasError {
				t.Errorf("unexpected errors: %v", ctx.Errors)
			}
		})
	}
}

// TestContractPassOldExpression tests validation of 'old' expressions
func TestContractPassOldExpression(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errorMsg    string
		expectError bool
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
		// Note: The parser already rejects 'old' in preconditions, so we can't test
		// this at the semantic level. The parser error is:
		// "'old' keyword can only be used in postconditions"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			// Run all passes in order
			ctx := semantic.NewPassContext()

			// Pass 1: Declaration
			declPass := semantic.NewDeclarationPass()
			if err := declPass.Run(program, ctx); err != nil {
				t.Fatalf("declaration pass error: %v", err)
			}

			// Pass 2: Type Resolution
			typePass := semantic.NewTypeResolutionPass()
			if err := typePass.Run(program, ctx); err != nil {
				t.Fatalf("type resolution pass error: %v", err)
			}

			// Skip Pass 3: Validation (not needed for contract validation)

			// Pass 4: Contract Validation
			contractPass := semantic.NewContractPass()
			if err := contractPass.Run(program, ctx); err != nil {
				t.Fatalf("contract pass error: %v", err)
			}

			// Check for expected errors
			hasError := len(ctx.Errors) > 0
			if tt.expectError && !hasError {
				t.Fatalf("expected error containing '%s', got none", tt.errorMsg)
			}

			if tt.expectError {
				found := false
				for _, err := range ctx.Errors {
					if strings.Contains(err, tt.errorMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, ctx.Errors)
				}
			} else if hasError {
				t.Errorf("unexpected errors: %v", ctx.Errors)
			}
		})
	}
}

// TestContractPassMultipleConditions tests multiple preconditions and postconditions
func TestContractPassMultipleConditions(t *testing.T) {
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

	// Run all passes in order
	ctx := semantic.NewPassContext()

	// Pass 1: Declaration
	declPass := semantic.NewDeclarationPass()
	if err := declPass.Run(program, ctx); err != nil {
		t.Fatalf("declaration pass error: %v", err)
	}

	// Pass 2: Type Resolution
	typePass := semantic.NewTypeResolutionPass()
	if err := typePass.Run(program, ctx); err != nil {
		t.Fatalf("type resolution pass error: %v", err)
	}

	// Skip Pass 3: Validation (not needed for contract validation)

	// Pass 4: Contract Validation
	contractPass := semantic.NewContractPass()
	if err := contractPass.Run(program, ctx); err != nil {
		t.Fatalf("contract pass error: %v", err)
	}

	// Should have no errors
	if len(ctx.Errors) > 0 {
		t.Errorf("unexpected errors: %v", ctx.Errors)
	}
}

// TestContractPassInClassMethod tests contracts in class methods
func TestContractPassInClassMethod(t *testing.T) {
	input := `
type TStack = class
private
	FCount: Integer;
public
	procedure Push(Item: Integer);
	require
		FCount < 100;
	begin
		FCount := FCount + 1;
	end;

	function Pop: Integer;
	require
		FCount > 0;
	begin
		Result := 0;
		FCount := FCount - 1;
	end;
	ensure
		FCount >= 0;
end;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Run all passes in order
	ctx := semantic.NewPassContext()

	// Pass 1: Declaration
	declPass := semantic.NewDeclarationPass()
	if err := declPass.Run(program, ctx); err != nil {
		t.Fatalf("declaration pass error: %v", err)
	}

	// Pass 2: Type Resolution
	typePass := semantic.NewTypeResolutionPass()
	if err := typePass.Run(program, ctx); err != nil {
		t.Fatalf("type resolution pass error: %v", err)
	}

	// Skip Pass 3: Validation (not needed for contract validation)

	// Pass 4: Contract Validation
	contractPass := semantic.NewContractPass()
	if err := contractPass.Run(program, ctx); err != nil {
		t.Fatalf("contract pass error: %v", err)
	}

	// Should have no errors
	if len(ctx.Errors) > 0 {
		t.Errorf("unexpected errors: %v", ctx.Errors)
	}
}
