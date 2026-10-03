package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestAnonymousCallable_RecursiveSignatures(t *testing.T) {
	for _, signature := range []string{
		"function: procedure",
		"function: function: procedure",
		"function: procedure of object",
		"function: Integer of object",
		"function(var callback: procedure(const value: Integer); lazy factory: function: String): procedure(var value: Integer)",
	} {
		for _, named := range []bool{false, true} {
			prefix, annotation := "", signature
			if named {
				prefix, annotation = "type TCallback = "+signature+";\n", "TCallback"
			}
			t.Run(prefix+signature, func(t *testing.T) {
				p := parser.New(lexer.New(prefix + "var callback: " + annotation + ";\nvar tail: Integer;"))
				program := p.ParseProgram()
				if len(p.Errors()) != 0 {
					t.Fatal(p.Errors())
				}
				a := NewAnalyzer()
				if err := a.Analyze(program); err != nil {
					t.Fatal(err)
				}
				var node *ast.FunctionPointerTypeNode
				if named {
					node = program.Statements[0].(*ast.TypeDeclaration).FunctionPointerType
				} else {
					node = program.Statements[0].(*ast.VarDeclStatement).Type.(*ast.FunctionPointerTypeNode)
				}
				symbol, ok := a.symbols.Resolve("callback")
				if !ok {
					t.Fatal("callback declaration lost")
				}
				checkCallableASTType(t, a, node, types.GetUnderlyingType(symbol.Type))
				if tail, ok := a.symbols.Resolve("tail"); !ok || tail.Type != types.INTEGER {
					t.Fatalf("following declaration = %v", tail)
				}
			})
		}
	}
}

// Check the source AST against its consumer-visible type metadata, including
// nested ownership and modes that string reconstruction previously discarded.
func checkCallableASTType(t *testing.T, a *Analyzer, node *ast.FunctionPointerTypeNode, resolved types.Type) {
	t.Helper()
	var signature *types.FunctionPointerType
	switch value := resolved.(type) {
	case *types.FunctionPointerType:
		if node.OfObject {
			t.Fatal("method ownership lost")
		}
		signature = value
	case *types.MethodPointerType:
		if !node.OfObject {
			t.Fatal("nested method ownership leaked to outer signature")
		}
		signature = &value.FunctionPointerType
	default:
		t.Fatalf("callable resolved to %T", resolved)
	}
	if metadata := a.semanticInfo.GetResolvedType(node); metadata == nil || !metadata.Equals(resolved) {
		t.Fatalf("callable AST metadata = %v, want %v", metadata, resolved)
	}
	if len(signature.Parameters) != len(node.Parameters) {
		t.Fatalf("parameter count = %d, want %d", len(signature.Parameters), len(node.Parameters))
	}
	for i, param := range node.Parameters {
		checkCallableParameterModes(t, signature, i, param)
		if child, ok := param.Type.(*ast.FunctionPointerTypeNode); ok {
			checkCallableASTType(t, a, child, signature.Parameters[i])
		}
	}
	if child, ok := node.ReturnType.(*ast.FunctionPointerTypeNode); ok {
		checkCallableASTType(t, a, child, signature.ReturnType)
	} else if node.ReturnType == nil && !signature.IsProcedure() {
		t.Fatal("procedure acquired a return type")
	}
}

func checkCallableParameterModes(t *testing.T, signature *types.FunctionPointerType, index int, param *ast.Parameter) {
	t.Helper()
	if len(signature.VarParams) <= index || len(signature.ConstParams) <= index || len(signature.LazyParams) <= index ||
		signature.VarParams[index] != param.ByRef || signature.ConstParams[index] != param.IsConst || signature.LazyParams[index] != param.IsLazy {
		t.Fatalf("parameter %d lost its passing mode", index)
	}
}

func TestAnonymousCallable_AssignmentResultRecovery(t *testing.T) {
	for _, destination := range []string{"Integer", "JSONVariant"} {
		p := parser.New(lexer.New("var source: function(x: Integer): procedure;\nvar target: " + destination + ";\ntarget := source;\nvar tail := 42;"))
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			t.Fatal(p.Errors())
		}
		a := NewAnalyzer()
		_ = a.Analyze(program)
		assignment := program.Statements[2].(*ast.AssignmentStatement)
		result, ok := a.semanticInfo.GetResolvedType(assignment.Value).(*types.FunctionPointerType)
		if !ok || !result.IsProcedure() || !a.semanticInfo.IsImplicitCall(assignment.Value) {
			t.Fatalf("%s: assignment lost one-call procedure result: %v", destination, result)
		}
		if tail, ok := a.symbols.Resolve("tail"); !ok || tail.Type != types.INTEGER {
			t.Fatalf("following declaration = %v", tail)
		}
	}
}

func TestAnonymousCallable_MemberSignatures(t *testing.T) {
	for _, owner := range []string{"class", "record"} {
		t.Run(owner, func(t *testing.T) {
			a := parseAndAnalyze(t, "type THolder = "+owner+`
 Factory: function: procedure;
 function Echo(callback: function: procedure): function: procedure;
 begin Result := @callback; end;
end;
var tail: Integer;`)
			if len(a.Errors()) != 0 {
				t.Fatal(a.Errors())
			}
		})
	}
}

func TestAnonymousCallable_RecordStaticCallOwnership(t *testing.T) {
	for _, supplier := range []string{"R.Stored", "R.Factory", "item.Stored", "item.Factory", "meta.Stored", "meta.Factory"} {
		t.Run(supplier, func(t *testing.T) {
			p := parser.New(lexer.New("type TProc = procedure; type TFactory = function: TProc;\n" +
				"type R = record class var Stored: TFactory; property Factory: TFactory read Stored write Stored; end;\n" +
				"var item: R; var meta := R; var target: JSONVariant;\n" +
				"target := " + supplier + ";"))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatal(p.Errors())
			}
			a := NewAnalyzer()
			_ = a.Analyze(program)
			assignment := program.Statements[len(program.Statements)-1].(*ast.AssignmentStatement)
			member := assignment.Value.(*ast.MemberAccessExpression)
			result, ok := types.GetUnderlyingType(a.semanticInfo.GetResolvedType(member)).(*types.FunctionPointerType)
			if !ok || !result.IsProcedure() || !a.semanticInfo.IsImplicitCall(member) {
				t.Fatalf("member did not commit one procedure result: %v", result)
			}
			record, isMeta := recordReceiverType(a.semanticInfo.GetResolvedType(member.Object))
			if record == nil || isMeta != (supplier != "item.Stored" && supplier != "item.Factory") {
				t.Fatalf("receiver ownership changed: record=%v meta=%v", record, isMeta)
			}
			if a.semanticInfo.IsImplicitCall(member.Object) {
				t.Fatal("record receiver unexpectedly marked as a call")
			}
		})
	}
}
