package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// External Method Parsing Tests
// ============================================================================

func TestExternalMethodParsing(t *testing.T) {
	t.Run("external method without name", func(t *testing.T) {
		input := `
type TExternal = class external
  procedure Hello; external;
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ClassDecl)
		if len(stmt.Methods) != 1 {
			t.Fatalf("Expected 1 method, got %d", len(stmt.Methods))
		}

		method := stmt.Methods[0]
		if !method.IsExternal {
			t.Error("Expected IsExternal to be true")
		}

		if method.ExternalName != "" {
			t.Errorf("Expected empty ExternalName, got %q", method.ExternalName)
		}
	})

	t.Run("external method with name", func(t *testing.T) {
		input := `
type TExternal = class external
  procedure Hello; external 'world';
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ClassDecl)
		method := stmt.Methods[0]

		if !method.IsExternal {
			t.Error("Expected IsExternal to be true")
		}

		if method.ExternalName != "world" {
			t.Errorf("Expected ExternalName 'world', got %q", method.ExternalName)
		}
	})

	t.Run("external function with name", func(t *testing.T) {
		input := `
type TExternal = class external
  function GetValue: Integer; external 'getValue';
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ClassDecl)
		method := stmt.Methods[0]

		if !method.IsExternal {
			t.Error("Expected IsExternal to be true")
		}

		if method.ExternalName != "getValue" {
			t.Errorf("Expected ExternalName 'getValue', got %q", method.ExternalName)
		}

		if method.ReturnType == nil || method.ReturnType.String() != "Integer" {
			t.Error("Expected return type Integer")
		}
	})

	t.Run("regular method is not external", func(t *testing.T) {
		input := `
type TRegular = class
  procedure DoSomething;
  begin
  end;
end;
`
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ClassDecl)
		method := stmt.Methods[0]

		if method.IsExternal {
			t.Error("Regular method should not be external")
		}

		if method.ExternalName != "" {
			t.Errorf("Regular method should have empty ExternalName, got %q", method.ExternalName)
		}
	})
}
