package bytecode

import (
	"bytes"
	"strings"
	"testing"
)

func TestDisassembleSimpleChunk(t *testing.T) {
	chunk := NewChunk("test")

	// Create a simple program: load two constants, add them, return
	const1 := chunk.AddConstant(IntValue(10))
	const2 := chunk.AddConstant(IntValue(32))

	chunk.Write(OpLoadConst, 0, uint16(const1), 1)
	chunk.Write(OpLoadConst, 0, uint16(const2), 1)
	chunk.WriteSimple(OpAddInt, 2)
	chunk.WriteSimple(OpReturn, 3)

	// Disassemble
	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)
	d.Disassemble()

	output := buf.String()

	// Check that output contains expected elements
	expectedStrings := []string{
		"== test ==",
		"Instructions: 4",
		"Constants: 2",
		"Constants Pool:",
		"[0000] 10",
		"[0001] 32",
		"Bytecode:",
		"LOAD_CONST",
		"ADD_INT",
		"RETURN",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Disassembly missing expected string %q\nOutput:\n%s", expected, output)
		}
	}
}

func TestDisassembleJumpInstructions(t *testing.T) {
	chunk := NewChunk("test_jumps")

	// Create conditional jump
	chunk.WriteSimple(OpLoadTrue, 1)
	jumpIdx := chunk.EmitJump(OpJumpIfFalse, 1)
	chunk.WriteSimple(OpLoadConst, 2)
	chunk.WriteSimple(OpLoadConst, 2)
	err := chunk.PatchJump(jumpIdx)
	if err != nil {
		t.Fatalf("PatchJump failed: %v", err)
	}
	chunk.WriteSimple(OpReturn, 3)

	// Disassemble
	output := DisassembleToString(chunk)

	// Check for jump information
	if !strings.Contains(output, "JUMP_IF_FALSE") {
		t.Error("Missing JUMP_IF_FALSE in output")
	}

	// Should show jump target
	if !strings.Contains(output, "->") {
		t.Error("Missing jump target indicator '->' in output")
	}
}

func TestDisassembleLoopInstructions(t *testing.T) {
	chunk := NewChunk("test_loop")

	// Create a simple loop
	loopStart := chunk.WriteSimple(OpLoadConst, 1)
	chunk.WriteSimple(OpLoadConst, 2)
	chunk.WriteSimple(OpAddInt, 2)
	err := chunk.EmitLoop(loopStart, 3)
	if err != nil {
		t.Fatalf("EmitLoop failed: %v", err)
	}

	// Disassemble
	output := DisassembleToString(chunk)

	// Check for loop instruction
	if !strings.Contains(output, "LOOP") {
		t.Error("Missing LOOP in output")
	}

	// Should show backward jump
	if !strings.Contains(output, "->") {
		t.Error("Missing jump target indicator in output")
	}
}

func TestDisassembleCallInstructions(t *testing.T) {
	chunk := NewChunk("test_calls")

	// Function call with 3 arguments
	chunk.Write(OpCall, 3, 10, 1)

	// Method call with 2 arguments
	chunk.Write(OpCallMethod, 2, 5, 2)

	// Disassemble
	output := DisassembleToString(chunk)

	// Check for call information
	if !strings.Contains(output, "CALL") {
		t.Error("Missing CALL in output")
	}

	if !strings.Contains(output, "args=3") {
		t.Error("Missing argument count in CALL output")
	}

	if !strings.Contains(output, "CALL_METHOD") {
		t.Error("Missing CALL_METHOD in output")
	}

	if !strings.Contains(output, "args=2") {
		t.Error("Missing argument count in CALL_METHOD output")
	}
}

func TestDisassembleRange(t *testing.T) {
	chunk := NewChunk("test")

	// Add several instructions
	for i := 0; i < 10; i++ {
		chunk.WriteSimple(OpAddInt, i+1)
	}

	// Disassemble only instructions 3-6
	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)
	d.DisassembleRange(3, 7)

	output := buf.String()

	// Should contain instructions 3-6
	if !strings.Contains(output, "0003") {
		t.Error("Missing instruction 0003")
	}
	if !strings.Contains(output, "0006") {
		t.Error("Missing instruction 0006")
	}

	// Should NOT contain instructions 0-2 or 7-9
	if strings.Contains(output, "0000") {
		t.Error("Should not contain instruction 0000")
	}
	if strings.Contains(output, "0007") {
		t.Error("Should not contain instruction 0007")
	}
}

func TestDisassembleLineNumbers(t *testing.T) {
	chunk := NewChunk("test")

	// Multiple instructions on line 5
	chunk.WriteSimple(OpLoadConst, 5)
	chunk.WriteSimple(OpLoadConst, 5)
	chunk.WriteSimple(OpLoadConst, 5)

	// Instructions on line 10
	chunk.WriteSimple(OpAddInt, 10)

	// More on line 10
	chunk.WriteSimple(OpAddInt, 10)

	// Disassemble
	output := DisassembleToString(chunk)

	// Check line numbers appear
	if !strings.Contains(output, "   5 ") {
		t.Error("Missing line number 5")
	}

	if !strings.Contains(output, "  10 ") {
		t.Error("Missing line number 10")
	}

	// Check pipe character for same line
	if !strings.Contains(output, "   |") {
		t.Error("Missing pipe character for continued line")
	}
}

func TestDisassembleConstantTypes(t *testing.T) {
	chunk := NewChunk("test")

	// Add various constant types
	chunk.AddConstant(IntValue(42))
	chunk.AddConstant(FloatValue(3.14))
	chunk.AddConstant(StringValue("hello"))
	chunk.AddConstant(BoolValue(true))
	chunk.AddConstant(NilValue())

	// Use them
	chunk.Write(OpLoadConst, 0, 0, 1)
	chunk.Write(OpLoadConst, 0, 1, 1)
	chunk.Write(OpLoadConst, 0, 2, 1)
	chunk.Write(OpLoadConst, 0, 3, 1)
	chunk.Write(OpLoadConst, 0, 4, 1)

	// Disassemble
	output := DisassembleToString(chunk)

	// Check constant representations
	expectedConstants := []string{
		"42",
		"3.14",
		`"hello"`,
		"true",
		"nil",
	}

	for _, expected := range expectedConstants {
		if !strings.Contains(output, expected) {
			t.Errorf("Missing constant %q in output:\n%s", expected, output)
		}
	}
}

func TestDisassembleInstructionToString(t *testing.T) {
	chunk := NewChunk("test")
	chunk.AddConstant(IntValue(42))
	chunk.Write(OpLoadConst, 0, 0, 1)
	chunk.WriteSimple(OpAddInt, 2)

	// Disassemble single instruction
	output1 := DisassembleInstructionToString(chunk, 0)
	if !strings.Contains(output1, "LOAD_CONST") {
		t.Error("Missing LOAD_CONST in single instruction output")
	}

	output2 := DisassembleInstructionToString(chunk, 1)
	if !strings.Contains(output2, "ADD_INT") {
		t.Error("Missing ADD_INT in single instruction output")
	}
}

func TestDisassembleForLoop(t *testing.T) {
	chunk := NewChunk("test_for")

	// For loop instruction
	chunk.Write(OpForPrep, 5, 100, 1) // loopVar=5, offset=100

	// Disassemble
	output := DisassembleToString(chunk)

	// Check for loop variable and offset
	if !strings.Contains(output, "FOR_PREP") {
		t.Error("Missing FOR_PREP in output")
	}

	if !strings.Contains(output, "var=5") {
		t.Error("Missing loop variable in output")
	}
}

func TestDisassembleClosure(t *testing.T) {
	chunk := NewChunk("test_closure")

	// Closure with 3 upvalues
	chunk.Write(OpClosure, 3, 10, 1)

	// Disassemble
	output := DisassembleToString(chunk)

	// Check for upvalue count
	if !strings.Contains(output, "CLOSURE") {
		t.Error("Missing CLOSURE in output")
	}

	if !strings.Contains(output, "upvalues=3") {
		t.Error("Missing upvalue count in output")
	}
}

func TestDisassembleObjectOperations(t *testing.T) {
	chunk := NewChunk("test_objects")

	// Object operations
	chunk.Write(OpNewObject, 0, 5, 1)    // class index 5
	chunk.Write(OpGetField, 0, 10, 2)    // field index 10
	chunk.Write(OpSetProperty, 0, 15, 3) // property index 15

	// Disassemble
	output := DisassembleToString(chunk)

	expectedOps := []string{
		"NEW_OBJECT",
		"GET_FIELD",
		"SET_PROPERTY",
	}

	for _, op := range expectedOps {
		if !strings.Contains(output, op) {
			t.Errorf("Missing %s in output", op)
		}
	}
}

func TestDisassembleArrayOperations(t *testing.T) {
	chunk := NewChunk("test_arrays")

	// Array operations
	chunk.Write(OpNewArray, 0, 5, 1) // 5 elements
	chunk.WriteSimple(OpArrayLength, 2)
	chunk.WriteSimple(OpArrayGet, 3)

	// Disassemble
	output := DisassembleToString(chunk)

	if !strings.Contains(output, "NEW_ARRAY") {
		t.Error("Missing NEW_ARRAY in output")
	}

	if !strings.Contains(output, "ARRAY_LENGTH") {
		t.Error("Missing ARRAY_LENGTH in output")
	}

	if !strings.Contains(output, "ARRAY_GET") {
		t.Error("Missing ARRAY_GET in output")
	}
}

func TestDisassembleExceptionHandling(t *testing.T) {
	chunk := NewChunk("test_exceptions")

	// Try-catch-finally
	tryIdx := chunk.EmitJump(OpTry, 1)
	chunk.WriteSimple(OpLoadConst, 2)
	chunk.PatchJump(tryIdx)
	catchIdx := chunk.EmitJump(OpCatch, 3)
	chunk.WriteSimple(OpPrint, 4)
	chunk.PatchJump(catchIdx)
	chunk.WriteSimple(OpFinally, 5)

	// Disassemble
	output := DisassembleToString(chunk)

	exceptionOps := []string{
		"TRY",
		"CATCH",
		"FINALLY",
	}

	for _, op := range exceptionOps {
		if !strings.Contains(output, op) {
			t.Errorf("Missing %s in output", op)
		}
	}
}

func BenchmarkDisassemble(b *testing.B) {
	// Create a chunk with 1000 instructions
	chunk := NewChunk("bench")
	for i := 0; i < 1000; i++ {
		chunk.AddConstant(IntValue(int64(i)))
		chunk.Write(OpLoadConst, 0, uint16(i), i/10+1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DisassembleToString(chunk)
	}
}

func BenchmarkDisassembleInstruction(b *testing.B) {
	chunk := NewChunk("bench")
	chunk.AddConstant(IntValue(42))
	chunk.Write(OpLoadConst, 0, 0, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DisassembleInstructionToString(chunk, 0)
	}
}
