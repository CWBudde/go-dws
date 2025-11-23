package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Lambda Expression Semantic Analysis Tests
// ============================================================================

// TestLambdaExpressionBasic tests basic lambda expression analysis with explicit types.
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

// TestLambdaParameterTypeInference tests parameter type inference from context.
func TestLambdaParameterTypeInference(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "infer single parameter type from variable declaration",
			input: `
				type TIntFunc = function(x: Integer): Integer;
				var f: TIntFunc := lambda(x) => x * 2;
			`,
		},
		{
			name: "infer multiple parameter types",
			input: `
				type TBinaryFunc = function(a, b: Integer): Integer;
				var add: TBinaryFunc := lambda(a, b) => a + b;
			`,
		},
		{
			name: "infer parameter type from function call argument",
			input: `
				type TFunc = function(n: Integer): Integer;
				function Apply(value: Integer; f: TFunc): Integer;
				begin
					Result := f(value);
				end;
				var result := Apply(5, lambda(n) => n * 2);
			`,
		},
		{
			name: "partial type annotation - some explicit some inferred",
			input: `
				type TMixedFunc = function(a: Integer; b: String): String;
				var f: TMixedFunc := lambda(a: Integer; b) => IntToStr(a) + b;
			`,
		},
		{
			name: "infer parameter and return types",
			input: `
				type TFunc = function(x: Integer): Integer;
				var square: TFunc := lambda(x) => x * x;
			`,
		},
		{
			name: "nested lambda with inference",
			input: `
				type TFunc = function(n: Integer): Integer;
				type THigherOrder = function(f: TFunc): TFunc;
				var composer: THigherOrder := lambda(f) => lambda(x: Integer) => f(f(x));
			`,
		},
		{
			name: "lambda in return statement",
			input: `
				type TFunc = function(x: Integer): Integer;
				function MakeMultiplier(factor: Integer): TFunc;
				begin
					Result := lambda(x) => x * factor;
				end;
			`,
		},
		{
			name: "infer with procedure type",
			input: `
				type TProc = procedure(x: Integer);
				var p: TProc := lambda(x) begin PrintLn(x); end;
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

// TestLambdaParameterTypeInferenceErrors tests error cases for parameter type inference.
func TestLambdaParameterTypeInferenceErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "parameter count mismatch - too few",
			input: `
				type TBinaryFunc = function(a, b: Integer): Integer;
				var f: TBinaryFunc := lambda(x) => x * 2;
			`,
			expectedErr: "lambda has 1 parameter but expected function type has 2 parameters",
		},
		{
			name: "parameter count mismatch - too many",
			input: `
				type TUnaryFunc = function(x: Integer): Integer;
				var f: TUnaryFunc := lambda(a, b) => a + b;
			`,
			expectedErr: "lambda has 2 parameters but expected function type has 1 parameter",
		},
		{
			name: "incompatible explicit parameter type",
			input: `
				type TIntFunc = function(x: Integer): Integer;
				var f: TIntFunc := lambda(x: String) => Length(x);
			`,
			expectedErr: "parameter 'x' has type String but expected type requires Integer",
		},
		{
			name: "incompatible return type with explicit type",
			input: `
				type TFunc = function(x: Integer): Integer;
				var f: TFunc := lambda(x): String => IntToStr(x);
			`,
			expectedErr: "lambda return type String incompatible with expected return type Integer",
		},
		{
			name: "incompatible inferred return type",
			input: `
				type TFunc = function(x: Integer): Integer;
				var f: TFunc := lambda(x) => IntToStr(x);
			`,
			expectedErr: "inferred lambda return type String incompatible with expected return type Integer",
		},
		{
			name: "no context for inference - still requires explicit types",
			input: `
				var f := lambda(x) => x * 2;
			`,
			expectedErr: "not fully implemented",
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

			found := false
			for _, err := range a.Errors() {
				if strings.Contains(err, tt.expectedErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing '%s', got errors: %v", tt.expectedErr, a.Errors())
			}
		})
	}
}

// TestLambdaInferenceEdgeCases tests additional edge cases for lambda type inference.
func TestLambdaInferenceEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "lambda in array literal with inferred types",
			input: `
				type TFunc = function(n: Integer): Integer;
				type TFuncArray = array of TFunc;
				var funcs: TFuncArray := [
					lambda(x) => x * 2,
					lambda(x) => x + 1,
					lambda(x) => x - 1
				];
			`,
		},
		{
			name: "lambda with closure capturing outer variables",
			input: `
				type TFunc = function(x: Integer): Integer;
				var multiplier := 5;
				var f: TFunc := lambda(x) => x * multiplier;
			`,
		},
		{
			name: "chained function calls with lambda inference",
			input: `
				type TMapper = function(x: Integer): Integer;
				function Apply(f: TMapper; n: Integer): Integer;
				begin
					Result := f(n);
				end;
				function Chain(f: TMapper; n: Integer): Integer;
				begin
					Result := Apply(f, n);
				end;
				var result := Chain(lambda(x) => x * 2, 10);
			`,
		},
		{
			name: "lambda assigned to record field",
			input: `
				type TFunc = function(x: Integer): Integer;
				type TRecord = record
					operation: TFunc;
				end;
				var rec: TRecord;
				begin
					rec.operation := lambda(x) => x * 2;
				end;
			`,
		},
		{
			name: "lambda with same parameter name as outer variable",
			input: `
				type TFunc = function(x: Integer): Integer;
				var x := 10;
				var f: TFunc := lambda(x) => x * 2;  // Parameter 'x' shadows outer 'x'
			`,
		},
		{
			name: "multiple lambdas in single expression",
			input: `
				type TBinaryOp = function(a, b: Integer): Integer;
				function Combine(op1, op2: TBinaryOp; x, y: Integer): Integer;
				begin
					Result := op1(x, y) + op2(x, y);
				end;
				var result := Combine(
					lambda(a, b) => a + b,
					lambda(a, b) => a * b,
					3, 4
				);
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

// TestLambdaInferenceComplexErrors tests complex error scenarios for lambda type inference.
func TestLambdaInferenceComplexErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name: "lambda in array with wrong element type",
			input: `
				type TIntFunc = function(x: Integer): Integer;
				type TFuncArray = array of TIntFunc;
				var funcs: TFuncArray := [
					lambda(x) => IntToStr(x)  // Returns String, not Integer
				];
			`,
			expectedErr: "inferred lambda return type String incompatible with expected return type Integer",
		},
		{
			name: "nested lambda with wrong inner return type",
			input: `
				type TIntFunc = function(x: Integer): Integer;
				function MakeFunc(): TIntFunc;
				var inner: TIntFunc;
				begin
					inner := lambda(x) => IntToStr(x);  // Wrong return type
					Result := inner;
				end;
			`,
			expectedErr: "inferred lambda return type String incompatible with expected return type Integer",
		},
		{
			name: "lambda with too few parameters in call",
			input: `
				type TBinaryFunc = function(a, b: Integer): Integer;
				function Process(f: TBinaryFunc): Integer;
				begin
					Result := f(1, 2);
				end;
				var result := Process(lambda(x) => x * 2);  // Only 1 param, needs 2
			`,
			expectedErr: "lambda has 1 parameter but expected function type has 2 parameters",
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

			found := false
			for _, err := range a.Errors() {
				if strings.Contains(err, tt.expectedErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing '%s', got errors: %v", tt.expectedErr, a.Errors())
			}
		})
	}
}

// TestVariadicLambdaInference tests lambda type inference in variadic function contexts.
func TestVariadicLambdaInference(t *testing.T) {
	tests := []struct {
		checkFn  func(*testing.T, *Analyzer)
		name     string
		input    string
		expectOk bool
	}{
		{
			name: "simple variadic with lambda - array of const",
			input: `
				type TMapper = function(x: Integer): String;
				procedure ApplyAll(const funcs: array of const);
				begin
				end;

				begin
					ApplyAll([lambda(n) => IntToStr(n), lambda(m) => IntToStr(m*2)]);
				end;
			`,
			expectOk: false, // array of const accepts any type, so lambdas will be Variant
			checkFn:  nil,
		},
		{
			name: "variadic with typed lambda parameter",
			input: `
				type TIntMapper = function(x: Integer): Integer;
				procedure ApplyToAll(const funcs: array of TIntMapper);
				begin
				end;

				begin
					ApplyToAll([lambda(x) => x * 2, lambda(y) => y + 1]);
				end;
			`,
			expectOk: true,
			checkFn: func(t *testing.T, a *Analyzer) {
				// Both lambdas should have Integer parameter type inferred
				// This is verified by the absence of errors
			},
		},
		{
			name: "variadic Integer mapper - multiple lambdas",
			input: `
				procedure ProcessInts(const mappers: array of function(Integer): Integer);
				begin
				end;

				begin
					ProcessInts([
						lambda(x) => x * 2,
						lambda(y) => y + 10,
						lambda(z) => z - 5
					]);
				end;
			`,
			expectOk: true,
			checkFn: func(t *testing.T, a *Analyzer) {
				// All three lambdas should infer Integer parameter
			},
		},
		{
			name: "mixed fixed and variadic parameters",
			input: `
				type TTransform = function(String): String;
				procedure Transform(prefix: String; const transforms: array of TTransform);
				begin
				end;

				begin
					Transform('Hello', [
						lambda(s) => UpperCase(s),
						lambda(t) => LowerCase(t)
					]);
				end;
			`,
			expectOk: true,
			checkFn: func(t *testing.T, a *Analyzer) {
				// Both lambdas should infer String parameter
			},
		},
		{
			name: "variadic with return type inference",
			input: `
				procedure Collect(const getters: array of function(): Integer);
				begin
				end;

				var x: Integer := 42;
				begin
					Collect([
						lambda() => x,
						lambda() => x + 1
					]);
				end;
			`,
			expectOk: true,
			checkFn:  nil,
		},
		{
			name: "variadic lambda type mismatch error",
			input: `
				procedure ApplyToInts(const funcs: array of function(Integer): Integer);
				begin
				end;

				begin
					ApplyToInts([
						lambda(x: String) => Length(x)
					]);
				end;
			`,
			expectOk: false, // Explicit String parameter conflicts with expected Integer
			checkFn:  nil,
		},
		{
			name: "empty variadic array",
			input: `
				type TFunc = function(Integer): Integer;
				procedure Apply(const funcs: array of TFunc);
				begin
				end;

				begin
					Apply([]);
				end;
			`,
			expectOk: true,
			checkFn:  nil,
		},
		{
			name: "single lambda in variadic",
			input: `
				procedure MapStrings(const mappers: array of function(String): String);
				begin
				end;

				begin
					MapStrings([lambda(s) => UpperCase(s)]);
				end;
			`,
			expectOk: true,
			checkFn:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := analyzeSource(t, tt.input)

			hasErrors := len(a.Errors()) > 0 || err != nil

			if tt.expectOk && hasErrors {
				t.Errorf("expected no errors, got: %v (analyzer errors: %v)", err, a.Errors())
			} else if !tt.expectOk && !hasErrors {
				t.Error("expected errors but got none")
			}

			if tt.expectOk && tt.checkFn != nil {
				tt.checkFn(t, a)
			}
		})
	}
}

// ============================================================================
// Overload Detection with Lambda Arguments
// ============================================================================

func TestOverloadDetectionWithLambdas(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		errorMsg string
		expectOk bool
	}{
		{
			name: "overloaded function with lambda - should detect and error",
			input: `
function Process(x: Integer; f: function(Integer): Integer): Integer; overload;
begin
	Result := f(x);
end;

function Process(x: String; f: function(String): String): String; overload;
begin
	Result := f(x);
end;

begin
	// This should trigger task 9.21.5 detection
	Process(5, lambda(n) => n * 2);
end;
			`,
			expectOk: false,
			errorMsg: "lambda type inference not yet supported for overloaded function",
		},
		{
			name: "overloaded function with explicit lambda types - should work",
			input: `
function Process(x: Integer; f: function(Integer): Integer): Integer; overload;
begin
	Result := f(x);
end;

function Process(x: String; f: function(String): String): String; overload;
begin
	Result := f(x);
end;

begin
	// Explicit types - should work fine
	Process(5, lambda(n: Integer): Integer => n * 2);
end;
			`,
			expectOk: true,
			errorMsg: "",
		},
		{
			name: "non-overloaded function with lambda - should work with inference",
			input: `
function Apply(x: Integer; f: function(Integer): Integer): Integer;
begin
	Result := f(x);
end;

begin
	// No overloading - lambda inference works
	Apply(5, lambda(n) => n * 2);
end;
			`,
			expectOk: true,
			errorMsg: "",
		},
		{
			name: "overloaded function with multiple lambda arguments",
			input: `
function Combine(f: function(Integer): Integer; g: function(Integer): Integer): Integer; overload;
begin
	Result := f(1) + g(2);
end;

function Combine(f: function(String): String; g: function(String): String): String; overload;
begin
	Result := f('a') + g('b');
end;

begin
	// Multiple lambdas without types - should detect both
	Combine(lambda(x) => x * 2, lambda(y) => y + 1);
end;
			`,
			expectOk: false,
			errorMsg: "lambda type inference not yet supported for overloaded function",
		},
		{
			name: "overloaded function with mixed arguments (lambda + non-lambda)",
			input: `
function Process(x: Integer; f: function(Integer): Integer): Integer; overload;
begin
	Result := f(x);
end;

function Process(x: String; f: function(String): String): String; overload;
begin
	Result := f(x);
end;

begin
	// Lambda at position 1 (second argument)
	Process(42, lambda(n) => n * 2);
end;
			`,
			expectOk: false,
			errorMsg: "lambda type inference not yet supported for overloaded function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := analyzeSource(t, tt.input)

			hasErrors := len(a.Errors()) > 0 || err != nil

			if tt.expectOk && hasErrors {
				t.Errorf("expected no errors, got: %v (analyzer errors: %v)", err, a.Errors())
			} else if !tt.expectOk && !hasErrors {
				t.Error("expected errors but got none")
			}

			// Check for specific error message if provided
			if !tt.expectOk && tt.errorMsg != "" {
				found := false
				for _, errMsg := range a.Errors() {
					// Simple substring match using Go's strings package would be better,
					// but let's do a manual check for clarity
					if len(errMsg) >= len(tt.errorMsg) {
						for i := 0; i <= len(errMsg)-len(tt.errorMsg); i++ {
							if errMsg[i:i+len(tt.errorMsg)] == tt.errorMsg {
								found = true
								break
							}
						}
					}
					if found {
						break
					}
				}
				if !found {
					t.Errorf("expected error containing '%s', but didn't find it in errors: %v",
						tt.errorMsg, a.Errors())
				}
			}
		})
	}
}
