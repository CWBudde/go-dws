package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// These tests observe the entry points themselves, so converting an original
// ErrorValue or replacing ctx.Exception with a new wrapper cannot go unnoticed.
func TestPropertyIndexRuntime_OriginalFailures(t *testing.T) {
	for _, path := range []string{"interface", "resolved"} {
		for _, kind := range []string{"receiver", "varIndex", "constIndex", "exception"} {
			t.Run(path+"/"+kind, func(t *testing.T) {
				e, ctx := createTestEvaluator(), createTestContext()
				original := &runtime.ErrorValue{Message: "original value [line: 7, column: 3]"}
				exception := &runtime.ExceptionValue{Message: "original exception"}
				receiver := &ast.Identifier{Value: "Receiver"}
				ctx.Env().Define("Receiver", &runtime.IntegerValue{Value: 1})
				ctx.Env().Define("Broken", original)
				ctx.Env().Define("X", &runtime.IntegerValue{Value: 5})
				prop := &types.PropertyInfo{Name: "P", IsIndexed: true, IndexParamNames: []string{"A", "B", "C"}, IndexParamModes: []types.PropertyIndexMode{types.PropertyIndexVar, types.PropertyIndexConst, types.PropertyIndexValue}}
				indices := []ast.Expression{&ast.Identifier{Value: "X"}, &ast.IntegerLiteral{Value: 1}, &ast.Identifier{Value: "MustNotRun"}}
				switch kind {
				case "receiver":
					ctx.Env().Define("Receiver", original)
				case "varIndex":
					indices[0] = &ast.IndexExpression{Left: &ast.Identifier{Value: "Broken"}, Index: &ast.IntegerLiteral{Value: 0}}
				case "constIndex":
					indices[1] = &ast.Identifier{Value: "Broken"}
				case "exception":
					// A real function re-raises its handler's original exception object.
					ctx.SetHandlerException(exception)
					e.typeSystem.RegisterFunction("Fail", &ast.FunctionDecl{Name: &ast.Identifier{Value: "Fail"}, ReturnType: &ast.TypeAnnotation{Name: "Integer"}, Body: &ast.BlockStatement{Statements: []ast.Statement{&ast.RaiseStatement{}}}})
					indices[1] = &ast.CallExpression{Function: &ast.Identifier{Value: "Fail"}}
				}
				var result Value
				if path == "interface" {
					iface := types.NewInterfaceType("I")
					iface.Properties["p"] = prop
					info := ast.NewSemanticInfo()
					info.SetResolvedType(receiver, iface)
					e.SetSemanticInfo(info)
					member := &ast.MemberAccessExpression{Object: receiver, Member: &ast.Identifier{Value: "P"}}
					var node ast.Expression = member
					for _, index := range indices {
						node = &ast.IndexExpression{Left: node, Index: index}
					}
					_, _, _, handled, err := e.resolveInterfaceIndexedProperty(node.(*ast.IndexExpression), ctx)
					if !handled {
						t.Fatal("interface contract not handled")
					}
					result = err
				} else {
					read := &ast.InheritedPropertyReadBinding{Read: &ast.MemberAccessExpression{Object: receiver, Member: &ast.Identifier{Value: "P"}}, Property: prop, Owner: "Lexical"}
					result = e.evalIndexedCompatibilityRead(&ast.IndexedPropertyReadBinding{Read: read, Indices: indices}, ctx)
				}
				if kind == "exception" {
					if ctx.Exception() != exception {
						t.Fatalf("exception identity lost: %#v", ctx.Exception())
					}
					if isError(result) {
						t.Fatalf("exception replaced by error: %v", result)
					}
				} else if result != original {
					t.Fatalf("ErrorValue identity lost: got %v, want original %v", result, original)
				}
			})
		}
	}
}
