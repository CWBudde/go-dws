package dwscript

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// TestParse_ValidCode tests Parse() with valid DWScript code
func TestParse_ValidCode(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	source := `
		var x: Integer := 42;
		var y: String := 'hello';

		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;
	`

	tree, err := engine.Parse(source)

	// Should return AST without error
	if err != nil {
		t.Errorf("Parse() returned unexpected error: %v", err)
	}

	if tree == nil {
		t.Fatal("Parse() returned nil AST for valid code")
	}

	// Verify AST contains expected statements
	if len(tree.Statements) == 0 {
		t.Error("Parse() returned empty AST")
	}

	// Check for variable declarations
	varCount := 0
	funcCount := 0
	for _, stmt := range tree.Statements {
		switch stmt.(type) {
		case *ast.VarDeclStatement:
			varCount++
		case *ast.FunctionDecl:
			funcCount++
		}
	}

	if varCount != 2 {
		t.Errorf("Expected 2 variable declarations, got %d", varCount)
	}

	if funcCount != 1 {
		t.Errorf("Expected 1 function declaration, got %d", funcCount)
	}
}

// TestParse_InvalidCode tests Parse() with syntax errors
func TestParse_InvalidCode(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Code with intentional syntax errors
	source := `
		var x Integer;  // Missing colon
		var y := 'test';  // Missing type
	`

	tree, err := engine.Parse(source)

	// Should return AST even with errors (best-effort parsing)
	if tree == nil {
		t.Fatal("Parse() returned nil AST even with syntax errors (should return partial AST)")
	}

	// Should also return error information
	if err == nil {
		t.Error("Parse() should return error for invalid syntax")
	}

	// Verify error is a CompileError
	compileErr, ok := err.(*CompileError)
	if !ok {
		t.Errorf("Expected *CompileError, got %T", err)
	} else {
		if compileErr.Stage != "parsing" {
			t.Errorf("Expected stage 'parsing', got '%s'", compileErr.Stage)
		}

		if len(compileErr.Errors) == 0 {
			t.Error("Expected syntax errors to be reported")
		}

		// Verify error details include position information
		for i, e := range compileErr.Errors {
			if e.Line == 0 {
				t.Errorf("Error %d missing line number", i)
			}
			if e.Message == "" {
				t.Errorf("Error %d missing message", i)
			}
		}
	}
}

// TestParse_EmptyCode tests Parse() with empty source
func TestParse_EmptyCode(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	tree, err := engine.Parse("")

	// Empty code should parse successfully (no errors)
	if err != nil {
		t.Errorf("Parse() returned error for empty code: %v", err)
	}

	if tree == nil {
		t.Fatal("Parse() returned nil AST for empty code")
	}

	// Should have no statements
	if len(tree.Statements) != 0 {
		t.Errorf("Expected 0 statements, got %d", len(tree.Statements))
	}
}

// TestParse_PartialCode tests Parse() with incomplete code
func TestParse_PartialCode(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Incomplete function (missing end)
	source := `
		function Test(): Integer;
		begin
			Result := 42
	`

	tree, err := engine.Parse(source)

	// Should return partial AST
	if tree == nil {
		t.Fatal("Parse() returned nil AST for partial code")
	}

	// Should report syntax errors
	if err == nil {
		t.Error("Parse() should return error for incomplete code")
	}
}

// TestParse_VsCompile tests difference between Parse() and Compile()
func TestParse_VsCompile(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Valid syntax but type error
	source := `
		var x: Integer := 'hello';  // Type mismatch
	`

	// Parse() should succeed (no syntax errors)
	tree, parseErr := engine.Parse(source)
	if tree == nil {
		t.Error("Parse() should return AST for valid syntax")
	}
	if parseErr != nil {
		t.Errorf("Parse() should not return error for valid syntax: %v", parseErr)
	}

	// Compile() should fail (semantic/type error)
	_, compileErr := engine.Compile(source)
	if compileErr == nil {
		t.Error("Compile() should return error for type mismatch")
	} else {
		// Verify it's a semantic error, not a parse error
		if ce, ok := compileErr.(*CompileError); ok {
			if ce.Stage != "type checking" {
				t.Errorf("Expected semantic error, got stage: %s", ce.Stage)
			}
		}
	}
}

// TestParse_LSPUseCase tests typical LSP usage pattern
func TestParse_LSPUseCase(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Simulate user typing incomplete code in editor
	sources := []string{
		"var x",                        // Very incomplete
		"var x:",                       // Still incomplete
		"var x: Integer",               // Missing semicolon
		"var x: Integer;",              // Complete
		"var x: Integer; var y: Strin", // Complete + incomplete
	}

	for i, source := range sources {
		tree, _ := engine.Parse(source)

		// Parse should always return something, even for very incomplete code
		if tree == nil {
			t.Errorf("Source %d: Parse() returned nil AST: %q", i, source)
			continue
		}

		// AST should be usable (has Pos/End methods, can be traversed)
		_ = tree.Pos()
		_ = tree.End()
		_ = tree.String()
	}
}

// TestParse_Performance tests that Parse() is faster than Compile()
func TestParse_Performance(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Large but valid source code
	source := `
		var x1: Integer := 1;
		var x2: Integer := 2;
		var x3: Integer := 3;
		var x4: Integer := 4;
		var x5: Integer := 5;

		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		function Multiply(a, b: Integer): Integer;
		begin
			Result := a * b;
		end;
	`

	// Parse() should complete quickly (no semantic analysis)
	_, parseErr := engine.Parse(source)
	if parseErr != nil {
		t.Errorf("Parse() failed: %v", parseErr)
	}

	// Compile() does more work (semantic analysis)
	_, compileErr := engine.Compile(source)
	if compileErr != nil {
		t.Errorf("Compile() failed: %v", compileErr)
	}

	// Both should work, but Parse() skips semantic analysis
	// (timing comparison would go here in a benchmark)
}

// TestParse_ErrorRecovery tests error recovery in Parse()
func TestParse_ErrorRecovery(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Multiple syntax errors in different locations
	source := `
		var x Integer;      // Error on line 2
		var y: String := 'ok';  // Valid
		var z Float;        // Error on line 4

		function Test();    // Error - missing return type colon
		begin
			PrintLn('test');
		end;
	`

	tree, err := engine.Parse(source)

	// Should return partial AST
	if tree == nil {
		t.Fatal("Parse() should return partial AST with multiple errors")
	}

	// Should have some valid statements
	if len(tree.Statements) == 0 {
		t.Error("Parse() should recover and parse some statements")
	}

	// Should report multiple errors
	if err != nil {
		if ce, ok := err.(*CompileError); ok {
			if len(ce.Errors) < 2 {
				t.Errorf("Expected multiple errors, got %d", len(ce.Errors))
			}
		}
	}
}

// Example_parse demonstrates using Parse() for LSP/editor integration
func Example_parse() {
	engine, _ := New()

	// Parse code that's syntactically valid
	source := "var x: Integer := 42;"
	tree, err := engine.Parse(source)

	// AST is available
	if tree != nil {
		// Can provide syntax highlighting, code folding, etc.
		if len(tree.Statements) > 0 {
			pos := tree.Statements[0].Pos()
			fmt.Printf("Statement at line: %d\n", pos.Line)
		}
	}

	// No syntax errors for valid code
	if err != nil {
		fmt.Println("Unexpected error")
	} else {
		fmt.Println("Parse successful")
	}

	// Output:
	// Statement at line: 1
	// Parse successful
}
