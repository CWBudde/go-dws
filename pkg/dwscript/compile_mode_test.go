package dwscript

import "testing"

func TestEngineEvalBytecodeMode(t *testing.T) {
	script := `
var x: Integer := 40;
var y: Integer := 2;
x := x + y;
`

	engine, err := New(WithCompileMode(CompileModeBytecode))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	result, err := engine.Eval(script)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("bytecode eval reported failure")
	}
}

func TestCompileStoresBytecodeChunk(t *testing.T) {
	engine, err := New(WithCompileMode(CompileModeBytecode))
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	program, err := engine.Compile(`var n: Integer := 1;`)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if program.bytecodeChunk == nil {
		t.Fatalf("expected bytecode chunk to be populated when compiling in bytecode mode")
	}
}
