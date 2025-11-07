package ast_test

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// Example demonstrates basic AST traversal
func Example() {
	// Create an engine and compile a simple program
	engine, _ := dwscript.New()
	program, _ := engine.Compile("var x: Integer := 42;")

	// Get the AST
	tree := program.AST()

	// Walk through statements
	for _, stmt := range tree.Statements {
		fmt.Printf("Statement type: %T\n", stmt)
	}

	// Output:
	// Statement type: *ast.VarDeclStatement
}

// Example_functionDeclaration demonstrates finding function declarations in the AST
func Example_functionDeclaration() {
	engine, _ := dwscript.New()
	program, _ := engine.Compile(`
		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;
	`)

	tree := program.AST()

	// Find all function declarations
	for _, stmt := range tree.Statements {
		if funcDecl, ok := stmt.(*ast.FunctionDecl); ok {
			fmt.Printf("Function: %s\n", funcDecl.Name.Value)
			fmt.Printf("Parameters: %d\n", len(funcDecl.Parameters))
			if funcDecl.ReturnType != nil {
				fmt.Printf("Return type: %s\n", funcDecl.ReturnType.Name)
			}
		}
	}

	// Output:
	// Function: Add
	// Parameters: 2
	// Return type: Integer
}

// Example_walkingAST demonstrates walking the AST to find specific node types
func Example_walkingAST() {
	engine, _ := dwscript.New()
	program, _ := engine.Compile(`
		var x: Integer;
		var y: String;
		begin
			x := 10;
			y := 'hello';
		end
	`)

	tree := program.AST()

	// Count variable declarations
	varCount := 0
	for _, stmt := range tree.Statements {
		if _, ok := stmt.(*ast.VarDeclStatement); ok {
			varCount++
		}
	}

	fmt.Printf("Variable declarations: %d\n", varCount)

	// Output:
	// Variable declarations: 2
}

// Example_positionInformation demonstrates accessing source position information
func Example_positionInformation() {
	engine, _ := dwscript.New()
	program, _ := engine.Compile("var x: Integer := 42;")

	tree := program.AST()

	for _, stmt := range tree.Statements {
		pos := stmt.Pos()
		end := stmt.End()
		fmt.Printf("Statement at line %d, col %d to line %d, col %d\n",
			pos.Line, pos.Column, end.Line, end.Column)
	}

	// Output:
	// Statement at line 1, col 1 to line 1, col 22
}
