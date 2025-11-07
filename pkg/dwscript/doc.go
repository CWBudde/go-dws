// Package dwscript provides a high-level API for embedding the DWScript interpreter
// in Go applications.
//
// DWScript is a full-featured Object Pascal-based scripting language, originally
// written in Delphi. This package provides 100% language compatibility while using
// idiomatic Go patterns.
//
// # Basic Usage
//
// The simplest way to use dwscript is with the Eval method:
//
//	engine, err := dwscript.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	result, err := engine.Eval(`
//	    var x: Integer := 42;
//	    PrintLn('The answer is ' + IntToStr(x));
//	`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Output) // "The answer is 42"
//
// # Compiling and Running
//
// For better performance when running the same script multiple times,
// compile once and run many times:
//
//	engine, _ := dwscript.New()
//
//	// Compile once
//	program, err := engine.Compile(`
//	    function Fibonacci(n: Integer): Integer;
//	    begin
//	        if n <= 1 then
//	            Result := n
//	        else
//	            Result := Fibonacci(n-1) + Fibonacci(n-2);
//	    end;
//	    PrintLn(IntToStr(Fibonacci(10)));
//	`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Run many times
//	for i := 0; i < 10; i++ {
//	    result, _ := program.Run()
//	    fmt.Println(result.Output)
//	}
//
// # Structured Errors
//
// The package provides structured error information with precise position data,
// perfect for IDE integration and error reporting:
//
//	engine, _ := dwscript.New()
//	_, err := engine.Compile(`
//	    var x: Integer := "not a number"; // Type error
//	`)
//
//	if compileErr, ok := err.(*dwscript.CompileError); ok {
//	    for _, e := range compileErr.Errors {
//	        fmt.Printf("Error at line %d, column %d: %s\n",
//	            e.Line, e.Column, e.Message)
//	        fmt.Printf("  Severity: %s, Code: %s\n", e.Severity, e.Code)
//	    }
//	}
//
// All positions use 1-based line and column numbering, matching most editors
// and IDEs. The Length field indicates the span of the error in characters.
//
// # AST Access
//
// Access the Abstract Syntax Tree for advanced use cases like code analysis,
// refactoring tools, or custom linters:
//
//	program, _ := engine.Compile(`
//	    var x: Integer := 42;
//	    if x > 0 then
//	        PrintLn('positive');
//	`)
//
//	// Get the AST
//	tree := program.AST()
//
//	// Traverse using the visitor pattern
//	ast.Inspect(tree, func(node ast.Node) bool {
//	    if fn, ok := node.(*ast.FunctionDecl); ok {
//	        fmt.Printf("Found function: %s at line %d\n",
//	            fn.Name.Value, fn.Pos().Line)
//	    }
//	    return true // continue traversal
//	})
//
// All AST nodes include both Pos() and End() methods that return precise
// position information for the entire node span.
//
// # Symbol Information
//
// Extract symbol table information for features like autocomplete, go-to-definition,
// and hover information:
//
//	program, _ := engine.Compile(`
//	    var count: Integer := 0;
//
//	    function Increment(): Integer;
//	    begin
//	        count := count + 1;
//	        Result := count;
//	    end;
//	`)
//
//	// Get all symbols
//	symbols := program.Symbols()
//	for _, sym := range symbols {
//	    fmt.Printf("%s (%s): %s at line %d\n",
//	        sym.Name, sym.Kind, sym.Type, sym.Position.Line)
//	}
//
// # Type Information
//
// Query type information at specific positions in the code:
//
//	program, _ := engine.Compile(`
//	    var x: Integer := 42;
//	    var y := x + 10;
//	`)
//
//	// Get type at position (line 2, column 5)
//	pos := token.Position{Line: 2, Column: 5}
//	if typeStr, ok := program.TypeAt(pos); ok {
//	    fmt.Printf("Type at position: %s\n", typeStr) // "Integer"
//	}
//
// # Parse-Only Mode
//
// For LSP servers and IDEs that need fast syntax checking without full
// type checking, use the Parse method:
//
//	engine, _ := dwscript.New()
//
//	// Parse without type checking (faster)
//	tree, err := engine.Parse(`
//	    var x: Integer := 42;
//	    if x > 0 then
//	        PrintLn('positive');
//	`)
//
//	if err != nil {
//	    // Only syntax errors, no type errors
//	    if compileErr, ok := err.(*dwscript.CompileError); ok {
//	        for _, e := range compileErr.Errors {
//	            fmt.Printf("Syntax error: %s\n", e.Message)
//	        }
//	    }
//	}
//
//	// AST is still available even if there were errors
//	if tree != nil {
//	    // Use for syntax highlighting, outline view, etc.
//	}
//
// # LSP Integration
//
// This package is designed to support Language Server Protocol (LSP) implementations.
// For a complete LSP server implementation, see: https://github.com/cwbudde/go-dws-lsp
//
// Key features for LSP support:
//   - Structured errors with precise position information
//   - Fast parse-only mode (Parse method)
//   - Complete AST access with position metadata
//   - Symbol table extraction
//   - Type information at position
//   - Visitor pattern for AST traversal
//
// # Configuration Options
//
// Configure the engine with functional options:
//
//	engine, _ := dwscript.New(
//	    dwscript.WithMaxRecursionDepth(2048),
//	    dwscript.WithOutput(os.Stdout),
//	    dwscript.WithTypeCheck(true), // Enable type checking
//	    dwscript.WithCompileMode(dwscript.CompileModeBytecode), // Use bytecode VM (experimental)
//	)
//
// # Foreign Function Interface (FFI)
//
// Register Go functions to be called from DWScript:
//
//	engine, _ := dwscript.New()
//
//	engine.RegisterFunction("GoAdd", func(a, b int64) int64 {
//	    return a + b
//	})
//
//	result, _ := engine.Eval(`
//	    var x := GoAdd(10, 32);
//	    PrintLn(IntToStr(x));
//	`)
//
// # Position Coordinate System
//
// All position information uses 1-based indexing for both lines and columns:
//   - Line numbers start at 1 (not 0)
//   - Column numbers start at 1 (not 0)
//   - This matches most text editors and IDEs
//
// Example positions for the string "var x := 42;":
//   - 'v' is at Line: 1, Column: 1
//   - 'x' is at Line: 1, Column: 5
//   - '4' is at Line: 1, Column: 10
//
// # Error Codes and Severity
//
// Structured errors include severity levels:
//   - SeverityError: Compilation errors that prevent execution
//   - SeverityWarning: Warnings that don't prevent execution
//   - SeverityInfo: Informational messages
//   - SeverityHint: Suggestions for code improvement
//
// Common error codes:
//   - "E001": Syntax error
//   - "E002": Type mismatch
//   - "E003": Undefined variable
//   - "W001": Unused variable
//   - "W002": Deprecated feature
//
// # Thread Safety
//
// Engine instances are safe for concurrent use. However, Program and Result
// instances are not thread-safe and should not be shared across goroutines
// without external synchronization.
//
// # Compatibility
//
// This implementation aims for 100% compatibility with the original DWScript
// language specification. See https://www.delphitools.info/dwscript/ for
// the complete language reference.
//
// Minimum Go version: 1.21
package dwscript
