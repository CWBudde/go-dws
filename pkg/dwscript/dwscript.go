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

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
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
	if e.options.TypeCheck {
		analyzer := semantic.NewAnalyzer()
		if err := analyzer.Analyze(program); err != nil {
			// Convert semantic errors
			errors := convertSemanticError(err)
			return nil, &CompileError{
				Stage:  "type checking",
				Errors: errors,
			}
		}
	}

	return &Program{
		ast:     program,
		options: e.options,
	}, nil
}

// convertSemanticError converts semantic analysis errors to structured Error objects.
func convertSemanticError(err error) []*Error {
	// Check if it's already a semantic.AnalysisError
	if analysisErr, ok := err.(*semantic.AnalysisError); ok {
		errors := make([]*Error, 0, len(analysisErr.Errors))
		for _, errStr := range analysisErr.Errors {
			// Parse position from error string
			line, column := 0, 0
			message := errStr

			// Try to extract position
			var parsed bool
			_, parseErr := fmt.Sscanf(errStr, "%s at %d:%d", &message, &line, &column)
			if parseErr == nil {
				parsed = true
				if idx := strings.LastIndex(errStr, " at "); idx >= 0 {
					message = errStr[:idx]
				}
			}

			if !parsed {
				message = errStr
			}

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

	// Set external functions in options for interpreter
	e.options.ExternalFunctions = e.externalFunctions

	// Create interpreter with output writer and options
	interpreter := interp.NewWithOptions(output, &e.options)

	// Execute
	result := interpreter.Eval(program.ast)

	// Check for runtime errors
	if result != nil && result.Type() == "ERROR" {
		outputStr := ""
		if buf, ok := output.(*bytes.Buffer); ok {
			outputStr = buf.String()
		}
		return &Result{
				Output:  outputStr,
				Success: false,
			}, &RuntimeError{
				Message: result.String(),
			}
	}

	// Extract output if it was a buffer
	outputStr := ""
	if buf, ok := output.(*bytes.Buffer); ok {
		outputStr = buf.String()
	}

	return &Result{
		Output:  outputStr,
		Success: true,
	}, nil
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
	ast     *ast.Program
	options Options
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
