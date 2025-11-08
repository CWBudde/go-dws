// Package dwscript provides a high-level API for embedding the DWScript interpreter
// in Go applications.
//
// Example usage:
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
package dwscript

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/cwbudde/go-dws/internal/bytecode"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Engine is the main entry point for the DWScript interpreter.
// It provides a high-level API for compiling and executing DWScript programs.
type Engine struct {
	options           Options
	externalFunctions *interp.ExternalFunctionRegistry
}

// New creates a new DWScript engine with the given options.
// If no options are provided, sensible defaults are used.
func New(opts ...Option) (*Engine, error) {
	engine := &Engine{
		options:           defaultOptions(),
		externalFunctions: interp.NewExternalFunctionRegistry(),
	}

	for _, opt := range opts {
		if err := opt(&engine.options); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return engine, nil
}

// Compile parses and type-checks the given DWScript source code,
// returning a compiled Program that can be executed multiple times.
//
// This is useful when you want to compile once and run many times,
// as it avoids re-parsing and re-checking the source code.
func (e *Engine) Compile(source string) (*Program, error) {
	// Tokenize
	l := lexer.New(source)

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		// Convert parser errors to public Error type
		errors := make([]*Error, 0, len(p.Errors()))
		for _, perr := range p.Errors() {
			errors = append(errors, &Error{
				Message:  perr.Message,
				Line:     perr.Pos.Line,
				Column:   perr.Pos.Column,
				Length:   perr.Length,
				Severity: SeverityError,
				Code:     perr.Code,
			})
		}
		return nil, &CompileError{
			Stage:  "parsing",
			Errors: errors,
		}
	}

	// Type check (if enabled)
	var analyzer *semantic.Analyzer
	if e.options.TypeCheck {
		analyzer = semantic.NewAnalyzer()
		if err := analyzer.Analyze(program); err != nil {
			// Convert semantic errors
			errors := convertSemanticError(err)
			return nil, &CompileError{
				Stage:  "type checking",
				Errors: errors,
			}
		}
	}

	var chunk *bytecode.Chunk
	if e.options.CompileMode == CompileModeBytecode {
		bc := bytecode.NewCompiler("dwscript")
		var err error
		chunk, err = bc.Compile(program)
		if err != nil {
			return nil, newBytecodeCompileError(err)
		}
	}

	return &Program{
		ast:           program,
		analyzer:      analyzer,
		options:       e.options,
		bytecodeChunk: chunk,
	}, nil
}

// Parse parses the given DWScript source code and returns the AST without
// performing semantic analysis or type checking.
//
// This method is designed for use cases where you need the AST quickly,
// such as in Language Server Protocol (LSP) implementations, code formatters,
// syntax highlighters, or other editor tooling. It provides:
//
//   - Fast parsing without expensive semantic checks
//   - Best-effort AST construction (returns partial AST even with syntax errors)
//   - Structured syntax error information for diagnostics
//
// Unlike Compile(), Parse() will return a (potentially partial) AST even when
// syntax errors are present. This allows editors to provide features like
// syntax highlighting, code folding, and outline views even for invalid code.
//
// The returned AST should not be used for execution, as it has not been
// type-checked and may be incomplete. Use Compile() instead if you need to
// execute the code.
//
// Example usage in an LSP server:
//
//	engine, _ := dwscript.New()
//	tree, errs := engine.Parse(documentText)
//
//	// Tree is available even if there are errors
//	// Provide syntax highlighting based on the AST
//	highlightCode(tree)
//
//	// Report syntax errors to the editor
//	if errs != nil {
//	    if compileErr, ok := errs.(*dwscript.CompileError); ok {
//	        for _, err := range compileErr.Errors {
//	            reportDiagnostic(err.Line, err.Column, err.Message)
//	        }
//	    }
//	}
func (e *Engine) Parse(source string) (*ast.Program, error) {
	// Tokenize
	l := lexer.New(source)

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	// Always return the AST (even if there are errors)
	// This is the key difference from Compile() - best-effort parsing

	// If there are parse errors, convert them to structured errors
	if len(p.Errors()) > 0 {
		errors := make([]*Error, 0, len(p.Errors()))
		for _, perr := range p.Errors() {
			errors = append(errors, &Error{
				Message:  perr.Message,
				Line:     perr.Pos.Line,
				Column:   perr.Pos.Column,
				Length:   perr.Length,
				Severity: SeverityError,
				Code:     perr.Code,
			})
		}

		// Return both the partial AST and the errors
		// This allows editors to work with incomplete code
		return program, &CompileError{
			Stage:  "parsing",
			Errors: errors,
		}
	}

	// No errors - return the complete AST
	return program, nil
}

// convertSemanticError converts semantic analysis errors to structured Error objects.
// It extracts position information from error strings that contain " at line:column" format.
func convertSemanticError(err error) []*Error {
	// Check if it's already a semantic.AnalysisError
	if analysisErr, ok := err.(*semantic.AnalysisError); ok {
		errors := make([]*Error, 0, len(analysisErr.Errors))
		for _, errStr := range analysisErr.Errors {
			// Extract position from error string if present
			// Format: "error message at line:column"
			line, column, message := extractPositionFromError(errStr)

			errors = append(errors, &Error{
				Message:  message,
				Line:     line,
				Column:   column,
				Length:   0,
				Severity: SeverityError,
				Code:     "E_SEMANTIC",
			})
		}
		return errors
	}

	// Fallback: single error
	return []*Error{
		{
			Message:  err.Error(),
			Line:     0,
			Column:   0,
			Length:   0,
			Severity: SeverityError,
			Code:     "E_SEMANTIC",
		},
	}
}

func newBytecodeCompileError(err error) *CompileError {
	return &CompileError{
		Stage: "bytecode",
		Errors: []*Error{
			{
				Message:  err.Error(),
				Line:     0,
				Column:   0,
				Length:   0,
				Severity: SeverityError,
				Code:     "E_BYTECODE_COMPILE",
			},
		},
	}
}

// extractPositionFromError extracts position information from an error string.
// Returns (line, column, message) where line and column are 0 if not found.
// Handles error formats like:
//   - "error message at 10:5"
//   - "error message"
func extractPositionFromError(errStr string) (int, int, string) {
	// Look for " at line:column" pattern at the end of the string
	idx := strings.LastIndex(errStr, " at ")
	if idx == -1 {
		return 0, 0, errStr
	}

	// Extract the position part
	posPart := errStr[idx+4:] // Skip " at "
	message := errStr[:idx]

	// Try to parse "line:column"
	var line, column int
	n, err := fmt.Sscanf(posPart, "%d:%d", &line, &column)
	if err != nil || n != 2 {
		// Couldn't parse position, return original message
		return 0, 0, errStr
	}

	return line, column, message
}

// Run executes a previously compiled Program and returns the result.
func (e *Engine) Run(program *Program) (*Result, error) {
	if program == nil {
		return nil, fmt.Errorf("program is nil")
	}

	// Determine output writer
	output := e.options.Output
	if output == nil {
		output = &bytes.Buffer{}
	}

	if e.options.CompileMode == CompileModeBytecode {
		return e.runBytecode(program, output)
	}

	return e.runInterpreter(program, output)
}

func (e *Engine) runInterpreter(program *Program, output io.Writer) (*Result, error) {
	e.options.ExternalFunctions = e.externalFunctions
	interpreter := interp.NewWithOptions(output, &e.options)
	value := interpreter.Eval(program.ast)

	if value != nil && value.Type() == "ERROR" {
		return &Result{
				Output:  extractOutput(output),
				Success: false,
			}, &RuntimeError{
				Message: value.String(),
			}
	}

	return &Result{
		Output:  extractOutput(output),
		Success: true,
	}, nil
}

func (e *Engine) runBytecode(program *Program, output io.Writer) (*Result, error) {
	chunk, err := program.ensureBytecodeChunk()
	if err != nil {
		return nil, err
	}

	vm := bytecode.NewVMWithOutput(output)
	if _, err := vm.Run(chunk); err != nil {
		if runtimeErr, ok := err.(*bytecode.RuntimeError); ok {
			return &Result{
					Output:  extractOutput(output),
					Success: false,
				}, &RuntimeError{
					Message: runtimeErr.Error(),
				}
		}

		return &Result{
			Output:  extractOutput(output),
			Success: false,
		}, err
	}

	return &Result{
		Output:  extractOutput(output),
		Success: true,
	}, nil
}

func extractOutput(output io.Writer) string {
	if buf, ok := output.(*bytes.Buffer); ok {
		return buf.String()
	}
	return ""
}

// Eval is a convenience method that compiles and runs the source code in one call.
// This is equivalent to calling Compile() followed by Run().
//
// For better performance when executing the same code multiple times,
// use Compile() once and then call Run() multiple times.
//
// The output is captured and returned in the Result. If you want output to go
// to a specific writer, use WithOutput option when creating the engine.
func (e *Engine) Eval(source string) (*Result, error) {
	program, err := e.Compile(source)
	if err != nil {
		return nil, err
	}

	// If no output was specified, capture to a buffer
	if e.options.Output == nil {
		var buf bytes.Buffer
		oldOutput := e.options.Output
		e.options.Output = &buf
		result, err := e.Run(program)
		e.options.Output = oldOutput
		return result, err
	}

	return e.Run(program)
}

// Program represents a compiled DWScript program.
// It can be executed multiple times without re-compilation.
type Program struct {
	ast           *ast.Program
	analyzer      *semantic.Analyzer
	options       Options
	bytecodeChunk *bytecode.Chunk
}

// AST returns the Abstract Syntax Tree of the compiled program.
//
// This method provides read-only access to the parsed and type-checked AST.
// The AST can be used for static analysis, code transformation, or tooling.
//
// Note: Modifications to the returned AST will not affect program execution,
// as the interpreter works with the original AST captured during compilation.
//
// Example usage:
//
//	program, _ := engine.Compile("var x: Integer := 42;")
//	tree := program.AST()
//	for _, stmt := range tree.Statements {
//	    fmt.Printf("Statement: %s\n", stmt.String())
//	}
func (p *Program) AST() *ast.Program {
	return p.ast
}

func (p *Program) ensureBytecodeChunk() (*bytecode.Chunk, error) {
	if p == nil {
		return nil, fmt.Errorf("program is nil")
	}
	if p.bytecodeChunk != nil {
		return p.bytecodeChunk, nil
	}
	if p.ast == nil {
		return nil, fmt.Errorf("bytecode compilation requires AST")
	}

	compiler := bytecode.NewCompiler("dwscript")
	chunk, err := compiler.Compile(p.ast)
	if err != nil {
		return nil, err
	}
	p.bytecodeChunk = chunk
	return chunk, nil
}

// Result represents the result of executing a DWScript program.
type Result struct {
	// Output contains all text written to stdout during program execution.
	Output string

	// Success indicates whether the program completed without runtime errors.
	Success bool
}

// CompileError is returned when source code fails to compile or type-check.
type CompileError struct {
	// Stage indicates which compilation stage failed ("parsing" or "type checking").
	Stage string

	// Errors contains one or more structured errors describing what went wrong.
	// Each error includes position information, severity, and error codes for LSP integration.
	Errors []*Error
}

func (e *CompileError) Error() string {
	if len(e.Errors) == 1 {
		return fmt.Sprintf("%s error: %s", e.Stage, e.Errors[0].Error())
	}

	// Format multiple errors
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s errors (%d):\n", e.Stage, len(e.Errors))
	for i, err := range e.Errors {
		if i < 10 || i == len(e.Errors)-1 { // Show first 10 and last error
			fmt.Fprintf(&buf, "  - %s\n", err.Error())
		} else if i == 10 {
			fmt.Fprintf(&buf, "  ... and %d more errors\n", len(e.Errors)-11)
		}
	}
	return buf.String()
}

// HasErrors returns true if there are any errors (not just warnings).
func (e *CompileError) HasErrors() bool {
	for _, err := range e.Errors {
		if err.IsError() {
			return true
		}
	}
	return false
}

// HasWarnings returns true if there are any warnings.
func (e *CompileError) HasWarnings() bool {
	for _, err := range e.Errors {
		if err.IsWarning() {
			return true
		}
	}
	return false
}

// RuntimeError is returned when a program fails during execution.
type RuntimeError struct {
	// Message describes the runtime error.
	Message string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("runtime error: %s", e.Message)
}

// SetOutput sets the writer where program output (PrintLn, etc.) will be written.
// This is used internally by the engine but exposed for advanced use cases.
func (e *Engine) SetOutput(w io.Writer) {
	e.options.Output = w
}
