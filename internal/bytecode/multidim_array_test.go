package bytecode_test

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/bytecode"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// runBothInterpreters runs the same source with both AST interpreter and bytecode VM
// and compares their outputs.
func runBothInterpreters(t *testing.T, source string) (astResult, vmResult string) {
	t.Helper()

	// Parse once
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Run AST interpreter
	var astBuf bytes.Buffer
	astInterp := interp.New(&astBuf)
	result := astInterp.Eval(program)
	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("AST interpreter error: %v", result)
	}
	astResult = astBuf.String()

	// Compile to bytecode
	compiler := bytecode.NewCompiler("test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compilation error: %v", err)
	}

	// Run bytecode VM
	var vmBuf bytes.Buffer
	vm := bytecode.NewVMWithOutput(&vmBuf)
	_, err = vm.Run(chunk)
	if err != nil {
		t.Fatalf("VM runtime error: %v", err)
	}
	vmResult = vmBuf.String()

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
