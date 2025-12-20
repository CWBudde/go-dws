package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func parseEnumDecl(t *testing.T, input string) *ast.EnumDecl {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parser errors: %v", errs)
	}
	if lexErrs := p.LexerErrors(); len(lexErrs) > 0 {
		t.Fatalf("lexer errors: %v", lexErrs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	enumDecl, ok := program.Statements[0].(*ast.EnumDecl)
	if !ok {
		t.Fatalf("statement is not *ast.EnumDecl, got %T", program.Statements[0])
	}

	return enumDecl
}

func TestVisitEnumDecl_ValueExpr(t *testing.T) {
	input := `type TEnumAlpha = (eAlpha = Ord('A'), eBeta, eGamma = 1+2);`
	enumDecl := parseEnumDecl(t, input)

	env := runtime.NewEnvironment()
	ctx := NewExecutionContext(env)
	typeSystem := interptypes.NewTypeSystem()
	refCountMgr := runtime.NewRefCountManager()
	eval := NewEvaluator(typeSystem, nil, nil, nil, nil, refCountMgr)

	result := eval.VisitEnumDecl(enumDecl, ctx)
	if isError(result) {
		t.Fatalf("VisitEnumDecl returned error: %v", result)
	}

	alphaVal, ok := env.Get("eAlpha")
	if !ok {
		t.Fatalf("eAlpha not defined in environment")
	}
	alphaEnum, ok := alphaVal.(*runtime.EnumValue)
	if !ok {
		t.Fatalf("eAlpha is not EnumValue, got %T", alphaVal)
	}
	if alphaEnum.OrdinalValue != int('A') {
		t.Fatalf("eAlpha ordinal = %d, want %d", alphaEnum.OrdinalValue, int('A'))
	}

	betaVal, ok := env.Get("eBeta")
	if !ok {
		t.Fatalf("eBeta not defined in environment")
	}
	betaEnum, ok := betaVal.(*runtime.EnumValue)
	if !ok {
		t.Fatalf("eBeta is not EnumValue, got %T", betaVal)
	}
	if betaEnum.OrdinalValue != int('A')+1 {
		t.Fatalf("eBeta ordinal = %d, want %d", betaEnum.OrdinalValue, int('A')+1)
	}

	gammaVal, ok := env.Get("eGamma")
	if !ok {
		t.Fatalf("eGamma not defined in environment")
	}
	gammaEnum, ok := gammaVal.(*runtime.EnumValue)
	if !ok {
		t.Fatalf("eGamma is not EnumValue, got %T", gammaVal)
	}
	if gammaEnum.OrdinalValue != 3 {
		t.Fatalf("eGamma ordinal = %d, want %d", gammaEnum.OrdinalValue, 3)
	}
}
