package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Constructor/Destructor Tests
// ============================================================================

// TestConstructorDeclaration tests parsing constructor declarations in classes
func TestConstructorDeclaration(t *testing.T) {
	input := `
type TExample = class
public
	constructor Create(AValue: Integer);
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("Expected ClassDecl, got %T", program.Statements[0])
	}

	if len(classDecl.Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(classDecl.Methods))
	}

	constructor := classDecl.Methods[0]
	if !constructor.IsConstructor {
		t.Errorf("Expected IsConstructor to be true")
	}

	if constructor.Name.Value != "Create" {
		t.Errorf("Expected constructor name 'Create', got '%s'", constructor.Name.Value)
	}

	if constructor.Visibility != ast.VisibilityPublic {
		t.Errorf("Expected public visibility, got %v", constructor.Visibility)
	}
}

// TestConstructorImplementation tests parsing constructor implementations outside class
func TestConstructorImplementation(t *testing.T) {
	input := `
type TExample = class
	constructor Create(AValue: Integer);
end;

constructor TExample.Create(AValue: Integer);
begin
	FValue := AValue;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(program.Statements))
	}

	// Second statement should be the constructor implementation
	constructorImpl, ok := program.Statements[1].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl for constructor implementation, got %T", program.Statements[1])
	}

	if !constructorImpl.IsConstructor {
		t.Errorf("Expected IsConstructor to be true for implementation")
	}

	if constructorImpl.Name.Value != "Create" {
		t.Errorf("Expected constructor name 'Create', got '%s'", constructorImpl.Name.Value)
	}

	if constructorImpl.Body == nil {
		t.Errorf("Expected constructor body to be present")
	}

	if len(constructorImpl.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(constructorImpl.Parameters))
	}
}

// TestDestructorDeclaration tests parsing destructor declarations in classes
func TestDestructorDeclaration(t *testing.T) {
	input := `
type TExample = class
public
	destructor Destroy;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("Expected ClassDecl, got %T", program.Statements[0])
	}

	if len(classDecl.Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(classDecl.Methods))
	}

	destructor := classDecl.Methods[0]
	if !destructor.IsDestructor {
		t.Errorf("Expected IsDestructor to be true")
	}

	if destructor.Name.Value != "Destroy" {
		t.Errorf("Expected destructor name 'Destroy', got '%s'", destructor.Name.Value)
	}
}

// TestDestructorImplementation tests parsing destructor implementations outside class
func TestDestructorImplementation(t *testing.T) {
	input := `
type TExample = class
	destructor Destroy;
end;

destructor TExample.Destroy;
begin
	// Cleanup code here
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %d", len(program.Statements))
	}

	// Second statement should be the destructor implementation
	destructorImpl, ok := program.Statements[1].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl for destructor implementation, got %T", program.Statements[1])
	}

	if !destructorImpl.IsDestructor {
		t.Errorf("Expected IsDestructor to be true for implementation")
	}

	if destructorImpl.Name.Value != "Destroy" {
		t.Errorf("Expected destructor name 'Destroy', got '%s'", destructorImpl.Name.Value)
	}

	if destructorImpl.Body == nil {
		t.Errorf("Expected destructor body to be present")
	}
}

// TestMultipleConstructors tests parsing a class with multiple constructors (overloaded)
func TestMultipleConstructors(t *testing.T) {
	input := `
type TExample = class
	constructor Create;
	constructor CreateWithValue(AValue: Integer);
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("Expected ClassDecl, got %T", program.Statements[0])
	}

	if len(classDecl.Methods) != 2 {
		t.Fatalf("Expected 2 methods, got %d", len(classDecl.Methods))
	}

	// Check both are marked as constructors
	for i, method := range classDecl.Methods {
		if !method.IsConstructor {
			t.Errorf("Method %d: Expected IsConstructor to be true", i)
		}
	}

	// Check names
	if classDecl.Methods[0].Name.Value != "Create" {
		t.Errorf("Expected first constructor name 'Create', got '%s'", classDecl.Methods[0].Name.Value)
	}

	if classDecl.Methods[1].Name.Value != "CreateWithValue" {
		t.Errorf("Expected second constructor name 'CreateWithValue', got '%s'", classDecl.Methods[1].Name.Value)
	}
}

// TestConstructorWithVisibility tests constructors with different visibility levels
func TestConstructorWithVisibility(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		visibility ast.Visibility
	}{
		{
			name: "public constructor",
			input: `
type TExample = class
public
	constructor Create;
end;`,
			visibility: ast.VisibilityPublic,
		},
		{
			name: "private constructor",
			input: `
type TExample = class
private
	constructor Create;
end;`,
			visibility: ast.VisibilityPrivate,
		},
		{
			name: "protected constructor",
			input: `
type TExample = class
protected
	constructor Create;
end;`,
			visibility: ast.VisibilityProtected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			classDecl := program.Statements[0].(*ast.ClassDecl)
			constructor := classDecl.Methods[0]

			if constructor.Visibility != tt.visibility {
				t.Errorf("Expected visibility %v, got %v", tt.visibility, constructor.Visibility)
			}
		})
	}
}

// TestConstructorDestructorMixed tests a class with both constructor and destructor
func TestConstructorDestructorMixed(t *testing.T) {
	input := `
type TExample = class
public
	constructor Create(AValue: Integer);
	destructor Destroy;
	function GetValue: Integer;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	classDecl := program.Statements[0].(*ast.ClassDecl)

	if len(classDecl.Methods) != 3 {
		t.Fatalf("Expected 3 methods, got %d", len(classDecl.Methods))
	}

	// Verify constructor
	if !classDecl.Methods[0].IsConstructor {
		t.Errorf("Expected first method to be constructor")
	}

	// Verify destructor
	if !classDecl.Methods[1].IsDestructor {
		t.Errorf("Expected second method to be destructor")
	}

	// Verify regular method
	if classDecl.Methods[2].IsConstructor || classDecl.Methods[2].IsDestructor {
		t.Errorf("Expected third method to be regular method")
	}
}
