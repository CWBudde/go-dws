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
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Hello"},
					Value: "Hello",
				},
				Parameters: []*Parameter{},
				ReturnType: nil,
			},
			expected: "procedure Hello",
		},
		{
			name: "procedure with one parameter",
			method: &InterfaceMethodDecl{
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "SetValue"},
					Value: "SetValue",
				},
				Parameters: []*Parameter{
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
							Value: "x",
						},
						Type: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
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
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "GetValue"},
					Value: "GetValue",
				},
				Parameters: []*Parameter{},
				ReturnType: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
			expected: "function GetValue: Integer",
		},
		{
			name: "function with parameters and return type",
			method: &InterfaceMethodDecl{
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Add"},
					Value: "Add",
				},
				Parameters: []*Parameter{
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "a"},
							Value: "a",
						},
						Type: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
						ByRef: false,
					},
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "b"},
							Value: "b",
						},
						Type: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
						ByRef: false,
					},
				},
				ReturnType: &TypeAnnotation{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
					Name:  "Integer",
				},
			},
			expected: "function Add(a: Integer; b: Integer): Integer",
		},
		{
			name: "procedure with var parameter",
			method: &InterfaceMethodDecl{
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Swap"},
					Value: "Swap",
				},
				Parameters: []*Parameter{
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "a"},
							Value: "a",
						},
						Type: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
						ByRef: true,
					},
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "b"},
							Value: "b",
						},
						Type: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
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
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "IEmpty"},
					Value: "IEmpty",
				},
				Parent:  nil,
				Methods: []*InterfaceMethodDecl{},
			},
			expected: "type IEmpty = interface\nend",
		},
		{
			name: "interface with parent (inheritance)",
			iface: &InterfaceDecl{
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "IDescendent"},
					Value: "IDescendent",
				},
				Parent: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "IBase"},
					Value: "IBase",
				},
				Methods: []*InterfaceMethodDecl{},
			},
			expected: "type IDescendent = interface(IBase)\nend",
		},
		{
			name: "interface with single method",
			iface: &InterfaceDecl{
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "IMy"},
					Value: "IMy",
				},
				Parent: nil,
				Methods: []*InterfaceMethodDecl{
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Hello"},
							Value: "Hello",
						},
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
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "ICounter"},
					Value: "ICounter",
				},
				Parent: nil,
				Methods: []*InterfaceMethodDecl{
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Increment"},
							Value: "Increment",
						},
						Parameters: []*Parameter{},
						ReturnType: nil,
					},
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "GetValue"},
							Value: "GetValue",
						},
						Parameters: []*Parameter{},
						ReturnType: &TypeAnnotation{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Integer"},
							Name:  "Integer",
						},
					},
				},
			},
			expected: "type ICounter = interface\n  procedure Increment;\n  function GetValue: Integer;\nend",
		},
		{
			name: "interface with parent and methods (IDescendent extends IBase)",
			iface: &InterfaceDecl{
				BaseNode: BaseNode{
					Token: lexer.Token{Type: lexer.TYPE, Literal: "type"},
				},
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "IDescendent"},
					Value: "IDescendent",
				},
				Parent: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "IBase"},
					Value: "IBase",
				},
				Methods: []*InterfaceMethodDecl{
					{
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "B"},
							Value: "B",
						},
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
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TTest"},
					Value: "TTest",
				},
				Parent: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TObject"},
					Value: "TObject",
				},
				Interfaces: []*Identifier{
					{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "IMy"},
						Value: "IMy",
					},
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
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TImpAB"},
					Value: "TImpAB",
				},
				Parent: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TObject"},
					Value: "TObject",
				},
				Interfaces: []*Identifier{
					{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "IIntfA"},
						Value: "IIntfA",
					},
					{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "IIntfB"},
						Value: "IIntfB",
					},
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
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TSimple"},
					Value: "TSimple",
				},
				Parent: nil,
				Interfaces: []*Identifier{
					{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "IMyInterface"},
						Value: "IMyInterface",
					},
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
				Name: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TTest"},
					Value: "TTest",
				},
				Parent: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "TObject"},
					Value: "TObject",
				},
				Interfaces: []*Identifier{
					{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "IIntf"},
						Value: "IIntf",
					},
				},
				Fields: []*FieldDecl{},
				Methods: []*FunctionDecl{
					{
						BaseNode: BaseNode{
							Token: lexer.Token{Type: lexer.PROCEDURE, Literal: "procedure"},
						},
						Name: &Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "Hello"},
							Value: "Hello",
						},
						Parameters: []*Parameter{},
						ReturnType: nil,
						Body: &BlockStatement{
							Token:      lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
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
