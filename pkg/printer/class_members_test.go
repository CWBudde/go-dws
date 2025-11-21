package printer_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// TestPrintParameter tests parameter printing
func TestPrintParameter(t *testing.T) {
	// Parameters are tested indirectly through function declarations
	// This test ensures different parameter types work correctly
	tests := []struct {
		name     string
		node     *ast.FunctionDecl
		contains string
	}{
		{
			name: "function with const parameter",
			node: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Test"},
						},
					},
					Value: "Test",
				},
				Parameters: []*ast.Parameter{
					{
						IsConst: true,
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
				ReturnType: &ast.TypeAnnotation{Name: "Integer"},
			},
			contains: "constx:Integer",
		},
		{
			name: "function with var parameter",
			node: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Test"},
						},
					},
					Value: "Test",
				},
				Parameters: []*ast.Parameter{
					{
						ByRef: true,
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "y"},
								},
							},
							Value: "y",
						},
						Type: &ast.TypeAnnotation{Name: "String"},
					},
				},
				ReturnType: &ast.TypeAnnotation{Name: "Boolean"},
			},
			contains: "vary:String",
		},
		{
			name: "function with default value",
			node: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Test"},
						},
					},
					Value: "Test",
				},
				Parameters: []*ast.Parameter{
					{
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "z"},
								},
							},
							Value: "z",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
						DefaultValue: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "42"},
								},
							},
							Value: 42,
						},
					},
				},
				ReturnType: &ast.TypeAnnotation{Name: "Integer"},
			},
			contains: "z:Integer=42",
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

// TestPrintFieldDecl tests field declaration printing
func TestPrintFieldDecl(t *testing.T) {
	// Field declarations are tested through class declarations
	tests := []struct {
		name     string
		node     *ast.ClassDecl
		contains string
	}{
		{
			name: "class with private field",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Fields: []*ast.FieldDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "fValue"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "fValue"},
								},
							},
							Value: "fValue",
						},
						Type:       &ast.TypeAnnotation{Name: "Integer"},
						Visibility: ast.VisibilityPrivate,
					},
				},
			},
			contains: "private fValue: Integer",
		},
		{
			name: "class with field with init value",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Fields: []*ast.FieldDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Count"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Count"},
								},
							},
							Value: "Count",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
						InitValue: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "0"},
								},
							},
							Value: 0,
						},
					},
				},
			},
			contains: "Count: Integer = 0",
		},
		{
			name: "class with field type inference",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Fields: []*ast.FieldDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Name"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Name"},
								},
							},
							Value: "Name",
						},
						InitValue: &ast.StringLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.STRING, Literal: "\"test\""},
								},
							},
							Value: "test",
						},
					},
				},
			},
			contains: "Name := \"test\"",
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

// TestPrintPropertyDecl tests property declaration printing
func TestPrintPropertyDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ClassDecl
		contains string
	}{
		{
			name: "class with property",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Properties: []*ast.PropertyDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.PROPERTY, Literal: "property"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Value"},
								},
							},
							Value: "Value",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
						ReadSpec: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "fValue"},
								},
							},
							Value: "fValue",
						},
						WriteSpec: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "fValue"},
								},
							},
							Value: "fValue",
						},
					},
				},
			},
			contains: "property Value: Integer read fValue write fValue",
		},
		{
			name: "class with default property",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TTest"},
						},
					},
					Value: "TTest",
				},
				Properties: []*ast.PropertyDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.PROPERTY, Literal: "property"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Items"},
								},
							},
							Value: "Items",
						},
						Type: &ast.TypeAnnotation{Name: "String"},
						ReadSpec: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "GetItem"},
								},
							},
							Value: "GetItem",
						},
						IsDefault: true,
					},
				},
			},
			contains: "default",
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

// TestPrintRecordDecl tests record declaration printing
func TestPrintRecordDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.RecordDecl
		contains string
	}{
		{
			name: "simple record",
			node: &ast.RecordDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.RECORD, Literal: "record"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TPoint"},
						},
					},
					Value: "TPoint",
				},
				Fields: []*ast.FieldDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "X"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "X"},
								},
							},
							Value: "X",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Y"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Y"},
								},
							},
							Value: "Y",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
			},
			contains: "type TPoint = record",
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

// TestPrintRecordPropertyDecl tests record property declaration printing
func TestPrintRecordPropertyDecl(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.RecordPropertyDecl
		expected string
	}{
		{
			name: "record property with read/write",
			node: &ast.RecordPropertyDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.PROPERTY, Literal: "property"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Value"},
						},
					},
					Value: "Value",
				},
				Type:       &ast.TypeAnnotation{Name: "Integer"},
				ReadField:  "fValue",
				WriteField: "fValue",
			},
			expected: "propertyValue:IntegerreadfValuewritefValue",
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
