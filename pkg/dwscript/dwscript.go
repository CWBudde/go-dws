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

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// Engine is the main entry point for the DWScript interpreter.
// It provides a high-level API for compiling and executing DWScript programs.
type Engine struct {
	options Options
}

// New creates a new DWScript engine with the given options.
// If no options are provided, sensible defaults are used.
func New(opts ...Option) (*Engine, error) {
	engine := &Engine{
		options: defaultOptions(),
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
		return nil, &CompileError{
			Stage:  "parsing",
			Errors: p.Errors(),
		}
	}

	// Type check (if enabled)
	if e.options.TypeCheck {
		analyzer := semantic.NewAnalyzer()
		if err := analyzer.Analyze(program); err != nil {
			return nil, &CompileError{
				Stage:  "type checking",
				Errors: []string{err.Error()},
			}
		}
	}

	return &Program{
		ast:     program,
		options: e.options,
	}, nil
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

	// Create interpreter with output writer
	interpreter := interp.New(output)

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

	// Errors contains one or more error messages describing what went wrong.
	Errors []string
}

func (e *CompileError) Error() string {
	if len(e.Errors) == 1 {
		return fmt.Sprintf("%s error: %s", e.Stage, e.Errors[0])
	}
	return fmt.Sprintf("%s errors: %v", e.Stage, e.Errors)
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
