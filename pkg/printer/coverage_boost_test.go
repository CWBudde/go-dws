package printer_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/cwbudde/go-dws/pkg/token"
)

// Additional tests to boost coverage to 80%+

// TestPrintIfStatementDetailed tests different if statement variations
func TestPrintIfStatementDetailed(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.IfStatement
		contains string
	}{
		{
			name: "if with non-block consequence",
			node: &ast.IfStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IF, Literal: "if"},
				},
				Condition: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.TRUE, Literal: "true"},
						},
					},
					Value: true,
				},
				Consequence: &ast.ExpressionStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.IDENT, Literal: "x"},
					},
					Expression: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "x"},
							},
						},
						Value: "x",
					},
				},
			},
			contains: "if true then",
		},
		{
			name: "if-else if chain",
			node: &ast.IfStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IF, Literal: "if"},
				},
				Condition: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.TRUE, Literal: "true"},
						},
					},
					Value: true,
				},
				Consequence: &ast.ExpressionStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.IDENT, Literal: "a"},
					},
					Expression: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "a"},
							},
						},
						Value: "a",
					},
				},
				Alternative: &ast.IfStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.IF, Literal: "if"},
					},
					Condition: &ast.BooleanLiteral{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.FALSE, Literal: "false"},
							},
						},
						Value: false,
					},
					Consequence: &ast.ExpressionStatement{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "b"},
						},
						Expression: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "b"},
								},
							},
							Value: "b",
						},
					},
				},
			},
			contains: "else if",
		},
		{
			name: "if with non-block alternative",
			node: &ast.IfStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.IF, Literal: "if"},
				},
				Condition: &ast.BooleanLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.TRUE, Literal: "true"},
						},
					},
					Value: true,
				},
				Consequence: &ast.ExpressionStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.IDENT, Literal: "a"},
					},
					Expression: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "a"},
							},
						},
						Value: "a",
					},
				},
				Alternative: &ast.ExpressionStatement{
					BaseNode: ast.BaseNode{
						Token: token.Token{Type: token.IDENT, Literal: "b"},
					},
					Expression: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "b"},
							},
						},
						Value: "b",
					},
				},
			},
			contains: "else",
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

// TestPrintRecordDeclDetailed tests record with various members
func TestPrintRecordDeclDetailed(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.RecordDecl
		contains []string
	}{
		{
			name: "record with class constants",
			node: &ast.RecordDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.RECORD, Literal: "record"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TMyRecord"},
						},
					},
					Value: "TMyRecord",
				},
				Constants: []*ast.ConstDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.CONST, Literal: "const"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "MaxSize"},
								},
							},
							Value: "MaxSize",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
						Value: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "100"},
								},
							},
							Value: 100,
						},
						IsClassConst: true,
					},
				},
			},
			contains: []string{"class const", "MaxSize"},
		},
		{
			name: "record with class vars",
			node: &ast.RecordDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.RECORD, Literal: "record"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TMyRecord"},
						},
					},
					Value: "TMyRecord",
				},
				ClassVars: []*ast.FieldDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Counter"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "Counter"},
								},
							},
							Value: "Counter",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
			},
			contains: []string{"class var", "Counter"},
		},
		{
			name: "record with methods and properties",
			node: &ast.RecordDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.RECORD, Literal: "record"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TMyRecord"},
						},
					},
					Value: "TMyRecord",
				},
				Methods: []*ast.FunctionDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.FUNCTION, Literal: "function"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "ToString"},
								},
							},
							Value: "ToString",
						},
						ReturnType: &ast.TypeAnnotation{Name: "String"},
					},
				},
				Properties: []ast.RecordPropertyDecl{
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
						Type:      &ast.TypeAnnotation{Name: "Integer"},
						ReadField: "fValue",
					},
				},
			},
			contains: []string{"function ToString", "property Value"},
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
			for _, expected := range tt.contains {
				if !contains(result, expected) {
					t.Errorf("Expected output to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// TestPrintHelperDeclDetailed tests helper with class vars and consts
func TestPrintHelperDeclDetailed(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.HelperDecl
		contains []string
	}{
		{
			name: "helper with class vars and consts",
			node: &ast.HelperDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.HELPER, Literal: "helper"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TIntHelper"},
						},
					},
					Value: "TIntHelper",
				},
				ForType: &ast.TypeAnnotation{Name: "Integer"},
				ClassVars: []*ast.FieldDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "GlobalCount"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "GlobalCount"},
								},
							},
							Value: "GlobalCount",
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
				ClassConsts: []*ast.ConstDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.CONST, Literal: "const"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "MaxValue"},
								},
							},
							Value: "MaxValue",
						},
						Value: &ast.IntegerLiteral{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.INT, Literal: "999"},
								},
							},
							Value: 999,
						},
					},
				},
				Methods: []*ast.FunctionDecl{},
				Properties: []*ast.PropertyDecl{
					{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.PROPERTY, Literal: "property"},
						},
						Name: &ast.Identifier{
							TypedExpressionBase: ast.TypedExpressionBase{
								BaseNode: ast.BaseNode{
									Token: token.Token{Type: token.IDENT, Literal: "IsPositive"},
								},
							},
							Value: "IsPositive",
						},
						Type: &ast.TypeAnnotation{Name: "Boolean"},
					},
				},
			},
			contains: []string{"class var", "class const", "property IsPositive"},
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
			for _, expected := range tt.contains {
				if !contains(result, expected) {
					t.Errorf("Expected output to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// TestPrintClassDeclDetailed tests class with various flags
func TestPrintClassDeclDetailed(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.ClassDecl
		contains []string
	}{
		{
			name: "partial abstract class",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TBase"},
						},
					},
					Value: "TBase",
				},
				IsPartial:  true,
				IsAbstract: true,
			},
			contains: []string{"partial class", "abstract"},
		},
		{
			name: "external class",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TExternal"},
						},
					},
					Value: "TExternal",
				},
				IsExternal:   true,
				ExternalName: "JsObject",
			},
			contains: []string{"external", "'JsObject'"},
		},
		{
			name: "class with parent and interfaces",
			node: &ast.ClassDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.CLASS, Literal: "class"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TChild"},
						},
					},
					Value: "TChild",
				},
				Parent: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "TBase"},
						},
					},
					Value: "TBase",
				},
				Interfaces: []*ast.Identifier{
					{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "IComparable"},
							},
						},
						Value: "IComparable",
					},
					{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "IDisposable"},
							},
						},
						Value: "IDisposable",
					},
				},
			},
			contains: []string{"TBase", "IComparable", "IDisposable"},
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
			for _, expected := range tt.contains {
				if !contains(result, expected) {
					t.Errorf("Expected output to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

// TestPrintVarDeclStatementDetailed tests var declarations with external
func TestPrintVarDeclStatementDetailed(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.VarDeclStatement
		contains string
	}{
		{
			name: "external variable",
			node: &ast.VarDeclStatement{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.VAR, Literal: "var"},
				},
				Names: []*ast.Identifier{
					{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: token.Token{Type: token.IDENT, Literal: "console"},
							},
						},
						Value: "console",
					},
				},
				Type:         &ast.TypeAnnotation{Name: "TConsole"},
				IsExternal:   true,
				ExternalName: "window.console",
			},
			contains: "external 'window.console'",
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

// TestPrintFunctionDeclDetailed tests function with forward and external
func TestPrintFunctionDeclDetailed(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.FunctionDecl
		contains string
	}{
		{
			name: "forward declaration",
			node: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "DoWork"},
						},
					},
					Value: "DoWork",
				},
				ReturnType: &ast.TypeAnnotation{Name: "Integer"},
				IsForward:  true,
			},
			contains: "forward",
		},
		{
			name: "external function",
			node: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Alert"},
						},
					},
					Value: "Alert",
				},
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
				IsExternal:   true,
				ExternalName: "window.alert",
			},
			contains: "external 'window.alert'",
		},
		{
			name: "reintroduce method",
			node: &ast.FunctionDecl{
				BaseNode: ast.BaseNode{
					Token: token.Token{Type: token.FUNCTION, Literal: "function"},
				},
				Name: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: token.Token{Type: token.IDENT, Literal: "Create"},
						},
					},
					Value: "Create",
				},
				ReturnType:    &ast.TypeAnnotation{Name: "TObject"},
				IsReintroduce: true,
			},
			contains: "reintroduce",
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
