package bytecode

import (
	"bytes"
	"strings"
	"testing"
)

// TestDisassembler_TryDisassembleMiscOp tests the tryDisassembleMiscOp function
func TestDisassembler_TryDisassembleMiscOp(t *testing.T) {
	tests := []struct {
		name           string
		expectedOutput string
		inst           Instruction
		expectedResult bool
	}{
		{
			name:           "OpCase",
			inst:           MakeInstruction(OpCase, 0, 5),
			expectedOutput: "CASE",
			expectedResult: true,
		},
		{
			name:           "OpVariantToType",
			inst:           MakeInstruction(OpVariantToType, 0, 3),
			expectedOutput: "VARIANT_TO_TYPE",
			expectedResult: true,
		},
		{
			name:           "OpLoadConst_not_misc",
			inst:           MakeInstruction(OpLoadConst, 0, 0),
			expectedOutput: "",
			expectedResult: false,
		},
		{
			name:           "OpAdd_not_misc",
			inst:           MakeSimpleInstruction(OpAddInt),
			expectedOutput: "",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := NewChunk("test")
			chunk.WriteInstruction(tt.inst, 1)

			var buf bytes.Buffer
			d := NewDisassembler(chunk, &buf)

			result := d.tryDisassembleMiscOp(tt.inst, tt.inst.OpCode())

			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}

			if tt.expectedResult {
				output := buf.String()
				if !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
				}
			}
		})
	}
}

// TestDisassembler_InvokeInstruction tests the invokeInstruction function
func TestDisassembler_InvokeInstruction(t *testing.T) {
	tests := []struct {
		name           string
		expectedOutput []string
		inst           Instruction
	}{
		{
			name: "invoke_with_1_arg",
			inst: MakeInstruction(OpInvoke, 1, 5),
			expectedOutput: []string{
				"INVOKE",
				"args=1",
				"method=5",
			},
		},
		{
			name: "invoke_with_3_args",
			inst: MakeInstruction(OpInvoke, 3, 10),
			expectedOutput: []string{
				"INVOKE",
				"args=3",
				"method=10",
			},
		},
		{
			name: "invoke_with_0_args",
			inst: MakeInstruction(OpInvoke, 0, 2),
			expectedOutput: []string{
				"INVOKE",
				"args=0",
				"method=2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := NewChunk("test")
			chunk.WriteInstruction(tt.inst, 1)

			var buf bytes.Buffer
			d := NewDisassembler(chunk, &buf)

			d.invokeInstruction(tt.inst)

			output := buf.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

// TestDisassembler_DisassembleRange tests disassembling a range of instructions
func TestDisassembler_DisassembleRange(t *testing.T) {
	chunk := NewChunk("test_range")
	chunk.WriteInstruction(MakeSimpleInstruction(OpLoadNil), 1)
	chunk.WriteInstruction(MakeSimpleInstruction(OpLoadTrue), 2)
	chunk.WriteInstruction(MakeSimpleInstruction(OpLoadFalse), 3)
	chunk.WriteInstruction(MakeSimpleInstruction(OpReturn), 4)

	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)

	// Disassemble middle range
	d.DisassembleRange(1, 3)

	output := buf.String()

	// Should contain the chunk name
	if !strings.Contains(output, "test_range") {
		t.Error("Output should contain chunk name")
	}

	// Should contain instructions 1-2
	if !strings.Contains(output, "LOAD_TRUE") {
		t.Error("Output should contain LOAD_TRUE")
	}
	if !strings.Contains(output, "LOAD_FALSE") {
		t.Error("Output should contain LOAD_FALSE")
	}

	// Should NOT contain instruction 0 or 3
	// (This is harder to verify precisely, but we can check the range notation)
	if !strings.Contains(output, "1-2") {
		t.Error("Output should indicate range 1-2")
	}
}

// TestDisassembler_ClosureInstruction tests disassembling closure instructions
func TestDisassembler_ClosureInstruction(t *testing.T) {
	chunk := NewChunk("test")
	inst := MakeInstruction(OpClosure, 2, 5) // 2 upvalues, func at index 5
	chunk.WriteInstruction(inst, 1)

	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)

	d.closureInstruction(inst)

	output := buf.String()

	expectedStrings := []string{
		"CLOSURE",
		"upvalues=2",
		"func=5",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got %q", expected, output)
		}
	}
}

// TestDisassembler_CallInstruction tests disassembling call instructions
func TestDisassembler_CallInstruction(t *testing.T) {
	chunk := NewChunk("test")
	inst := MakeInstruction(OpCall, 3, 7) // 3 args, func at index 7
	chunk.WriteInstruction(inst, 1)

	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)

	d.callInstruction(inst)

	output := buf.String()

	expectedStrings := []string{
		"CALL",
		"args=3",
		"func=7",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got %q", expected, output)
		}
	}
}

// TestDisassembler_ReturnInstruction tests disassembling return instructions
func TestDisassembler_ReturnInstruction(t *testing.T) {
	tests := []struct {
		name           string
		expectedOutput string
		inst           Instruction
	}{
		{
			name:           "return_with_value",
			inst:           MakeInstruction(OpReturn, 1, 0),
			expectedOutput: "with value",
		},
		{
			name:           "return_without_value",
			inst:           MakeInstruction(OpReturn, 0, 0),
			expectedOutput: "RETURN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := NewChunk("test")
			chunk.WriteInstruction(tt.inst, 1)

			var buf bytes.Buffer
			d := NewDisassembler(chunk, &buf)

			d.returnInstruction(tt.inst)

			output := buf.String()
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}

// TestDisassembler_JumpInstruction tests disassembling jump instructions
func TestDisassembler_JumpInstruction(t *testing.T) {
	chunk := NewChunk("test")
	// Create a jump forward by 5 instructions
	inst := MakeInstruction(OpJump, 0, 5)
	chunk.WriteInstruction(inst, 1)

	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)

	// offset 0, sign +1 (forward jump)
	d.jumpInstruction(inst, 0, 1)

	output := buf.String()

	if !strings.Contains(output, "JUMP") {
		t.Error("Output should contain JUMP")
	}

	// Should show the jump offset and target
	if !strings.Contains(output, "->") {
		t.Error("Output should contain '->' to show jump target")
	}
}

// TestDisassembler_ForInstruction tests disassembling for-loop instructions
func TestDisassembler_ForInstruction(t *testing.T) {
	chunk := NewChunk("test")
	// For loop: var at index 2, jump forward by 10
	inst := MakeInstruction(OpForLoop, 2, 10)
	chunk.WriteInstruction(inst, 1)

	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)

	d.forInstruction(inst, 5) // offset 5

	output := buf.String()

	expectedStrings := []string{
		"var=2",
		"jump=",
		"->",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got %q", expected, output)
		}
	}
}

// TestDisassembler_ByteInstruction tests disassembling byte operand instructions
func TestDisassembler_ByteInstruction(t *testing.T) {
	chunk := NewChunk("test")
	inst := MakeInstruction(OpCase, 0, 42)
	chunk.WriteInstruction(inst, 1)

	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)

	d.byteInstruction(inst, "jumpTable")

	output := buf.String()

	if !strings.Contains(output, "42") {
		t.Error("Output should contain the byte operand value 42")
	}
	if !strings.Contains(output, "jumpTable") {
		t.Error("Output should contain the operand name 'jumpTable'")
	}
}

// TestDisassembler_ConstantInstruction tests disassembling constant instructions
func TestDisassembler_ConstantInstruction(t *testing.T) {
	chunk := NewChunk("test")
	chunk.AddConstant(IntValue(42))
	chunk.AddConstant(StringValue("hello"))

	inst1 := MakeInstruction(OpLoadConst, 0, 0)
	inst2 := MakeInstruction(OpLoadConst, 0, 1)

	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)

	d.constantInstruction(inst1, 0)
	output1 := buf.String()

	if !strings.Contains(output1, "42") {
		t.Error("Output should contain constant value 42")
	}

	buf.Reset()
	d.constantInstruction(inst2, 1)
	output2 := buf.String()

	if !strings.Contains(output2, "hello") {
		t.Error("Output should contain constant value 'hello'")
	}
}

// TestDisassembler_SimpleInstruction tests disassembling simple instructions
func TestDisassembler_SimpleInstruction(t *testing.T) {
	tests := []struct {
		name           string
		expectedOutput string
		inst           Instruction
	}{
		{
			name:           "OpLoadNil",
			inst:           MakeSimpleInstruction(OpLoadNil),
			expectedOutput: "LOAD_NIL",
		},
		{
			name:           "OpLoadTrue",
			inst:           MakeSimpleInstruction(OpLoadTrue),
			expectedOutput: "LOAD_TRUE",
		},
		{
			name:           "OpPop",
			inst:           MakeSimpleInstruction(OpPop),
			expectedOutput: "POP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := NewChunk("test")
			chunk.WriteInstruction(tt.inst, 1)

			var buf bytes.Buffer
			d := NewDisassembler(chunk, &buf)

			d.simpleInstruction(tt.inst)

			output := buf.String()
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}

// TestDisassembler_FullDisassembly tests complete disassembly
func TestDisassembler_FullDisassembly(t *testing.T) {
	chunk := NewChunk("complete_test")
	chunk.AddConstant(IntValue(42))
	chunk.AddConstant(StringValue("test"))

	chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 1), 2)
	chunk.WriteInstruction(MakeSimpleInstruction(OpAddInt), 2)
	chunk.WriteInstruction(MakeSimpleInstruction(OpReturn), 3)

	var buf bytes.Buffer
	d := NewDisassembler(chunk, &buf)

	d.Disassemble()

	output := buf.String()

	// Should contain chunk name
	if !strings.Contains(output, "complete_test") {
		t.Error("Output should contain chunk name")
	}

	// Should contain all instructions
	if !strings.Contains(output, "LOAD_CONST") {
		t.Error("Output should contain LOAD_CONST")
	}
	if !strings.Contains(output, "ADD_INT") {
		t.Error("Output should contain ADD_INT")
	}
	if !strings.Contains(output, "RETURN") {
		t.Error("Output should contain RETURN")
	}

	// Should contain constant values
	if !strings.Contains(output, "42") {
		t.Error("Output should contain constant 42")
	}
	if !strings.Contains(output, "test") {
		t.Error("Output should contain constant 'test'")
	}
}

// TestDisassembleToString tests the DisassembleToString helper
func TestDisassembleToString(t *testing.T) {
	chunk := NewChunk("string_test")
	chunk.AddConstant(IntValue(100))
	chunk.WriteInstruction(MakeInstruction(OpLoadConst, 0, 0), 1)
	chunk.WriteInstruction(MakeSimpleInstruction(OpReturn), 1)

	output := DisassembleToString(chunk)

	if !strings.Contains(output, "string_test") {
		t.Error("Output should contain chunk name")
	}
	if !strings.Contains(output, "100") {
		t.Error("Output should contain constant value")
	}
	if !strings.Contains(output, "LOAD_CONST") {
		t.Error("Output should contain LOAD_CONST")
	}
	if !strings.Contains(output, "RETURN") {
		t.Error("Output should contain RETURN")
	}
}
