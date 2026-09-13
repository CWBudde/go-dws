package types

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestTypeSystem_LexicalUnitOwnership(t *testing.T) {
	p := parser.New(lexer.New(`function Outer: Integer;
 function Nested: Integer; begin Result := 1 end;
 begin var callback := lambda => Nested(); Result := callback() end;`))
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatal(p.Errors())
	}
	ts := NewTypeSystem()
	ts.RegisterNodeUnit(program, "MyUNIT")
	count := 0
	ast.Inspect(program, func(node ast.Node) bool {
		if node != nil {
			count++
			if owner, ok := ts.NodeUnit(node); !ok || owner != "myunit" {
				t.Fatalf("owner of %T = %q, %v", node, owner, ok)
			}
		}
		return true
	})
	if count < 10 {
		t.Fatalf("unexpectedly short source tree: %d", count)
	}
	mainNode := &ast.Identifier{Value: "MainCallback"}
	ts.RegisterNodeUnit(mainNode, "")
	if owner, ok := ts.NodeUnit(mainNode); !ok || owner != "" {
		t.Fatalf("explicit main owner = %q, %v", owner, ok)
	}
	if _, ok := ts.NodeUnit(&ast.Identifier{Value: "Generated"}); ok {
		t.Fatal("generated node was registered")
	}
}

func TestTypeSystem_GeneratedWrapperOwnership(t *testing.T) {
	ts := NewTypeSystem()
	source := &ast.Identifier{Value: "Implementation"}
	inherited := &ast.Identifier{Value: "InheritedContract"}
	ts.RegisterNodeUnit(source, "UnitB")
	ts.RegisterNodeUnit(inherited, "UnitA")
	wrapper := &ast.BinaryExpression{Operator: "and", Left: source, Right: inherited}
	ts.RegisterNodeAlias(wrapper, source)
	if owner, ok := ts.NodeUnit(wrapper); !ok || owner != "unitb" {
		t.Fatalf("wrapper owner = %q, %v", owner, ok)
	}
	if owner, ok := ts.NodeUnit(inherited); !ok || owner != "unita" {
		t.Fatalf("inherited expression owner = %q, %v", owner, ok)
	}
}
