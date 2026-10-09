package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestRecordIndexDeclarations_IncompleteInlineReachedTypes(t *testing.T) {
	source := `var r: record
 function Get(x: Integer): Integer; begin Result := 0; end;
 property P[const a: Integer]: Integer read Get;
 property Q[var b: Missing : Integer read (1);
end;
Tail; {$ERROR 'unread'}`
	node := parseStoppedInlineRecord(t, source, 2)
	if !node.Properties[0].IndexParams[0].IsConst || !node.Properties[1].IndexParams[0].ByRef {
		t.Fatal("lost reached parameter modes")
	}
	a := NewAnalyzer()
	a.SetCompileStopped(true)
	resolved, err := a.resolveRecordTypeNode(node)
	if err != nil || resolved != nil {
		t.Fatalf("incomplete record resolved as a completed type: %v, %v", resolved, err)
	}
	info := a.GetSemanticInfo()
	if info.GetResolvedType(node) != nil {
		t.Fatal("registered incomplete record type")
	}
	if info.GetResolvedType(node.Properties[0].Type) != types.INTEGER {
		t.Fatal("completed property type was not reached")
	}
	if info.GetResolvedType(node.Properties[1].IndexParams[0].Type) != types.VARIANT {
		t.Fatal("lost unknown named index Variant recovery")
	}
	assertUnreadRecordPropertyAccessors(t, node.Properties[1])
}

func TestRecordIndexDeclarations_IncompleteInlinePropertyTypeError(t *testing.T) {
	source := `var r: record
 function Get(x: String): Integer; begin Result := 0; end;
 property P[var a: Integer]: Integer read Get;
 property Q[const a: Integer]: MissingType read Get;
 property U[var a: MissingIndex : Integer read (1);
end;
Tail; {$ERROR 'unread'}`
	node := parseStoppedInlineRecord(t, source, 3)
	if !node.Properties[1].IndexParams[0].IsConst || !node.Properties[2].IndexParams[0].ByRef {
		t.Fatal("lost reached parameter modes")
	}
	a := NewAnalyzer()
	a.SetCompileStopped(true)
	resolved, err := a.resolveRecordTypeNode(node)
	if err != nil || resolved != nil {
		t.Fatalf("property failure escaped incomplete resolution: %v, %v", resolved, err)
	}
	info := a.GetSemanticInfo()
	if info.GetResolvedType(node) != nil || info.GetResolvedType(node.Properties[1].Type) != nil {
		t.Fatal("registered incomplete record or invalid property type")
	}
	if info.GetResolvedType(node.Properties[0].Type) != types.INTEGER ||
		info.GetResolvedType(node.Properties[1].IndexParams[0].Type) != types.INTEGER ||
		info.GetResolvedType(node.Properties[2].IndexParams[0].Type) != types.VARIANT {
		t.Fatal("lost reached property/index resolution")
	}
	assertUnreadRecordPropertyAccessors(t, node.Properties[2])
}

func parseStoppedInlineRecord(t *testing.T, source string, propertyCount int) *ast.RecordTypeNode {
	t.Helper()
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	errs := p.Errors()
	if len(errs) != 1 || !errs[0].Stop || errs[0].Message != `"]" expected` {
		t.Fatal(errs)
	}
	if len(program.Statements) != 1 {
		t.Fatal("stop read later statements")
	}
	node := program.Statements[0].(*ast.VarDeclStatement).Type.(*ast.RecordTypeNode)
	if !node.Incomplete || len(node.Properties) != propertyCount || node.Properties[propertyCount-1].Type != nil {
		t.Fatal("lost reached partial inline metadata")
	}
	return node
}

func assertUnreadRecordPropertyAccessors(t *testing.T, partial ast.RecordPropertyDecl) {
	t.Helper()
	if partial.ReadExpr != nil || partial.ReadField != "" {
		t.Fatal("partial property acquired unread accessors")
	}
}
