package bytecode_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/bytecode"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestVMParityBasic tests that the bytecode VM produces the same output as the AST interpreter.
func TestVMParityBasic(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "PrintLn basic",
			source: `PrintLn('Hello, World!');`,
		},
		{
			name:   "Multiple PrintLn",
			source: `PrintLn('Line 1'); PrintLn('Line 2'); PrintLn('Line 3');`,
		},
		{
			name: "Integer variables",
			source: `
				var x: Integer := 42;
				PrintLn(IntToStr(x));
			`,
		},
		{
			name: "String variables",
			source: `
				var s: String := 'test';
				PrintLn(s);
			`,
		},
		{
			name: "Arithmetic",
			source: `
				var a: Integer := 10;
				var b: Integer := 20;
				var c: Integer := a + b;
				PrintLn(IntToStr(c));
			`,
		},
		{
			name: "If statement",
			source: `
				var x: Integer := 10;
				if x > 5 then
					PrintLn('Greater')
				else
					PrintLn('Less or equal');
			`,
		},
		{
			name: "While loop",
			source: `
				var i: Integer := 0;
				while i < 3 do begin
					PrintLn(IntToStr(i));
					i := i + 1;
				end;
			`,
		},
		// TODO: For loop not yet supported in bytecode compiler
		// {
		// 	name: "For loop",
		// 	source: `
		// 		var i: Integer;
		// 		for i := 1 to 3 do
		// 			PrintLn(IntToStr(i));
		// 	`,
		// },
		// TODO: Result variable not yet supported in bytecode compiler
		// {
		// 	name: "Function call",
		// 	source: `
		// 		function Add(a, b: Integer): Integer;
		// 		begin
		// 			Result := a + b;
		// 		end;
		//
		// 		PrintLn(IntToStr(Add(5, 7)));
		// 	`,
		// },
		{
			name: "String functions",
			source: `
				var s: String := 'Hello';
				PrintLn(IntToStr(Length(s)));
				PrintLn(Copy(s, 2, 3));
			`,
		},
		{
			name: "Ord",
			source: `
				PrintLn(IntToStr(Ord('A')));
			`,
		},
		{
			name: "Type conversion",
			source: `
				var x: Integer := 42;
				PrintLn(IntToStr(x));
				var y: Integer := StrToInt('100');
				PrintLn(IntToStr(y));
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run with AST interpreter
			astOutput := runWithInterpreter(t, tt.source)

			// Run with bytecode VM
			bcOutput := runWithBytecode(t, tt.source)

			// Compare outputs
			if astOutput != bcOutput {
				t.Errorf("Output mismatch:\nAST output:\n%s\nBytecode output:\n%s", astOutput, bcOutput)
			}
		})
	}
}

func runWithInterpreter(t *testing.T, source string) string {
	t.Helper()

	var output bytes.Buffer
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	interpreter := interp.New(&output)
	result := interpreter.Eval(program)

	if result != nil && result.Type() == "ERROR" {
		t.Fatalf("Interpreter error: %v", result)
	}

	return output.String()
}

func runWithBytecode(t *testing.T, source string) string {
	t.Helper()

	var output bytes.Buffer
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	compiler := bytecode.NewCompiler("test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compiler error: %v", err)
	}

	vm := bytecode.NewVMWithOutput(&output)
	_, err = vm.Run(chunk)
	if err != nil {
		t.Fatalf("VM error: %v", err)
	}

	return output.String()
}

// TestBytecodeDisassemblerOutput tests that the disassembler produces valid output.
func TestBytecodeDisassemblerOutput(t *testing.T) {
	source := `
		var x: Integer := 42;
		PrintLn(IntToStr(x));
	`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	compiler := bytecode.NewCompiler("test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compiler error: %v", err)
	}

	var disasmOutput bytes.Buffer
	disasm := bytecode.NewDisassembler(chunk, &disasmOutput)
	disasm.Disassemble()

	output := disasmOutput.String()

	// Verify the output contains expected instructions
	expectedInstructions := []string{
		"LOAD_CONST",
		"STORE_GLOBAL",
		"LOAD_GLOBAL",
		"HALT",
	}

	for _, instr := range expectedInstructions {
		if !strings.Contains(output, instr) {
			t.Errorf("Disassembler output missing expected instruction: %s\nOutput:\n%s", instr, output)
		}
	}
}

// BenchmarkVMVsInterpreter compares performance of bytecode VM vs AST interpreter.
func BenchmarkVMVsInterpreter(b *testing.B) {
	source := `
		var sum: Integer := 0;
		var i: Integer := 1;
		while i <= 100 do begin
			sum := sum + i;
			i := i + 1;
		end;
		PrintLn(IntToStr(sum));
	`

	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		b.Fatalf("Parser errors: %v", p.Errors())
	}

	// Compile once for bytecode
	compiler := bytecode.NewCompiler("bench")
	chunk, err := compiler.Compile(program)
	if err != nil {
		b.Fatalf("Compiler error: %v", err)
	}

	b.Run("AST Interpreter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var output bytes.Buffer
			interpreter := interp.New(&output)
			_ = interpreter.Eval(program)
		}
	})

	b.Run("Bytecode VM", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var output bytes.Buffer
			vm := bytecode.NewVMWithOutput(&output)
			_, _ = vm.Run(chunk)
		}
	})
}

// TestVMParityIsExpressions tests that the bytecode VM produces the same output as the AST interpreter for 'is' expressions.
func TestVMParityIsExpressions(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "is True basic",
			source: `
				var x: Boolean := True;
				if x is True then
					PrintLn('True')
				else
					PrintLn('False');
			`,
		},
		{
			name: "is False basic",
			source: `
				var x: Boolean := False;
				if x is False then
					PrintLn('False')
				else
					PrintLn('Not false');
			`,
		},
		{
			name: "is with integer to boolean conversion",
			source: `
				var x: Integer := 1;
				if x is True then
					PrintLn('Truthy')
				else
					PrintLn('Falsey');
			`,
		},
		{
			name: "is with zero to boolean conversion",
			source: `
				var x: Integer := 0;
				if x is False then
					PrintLn('Zero is falsey')
				else
					PrintLn('Zero is truthy');
			`,
		},
		{
			name: "is True with precedence",
			source: `
				var a: Boolean := True;
				var b: Boolean := True;
				if a is True and b is True then
					PrintLn('Both true')
				else
					PrintLn('Not both true');
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run with AST interpreter
			astOutput := runWithInterpreter(t, tt.source)

			// Run with bytecode VM
			bcOutput := runWithBytecode(t, tt.source)

			// Compare outputs
			if astOutput != bcOutput {
				t.Errorf("Output mismatch:\nAST output:\n%s\nBytecode output:\n%s", astOutput, bcOutput)
			}
		})
	}
}
