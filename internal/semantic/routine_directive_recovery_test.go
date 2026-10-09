package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestAnalyze_RoutineRecoveryBindsQualifiedPrototype(t *testing.T) {
	source := "type T = class class procedure P(X: type Integer = 1); static; end;\nclass procedure T.P(X: Integer); export; begin end;"
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 1 || !errs[0].Stop || errs[0].Message != "BEGIN expected" {
		t.Fatalf("parser: %v", errs)
	}
	fn := program.Statements[1].(*ast.FunctionDecl)
	if !fn.BodyMissingBegin || fn.Body != nil || fn.ClassName == nil {
		t.Fatalf("qualified carrier: %+v", fn)
	}
	a := NewAnalyzer()
	a.SetCompileStopped(true)
	a.SetSymbolDictionaryDiagnostics(false)
	if err := a.Analyze(program); err != nil {
		t.Fatalf("prototype binding: %v", err)
	}
	class := a.getClassType("T")
	overloads := class.GetMethodOverloads("P")
	if len(overloads) != 1 || overloads[0].IsForwarded || class.ForwardedMethods["p"] {
		t.Fatalf("prototype not bound: %+v", class)
	}
	checkRoutineRecoveredPrototypeMetadata(t, overloads[0])
	if a.currentFunction != nil {
		t.Fatal("routine prefix context escaped its declaration")
	}
}

func checkRoutineRecoveredPrototypeMetadata(t *testing.T, method *types.MethodInfo) {
	t.Helper()
	if !method.IsStatic || len(method.Signature.StrictParams) != 1 || !method.Signature.StrictParams[0] {
		t.Fatalf("static/strict metadata lost: %+v", method)
	}
	if len(method.Signature.DefaultValues) != 1 {
		t.Fatalf("default lost: %+v", method.Signature)
	}
	value, ok := method.Signature.DefaultValues[0].(*ast.IntegerLiteral)
	if !ok || value.Value != 1 {
		t.Fatalf("default not retained: %v", method.Signature.DefaultValues)
	}
}
