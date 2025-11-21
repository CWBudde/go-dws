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
		value    Value
		check    func(Value) bool
		name     string
		wantType ValueType
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

func TestChunkOptimizeRemovesLiteralPushPopPairs(t *testing.T) {
	chunk := NewChunk("opt")
	constIdx := chunk.AddConstant(IntValue(42))

	chunk.Write(OpLoadConst, 0, uint16(constIdx), 1)
	chunk.WriteSimple(OpPop, 1)
	chunk.WriteSimple(OpLoadTrue, 2)
	chunk.WriteSimple(OpPop, 2)
	chunk.WriteSimple(OpHalt, 3)

	chunk.Optimize()

	if len(chunk.Code) != 1 {
		t.Fatalf("expected 1 instruction after optimization, got %d", len(chunk.Code))
	}
	if chunk.Code[0].OpCode() != OpHalt {
		t.Fatalf("expected remaining instruction to be OpHalt, got %v", chunk.Code[0].OpCode())
	}
	if line := chunk.GetLine(0); line != 3 {
		t.Fatalf("expected OpHalt line to remain 3, got %d", line)
	}
}

func TestChunkOptimizeUpdatesJumpOffsets(t *testing.T) {
	chunk := NewChunk("jump")
	constIdx := chunk.AddConstant(IntValue(1))

	chunk.Write(OpJump, 0, 2, 1) // jump over next two instructions
	chunk.Write(OpLoadConst, 0, uint16(constIdx), 2)
	chunk.WriteSimple(OpPop, 2)
	chunk.Write(OpReturn, 0, 0, 3)

	chunk.Optimize()

	if len(chunk.Code) != 2 {
		t.Fatalf("expected 2 instructions after optimization, got %d", len(chunk.Code))
	}
	if chunk.Code[0].OpCode() != OpJump {
		t.Fatalf("expected first instruction to remain OpJump, got %v", chunk.Code[0].OpCode())
	}
	if chunk.Code[0].B() != 0 {
		t.Fatalf("expected jump offset to be 0 after removing dead instructions, got %d", chunk.Code[0].B())
	}
	if chunk.Code[1].OpCode() != OpReturn {
		t.Fatalf("expected OpReturn after jump, got %v", chunk.Code[1].OpCode())
	}
}

func TestChunkOptimizeRemovesStackShuffleNoops(t *testing.T) {
	chunk := NewChunk("stack")
	chunk.WriteSimple(OpDup, 1)
	chunk.WriteSimple(OpPop, 1)
	chunk.WriteSimple(OpDup2, 2)
	chunk.WriteSimple(OpPop, 2)
	chunk.WriteSimple(OpPop, 2)
	chunk.WriteSimple(OpSwap, 3)
	chunk.WriteSimple(OpSwap, 3)
	chunk.WriteSimple(OpRotate3, 4)
	chunk.WriteSimple(OpRotate3, 4)
	chunk.WriteSimple(OpRotate3, 4)
	chunk.WriteSimple(OpHalt, 5)

	chunk.Optimize()

	if len(chunk.Code) != 1 || chunk.Code[0].OpCode() != OpHalt {
		t.Fatalf("expected only OpHalt after removing redundant shuffles, got %v", chunk.Code)
	}
}

func TestChunkOptimizeStackShuffleRemainders(t *testing.T) {
	chunk := NewChunk("stack-rem")
	chunk.WriteSimple(OpSwap, 1)
	chunk.WriteSimple(OpSwap, 1)
	chunk.WriteSimple(OpSwap, 1) // odd count -> one swap remains
	chunk.WriteSimple(OpRotate3, 2)
	chunk.WriteSimple(OpRotate3, 2)
	chunk.WriteSimple(OpRotate3, 2)
	chunk.WriteSimple(OpRotate3, 2) // 4 rotations -> 1 remains

	chunk.Optimize()

	if len(chunk.Code) != 2 {
		t.Fatalf("expected two instructions (swap + rotate3), got %d", len(chunk.Code))
	}
	if chunk.Code[0].OpCode() != OpSwap {
		t.Fatalf("expected OpSwap to remain first, got %v", chunk.Code[0].OpCode())
	}
	if chunk.Code[1].OpCode() != OpRotate3 {
		t.Fatalf("expected OpRotate3 to remain second, got %v", chunk.Code[1].OpCode())
	}
}

func TestChunkOptimizeEliminatesDeadCodeAfterReturn(t *testing.T) {
	chunk := NewChunk("dead")
	chunk.Write(OpReturn, 0, 0, 1)
	constIdx := chunk.AddConstant(IntValue(1))
	chunk.Write(OpLoadConst, 0, uint16(constIdx), 2)
	chunk.WriteSimple(OpHalt, 3)

	chunk.Optimize()

	if len(chunk.Code) != 1 || chunk.Code[0].OpCode() != OpReturn {
		t.Fatalf("expected only OpReturn to remain, got %v", chunk.Code)
	}
}

func TestChunkOptimizePreservesJumpTargets(t *testing.T) {
	chunk := NewChunk("dead-target")
	chunk.Write(OpJump, 0, 2, 1) // jumps to final instruction
	chunk.Write(OpReturn, 0, 0, 2)
	constIdx := chunk.AddConstant(IntValue(2))
	chunk.Write(OpLoadConst, 0, uint16(constIdx), 3)
	chunk.WriteSimple(OpHalt, 4)

	chunk.Optimize()

	if len(chunk.Code) != 2 {
		t.Fatalf("expected OpJump and OpHalt to remain, got %v", chunk.Code)
	}
	if chunk.Code[0].OpCode() != OpJump || chunk.Code[1].OpCode() != OpHalt {
		t.Fatalf("unexpected instruction sequence after DCE: %v", chunk.Code)
	}
}

func TestChunkOptimizePropagatesLocalConstantLoads(t *testing.T) {
	chunk := NewChunk("const-local")
	valIdx := chunk.AddConstant(IntValue(7))
	chunk.Write(OpLoadConst, 0, uint16(valIdx), 1)
	chunk.Write(OpStoreLocal, 0, 0, 1)
	chunk.Write(OpLoadLocal, 0, 0, 2)
	chunk.Write(OpReturn, 1, 0, 3)

	chunk.Optimize()

	if len(chunk.Code) != 4 {
		t.Fatalf("expected 4 instructions, got %d", len(chunk.Code))
	}
	switch chunk.Code[2].OpCode() {
	case OpLoadConst, OpLoadConst0, OpLoadConst1:
	default:
		t.Fatalf("expected LoadLocal to be replaced with LoadConst variant, got %v", chunk.Code[2].OpCode())
	}
}

func TestChunkOptimizeFoldsConstantAddition(t *testing.T) {
	chunk := NewChunk("const-add")
	aIdx := chunk.AddConstant(IntValue(2))
	bIdx := chunk.AddConstant(IntValue(3))
	chunk.Write(OpLoadConst, 0, uint16(aIdx), 1)
	chunk.Write(OpLoadConst, 0, uint16(bIdx), 1)
	chunk.WriteSimple(OpAddInt, 1)
	chunk.Write(OpReturn, 1, 0, 2)

	chunk.Optimize()

	if len(chunk.Code) != 2 {
		t.Fatalf("expected folded sequence (2 instructions), got %d", len(chunk.Code))
	}
	if chunk.Code[0].OpCode() != OpLoadConst {
		t.Fatalf("expected first instruction to be LoadConst result, got %v", chunk.Code[0].OpCode())
	}
	if chunk.Code[1].OpCode() != OpReturn {
		t.Fatalf("expected trailing OpReturn, got %v", chunk.Code[1].OpCode())
	}
}

func TestChunkOptimizeInlinesSmallFunction(t *testing.T) {
	chunk := NewChunk("inline")
	fnChunk := NewChunk("fn")
	constIdx := fnChunk.AddConstant(IntValue(99))
	fnChunk.Write(OpLoadConst, 0, uint16(constIdx), 1)
	fnChunk.Write(OpReturn, 1, 0, 1)
	fn := NewFunctionObject("const99", fnChunk, 0)
	fnValue := FunctionValue(fn)
	fConst := chunk.AddConstant(fnValue)
	chunk.Write(OpCall, 0, uint16(fConst), 1)
	chunk.Write(OpReturn, 1, 0, 2)

	chunk.Optimize()

	if len(chunk.Code) != 2 {
		t.Fatalf("expected 2 instructions after inlining, got %d", len(chunk.Code))
	}
	if chunk.Code[0].OpCode() != OpLoadConst {
		t.Fatalf("expected first instruction to load constant, got %v", chunk.Code[0].OpCode())
	}
	if chunk.Code[1].OpCode() != OpReturn {
		t.Fatalf("expected trailing OpReturn, got %v", chunk.Code[1].OpCode())
	}
}

func TestChunkOptimizeCanDisablePasses(t *testing.T) {
	chunk := NewChunk("opt")
	constIdx := chunk.AddConstant(IntValue(1))

	chunk.Write(OpLoadConst, 0, uint16(constIdx), 1)
	chunk.WriteSimple(OpPop, 1)
	chunk.WriteSimple(OpHalt, 2)

	chunk.Optimize(
		WithOptimizationPass(PassLiteralDiscard, false),
		WithOptimizationPass(PassStackShuffle, false),
		WithOptimizationPass(PassDeadCode, false),
	)

	if len(chunk.Code) != 3 {
		t.Fatalf("expected optimizer to skip passes when disabled, got %d instructions", len(chunk.Code))
	}
	if chunk.Code[0].OpCode() != OpLoadConst || chunk.Code[1].OpCode() != OpPop {
		t.Fatalf("expected literal instructions untouched when pass disabled, got %v", chunk.Code)
	}
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

// TestSetTryInfo tests the SetTryInfo method
func TestSetTryInfo(t *testing.T) {
	t.Run("set try info on valid chunk", func(t *testing.T) {
		chunk := NewChunk("test")
		info := TryInfo{
			CatchTarget:   21,
			FinallyTarget: -1,
			HasCatch:      true,
			HasFinally:    false,
		}
		chunk.SetTryInfo(10, info)

		retrieved, ok := chunk.TryInfoAt(10)
		if !ok {
			t.Error("TryInfoAt should find the info we just set")
		}
		if retrieved.CatchTarget != 21 || !retrieved.HasCatch {
			t.Errorf("Retrieved info = %+v, want %+v", retrieved, info)
		}
	})

	t.Run("set try info on nil chunk", func(t *testing.T) {
		var chunk *Chunk = nil
		// Should not panic
		chunk.SetTryInfo(0, TryInfo{})
	})

	t.Run("set multiple try infos", func(t *testing.T) {
		chunk := NewChunk("test")
		info1 := TryInfo{CatchTarget: 21, FinallyTarget: -1, HasCatch: true, HasFinally: false}
		info2 := TryInfo{CatchTarget: 41, FinallyTarget: 50, HasCatch: true, HasFinally: true}

		chunk.SetTryInfo(10, info1)
		chunk.SetTryInfo(30, info2)

		r1, ok1 := chunk.TryInfoAt(10)
		r2, ok2 := chunk.TryInfoAt(30)

		if !ok1 || !ok2 {
			t.Error("Both try infos should be retrievable")
		}
		if r1.CatchTarget != 21 || r2.CatchTarget != 41 {
			t.Errorf("Try infos not stored correctly")
		}
	})

	t.Run("overwrite existing try info", func(t *testing.T) {
		chunk := NewChunk("test")
		info1 := TryInfo{CatchTarget: 21, FinallyTarget: -1, HasCatch: true, HasFinally: false}
		info2 := TryInfo{CatchTarget: 26, FinallyTarget: 30, HasCatch: true, HasFinally: true}

		chunk.SetTryInfo(10, info1)
		chunk.SetTryInfo(10, info2)

		retrieved, _ := chunk.TryInfoAt(10)
		if retrieved.CatchTarget != 26 {
			t.Errorf("Try info should be overwritten, got CatchTarget = %d, want 26", retrieved.CatchTarget)
		}
	})
}

// TestTryInfoAt tests the TryInfoAt method
func TestTryInfoAt(t *testing.T) {
	t.Run("retrieve existing try info", func(t *testing.T) {
		chunk := NewChunk("test")
		info := TryInfo{CatchTarget: 21, FinallyTarget: -1, HasCatch: true, HasFinally: false}
		chunk.SetTryInfo(10, info)

		retrieved, ok := chunk.TryInfoAt(10)
		if !ok {
			t.Error("TryInfoAt(10) should return true for existing info")
		}
		if retrieved.CatchTarget != 21 {
			t.Errorf("Retrieved CatchTarget = %d, want 21", retrieved.CatchTarget)
		}
	})

	t.Run("retrieve non-existent try info", func(t *testing.T) {
		chunk := NewChunk("test")
		_, ok := chunk.TryInfoAt(100)
		if ok {
			t.Error("TryInfoAt(100) should return false for non-existent info")
		}
	})

	t.Run("retrieve from nil chunk", func(t *testing.T) {
		var chunk *Chunk = nil
		_, ok := chunk.TryInfoAt(10)
		if ok {
			t.Error("TryInfoAt on nil chunk should return false")
		}
	})

	t.Run("retrieve from chunk with nil tryInfos map", func(t *testing.T) {
		chunk := NewChunk("test")
		// tryInfos map is nil initially
		_, ok := chunk.TryInfoAt(10)
		if ok {
			t.Error("TryInfoAt should return false when tryInfos map is nil")
		}
	})
}

// TestChunkValuesEqual tests the valuesEqual method
func TestChunkValuesEqual(t *testing.T) {
	chunk := NewChunk("test")

	t.Run("nil values equal", func(t *testing.T) {
		if !chunk.valuesEqual(NilValue(), NilValue()) {
			t.Error("nil values should be equal")
		}
	})

	t.Run("bool values equal", func(t *testing.T) {
		if !chunk.valuesEqual(BoolValue(true), BoolValue(true)) {
			t.Error("equal bool values should be equal")
		}
		if !chunk.valuesEqual(BoolValue(false), BoolValue(false)) {
			t.Error("equal bool values should be equal")
		}
	})

	t.Run("bool values not equal", func(t *testing.T) {
		if chunk.valuesEqual(BoolValue(true), BoolValue(false)) {
			t.Error("different bool values should not be equal")
		}
	})

	t.Run("int values equal", func(t *testing.T) {
		if !chunk.valuesEqual(IntValue(42), IntValue(42)) {
			t.Error("equal int values should be equal")
		}
	})

	t.Run("int values not equal", func(t *testing.T) {
		if chunk.valuesEqual(IntValue(42), IntValue(43)) {
			t.Error("different int values should not be equal")
		}
	})

	t.Run("float values equal", func(t *testing.T) {
		if !chunk.valuesEqual(FloatValue(3.14), FloatValue(3.14)) {
			t.Error("equal float values should be equal")
		}
	})

	t.Run("float values not equal", func(t *testing.T) {
		if chunk.valuesEqual(FloatValue(3.14), FloatValue(3.15)) {
			t.Error("different float values should not be equal")
		}
	})

	t.Run("string values equal", func(t *testing.T) {
		if !chunk.valuesEqual(StringValue("hello"), StringValue("hello")) {
			t.Error("equal string values should be equal")
		}
	})

	t.Run("string values not equal", func(t *testing.T) {
		if chunk.valuesEqual(StringValue("hello"), StringValue("world")) {
			t.Error("different string values should not be equal")
		}
	})

	t.Run("different types not equal", func(t *testing.T) {
		if chunk.valuesEqual(IntValue(42), StringValue("42")) {
			t.Error("values of different types should not be equal")
		}
		if chunk.valuesEqual(IntValue(42), FloatValue(42.0)) {
			t.Error("int and float should not be equal (different types)")
		}
	})

	t.Run("complex types not equal", func(t *testing.T) {
		arr1 := ArrayValue(NewArrayInstance([]Value{IntValue(1)}))
		arr2 := ArrayValue(NewArrayInstance([]Value{IntValue(1)}))
		if chunk.valuesEqual(arr1, arr2) {
			t.Error("complex types (arrays) should not deduplicate")
		}
	})

	t.Run("object values not equal", func(t *testing.T) {
		obj1 := ObjectValue(NewObjectInstance("Test"))
		obj2 := ObjectValue(NewObjectInstance("Test"))
		if chunk.valuesEqual(obj1, obj2) {
			t.Error("complex types (objects) should not deduplicate")
		}
	})
}

// TestAddConstantDeduplication tests that AddConstant deduplicates properly
func TestAddConstantDeduplication(t *testing.T) {
	chunk := NewChunk("test")

	t.Run("deduplicate int constants", func(t *testing.T) {
		idx1 := chunk.AddConstant(IntValue(42))
		idx2 := chunk.AddConstant(IntValue(42))
		if idx1 != idx2 {
			t.Errorf("AddConstant should deduplicate: idx1=%d, idx2=%d", idx1, idx2)
		}
		if len(chunk.Constants) != 1 {
			t.Errorf("Constants length = %d, want 1", len(chunk.Constants))
		}
	})

	t.Run("deduplicate string constants", func(t *testing.T) {
		chunk := NewChunk("test")
		idx1 := chunk.AddConstant(StringValue("hello"))
		idx2 := chunk.AddConstant(StringValue("hello"))
		if idx1 != idx2 {
			t.Errorf("AddConstant should deduplicate strings: idx1=%d, idx2=%d", idx1, idx2)
		}
		if len(chunk.Constants) != 1 {
			t.Errorf("Constants length = %d, want 1", len(chunk.Constants))
		}
	})

	t.Run("deduplicate nil constants", func(t *testing.T) {
		chunk := NewChunk("test")
		idx1 := chunk.AddConstant(NilValue())
		idx2 := chunk.AddConstant(NilValue())
		if idx1 != idx2 {
			t.Errorf("AddConstant should deduplicate nil: idx1=%d, idx2=%d", idx1, idx2)
		}
		if len(chunk.Constants) != 1 {
			t.Errorf("Constants length = %d, want 1", len(chunk.Constants))
		}
	})

	t.Run("do not deduplicate different values", func(t *testing.T) {
		chunk := NewChunk("test")
		idx1 := chunk.AddConstant(IntValue(42))
		idx2 := chunk.AddConstant(IntValue(43))
		if idx1 == idx2 {
			t.Error("AddConstant should not deduplicate different values")
		}
		if len(chunk.Constants) != 2 {
			t.Errorf("Constants length = %d, want 2", len(chunk.Constants))
		}
	})

	t.Run("do not deduplicate complex types", func(t *testing.T) {
		chunk := NewChunk("test")
		arr1 := ArrayValue(NewArrayInstance([]Value{IntValue(1)}))
		arr2 := ArrayValue(NewArrayInstance([]Value{IntValue(1)}))
		idx1 := chunk.AddConstant(arr1)
		idx2 := chunk.AddConstant(arr2)
		if idx1 == idx2 {
			t.Error("AddConstant should not deduplicate complex types")
		}
		if len(chunk.Constants) != 2 {
			t.Errorf("Constants length = %d, want 2 (no dedup for arrays)", len(chunk.Constants))
		}
	})
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
