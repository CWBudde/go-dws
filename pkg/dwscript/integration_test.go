// Copyright (c) 2024 MeKo-Tech
// SPDX-License-Identifier: MIT

package dwscript

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestIntegration_ParseASTSymbols tests the complete workflow: Parse → AST → Symbols (Task 10.22)
func TestIntegration_ParseASTSymbols(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	source := `
		var x: Integer := 42;
		var message: String := 'Hello';
		const PI = 3.14;

		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		function Greet(name: String): String;
		begin
			Result := 'Hello, ' + name;
		end;
	`

	// Step 1: Compile to get full Program
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Step 2: Get AST
	tree := program.AST()
	if tree == nil {
		t.Fatal("Program.AST() returned nil")
	}

	// Verify AST structure
	if len(tree.Statements) == 0 {
		t.Fatal("Expected statements in AST")
	}

	// Step 3: Get Symbols
	symbols := program.Symbols()
	if len(symbols) == 0 {
		t.Fatal("Expected symbols")
	}

	// Verify we have expected symbols
	symbolNames := make(map[string]bool)
	for _, sym := range symbols {
		symbolNames[sym.Name] = true
	}

	expectedSymbols := []string{"x", "message", "PI", "Add", "Greet"}
	for _, expected := range expectedSymbols {
		if !symbolNames[expected] {
			t.Errorf("Expected symbol %q not found", expected)
		}
	}

	// Verify symbol details
	for _, sym := range symbols {
		// Note: Position information is not currently stored in symbol table
		// (this is a known limitation - positions are available on AST nodes)

		// All symbols should have a kind
		if sym.Kind == "" {
			t.Errorf("Symbol %q has empty kind", sym.Name)
		}

		// All symbols should have a type
		if sym.Type == "" {
			t.Errorf("Symbol %q has empty type", sym.Name)
		}
	}
}

// TestIntegration_ErrorRecovery tests error recovery and partial results (Task 10.22)
func TestIntegration_ErrorRecovery(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Code with both syntax and semantic errors
	source := `
		var x: Integer := 42;
		var y String;  // Syntax error: missing colon
		var z: Integer := x + 10;  // Valid after error

		function Test(): Integer;
		begin
			Result := undefined_var;  // Semantic error: undefined variable
		end;
	`

	// Use Parse() for syntax-only checking
	tree, parseErr := engine.Parse(source)

	// Should get partial AST even with errors
	if tree == nil {
		t.Fatal("Parse() should return partial AST even with syntax errors")
	}

	// Should report syntax error
	if parseErr == nil {
		t.Error("Parse() should report syntax error")
	}

	if compileErr, ok := parseErr.(*CompileError); ok {
		if compileErr.Stage != "parsing" {
			t.Errorf("Expected parsing stage, got %q", compileErr.Stage)
		}

		// Should have at least one error
		if len(compileErr.Errors) == 0 {
			t.Error("Expected at least one syntax error")
		}

		// Check error has position info
		for i, e := range compileErr.Errors {
			if e.Line == 0 && e.Column == 0 {
				t.Errorf("Error %d missing position information", i)
			}
			if e.Message == "" {
				t.Errorf("Error %d missing message", i)
			}
		}
	}

	// Now test Compile() for semantic errors
	_, compileErr := engine.Compile(source)
	if compileErr == nil {
		t.Error("Compile() should report errors")
	}

	// Should have structured error information
	if ce, ok := compileErr.(*CompileError); ok {
		if len(ce.Errors) == 0 {
			t.Error("Expected errors in CompileError")
		}

		// Verify errors have position and message
		for i, e := range ce.Errors {
			if e.Message == "" {
				t.Errorf("Error %d has no message", i)
			}
			// Note: Position might be 0:0 for some errors, that's OK
		}
	}
}

// TestIntegration_PositionMapping tests accurate position tracking (Task 10.22)
func TestIntegration_PositionMapping(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	source := `var x: Integer := 42;
var y: String := 'hello';

function Add(a, b: Integer): Integer;
begin
  Result := a + b;
end;`

	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	tree := program.AST()

	// Test position accuracy for each statement
	tests := []struct {
		stmtIndex int
		wantLine  int
		desc      string
	}{
		{0, 1, "first variable declaration"},
		{1, 2, "second variable declaration"},
		{2, 4, "function declaration"},
	}

	for _, tt := range tests {
		if tt.stmtIndex >= len(tree.Statements) {
			t.Errorf("Statement index %d out of bounds (have %d statements)", tt.stmtIndex, len(tree.Statements))
			continue
		}

		stmt := tree.Statements[tt.stmtIndex]
		pos := stmt.Pos()

		if pos.Line != tt.wantLine {
			t.Errorf("%s: got line %d, want line %d", tt.desc, pos.Line, tt.wantLine)
		}

		// Verify End() is after or equal to Pos()
		end := stmt.End()
		if end.Line < pos.Line {
			t.Errorf("%s: End().Line (%d) < Pos().Line (%d)", tt.desc, end.Line, pos.Line)
		}
		if end.Line == pos.Line && end.Column < pos.Column {
			t.Errorf("%s: End().Column (%d) < Pos().Column (%d) on same line", tt.desc, end.Column, pos.Column)
		}
	}

	// Verify symbols are extracted correctly
	symbols := program.Symbols()
	if len(symbols) == 0 {
		t.Error("Expected symbols to be extracted")
	}

	// Find user-defined symbols
	userSymbols := map[string]bool{"x": false, "y": false, "Add": false}
	for _, sym := range symbols {
		if _, exists := userSymbols[sym.Name]; exists {
			userSymbols[sym.Name] = true
		}
	}

	// Verify all expected symbols were found
	for name, found := range userSymbols {
		if !found {
			t.Errorf("Expected symbol %q not found", name)
		}
	}
}

// TestIntegration_RealCodeSample tests with actual DWScript code (Task 10.22)
func TestIntegration_RealCodeSample(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Real DWScript code sample
	source := `
		// Fibonacci calculator
		function Fibonacci(n: Integer): Integer;
		begin
			if n <= 1 then
				Result := n
			else
				Result := Fibonacci(n - 1) + Fibonacci(n - 2);
		end;

		// Main program
		var i: Integer;
		for i := 0 to 10 do
			PrintLn(IntToStr(i) + ': ' + IntToStr(Fibonacci(i)));
	`

	// Parse and compile
	program, err := engine.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Get AST
	tree := program.AST()
	if tree == nil {
		t.Fatal("AST is nil")
	}

	// Count node types using visitor
	counts := make(map[string]int)
	ast.Inspect(tree, func(node ast.Node) bool {
		if node != nil {
			switch node.(type) {
			case *ast.Program:
				counts["Program"]++
			case *ast.FunctionDecl:
				counts["FunctionDecl"]++
			case *ast.IfStatement:
				counts["IfStatement"]++
			case *ast.ForStatement:
				counts["ForStatement"]++
			case *ast.VarDeclStatement:
				counts["VarDecl"]++
			case *ast.CallExpression:
				counts["CallExpression"]++
			case *ast.BinaryExpression:
				counts["BinaryExpression"]++
			}
		}
		return true
	})

	// Verify expected structure
	if counts["Program"] != 1 {
		t.Errorf("Expected 1 Program, got %d", counts["Program"])
	}
	if counts["FunctionDecl"] < 1 {
		t.Error("Expected at least 1 FunctionDecl")
	}
	if counts["IfStatement"] < 1 {
		t.Error("Expected at least 1 IfStatement")
	}

	// Get symbols
	symbols := program.Symbols()

	// Should have Fibonacci function
	foundFib := false
	for _, sym := range symbols {
		if sym.Name == "Fibonacci" {
			foundFib = true
			if sym.Kind != "function" {
				t.Errorf("Fibonacci kind = %q, want %q", sym.Kind, "function")
			}
			if !strings.Contains(sym.Type, "Integer") {
				t.Errorf("Fibonacci type = %q, should contain 'Integer'", sym.Type)
			}
		}
	}

	if !foundFib {
		t.Error("Fibonacci function not found in symbols")
	}
}

// TestIntegration_NoRegressions tests that existing functionality still works (Task 10.22)
func TestIntegration_NoRegressions(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test basic compilation still works
	program, err := engine.Compile("var x: Integer := 42;")
	if err != nil {
		t.Errorf("Basic compilation failed: %v", err)
	}

	// Test that compiled program provides AST and symbols
	tree := program.AST()
	if tree == nil {
		t.Error("Program.AST() returned nil")
	}

	symbols := program.Symbols()
	if len(symbols) == 0 {
		t.Error("Program.Symbols() returned no symbols")
	}

	// Test execution still works via Eval
	result, err := engine.Eval("var x: Integer := 42;")
	if err != nil {
		t.Errorf("Eval failed: %v", err)
	}

	// Test result can be retrieved
	if result == nil {
		t.Error("Result is nil")
	}

	// Test various language features
	tests := []struct {
		name   string
		source string
	}{
		{"variables", "var x: Integer := 10; var y := x + 5;"},
		{"strings", "var s: String := 'hello world';"},
		{"booleans", "var b: Boolean := true;"},
		{"functions", "function Test(): Integer; begin Result := 42; end;"},
		{"if statements", "var x := 5; if x > 0 then x := 10;"},
		{"while loops", "var i := 0; while i < 5 do i := i + 1;"},
		{"for loops", "var i: Integer; for i := 0 to 10 do PrintLn(IntToStr(i));"},
		{"arrays", "var arr: array of Integer; arr := [1, 2, 3];"},
		{"constants", "const PI = 3.14; var r := 2.0; var area := PI * r * r;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Compile(tt.source)
			if err != nil {
				t.Errorf("Compilation failed for %s: %v", tt.name, err)
			}
		})
	}
}

// TestIntegration_LSPWorkflow tests typical LSP usage patterns (Task 10.22)
func TestIntegration_LSPWorkflow(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	source := `
		var x: Integer := 42;
		var message: String := 'Hello, World!';

		function Double(n: Integer): Integer;
		begin
			Result := n * 2;
		end;
	`

	// Typical LSP workflow:

	// 1. Parse for syntax highlighting (fast, no type checking)
	tree, parseErr := engine.Parse(source)
	if tree == nil {
		t.Fatal("Parse() returned nil AST")
	}
	if parseErr != nil {
		t.Errorf("Parse() returned unexpected error: %v", parseErr)
	}

	// 2. Get AST for code structure (outline view, breadcrumbs)
	if len(tree.Statements) == 0 {
		t.Error("Expected statements in AST")
	}

	// Use visitor to find declarations for outline
	var functions []string
	var variables []string
	ast.Inspect(tree, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FunctionDecl:
			if n.Name != nil {
				functions = append(functions, n.Name.Value)
			}
		case *ast.VarDeclStatement:
			for _, name := range n.Names {
				variables = append(variables, name.Value)
			}
		}
		return true
	})

	if len(functions) != 1 || functions[0] != "Double" {
		t.Errorf("Expected [Double], got %v", functions)
	}
	if len(variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(variables))
	}

	// 3. Full compile for diagnostics
	program, compileErr := engine.Compile(source)
	if compileErr != nil {
		t.Errorf("Compile() failed: %v", compileErr)
	}

	// 4. Get symbols for autocomplete
	symbols := program.Symbols()
	if len(symbols) == 0 {
		t.Error("Expected symbols for autocomplete")
	}

	// Verify symbols are usable for autocomplete
	for _, sym := range symbols {
		// Each symbol should have name, type, and kind
		if sym.Name == "" {
			t.Errorf("Symbol has empty name")
		}
		if sym.Type == "" {
			t.Errorf("Symbol %q has empty type", sym.Name)
		}
		if sym.Kind == "" {
			t.Errorf("Symbol %q has empty kind", sym.Name)
		}
	}

	// Verify expected symbols are present
	expectedSyms := map[string]bool{"x": false, "message": false, "Double": false}
	for _, sym := range symbols {
		if _, exists := expectedSyms[sym.Name]; exists {
			expectedSyms[sym.Name] = true
		}
	}

	for name, found := range expectedSyms {
		if !found {
			t.Errorf("Expected symbol %q not found in autocomplete list", name)
		}
	}
}

// TestIntegration_ErrorPositions tests that all errors have accurate positions (Task 10.22)
func TestIntegration_ErrorPositions(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Code with error on specific line
	source := `var x: Integer := 42;
var y: Integer := x + 10;
var z: Integer := undefined_variable;  // Error on line 3
var w: Integer := 100;`

	_, err = engine.Compile(source)
	if err == nil {
		t.Fatal("Expected compilation error for undefined variable")
	}

	compileErr, ok := err.(*CompileError)
	if !ok {
		t.Fatalf("Expected *CompileError, got %T", err)
	}

	// Should have at least one error
	if len(compileErr.Errors) == 0 {
		t.Fatal("Expected at least one error")
	}

	// Find the undefined variable error
	foundError := false
	for _, e := range compileErr.Errors {
		if strings.Contains(e.Message, "undefined") || strings.Contains(e.Message, "Undefined") {
			foundError = true

			// Should be on line 3
			if e.Line != 3 {
				t.Errorf("Error line = %d, want 3", e.Line)
			}

			// Should have column information
			if e.Column == 0 {
				t.Error("Error missing column information")
			}

			t.Logf("Error at %d:%d: %s", e.Line, e.Column, e.Message)
		}
	}

	if !foundError {
		t.Error("Did not find undefined variable error")
	}
}

// TestIntegration_MultipleErrors tests handling of multiple errors (Task 10.22)
func TestIntegration_MultipleErrors(t *testing.T) {
	engine, err := New(WithTypeCheck(true))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Code with multiple errors
	source := `
		var x: Integer := 'string';  // Type error line 2
		var y: String := 42;          // Type error line 3
		var z: Integer := undefined;  // Undefined error line 4
	`

	_, err = engine.Compile(source)
	if err == nil {
		t.Fatal("Expected compilation errors")
	}

	compileErr, ok := err.(*CompileError)
	if !ok {
		t.Fatalf("Expected *CompileError, got %T", err)
	}

	// Should report multiple errors
	if len(compileErr.Errors) < 2 {
		t.Errorf("Expected at least 2 errors, got %d", len(compileErr.Errors))
	}

	// All errors should have position information
	for i, e := range compileErr.Errors {
		if e.Line == 0 {
			t.Errorf("Error %d missing line number", i)
		}
		if e.Message == "" {
			t.Errorf("Error %d missing message", i)
		}

		t.Logf("Error %d at %d:%d: %s", i+1, e.Line, e.Column, e.Message)
	}
}
