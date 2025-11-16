package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestTypeExpressionEndPos tests that EndPos is correctly set for various type expressions.
func TestTypeExpressionEndPos(t *testing.T) {
	tests := []struct {
		checkPos func(*testing.T, ast.Statement)
		name     string
		input    string
	}{
		{
			name:  "TypeAnnotation simple",
			input: "var x: Integer;",
			checkPos: func(t *testing.T, stmt ast.Statement) {
				varStmt, ok := stmt.(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("expected *ast.VarDeclStatement, got %T", stmt)
				}
				if varStmt.Type == nil {
					t.Fatal("varStmt.Type is nil")
				}
				// TypeAnnotation should end after "Integer"
				// "var x: Integer;" - Integer ends at column 15
				if varStmt.Type.End().Column != 15 {
					t.Errorf("expected Type.End().Column = 15, got %d", varStmt.Type.End().Column)
				}
			},
		},
		{
			name:  "ArrayType dynamic",
			input: "var arr: array of Integer;",
			checkPos: func(t *testing.T, stmt ast.Statement) {
				varStmt, ok := stmt.(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("expected *ast.VarDeclStatement, got %T", stmt)
				}
				if varStmt.Type == nil {
					t.Fatal("varStmt.Type is nil")
				}
				arrayType, ok := varStmt.Type.(*ast.ArrayTypeNode)
				if !ok {
					t.Fatalf("expected *ast.ArrayTypeNode, got %T", varStmt.Type)
				}
				// "var arr: array of Integer;" - Integer ends at column 26
				if arrayType.End().Column != 26 {
					t.Errorf("expected ArrayType.End().Column = 26, got %d", arrayType.End().Column)
				}
			},
		},
		{
			name:  "SetType",
			input: "var s: set of TEnum;",
			checkPos: func(t *testing.T, stmt ast.Statement) {
				varStmt, ok := stmt.(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("expected *ast.VarDeclStatement, got %T", stmt)
				}
				if varStmt.Type == nil {
					t.Fatal("varStmt.Type is nil")
				}
				setType, ok := varStmt.Type.(*ast.SetTypeNode)
				if !ok {
					t.Fatalf("expected *ast.SetTypeNode, got %T", varStmt.Type)
				}
				// "var s: set of TEnum;" - TEnum ends at column 20
				if setType.End().Column != 20 {
					t.Errorf("expected SetType.End().Column = 20, got %d", setType.End().Column)
				}
			},
		},
		{
			name:  "StaticArrayType",
			input: "var a: array[1..10] of String;",
			checkPos: func(t *testing.T, stmt ast.Statement) {
				varStmt, ok := stmt.(*ast.VarDeclStatement)
				if !ok {
					t.Fatalf("expected *ast.VarDeclStatement, got %T", stmt)
				}
				if varStmt.Type == nil {
					t.Fatal("varStmt.Type is nil")
				}
				arrayType, ok := varStmt.Type.(*ast.ArrayTypeNode)
				if !ok {
					t.Fatalf("expected *ast.ArrayTypeNode, got %T", varStmt.Type)
				}
				// "var a: array[1..10] of String;" - String ends at column 30
				if arrayType.End().Column != 30 {
					t.Errorf("expected StaticArray.End().Column = 30, got %d", arrayType.End().Column)
				}
			},
		},
		{
			name:  "FunctionPointerType in type declaration",
			input: "type TFunc = function(x: Integer): Boolean;",
			checkPos: func(t *testing.T, stmt ast.Statement) {
				typeDecl, ok := stmt.(*ast.TypeDeclaration)
				if !ok {
					t.Fatalf("expected *ast.TypeDeclaration, got %T", stmt)
				}
				if typeDecl.FunctionPointerType == nil {
					t.Fatal("typeDecl.FunctionPointerType is nil")
				}
				// "type TFunc = function(x: Integer): Boolean;"
				// Col: 123456789012345678901234567890123456789012345
				//      type TFunc = function(x: Integer): Boolean;
				// Boolean is at columns 36-42, ends at 43
				if typeDecl.FunctionPointerType.End().Column != 43 {
					t.Errorf("expected FunctionPointer.End().Column = 43, got %d", typeDecl.FunctionPointerType.End().Column)
				}
			},
		},
		{
			name:  "ProcedurePointerType in type declaration",
			input: "type TProc = procedure(msg: String);",
			checkPos: func(t *testing.T, stmt ast.Statement) {
				typeDecl, ok := stmt.(*ast.TypeDeclaration)
				if !ok {
					t.Fatalf("expected *ast.TypeDeclaration, got %T", stmt)
				}
				if typeDecl.FunctionPointerType == nil {
					t.Fatal("typeDecl.FunctionPointerType is nil")
				}
				// "type TProc = procedure(msg: String);"
				// Col: 12345678901234567890123456789012345678
				//      type TProc = procedure(msg: String);
				// ) is at column 35, EndPos is after it at column 36
				if typeDecl.FunctionPointerType.End().Column != 36 {
					t.Errorf("expected ProcedurePointer.End().Column = 36, got %d", typeDecl.FunctionPointerType.End().Column)
				}
			},
		},
		{
			name:  "MethodPointerType in type declaration",
			input: "type TMethod = procedure(x: Integer) of object;",
			checkPos: func(t *testing.T, stmt ast.Statement) {
				typeDecl, ok := stmt.(*ast.TypeDeclaration)
				if !ok {
					t.Fatalf("expected *ast.TypeDeclaration, got %T", stmt)
				}
				if typeDecl.FunctionPointerType == nil {
					t.Fatal("typeDecl.FunctionPointerType is nil")
				}
				// "type TMethod = procedure(x: Integer) of object;"
				// Col: 123456789012345678901234567890123456789012345678
				//      type TMethod = procedure(x: Integer) of object;
				// object ends at column 46, EndPos is after at column 47
				if typeDecl.FunctionPointerType.End().Column != 47 {
					t.Errorf("expected MethodPointer.End().Column = 47, got %d", typeDecl.FunctionPointerType.End().Column)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			if len(program.Statements) == 0 {
				t.Fatal("no statements parsed")
			}

			tt.checkPos(t, program.Statements[0])
		})
	}
}
