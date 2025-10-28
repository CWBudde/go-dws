package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Task 9.216: Lambda Expression Semantic Analysis Tests
// ============================================================================

// TestLambdaExpressionBasic tests basic lambda expression analysis with explicit types.
// Task 9.216: Test valid lambda expressions
func TestLambdaExpressionBasic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple lambda with explicit types",
			input: `
				var f := lambda(x: Integer): Integer begin Result := x * 2; end;
			`,
		},
		{
			name: "lambda with multiple parameters",
			input: `
				var add := lambda(a: Integer; b: Integer): Integer begin Result := a + b; end;
			`,
		},
		{
			name: "shorthand lambda",
			input: `
				var double := lambda(x: Integer): Integer => x * 2;
			`,
		},
		{
			name: "procedure lambda (no return type)",
			input: `
				var print := lambda(msg: String) begin PrintLn(msg); end;
			`,
		},
		{
			name: "lambda with String parameters",
			input: `
				var concat := lambda(a: String; b: String): String begin Result := a + b; end;
			`,
		},
		{
			name: "lambda with Boolean return",
			input: `
				var isEven := lambda(n: Integer): Boolean begin Result := n mod 2 = 0; end;
			`,
		},
		{
			name: "nested lambdas in assignment",
			input: `
				var outer := lambda(x: Integer): Integer begin
					var inner := lambda(y: Integer): Integer begin Result := x + y; end;
					Result := x * 2;
				end;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}
		})
	}
}

// TestLambdaReturnTypeInference tests return type inference from lambda body.
// Task 9.216: Test return type inference
func TestLambdaReturnTypeInference(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "infer Integer from return statement",
			input: `
				var f := lambda(x: Integer) begin Result := 42; end;
			`,
		},
		{
			name: "infer String from return statement",
			input: `
				var f := lambda(x: Integer) begin Result := 'hello'; end;
			`,
		},
		{
			name: "infer Boolean from return statement",
			input: `
				var f := lambda(x: Integer) begin Result := true; end;
			`,
		},
		{
			name: "infer VOID from empty body",
			input: `
				var f := lambda(x: Integer) begin end;
			`,
		},
		{
			name: "infer from arithmetic expression",
			input: `
				var f := lambda(x: Integer) begin Result := x + 1; end;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}
		})
	}
}

// TestLambdaParameterScoping tests that lambda parameters are properly scoped.
// Task 9.216: Test parameter scoping
func TestLambdaParameterScoping(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "lambda parameter shadows outer variable",
			input: `
				var x: Integer := 10;
				var f := lambda(x: Integer): Integer begin Result := x * 2; end;
			`,
		},
		{
			name: "lambda can access outer scope",
			input: `
				var multiplier: Integer := 5;
				var f := lambda(x: Integer): Integer begin Result := x * multiplier; end;
			`,
		},
		{
			name: "multiple lambdas with same parameter names",
			input: `
				var f1 := lambda(x: Integer): Integer begin Result := x + 1; end;
				var f2 := lambda(x: Integer): Integer begin Result := x * 2; end;
			`,
		},
		{
			name: "lambda parameters in nested lambdas",
			input: `
				var f := lambda(x: Integer): Integer begin
					var g := lambda(y: Integer): Integer begin Result := x + y; end;
					Result := x;
				end;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}
		})
	}
}

// TestLambdaErrors tests error cases in lambda analysis.
// Task 9.216: Test error handling
func TestLambdaErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "duplicate parameter names",
			input: `
				var f := lambda(x: Integer; x: Integer): Integer begin Result := x; end;
			`,
			expectedErr: "duplicate parameter name",
		},
		{
			name: "unknown parameter type",
			input: `
				var f := lambda(x: UnknownType): Integer begin Result := 42; end;
			`,
			expectedErr: "unknown parameter type",
		},
		{
			name: "unknown return type",
			input: `
				var f := lambda(x: Integer): UnknownType begin Result := 42; end;
			`,
			expectedErr: "unknown return type",
		},
		{
			name: "missing parameter type annotation",
			input: `
				var f := lambda(x): Integer begin Result := x * 2; end;
			`,
			expectedErr: "not fully implemented",
		},
		{
			name: "conflicting return types",
			input: `
				var f := lambda(x: Integer) begin
					if x > 0 then
						Result := 42
					else
						Result := 'error';
				end;
			`,
			expectedErr: "conflicting return types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			_ = a.Analyze(program)

			// Check that we got an error
			if len(a.Errors()) == 0 {
				t.Errorf("expected semantic error containing '%s', got no errors", tt.expectedErr)
				return
			}

			// Check that the error message contains the expected substring
			foundExpected := false
			for _, err := range a.Errors() {
				if strings.Contains(err, tt.expectedErr) {
					foundExpected = true
					break
				}
			}
			if !foundExpected {
				t.Errorf("expected error containing '%s', got: %v", tt.expectedErr, a.Errors())
			}
		})
	}
}

// TestLambdaInFunctionCall tests lambdas used directly in function calls.
// Task 9.216: Test lambda in expression context
// Note: Function pointer type parameters in function declarations are not yet fully supported
// so we test lambda in assignment context instead
func TestLambdaInFunctionCall(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "lambda in variable assignment",
			input: `
				var f := lambda(n: Integer): Integer begin Result := n * 2; end;
				var x: Integer := 5;
			`,
		},
		{
			name: "shorthand lambda in assignment",
			input: `
				var f := lambda(n: Integer): Integer => n * 2;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}
		})
	}
}

// TestClosureCapture tests that lambdas correctly capture variables from outer scopes.
// Task 9.217: Test closure capture analysis
func TestClosureCapture(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedCaptures []string
	}{
		{
			name: "capture single outer variable",
			input: `
				var multiplier: Integer := 5;
				var f := lambda(x: Integer): Integer begin Result := x * multiplier; end;
			`,
			expectedCaptures: []string{"multiplier"},
		},
		{
			name: "capture multiple outer variables",
			input: `
				var a: Integer := 10;
				var b: Integer := 20;
				var f := lambda(x: Integer): Integer begin Result := x + a + b; end;
			`,
			expectedCaptures: []string{"a", "b"},
		},
		{
			name: "no capture when using only parameters",
			input: `
				var f := lambda(x: Integer; y: Integer): Integer begin Result := x + y; end;
			`,
			expectedCaptures: nil,
		},
		{
			name: "capture outer variable in outer lambda",
			input: `
				var outer: Integer := 100;
				var f := lambda(x: Integer): Integer begin Result := x + outer; end;
			`,
			expectedCaptures: []string{"outer"},
		},
		{
			name: "no capture of local variables",
			input: `
				var f := lambda(x: Integer): Integer begin
					var local: Integer := 5;
					Result := x + local;
				end;
			`,
			expectedCaptures: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}

			// Find the lambda expression in the AST
			var lambdaExpr *ast.LambdaExpression
			for _, stmt := range program.Statements {
				if varDecl, ok := stmt.(*ast.VarDeclStatement); ok {
					if lambda, ok := varDecl.Value.(*ast.LambdaExpression); ok {
						lambdaExpr = lambda
						break
					}
				}
			}

			if lambdaExpr == nil {
				t.Fatal("lambda expression not found in program")
			}

			// Check captured variables
			if tt.expectedCaptures == nil {
				if len(lambdaExpr.CapturedVars) != 0 {
					t.Errorf("expected no captures, got: %v", lambdaExpr.CapturedVars)
				}
			} else {
				// Check that all expected captures are present
				captureMap := make(map[string]bool)
				for _, v := range lambdaExpr.CapturedVars {
					captureMap[v] = true
				}

				for _, expected := range tt.expectedCaptures {
					if !captureMap[expected] {
						t.Errorf("expected to capture '%s', but it was not captured. Captured: %v",
							expected, lambdaExpr.CapturedVars)
					}
				}

				// Check that we didn't capture extra variables
				if len(lambdaExpr.CapturedVars) != len(tt.expectedCaptures) {
					t.Errorf("expected %d captures, got %d: %v",
						len(tt.expectedCaptures), len(lambdaExpr.CapturedVars), lambdaExpr.CapturedVars)
				}
			}
		})
	}
}

// TestLambdaWithResultVariable tests that Result variable is properly added to scope.
// Task 9.216: Test Result variable in lambda scope
func TestLambdaWithResultVariable(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "lambda with Result assignment",
			input: `
				var f := lambda(x: Integer): Integer begin
					Result := x * 2;
				end;
			`,
		},
		{
			name: "lambda with multiple Result assignments",
			input: `
				var f := lambda(x: Integer): Integer begin
					if x > 0 then
						Result := x
					else
						Result := -x;
				end;
			`,
		},
		{
			name: "lambda with Result in expression",
			input: `
				var f := lambda(x: Integer): Integer begin
					Result := 0;
					Result := Result + x;
				end;
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			a := NewAnalyzer()
			err := a.Analyze(program)
			if err != nil {
				t.Errorf("expected no semantic errors, got: %v", err)
			}
			if len(a.Errors()) != 0 {
				t.Errorf("expected no errors, got: %v", a.Errors())
			}
		})
	}
}
