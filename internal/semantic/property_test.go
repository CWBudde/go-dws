package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Property Semantic Analysis Tests (Task 8.52)
// ============================================================================

// TestPropertyDeclaration tests valid property declarations with various access patterns
func TestPropertyDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "field-backed property",
			input: `
type TTest = class
	FName: String;
	property Name: String read FName write FName;
end;`,
		},
		{
			name: "method-backed property",
			input: `
type TTest = class
	function GetCount: Integer; begin Result := 0; end;
	procedure SetCount(value: Integer); begin end;
	property Count: Integer read GetCount write SetCount;
end;`,
		},
		{
			name: "read-only property",
			input: `
type TTest = class
	FSize: Integer;
	property Size: Integer read FSize;
end;`,
		},
		{
			name: "write-only property",
			input: `
type TTest = class
	FData: String;
	property Data: String write FData;
end;`,
		},
		{
			name: "indexed property with methods",
			input: `
type TTest = class
	function GetItem(i: Integer): String; begin Result := ''; end;
	procedure SetItem(i: Integer; value: String); begin end;
	property Items[i: Integer]: String read GetItem write SetItem;
end;`,
		},
		{
			name: "default indexed property",
			input: `
type TTest = class
	function GetItem(i: Integer): String; begin Result := ''; end;
	procedure SetItem(i: Integer; value: String); begin end;
	property Items[i: Integer]: String read GetItem write SetItem; default;
end;`,
		},
		{
			name: "multi-index property",
			input: `
type TTest = class
	function GetData(x, y: Integer): Float; begin Result := 0.0; end;
	procedure SetData(x, y: Integer; value: Float); begin end;
	property Data[x, y: Integer]: Float read GetData write SetData;
end;`,
		},
		{
			name: "auto-property generates backing field",
			input: `
type TTest = class
	FValue: Integer;
	property Value: Integer;
end;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if err != nil {
				t.Errorf("unexpected semantic error: %v", err)
			}
		})
	}
}

// TestPropertyErrors tests various error cases in property declarations
func TestPropertyErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "duplicate property name",
			input: `
type TTest = class
	FValue: Integer;
	property Value: Integer read FValue;
	property Value: String read FValue;
end;`,
			expectedError: "duplicate property 'Value'",
		},
		{
			name: "read field not found",
			input: `
type TTest = class
	property Name: String read FName;
end;`,
			expectedError: "read specifier 'FName' not found",
		},
		{
			name: "write field not found",
			input: `
type TTest = class
	property Name: String write FName;
end;`,
			expectedError: "write specifier 'FName' not found",
		},
		{
			name: "read method not found",
			input: `
type TTest = class
	property Count: Integer read GetCount;
end;`,
			expectedError: "read specifier 'GetCount' not found",
		},
		{
			name: "write method not found",
			input: `
type TTest = class
	property Count: Integer write SetCount;
end;`,
			expectedError: "write specifier 'SetCount' not found",
		},
		{
			name: "read field type mismatch",
			input: `
type TTest = class
	FValue: String;
	property Count: Integer read FValue;
end;`,
			expectedError: "read field 'FValue' has type String, expected Integer",
		},
		{
			name: "write field type mismatch",
			input: `
type TTest = class
	FValue: String;
	property Count: Integer write FValue;
end;`,
			expectedError: "write field 'FValue' has type String, expected Integer",
		},
		{
			name: "getter wrong return type",
			input: `
type TTest = class
	function GetCount: String; begin Result := ''; end;
	property Count: Integer read GetCount;
end;`,
			expectedError: "getter method 'GetCount' returns String, expected Integer",
		},
		{
			name: "setter wrong parameter type",
			input: `
type TTest = class
	procedure SetCount(value: String); begin end;
	property Count: Integer write SetCount;
end;`,
			expectedError: "setter method 'SetCount' value parameter has type String, expected Integer",
		},
		{
			name: "setter not void",
			input: `
type TTest = class
	function SetCount(value: Integer): Boolean; begin Result := true; end;
	property Count: Integer write SetCount;
end;`,
			expectedError: "setter method 'SetCount' must return void",
		},
		{
			name: "getter has parameters for non-indexed property",
			input: `
type TTest = class
	function GetValue(x: Integer): Integer; begin Result := 0; end;
	property Value: Integer read GetValue;
end;`,
			expectedError: "getter method 'GetValue' has 1 parameters, expected 0",
		},
		{
			name: "setter missing value parameter for non-indexed property",
			input: `
type TTest = class
	procedure SetValue; begin end;
	property Value: Integer write SetValue;
end;`,
			expectedError: "setter method 'SetValue' has 0 parameters, expected 1",
		},
		{
			name: "default property not indexed",
			input: `
type TTest = class
	FValue: Integer;
	property Value: Integer read FValue; default;
end;`,
			expectedError: "default property 'Value' must be an indexed property",
		},
		{
			name: "multiple default properties",
			input: `
type TTest = class
	function GetItem(i: Integer): String; begin Result := ''; end;
	function GetData(i: Integer): Integer; begin Result := 0; end;
	property Items[i: Integer]: String read GetItem; default;
	property Data[i: Integer]: Integer read GetData; default;
end;`,
			expectedError: "already has default property 'Items'",
		},
		{
			name: "indexed property getter missing index parameter",
			input: `
type TTest = class
	function GetItem: String; begin Result := ''; end;
	property Items[index: Integer]: String read GetItem;
end;`,
			expectedError: "getter method 'GetItem' has 0 parameters, expected 1",
		},
		{
			name: "indexed property setter missing index parameter",
			input: `
type TTest = class
	procedure SetItem(value: String); begin end;
	property Items[index: Integer]: String write SetItem;
end;`,
			expectedError: "setter method 'SetItem' has 1 parameters, expected 2",
		},
		{
			name: "indexed property getter wrong index type",
			input: `
type TTest = class
	function GetItem(i: String): Integer; begin Result := 0; end;
	property Items[i: Integer]: Integer read GetItem;
end;`,
			expectedError: "getter method 'GetItem' parameter 1 has type String, expected Integer",
		},
		{
			name: "indexed property setter wrong index type",
			input: `
type TTest = class
	procedure SetItem(i: String; value: Integer); begin end;
	property Items[i: Integer]: Integer write SetItem;
end;`,
			expectedError: "setter method 'SetItem' parameter 1 has type String, expected Integer",
		},
		{
			name: "property with unknown type",
			input: `
type TTest = class
	property Value: UnknownType read FValue;
end;`,
			expectedError: "unknown type 'UnknownType'",
		},
		{
			name: "property with unknown index parameter type",
			input: `
type TTest = class
	function GetItem(i: UnknownType): String; begin Result := ''; end;
	property Items[i: UnknownType]: String read GetItem;
end;`,
			expectedError: "unknown type 'UnknownType'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if err == nil {
				t.Errorf("expected error containing '%s', got no error", tt.expectedError)
				return
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.expectedError) {
				t.Errorf("expected error containing '%s', got '%s'", tt.expectedError, errMsg)
			}
		})
	}
}

// TestPropertyInheritance tests property access through inheritance
func TestPropertyInheritance(t *testing.T) {
	input := `
type TBase = class
	FValue: Integer;
	property Value: Integer read FValue write FValue;
end;

type TDerived = class(TBase)
	FName: String;
	property Name: String read FName write FName;
end;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("unexpected semantic error: %v", err)
	}

	// Verify both classes have their properties registered
	baseClass := analyzer.classes["TBase"]
	if baseClass == nil {
		t.Fatal("TBase class not found")
	}
	if _, found := baseClass.Properties["Value"]; !found {
		t.Error("TBase should have property 'Value'")
	}

	derivedClass := analyzer.classes["TDerived"]
	if derivedClass == nil {
		t.Fatal("TDerived class not found")
	}
	if _, found := derivedClass.Properties["Name"]; !found {
		t.Error("TDerived should have property 'Name'")
	}

	// Verify derived class can access parent property through inheritance
	if _, found := derivedClass.GetProperty("Value"); !found {
		t.Error("TDerived should have access to inherited property 'Value'")
	}
}

// TestPropertyAccessKinds tests that PropertyInfo is correctly populated with access kinds
func TestPropertyAccessKinds(t *testing.T) {
	input := `
type TTest = class
	FName: String;
	function GetCount: Integer; begin Result := 0; end;
	procedure SetCount(value: Integer); begin end;

	property Name: String read FName write FName;
	property Count: Integer read GetCount write SetCount;
	property Size: String read FName;
end;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("unexpected semantic error: %v", err)
	}

	class := analyzer.classes["TTest"]
	if class == nil {
		t.Fatal("TTest class not found")
	}

	// Check Name property (field-backed)
	nameProp, found := class.Properties["Name"]
	if !found {
		t.Fatal("Name property not found")
	}
	if nameProp.ReadKind != types.PropAccessField {
		t.Errorf("Name property ReadKind should be PropAccessField, got %v", nameProp.ReadKind)
	}
	if nameProp.WriteKind != types.PropAccessField {
		t.Errorf("Name property WriteKind should be PropAccessField, got %v", nameProp.WriteKind)
	}

	// Check Count property (method-backed)
	countProp, found := class.Properties["Count"]
	if !found {
		t.Fatal("Count property not found")
	}
	if countProp.ReadKind != types.PropAccessMethod {
		t.Errorf("Count property ReadKind should be PropAccessMethod, got %v", countProp.ReadKind)
	}
	if countProp.WriteKind != types.PropAccessMethod {
		t.Errorf("Count property WriteKind should be PropAccessMethod, got %v", countProp.WriteKind)
	}

	// Check Size property (no write access)
	sizeProp, found := class.Properties["Size"]
	if !found {
		t.Fatal("Size property not found")
	}
	if sizeProp.ReadKind != types.PropAccessField {
		t.Errorf("Size property ReadKind should be PropAccessField, got %v", sizeProp.ReadKind)
	}
	if sizeProp.WriteKind != types.PropAccessNone {
		t.Errorf("Size property WriteKind should be PropAccessNone, got %v", sizeProp.WriteKind)
	}
}

// TestIndexedPropertyValidation tests indexed property parameter validation
func TestIndexedPropertyValidation(t *testing.T) {
	input := `
type TTest = class
	function GetItem(i: Integer): String; begin Result := ''; end;
	procedure SetItem(i: Integer; value: String); begin end;
	function GetData(x, y: Integer): Float; begin Result := 0.0; end;

	property Items[i: Integer]: String read GetItem write SetItem;
	property Data[x, y: Integer]: Float read GetData; default;
end;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err != nil {
		t.Errorf("unexpected semantic error: %v", err)
	}

	class := analyzer.classes["TTest"]
	if class == nil {
		t.Fatal("TTest class not found")
	}

	// Check Items property
	itemsProp, found := class.Properties["Items"]
	if !found {
		t.Fatal("Items property not found")
	}
	if !itemsProp.IsIndexed {
		t.Error("Items property should be indexed")
	}
	if itemsProp.IsDefault {
		t.Error("Items property should not be default")
	}

	// Check Data property
	dataProp, found := class.Properties["Data"]
	if !found {
		t.Fatal("Data property not found")
	}
	if !dataProp.IsIndexed {
		t.Error("Data property should be indexed")
	}
	if !dataProp.IsDefault {
		t.Error("Data property should be default")
	}
}
