package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Test simple interface declaration
// ============================================================================

func TestSimpleInterfaceDeclaration(t *testing.T) {
	input := `
type IMyInterface = interface
	procedure A;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("Expected *ast.InterfaceDecl, got %T", program.Statements[0])
	}

	if stmt.Name.Value != "IMyInterface" {
		t.Errorf("Expected interface name 'IMyInterface', got '%s'", stmt.Name.Value)
	}

	if stmt.Parent != nil {
		t.Error("Expected no parent interface, got non-nil parent")
	}

	if len(stmt.Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(stmt.Methods))
	}

	if stmt.Methods[0].Name.Value != "A" {
		t.Errorf("Expected method name 'A', got '%s'", stmt.Methods[0].Name.Value)
	}

	if stmt.Methods[0].ReturnType != nil {
		t.Error("Procedure should have nil return type")
	}
}

// ============================================================================
// Test interface with multiple methods
// ============================================================================

func TestInterfaceMultipleMethods(t *testing.T) {
	input := `
type ICounter = interface
	procedure Increment;
	procedure Decrement;
	function GetValue: Integer;
	procedure SetValue(x: Integer);
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("Expected *ast.InterfaceDecl, got %T", program.Statements[0])
	}

	if len(stmt.Methods) != 4 {
		t.Fatalf("Expected 4 methods, got %d", len(stmt.Methods))
	}

	// Test procedure with no parameters
	if stmt.Methods[0].Name.Value != "Increment" {
		t.Errorf("Expected method name 'Increment', got '%s'", stmt.Methods[0].Name.Value)
	}
	if stmt.Methods[0].ReturnType != nil {
		t.Error("Increment should have nil return type")
	}
	if len(stmt.Methods[0].Parameters) != 0 {
		t.Errorf("Increment should have 0 parameters, got %d", len(stmt.Methods[0].Parameters))
	}

	// Test function with return type
	if stmt.Methods[2].Name.Value != "GetValue" {
		t.Errorf("Expected method name 'GetValue', got '%s'", stmt.Methods[2].Name.Value)
	}
	if stmt.Methods[2].ReturnType == nil {
		t.Error("GetValue should have return type")
	} else if stmt.Methods[2].ReturnType.String() != "Integer" {
		t.Errorf("Expected return type 'Integer', got '%s'", stmt.Methods[2].ReturnType.String())
	}

	// Test procedure with parameters
	if stmt.Methods[3].Name.Value != "SetValue" {
		t.Errorf("Expected method name 'SetValue', got '%s'", stmt.Methods[3].Name.Value)
	}
	if len(stmt.Methods[3].Parameters) != 1 {
		t.Errorf("SetValue should have 1 parameter, got %d", len(stmt.Methods[3].Parameters))
	} else {
		if stmt.Methods[3].Parameters[0].Name.Value != "x" {
			t.Errorf("Expected parameter name 'x', got '%s'", stmt.Methods[3].Parameters[0].Name.Value)
		}
	}
}

// ============================================================================
// Test interface inheritance
// ============================================================================

func TestInterfaceInheritance(t *testing.T) {
	input := `
type IDerived = interface(IBase)
	procedure B;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("Expected *ast.InterfaceDecl, got %T", program.Statements[0])
	}

	if stmt.Name.Value != "IDerived" {
		t.Errorf("Expected interface name 'IDerived', got '%s'", stmt.Name.Value)
	}

	if stmt.Parent == nil {
		t.Fatal("Expected parent interface, got nil")
	}

	if stmt.Parent.Value != "IBase" {
		t.Errorf("Expected parent name 'IBase', got '%s'", stmt.Parent.Value)
	}

	if len(stmt.Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(stmt.Methods))
	}

	if stmt.Methods[0].Name.Value != "B" {
		t.Errorf("Expected method name 'B', got '%s'", stmt.Methods[0].Name.Value)
	}
}

// ============================================================================
// Test external interfaces
// ============================================================================

func TestExternalInterface(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		externalName string
		isExternal   bool
	}{
		{
			name: "external without name",
			input: `
type IExternal = interface external
	procedure DoSomething;
end;
`,
			isExternal:   true,
			externalName: "",
		},
		{
			name: "external with name",
			input: `
type IExternal = interface external 'IDispatch'
	procedure DoSomething;
end;
`,
			isExternal:   true,
			externalName: "IDispatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.InterfaceDecl)
			if !ok {
				t.Fatalf("Expected *ast.InterfaceDecl, got %T", program.Statements[0])
			}

			if stmt.IsExternal != tt.isExternal {
				t.Errorf("Expected IsExternal=%v, got %v", tt.isExternal, stmt.IsExternal)
			}

			if stmt.ExternalName != tt.externalName {
				t.Errorf("Expected ExternalName='%s', got '%s'", tt.externalName, stmt.ExternalName)
			}
		})
	}
}

// ============================================================================
// Test class implementing interfaces
// ============================================================================

func TestClassImplementsInterface(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		className          string
		parentClass        string
		expectedInterfaces []string
		hasParent          bool
	}{
		{
			name: "class with single interface",
			input: `
type TTest = class(IMyInterface)
end;
`,
			className:          "TTest",
			parentClass:        "",
			expectedInterfaces: []string{"IMyInterface"},
			hasParent:          false,
		},
		{
			name: "class with multiple interfaces",
			input: `
type TTest = class(IInterface1, IInterface2)
end;
`,
			className:          "TTest",
			parentClass:        "",
			expectedInterfaces: []string{"IInterface1", "IInterface2"},
			hasParent:          false,
		},
		{
			name: "class with parent and interface",
			input: `
type TTest = class(TParent, IMyInterface)
end;
`,
			className:          "TTest",
			parentClass:        "TParent",
			expectedInterfaces: []string{"IMyInterface"},
			hasParent:          true,
		},
		{
			name: "class with parent and multiple interfaces",
			input: `
type TTest = class(TParent, IInterface1, IInterface2, IInterface3)
end;
`,
			className:          "TTest",
			parentClass:        "TParent",
			expectedInterfaces: []string{"IInterface1", "IInterface2", "IInterface3"},
			hasParent:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ClassDecl)
			if !ok {
				t.Fatalf("Expected *ast.ClassDecl, got %T", program.Statements[0])
			}

			if stmt.Name.Value != tt.className {
				t.Errorf("Expected class name '%s', got '%s'", tt.className, stmt.Name.Value)
			}

			// Check parent
			if tt.hasParent {
				if stmt.Parent == nil {
					t.Fatal("Expected parent class, got nil")
				}
				if stmt.Parent.Value != tt.parentClass {
					t.Errorf("Expected parent '%s', got '%s'", tt.parentClass, stmt.Parent.Value)
				}
			} else {
				if stmt.Parent != nil {
					t.Errorf("Expected no parent, got '%s'", stmt.Parent.Value)
				}
			}

			// Check interfaces
			if len(stmt.Interfaces) != len(tt.expectedInterfaces) {
				t.Fatalf("Expected %d interfaces, got %d", len(tt.expectedInterfaces), len(stmt.Interfaces))
			}

			for i, expected := range tt.expectedInterfaces {
				if stmt.Interfaces[i].Value != expected {
					t.Errorf("Expected interface[%d] '%s', got '%s'", i, expected, stmt.Interfaces[i].Value)
				}
			}
		})
	}
}

// ============================================================================
// Test forward interface declarations
// ============================================================================

func TestForwardInterfaceDeclaration(t *testing.T) {
	input := `
type IForward = interface;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("Expected *ast.InterfaceDecl, got %T", program.Statements[0])
	}

	if stmt.Name.Value != "IForward" {
		t.Errorf("Expected interface name 'IForward', got '%s'", stmt.Name.Value)
	}

	if len(stmt.Methods) != 0 {
		t.Errorf("Forward declaration should have 0 methods, got %d", len(stmt.Methods))
	}
}

// ============================================================================
// Test parsing errors
// ============================================================================

func TestInterfaceParsingErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "missing end keyword",
			input: `
type ITest = interface
	procedure A;
`,
		},
		{
			name: "missing semicolon after end",
			input: `
type ITest = interface
	procedure A;
end
`,
		},
		{
			name: "method with body (not allowed)",
			input: `
type ITest = interface
	procedure A begin end;
end;
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			if len(p.Errors()) == 0 {
				t.Error("Expected parser errors, got none")
			}
		})
	}
}

// Note: checkParserErrors is defined in parser_test.go
