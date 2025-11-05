package bytecode

import (
	"testing"
)

func TestNewChunk(t *testing.T) {
	chunk := NewChunk("test")

	if chunk.Name != "test" {
		t.Errorf("NewChunk() name = %q, want %q", chunk.Name, "test")
	}

	if len(chunk.Code) != 0 {
		t.Errorf("NewChunk() should have empty code, got %d instructions", len(chunk.Code))
	}

	if len(chunk.Constants) != 0 {
		t.Errorf("NewChunk() should have empty constants, got %d constants", len(chunk.Constants))
	}

	if len(chunk.Lines) != 0 {
		t.Errorf("NewChunk() should have empty lines, got %d line entries", len(chunk.Lines))
	}
}

func TestWriteInstruction(t *testing.T) {
	chunk := NewChunk("test")

	// Write a few instructions
	idx1 := chunk.WriteSimple(OpHalt, 1)
	idx2 := chunk.Write(OpLoadConst, 0, 5, 2)
	idx3 := chunk.WriteSimple(OpAddInt, 3)

	// Check indices
	if idx1 != 0 {
		t.Errorf("First instruction index = %d, want 0", idx1)
	}
	if idx2 != 1 {
		t.Errorf("Second instruction index = %d, want 1", idx2)
	}
	if idx3 != 2 {
		t.Errorf("Third instruction index = %d, want 2", idx3)
	}

	// Check code was written
	if len(chunk.Code) != 3 {
		t.Errorf("Code length = %d, want 3", len(chunk.Code))
	}

	// Check instructions
	if chunk.Code[0].OpCode() != OpHalt {
		t.Errorf("First instruction opcode = %v, want OpHalt", chunk.Code[0].OpCode())
	}
	if chunk.Code[1].OpCode() != OpLoadConst {
		t.Errorf("Second instruction opcode = %v, want OpLoadConst", chunk.Code[1].OpCode())
	}
	if chunk.Code[1].B() != 5 {
		t.Errorf("Second instruction operand = %d, want 5", chunk.Code[1].B())
	}
}

func TestAddConstant(t *testing.T) {
	chunk := NewChunk("test")

	// Add different types of constants
	idx1 := chunk.AddConstant(IntValue(42))
	idx2 := chunk.AddConstant(FloatValue(3.14))
	idx3 := chunk.AddConstant(StringValue("hello"))
	idx4 := chunk.AddConstant(BoolValue(true))
	idx5 := chunk.AddConstant(NilValue())

	// Check indices
	if idx1 != 0 || idx2 != 1 || idx3 != 2 || idx4 != 3 || idx5 != 4 {
		t.Errorf("Constant indices not sequential: %d, %d, %d, %d, %d", idx1, idx2, idx3, idx4, idx5)
	}

	// Check constant count
	if chunk.ConstantCount() != 5 {
		t.Errorf("ConstantCount() = %d, want 5", chunk.ConstantCount())
	}

	// Check constants can be retrieved
	if val := chunk.GetConstant(0); val.AsInt() != 42 {
		t.Errorf("GetConstant(0) = %v, want 42", val)
	}
	if val := chunk.GetConstant(1); val.AsFloat() != 3.14 {
		t.Errorf("GetConstant(1) = %v, want 3.14", val)
	}
	if val := chunk.GetConstant(2); val.AsString() != "hello" {
		t.Errorf("GetConstant(2) = %q, want %q", val.AsString(), "hello")
	}
}

func TestConstantDeduplication(t *testing.T) {
	chunk := NewChunk("test")

	// Add same constant multiple times
	idx1 := chunk.AddConstant(IntValue(42))
	idx2 := chunk.AddConstant(IntValue(42))
	idx3 := chunk.AddConstant(IntValue(42))

	// Should return same index
	if idx1 != idx2 || idx2 != idx3 {
		t.Errorf("Constant deduplication failed: %d, %d, %d", idx1, idx2, idx3)
	}

	// Should only have one constant
	if chunk.ConstantCount() != 1 {
		t.Errorf("ConstantCount() = %d, want 1 (deduplication failed)", chunk.ConstantCount())
	}

	// Same for strings
	idx4 := chunk.AddConstant(StringValue("test"))
	idx5 := chunk.AddConstant(StringValue("test"))

	if idx4 != idx5 {
		t.Errorf("String deduplication failed: %d != %d", idx4, idx5)
	}

	if chunk.ConstantCount() != 2 {
		t.Errorf("ConstantCount() = %d, want 2", chunk.ConstantCount())
	}
}

func TestLineInfoEncoding(t *testing.T) {
	chunk := NewChunk("test")

	// Write multiple instructions on same line
	chunk.WriteSimple(OpLoadConst, 1)
	chunk.WriteSimple(OpLoadConst, 1)
	chunk.WriteSimple(OpLoadConst, 1)

	// Should only have one LineInfo entry (run-length encoding)
	if len(chunk.Lines) != 1 {
		t.Errorf("Lines length = %d, want 1 (run-length encoding failed)", len(chunk.Lines))
	}

	// Write instruction on different line
	chunk.WriteSimple(OpAddInt, 2)

	// Should now have two entries
	if len(chunk.Lines) != 2 {
		t.Errorf("Lines length = %d, want 2", len(chunk.Lines))
	}

	// Write more on line 2
	chunk.WriteSimple(OpAddInt, 2)
	chunk.WriteSimple(OpAddInt, 2)

	// Should still have two entries
	if len(chunk.Lines) != 2 {
		t.Errorf("Lines length = %d, want 2", len(chunk.Lines))
	}
}

func TestGetLine(t *testing.T) {
	chunk := NewChunk("test")

	// Line 1: instructions 0-2
	chunk.WriteSimple(OpLoadConst, 1)
	chunk.WriteSimple(OpLoadConst, 1)
	chunk.WriteSimple(OpLoadConst, 1)

	// Line 5: instructions 3-4
	chunk.WriteSimple(OpAddInt, 5)
	chunk.WriteSimple(OpAddInt, 5)

	// Line 10: instruction 5
	chunk.WriteSimple(OpReturn, 10)

	// Test line retrieval
	tests := []struct {
		instructionIndex int
		expectedLine     int
	}{
		{0, 1},
		{1, 1},
		{2, 1},
		{3, 5},
		{4, 5},
		{5, 10},
	}

	for _, tt := range tests {
		line := chunk.GetLine(tt.instructionIndex)
		if line != tt.expectedLine {
			t.Errorf("GetLine(%d) = %d, want %d", tt.instructionIndex, line, tt.expectedLine)
		}
	}
}

func TestPatchJump(t *testing.T) {
	chunk := NewChunk("test")

	// Emit a jump instruction
	jumpIdx := chunk.EmitJump(OpJumpIfFalse, 1)

	// Write some instructions
	chunk.WriteSimple(OpLoadConst, 2)
	chunk.WriteSimple(OpLoadConst, 2)
	chunk.WriteSimple(OpAddInt, 2)

	// Patch the jump
	err := chunk.PatchJump(jumpIdx)
	if err != nil {
		t.Fatalf("PatchJump() error = %v", err)
	}

	// Check the jump offset
	inst := chunk.Code[jumpIdx]
	offset := inst.SignedB()

	// Should jump 3 instructions forward (to instruction 4)
	if offset != 3 {
		t.Errorf("Jump offset = %d, want 3", offset)
	}
}

func TestEmitLoop(t *testing.T) {
	chunk := NewChunk("test")

	// Write some instructions
	loopStart := chunk.WriteSimple(OpLoadConst, 1)
	chunk.WriteSimple(OpLoadConst, 2)
	chunk.WriteSimple(OpAddInt, 2)

	// Emit backward jump to loopStart
	err := chunk.EmitLoop(loopStart, 3)
	if err != nil {
		t.Fatalf("EmitLoop() error = %v", err)
	}

	// Check the loop instruction
	loopInst := chunk.Code[len(chunk.Code)-1]
	if loopInst.OpCode() != OpLoop {
		t.Errorf("Loop instruction opcode = %v, want OpLoop", loopInst.OpCode())
	}

	// Offset should be negative (backward jump)
	offset := loopInst.SignedB()
	if offset >= 0 {
		t.Errorf("Loop offset = %d, should be negative", offset)
	}
}

func TestValueConstructors(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		wantType ValueType
		check    func(Value) bool
	}{
		{
			name:     "NilValue",
			value:    NilValue(),
			wantType: ValueNil,
			check:    func(v Value) bool { return v.IsNil() },
		},
		{
			name:     "BoolValue true",
			value:    BoolValue(true),
			wantType: ValueBool,
			check:    func(v Value) bool { return v.IsBool() && v.AsBool() == true },
		},
		{
			name:     "BoolValue false",
			value:    BoolValue(false),
			wantType: ValueBool,
			check:    func(v Value) bool { return v.IsBool() && v.AsBool() == false },
		},
		{
			name:     "IntValue",
			value:    IntValue(42),
			wantType: ValueInt,
			check:    func(v Value) bool { return v.IsInt() && v.AsInt() == 42 },
		},
		{
			name:     "FloatValue",
			value:    FloatValue(3.14),
			wantType: ValueFloat,
			check:    func(v Value) bool { return v.IsFloat() && v.AsFloat() == 3.14 },
		},
		{
			name:     "StringValue",
			value:    StringValue("hello"),
			wantType: ValueString,
			check:    func(v Value) bool { return v.IsString() && v.AsString() == "hello" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", tt.value.Type, tt.wantType)
			}
			if !tt.check(tt.value) {
				t.Errorf("Value check failed for %v", tt.value)
			}
		})
	}
}

func TestValueString(t *testing.T) {
	tests := []struct {
		value Value
		want  string
	}{
		{NilValue(), "nil"},
		{BoolValue(true), "true"},
		{BoolValue(false), "false"},
		{IntValue(42), "42"},
		{FloatValue(3.14), "3.14"},
		{StringValue("hello"), `"hello"`},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChunkValidate(t *testing.T) {
	t.Run("Valid chunk", func(t *testing.T) {
		chunk := NewChunk("test")
		constIdx := chunk.AddConstant(IntValue(42))
		chunk.Write(OpLoadConst, 0, uint16(constIdx), 1)

		if err := chunk.Validate(); err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("Invalid constant reference", func(t *testing.T) {
		chunk := NewChunk("test")
		// Reference constant 5 but no constants exist
		chunk.Write(OpLoadConst, 0, 5, 1)

		if err := chunk.Validate(); err == nil {
			t.Error("Validate() error = nil, want error for invalid constant reference")
		}
	})

	t.Run("LOAD_CONST_0 without constants", func(t *testing.T) {
		chunk := NewChunk("test")
		chunk.WriteSimple(OpLoadConst0, 1)

		if err := chunk.Validate(); err == nil {
			t.Error("Validate() error = nil, want error for LOAD_CONST_0 without constants")
		}
	})
}

func TestChunkStats(t *testing.T) {
	chunk := NewChunk("test_function")

	// Add some instructions and constants
	chunk.AddConstant(IntValue(42))
	chunk.AddConstant(FloatValue(3.14))
	chunk.WriteSimple(OpLoadConst, 1)
	chunk.WriteSimple(OpLoadConst, 1)
	chunk.WriteSimple(OpAddInt, 2)
	chunk.WriteSimple(OpReturn, 3)

	stats := chunk.GetStats()

	if stats.Name != "test_function" {
		t.Errorf("Stats.Name = %q, want %q", stats.Name, "test_function")
	}
	if stats.InstructionCount != 4 {
		t.Errorf("Stats.InstructionCount = %d, want 4", stats.InstructionCount)
	}
	if stats.ConstantCount != 2 {
		t.Errorf("Stats.ConstantCount = %d, want 2", stats.ConstantCount)
	}
	if stats.CodeBytes != 16 { // 4 instructions * 4 bytes
		t.Errorf("Stats.CodeBytes = %d, want 16", stats.CodeBytes)
	}
	if stats.UniqueLines != 3 {
		t.Errorf("Stats.UniqueLines = %d, want 3", stats.UniqueLines)
	}
}

func TestValueIsNumber(t *testing.T) {
	tests := []struct {
		value    Value
		isNumber bool
	}{
		{IntValue(42), true},
		{FloatValue(3.14), true},
		{StringValue("42"), false},
		{BoolValue(true), false},
		{NilValue(), false},
	}

	for _, tt := range tests {
		if got := tt.value.IsNumber(); got != tt.isNumber {
			t.Errorf("%v.IsNumber() = %v, want %v", tt.value, got, tt.isNumber)
		}
	}
}

func TestIntToFloatConversion(t *testing.T) {
	intVal := IntValue(42)
	floatResult := intVal.AsFloat()

	if floatResult != 42.0 {
		t.Errorf("IntValue(42).AsFloat() = %f, want 42.0", floatResult)
	}
}

func BenchmarkWriteInstruction(b *testing.B) {
	chunk := NewChunk("bench")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunk.WriteSimple(OpAddInt, 1)
	}
}

func BenchmarkAddConstant(b *testing.B) {
	chunk := NewChunk("bench")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunk.AddConstant(IntValue(int64(i)))
	}
}

func BenchmarkGetLine(b *testing.B) {
	chunk := NewChunk("bench")
	// Create chunk with 1000 instructions across 100 lines
	for i := 0; i < 1000; i++ {
		chunk.WriteSimple(OpAddInt, i/10+1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chunk.GetLine(i % 1000)
	}
}

func BenchmarkAddConstantWithDedup(b *testing.B) {
	chunk := NewChunk("bench")
	// Pre-fill with some constants
	for i := 0; i < 100; i++ {
		chunk.AddConstant(IntValue(int64(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will trigger deduplication
		chunk.AddConstant(IntValue(42))
	}
}
