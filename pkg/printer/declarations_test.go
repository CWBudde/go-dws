package printer_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Helper function to convert int to pointer
func intPtr(i int) *int {
	return &i
}

// TestPrintEnumDecl tests enum declaration printing
func TestPrintEnumDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.EnumDecl
		expected string
	}{
		{
			name: "simple enum",
			node: &ast.EnumDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TColor"},
						},
					},
					Value: "TColor",
				},
				Values: []ast.EnumValue{
					{Name: "Red"},
					{Name: "Green"},
					{Name: "Blue"},
				},
			},
			expected: "typeTColor=(Red,Green,Blue)",
		},
		{
			name: "enum with explicit values",
			node: &ast.EnumDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TStatus"},
						},
					},
					Value: "TStatus",
				},
				Values: []ast.EnumValue{
					{Name: "Idle", Value: intPtr(0)},
					{Name: "Running", Value: intPtr(1)},
					{Name: "Stopped", Value: intPtr(2)},
				},
			},
			expected: "typeTStatus=(Idle=0,Running=1,Stopped=2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestPrintArrayDecl tests array declaration printing
func TestPrintArrayDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ArrayDecl
		contains string
	}{
		{
			name: "simple array declaration",
			node: &ast.ArrayDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TIntArray"},
						},
					},
					Value: "TIntArray",
				},
				ArrayType: &ast.ArrayTypeAnnotation{
					ElementType: &ast.TypeAnnotation{Name: "Integer"},
				},
			},
			contains: "typeTIntArray=array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintSetDecl tests set declaration printing
func TestPrintSetDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.SetDecl
		expected string
	}{
		{
			name: "simple set declaration",
			node: &ast.SetDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TCharSet"},
						},
					},
					Value: "TCharSet",
				},
				ElementType: &ast.TypeAnnotation{Name: "Char"},
			},
			expected: "typeTCharSet=setofChar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestPrintInterfaceDecl tests interface declaration printing
func TestPrintInterfaceDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.InterfaceDecl
		contains string
	}{
		{
			name: "simple interface",
			node: &ast.InterfaceDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INTERFACE, Literal: "interface"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "IComparable"},
						},
					},
					Value: "IComparable",
				},
				Methods: []*ast.InterfaceMethodDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.FUNCTION, Literal: "function"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "CompareTo"},
								},
							},
							Value: "CompareTo",
						},
						Parameters: []*ast.Parameter{
							{
								Name: &ast.Identifier{
									TypedExpressionBase: ast.TypedExpressionBase{
										BaseNode: ast.BaseNode{
											Token: token.Token{Type: token.IDENT, Literal: "other"},
										},
									},
									Value: "other",
								},
								Type: &ast.TypeAnnotation{Name: "IComparable"},
							},
						},
						ReturnType: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
			},
			contains: "type IComparable = interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintInterfaceMethodDecl tests interface method declaration printing
func TestPrintInterfaceMethodDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.InterfaceDecl
		contains string
	}{
		{
			name: "interface with procedure",
			node: &ast.InterfaceDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.INTERFACE, Literal: "interface"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "ITest"},
						},
					},
					Value: "ITest",
				},
				Methods: []*ast.InterfaceMethodDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.PROCEDURE, Literal: "procedure"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "DoSomething"},
								},
							},
							Value: "DoSomething",
						},
					},
				},
			},
			contains: "procedure DoSomething",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintTypeDeclaration tests type alias declaration printing
func TestPrintTypeDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.TypeDeclaration
		expected string
	}{
		{
			name: "type alias",
			node: &ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TMyInt"},
						},
					},
					Value: "TMyInt",
				},
				AliasedType: &ast.TypeAnnotation{Name: "Integer"},
			},
			expected: "typeTMyInt=Integer",
		},
		{
			name: "subrange type",
			node: &ast.TypeDeclaration{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.TYPE, Literal: "type"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TDigit"},
						},
					},
					Value: "TDigit",
				},
				IsSubrange: true,
				LowBound: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "0"},
						},
					},
					Value: 0,
				},
				HighBound: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "9"},
						},
					},
					Value: 9,
				},
			},
			expected: "typeTDigit=0..9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestPrintUnitDeclaration tests unit declaration printing
func TestPrintUnitDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.UnitDeclaration
		contains string
	}{
		{
			name: "simple unit",
			node: &ast.UnitDeclaration{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.UNIT, Literal: "unit"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "MyUnit"},
						},
					},
					Value: "MyUnit",
				},
				InterfaceSection: &ast.BlockStatement{
					Statements: []ast.Statement{},
				},
				ImplementationSection: &ast.BlockStatement{
					Statements: []ast.Statement{},
				},
			},
			contains: "unit MyUnit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintHelperDecl tests helper declaration printing
func TestPrintHelperDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.HelperDecl
		contains string
	}{
		{
			name: "simple helper",
			node: &ast.HelperDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.HELPER, Literal: "helper"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TStringHelper"},
						},
					},
					Value: "TStringHelper",
				},
				ForType: &ast.TypeAnnotation{Name: "String"},
				Methods: []*ast.FunctionDecl{},
			},
			contains: "type TStringHelper = helper for String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format:      printer.FormatDWScript,
				Style:       printer.StyleDetailed,
				IndentWidth: 2,
			})
			result := p.Print(tt.node)
			if !contains(result, tt.contains) {
				t.Errorf("Expected output to contain %q, got: %q", tt.contains, result)
			}
		})
	}
}

// TestPrintArrayTypeAnnotation tests array type annotation printing
func TestPrintArrayTypeAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ArrayTypeAnnotation
		expected string
	}{
		{
			name: "dynamic array",
			node: &ast.ArrayTypeAnnotation{
				ElementType: &ast.TypeAnnotation{Name: "Integer"},
			},
			expected: "arrayofInteger",
		},
		{
			name: "static array",
			node: &ast.ArrayTypeAnnotation{
				LowBound: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "0"},
						},
					},
					Value: 0,
				},
				HighBound: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.INT, Literal: "9"},
						},
					},
					Value: 9,
				},
				ElementType: &ast.TypeAnnotation{Name: "String"},
			},
			expected: "array[0..9]ofString",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestPrintArrayTypeNode tests array type node printing
func TestPrintArrayTypeNode(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ArrayTypeNode
		expected string
	}{
		{
			name: "dynamic array type",
			node: &ast.ArrayTypeNode{
				ElementType: &ast.TypeAnnotation{Name: "Float"},
			},
			expected: "arrayofFloat",
		},
		{
			name: "indexed array type",
			node: &ast.ArrayTypeNode{
				IndexType:   &ast.TypeAnnotation{Name: "TColor"},
				ElementType: &ast.TypeAnnotation{Name: "String"},
			},
			expected: "array[TColor]ofString",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestPrintSetTypeNode tests set type node printing
func TestPrintSetTypeNode(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.SetTypeNode
		expected string
	}{
		{
			name: "set type",
			node: &ast.SetTypeNode{
				ElementType: &ast.TypeAnnotation{Name: "Byte"},
			},
			expected: "setofByte",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestPrintClassOfTypeNode tests class-of type node printing
func TestPrintClassOfTypeNode(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ClassOfTypeNode
		expected string
	}{
		{
			name: "class of type",
			node: &ast.ClassOfTypeNode{
				ClassType: &ast.TypeAnnotation{Name: "TComponent"},
			},
			expected: "classofTComponent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}

// TestPrintFunctionPointerTypeNode tests function pointer type node printing
func TestPrintFunctionPointerTypeNode(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.FunctionPointerTypeNode
		expected string
	}{
		{
			name: "function pointer",
			node: &ast.FunctionPointerTypeNode{
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "x"},
								},
							},
							Value: "x",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
				ReturnType: &ast.TypeAnnotation{Name: "String"},
			},
			expected: "function(x:Integer):String",
		},
		{
			name: "procedure pointer",
			node: &ast.FunctionPointerTypeNode{
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "msg"},
								},
							},
							Value: "msg",
						},
						Type: &ast.TypeAnnotation{Name: "String"},
					},
				},
			},
			expected: "procedure(msg:String)",
		},
		{
			name: "method pointer (of object)",
			node: &ast.FunctionPointerTypeNode{
				Parameters: []*ast.Parameter{},
				ReturnType: &ast.TypeAnnotation{Name: "Boolean"},
				OfObject:   true,
			},
			expected: "function:Booleanof object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer.New(printer.Options{
				Format: printer.FormatDWScript,
				Style:  printer.StyleCompact,
			})
			result := p.Print(tt.node)
			if result != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, result)
			}
		})
	}
}
