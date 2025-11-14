package bytecode

import (
	"testing"
)

func TestInstructionEncoding(t *testing.T) {
	tests := []struct {
		name     string
		op       OpCode
		a        byte
		b        uint16
		expected Instruction
	}{
		{
			name:     "Simple instruction with no operands",
			op:       OpHalt,
			a:        0,
			b:        0,
			expected: Instruction(OpHalt),
		},
		{
			name:     "Load constant at index 42",
			op:       OpLoadConst,
			a:        0,
			b:        42,
			expected: Instruction(uint32(OpLoadConst) | (42 << 16)),
		},
		{
			name:     "Load local at index 5",
			op:       OpLoadLocal,
			a:        0,
			b:        5,
			expected: Instruction(uint32(OpLoadLocal) | (5 << 16)),
		},
		{
			name:     "Jump with offset 100",
			op:       OpJump,
			a:        0,
			b:        100,
			expected: Instruction(uint32(OpJump) | (100 << 16)),
		},
		{
			name:     "Call with 3 arguments and function index 10",
			op:       OpCall,
			a:        3,
			b:        10,
			expected: Instruction(uint32(OpCall) | (3 << 8) | (10 << 16)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := MakeInstruction(tt.op, tt.a, tt.b)
			if inst != tt.expected {
				t.Errorf("MakeInstruction() = 0x%08X, want 0x%08X", inst, tt.expected)
			}
		})
	}
}

func TestInstructionDecoding(t *testing.T) {
	tests := []struct {
		name string
		inst Instruction
		op   OpCode
		a    byte
		b    uint16
	}{
		{
			name: "Decode OpHalt",
			inst: MakeSimpleInstruction(OpHalt),
			op:   OpHalt,
			a:    0,
			b:    0,
		},
		{
			name: "Decode OpLoadConst with index 42",
			inst: MakeInstruction(OpLoadConst, 0, 42),
			op:   OpLoadConst,
			a:    0,
			b:    42,
		},
		{
			name: "Decode OpCall with 3 args and func 10",
			inst: MakeInstruction(OpCall, 3, 10),
			op:   OpCall,
			a:    3,
			b:    10,
		},
		{
			name: "Decode OpAddInt",
			inst: MakeSimpleInstruction(OpAddInt),
			op:   OpAddInt,
			a:    0,
			b:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if op := tt.inst.OpCode(); op != tt.op {
				t.Errorf("OpCode() = %v, want %v", op, tt.op)
			}
			if a := tt.inst.A(); a != tt.a {
				t.Errorf("A() = %v, want %v", a, tt.a)
			}
			if b := tt.inst.B(); b != tt.b {
				t.Errorf("B() = %v, want %v", b, tt.b)
			}
		})
	}
}

func TestInstructionABC(t *testing.T) {
	// Test instruction with three byte operands
	inst := MakeInstructionABC(OpCall, 3, 5, 7)

	if op := inst.OpCode(); op != OpCall {
		t.Errorf("OpCode() = %v, want %v", op, OpCall)
	}
	if a := inst.A(); a != 3 {
		t.Errorf("A() = %v, want 3", a)
	}
	if b := inst.B(); b != 0x0705 {
		t.Errorf("B() = 0x%04X, want 0x0705", b)
	}
	if c := inst.C(); c != 7 {
		t.Errorf("C() = %v, want 7", c)
	}
}

func TestSignedBOperand(t *testing.T) {
	tests := []struct {
		name   string
		b      uint16
		signed int16
	}{
		{
			name:   "Positive offset",
			b:      100,
			signed: 100,
		},
		{
			name:   "Negative offset",
			b:      0xFFFF, // -1 in two's complement
			signed: -1,
		},
		{
			name:   "Large negative offset",
			b:      0xFF9C, // -100 in two's complement
			signed: -100,
		},
		{
			name:   "Zero offset",
			b:      0,
			signed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := MakeInstruction(OpJump, 0, tt.b)
			if signed := inst.SignedB(); signed != tt.signed {
				t.Errorf("SignedB() = %v, want %v", signed, tt.signed)
			}
		})
	}
}

func TestOpCodeNames(t *testing.T) {
	tests := []struct {
		name string
		op   OpCode
	}{
		{"LOAD_CONST", OpLoadConst},
		{"ADD_INT", OpAddInt},
		{"ADD_FLOAT", OpAddFloat},
		{"JUMP", OpJump},
		{"CALL", OpCall},
		{"RETURN", OpReturn},
		{"HALT", OpHalt},
		{"PRINT", OpPrint},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OpCodeNames[tt.op]; got != tt.name {
				t.Errorf("OpCodeNames[%v] = %q, want %q", tt.op, got, tt.name)
			}
		})
	}
}

func TestInstructionString(t *testing.T) {
	tests := []struct {
		want string
		inst Instruction
	}{
		{"HALT", MakeSimpleInstruction(OpHalt)},
		{"LOAD_CONST", MakeInstruction(OpLoadConst, 0, 42)},
		{"ADD_INT", MakeInstruction(OpAddInt, 0, 0)},
		{"JUMP", MakeInstruction(OpJump, 0, 100)},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.inst.String(); got != tt.want {
				t.Errorf("Instruction.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpcodeCount(t *testing.T) {
	// Verify we have exactly 115 opcodes (under 128 for optimal switch performance)
	const expectedOpcodeCount = 115

	// Count non-empty opcode names
	count := 0
	for _, name := range OpCodeNames {
		if name != "" {
			count++
		}
	}

	if count != expectedOpcodeCount {
		t.Errorf("Expected %d opcodes, got %d", expectedOpcodeCount, count)
	}

	// Verify all opcodes are under 128 for optimal Go switch performance
	if OpDebugger >= 128 {
		t.Errorf("Highest opcode (OpDebugger) is %d, should be < 128 for optimal switch performance", OpDebugger)
	}
}

func BenchmarkMakeInstruction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MakeInstruction(OpLoadConst, 0, 42)
	}
}

func BenchmarkInstructionDecode(b *testing.B) {
	inst := MakeInstruction(OpCall, 3, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = inst.OpCode()
		_ = inst.A()
		_ = inst.B()
	}
}
