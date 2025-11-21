package bytecode_test

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/bytecode"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// NewTestLexer creates a lexer for the given source code.
// This is a convenience helper for tests to avoid repetitive lexer creation.
func NewTestLexer(source string) *lexer.Lexer {
	return lexer.New(source)
}

// NewTestParser creates a parser for the given lexer.
// This is a convenience helper for tests to avoid repetitive parser creation.
func NewTestParser(l *lexer.Lexer) *parser.Parser {
	return parser.New(l)
}

// NewTestProgram parses source code and returns the program, failing the test on parse errors.
// This is a convenience helper for tests to avoid repetitive parsing and error checking.
func NewTestProgram(t *testing.T, source string) *ast.Program {
	t.Helper()
	l := NewTestLexer(source)
	p := NewTestParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	return program
}

// NewTestASTInterpreter creates and runs an AST interpreter, returning the output.
// This is a convenience helper for tests to avoid repetitive interpreter setup.
func NewTestASTInterpreter(t *testing.T, program *ast.Program) string {
	t.Helper()
	var buf bytes.Buffer
	interp := interp.New(&buf)
	result := interp.Eval(program)
	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("AST interpreter error: %v", result)
	}
	return buf.String()
}

// NewTestBytecodeCompiler creates a bytecode compiler for the given program.
// This is a convenience helper for tests to avoid repetitive compiler creation.
func NewTestBytecodeCompiler(t *testing.T, program *ast.Program) *bytecode.Chunk {
	t.Helper()
	compiler := bytecode.NewCompiler("test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compilation error: %v", err)
	}
	return chunk
}

// NewTestVM creates and runs a bytecode VM, returning the output.
// This is a convenience helper for tests to avoid repetitive VM setup.
func NewTestVM(t *testing.T, chunk *bytecode.Chunk) string {
	t.Helper()
	var buf bytes.Buffer
	vm := bytecode.NewVMWithOutput(&buf)
	_, err := vm.Run(chunk)
	if err != nil {
		t.Fatalf("VM runtime error: %v", err)
	}
	return buf.String()
}

// runBothInterpreters runs the same source with both AST interpreter and bytecode VM
// and compares their outputs.
func runBothInterpreters(t *testing.T, source string) (astResult, vmResult string) {
	t.Helper()

	// Parse once
	program := NewTestProgram(t, source)

	// Run AST interpreter
	astResult = NewTestASTInterpreter(t, program)

	// Compile to bytecode
	chunk := NewTestBytecodeCompiler(t, program)

	// Run bytecode VM
	vmResult = NewTestVM(t, chunk)

	return astResult, vmResult
}

// TestMultiDimensionalArrayBytecode tests multi-dimensional array creation in bytecode VM
// by comparing its output with the AST interpreter.
func TestMultiDimensionalArrayBytecode(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "2D array - basic creation",
			source: `
				var arr: array of array of Integer;
				arr := new Integer[3, 4];
				PrintLn(IntToStr(Length(arr)));
			`,
		},
		{
			name: "2D array - inner length",
			source: `
				var arr: array of array of Integer;
				arr := new Integer[3, 4];
				PrintLn(IntToStr(Length(arr[0])));
			`,
		},
		{
			name: "2D array - set and get value",
			source: `
				var arr: array of array of Integer;
				arr := new Integer[3, 4];
				arr[1][2] := 42;
				PrintLn(IntToStr(arr[1][2]));
			`,
		},
		{
			name: "2D array - with expressions",
			source: `
				var M, N: Integer;
				var arr: array of array of Integer;
				M := 5;
				N := 10;
				arr := new Integer[M, N];
				PrintLn(IntToStr(Length(arr) * Length(arr[0])));
			`,
		},
		{
			name: "3D array - basic creation",
			source: `
				var arr: array of array of array of Integer;
				arr := new Integer[2, 3, 4];
				PrintLn(IntToStr(Length(arr)));
			`,
		},
		{
			name: "3D array - nested access",
			source: `
				var arr: array of array of array of Integer;
				arr := new Integer[2, 3, 4];
				arr[1][2][3] := 99;
				PrintLn(IntToStr(arr[1][2][3]));
			`,
		},
		{
			name: "single dimension should still work",
			source: `
				var arr: array of Integer;
				arr := new Integer[10];
				PrintLn(IntToStr(Length(arr)));
			`,
		},
		{
			name: "2D array - Float type",
			source: `
				var arr: array of array of Float;
				arr := new Float[2, 3];
				arr[0][1] := 3.14;
				PrintLn(IntToStr(Integer(arr[0][1] * 100)));
			`,
		},
		{
			name: "2D array - String type",
			source: `
				var arr: array of array of String;
				var s: String;
				arr := new String[2, 3];
				arr[0][0] := 'hello';
				arr[1][2] := 'world';
				s := arr[0][0] + ' ' + arr[1][2];
				PrintLn(IntToStr(Length(s)));
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			astOutput, vmOutput := runBothInterpreters(t, tt.source)
			if astOutput != vmOutput {
				t.Errorf("AST and VM outputs differ:\nAST: %q\nVM:  %q", astOutput, vmOutput)
			}
		})
	}
}
