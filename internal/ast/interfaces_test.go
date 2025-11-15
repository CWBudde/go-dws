package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// ============================================================================
// InterfaceMethodDecl Tests
// ============================================================================

func TestInterfaceMethodDeclString(t *testing.T) {
	tests := []struct {
		name     string
		method   *InterfaceMethodDecl
		expected string
	}{
		{
			name: "simple procedure (no parameters, no return type)",
			method: &InterfaceMethodDecl{
				Name:       NewTestIdentifier("Hello"),
				Parameters: []*Parameter{},
				ReturnType: nil,
			},
			expected: "procedure Hello",
		},
		{
			name: "procedure with one parameter",
			method: &InterfaceMethodDecl{
				Name: NewTestIdentifier("SetValue"),
				Parameters: []*Parameter{
					{
						Name:  NewTestIdentifier("x"),
						Type:  NewTestTypeAnnotation("Integer"),
						ByRef: false,
					},
				},
				ReturnType: nil,
			},
			expected: "procedure SetValue(x: Integer)",
		},
		{
			name: "function with return type",
			method: &InterfaceMethodDecl{
				Name:       NewTestIdentifier("GetValue"),
				Parameters: []*Parameter{},
				ReturnType: NewTestTypeAnnotation("Integer"),
			},
			expected: "function GetValue: Integer",
		},
		{
			name: "function with parameters and return type",
			method: &InterfaceMethodDecl{
				Name: NewTestIdentifier("Add"),
				Parameters: []*Parameter{
					{
						Name:  NewTestIdentifier("a"),
						Type:  NewTestTypeAnnotation("Integer"),
						ByRef: false,
					},
					{
						Name:  NewTestIdentifier("b"),
						Type:  NewTestTypeAnnotation("Integer"),
						ByRef: false,
					},
				},
				ReturnType: NewTestTypeAnnotation("Integer"),
			},
			expected: "function Add(a: Integer; b: Integer): Integer",
		},
		{
			name: "procedure with var parameter",
			method: &InterfaceMethodDecl{
				Name: NewTestIdentifier("Swap"),
				Parameters: []*Parameter{
					{
						Name:  NewTestIdentifier("a"),
						Type:  NewTestTypeAnnotation("Integer"),
						ByRef: true,
					},
					{
						Name:  NewTestIdentifier("b"),
						Type:  NewTestTypeAnnotation("Integer"),
						ByRef: true,
					},
				},
				ReturnType: nil,
			},
			expected: "procedure Swap(var a: Integer; var b: Integer)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method.String()
			if result != tt.expected {
				t.Errorf("InterfaceMethodDecl.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// InterfaceDecl Tests
// ============================================================================

func TestInterfaceDeclString(t *testing.T) {
	tests := []struct {
		name     string
		iface    *InterfaceDecl
		expected string
	}{
		{
			name: "simple interface without methods",
			iface: &InterfaceDecl{
									BaseNode: BaseNode{Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name:    NewTestIdentifier("IEmpty"),
				Parent:  nil,
				Methods: []*InterfaceMethodDecl{},
			},
			expected: "type IEmpty = interface\nend",
		},
		{
			name: "interface with parent (inheritance)",
			iface: &InterfaceDecl{
									BaseNode: BaseNode{Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name:    NewTestIdentifier("IDescendent"),
				Parent:  NewTestIdentifier("IBase"),
				Methods: []*InterfaceMethodDecl{},
			},
			expected: "type IDescendent = interface(IBase)\nend",
		},
		{
			name: "interface with single method",
			iface: &InterfaceDecl{
									BaseNode: BaseNode{Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name:   NewTestIdentifier("IMy"),
				Parent: nil,
				Methods: []*InterfaceMethodDecl{
					{
						Name:       NewTestIdentifier("Hello"),
						Parameters: []*Parameter{},
						ReturnType: nil,
					},
				},
			},
			expected: "type IMy = interface\n  procedure Hello;\nend",
		},
		{
			name: "interface with multiple methods",
			iface: &InterfaceDecl{
									BaseNode: BaseNode{Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name:   NewTestIdentifier("ICounter"),
				Parent: nil,
				Methods: []*InterfaceMethodDecl{
					{
						Name:       NewTestIdentifier("Increment"),
						Parameters: []*Parameter{},
						ReturnType: nil,
					},
					{
						Name:       NewTestIdentifier("GetValue"),
						Parameters: []*Parameter{},
						ReturnType: NewTestTypeAnnotation("Integer"),
					},
				},
			},
			expected: "type ICounter = interface\n  procedure Increment;\n  function GetValue: Integer;\nend",
		},
		{
			name: "interface with parent and methods (IDescendent extends IBase)",
			iface: &InterfaceDecl{
									BaseNode: BaseNode{Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				Name:   NewTestIdentifier("IDescendent"),
				Parent: NewTestIdentifier("IBase"),
				Methods: []*InterfaceMethodDecl{
					{
						Name:       NewTestIdentifier("B"),
						Parameters: []*Parameter{},
						ReturnType: nil,
					},
				},
			},
			expected: "type IDescendent = interface(IBase)\n  procedure B;\nend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.iface.String()
			if result != tt.expected {
				t.Errorf("InterfaceDecl.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// ClassDecl with Interfaces Tests
// ============================================================================

func TestClassDeclWithInterfacesString(t *testing.T) {
	tests := []struct {
		name      string
		classDecl *ClassDecl
		expected  string
	}{
		{
			name: "class implementing single interface",
			classDecl: &ClassDecl{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.TYPE, Literal: "type"}},
				Name:     NewTestIdentifier("TTest"),
				Parent:   NewTestIdentifier("TObject"),
				Interfaces: []*Identifier{
					NewTestIdentifier("IMy"),
				},
				Fields:  []*FieldDecl{},
				Methods: []*FunctionDecl{},
			},
			expected: "type TTest = class(TObject, IMy)\nend",
		},
		{
			name: "class implementing multiple interfaces",
			classDecl: &ClassDecl{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.TYPE, Literal: "type"}},
				Name:     NewTestIdentifier("TImpAB"),
				Parent:   NewTestIdentifier("TObject"),
				Interfaces: []*Identifier{
					NewTestIdentifier("IIntfA"),
					NewTestIdentifier("IIntfB"),
				},
				Fields:  []*FieldDecl{},
				Methods: []*FunctionDecl{},
			},
			expected: "type TImpAB = class(TObject, IIntfA, IIntfB)\nend",
		},
		{
			name: "class with no parent but implementing interface",
			classDecl: &ClassDecl{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.TYPE, Literal: "type"}},
				Name:     NewTestIdentifier("TSimple"),
				Parent:   nil,
				Interfaces: []*Identifier{
					NewTestIdentifier("IMyInterface"),
				},
				Fields:  []*FieldDecl{},
				Methods: []*FunctionDecl{},
			},
			expected: "type TSimple = class(IMyInterface)\nend",
		},
		{
			name: "class with parent, interfaces, and methods",
			classDecl: &ClassDecl{
				BaseNode: BaseNode{Token: lexer.Token{Type: lexer.TYPE, Literal: "type"}},
				Name:     NewTestIdentifier("TTest"),
				Parent:   NewTestIdentifier("TObject"),
				Interfaces: []*Identifier{
					NewTestIdentifier("IIntf"),
				},
				Fields: []*FieldDecl{},
				Methods: []*FunctionDecl{
					{
													BaseNode: BaseNode{Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
						Name:       NewTestIdentifier("Hello"),
						Parameters: []*Parameter{},
						ReturnType: nil,
						Body: &BlockStatement{
							BaseNode: BaseNode{Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
							Statements: []Statement{},
						},
					},
				},
			},
			expected: "type TTest = class(TObject, IIntf)\n  procedure Hello begin\n  end;\nend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.classDecl.String()
			if result != tt.expected {
				t.Errorf("ClassDecl.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
