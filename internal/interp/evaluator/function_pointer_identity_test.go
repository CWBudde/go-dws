package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestFunctionPointerPreservesResolvedParameterIdentity(t *testing.T) {
	p := parser.New(lexer.New("function Read(values: array[-2..2] of Integer): Integer; begin Result := values[0]; end;"))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatal(p.Errors())
	}
	analyzer := semantic.NewAnalyzer()
	if err := analyzer.Analyze(program); err != nil {
		t.Fatal(err)
	}
	declaration, ok := program.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected function declaration, got %T", program.Statements[0])
	}
	info := analyzer.GetSemanticInfo()
	evaluator := NewEvaluator(nil, nil, nil, nil, info, nil)
	pointerType := evaluator.buildFunctionPointerType(declaration, nil)
	if pointerType == nil {
		t.Fatal("missing resolved pointer signature")
	}
	if pointerType.Parameters[0] != info.GetResolvedType(declaration.Parameters[0].Type) {
		t.Fatal("address-of pointer replaced the resolved array parameter type")
	}
	if pointerType.ReturnType != info.GetResolvedType(declaration.ReturnType) {
		t.Fatal("resolved return type identity lost")
	}
}
