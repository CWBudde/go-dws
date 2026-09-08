package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Resolved types must survive the compile/execute boundary without a runtime
// registry or a second parse of signatures with named parameters.
func TestResolveTypeFromAnnotation_SemanticIdentity(t *testing.T) {
	for _, source := range []string{
		"var value: array[-2..2] of Integer;",
		"type TItem = (First, Second); var value: array[TItem] of String;",
		"var value: function(var x: Integer): String;",
		"type TBase = class end; type TChild = class(TBase) end; var value: class of TChild;",
	} {
		t.Run(source, func(t *testing.T) {
			p := parser.New(lexer.New(source))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parse: %v", p.Errors())
			}
			analyzer := semantic.NewAnalyzer()
			if err := analyzer.Analyze(program); err != nil {
				t.Fatalf("analyze: %v", err)
			}
			decl, ok := program.Statements[len(program.Statements)-1].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("last statement: %T", program.Statements[len(program.Statements)-1])
			}
			info := analyzer.GetSemanticInfo()
			expected := info.GetResolvedType(decl.Type)
			if expected == nil {
				t.Fatal("missing resolved declaration type")
			}
			evaluator := NewEvaluator(nil, nil, nil, nil, info, nil)
			actual, err := evaluator.ResolveTypeFromAnnotation(decl.Type, nil)
			if err != nil {
				t.Fatal(err)
			}
			if actual != expected {
				t.Fatalf("runtime reconstructed %v instead of retaining semantic identity", actual)
			}
			if class, ok := actual.(*types.ClassOfType); ok && class.ClassType.Parent == nil {
				t.Fatal("lost parent class")
			}
		})
	}
}

func TestSemanticTypes_ContextualLiteral(t *testing.T) {
	for _, source := range []string{
		"type TValues = array of Integer; var value: TValues := [1, 2];",
		"type TItem = (First, Second); type TValues = set of TItem; var value: TValues := [];",
	} {
		t.Run(source, func(t *testing.T) {
			p := parser.New(lexer.New(source))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatalf("parse: %v", p.Errors())
			}
			analyzer := semantic.NewAnalyzer()
			if err := analyzer.Analyze(program); err != nil {
				t.Fatalf("analyze: %v", err)
			}
			decl := program.Statements[len(program.Statements)-1].(*ast.VarDeclStatement)
			info := analyzer.GetSemanticInfo()
			resolved := info.GetResolvedType(decl.Value)
			if resolved == nil {
				t.Fatal("original contextual literal has no resolved type")
			}
			annotation := info.GetType(decl.Value)
			if annotation == nil || info.GetResolvedType(annotation) != resolved {
				t.Fatal("contextual literal annotation lost resolved type identity")
			}
		})
	}
}
