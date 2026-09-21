package semantic

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

func TestInterfaceSelfReference_CanonicalType(t *testing.T) {
	const source = `type INode = interface
   function Next: inode;
   procedure Link(other: INODE);
   property Child: INode read Next;
end;
type IChild = interface(INode)
   function Copy: IChild;
end;`
	p := parser.New(lexer.New(source))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) != 0 {
		t.Fatalf("parse errors: %v", errors)
	}
	a := NewAnalyzer()
	if err := a.Analyze(program); err != nil {
		t.Fatalf("analysis failed: %v", a.Errors())
	}
	node := a.getInterfaceType("INode")
	if node == nil {
		t.Fatal("interface was not registered")
	}
	if node.Methods["next"].ReturnType != node {
		t.Error("self return type does not reference the canonical interface")
	}
	if node.Methods["link"].Parameters[0] != node {
		t.Error("self parameter type does not reference the canonical interface")
	}
	if node.Properties["child"].Type != node {
		t.Error("self property type does not reference the canonical interface")
	}
	child := a.getInterfaceType("IChild")
	if child == nil || child.Parent != node || child.Methods["copy"].ReturnType != child {
		t.Error("derived interface lost its canonical parent or self type")
	}
}

func TestInterfaceSelfReference_InvalidNames(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		want   string
	}{
		{"unknown return", `type INode = interface function Next: IMissing; end;`, "unknown return type 'IMissing'"},
		{"unknown parameter", `type INode = interface procedure Link(other: IMissing); end;`, "unknown parameter type 'IMissing'"},
		{"unknown property", `type INode = interface property Child: IMissing read Next; end;`, "unknown type 'IMissing'"},
		{"self parent", `type INode = interface(INode) end;`, "parent interface 'INode' not found"},
		{"future interface", `type INode = interface function Next: IFuture; end; type IFuture = interface end;`, "unknown return type 'IFuture'"},
		{"duplicate declaration", `type INode = interface function Next: INode; end; type inode = interface end;`, "already exists"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expectError(t, tc.source, tc.want)
		})
	}
}
