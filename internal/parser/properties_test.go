package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// Basic Property Parsing Tests
// ============================================================================

func TestPropertyBasicReadWrite(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		propName    string
		propType    string
		readSpec    string
		writeSpec   string
		isReadOnly  bool
		isWriteOnly bool
	}{
		{
			name:        "field-backed property",
			input:       "property Name: String read FName write FName;",
			propName:    "Name",
			propType:    "String",
			readSpec:    "FName",
			writeSpec:   "FName",
			isReadOnly:  false,
			isWriteOnly: false,
		},
		{
			name:        "method-backed property",
			input:       "property Count: Integer read GetCount write SetCount;",
			propName:    "Count",
			propType:    "Integer",
			readSpec:    "GetCount",
			writeSpec:   "SetCount",
			isReadOnly:  false,
			isWriteOnly: false,
		},
		{
			name:        "mixed field and method",
			input:       "property Value: Float read FValue write SetValue;",
			propName:    "Value",
			propType:    "Float",
			readSpec:    "FValue",
			writeSpec:   "SetValue",
			isReadOnly:  false,
			isWriteOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)

			// Parse the property declaration
			prop := p.parsePropertyDeclaration()

			// Check for parser errors
			errors := p.Errors()
			if len(errors) != 0 {
				t.Fatalf("Parser had %d errors:\n%v", len(errors), errors)
			}

			if prop == nil {
				t.Fatal("parsePropertyDeclaration() returned nil")
			}

			// Verify property name
			if prop.Name == nil {
				t.Fatal("Property Name is nil")
			}
			if prop.Name.Value != tt.propName {
				t.Errorf("Property name: expected=%q, got=%q", tt.propName, prop.Name.Value)
			}

			// Verify property type
			if prop.Type == nil {
				t.Fatal("Property Type is nil")
			}
			if prop.Type.String() != tt.propType {
				t.Errorf("Property type: expected=%q, got=%q", tt.propType, prop.Type.String())
			}

			// Verify read spec
			if !tt.isWriteOnly {
				if prop.ReadSpec == nil {
					t.Fatal("Property ReadSpec is nil")
				}
				readIdent, ok := prop.ReadSpec.(*ast.Identifier)
				if !ok {
					t.Fatalf("ReadSpec is not an Identifier, got=%T", prop.ReadSpec)
				}
				if readIdent.Value != tt.readSpec {
					t.Errorf("ReadSpec: expected=%q, got=%q", tt.readSpec, readIdent.Value)
				}
			}

			// Verify write spec
			if !tt.isReadOnly {
				if prop.WriteSpec == nil {
					t.Fatal("Property WriteSpec is nil")
				}
				writeIdent, ok := prop.WriteSpec.(*ast.Identifier)
				if !ok {
					t.Fatalf("WriteSpec is not an Identifier, got=%T", prop.WriteSpec)
				}
				if writeIdent.Value != tt.writeSpec {
					t.Errorf("WriteSpec: expected=%q, got=%q", tt.writeSpec, writeIdent.Value)
				}
			}

			// Verify indexed params and default flag
			if prop.IndexParams != nil {
				t.Errorf("IndexParams should be nil for non-indexed property, got=%v", prop.IndexParams)
			}
			if prop.IsDefault {
				t.Error("IsDefault should be false for non-default property")
			}
		})
	}
}

func TestPropertyReadOnly(t *testing.T) {
	input := "property Size: Integer read FSize;"

	l := lexer.New(input)
	p := New(l)
	prop := p.parsePropertyDeclaration()

	errors := p.Errors()
	if len(errors) != 0 {
		t.Fatalf("Parser had %d errors:\n%v", len(errors), errors)
	}

	if prop == nil {
		t.Fatal("parsePropertyDeclaration() returned nil")
	}

	// Verify read spec exists
	if prop.ReadSpec == nil {
		t.Fatal("Read-only property should have ReadSpec")
	}

	readIdent, ok := prop.ReadSpec.(*ast.Identifier)
	if !ok {
		t.Fatalf("ReadSpec is not an Identifier, got=%T", prop.ReadSpec)
	}
	if readIdent.Value != "FSize" {
		t.Errorf("ReadSpec: expected=%q, got=%q", "FSize", readIdent.Value)
	}

	// Verify write spec is nil
	if prop.WriteSpec != nil {
		t.Errorf("Read-only property should not have WriteSpec, got=%v", prop.WriteSpec)
	}
}

func TestPropertyWriteOnly(t *testing.T) {
	input := "property Output: String write SetOutput;"

	l := lexer.New(input)
	p := New(l)
	prop := p.parsePropertyDeclaration()

	errors := p.Errors()
	if len(errors) != 0 {
		t.Fatalf("Parser had %d errors:\n%v", len(errors), errors)
	}

	if prop == nil {
		t.Fatal("parsePropertyDeclaration() returned nil")
	}

	// Verify read spec is nil
	if prop.ReadSpec != nil {
		t.Errorf("Write-only property should not have ReadSpec, got=%v", prop.ReadSpec)
	}

	// Verify write spec exists
	if prop.WriteSpec == nil {
		t.Fatal("Write-only property should have WriteSpec")
	}

	writeIdent, ok := prop.WriteSpec.(*ast.Identifier)
	if !ok {
		t.Fatalf("WriteSpec is not an Identifier, got=%T", prop.WriteSpec)
	}
	if writeIdent.Value != "SetOutput" {
		t.Errorf("WriteSpec: expected=%q, got=%q", "SetOutput", writeIdent.Value)
	}
}

func TestPropertyExpressionRead(t *testing.T) {
	// Test expression-based read specifier: property Doubled: Integer read (FValue * 2);
	input := "property Doubled: Integer read (FValue * 2);"

	l := lexer.New(input)
	p := New(l)
	prop := p.parsePropertyDeclaration()

	errors := p.Errors()
	if len(errors) != 0 {
		t.Fatalf("Parser had %d errors:\n%v", len(errors), errors)
	}

	if prop == nil {
		t.Fatal("parsePropertyDeclaration() returned nil")
	}

	// Verify read spec is an expression (not just an identifier)
	if prop.ReadSpec == nil {
		t.Fatal("Property ReadSpec is nil")
	}

	// Should be a BinaryExpression (FValue * 2)
	_, isBinary := prop.ReadSpec.(*ast.BinaryExpression)
	if !isBinary {
		t.Errorf("ReadSpec should be BinaryExpression for (FValue * 2), got=%T", prop.ReadSpec)
	}
}

// ============================================================================
// Indexed, Default, and Auto Properties
// ============================================================================

func TestPropertyIndexed(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		propName   string
		indexNames []string
		indexTypes []string
		indexCount int
	}{
		{
			name:       "single index parameter",
			input:      "property Items[index: Integer]: String read GetItem write SetItem;",
			propName:   "Items",
			indexCount: 1,
			indexNames: []string{"index"},
			indexTypes: []string{"Integer"},
		},
		{
			name:       "multiple index parameters",
			input:      "property Data[x, y: Integer]: Float read GetData write SetData;",
			propName:   "Data",
			indexCount: 2,
			indexNames: []string{"x", "y"},
			indexTypes: []string{"Integer", "Integer"},
		},
		{
			name:       "mixed index parameter types",
			input:      "property Matrix[row: Integer; col: Integer]: Float read GetCell;",
			propName:   "Matrix",
			indexCount: 2,
			indexNames: []string{"row", "col"},
			indexTypes: []string{"Integer", "Integer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			prop := p.parsePropertyDeclaration()

			errors := p.Errors()
			if len(errors) != 0 {
				t.Fatalf("Parser had %d errors:\n%v", len(errors), errors)
			}

			if prop == nil {
				t.Fatal("parsePropertyDeclaration() returned nil")
			}

			// Verify property name
			if prop.Name.Value != tt.propName {
				t.Errorf("Property name: expected=%q, got=%q", tt.propName, prop.Name.Value)
			}

			// Verify index parameters
			if prop.IndexParams == nil {
				t.Fatal("Indexed property should have IndexParams")
			}

			if len(prop.IndexParams) != tt.indexCount {
				t.Fatalf("IndexParams count: expected=%d, got=%d", tt.indexCount, len(prop.IndexParams))
			}

			// Verify each parameter
			for i, param := range prop.IndexParams {
				if param.Name.Value != tt.indexNames[i] {
					t.Errorf("Index param %d name: expected=%q, got=%q", i, tt.indexNames[i], param.Name.Value)
				}
				if param.Type.String() != tt.indexTypes[i] {
					t.Errorf("Index param %d type: expected=%q, got=%q", i, tt.indexTypes[i], param.Type.String())
				}
			}
		})
	}
}

func TestPropertyDefault(t *testing.T) {
	input := "property Items[index: Integer]: String read GetItem write SetItem; default;"

	l := lexer.New(input)
	p := New(l)
	prop := p.parsePropertyDeclaration()

	errors := p.Errors()
	if len(errors) != 0 {
		t.Fatalf("Parser had %d errors:\n%v", len(errors), errors)
	}

	if prop == nil {
		t.Fatal("parsePropertyDeclaration() returned nil")
	}

	// Verify IsDefault flag
	if !prop.IsDefault {
		t.Error("Expected IsDefault=true for default property")
	}

	// Default properties must be indexed
	if prop.IndexParams == nil || len(prop.IndexParams) == 0 {
		t.Error("Default property must have index parameters")
	}
}

func TestPropertyAuto(t *testing.T) {
	input := "property Name: String;"

	l := lexer.New(input)
	p := New(l)
	prop := p.parsePropertyDeclaration()

	errors := p.Errors()
	if len(errors) != 0 {
		t.Fatalf("Parser had %d errors:\n%v", len(errors), errors)
	}

	if prop == nil {
		t.Fatal("parsePropertyDeclaration() returned nil")
	}

	// Auto-property should have generated read/write specs
	// The backing field name should be FName (F + property name)
	if prop.ReadSpec == nil {
		t.Error("Auto-property should have generated ReadSpec")
	}

	if prop.WriteSpec == nil {
		t.Error("Auto-property should have generated WriteSpec")
	}

	// Both should point to the same generated field
	readIdent, ok1 := prop.ReadSpec.(*ast.Identifier)
	writeIdent, ok2 := prop.WriteSpec.(*ast.Identifier)

	if !ok1 || !ok2 {
		t.Fatal("Auto-property read/write specs should be identifiers")
	}

	expectedField := "FName" // F + property name
	if readIdent.Value != expectedField {
		t.Errorf("Auto-property ReadSpec: expected=%q, got=%q", expectedField, readIdent.Value)
	}
	if writeIdent.Value != expectedField {
		t.Errorf("Auto-property WriteSpec: expected=%q, got=%q", expectedField, writeIdent.Value)
	}
}

func TestPropertyErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "missing type",
			input:         "property Name read FName;",
			expectedError: "expected",
		},
		{
			name:          "missing read and write",
			input:         "property Name: String;", // This should be auto-property, tested later
			expectedError: "",                       // Will handle in auto-property test
		},
		{
			name:          "missing semicolon",
			input:         "property Name: String read FName",
			expectedError: "expected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip auto-property test for now (tested in Batch 2)
			if tt.name == "missing read and write" {
				t.Skip("Auto-property test - will be implemented in Batch 2")
			}

			l := lexer.New(tt.input)
			p := New(l)
			prop := p.parsePropertyDeclaration()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Fatalf("Expected parser error for %q, but got none. Prop=%v", tt.input, prop)
			}
		})
	}
}

// ============================================================================
// Class Integration Tests
// ============================================================================

func TestClassWithProperties(t *testing.T) {
	input := `
type TPerson = class
	private
		FName: String;
		FAge: Integer;
	public
		property Name: String read FName write FName;
		property Age: Integer read FAge write FAge;
		property IsAdult: Boolean read (FAge >= 18);
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	// Verify class has properties
	if classDecl.Properties == nil {
		t.Fatal("Class should have Properties field")
	}

	if len(classDecl.Properties) != 3 {
		t.Fatalf("Expected 3 properties, got %d", len(classDecl.Properties))
	}

	// Verify first property: property Name: String read FName write FName;
	prop1 := classDecl.Properties[0]
	if prop1.Name.Value != "Name" {
		t.Errorf("Property 1 name: expected='Name', got=%q", prop1.Name.Value)
	}
	if prop1.Type.String() != "String" {
		t.Errorf("Property 1 type: expected='String', got=%q", prop1.Type.String())
	}

	// Verify second property
	prop2 := classDecl.Properties[1]
	if prop2.Name.Value != "Age" {
		t.Errorf("Property 2 name: expected='Age', got=%q", prop2.Name.Value)
	}

	// Verify third property (expression-based read)
	prop3 := classDecl.Properties[2]
	if prop3.Name.Value != "IsAdult" {
		t.Errorf("Property 3 name: expected='IsAdult', got=%q", prop3.Name.Value)
	}
	if prop3.WriteSpec != nil {
		t.Error("IsAdult property should be read-only (no WriteSpec)")
	}
}

func TestPropertyAccessInExpression(t *testing.T) {
	// Property access is parsed as member access
	// This is already handled by existing expression parser
	input := "obj.Name"

	l := lexer.New(input)
	p := New(l)

	expr := p.parseExpressionCursor(LOWEST)

	errors := p.Errors()
	if len(errors) != 0 {
		t.Fatalf("Parser had %d errors:\n%v", len(errors), errors)
	}

	// Should be parsed as MemberAccessExpression
	memberExpr, ok := expr.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("Expected MemberAccessExpression, got=%T", expr)
	}

	if memberExpr.Member.Value != "Name" {
		t.Errorf("Member: expected='Name', got=%q", memberExpr.Member.Value)
	}

	// Property translation happens in semantic analysis/interpreter, not parser
}

func TestPropertyAssignment(t *testing.T) {
	// Property assignment is parsed as assignment to member
	// This is already handled by existing statement parser
	input := "obj.Name := 'John';"

	l := lexer.New(input)
	p := New(l)

	stmt := p.parseStatementCursor()

	errors := p.Errors()
	if len(errors) != 0 {
		t.Fatalf("Parser had %d errors:\n%v", len(errors), errors)
	}

	// Should be parsed as AssignmentStatement
	assignStmt, ok := stmt.(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("Expected AssignmentStatement, got=%T", stmt)
	}

	// Target should be MemberAccessExpression
	memberExpr, ok := assignStmt.Target.(*ast.MemberAccessExpression)
	if !ok {
		t.Fatalf("Assignment Target should be MemberAccessExpression, got=%T", assignStmt.Target)
	}

	if memberExpr.Member.Value != "Name" {
		t.Errorf("Member: expected='Name', got=%q", memberExpr.Member.Value)
	}

	// Property translation happens in semantic analysis/interpreter, not parser
}

// ============================================================================
// Class Property Tests
// ============================================================================

func TestClassProperty(t *testing.T) {
	input := `
type TCounter = class
	private
		class var FCount: Integer;
	public
		class property Count: Integer read GetCount write SetCount;
		class property Version: String read GetVersion;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	classDecl, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ClassDecl. got=%T",
			program.Statements[0])
	}

	// Verify class has 2 properties
	if len(classDecl.Properties) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(classDecl.Properties))
	}

	// Verify first property: class property Count
	prop1 := classDecl.Properties[0]
	if prop1.Name.Value != "Count" {
		t.Errorf("Property 1 name: expected='Count', got=%q", prop1.Name.Value)
	}
	if !prop1.IsClassProperty {
		t.Error("Property 1 should be marked as class property")
	}
	if prop1.Type.String() != "Integer" {
		t.Errorf("Property 1 type: expected='Integer', got=%q", prop1.Type.String())
	}

	// Verify it has both read and write
	if prop1.ReadSpec == nil {
		t.Error("Property 1 should have ReadSpec")
	}
	if prop1.WriteSpec == nil {
		t.Error("Property 1 should have WriteSpec")
	}

	// Verify second property: class property Version (read-only)
	prop2 := classDecl.Properties[1]
	if prop2.Name.Value != "Version" {
		t.Errorf("Property 2 name: expected='Version', got=%q", prop2.Name.Value)
	}
	if !prop2.IsClassProperty {
		t.Error("Property 2 should be marked as class property")
	}
	if prop2.ReadSpec == nil {
		t.Error("Property 2 should have ReadSpec")
	}
	if prop2.WriteSpec != nil {
		t.Error("Property 2 should be read-only (no WriteSpec)")
	}
}

func TestClassPropertyString(t *testing.T) {
	input := `
type TTest = class
	class property Count: Integer read GetCount write SetCount;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	classDecl := program.Statements[0].(*ast.ClassDecl)
	prop := classDecl.Properties[0]

	// Verify String() output includes "class property"
	propStr := prop.String()
	expected := "class property Count: Integer read GetCount write SetCount;"
	if propStr != expected {
		t.Errorf("Property String():\nExpected: %q\nGot:      %q", expected, propStr)
	}
}

func TestMixedInstanceAndClassProperties(t *testing.T) {
	input := `
type TMixed = class
	private
		FName: String;
	public
		property Name: String read FName write FName;
		class property Count: Integer read GetCount;
end;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	classDecl := program.Statements[0].(*ast.ClassDecl)

	if len(classDecl.Properties) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(classDecl.Properties))
	}

	// First property should be instance property
	prop1 := classDecl.Properties[0]
	if prop1.IsClassProperty {
		t.Error("Property 'Name' should be an instance property (IsClassProperty=false)")
	}

	// Second property should be class property
	prop2 := classDecl.Properties[1]
	if !prop2.IsClassProperty {
		t.Error("Property 'Count' should be a class property (IsClassProperty=true)")
	}
}
